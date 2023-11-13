package database

import (
	"Bitcoin/src/merkle"
	"Bitcoin/src/model"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	BlockContentTable = "BlockContent"
)

type IBlockContentDB interface {
	SaveBlockContent(key []byte, content *merkle.MerkleTree[*model.Transaction]) error
	GetBlockContent(hash []byte) (*merkle.MerkleTree[*model.Transaction], error)
	Close() error
}

type BlockContentDB struct {
	IBaseDB[merkle.MerkleTree[*model.Transaction]]
}

func NewBlockContentDB(db *leveldb.DB) IBlockContentDB {
	basedb := &BaseDB[merkle.MerkleTree[*model.Transaction]]{Database: db}
	blockContentDB := &BlockContentDB{IBaseDB: basedb}
	return blockContentDB
}

func (db *BlockContentDB) SaveBlockContent(key []byte, content *merkle.MerkleTree[*model.Transaction]) error {
	return db.Save([]byte(BlockContentTable), key, content)
}

func (db *BlockContentDB) GetBlockContent(key []byte) (*merkle.MerkleTree[*model.Transaction], error) {
	return db.Get([]byte(BlockContentTable), key)
}
