package model

import (
	"Bitcoin/src/bitcoin"
	"Bitcoin/src/cryptography"
	"Bitcoin/src/merkle"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"Bitcoin/test"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

type testBlock struct {
	Id         uint64    `json:"id,omitempty"`
	Hash       string    `json:"hash,omitempty"`
	Prevhash   string    `json:"prevhash,omitempty"`
	RootHash   string    `json:"roothash,omitempty"`
	Nonce      uint32    `json:"nonce,omitempty"`
	Difficulty string    `json:"difficulty,omitempty"`
	Timestamp  time.Time `json:"timestamp,omitempty"`
}

func (s *testBlock) equal(block *model.Block) (bool, string, string, string) {
	if block.Id != s.Id {
		return false, "id", fmt.Sprintf("%d", s.Id), fmt.Sprintf("%d", block.Id)
	}

	actualHash := hex.EncodeToString(block.Hash)
	if actualHash != s.Hash {
		return false, "hash", s.Hash, actualHash
	}

	actualPrevHash := hex.EncodeToString(block.Prevhash)
	if actualPrevHash != s.Prevhash {
		return false, "prevhash", s.Prevhash, actualPrevHash
	}

	actualRootHash := hex.EncodeToString(block.RootHash)
	if actualRootHash != s.RootHash {
		return false, "roothash", s.RootHash, actualRootHash
	}

	if s.Nonce != block.Nonce {
		return false, "nonce", fmt.Sprintf("%d", s.Nonce), fmt.Sprintf("%d", block.Nonce)
	}

	expectDifficulty, _ := strconv.ParseFloat(s.Difficulty, 64)
	if expectDifficulty != block.Difficulty {
		return false, "difficulty", fmt.Sprintf("%.0f", expectDifficulty), fmt.Sprintf("%.0f", block.Difficulty)
	}

	if s.Timestamp.UnixMilli() != block.Time.UnixMilli() {
		return false, "timestamp", fmt.Sprintf("%v", s.Timestamp), fmt.Sprintf("%v", block.Time)
	}

	return true, "", "", ""
}

func Test_Block_ComputeHash_Hash_Not_Change(t *testing.T) {
	block, err := test.NewBlock(1, 10)
	if err != nil {
		t.Fatalf("new block error: %v", err)
	}

	hash, err := block.ComputeHash()
	if err != nil {
		t.Fatalf("compute hash error: %v", err)
	}

	if bytes.Equal(hash, block.Hash) {
		t.Log("block hash didn't changed after serialize")
	} else {
		t.Fatalf("block hash is changed from [%x] to [%x]", block.Hash, hash)
	}
}

func Test_Block_Marshal(t *testing.T) {
	block, err := test.NewBlock(1, 10)
	if err != nil {
		t.Fatalf("new block error: %v", err)
	}

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	t.Logf("json: %s", string(data))

	var s testBlock
	err = json.Unmarshal(data, &s)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}

	same, field, expect, actual := s.equal(block)
	if !same {
		t.Fatalf("%s mismatch, expect: %s actual: %s", field, expect, actual)
	}
}

func Test_Block_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("block.json")
	if err != nil {
		t.Fatalf("read file error: %s", err)
	}

	var s testBlock
	err = json.Unmarshal(data, &s)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}

	var block model.Block
	err = json.Unmarshal(data, &block)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}

	same, field, expect, actual := s.equal(&block)
	if !same {
		t.Fatalf("%s mismatch, expect: %s actual: %s", field, expect, actual)
	}
}

func Test_Difficulty(t *testing.T) {
	hash, err := cryptography.Hash("hello")
	if err != nil {
		t.Fatalf("compute hash err: %v", err)
	}

	for z := uint64(1); z <= 5; z++ {
		difficulty := bitcoin.ComputeDifficulty(bitcoin.MakeDifficulty(z))
		log.Printf("difficulty: %x", difficulty)

		for k := z; k < z+5; k++ {
			for i := uint64(0); i < k; i++ {
				p := i / 8
				q := 7 - i%8
				hash[p] = (hash[p] | (1 << q)) ^ (1 << q)
			}
			t.Logf("hash: %x", hash)

			actual := bitcoin.ComputeDifficulty(hash)
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

	rootHash := block.Body.Table[len(block.Body.Table)-1][0].Hash
	if !bytes.Equal(rootHash, block.RootHash) {
		log.Fatalf("block root hash mismatch, expect: %x, actual: %x", block.RootHash, rootHash)
	}

	same, field, expect, actual := equal(req, block)
	if !same {
		t.Fatalf("%s mismatch, expect: %s actual: %s", field, expect, actual)
	}
}

func Test_Block_To(t *testing.T) {
	block, err := test.NewBlock(1, 10)
	if err != nil {
		t.Fatalf("new block err: %v", err)
	}

	req, err := model.BlockTo(block)
	if err != nil {
		t.Fatalf("block to err: %v", err)
	}

	var tree merkle.MerkleTree[*model.Transaction]
	err = json.Unmarshal(req.Content, &tree)
	if err != nil {
		log.Fatalf("unmarshal tree error: %v", err)
	}

	rootHash := tree.Table[len(tree.Table)-1][0].Hash
	if !bytes.Equal(rootHash, req.RootHash) {
		log.Fatalf("block req root hash mismatch, expect: %x, actual: %x", req.RootHash, rootHash)
	}

	same, field, expect, actual := equal(req, block)
	if !same {
		t.Fatalf("%s mismatch, expect: %s actual: %s", field, expect, actual)
	}
}

func Test_Block_FindHash(t *testing.T) {
	block, err := test.NewBlock(1, 10)
	if err != nil {
		t.Fatalf("new block err: %v", err)
	}

	hash, err := block.FindHash(context.TODO())
	if err != nil {
		t.Fatalf("find hash err: %v", err)
	}

	if block.Nonce == 0 {
		t.Fatalf("block %x nonce should not be 0", block.Hash)
	}

	difficulty := bitcoin.ComputeDifficulty(hash)
	if difficulty > block.Difficulty {
		t.Fatalf("the difficulty %.0f of block hash %x nonce > %.0f", difficulty, block.Hash, block.Difficulty)
	}
}

func newBlockReq() (*protocol.BlockReq, error) {
	prevHash, err := cryptography.Hash("prev")
	if err != nil {
		return nil, err
	}

	txs := make([]*model.Transaction, 4)
	for i := 0; i < len(txs); i++ {
		txs[i], err = test.NewTransaction()
		if err != nil {
			return nil, err
		}
	}

	tree, err := merkle.BuildTree(txs)
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

	block.Difficulty = bitcoin.ComputeDifficulty(hash)

	return block, nil
}

func equal(req *protocol.BlockReq, block *model.Block) (bool, string, string, string) {
	if block.Id != req.Id {
		return false, "id", fmt.Sprintf("%d", req.Id), fmt.Sprintf("%d", block.Id)
	}

	if !bytes.Equal(block.Hash, req.Hash) {
		return false, "hash", hex.EncodeToString(req.Hash), hex.EncodeToString(block.Hash)
	}

	if !bytes.Equal(block.Prevhash, req.Prevhash) {
		return false, "prevhash", hex.EncodeToString(req.Prevhash), hex.EncodeToString(block.Prevhash)
	}

	if !bytes.Equal(block.RootHash, req.RootHash) {
		return false, "rootHash", hex.EncodeToString(req.RootHash), hex.EncodeToString(block.RootHash)
	}

	if req.Nonce != block.Nonce {
		return false, "nonce", fmt.Sprintf("%d", req.Nonce), fmt.Sprintf("%d", block.Nonce)
	}

	if req.Difficulty != block.Difficulty {
		return false, "difficulty", fmt.Sprintf("%.0f", req.Difficulty), fmt.Sprintf("%.0f", block.Difficulty)
	}

	if req.Timestamp != block.Time.UnixMilli() {
		return false, "timestamp", fmt.Sprintf("%v", req.Timestamp), fmt.Sprintf("%v", block.Time)
	}

	return true, "", "", ""
}
