package service

import (
	"Bitcoin/src/errors"
	"bytes"
	"time"
)

type Identity interface {
	ComputeHash() ([]byte, error)
}

func validateHash[T Identity](hash []byte, model Identity) ([]byte, error) {
	computeHash, err := model.ComputeHash()
	if err != nil {
		return nil, errors.ErrIdentityInvalid
	}

	if !bytes.Equal(computeHash, hash) {
		return nil, errors.ErrIdentityHashInvalid
	}
	return computeHash, nil
}

func validateTimestamp(t time.Time) error {
	if t.Compare(time.Now().Add(2*time.Hour)) >= 0 {
		return errors.ErrIdentityTooEarly
	}
	return nil
}
