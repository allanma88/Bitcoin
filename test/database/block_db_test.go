package database

import (
	"Bitcoin/src/database"
	"Bitcoin/test"
	"bytes"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

func Test_BlockDB_Get(t *testing.T) {
	db, err := leveldb.OpenFile(DBPath, nil)
	if err != nil {
		t.Fatalf("open %s error: %s", DBPath, err)
	}

	defer cleanUp(db, DBPath)

	block := test.NewBlock(1, 10, nil)

	blockdb := database.NewBlockDB(db)
	err = blockdb.SaveBlock(block)
	if err != nil {
		t.Fatalf("save block error: %s", err)
	}

	newBlock, err := blockdb.GetBlock(block.Hash)
	if err != nil {
		t.Fatalf("get block error: %s", err)
	}
	t.Logf("get new block %x", newBlock.Hash)

	newHash, err := newBlock.ComputeHash()
	if err != nil {
		t.Fatalf("compute hash error: %s", err)
	}
	if !bytes.Equal(newBlock.Hash, newHash) {
		t.Fatalf("block is invalid, its hash is changed from %x to %x", newBlock.Hash, newHash)
	}

	if !bytes.Equal(block.Hash, newBlock.Hash) {
		t.Errorf("block hash are not identical, expect: %x, actual: %x", block.Hash, newBlock.Hash)
	}
}
