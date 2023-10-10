package db

import (
	"Bitcoin/src/model"
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type ITransactionDB interface {
	SaveTx(tx *model.Transaction) error
	GetTx(hash []byte) (*model.Transaction, error)
	Close() error
}

type TransactionDB struct {
	*leveldb.DB
}

func NewTransactionDB(path string) (ITransactionDB, error) {
	db, err := leveldb.OpenFile(path, nil)
	return &TransactionDB{DB: db}, err
}

func (db *TransactionDB) SaveTx(tx *model.Transaction) error {
	bytes, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	opt := &opt.WriteOptions{}
	err = db.DB.Put(tx.Id, bytes, opt)
	return err
}

func (db *TransactionDB) GetTx(hash []byte) (*model.Transaction, error) {
	opt := &opt.ReadOptions{}
	bytes, err := db.DB.Get(hash, opt)
	if err != nil {
		return nil, err
	}

	var tx model.Transaction
	err = json.Unmarshal(bytes, &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (db *TransactionDB) Close() error {
	return db.DB.Close()
}
