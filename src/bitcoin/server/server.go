package server

import (
	"Bitcoin/src/config"
	"Bitcoin/src/database"
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"Bitcoin/src/service"
	"context"
	"log"
	"sync"
)

const (
	TxBroadcastQueueSize    = 10000
	BlockBroadcastQueueSize = 10000
	PullBlockQueueSize      = 100
	BlocksPerSave           = 100
)

type BitcoinServer struct {
	protocol.TransactionServer
	protocol.BlockServer
	cfg                 *config.Config
	nodeService         *service.NodeService
	chainService        *service.ChainService
	txService           *service.TransactionService
	blockService        *service.BlockService
	syncService         *service.SyncService
	mineService         *service.MineService
	mempool             *service.MemPool
	txBroadcastQueue    chan *model.Transaction
	blockBroadcastQueue chan *model.Block
	syncBlockQueue      chan string
	cancelFunc          context.CancelCauseFunc
	exiting             bool
}

func NewBitcoinServer(cfg *config.Config, blockdb database.IBlockDB, cancelFunc context.CancelCauseFunc) (*BitcoinServer, error) {
	server := &BitcoinServer{
		cfg:                 cfg,
		nodeService:         service.NewNodeService(cfg.Endpoint, cfg.Bootstraps),
		chainService:        service.NewChainService(),
		txService:           service.NewTransactionService(blockdb),
		blockService:        service.NewBlockService(blockdb),
		mempool:             service.NewMemPool(int(cfg.MaxTxSizePerBlock)),
		txBroadcastQueue:    make(chan *model.Transaction, TxBroadcastQueueSize),
		blockBroadcastQueue: make(chan *model.Block, BlockBroadcastQueueSize),
		syncBlockQueue:      make(chan string, PullBlockQueueSize),
		cancelFunc:          cancelFunc,
	}
	server.syncService = service.NewSyncService(server.chainService, server.nodeService, server.addBlock)
	server.mineService = service.NewMineService(cfg, server.txService, server.mempool)

	if err := server.load(); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *BitcoinServer) AddTx(ctx context.Context, request *protocol.TransactionReq) (*protocol.TransactionReply, error) {
	tx := model.TransactionFrom(request)

	log.Printf("received transaction: %x", tx.Hash)
	f := func(hash []byte) *model.Transaction {
		return s.mempool.Get(hash)
	}

	if err := s.txService.ValidateTx(tx, f); err != nil {
		log.Printf("validate transaction %x failed: %v", tx.Hash, err)
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("validated transaction: %x", tx.Hash)

	s.mempool.Put(tx)
	log.Printf("puted transaction on mempool: %x", tx.Hash)

	s.txBroadcastQueue <- tx

	log.Printf("broadcast the transaction: %x", tx.Hash)

	if err := s.nodeService.AddAddrs(request.Nodes); err != nil {
		log.Printf("add nodes failed: %v", err)
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("added to the node list: %x", tx.Hash)

	return &protocol.TransactionReply{Result: true}, nil
}

func (s *BitcoinServer) NewBlock(ctx context.Context, request *protocol.BlockReq) (*protocol.BlockReply, error) {
	block, err := model.BlockFrom(request)
	if err != nil {
		return &protocol.BlockReply{Result: false}, err
	}
	log.Printf("received block: %x", block.Hash)

	err = s.addBlock(block)
	//if the prev block doesn't exist, then maybe we fall behind with current chain or a new main chain show up,
	// so we need sync with the request node
	if err == errors.ErrPrevBlockNotFound {
		if len(s.syncBlockQueue) < PullBlockQueueSize {
			s.syncBlockQueue <- request.Node
		}
	}
	if err != nil {
		return &protocol.BlockReply{Result: false}, err
	}

	return &protocol.BlockReply{Result: true}, nil
}

func (s *BitcoinServer) GetBlocks(ctx context.Context, request *protocol.GetBlocksReq) (*protocol.GetBlocksReply, error) {
	mainChain := s.chainService.GetMainChain()
	blocks, end, err := s.blockService.GetBlocks(mainChain.LastBlockHash, request.Blockhashes)
	if blocks == nil || err != nil {
		return &protocol.GetBlocksReply{}, err
	}

	blockReqs := make([]*protocol.BlockReq, len(blocks))
	for i := 0; i < len(blocks); i++ {
		blockReq, err := model.BlockTo(blocks[i])
		if err != nil {
			return &protocol.GetBlocksReply{}, err
		}
		blockReqs[i] = blockReq
	}
	reply := &protocol.GetBlocksReply{
		Blocks: blockReqs,
		End:    end,
	}
	return reply, nil
}

func (s *BitcoinServer) BroadcastTx() {
	for tx := range s.txBroadcastQueue {
		s.nodeService.SendTx(tx)

		if s.exiting {
			break
		}
	}
}

func (s *BitcoinServer) BroadcastBlock() {
	for block := range s.blockBroadcastQueue {
		s.nodeService.SendBlock(block)

		if s.exiting {
			break
		}
	}
}

func (s *BitcoinServer) SyncBlocks(wait *sync.WaitGroup) {
	for addr := range s.syncBlockQueue {
		// cancel mining
		s.cancelFunc(errors.ErrServerCancelMining)
		// pending the mining task
		wait.Add(1)
		s.syncService.SyncBlocks(addr)
		wait.Done()

		if s.exiting {
			break
		}
	}
}

func (s *BitcoinServer) MineBlock(ctx context.Context, wait *sync.WaitGroup) {
	for !s.exiting {
		mainChain := s.chainService.GetMainChain()
		lastBlock, err := s.blockService.GetBlock(mainChain.LastBlockHash, false)
		if err != nil {
			log.Printf("get last block of main chain error: %v", err)
			continue
		}

		block, err := s.mineService.MineBlock(lastBlock, ctx, wait)
		if err != nil {
			log.Printf("mine block error: %v", err)
			continue
		}

		if err = s.addBlock(block); err != nil {
			log.Printf("add block error: %v", err)
			continue
		}

		if block.Number%BlocksPerSave == 0 {
			if err := s.chainService.Save(s.cfg.DataDir); err != nil {
				log.Printf("chain service save error: %v", err)
				continue
			}
		}
	}
}

func (s *BitcoinServer) Shutdown() error {
	if err := s.mempool.Save(s.cfg.DataDir); err != nil {
		return err
	}

	if err := s.chainService.Save(s.cfg.DataDir); err != nil {
		return err
	}

	s.exiting = true
	return nil
}

func (s *BitcoinServer) load() error {
	if err := s.mempool.Load(s.cfg.DataDir); err != nil {
		return err
	}

	chains, err := s.chainService.Load(s.cfg.DataDir)
	if err != nil {
		return err
	}

	for _, chain := range chains {
		blockHashes := [][]byte{chain.LastBlockHash}
		for len(blockHashes) > 0 {
			blockHash := blockHashes[0]
			blockHashes = blockHashes[1:]
			blocks, err := s.blockService.FilterBlock(blockHash)
			if err != nil {
				return err
			}
			for _, block := range blocks {
				//TODO: better apply algorithm, only apply the utxo once, no need rollback
				s.applyBlock(block)
				blockHashes = append(blockHashes, block.Hash)
			}
		}
	}

	return nil
}

func (s *BitcoinServer) addBlock(block *model.Block) error {
	err := s.blockService.Validate(block)
	if err != nil {
		return err
	}
	log.Printf("validated block: %x", block.Hash)

	reward := block.GetNextReward(s.cfg.InitRewrad, s.cfg.BlocksPerRewrad)
	txs := block.GetTxs()

	if err = s.txService.ValidateOnChainTxs(txs, block.Hash, reward); err != nil {
		return err
	}

	if err = s.applyBlock(block); err != nil {
		log.Printf("apply block %x failed: %v", block.Hash, err)
		return err
	}

	s.mempool.Remove(block.GetTxs())

	if err = s.blockService.SaveBlock(block); err != nil {
		log.Printf("save block %x failed: %v", block.Hash, err)
		return err
	}
	log.Printf("saved block: %x", block.Hash)

	s.blockBroadcastQueue <- block

	return nil
}

func (s *BitcoinServer) applyBlock(block *model.Block) error {
	applyChain, rollbackChain := s.chainService.ApplyChain(block)

	if applyChain != nil && rollbackChain != nil {
		applyBlocks, rollbackBlocks, err := s.blockService.GetBlocksOfChain(applyChain, rollbackChain)
		if err != nil {
			return err
		}
		if err := s.chainService.SwitchBalances(rollbackBlocks, applyBlocks); err != nil {
			return err
		}
	} else {
		if s.cfg.Server != block.Miner {
			s.cancelFunc(errors.ErrServerCancelMining)
		}

		if err := s.chainService.ApplyBalance(block); err != nil {
			return err
		}
	}
	return nil
}
