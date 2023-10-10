package test

import (
	"Bitcoin/src/cryptography"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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
