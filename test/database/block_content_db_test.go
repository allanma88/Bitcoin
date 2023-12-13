package database

import (
	"Bitcoin/src/database"
	"Bitcoin/test"
	"bytes"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

func Test_BlockContentDB_Get(t *testing.T) {
	db, err := leveldb.OpenFile(DBPath, nil)
	if err != nil {
		t.Fatalf("open %s error: %v", DBPath, err)
	}

	defer cleanUp(db, DBPath)

	tx := test.NewTransaction([]byte{})
	blockContentDb := database.NewBlockContentDB(db)
	err = blockContentDb.SaveTx(tx)
	if err != nil {
		t.Fatalf("save transaction error: %v", err)
	}

	newTx, err := blockContentDb.GetTx(tx.Hash)
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
