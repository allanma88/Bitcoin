package database

import (
	"Bitcoin/src/model"
)

const (
	BlockTable = "Block"
)

type IBlockDB interface {
	SaveBlock(block *model.Block) error
	GetBlock(hash []byte) (*model.Block, error)
	Close() error
}

type BlockDB struct {
	BaseDB[*model.Block]
}

func (db *BlockDB) SaveBlock(block *model.Block) error {
	return db.Save([]byte(BlockTable), block.Id, block)
}

func (db *BlockDB) GetBlock(hash []byte) (*model.Block, error) {
	var block model.Block
	err := db.Get([]byte(BlockTable), hash, &block)
	return &block, err
}
