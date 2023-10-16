package server

import (
	"Bitcoin/src/config"
	"Bitcoin/src/db"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"Bitcoin/src/service"
	"context"
	"log"
)

type BitcoinServer struct {
	protocol.TransactionServer
	txService   *service.TransactionService
	nodeService *service.NodeService
	pendingTxs  []*model.Transaction
	txQueue     chan *model.Transaction
}

func NewBitcoinServer(cfg *config.Config) (*BitcoinServer, error) {
	db, err := db.NewTransactionDB(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	server := &BitcoinServer{
		txService:   service.NewTransactionService(db),
		nodeService: service.NewNodeService(cfg),
		pendingTxs:  make([]*model.Transaction, 0, model.PendingTxSize),
		txQueue:     make(chan *model.Transaction, model.TxQueueSize),
	}
	return server, nil
}

func (s *BitcoinServer) ExecuteTx(ctx context.Context, request *protocol.TransactionReq) (*protocol.TransactionReply, error) {
	tx, err := model.TransactionFrom(request)
	if err != nil {
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("received transaction: %x", tx.Id)

	err = s.txService.Validate(tx)
	if err != nil {
		log.Printf("validate transaction %x failed: %v", tx.Id, err)
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("validated transaction: %x", tx.Id)

	err = s.txService.SaveTx(tx)
	if err != nil {
		log.Printf("save transaction %x failed: %v", tx.Id, err)
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("saved transaction: %x", tx.Id)

	s.pendingTxs = append(s.pendingTxs, tx)
	log.Printf("append to pending txs: %x", tx.Id)

	go func() { s.txQueue <- tx }()
	log.Printf("broadcast the transaction: %x", tx.Id)

	err = s.nodeService.AddAddrs(request.Nodes)
	if err != nil {
		log.Printf("add nodes failed: %v", err)
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("added to the node list: %x", tx.Id)

	return &protocol.TransactionReply{Result: true}, nil
}

func (s *BitcoinServer) BroadcastTx() {
	for tx := range s.txQueue {
		s.nodeService.SendTx(tx)
	}
}
