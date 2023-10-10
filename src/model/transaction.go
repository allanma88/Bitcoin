package model

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/protocol"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/peteprogrammer/go-automapper"
)

type In struct {
	PrevHash  []byte `json:"prevHash,omitempty"`
	Index     uint32 `json:"index,omitempty"`
	Signature []byte `json:"signature,omitempty"`
}

type Out struct {
	Pubkey []byte `json:"pubkey,omitempty"`
	Value  uint64 `json:"value,omitempty"`
}

type Transaction struct {
	Id        []byte               `json:"id,omitempty"`
	InLen     uint32               `json:"in_len,omitempty"`
	OutLen    uint32               `json:"out_len,omitempty"`
	Ins       []*In                `json:"ins,omitempty"`
	Outs      []*Out               `json:"outs,omitempty"`
	Timestamp *timestamp.Timestamp `json:"timestamp,omitempty"`
}

func (tx *Transaction) ComputeHash() ([]byte, error) {
	originalHash := tx.Id
	tx.Id = []byte{}

	hash, err := cryptography.Hash(tx)

	tx.Id = originalHash
	return hash, err
}

func TransactionFrom(request *protocol.TransactionReq) (*Transaction, error) {
	var tx Transaction
	automapper.MapLoose(request, &tx)
	return &tx, nil
}

func TransactionTo(tx *Transaction, nodes []string) *protocol.TransactionReq {
	var request protocol.TransactionReq
	automapper.MapLoose(tx, &request)
	request.Nodes = nodes
	return &request
}
