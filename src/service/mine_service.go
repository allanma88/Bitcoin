package service

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/config"
	"Bitcoin/src/model"
	"context"
	"sync"
	"time"
)

type MineService struct {
	cfg       *config.Config
	txService *TransactionService
	mempool   *MemPool
}

func NewMineService(cfg *config.Config, txService *TransactionService, mempool *MemPool) *MineService {
	return &MineService{
		cfg:       cfg,
		txService: txService,
		mempool:   mempool,
	}
}

func (s *MineService) MineBlock(lastBlock *model.Block, ctx context.Context, wait *sync.WaitGroup) (*model.Block, error) {
	reward := lastBlock.GetNextReward(s.cfg.InitRewrad, s.cfg.BlocksPerRewrad)

	txs, err := s.fetchTxs(reward)
	if err != nil {
		return nil, err
	}

	wait.Wait()

	return s.mineBlock(lastBlock, txs, ctx)
}

func (s *MineService) mineBlock(lastBlock *model.Block, transactions []*model.Transaction, ctx context.Context) (*model.Block, error) {
	content, err := collection.BuildTree(transactions)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	difficulty := lastBlock.GetNextDifficulty(s.cfg.BlocksPerDifficulty, s.cfg.BlockInterval)
	totalInterval := lastBlock.GetNextTotalInterval(now, s.cfg.BlocksPerDifficulty)

	block := &model.Block{
		Number:        lastBlock.Number + 1,
		Prevhash:      lastBlock.Hash,
		RootHash:      content.Table[len(content.Table)-1][0].Hash,
		Difficulty:    difficulty,
		Time:          now,
		TotalInterval: totalInterval,
		Miner:         s.cfg.Server,
		Body:          content,
	}

	hash, err := block.FindHash(ctx)
	if err != nil {
		return nil, err
	}
	block.Hash = hash

	for _, tx := range transactions {
		tx.BlockHash = hash
	}

	return block, nil
}

func (s *MineService) fetchTxs(reward uint64) ([]*model.Transaction, error) {
	txs := s.mempool.TopMax(int(s.cfg.MaxTxSizePerBlock))

	totalFee, err := s.txService.ValidateTxs(txs, nil)
	if err != nil {
		return nil, err
	}

	coinbaseTx, err := model.MakeCoinbaseTx(s.cfg.MinerPubkey, reward+totalFee)
	if err != nil {
		return nil, err
	}

	txs = append([]*model.Transaction{coinbaseTx}, txs...)
	return txs, nil
}
