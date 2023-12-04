package service

import (
	"Bitcoin/src/config"
	"Bitcoin/src/database"
	"Bitcoin/src/errors"
	"Bitcoin/src/infra"
	"Bitcoin/src/merkle"
	"Bitcoin/src/model"
	"bytes"
	"context"
	"log"
	"time"
)

//TODO: more test cases

type BlockService struct {
	blockDB        database.IBlockDB
	blockContentDB database.IBlockContentDB
	cfg            *config.Config
}

func NewBlockService(blockDB database.IBlockDB, blockContentDB database.IBlockContentDB, cfg *config.Config) *BlockService {
	return &BlockService{
		blockDB:        blockDB,
		blockContentDB: blockContentDB,
		cfg:            cfg,
	}
}

func (service *BlockService) GetBlocks() []*model.Block {
	return nil
}

func (service *BlockService) SaveBlock(block *model.Block) error {
	err := service.blockDB.SaveBlock(block)
	if err != nil {
		return err
	}

	return service.blockContentDB.SaveBlockContent(block.RootHash, block.Body)
}

func (service *BlockService) Validate(block *model.Block) error {
	hash, err := validateHash[*model.Block](block.Hash, block)
	if err != nil {
		return err
	}

	err = validateTimestamp(block.Time)
	if err != nil {
		return err
	}

	existBlock, err := service.blockDB.GetBlock(hash)
	if err != nil {
		return err
	}
	if existBlock != nil {
		return errors.ErrBlockExist
	}

	prevBlock, err := service.blockDB.GetBlock(block.Prevhash)
	if err != nil {
		return err
	}
	if prevBlock == nil {
		return errors.ErrPrevBlockNotFound
	}
	if block.Number != prevBlock.Number+1 {
		return errors.ErrBlockNumberInvalid
	}
	block.PrevBlock = prevBlock

	err = validateDifficulty(block.Hash, block.Difficulty)
	if err != nil {
		return err
	}

	err = validateRootHash(block.RootHash, block.Body)
	if err != nil {
		return err
	}

	return nil
}

func (service *BlockService) MineBlock(lastBlock *model.Block, transactions []*model.Transaction, ctx context.Context) (*model.Block, error) {
	content, err := merkle.BuildTree(transactions)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	difficulty := lastBlock.GetNextDifficulty(service.cfg.BlocksPerDifficulty, service.cfg.BlockInterval)
	totalInterval := lastBlock.GetNextTotalInterval(now, service.cfg.BlocksPerDifficulty)

	block := &model.Block{
		Number:        lastBlock.Number + 1,
		Prevhash:      lastBlock.Hash,
		RootHash:      content.Table[len(content.Table)-1][0].Hash,
		Difficulty:    difficulty,
		Time:          now,
		TotalInterval: totalInterval,
		Miner:         service.cfg.Server,
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

func validateDifficulty(hash []byte, difficulty float64) error {
	actual := infra.ComputeDifficulty(hash)
	if actual > difficulty {
		return errors.ErrBlockNonceInvalid
	}
	return nil
}

func validateRootHash(roothash []byte, tree *merkle.MerkleTree[*model.Transaction]) error {
	valid, err := tree.Validate()
	if err != nil {
		return err
	}

	if !valid {
		log.Printf("content is invalid")
		return errors.ErrBlockContentInvalid
	}

	if !bytes.Equal(roothash, tree.Table[len(tree.Table)-1][0].Hash) {
		log.Printf("content hash mismatch with root hash")
		return errors.ErrBlockContentInvalid
	}

	return nil
}
