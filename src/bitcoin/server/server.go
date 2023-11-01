package server

import (
	"Bitcoin/src/config"
	"Bitcoin/src/database"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"Bitcoin/src/service"
	"context"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)

type BitcoinServer struct {
	protocol.TransactionServer
	txService           *service.TransactionService
	blockService        *service.BlockService
	nodeService         *service.NodeService
	txBroadcastQueue    chan *model.Transaction
	blockBroadcastQueue chan *model.Block
	mineQueue           chan *model.Transaction
}

func NewBitcoinServer(cfg *config.Config) (*BitcoinServer, error) {
	db, err := leveldb.OpenFile(cfg.DataDir, nil)
	if err != nil {
		return nil, err
	}

	txdb := database.NewTransactionDB(db)
	blockdb := database.NewBlockDB(db)
	blockContentDb := database.NewBlockContentDB(db)
	server := &BitcoinServer{
		nodeService:         service.NewNodeService(cfg),
		txService:           service.NewTransactionService(txdb),
		blockService:        service.NewBlockService(blockdb, blockContentDb, cfg),
		txBroadcastQueue:    make(chan *model.Transaction, model.TxBroadcastQueueSize),
		blockBroadcastQueue: make(chan *model.Block, model.BlockBroadcastQueueSize),
		mineQueue:           make(chan *model.Transaction, model.MaxTxSizePerBlock),
	}
	return server, nil
}

func (s *BitcoinServer) AddTx(ctx context.Context, request *protocol.TransactionReq) (*protocol.TransactionReply, error) {
	tx := model.TransactionFrom(request)

	log.Printf("received transaction: %x", tx.Hash)

	err := s.txService.Validate(tx)
	if err != nil {
		log.Printf("validate transaction %x failed: %v", tx.Hash, err)
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("validated transaction: %x", tx.Hash)

	err = s.txService.SaveTx(tx)
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

func (s *BitcoinServer) AddBlock(ctx context.Context, request *protocol.BlockReq) (*protocol.BlockReply, error) {
	block, err := model.BlockFrom(request)
	if err != nil {
		return &protocol.BlockReply{Result: false}, err
	}
	log.Printf("received block: %x", block.Hash)

	err = s.blockService.Validate(block)
	if err != nil {
		log.Printf("validate block %x failed: %v", block.Hash, err)
		return &protocol.BlockReply{Result: false}, err
	}
	log.Printf("validated block: %x", block.Hash)

	err = s.blockService.SaveBlock(block)
	if err != nil {
		log.Printf("save block %x failed: %v", block.Hash, err)
		return &protocol.BlockReply{Result: false}, err
	}
	log.Printf("saved block: %x", block.Hash)

	go func() {
		s.blockBroadcastQueue <- block
	}()
	log.Printf("broadcast the block: %x", block.Hash)

	err = s.nodeService.AddAddrs(request.Nodes)
	if err != nil {
		log.Printf("add nodes failed: %v", err)
		return &protocol.BlockReply{Result: false}, err
	}
	log.Printf("added to the node list: %x", block.Hash)

	return &protocol.BlockReply{Result: true}, nil
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

func (s *BitcoinServer) MineBlock() {
	for {
		txs := make([]*model.Transaction, model.MaxTxSizePerBlock)
		for i := 0; i < model.MaxTxSizePerBlock; i++ {
			txs[i] = <-s.mineQueue
		}

		block, err := s.blockService.MineBlock(txs)
		if err != nil {
			log.Printf("mine block error: %v", err)
			continue
		}

		s.txService.RemoveTxs(txs)

		s.blockBroadcastQueue <- block
	}
}
