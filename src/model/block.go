package model

import (
	"Bitcoin/src/merkle"

	"github.com/golang/protobuf/ptypes/timestamp"
)

type Block struct {
	Id         []byte               `json:"id,omitempty"`
	RootHash   []byte               `json:"roothash,omitempty"`
	Nonce      uint32               `json:"nonce,omitempty"`
	Difficulty uint32               `json:"difficulty,omitempty"`
	Prevhash   []byte               `json:"prevhash,omitempty"`
	Timestamp  *timestamp.Timestamp `json:"timestamp,omitempty"`
	Content    *merkle.MerkleTree   `json:"-"`
}
