package model

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/errors"
	"Bitcoin/src/infra"
	"Bitcoin/src/merkle"
	"Bitcoin/src/protocol"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/peteprogrammer/go-automapper"
)

const (
	InitReward = 50 //TODO: make configurable
)

type Block struct {
	Number        uint64
	Hash          []byte
	Prevhash      []byte
	RootHash      []byte
	Nonce         uint32
	Difficulty    float64
	Time          time.Time
	Body          *merkle.MerkleTree[*Transaction]
	TotalInterval uint64
	Miner         string
}

type prettyBlock struct {
	Number        uint64    `json:"number,omitempty"`
	Hash          string    `json:"hash,omitempty"`
	Prevhash      string    `json:"prevhash,omitempty"`
	RootHash      string    `json:"roothash,omitempty"`
	Nonce         uint32    `json:"nonce,omitempty"`
	Difficulty    string    `json:"difficulty,omitempty"`
	Timestamp     time.Time `json:"timestamp,omitempty"`
	TotalInterval uint64
	Miner         string
}

func (block *Block) MarshalJSON() ([]byte, error) {
	var s = prettyBlock{
		Number:        block.Number,
		Hash:          hex.EncodeToString(block.Hash),
		Prevhash:      hex.EncodeToString(block.Prevhash),
		RootHash:      hex.EncodeToString(block.RootHash),
		Nonce:         block.Nonce,
		Difficulty:    fmt.Sprintf("%.0f", block.Difficulty),
		Timestamp:     block.Time,
		TotalInterval: block.TotalInterval,
		Miner:         block.Miner,
	}
	return json.Marshal(s)
}

func (block *Block) UnmarshalJSON(data []byte) error {
	var s prettyBlock
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	block.Number = s.Number

	block.Hash, err = hex.DecodeString(s.Hash)
	if err != nil {
		return err
	}

	block.Prevhash, err = hex.DecodeString(s.Prevhash)
	if err != nil {
		return err
	}

	block.RootHash, err = hex.DecodeString(s.RootHash)
	if err != nil {
		return err
	}

	difficulty, err := strconv.ParseFloat(s.Difficulty, 64)
	if err != nil {
		return err
	}

	block.Difficulty = difficulty
	block.Nonce = s.Nonce
	block.Time = s.Timestamp
	block.TotalInterval = s.TotalInterval
	block.Miner = s.Miner
	return err
}

func BlockFrom(request *protocol.BlockReq) (*Block, error) {
	var block Block
	var tree merkle.MerkleTree[*Transaction]

	err := json.Unmarshal(request.Content, &tree)
	if err != nil {
		return nil, err
	}
	block.Body = &tree

	automapper.MapLoose(request, &block)
	block.Time = time.UnixMilli(request.Timestamp)
	return &block, nil
}

func BlockTo(block *Block) (*protocol.BlockReq, error) {
	var request protocol.BlockReq
	automapper.MapLoose(block, &request)
	request.Timestamp = block.Time.UnixMilli()

	body, err := json.Marshal(block.Body)
	if err != nil {
		return nil, err
	}

	request.Content = body
	return &request, nil
}

func (block *Block) FindHash(ctx context.Context) ([]byte, error) {
	var nonce uint32
	//return err if not find valid hash
	for nonce = 1; nonce < math.MaxUint32; nonce++ {
		err := context.Cause(ctx)
		if err != nil {
			return nil, err
		}

		block.Nonce = nonce

		hash, err := block.ComputeHash()
		if err != nil {
			return nil, err
		}

		actual := infra.ComputeDifficulty(hash)
		if actual <= block.Difficulty {
			return hash, nil
		}
	}

	block.Nonce = 0
	return nil, errors.ErrBlockContentInvalid
}

func (block *Block) ComputeHash() ([]byte, error) {
	originalHash := block.Hash
	block.Hash = []byte{}

	hash, err := cryptography.Hash(block)

	block.Hash = originalHash
	return hash, err
}

func (block *Block) GetTxs() []*Transaction {
	return block.Body.GetVals()
}

func (block *Block) GetNextDifficulty(blocksPerDifficulty, expectBlockInterval uint64) float64 {
	difficulty := block.Difficulty
	if block.Number%blocksPerDifficulty == 0 {
		avgInterval := block.TotalInterval / (blocksPerDifficulty)
		difficulty = block.Difficulty * float64((avgInterval / expectBlockInterval))
	}

	return difficulty
}

func (block *Block) GetNextReward(blocksPerRewrad uint64) uint64 {
	return InitReward / (block.Number/blocksPerRewrad + 1)
}

func (block *Block) GetNextTotalInterval(t time.Time, blocksPerDifficulty uint64) uint64 {
	if block.Number%blocksPerDifficulty == 0 {
		return 0
	} else {
		return block.TotalInterval + uint64(t.Sub(block.Time).Milliseconds())
	}
}
