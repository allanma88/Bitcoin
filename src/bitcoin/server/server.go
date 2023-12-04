package server

import (
	"Bitcoin/src/config"
	"Bitcoin/src/database"
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"Bitcoin/src/service"
	"context"
	"fmt"
	"log"
)

const (
	InitReward              = 50
	TxBroadcastQueueSize    = 10
	BlockBroadcastQueueSize = 10
	BlockQueueSize          = 10
	PullBlockQueueSize      = 10000
	MaxTxSizePerBlock       = 10
)

type BitcoinServer struct {
	protocol.TransactionServer
	protocol.BlockServer
	cfg                 *config.Config
	nodeService         *service.NodeService
	utxoService         *service.UtxoService
	chainService        *service.ChainService
	txService           *service.TransactionService
	blockService        *service.BlockService
	txBroadcastQueue    chan *model.Transaction
	blockBroadcastQueue chan *model.Block
	mineQueue           chan *model.Transaction
	pullBlockQueue      chan string
	cancelFunc          context.CancelCauseFunc
}

func NewBitcoinServer(cfg *config.Config, txdb database.ITransactionDB, blockdb database.IBlockDB, blockContentDb database.IBlockContentDB, cancelFunc context.CancelCauseFunc) (*BitcoinServer, error) {
	server := &BitcoinServer{
		cfg:                 cfg,
		nodeService:         service.NewNodeService(cfg),
		utxoService:         service.NewUtxoService(),
		chainService:        service.NewChainService(), //TODO: how to set when server restart?
		txService:           service.NewTransactionService(txdb),
		blockService:        service.NewBlockService(blockdb, blockContentDb, cfg),
		txBroadcastQueue:    make(chan *model.Transaction, TxBroadcastQueueSize),
		blockBroadcastQueue: make(chan *model.Block, BlockBroadcastQueueSize),
		mineQueue:           make(chan *model.Transaction, MaxTxSizePerBlock),
		pullBlockQueue:      make(chan string, PullBlockQueueSize),
		cancelFunc:          cancelFunc,
	}
	return server, nil
}

func (s *BitcoinServer) AddTx(ctx context.Context, request *protocol.TransactionReq) (*protocol.TransactionReply, error) {
	tx := model.TransactionFrom(request)

	log.Printf("received transaction: %x", tx.Hash)

	_, err := s.txService.ValidateOffChainTx(tx)
	if err != nil {
		log.Printf("validate transaction %x failed: %v", tx.Hash, err)
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("validated transaction: %x", tx.Hash)

	err = s.txService.SaveOffChainTx(tx)
	if err != nil {
		log.Printf("save transaction %x failed: %v", tx.Hash, err)
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("saved transaction: %x", tx.Hash)

	go func() {
		s.txBroadcastQueue <- tx
		s.mineQueue <- tx
	}()
	log.Printf("broadcast the transaction: %x", tx.Hash)

	err = s.nodeService.AddAddrs(request.Nodes)
	if err != nil {
		log.Printf("add nodes failed: %v", err)
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("added to the node list: %x", tx.Hash)

	return &protocol.TransactionReply{Result: true}, nil
}

func (s *BitcoinServer) NewBlock(ctx context.Context, request *protocol.BlockReq) (*protocol.BlockReply, error) {
	block, err := s.addBlockReq(request)
	//if the prev block doesn't exist, then maybe we fall behind with current chain or a new main chain show up,
	// so we need sync with the request node
	if err == errors.ErrPrevBlockNotFound {
		go func() {
			s.pullBlockQueue <- request.Node
		}()
	}
	if err != nil {
		return &protocol.BlockReply{Result: false}, err
	}

	go func() {
		s.blockBroadcastQueue <- block
	}()

	return &protocol.BlockReply{Result: true}, nil
}

func (s *BitcoinServer) GetBlocks(ctx context.Context, request *protocol.GetBlocksReq) (*protocol.GetBlocksReply, error) {
	//TODO: return blocks of the main chain of the current node
	return nil, fmt.Errorf("GetBlocks API not implemented")
}

func (s *BitcoinServer) BroadcastTx() {
	for tx := range s.txBroadcastQueue {
		s.nodeService.SendTx(tx)
	}
}

func (s *BitcoinServer) BroadcastBlock() {
	for block := range s.blockBroadcastQueue {
		s.nodeService.SendBlock(block)
	}
}

func (s *BitcoinServer) PullBlocks() {
	for addr := range s.pullBlockQueue {
		for {
			//TODO: lastBlock should be not same each time if we saved any of blockReqs
			//TODO: maybe stop the mining when pull blocks?
			lastBlockHashes := s.chainService.GetChainHashes(10)
			blockReqs, end, err := s.nodeService.GetBlocks(lastBlockHashes, addr)
			if err != nil {
				log.Printf("get blocks from %v error: %v", addr, err)
				break
			}

			blocks := make([]*model.Block, 0, len(blockReqs))
			var block *model.Block
			for _, blockReq := range blockReqs {
				block, err = s.addBlockReq(blockReq)
				if err == errors.ErrBlockExist {
					// it's possible that my main chain is not main chain anymore,
					// so I pull all blocks of the new main chain, the new main chain is side chain previously,
					// some blocks of the new main chain already exist, so ignore this error
					continue
				}
				if err != nil {
					break
				}

				blocks = append(blocks, block)
			}

			for _, block := range blocks {
				s.blockBroadcastQueue <- block
			}

			if err != nil {
				log.Printf("add blocks from %v error: %v", addr, err)
				break
			}

			if len(blocks) > 0 && blocks[len(blocks)-1].Number == end {
				break
			}
		}
	}
}

func (s *BitcoinServer) MineBlock(ctx context.Context) {
	for {
		lastBlock := s.chainService.GetMainChain()
		reward := lastBlock.GetNextReward(s.cfg.BlocksPerRewrad)

		txs, err := s.receiveTxs(reward)
		if err != nil {
			//TODO: maybe fatal err?
			log.Printf("receive txs error: %v", err)
			continue
		}

		block, err := s.blockService.MineBlock(lastBlock, txs, ctx)
		if err != nil {
			//TODO: maybe fatal err?
			log.Printf("mine block error: %v", err)
			continue
		}

		if err = s.addBlock(block); err != nil {
			log.Printf("add block error: %v", err)
			continue
		}
		s.blockBroadcastQueue <- block
	}
}

func (s *BitcoinServer) addBlockReq(blockReq *protocol.BlockReq) (*model.Block, error) {
	block, err := model.BlockFrom(blockReq)
	if err != nil {
		return nil, err
	}
	log.Printf("received block: %x", block.Hash)

	if err = s.addBlock(block); err != nil {
		return nil, err
	}

	return block, nil
}

func (s *BitcoinServer) addBlock(block *model.Block) error {
	err := s.blockService.Validate(block)
	if err != nil {
		return err
	}
	log.Printf("validated block: %x", block.Hash)

	reward := block.GetNextReward(s.cfg.BlocksPerRewrad)
	txs := block.GetTxs()

	if err = s.txService.ValidateOnChainTxs(txs, block.Hash, reward); err != nil {
		return err
	}

	if err = s.applyBlock(block); err != nil {
		log.Printf("apply block %x failed: %v", block.Hash, err)
		return err
	}

	if err = s.blockService.SaveBlock(block); err != nil {
		log.Printf("save block %x failed: %v", block.Hash, err)
		return err
	}
	log.Printf("saved block: %x", block.Hash)

	return nil
}

func (s *BitcoinServer) applyBlock(block *model.Block) error {
	applyBlocks, rollbackBlocks := s.chainService.SetChain(block)

	if applyBlocks != nil && rollbackBlocks != nil {
		for i := 0; i < len(rollbackBlocks); i++ {
			txs := rollbackBlocks[i].GetTxs()
			if err := s.utxoService.RollbackBalances(txs); err != nil {
				return nil
			}
		}
		for i := len(applyBlocks) - 1; i >= 0; i-- {
			txs := applyBlocks[i].GetTxs()
			if err := s.utxoService.ApplyBalances(txs); err != nil {
				return err
			}
		}
	} else {
		if s.cfg.Server != block.Miner {
			s.cancelFunc(errors.ErrServerCancelMining)
		}

		txs := block.GetTxs()
		if err := s.utxoService.ApplyBalances(txs); err != nil {
			return err
		}
	}
	return nil
}

func (s *BitcoinServer) receiveTxs(reward uint64) ([]*model.Transaction, error) {
	txs := make([]*model.Transaction, MaxTxSizePerBlock)
	var totalFee uint64 = 0
	for i := 1; i < MaxTxSizePerBlock; i++ {
		txs[i] = <-s.mineQueue
		fee, err := s.txService.ValidateOffChainTx(txs[i])
		if err != nil {
			return nil, err
		}
		totalFee += fee
	}

	coinbaseTx, err := model.MakeCoinbaseTx(s.cfg.MinerPubkey, reward+totalFee)
	if err != nil {
		return nil, err
	}

	txs[0] = coinbaseTx
	return txs, nil
}
