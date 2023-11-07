package model

import (
	"Bitcoin/src/bitcoin"
	"Bitcoin/src/cryptography"
	"Bitcoin/src/errors"
	"Bitcoin/src/merkle"
	"Bitcoin/src/protocol"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/peteprogrammer/go-automapper"
)

type Block struct {
	Id         uint64             `json:"id,omitempty"`
	Hash       []byte             `json:"hash,omitempty"`
	Prevhash   []byte             `json:"prevhash,omitempty"`
	RootHash   []byte             `json:"roothash,omitempty"`
	Nonce      uint32             `json:"nonce,omitempty"`
	Difficulty float64            `json:"difficulty,omitempty"`
	Time       time.Time          `json:"timestamp,omitempty"`
	Body       *merkle.MerkleTree `json:"-"`
}

func (block *Block) MarshalJSON() ([]byte, error) {
	var s = struct {
		Id         uint64    `json:"id,omitempty"`
		Hash       string    `json:"hash,omitempty"`
		Prevhash   string    `json:"prevhash,omitempty"`
		RootHash   string    `json:"roothash,omitempty"`
		Nonce      uint32    `json:"nonce,omitempty"`
		Difficulty string    `json:"difficulty,omitempty"`
		Timestamp  time.Time `json:"timestamp,omitempty"`
	}{
		Id:         block.Id,
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
	var s struct {
		Id         uint64    `json:"id,omitempty"`
		Hash       string    `json:"hash,omitempty"`
		Prevhash   string    `json:"prevhash,omitempty"`
		RootHash   string    `json:"roothash,omitempty"`
		Nonce      uint32    `json:"nonce,omitempty"`
		Difficulty string    `json:"difficulty,omitempty"`
		Timestamp  time.Time `json:"timestamp,omitempty"`
	}

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	block.Id = s.Id

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
	var tree merkle.MerkleTree

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

func (block *Block) FindHash() ([]byte, error) {
	var nonce uint32
	//return err if not find valid hash
	err := errors.ErrBlockContentInvalid
	for nonce = 1; nonce < math.MaxUint32; nonce++ {
		block.Nonce = nonce

		hash, err := block.ComputeHash()
		if err != nil {
			break
		}

		actual := bitcoin.ComputeDifficulty(hash)
		if actual <= block.Difficulty {
			return hash, nil
		}
	}
	block.Nonce = 0
	return nil, err
}

func (block *Block) ComputeHash() ([]byte, error) {
	originalHash := block.Hash
	block.Hash = []byte{}

	hash, err := cryptography.Hash(block)

	block.Hash = originalHash
	return hash, err
}
