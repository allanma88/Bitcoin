package database

import (
	"Bitcoin/src/model"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	OnChainTxTable  = "OnChainTransaction"
	OffChainTxTable = "OffChainTransaction"
)

type ITransactionDB interface {
	SaveOffChainTx(tx *model.Transaction) error
	SaveOnChainTx(tx *model.Transaction) error
	GetTx(hash []byte) (*model.Transaction, error)
	GetOnChainTx(hash []byte) (*model.Transaction, error)
	Close() error
}

type TransactionDB struct {
	IBaseDB[model.Transaction]
}

func NewTransactionDB(db *leveldb.DB) *TransactionDB {
	basedb := &BaseDB[model.Transaction]{Database: db}
	txdb := &TransactionDB{IBaseDB: basedb}
	return txdb
}

func (db *TransactionDB) SaveOffChainTx(tx *model.Transaction) error {
	return db.Save([]byte(OffChainTxTable), tx.Hash, tx)
}

func (db *TransactionDB) SaveOnChainTx(tx *model.Transaction) error {
	return db.Move([]byte(OffChainTxTable), []byte(OnChainTxTable), tx.Hash, tx)
}

func (db *TransactionDB) GetTx(hash []byte) (*model.Transaction, error) {
	tx, err := db.Get([]byte(OnChainTxTable), hash)
	if tx != nil || err != nil {
		return tx, err
	}
	tx, err = db.Get([]byte(OffChainTxTable), hash)
	if tx != nil || err != nil {
		return tx, err
	}
	return nil, nil
}

// TODO: add test case
func (db *TransactionDB) GetOnChainTx(hash []byte) (*model.Transaction, error) {
	tx, err := db.Get([]byte(OnChainTxTable), hash)
	if err != nil {
		return nil, err
	}

	if tx != nil {
		if tx.BlockHash == nil || len(tx.BlockHash) == 0 {
			return nil, nil
		}
	}
	return tx, nil
}
