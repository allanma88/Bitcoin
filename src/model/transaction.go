package model

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/cryptography"
	"Bitcoin/src/protocol"
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/peteprogrammer/go-automapper"
)

type In struct {
	PrevHash  []byte
	PrevOut   *Out //TODO: waste memory
	Index     uint32
	Signature []byte
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
	Hash      []byte
	InLen     uint32
	OutLen    uint32
	Ins       []*In
	Outs      []*Out
	Timestamp time.Time
	BlockHash []byte
	Fee       uint64
}

type jTransaction struct {
	Hash      string    `json:"hash,omitempty"`
	InLen     uint32    `json:"in_len,omitempty"`
	OutLen    uint32    `json:"out_len,omitempty"`
	Ins       []*In     `json:"ins,omitempty"`
	Outs      []*Out    `json:"outs,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	BlockHash string    `json:"block_hash,omitempty"`
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {
	var jtx = jTransaction{
		Hash:      hex.EncodeToString(tx.Hash),
		InLen:     tx.InLen,
		OutLen:    tx.OutLen,
		Ins:       tx.Ins,
		Outs:      tx.Outs,
		Timestamp: tx.Timestamp,
		BlockHash: hex.EncodeToString(tx.BlockHash),
	}
	return json.Marshal(jtx)
}

func (tx *Transaction) UnmarshalJSON(data []byte) error {
	var jtx jTransaction

	err := json.Unmarshal(data, &jtx)
	if err != nil {
		return err
	}

	tx.Hash, err = hex.DecodeString(jtx.Hash)
	if err != nil {
		return err
	}

	tx.BlockHash, err = hex.DecodeString(jtx.BlockHash)
	if err != nil {
		return err
	}

	tx.InLen = jtx.InLen
	tx.OutLen = jtx.OutLen
	tx.Ins = jtx.Ins
	tx.Outs = jtx.Outs
	tx.Timestamp = jtx.Timestamp
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
	//TODO: more general way to compute hash, use tag and no need assign the value of each field
	ins := make([]*In, len(tx.Ins))
	for i := 0; i < len(tx.Ins); i++ {
		in := tx.Ins[i]
		ins[i] = &In{
			PrevHash:  in.PrevHash,
			Index:     in.Index,
			Signature: in.Signature,
		}
	}

	newtx := &Transaction{
		InLen:     tx.InLen,
		OutLen:    tx.OutLen,
		Ins:       ins,
		Outs:      tx.Outs,
		Timestamp: tx.Timestamp,
	}

	return cryptography.Hash(newtx)
}

func MakeCoinbaseTx(pubkey []byte, val uint64) (*Transaction, error) {
	tx := &Transaction{
		InLen:  0,
		OutLen: 1,
		Ins:    []*In{},
		Outs: []*Out{
			{
				Pubkey: pubkey,
				Value:  val,
			},
		},
		Timestamp: time.Now(),
	}

	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, err
	}

	tx.Hash = hash
	return tx, nil
}

func (tx *Transaction) Compare(other collection.Comparable) int {
	otherTx := other.(*Transaction)
	if tx.Fee < otherTx.Fee {
		return -1
	} else if tx.Fee == otherTx.Fee {
		return 0
	}
	return 1
}

func (tx *Transaction) Equal(other collection.Comparable) bool {
	otherTx := other.(*Transaction)
	return bytes.Equal(tx.Hash, otherTx.Hash)
}
