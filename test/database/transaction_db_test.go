package database

import (
	"Bitcoin/src/database"
	"Bitcoin/src/model"
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	DBPath = "bitcoin"
)

func Test_CRUD(t *testing.T) {
	ins := []*model.In{}
	outs := []*model.Out{}
	tx, err := newTransaction(ins, outs)
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
	}

	db, err := leveldb.OpenFile(DBPath, nil)
	if err != nil {
		t.Fatalf("open %s error: %s", DBPath, err)
	}

	txdb := database.NewTransactionDB(db)
	err = txdb.SaveTx(tx)
	if err != nil {
		t.Errorf("save transaction error: %s", err)
	}

	newTx, err := txdb.GetTx(tx.Hash)
	if err != nil {
		t.Errorf("get transaction error: %s", err)
	}

	if !bytes.Equal(tx.Hash, newTx.Hash) {
		t.Errorf("transaction hash are not identical, expect: %x, actual: %x", tx.Hash, newTx.Hash)
	}

	expectHash, err := tx.ComputeHash()
	if err != nil {
		t.Errorf("compute hash error: %s", err)
	}
	actualHash, err := newTx.ComputeHash()
	if err != nil {
		t.Errorf("compute hash error: %s", err)
	}
	if !bytes.Equal(expectHash, actualHash) {
		t.Errorf("transaction are not identical, expect: %x, actual: %x", expectHash, actualHash)
	}

	err = txdb.RemoveTx(tx)
	if err != nil {
		t.Errorf("remove tx error: %s", err)
	}

	newTx, err = txdb.GetTx(tx.Hash)
	if err != nil {
		t.Errorf("get tx error: %s", err)
	}
	if newTx != nil {
		t.Errorf("get an deleted tx %x", tx.Hash)
	}

	txdb.Close()
	os.RemoveAll(DBPath)
}

func newTransaction(ins []*model.In, outs []*model.Out) (*model.Transaction, error) {
	tx := &model.Transaction{
		InLen:     uint32(len(ins)),
		OutLen:    uint32(len(outs)),
		Ins:       ins,
		Outs:      outs,
		Timestamp: time.Now(),
	}
	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, err
	}
	tx.Hash = hash
	return tx, nil
}
