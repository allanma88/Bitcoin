package database

import (
	"Bitcoin/src/model"

	"github.com/syndtr/goleveldb/leveldb"
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
	IBaseDB[model.Block]
}

func NewBlockDB(db *leveldb.DB) IBlockDB {
	basedb := &BaseDB[model.Block]{Database: db}
	blockdb := &BlockDB{IBaseDB: basedb}
	return blockdb
}

func (db *BlockDB) SaveBlock(block *model.Block) error {
	return db.Save([]byte(BlockTable), block.Hash, block)
}

func (db *BlockDB) GetBlock(hash []byte) (*model.Block, error) {
	return db.Get([]byte(BlockTable), hash)
}
