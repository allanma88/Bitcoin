package model

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"Bitcoin/test"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"
)

func Test__Transaction_ComputeHash_Hash_Not_Change(t *testing.T) {
	tx, err := newTransaction()
	if err != nil {
		t.Fatalf("new transaction error: %v", err)
	}

	hash, err := tx.ComputeHash()
	if err != nil {
		t.Fatalf("compute hash error: %v", err)
	}

	if bytes.Equal(hash, tx.Hash) {
		t.Log("transaction hash didn't changed after serialize")
	} else {
		t.Fatalf("transaction hash is changed from [%x] to [%x]", tx.Hash, hash)
	}
}

func Test_Transaction_From(t *testing.T) {
	req, err := newTransactionReq()
	if err != nil {
		t.Fatalf("new transaction request error: %s", err)
	}

	tx := model.TransactionFrom(req)

	if !equalTx(req, tx) {
		t.Fatalf("transaction request are not equal with transaction")
	}
}

func Test_Transaction_To(t *testing.T) {
	tx, err := newTransaction()
	if err != nil {
		t.Fatalf("new transaction request error: %s", err)
	}

	req := model.TransactionTo(tx)

	if !equalTx(req, tx) {
		t.Fatalf("transaction request are not equal with transaction")
	}
}

func Test_Transaction_Marshal(t *testing.T) {
	tx, err := newTransaction()
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
	}

	data, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}
	t.Logf("json: %s", string(data))
}

func Test_Transaction_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("transaction.json")
	if err != nil {
		t.Fatalf("read file error: %s", err)
	}

	var tx model.Transaction
	err = json.Unmarshal(data, &tx)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}

	expect := "50b0df8443af29f0cf81dd4b27823fb41aabf3201d14cc3da8019050f5b8a967"
	actual := hex.EncodeToString(tx.Hash)
	if actual != expect {
		t.Fatalf("expect hash is %s, actual is %s", expect, actual)
	}
}

func newTransaction() (*model.Transaction, error) {
	privkey, pubkey, err := test.NewKeys()
	if err != nil {
		return nil, err
	}

	prevHash, err := cryptography.Hash("whatever")
	if err != nil {
		return nil, err
	}

	sig, err := cryptography.Sign(privkey, prevHash)
	if err != nil {
		return nil, err
	}

	ins := []*model.In{
		{
			PrevHash:  prevHash,
			Index:     0,
			Signature: sig,
		},
	}

	outs := []*model.Out{{
		Pubkey: pubkey,
		Value:  1,
	}}

	tx := &model.Transaction{
		InLen:     uint32(len(ins)),
		OutLen:    uint32(len(outs)),
		Ins:       ins,
		Outs:      outs,
		Timestamp: time.Now(),
	}

	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, err
	}
	tx.Hash = hash

	return tx, nil
}

func newTransactionReq() (*protocol.TransactionReq, error) {
	privkey, pubkey, err := test.NewKeys()
	if err != nil {
		return nil, err
	}

	prevHash, err := cryptography.Hash("whatever")
	if err != nil {
		return nil, err
	}

	sig, err := cryptography.Sign(privkey, prevHash)
	if err != nil {
		return nil, err
	}

	ins := []*protocol.InReq{
		{
			PrevHash:  prevHash,
			Index:     0,
			Signature: sig,
		},
	}

	out := &protocol.OutReq{
		Pubkey: pubkey,
		Value:  1,
	}
	outs := []*protocol.OutReq{out}
	tx := &protocol.TransactionReq{
		InLen:  uint32(len(ins)),
		OutLen: uint32(len(outs)),
		Ins:    ins,
		Outs:   outs,
		Time:   time.Now().UnixMilli(),
	}

	return tx, nil
}

func equalTx(req *protocol.TransactionReq, tx *model.Transaction) bool {
	if !bytes.Equal(req.Hash, tx.Hash) {
		log.Printf("hash: %x != %x", req.Hash, tx.Hash)
		return false
	}
	if req.Time != tx.Timestamp.UnixMilli() {
		log.Printf("time: %v != %v", req.Time, tx.Timestamp.UnixMilli())
		return false
	}
	if !equalIns(req.Ins, tx.Ins) {
		log.Print("ins not equal")
		return false
	}
	if !equalOuts(req.Outs, tx.Outs) {
		log.Print("outs not equal")
		return false
	}
	return true
}

func equalIns(reqs []*protocol.InReq, ins []*model.In) bool {
	if len(reqs) != len(ins) {
		return false
	}
	for i := 0; i < len(reqs); i++ {
		if !bytes.Equal(reqs[i].PrevHash, ins[i].PrevHash) || !bytes.Equal(reqs[i].Signature, ins[i].Signature) || reqs[i].Index != ins[i].Index {
			return false
		}
	}

	return true
}

func equalOuts(reqs []*protocol.OutReq, outs []*model.Out) bool {
	if len(reqs) != len(outs) {
		return false
	}
	for i := 0; i < len(reqs); i++ {
		if !bytes.Equal(reqs[i].Pubkey, outs[i].Pubkey) || reqs[i].Value != outs[i].Value {
			return false
		}
	}
	return true
}
