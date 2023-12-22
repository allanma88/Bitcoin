package database

import (
	"bytes"
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type IBaseDB interface {
	Save(prefix []byte, key, val any) error
	Get(prefix, key []byte, val any) (bool, error)
	Filter(prefix, start []byte) ([][]byte, error)
	Size(prefix []byte) (int64, error)
	StartBatch() IBatch
	EndBatch(batch IBatch) error
	Close() error
}

type BaseDB struct {
	Database *leveldb.DB
}

func (db *BaseDB) Save(prefix []byte, key, val any) error {
	keydata, err := serialize(key)
	if err != nil {
		return err
	}

	valdata, err := serialize(val)
	if err != nil {
		return err
	}

	opt := &opt.WriteOptions{}
	k := makeKey(prefix, keydata)
	return db.Database.Put(k, valdata, opt)
}

func (db *BaseDB) Get(prefix, key []byte, val any) (bool, error) {
	opt := &opt.ReadOptions{}
	k := makeKey(prefix, key)

	data, err := db.Database.Get(k, opt)
	if err == leveldb.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, json.Unmarshal(data, val)
}

func (db *BaseDB) Filter(prefix, start []byte) ([][]byte, error) {
	opt := &opt.ReadOptions{}
	iter := db.Database.NewIterator(util.BytesPrefix(prefix), opt)

	vals := make([][]byte, 0)
	for ok := iter.Seek(start); ok; ok = iter.Next() {
		data := iter.Value()
		vals = append(vals, data)
	}
	iter.Release()

	return vals, nil
}

func (db *BaseDB) Size(prefix []byte) (int64, error) {
	sizes, err := db.Database.SizeOf([]util.Range{*util.BytesPrefix(prefix)})
	if err != nil {
		return 0, err
	}
	return sizes[0], nil
}

func (db *BaseDB) StartBatch() IBatch {
	return &BaseBatch{Batch: new(leveldb.Batch)}
}

func (db *BaseDB) EndBatch(batch IBatch) error {
	opt := &opt.WriteOptions{}
	baseBatch := batch.(BaseBatch)
	return db.Database.Write(baseBatch.Batch, opt)
}

func (db *BaseDB) Close() error {
	return db.Database.Close()
}

type IBatch interface {
	Save(prefix []byte, key, val any) error
}

type BaseBatch struct {
	Batch *leveldb.Batch
}

func (batch BaseBatch) Save(prefix []byte, key, val any) error {
	keydata, err := serialize(key)
	if err != nil {
		return err
	}

	valdata, err := serialize(val)
	if err != nil {
		return err
	}

	k := makeKey(prefix, keydata)
	batch.Batch.Put(k, valdata)

	return nil
}

func serialize(v any) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	return json.Marshal(v)
}

func makeKey(prefix, key []byte) []byte {
	return bytes.Join([][]byte{prefix, key}, []byte("-"))
}
