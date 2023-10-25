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
	IBaseDB[*model.Transaction]
}

func NewTransactionDB(db *leveldb.DB) *TransactionDB {
	basedb := &BaseDB[*model.Transaction]{Database: db}
	txdb := &TransactionDB{IBaseDB: basedb}
	return txdb
}

func (db *TransactionDB) SaveTx(tx *model.Transaction) error {
	return db.Save([]byte(TxTable), tx.Id, tx)
}

func (db *TransactionDB) GetTx(hash []byte) (*model.Transaction, error) {
	var tx model.Transaction
	err := db.Get([]byte(TxTable), hash, &tx)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tx, nil
}
