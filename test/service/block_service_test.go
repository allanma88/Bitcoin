package service

import (
	"Bitcoin/src/config"
	"Bitcoin/src/database"
	"Bitcoin/src/merkle"
	"Bitcoin/src/model"
	"Bitcoin/src/service"
	"Bitcoin/test"
	"testing"
)

func Test_Validate_Succeed(t *testing.T) {
	prevBlock := test.NewBlock(1, 10, nil)
	block := test.NewBlock(2, 10, prevBlock.Hash)
	blockdb := newBlockDB()
	blockContentDB := newBlockContentDB()
	serv := service.NewBlockService(blockdb, blockContentDB, &config.Config{})

	blockdb.SaveBlock(prevBlock)

	err := serv.Validate(block)
	if err != nil {
		t.Fatalf("validate block failed: %v", err)
	}
	t.Logf("Block %x validate succeed", block.Hash)
}

func newBlockDB(blocks ...*model.Block) database.IBlockDB {
	basedb := newTestBaseDB[model.Block]()
	blockdb := &database.BlockDB{IBaseDB: basedb}
	for _, block := range blocks {
		blockdb.SaveBlock(block)
	}
	return blockdb
}

func newBlockContentDB() database.IBlockContentDB {
	basedb := newTestBaseDB[merkle.MerkleTree[*model.Transaction]]()
	blockContentDB := &database.BlockContentDB{IBaseDB: basedb}
	return blockContentDB
}
