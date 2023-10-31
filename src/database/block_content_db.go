package database

import (
	"Bitcoin/src/merkle"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	BlockContentTable = "BlockContent"
)

type IBlockContentDB interface {
	SaveBlockContent(key []byte, content *merkle.MerkleTree) error
	GetBlockContent(hash []byte) (*merkle.MerkleTree, error)
	Close() error
}

type BlockContentDB struct {
	IBaseDB[merkle.MerkleTree]
}

func NewBlockContentDB(db *leveldb.DB) IBlockContentDB {
	basedb := &BaseDB[merkle.MerkleTree]{Database: db}
	blockdb := &BlockContentDB{IBaseDB: basedb}
	return blockdb
}

func (db *BlockContentDB) SaveBlockContent(key []byte, content *merkle.MerkleTree) error {
	return db.Save([]byte(BlockContentTable), key, content)
}

func (db *BlockContentDB) GetBlockContent(key []byte) (*merkle.MerkleTree, error) {
	return db.Get([]byte(BlockContentTable), key)
}
