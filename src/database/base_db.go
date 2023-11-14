package database

import (
	"bytes"
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type IBaseDB[T any] interface {
	Save(prefix, key []byte, val *T) error
	Get(prefix, key []byte) (*T, error)
	Move(oldPrefix, newPrefix, key []byte, val *T) error
	Close() error
}

type BaseDB[T any] struct {
	Database *leveldb.DB
}

func (db *BaseDB[T]) Save(prefix, key []byte, val *T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	opt := &opt.WriteOptions{}
	k := makeKey(prefix, key)
	err = db.Database.Put(k, data, opt)
	return err
}

func (db *BaseDB[T]) Get(prefix, key []byte) (*T, error) {
	opt := &opt.ReadOptions{}
	k := makeKey(prefix, key)

	data, err := db.Database.Get(k, opt)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var val T
	err = json.Unmarshal(data, &val)
	if err != nil {
		return nil, err
	}
	return &val, nil
}

// TODO: test case
func (db *BaseDB[T]) Move(oldPrefix, newPrefix, key []byte, val *T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	tx, err := db.Database.OpenTransaction()
	if err != nil {
		return err
	}

	opt := &opt.WriteOptions{}
	oldK := makeKey(oldPrefix, key)
	err = tx.Delete(oldK, opt)
	if err != nil {
		tx.Discard()
		return err
	}

	newK := makeKey(newPrefix, key)
	err = tx.Put(newK, data, opt)
	if err != nil {
		tx.Discard()
		return err
	}

	return tx.Commit()
}

func (db *BaseDB[T]) Close() error {
	return db.Database.Close()
}

func makeKey(prefix, key []byte) []byte {
	return bytes.Join([][]byte{prefix, key}, []byte("-"))
}
