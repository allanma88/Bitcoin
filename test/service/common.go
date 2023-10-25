package service

import (
	"Bitcoin/src/database"
	"bytes"
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
)

type Identity interface {
	GetId() []byte
}

type TestBaseDB[T any] struct {
	Items map[string][]byte
}

func newTestBaseDB[T any]() database.IBaseDB[T] {
	items := make(map[string][]byte)
	txdb := &TestBaseDB[T]{Items: items}
	return txdb
}

func (db *TestBaseDB[T]) Save(prefix, key []byte, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	k := bytes.Join([][]byte{prefix, key}, []byte("-"))
	db.Items[string(k)] = data
	return nil
}

func (db *TestBaseDB[T]) Get(prefix, key []byte, val T) error {
	k := bytes.Join([][]byte{prefix, key}, []byte("-"))
	data, ok := db.Items[string(k)]
	if ok {
		err := json.Unmarshal(data, val)
		if err != nil {
			return err
		}
		return nil
	}
	return leveldb.ErrNotFound
}

func (db *TestBaseDB[T]) Close() error {
	return nil
}
