package database

import (
	"Bitcoin/src/database"
	"Bitcoin/src/model"
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

	block, err := test.NewBlock(1, 10)
	if err != nil {
		t.Fatalf("new block error: %s", err)
	}

	blockdb := database.NewBlockDB(db)
	err = blockdb.SaveBlock(block)
	if err != nil {
		t.Fatalf("save block error: %s", err)
	}

	newBlock, err := blockdb.GetBlock(block.Id, block.Hash)
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

func Test_BlockDB_Last(t *testing.T) {
	db, err := leveldb.OpenFile(DBPath, nil)
	if err != nil {
		t.Fatalf("open %s error: %v", DBPath, err)
	}
	defer cleanUp(db, DBPath)

	blockdb := database.NewBlockDB(db)
	blocks := make([]*model.Block, 5)
	for i := 0; i < len(blocks); i++ {
		blocks[i], err = test.NewBlock(uint64(i+1), 10)
		if err != nil {
			t.Fatalf("new block error: %s", err)
		}

		err = blockdb.SaveBlock(blocks[i])
		if err != nil {
			t.Fatalf("save block error: %s", err)
		} else {
			t.Logf("save block %x %d", blocks[i].Hash, blocks[i].Id)
		}
	}

	lastBlocks, err := blockdb.LastBlocks(len(blocks))
	if err != nil {
		t.Fatalf("last blocks error: %v", err)
	}

	for i, block := range lastBlocks {
		if block == nil {
			t.Logf("the %d block is nil", i)
		} else {
			t.Logf("get block %x %d", block.Hash, block.Id)
		}
	}

	if len(lastBlocks) != len(blocks) {
		t.Fatalf("should get %d values, but actually %d", len(blocks), len(lastBlocks))
	}

	for i := 0; i < len(blocks); i++ {
		if !bytes.Equal(blocks[i].Hash, lastBlocks[i].Hash) {
			t.Fatalf("should get %x %d, but %x %d", blocks[i].Hash, blocks[i].Id, lastBlocks[i].Hash, lastBlocks[i].Id)
		}
	}
}
