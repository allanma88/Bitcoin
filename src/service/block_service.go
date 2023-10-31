package service

import (
	"Bitcoin/src/config"
	"Bitcoin/src/database"
	"Bitcoin/src/errors"
	"Bitcoin/src/merkle"
	"Bitcoin/src/model"
	"bytes"
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

func (serv *BlockService) MineBlock(transactions []*model.Transaction) (*model.Block, error) {
	block, err := serv.MakeBlock(transactions)
	if err != nil {
		return nil, err
	}

	err = serv.SaveBlock(block)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (service *BlockService) MakeBlock(transactions []*model.Transaction) (*model.Block, error) {
	content, err := merkle.BuildTree(transactions)
	if err != nil {
		return nil, err
	}

	difficulty, err := service.FindDifficulty()
	if err != nil {
		return nil, err
	}

	block := &model.Block{
		RootHash:   content.Table[len(content.Table)-1][0].Hash,
		Difficulty: difficulty,
		Timestamp:  time.Now().UTC(),
		Body:       content,
	}

	hash, err := block.FindHash()
	if err != nil {
		return nil, err
	}
	block.Hash = hash

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

	err = validateTimestamp(block.Timestamp)
	if err != nil {
		return err
	}

	if block.Prevhash != nil {
		prev, err := service.blockDB.GetBlock(block.Prevhash)
		if err != nil {
			return err
		}
		if prev == nil {
			return errors.ErrBlockNotFound
		}
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

func (service *BlockService) FindDifficulty() (float64, error) {
	lastBlocks, err := service.blockDB.LastBlocks(int(service.cfg.AjustBlockNum))
	if err != nil {
		return 0, err
	}

	if len(lastBlocks) == 0 {
		return model.ComputeDifficulty(model.MakeDifficulty(int(service.cfg.InitDifficulty))), nil
	}

	lastBlock := lastBlocks[len(lastBlocks)-1]

	if lastBlock.Id%uint64(service.cfg.AjustBlockNum) != 0 {
		return lastBlock.Difficulty, nil
	}

	actualDuration, err := service.ComputeAvgBlockDuration(lastBlocks)
	if err != nil {
		return 0, err
	}

	difficulty := lastBlock.Difficulty * float64((actualDuration / service.cfg.BlockDuration).Microseconds())
	return difficulty, nil
}

func (service *BlockService) ComputeAvgBlockDuration(lastBlocks []*model.Block) (time.Duration, error) {

	var actualDuration time.Duration
	for i := 1; i < len(lastBlocks); i++ {
		actualDuration += lastBlocks[i].Timestamp.Sub(lastBlocks[i-1].Timestamp)
	}
	actualDuration = time.Duration(int64(actualDuration) / int64(len(lastBlocks)))
	return actualDuration, nil
}

func validateDifficulty(hash []byte, difficulty float64) error {
	actual := model.ComputeDifficulty(hash)
	if actual > difficulty {
		return errors.ErrBlockNonceInvalid
	}
	return nil
}

func validateRootHash(roothash []byte, tree *merkle.MerkleTree) error {
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
