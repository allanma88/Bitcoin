package model

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"Bitcoin/test"
	"bytes"
	"testing"

	"github.com/peteprogrammer/go-automapper"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_ComputeHash_Hash_Not_Change(t *testing.T) {
	tx := &model.Transaction{
		InLen:     0,
		OutLen:    0,
		Timestamp: timestamppb.Now(),
	}
	originalHash, err := cryptography.Hash(tx)
	if err != nil {
		t.Fatalf("transaction hash error: %s", err)
	}

	tx.Id = originalHash

	tx.ComputeHash()
	if bytes.Equal(originalHash, tx.Id) {
		t.Log("transaction hash didn't changed after serialize")
	} else {
		t.Fatalf("transaction hash is changed from [%x] to [%x]", originalHash, tx.Id)
	}
}

func Test_Convert(t *testing.T) {
	req, err := newTransactionReq()
	if err != nil {
		t.Fatalf("new transaction request error: %s", err)
	}

	var tx model.Transaction
	automapper.MapLoose(req, &tx)

	if !equalTx(req, &tx) {
		t.Fatalf("transaction request are not equal with transaction")
	}
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
		InLen:     uint32(len(ins)),
		OutLen:    uint32(len(outs)),
		Ins:       ins,
		Outs:      outs,
		Timestamp: timestamppb.Now(),
	}

	return tx, nil
}

func equalTx(req *protocol.TransactionReq, tx *model.Transaction) bool {
	if !bytes.Equal(req.Id, tx.Id) {
		return false
	}
	if req.Timestamp != tx.Timestamp || req.InLen != tx.InLen || req.OutLen != tx.OutLen {
		return false
	}
	if !equalIns(req.Ins, tx.Ins) {
		return false
	}
	if !equalOuts(req.Outs, tx.Outs) {
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
