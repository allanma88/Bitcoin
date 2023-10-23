package database

import (
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type BaseDB[T any] struct {
	Database *leveldb.DB
}

func (db *BaseDB[T]) Save(prefix, key []byte, val T) error {
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}

	opt := &opt.WriteOptions{}
	err = db.Database.Put(key, bytes, opt)
	return err
}

func (db *BaseDB[T]) Get(prefix, key []byte, val T) error {
	opt := &opt.ReadOptions{}
	bytes, err := db.Database.Get(key, opt)
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
