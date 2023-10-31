package model

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/errors"
	"Bitcoin/src/merkle"
	"Bitcoin/src/protocol"
	"encoding/hex"
	"encoding/json"
	"math"
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
	Timestamp  time.Time          `json:"timestamp,omitempty"`
	Body       *merkle.MerkleTree `json:"-"`
}

func (block *Block) MarshalJSON() ([]byte, error) {
	var s = struct {
		Id         uint64    `json:"id,omitempty"`
		Hash       string    `json:"hash,omitempty"`
		Prevhash   string    `json:"prevhash,omitempty"`
		RootHash   string    `json:"roothash,omitempty"`
		Nonce      uint32    `json:"nonce,omitempty"`
		Difficulty float64   `json:"difficulty,omitempty"`
		Timestamp  time.Time `json:"timestamp,omitempty"`
	}{
		Id:         block.Id,
		Hash:       hex.EncodeToString(block.Hash),
		Prevhash:   hex.EncodeToString(block.Prevhash),
		RootHash:   hex.EncodeToString(block.RootHash),
		Nonce:      block.Nonce,
		Difficulty: block.Difficulty,
		Timestamp:  block.Timestamp,
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
		Difficulty float64   `json:"difficulty,omitempty"`
		Timestamp  time.Time `json:"timestamp,omitempty"`
	}

	block.Id = s.Id

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

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

	block.Nonce = s.Nonce
	block.Difficulty = s.Difficulty
	block.Timestamp = s.Timestamp
	return err
}

func BlockFrom(request *protocol.BlockReq) (*Block, error) {
	var block Block

	err := json.Unmarshal(request.Content, block.Body)
	if err != nil {
		return nil, err
	}

	automapper.MapLoose(request, &block)
	block.Timestamp = time.UnixMilli(request.Timestamp)
	return &block, nil
}

func BlockTo(block *Block, nodes []string) (*protocol.BlockReq, error) {
	content, err := json.Marshal(block.Body)
	if err != nil {
		return nil, err
	}

	var request protocol.BlockReq
	automapper.MapLoose(block, &request)
	request.Timestamp = block.Timestamp.UnixMilli()

	request.Content = content
	request.Nodes = nodes
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

		actual := ComputeDifficulty(hash)
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

func ComputeDifficulty(hash []byte) float64 {
	var n float64 = 0

	for i := 0; i < len(hash); i++ {
		n = n + float64(hash[i])
		n = n * math.Pow(2, 8) //TODO: slow
	}

	return n
}

func MakeDifficulty(z int) []byte {
	difficulty := make([]byte, 32)
	for i := 0; i < 32; i++ {
		difficulty[i] = 255
	}
	for i := 0; i < z; i++ {
		p := i / 8
		q := 7 - i%8
		difficulty[p] = difficulty[p] ^ (1 << q)
	}
	return difficulty
}
