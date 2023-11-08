package database

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/database"
	"Bitcoin/test"
	"bytes"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

func Test_TransactionDB_Get(t *testing.T) {
	db, err := leveldb.OpenFile(DBPath, nil)
	if err != nil {
		t.Fatalf("open %s error: %v", DBPath, err)
	}

	defer cleanUp(db, DBPath)

	tx, err := test.NewTransaction()
	if err != nil {
		t.Fatalf("new transaction error: %v", err)
	}

	txdb := database.NewTransactionDB(db)
	err = txdb.SaveTx(tx)
	if err != nil {
		t.Fatalf("save transaction error: %v", err)
	}

	newTx, err := txdb.GetTx(tx.Hash)
	if err != nil {
		t.Fatalf("get transaction error: %v", err)
	}

	newHash, err := newTx.ComputeHash()
	if err != nil {
		t.Fatalf("compute hash error: %v", err)
	}
	if !bytes.Equal(newTx.Hash, newHash) {
		t.Fatalf("transaction is invalid, its hash is changed from %x to %x", newTx.Hash, newHash)
	}

	if !bytes.Equal(tx.Hash, newTx.Hash) {
		t.Fatalf("transaction hash are not identical, expect: %x, actual: %x", tx.Hash, newTx.Hash)
	}
}

func Test_TransactionDB_Remove(t *testing.T) {
	db, err := leveldb.OpenFile(DBPath, nil)
	if err != nil {
		t.Fatalf("open %s error: %v", DBPath, err)
	}

	defer cleanUp(db, DBPath)

	tx, err := test.NewTransaction()
	if err != nil {
		t.Fatalf("new transaction error: %v", err)
	}

	txdb := database.NewTransactionDB(db)
	err = txdb.SaveTx(tx)
	if err != nil {
		t.Fatalf("save transaction error: %s", err)
	}

	err = txdb.RemoveTx(tx.Hash)
	if err != nil {
		t.Fatalf("remove tx error: %sv", err)
	}

	newTx, err := txdb.GetTx(tx.Hash)
	if err != nil {
		t.Fatalf("get tx error: %v", err)
	}
	if newTx != nil {
		t.Fatalf("get an deleted tx %x", tx.Hash)
	}
}

func Test_TransactionDB_Remove_Not_Exist_Tx(t *testing.T) {
	db, err := leveldb.OpenFile(DBPath, nil)
	if err != nil {
		t.Fatalf("open %s error: %v", DBPath, err)
	}

	defer cleanUp(db, DBPath)

	txdb := database.NewTransactionDB(db)
	hash, err := cryptography.Hash("whatever")
	if err != nil {
		t.Fatalf("compute hash error: %v", err)
	}

	err = txdb.RemoveTx(hash)
	if err != nil {
		t.Fatalf("remove tx error: %v", err)
	}
}
