package service

import (
	"Bitcoin/src/config"
	"Bitcoin/src/database"
	"Bitcoin/src/merkle"
	"Bitcoin/src/model"
	"Bitcoin/src/service"
	"fmt"
	"testing"
	"time"
)

func Test_Validate_Succeed(t *testing.T) {
	block, err := makeBlock(20)
	if err != nil {
		t.Fatalf("make block error: %v", err)
	}

	blockdb := newBlockDB()
	blockContentDB := newBlockContentDB()
	serv := service.NewBlockService(blockdb, blockContentDB, &config.Config{})

	err = serv.Validate(block)
	if err != nil {
		t.Fatalf("validate block failed: %v", err)
	}
	t.Logf("Block %x validate succeed", block.Hash)
}

func makeBlock(difficultyLevel int) (*model.Block, error) {
	vals := make([]string, 5)
	for i := 0; i < len(vals); i++ {
		vals[i] = fmt.Sprintf("Hello%d", (i + 1))
	}

	content, err := merkle.BuildTree(vals)
	if err != nil {
		return nil, err
	}

	block := &model.Block{
		RootHash:   content.Table[len(content.Table)-1][0].Hash,
		Difficulty: model.ComputeDifficulty(model.MakeDifficulty(difficultyLevel)),
		Time:       time.Now().UTC(),
		Body:       content,
	}

	hash, err := block.FindHash()
	if err != nil {
		return nil, err
	}
	block.Hash = hash

	return block, nil
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
	basedb := newTestBaseDB[merkle.MerkleTree]()
	blockContentDB := &database.BlockContentDB{IBaseDB: basedb}
	return blockContentDB
}
