package database

import (
	"bytes"
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type IBaseDB[T any] interface {
	Save(prefix, key []byte, val T) error
	Get(prefix, key []byte, val T) error
	Close() error
}

type BaseDB[T any] struct {
	Database *leveldb.DB
}

func (db *BaseDB[T]) Save(prefix, key []byte, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	opt := &opt.WriteOptions{}
	k := bytes.Join([][]byte{prefix, key}, []byte("-"))
	err = db.Database.Put(k, data, opt)
	return err
}

func (db *BaseDB[T]) Get(prefix, key []byte, val T) error {
	opt := &opt.ReadOptions{}
	k := bytes.Join([][]byte{prefix, key}, []byte("-"))
	bytes, err := db.Database.Get(k, opt)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, val)
	if err != nil {
		return err
	}
	return nil
}

func (db *BaseDB[T]) Close() error {
	return db.Database.Close()
}
