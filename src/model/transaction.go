package model

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/protocol"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/peteprogrammer/go-automapper"
)

type In struct {
	PrevHash  []byte `json:"prevHash,omitempty"`
	Index     uint32 `json:"index,omitempty"`
	Signature []byte `json:"signature,omitempty"`
}

func (in *In) MarshalJSON() ([]byte, error) {
	var s = struct {
		PrevHash  string `json:"prevHash,omitempty"`
		Index     uint32 `json:"index,omitempty"`
		Signature string `json:"signature,omitempty"`
	}{
		PrevHash:  hex.EncodeToString(in.PrevHash),
		Index:     in.Index,
		Signature: base64.RawStdEncoding.EncodeToString(in.Signature),
	}
	return json.Marshal(s)
}

func (in *In) UnmarshalJSON(data []byte) error {
	var s struct {
		PrevHash  string `json:"prevHash,omitempty"`
		Index     uint32 `json:"index,omitempty"`
		Signature string `json:"signature,omitempty"`
	}

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	in.PrevHash, err = hex.DecodeString(s.PrevHash)
	if err != nil {
		return err
	}

	in.Signature, err = base64.RawStdEncoding.DecodeString(s.Signature)
	if err != nil {
		return err
	}

	in.Index = s.Index
	return err
}

type Out struct {
	Pubkey []byte `json:"pubkey,omitempty"`
	Value  uint64 `json:"value,omitempty"`
}

func (out *Out) MarshalJSON() ([]byte, error) {
	var s = struct {
		Pubkey string `json:"pubkey,omitempty"`
		Value  uint64 `json:"value,omitempty"`
	}{
		Pubkey: base64.RawStdEncoding.EncodeToString(out.Pubkey),
		Value:  out.Value,
	}
	return json.Marshal(s)
}

func (out *Out) UnmarshalJSON(data []byte) error {
	var s struct {
		Pubkey string `json:"pubkey,omitempty"`
		Value  uint64 `json:"value,omitempty"`
	}

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	out.Pubkey, err = base64.RawStdEncoding.DecodeString(s.Pubkey)
	if err != nil {
		return err
	}

	out.Value = s.Value
	return err
}

type Transaction struct {
	Hash      []byte    `json:"hash,omitempty"`
	InLen     uint32    `json:"in_len,omitempty"`
	OutLen    uint32    `json:"out_len,omitempty"`
	Ins       []*In     `json:"ins,omitempty"`
	Outs      []*Out    `json:"outs,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {
	var s = struct {
		Hash      string    `json:"hash,omitempty"`
		InLen     uint32    `json:"in_len,omitempty"`
		OutLen    uint32    `json:"out_len,omitempty"`
		Ins       []*In     `json:"ins,omitempty"`
		Outs      []*Out    `json:"outs,omitempty"`
		Timestamp time.Time `json:"timestamp,omitempty"`
	}{
		Hash:      hex.EncodeToString(tx.Hash),
		InLen:     tx.InLen,
		OutLen:    tx.OutLen,
		Ins:       tx.Ins,
		Outs:      tx.Outs,
		Timestamp: tx.Timestamp,
	}
	return json.Marshal(s)
}

func (tx *Transaction) UnmarshalJSON(data []byte) error {
	var s struct {
		Hash      string    `json:"hash,omitempty"`
		InLen     uint32    `json:"in_len,omitempty"`
		OutLen    uint32    `json:"out_len,omitempty"`
		Ins       []*In     `json:"ins,omitempty"`
		Outs      []*Out    `json:"outs,omitempty"`
		Timestamp time.Time `json:"timestamp,omitempty"`
	}

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	tx.Hash, err = hex.DecodeString(s.Hash)
	if err != nil {
		return err
	}

	tx.InLen = s.InLen
	tx.OutLen = s.OutLen
	tx.Ins = s.Ins
	tx.Outs = s.Outs
	tx.Timestamp = s.Timestamp
	return err
}

func TransactionFrom(request *protocol.TransactionReq) *Transaction {
	var tx Transaction
	automapper.MapLoose(request, &tx)

	tx.Timestamp = time.UnixMilli(request.Time)
	return &tx
}

func TransactionTo(tx *Transaction) *protocol.TransactionReq {
	var request protocol.TransactionReq
	automapper.MapLoose(tx, &request)

	request.Time = tx.Timestamp.UnixMilli()
	return &request
}

func (tx *Transaction) ComputeHash() ([]byte, error) {
	originalHash := tx.Hash
	tx.Hash = []byte{}

	hash, err := cryptography.Hash(tx)

	tx.Hash = originalHash
	return hash, err
}
