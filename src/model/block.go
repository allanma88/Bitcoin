package model

import (
	"Bitcoin/src/bitcoin"
	"Bitcoin/src/cryptography"
	"Bitcoin/src/errors"
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

type Block struct {
	Number     uint64                           `json:"number,omitempty"`
	Hash       []byte                           `json:"hash,omitempty"`
	Prevhash   []byte                           `json:"prevhash,omitempty"`
	RootHash   []byte                           `json:"roothash,omitempty"`
	Nonce      uint32                           `json:"nonce,omitempty"`
	Difficulty float64                          `json:"difficulty,omitempty"`
	Time       time.Time                        `json:"timestamp,omitempty"`
	Body       *merkle.MerkleTree[*Transaction] `json:"-"`
}

type prettyBlock struct {
	Number     uint64    `json:"number,omitempty"`
	Hash       string    `json:"hash,omitempty"`
	Prevhash   string    `json:"prevhash,omitempty"`
	RootHash   string    `json:"roothash,omitempty"`
	Nonce      uint32    `json:"nonce,omitempty"`
	Difficulty string    `json:"difficulty,omitempty"`
	Timestamp  time.Time `json:"timestamp,omitempty"`
}

func (block *Block) MarshalJSON() ([]byte, error) {
	var s = prettyBlock{
		Number:     block.Number,
		Hash:       hex.EncodeToString(block.Hash),
		Prevhash:   hex.EncodeToString(block.Prevhash),
		RootHash:   hex.EncodeToString(block.RootHash),
		Nonce:      block.Nonce,
		Difficulty: fmt.Sprintf("%.0f", block.Difficulty),
		Timestamp:  block.Time,
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

		actual := bitcoin.ComputeDifficulty(hash)
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
