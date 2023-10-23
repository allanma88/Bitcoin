package database

import (
	"Bitcoin/src/model"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	TxTable = "Transaction"
)

type ITransactionDB interface {
	SaveTx(tx *model.Transaction) error
	GetTx(hash []byte) (*model.Transaction, error)
	Close() error
}

type TransactionDB struct {
	BaseDB[*model.Transaction]
}

func NewTransactionDB(db *leveldb.DB) *TransactionDB {
	basedb := BaseDB[*model.Transaction]{Database: db}
	txdb := &TransactionDB{BaseDB: basedb}
	return txdb
}

func (db *TransactionDB) SaveTx(tx *model.Transaction) error {
	return db.BaseDB.Save([]byte(TxTable), tx.Id, tx)
}

func (db *TransactionDB) GetTx(hash []byte) (*model.Transaction, error) {
	var tx model.Transaction
	err := db.BaseDB.Get([]byte(TxTable), hash, &tx)
	return &tx, err
}
