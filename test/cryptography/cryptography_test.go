package crypto

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"
)

func Test_Hash_Field_Order_Doesnot_Matter(t *testing.T) {
	timestamp := time.Now()
	tx1 := &model.Transaction{
		InLen:     1,
		OutLen:    1,
		Timestamp: timestamp,
	}
	tx2 := &model.Transaction{
		OutLen:    1,
		InLen:     1,
		Timestamp: timestamp,
	}
	hash1, _ := cryptography.Hash(tx1)
	hash2, _ := cryptography.Hash(tx2)
	if bytes.Equal(hash1, hash2) {
		t.Log("Field order is not matter when serialize")
	} else {
		t.Fatalf("hash is not equal: %x != %x", hash1, hash2)
	}
}

func Test_Verify(t *testing.T) {
	privatKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Generate key err: %s", err)
	}

	hash, err := cryptography.Hash([]byte("HelloWorld"))
	if err != nil {
		t.Fatalf("Hash err: %s", err)
	}

	signature, err := ecdsa.SignASN1(rand.Reader, privatKey, hash)
	if err != nil {
		t.Fatalf("Sign err: %s", err)
	}

	pubkey, err := cryptography.EncodePublicKey(&privatKey.PublicKey)
	if err != nil {
		t.Fatalf("Encode public key err: %s", err)
	}

	valid, err := cryptography.Verify(pubkey, hash, signature)
	if err != nil {
		t.Fatalf("Verify signature err: %s", err)
	}

	if !valid {
		t.Fatalf("Verify failed")
	}
}

func Test_String_Hash(t *testing.T) {
	inputs := []string{
		"hello world",
		`{"inLen":1,"outLen":1,"ins":[{}],"outs":[{}],"timestamp":{"seconds":1359849600}}`,
	}
	expects := []string{
		"b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		"674b98e503495c34167d94f6c7198b3250513cdc72285ca03cb672968743a705",
	}
	for i := 0; i < len(inputs); i++ {
		hash, err := cryptography.Hash(inputs[i])
		if err != nil {
			t.Fatalf("Hash err: %s", err)
		}

		actual := hex.EncodeToString(hash)
		if actual != expects[i] {
			t.Fatalf("Wrong hash, expect: %v, actual: %v", expects[i], actual)
		}
	}
}

func Test_Transaction_Hash(t *testing.T) {
	timestamp, err := time.Parse("2006-Jan-02", "2013-Feb-03")
	if err != nil {
		t.Fatalf("Parse time err: %s", err)
	}
	tx := &model.Transaction{
		InLen:     1,
		OutLen:    1,
		Ins:       []*model.In{{}},
		Outs:      []*model.Out{{}},
		Timestamp: timestamp,
	}

	hash, err := cryptography.Hash(tx)
	if err != nil {
		t.Fatalf("Hash err: %s", err)
	}

	actual := hex.EncodeToString(hash)
	expect := "2c15e1e873280cf7f2842a661a085f93a64552f81d76527e98d664e46f8eade9"
	if actual != expect {
		t.Fatalf("Wrong hash, expect: %v, actual: %v", expect, actual)
	}
}

func Test_TransactionReq_Hash(t *testing.T) {
	timestamp, err := time.Parse("2006-Jan-02", "2013-Feb-03")
	if err != nil {
		t.Fatalf("Parse time err: %s", err)
	}
	req := &protocol.TransactionReq{
		InLen:  1,
		OutLen: 1,
		Ins:    []*protocol.InReq{{}},
		Outs:   []*protocol.OutReq{{}},
		Time:   timestamp.UnixMilli(),
	}

	hash, err := cryptography.Hash(req)
	if err != nil {
		t.Fatalf("Hash err: %s", err)
	}

	actual := hex.EncodeToString(hash)
	expect := "32594ab9b72aff0eb0e2126166c4524556e3880aff596d836572edccc57a9316"
	if actual != expect {
		t.Fatalf("Wrong hash, expect: %v, actual: %v", expect, actual)
	}
}
