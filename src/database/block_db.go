package database

import (
	"Bitcoin/src/model"
	"encoding/binary"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	BlockTable = "Block"
)

type IBlockDB interface {
	SaveBlock(block *model.Block) error
	GetBlock(id uint64, hash []byte) (*model.Block, error)
	LastBlocks(n int) ([]*model.Block, error)
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
	key := makeBlockKey(block.Id, block.Hash)
	return db.Save([]byte(BlockTable), key, block)
}

func (db *BlockDB) GetBlock(id uint64, hash []byte) (*model.Block, error) {
	key := makeBlockKey(id, hash)
	return db.Get([]byte(BlockTable), key)
}

func (db *BlockDB) LastBlocks(n int) ([]*model.Block, error) {
	return db.Last([]byte(BlockTable), n)
}

func makeBlockKey(id uint64, hash []byte) []byte {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, id)
	key = append(key, hash...)
	return key
}
