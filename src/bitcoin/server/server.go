package server

import (
	"Bitcoin/src/bitcoin"
	"Bitcoin/src/config"
	"Bitcoin/src/database"
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"Bitcoin/src/service"
	"context"
	"log"
)

const (
	InitReward              = 50
	TxBroadcastQueueSize    = 10
	BlockBroadcastQueueSize = 10
	BlockQueueSize          = 10
	MaxTxSizePerBlock       = 10
)

type BitcoinServer struct {
	protocol.TransactionServer
	protocol.BlockServer
	state               *bitcoin.State
	cfg                 *config.Config
	nodeService         *service.NodeService
	txService           *service.TransactionService
	blockService        *service.BlockService
	txBroadcastQueue    chan *model.Transaction
	blockQueue          chan *model.Block
	blockBroadcastQueue chan *model.Block
	mineQueue           chan *model.Transaction
	cancelFunc          context.CancelCauseFunc
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
		state:               bitcoin.NewState(cfg.InitDifficultyLevel), //TODO: how to set when server restart?
		cancelFunc:          cancelFunc,
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
		if s.state.GetLastBlockNumber()+1 == block.Number {
			s.cancelFunc(errors.ErrServerCancelMining)
		}
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

func (s *BitcoinServer) UpdateState() {
	//TODO: how to handle when server is restart
	for {
		for i := uint64(0); i < s.cfg.BlocksPerDifficulty; i++ {
			block := <-s.blockQueue

			err := s.txService.ChainOnTxs(block.GetTxs()...)
			if err != nil {
				log.Printf("chain on transaction err: %v", err)
			}
			s.state.Update(block.Number, block.Time)
		}
	}
}

func (s *BitcoinServer) MineBlock(ctx context.Context) {
	//TODO: stop mining when receive the next block
	for {
		lastBlockNumber, reward, difficulty := s.state.Get(s.cfg.BlocksPerDifficulty, s.cfg.BlocksPerRewrad, s.cfg.BlockInterval)

		txs, err := s.receiveTxs(reward)
		if err != nil {
			//TODO: maybe fatal err?
			log.Printf("receive txs error: %v", err)
			continue
		}

		block, err := s.blockService.MineBlock(lastBlockNumber, difficulty, txs, ctx)
		if err != nil {
			//TODO: maybe fatal err?
			log.Printf("mine block error: %v", err)
			continue
		}

		s.blockQueue <- block
		s.blockBroadcastQueue <- block
	}
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
