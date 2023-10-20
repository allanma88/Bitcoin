package cryptography

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"reflect"
)

func Hash(e any) ([]byte, error) {
	str, err := marshal(e)
	if err != nil {
		return nil, err
	}
	// log.Printf("%s", str)

	sha := sha256.New()
	sha.Write([]byte(str))
	hash := sha.Sum(nil)

	return hash, nil
}

func Verify(pubkey, hash, signature []byte) (bool, error) {
	publicKey, err := DecodePublicKey(pubkey)
	if err != nil {
		return false, err
	}
	valid := ecdsa.VerifyASN1(publicKey, hash, signature)
	return valid, nil
}

func Sign(privkey, hash []byte) ([]byte, error) {
	privateKey, err := DecodePrivateKey(privkey)
	if err != nil {
		return nil, err
	}
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hash)
	return signature, err
}

func EncodePrivateKey(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	x509Encoded, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	bytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
	return bytes, nil
}

func EncodePublicKey(publicKey *ecdsa.PublicKey) ([]byte, error) {
	x509EncodedPub, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	bytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
	return bytes, nil
}

func DecodePrivateKey(bytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(bytes)
	x509Encoded := block.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	return privateKey, err
}

func DecodePublicKey(bytes []byte) (*ecdsa.PublicKey, error) {
	blockPub, _ := pem.Decode(bytes)
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	if err != nil {
		return nil, err
	}
	publicKey := genericPublicKey.(*ecdsa.PublicKey)
	return publicKey, nil
}

func marshal(e any) ([]byte, error) {
	if reflect.TypeOf(e).Kind() == reflect.String {
		raw := json.RawMessage(e.(string))
		return raw.MarshalJSON()
	} else {
		return json.Marshal(e)
	}
}
