package database

import (
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
	err = txdb.SaveOffChainTx(tx)
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
