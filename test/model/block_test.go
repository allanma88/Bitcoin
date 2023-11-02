package model

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/merkle"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

func Test_Block_ComputeHash_Hash_Not_Change(t *testing.T) {
	block, err := newBlock()
	if err != nil {
		t.Fatalf("new block error: %v", err)
	}

	hash, err := block.ComputeHash()
	if err != nil {
		t.Fatalf("compute hash error: %v", err)
	}

	if bytes.Equal(hash, block.Hash) {
		t.Log("transaction hash didn't changed after serialize")
	} else {
		t.Fatalf("transaction hash is changed from [%x] to [%x]", block.Hash, hash)
	}
}

func Test_Block_Marshal(t *testing.T) {
	block, err := newBlock()
	if err != nil {
		t.Fatalf("new block error: %v", err)
	}

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	t.Logf("json: %s", string(data))
}

func Test_Block_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("block.json")
	if err != nil {
		t.Fatalf("read file error: %s", err)
	}

	var block model.Block
	err = json.Unmarshal(data, &block)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}

	expectHash := "236309f506a54077d267ec48dc49c4bbafca8b9782d3e03b09fd24157dc2788b"
	actualHash := hex.EncodeToString(block.Hash)
	if actualHash != expectHash {
		t.Fatalf("expect hash is %s, actual is %s", expectHash, actualHash)
	}

	expectDifficulty, _ := strconv.ParseFloat("4097519688046410717622170652225431290464303836039125771621162712638162707939328", 64)
	if expectDifficulty != block.Difficulty {
		t.Fatalf("expect difficulty is %f, actual is %f", expectDifficulty, block.Difficulty)
	}
}

func Test_Difficulty(t *testing.T) {
	hash, err := cryptography.Hash("hello")
	if err != nil {
		t.Fatalf("compute hash err: %v", err)
	}

	for z := 1; z <= 5; z++ {
		difficulty := model.ComputeDifficulty(model.MakeDifficulty(z))
		log.Printf("difficulty: %x", difficulty)

		for k := z; k < z+5; k++ {
			for i := 0; i < k; i++ {
				p := i / 8
				q := 7 - i%8
				hash[p] = (hash[p] | (1 << q)) ^ (1 << q)
			}
			t.Logf("hash: %x", hash)

			actual := model.ComputeDifficulty(hash)
			if actual > difficulty {
				t.Fatalf("compute difficulty error, %v should smaller then %v, but actally greater", actual, difficulty)
			}
		}
	}
}

func Test_Block_From(t *testing.T) {
	req, err := newBlockReq()
	if err != nil {
		t.Fatalf("new block req err: %v", err)
	}

	block, err := model.BlockFrom(req)
	if err != nil {
		t.Fatalf("block from err: %v", err)
	}

	if !equal(req, block) {
		t.Fatalf("block req %x is not equal with block %x", req.Hash, block.Hash)
	}
}

func Test_Block_To(t *testing.T) {
	block, err := newBlock()
	if err != nil {
		t.Fatalf("new block err: %v", err)
	}

	req, err := model.BlockTo(block)
	if err != nil {
		t.Fatalf("block to err: %v", err)
	}

	if !equal(req, block) {
		t.Fatalf("block req %x is not equal with block %x", req.Hash, block.Hash)
	}
}

func Test_Block_FindHash(t *testing.T) {
	z := 10
	block, err := newBlock1(z)
	if err != nil {
		t.Fatalf("new block err: %v", err)
	}

	hash, err := block.FindHash()
	if err != nil {
		t.Fatalf("find hash err: %v", err)
	}

	if block.Nonce == 0 {
		t.Fatalf("block %x nonce should not be 0", block.Hash)
	}

	difficulty := model.ComputeDifficulty(hash)
	if difficulty > block.Difficulty {
		t.Fatalf("the difficulty %.0f of block hash %x nonce > %.0f", difficulty, block.Hash, block.Difficulty)
	}
}

func newBlock() (*model.Block, error) {
	prevHash, err := cryptography.Hash("prev")
	if err != nil {
		return nil, err
	}

	tree, err := merkle.BuildTree[string]([]string{"Hello1", "Hello2"})
	if err != nil {
		return nil, err
	}

	rootHash := tree.Table[len(tree.Table)-1][0].Hash

	block := &model.Block{
		Id:         1,
		Prevhash:   prevHash,
		RootHash:   rootHash,
		Difficulty: model.ComputeDifficulty(model.MakeDifficulty(10)),
		Time:       time.Now(),
		Body:       tree,
	}

	hash, err := block.FindHash()
	if err != nil {
		return nil, err
	}
	block.Hash = hash

	return block, nil
}

func newBlock1(z int) (*model.Block, error) {
	prevHash, err := cryptography.Hash("prev")
	if err != nil {
		return nil, err
	}

	tree, err := merkle.BuildTree[string]([]string{"Hello1", "Hello2"})
	if err != nil {
		return nil, err
	}

	rootHash := tree.Table[len(tree.Table)-1][0].Hash

	block := &model.Block{
		Id:         1,
		Prevhash:   prevHash,
		RootHash:   rootHash,
		Difficulty: model.ComputeDifficulty(model.MakeDifficulty(z)),
		Time:       time.Now(),
		Body:       tree,
	}

	return block, nil
}

func newBlockReq() (*protocol.BlockReq, error) {
	prevHash, err := cryptography.Hash("prev")
	if err != nil {
		return nil, err
	}

	tree, err := merkle.BuildTree[string]([]string{"Hello1", "Hello2"})
	if err != nil {
		return nil, err
	}

	rootHash := tree.Table[len(tree.Table)-1][0].Hash

	content, err := json.Marshal(tree)
	if err != nil {
		return nil, err
	}

	block := &protocol.BlockReq{
		Id:        1,
		Prevhash:  prevHash,
		RootHash:  rootHash,
		Nonce:     10,
		Timestamp: time.Now().UnixMilli(),
		Content:   content,
	}

	hash, err := cryptography.Hash(block)
	if err != nil {
		return nil, err
	}
	block.Hash = hash

	block.Difficulty = model.ComputeDifficulty(hash)

	return block, nil
}

func equal(req *protocol.BlockReq, block *model.Block) bool {
	if req.Id != block.Id {
		return false
	}

	if !bytes.Equal(req.Hash, block.Hash) {
		return false
	}

	if !bytes.Equal(req.Prevhash, block.Prevhash) {
		return false
	}

	if !bytes.Equal(req.RootHash, block.RootHash) {
		return false
	}

	if req.Nonce != block.Nonce {
		return false
	}

	if req.Difficulty != block.Difficulty {
		return false
	}

	if req.Timestamp != block.Time.UnixMilli() {
		return false
	}

	var tree merkle.MerkleTree
	err := json.Unmarshal(req.Content, &tree)
	if err != nil {
		log.Fatalf("unmarshal tree error: %v", err)
	}

	rootHash := tree.Table[len(tree.Table)-1][0].Hash
	rootHash1 := block.Body.Table[len(block.Body.Table)-1][0].Hash
	if !bytes.Equal(rootHash, rootHash1) {
		return false
	}

	return true
}
