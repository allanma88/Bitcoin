package test

import (
	"Bitcoin/src/bitcoin"
	"Bitcoin/src/cryptography"
	"Bitcoin/src/merkle"
	"Bitcoin/src/model"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"time"
)

func NewKeys() ([]byte, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	privkey, err := cryptography.EncodePrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}
	pubkey, err := cryptography.EncodePublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	return privkey, pubkey, nil
}

func NewTransaction() (*model.Transaction, error) {
	privkey, pubkey, err := NewKeys()
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

func NewBlock(id uint64, difficultyLevel uint64) (*model.Block, error) {
	prevHash, err := cryptography.Hash("prev")
	if err != nil {
		return nil, err
	}

	txs := make([]*model.Transaction, 4)
	for i := 0; i < len(txs); i++ {
		txs[i], err = NewTransaction()
		if err != nil {
			return nil, err
		}
	}

	tree, err := merkle.BuildTree(txs)
	if err != nil {
		return nil, err
	}

	rootHash := tree.Table[len(tree.Table)-1][0].Hash

	block := &model.Block{
		Prevhash:   prevHash,
		Id:         id,
		RootHash:   rootHash,
		Difficulty: bitcoin.ComputeDifficulty(bitcoin.MakeDifficulty(difficultyLevel)),
		Time:       time.Now(),
		Body:       tree,
	}

	hash, err := block.FindHash()
	if err != nil {
		return nil, err
	}
	block.Hash = hash

	return block, nil
}
