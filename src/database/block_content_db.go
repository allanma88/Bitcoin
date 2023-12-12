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
	txDB ITransactionDB
}

func NewBlockContentDB(db *leveldb.DB) IBlockContentDB {
	basedb := &BaseDB[merkle.MerkleTree[*model.Transaction]]{Database: db}
	blockContentDB := &BlockContentDB{IBaseDB: basedb, txDB: NewTransactionDB(db)}
	return blockContentDB
}

func (db *BlockContentDB) SaveBlockContent(key []byte, content *merkle.MerkleTree[*model.Transaction]) error {
	err := db.Save([]byte(BlockContentTable), key, content)
	if err != nil {
		return err
	}

	//TODO: save an index of tx, not the entire tx, so not save duplicate with tx in the block
	//TODO: save txs and block content in one db transaction
	for _, tx := range content.GetVals() {
		err = db.txDB.SaveOnChainTx(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *BlockContentDB) GetBlockContent(key []byte) (*merkle.MerkleTree[*model.Transaction], error) {
	return db.Get([]byte(BlockContentTable), key)
}
