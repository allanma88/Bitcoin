package server

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/config"
	"Bitcoin/src/database"
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"Bitcoin/src/service"
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"
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
	lastBlocks          *collection.SortedSet[*model.Block]
	cfg                 *config.Config
	nodeService         *service.NodeService
	txService           *service.TransactionService
	blockService        *service.BlockService
	txBroadcastQueue    chan *model.Transaction
	blockQueue          chan *model.Block
	blockBroadcastQueue chan *model.Block
	mineQueue           chan *model.Transaction
	pullBlockQueue      chan string
	cancelFunc          context.CancelCauseFunc
	lock                sync.Mutex
}

func NewBitcoinServer(cfg *config.Config, txdb database.ITransactionDB, blockdb database.IBlockDB, blockContentDb database.IBlockContentDB, cancelFunc context.CancelCauseFunc) (*BitcoinServer, error) {
	server := &BitcoinServer{
		cfg:                 cfg,
		nodeService:         service.NewNodeService(cfg),
		txService:           service.NewTransactionService(txdb),
		blockService:        service.NewBlockService(blockdb, blockContentDb, cfg),
		txBroadcastQueue:    make(chan *model.Transaction, TxBroadcastQueueSize),
		blockQueue:          make(chan *model.Block, BlockQueueSize),
		blockBroadcastQueue: make(chan *model.Block, BlockBroadcastQueueSize),
		mineQueue:           make(chan *model.Transaction, MaxTxSizePerBlock),
		pullBlockQueue:      make(chan string, PullBlockQueueSize),
		lastBlocks:          collection.NewSortedSet[*model.Block](), //TODO: how to set when server restart?
		cancelFunc:          cancelFunc,
		lock:                sync.Mutex{},
	}
	return server, nil
}

func (s *BitcoinServer) AddTx(ctx context.Context, request *protocol.TransactionReq) (*protocol.TransactionReply, error) {
	tx := model.TransactionFrom(request)

	log.Printf("received transaction: %x", tx.Hash)

	fee, err := s.txService.Validate(tx)
	if err != nil {
		log.Printf("validate transaction %x failed: %v", tx.Hash, err)
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("validated transaction: %x", tx.Hash)
	tx.Fee = fee

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
	block, err := s.addBlock(request)
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
		s.blockQueue <- block
		s.blockBroadcastQueue <- block
	}()

	return &protocol.BlockReply{Result: true}, nil
}

func (s *BitcoinServer) GetBlocks(ctx context.Context, request *protocol.GetBlocksReq) (*protocol.GetBlocksReply, error) {
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

func (s *BitcoinServer) UpdateState() {
	//TODO: how to handle when server is restart
	for {
		blocks := make([]*model.Block, 0, BlockQueueSize)
		for {
			block, ok := <-s.blockQueue
			if ok {
				blocks = append(blocks, block)
			} else {
				break
			}
		}
		s.setLastBlocks(blocks)
	}
}

func (s *BitcoinServer) PullBlocks() {
	for addr := range s.pullBlockQueue {
		for {
			//TODO: lastBlock should be not same each time if we saved any of blockReqs
			//TODO: maybe stop the mining when pull blocks?
			lastBlockHashes := s.getLastBlockHashes(10)
			blockReqs, end, err := s.nodeService.GetBlocks(lastBlockHashes, addr)
			if err != nil {
				log.Printf("get blocks from %v error: %v", addr, err)
				break
			}

			blocks := make([]*model.Block, 0, len(blockReqs))
			var block *model.Block
			for _, blockReq := range blockReqs {
				block, err = s.validateBlock(blockReq)
				if err == errors.ErrBlockExist {
					// it's possible that my main chain is not main chain anymore,
					// so I pull all blocks of the new main chain, the new main chain is side chain previously,
					// some blocks of the new main chain already exist, so ignore this error
					continue
				}
				if err != nil {
					break
				}

				err = s.blockService.SaveBlock(block)
				if err != nil {
					log.Fatalf("save block %x failed: %v", block.Hash, err)
				}

				blocks = append(blocks, block)
			}

			for _, block := range blocks {
				s.blockQueue <- block
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
		lastBlock := s.getLastBlock()
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

		//TODO: save txs and block in one db transaction
		err = s.txService.ChainOnTxs(block.GetTxs()...)
		if err != nil {
			log.Printf("chain on transaction err: %v", err)
			continue
		}

		s.blockQueue <- block
		s.blockBroadcastQueue <- block
	}
}

func (s *BitcoinServer) validateBlock(blockReq *protocol.BlockReq) (*model.Block, error) {
	block, err := model.BlockFrom(blockReq)
	if err != nil {
		return nil, err
	}
	log.Printf("received block: %x", block.Hash)

	err = s.blockService.Validate(block)
	if err != nil {
		return nil, err
	}
	log.Printf("validated block: %x", block.Hash)
	return block, nil
}

func (s *BitcoinServer) addBlock(blockReq *protocol.BlockReq) (*model.Block, error) {
	block, err := s.validateBlock(blockReq)
	if err != nil {
		log.Printf("validate block %x failed: %v", block.Hash, err)
		return nil, err
	}

	//TODO: save txs and block in one db transaction
	err = s.txService.ChainOnTxs(block.GetTxs()...)
	if err != nil {
		log.Printf("chain on transaction err: %v", err)
		return nil, err
	}

	err = s.blockService.SaveBlock(block)
	if err != nil {
		log.Printf("save block %x failed: %v", block.Hash, err)
		return nil, err
	}
	log.Printf("saved block: %x", block.Hash)

	return block, nil
}

func (s *BitcoinServer) receiveTxs(reward uint64) ([]*model.Transaction, error) {
	txs := make([]*model.Transaction, MaxTxSizePerBlock)
	var totalFee uint64 = 0
	for i := 1; i < MaxTxSizePerBlock; i++ {
		txs[i] = <-s.mineQueue //TODO: need to validate the tx again since tx maybe invalid when we start to mine the block
		totalFee += txs[i].Fee
	}

	coinbaseTx, err := model.MakeCoinbaseTx(s.cfg.MinerPubkey, reward+totalFee)
	if err != nil {
		return nil, err
	}

	txs[0] = coinbaseTx
	return txs, nil
}

func (s *BitcoinServer) getLastBlock() *model.Block {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.lastBlocks.First()
}

func (s *BitcoinServer) getLastBlockHashes(n int) [][]byte {
	s.lock.Lock()
	defer s.lock.Unlock()

	blocks := s.lastBlocks.Top(n)
	blockHashes := make([][]byte, len(blocks))
	for i := 0; i < len(blocks); i++ {
		blockHashes[i] = blocks[i].Hash
	}
	return blockHashes
}

func (s *BitcoinServer) setLastBlocks(blocks []*model.Block) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, block := range blocks {
		lastBlock := s.lastBlocks.First()
		if s.cfg.Server != block.Miner && bytes.Equal(lastBlock.Hash, block.Prevhash) {
			s.cancelFunc(errors.ErrServerCancelMining)
		}

		s.lastBlocks.Remove(block.PrevBlock)
		s.lastBlocks.Insert(block)
	}
}
