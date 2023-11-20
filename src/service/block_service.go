package service

import (
	"Bitcoin/src/bitcoin"
	"Bitcoin/src/config"
	"Bitcoin/src/database"
	"Bitcoin/src/errors"
	"Bitcoin/src/merkle"
	"Bitcoin/src/model"
	"bytes"
	"context"
	"log"
	"time"
)

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

func (serv *BlockService) MineBlock(lastBlockNumber uint64, difficulty float64, transactions []*model.Transaction, ctx context.Context) (*model.Block, error) {
	block, err := serv.MakeBlock(lastBlockNumber+1, difficulty, transactions, ctx)
	if err != nil {
		return nil, err
	}

	err = serv.SaveBlock(block)
	if err != nil {
		return nil, err
	}
	return block, nil
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

func (service *BlockService) MakeBlock(number uint64, difficulty float64, transactions []*model.Transaction, ctx context.Context) (*model.Block, error) {
	content, err := merkle.BuildTree(transactions)
	if err != nil {
		return nil, err
	}

	block := &model.Block{
		Number:     number,
		RootHash:   content.Table[len(content.Table)-1][0].Hash,
		Difficulty: difficulty,
		Time:       time.Now().UTC(),
		Body:       content,
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
	actual := bitcoin.ComputeDifficulty(hash)
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
