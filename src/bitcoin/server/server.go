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
	syncService         *service.SyncService
	mineService         *service.MineService
	mempool             *service.MemPool
	txBroadcastQueue    chan *model.Transaction
	blockBroadcastQueue chan *model.Block
	syncBlockQueue      chan string
	cancelFunc          context.CancelCauseFunc
}

func NewBitcoinServer(cfg *config.Config, blockdb database.IBlockDB, blockContentDb database.IBlockContentDB, cancelFunc context.CancelCauseFunc) (*BitcoinServer, error) {
	server := &BitcoinServer{
		cfg:                 cfg,
		nodeService:         service.NewNodeService(cfg.Endpoint, cfg.Bootstraps),
		utxoService:         service.NewUtxoService(),
		chainService:        service.NewChainService(), //TODO: how to set when server restart?
		txService:           service.NewTransactionService(blockContentDb),
		blockService:        service.NewBlockService(blockdb, blockContentDb),
		mempool:             service.NewMemPool(int(cfg.MaxTxSizePerBlock)),
		txBroadcastQueue:    make(chan *model.Transaction, TxBroadcastQueueSize),
		blockBroadcastQueue: make(chan *model.Block, BlockBroadcastQueueSize),
		syncBlockQueue:      make(chan string, PullBlockQueueSize),
		cancelFunc:          cancelFunc,
	}
	server.syncService = service.NewSyncService(server.chainService, server.nodeService, server.addBlock)
	server.mineService = service.NewMineService(cfg, server.txService, server.mempool)

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
	blocks, end, err := s.blockService.GetBlocks(mainChain, request.Blockhashes)
	if err != nil {
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
	}
}

func (s *BitcoinServer) BroadcastBlock() {
	for block := range s.blockBroadcastQueue {
		s.nodeService.SendBlock(block)
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
	}
}

func (s *BitcoinServer) MineBlock(ctx context.Context, wait *sync.WaitGroup) {
	for {
		lastBlock := s.chainService.GetMainChain()
		block, err := s.mineService.MineBlock(lastBlock, ctx, wait)
		if err != nil {
			log.Printf("mine block error: %v", err)
			continue
		}

		if err = s.addBlock(block); err != nil {
			log.Printf("add block error: %v", err)
			continue
		}
	}
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

	if err = s.blockService.SaveBlock(block); err != nil {
		log.Printf("save block %x failed: %v", block.Hash, err)
		return err
	}
	log.Printf("saved block: %x", block.Hash)

	if err = s.applyBlock(block); err != nil {
		log.Printf("apply block %x failed: %v", block.Hash, err)
		return err
	}

	s.mempool.Remove(block.GetTxs())

	s.blockBroadcastQueue <- block

	return nil
}

func (s *BitcoinServer) applyBlock(block *model.Block) error {
	applyBlocks, rollbackBlocks := s.chainService.SetChain(block)

	if applyBlocks != nil && rollbackBlocks != nil {
		if err := s.utxoService.SwitchBalances(rollbackBlocks, applyBlocks); err != nil {
			return err
		}
	} else {
		if s.cfg.Server != block.Miner {
			s.cancelFunc(errors.ErrServerCancelMining)
		}

		if err := s.utxoService.ApplyBalances(block); err != nil {
			return err
		}
	}
	return nil
}
