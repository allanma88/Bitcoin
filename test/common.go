package test

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/cryptography"
	"Bitcoin/src/infra"
	"Bitcoin/src/model"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"time"
)

func NewKeys() ([]byte, []byte) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("generate keys error: %s", err)
	}
	privkey, err := cryptography.EncodePrivateKey(privateKey)
	if err != nil {
		log.Fatalf("encode private key error: %s", err)
	}
	pubkey, err := cryptography.EncodePublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatalf("encode public key error: %s", err)
	}
	return privkey, pubkey
}

func NewTransaction(blockHash []byte) *model.Transaction {
	privkey, pubkey := NewKeys()

	prevHash, err := cryptography.Hash("whatever")
	if err != nil {
		log.Fatalf("compute hash error: %s", err)
	}

	sig, err := cryptography.Sign(privkey, prevHash)
	if err != nil {
		log.Fatalf("sign prev hash error: %s", err)
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

	if blockHash == nil {
		blockHash, err = cryptography.Hash("block")
		if err != nil {
			log.Fatalf("compute block hash error: %s", err)
		}
	}

	tx := &model.Transaction{
		InLen:     uint32(len(ins)),
		OutLen:    uint32(len(outs)),
		Ins:       ins,
		Outs:      outs,
		Timestamp: time.Now(),
		BlockHash: blockHash,
	}

	hash, err := tx.ComputeHash()
	if err != nil {
		log.Fatalf("compute tx hash error: %s", err)
	}
	tx.Hash = hash

	return tx
}

func NewBlock(number uint64, difficultyLevel uint64, prevhash []byte) *model.Block {
	txs := make([]*model.Transaction, 4)
	for i := 0; i < len(txs); i++ {
		txs[i] = NewTransaction([]byte{})
	}

	tree, err := collection.BuildTree(txs)
	if err != nil {
		log.Fatalf("builder merkle tree error: %s", err)
	}

	if prevhash == nil {
		prevhash, err = cryptography.Hash("prev")
		if err != nil {
			log.Fatalf("compute prev hash error: %s", err)
		}
	}

	rootHash := tree.Table[len(tree.Table)-1][0].Hash

	block := &model.Block{
		Prevhash:   prevhash,
		Number:     number,
		RootHash:   rootHash,
		Difficulty: infra.ComputeDifficulty(infra.MakeDifficulty(difficultyLevel)),
		Time:       time.Now(),
		Body:       tree,
	}

	hash, err := block.FindHash(context.TODO())
	if err != nil {
		log.Fatalf("find block hash error: %s", err)
	}
	block.Hash = hash
	for _, tx := range txs {
		tx.BlockHash = hash
	}

	return block
}
