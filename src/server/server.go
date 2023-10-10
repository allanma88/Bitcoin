package main

import (
	"Bitcoin/src/db"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"Bitcoin/src/service"
	"context"
	"log"
)

type BitcoinServer struct {
	protocol.TransactionServer
	*service.TransactionService
	*service.NodeService
	pendingTxs []*model.Transaction
	txQueue    chan *model.Transaction
}

func NewBitcoinServer(path string) (*BitcoinServer, error) {
	db, err := db.NewTransactionDB(path)
	if err != nil {
		return nil, err
	}
	server := &BitcoinServer{
		TransactionService: service.NewTransactionService(db),
		NodeService:        service.NewNodeService(),
		pendingTxs:         make([]*model.Transaction, 0, model.PendingTxSize),
		txQueue:            make(chan *model.Transaction, model.TxQueueSize),
	}
	return server, nil
}

func (s *BitcoinServer) ExecuteTx(ctx context.Context, request *protocol.TransactionReq) (*protocol.TransactionReply, error) {
	tx, err := model.TransactionFrom(request)
	if err != nil {
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("received transaction: %x", tx.Id)

	err = s.Validate(tx)
	if err != nil {
		log.Printf("validate transaction %x failed: %v", tx.Id, err)
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("validated transaction: %x", tx.Id)

	err = s.SaveTx(tx)
	if err != nil {
		log.Printf("save transaction %x failed: %v", tx.Id, err)
		return &protocol.TransactionReply{Result: false}, err
	}
	log.Printf("saved transaction: %x", tx.Id)

	s.pendingTxs = append(s.pendingTxs, tx)
	s.txQueue <- tx //TODO: async?
	s.NodeService.AddNodes(request.Nodes)

	return &protocol.TransactionReply{Result: true}, nil
}

func (s *BitcoinServer) broadcastTx() {
	for tx := range s.txQueue {
		s.NodeService.SendTx(tx)
	}
}
