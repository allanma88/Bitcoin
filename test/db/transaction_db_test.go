package db

import (
	"Bitcoin/src/database"
	"Bitcoin/src/model"
	"bytes"
	"os"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	newTx, err := txdb.GetTx(tx.Id)
	if err != nil {
		t.Errorf("get transaction error: %s", err)
	}

	if !bytes.Equal(tx.Id, newTx.Id) {
		t.Errorf("transaction hash are not identical, expect: %x, actual: %x", tx.Id, newTx.Id)
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

	cleanUp(txdb)
}

func newTransaction(ins []*model.In, outs []*model.Out) (*model.Transaction, error) {
	tx := &model.Transaction{
		InLen:     uint32(len(ins)),
		OutLen:    uint32(len(outs)),
		Ins:       ins,
		Outs:      outs,
		Timestamp: timestamppb.Now(),
	}
	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, err
	}
	tx.Id = hash
	return tx, nil
}

func cleanUp(txdb database.ITransactionDB) {
	txdb.Close()
	os.RemoveAll(DBPath)
}
