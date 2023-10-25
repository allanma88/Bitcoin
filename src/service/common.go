package service

import (
	"Bitcoin/src/errors"
	"bytes"
	"time"
)

type Identity interface {
	ComputeHash() ([]byte, error)
}

func validateHash[T Identity](id []byte, model Identity) ([]byte, error) {
	hash, err := model.ComputeHash()
	if err != nil {
		return nil, errors.ErrIdentityInvalid
	}
	if !bytes.Equal(hash, id) {
		return nil, errors.ErrIdentityHashInvalid
	}
	return hash, nil
}

func validateTimestamp(t time.Time) error {
	if t.Compare(time.Now().Add(2*time.Hour)) >= 0 {
		return errors.ErrIdentityTooEarly
	}
	return nil
}
