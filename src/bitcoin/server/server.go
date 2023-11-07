package server

import (
	"Bitcoin/src/bitcoin"
	"Bitcoin/src/config"
	"Bitcoin/src/database"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"Bitcoin/src/service"
	"context"
	"log"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	InitReward = 50
)

type BitcoinServer struct {
	protocol.TransactionServer
	*service.BlockService
	*bitcoin.State
	lock                sync.Mutex
	cfg                 *config.Config
	txService           *service.TransactionService
	nodeService         *service.NodeService
	txBroadcastQueue    chan *model.Transaction
	blockQueue          chan *model.Block
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
	state := &bitcoin.State{
		Difficulty: bitcoin.ComputeDifficulty(bitcoin.MakeDifficulty(cfg.InitDifficulty)),
	}
	server := &BitcoinServer{
		lock:                sync.Mutex{},
		cfg:                 cfg,
		nodeService:         service.NewNodeService(cfg),
		txService:           service.NewTransactionService(txdb),
		BlockService:        service.NewBlockService(blockdb, blockContentDb, cfg),
		txBroadcastQueue:    make(chan *model.Transaction, model.TxBroadcastQueueSize),
		blockQueue:          make(chan *model.Block, model.BlockBroadcastQueueSize),
		blockBroadcastQueue: make(chan *model.Block, model.BlockBroadcastQueueSize),
		mineQueue:           make(chan *model.Transaction, model.MaxTxSizePerBlock),
		State:               state,
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

	err = s.BlockService.Validate(block)
	if err != nil {
		log.Printf("validate block %x failed: %v", block.Hash, err)
		return &protocol.BlockReply{Result: false}, err
	}
	log.Printf("validated block: %x", block.Hash)

	err = s.BlockService.SaveBlock(block)
	if err != nil {
		log.Printf("save block %x failed: %v", block.Hash, err)
		return &protocol.BlockReply{Result: false}, err
	}
	log.Printf("saved block: %x", block.Hash)

	go func() {
		s.blockQueue <- block
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

func (s *BitcoinServer) ReceiveBlock() {
	for {
		s.lock.Lock()
		if (s.LastBlockId+1)%uint64(s.cfg.BlocksPerDifficulty+1) == 0 {
			s.TotalInterval = 0
		}
		s.lock.Unlock()

		var prevBlock *model.Block = nil //TODO: how to handle when server is restart
		for i := 0; i < int(s.cfg.BlocksPerDifficulty); i++ {
			block := <-s.blockBroadcastQueue

			if prevBlock != nil {
				s.lock.Lock()
				s.TotalInterval += block.Time.Sub(prevBlock.Time)
				s.lock.Unlock()
			}

			prevBlock = block
		}
	}
}

func (s *BitcoinServer) MineBlock() {
	for {
		txs, err := s.receiveTxs()
		if err != nil {
			//TODO: maybe fatal err?
			log.Printf("receive txs error: %v", err)
			continue
		}

		s.lock.Lock()
		bitcoin.AdjustDifficulty(s.State, int(s.cfg.BlocksPerDifficulty), s.cfg.BlockInterval)
		s.lock.Unlock()

		block, err := s.BlockService.MineBlock(s.LastBlockId+1, s.Difficulty, txs)
		if err != nil {
			//TODO: maybe fatal err?
			log.Printf("mine block error: %v", err)
			continue
		}

		s.lock.Lock()
		s.LastBlockId = block.Id
		s.lock.Unlock()

		s.txService.RemoveTxs(txs)

		s.blockQueue <- block
		s.blockBroadcastQueue <- block
	}
}

func (s *BitcoinServer) receiveTxs() ([]*model.Transaction, error) {
	txs := make([]*model.Transaction, model.MaxTxSizePerBlock)
	var totalFee uint64 = 0
	for i := 1; i < model.MaxTxSizePerBlock; i++ {
		txs[i] = <-s.mineQueue
		totalFee += txs[i].Fee
	}

	reward := bitcoin.ComputeReward(s.LastBlockId, int(s.cfg.BlocksPerRewrad))
	coinbaseTx, err := model.MakeCoinbaseTx(s.cfg.MinerPubkey, reward+totalFee)
	if err != nil {
		return nil, err
	}

	txs[0] = coinbaseTx
	return txs, nil
}
