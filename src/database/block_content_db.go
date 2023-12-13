package database

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/model"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	BlockContentTable = "BlockContent"
	TxTable           = "Transaction"
)

type IBlockContentDB interface {
	SaveBlockContent(key []byte, content *collection.MerkleTree[*model.Transaction]) error
	GetBlockContent(hash []byte) (*collection.MerkleTree[*model.Transaction], error)
	SaveTx(tx *model.Transaction) error
	GetTx(hash []byte) (*model.Transaction, error)
	Close() error
}

type BlockContentDB struct {
	IBaseDB[collection.MerkleTree[*model.Transaction]]
	TxDB IBaseDB[model.Transaction]
}

func NewBlockContentDB(db *leveldb.DB) IBlockContentDB {
	basedb := &BaseDB[collection.MerkleTree[*model.Transaction]]{Database: db}
	blockContentDB := &BlockContentDB{IBaseDB: basedb, TxDB: &BaseDB[model.Transaction]{Database: db}}
	return blockContentDB
}

func (db *BlockContentDB) SaveBlockContent(key []byte, content *collection.MerkleTree[*model.Transaction]) error {
	err := db.Save([]byte(BlockContentTable), key, content)
	if err != nil {
		return err
	}

	//TODO: save an index of tx, not the entire tx, so not save duplicate with tx in the block
	//TODO: save txs and block content in one db transaction
	for _, tx := range content.GetVals() {
		err = db.SaveTx(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *BlockContentDB) GetBlockContent(key []byte) (*collection.MerkleTree[*model.Transaction], error) {
	content, err := db.Get([]byte(BlockContentTable), key)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(content.Table[0]); i++ {
		txhash := content.Table[0][i].Hash

		tx, err := db.GetTx(txhash)
		if err != nil {
			return nil, err
		}
		content.Table[0][i].Val = tx
	}
	return content, nil
}

func (db *BlockContentDB) SaveTx(tx *model.Transaction) error {
	return db.TxDB.Save([]byte(TxTable), tx.Hash, tx)
}

func (db *BlockContentDB) GetTx(hash []byte) (*model.Transaction, error) {
	return db.TxDB.Get([]byte(TxTable), hash)
}
