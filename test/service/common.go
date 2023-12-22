package service

import (
	"Bitcoin/src/database"
	"encoding/json"
	"errors"
)

type TestTable struct {
	Items map[string][]byte
	Keys  []string
}

func newTestTable() *TestTable {
	items := make(map[string][]byte)
	keys := make([]string, 0)
	txdb := &TestTable{Items: items, Keys: keys}
	return txdb
}

func (table *TestTable) Save(key, val []byte) error {
	table.Items[string(key)] = val
	table.Keys = append(table.Keys, string(key))

	return nil
}

func (table *TestTable) Get(key []byte) ([]byte, error) {
	data, ok := table.Items[string(key)]
	if ok {
		return data, nil
	}
	return nil, nil
}

func (table *TestTable) Remove(key []byte) error {
	delete(table.Items, string(key))
	return nil
}

type TestBaseDB struct {
	Tables map[string]*TestTable
}

func newTestBaseDB() database.IBaseDB {
	tables := make(map[string]*TestTable)
	txdb := &TestBaseDB{Tables: tables}
	return txdb
}

func (db *TestBaseDB) Save(prefix []byte, key, val any) error {
	keydata, err := serialize(key)
	if err != nil {
		return err
	}

	valdata, err := serialize(val)
	if err != nil {
		return err
	}

	table, ok := db.Tables[string(prefix)]

	if !ok {
		table = newTestTable()
		db.Tables[string(prefix)] = table
	}

	return table.Save(keydata, valdata)
}

func (db *TestBaseDB) Get(prefix, key []byte, v any) (bool, error) {
	table, ok := db.Tables[string(prefix)]
	if !ok {
		return false, nil
	}
	data, err := table.Get(key)
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(data, v)
}

func (db *TestBaseDB) Move(oldPrefix, newPrefix, key, val []byte) error {
	oldTable, ok := db.Tables[string(oldPrefix)]
	if ok {
		err := oldTable.Remove(key)
		if err != nil {
			return err
		}
	}

	table, ok := db.Tables[string(newPrefix)]
	if !ok {
		table = newTestTable()
		db.Tables[string(newPrefix)] = table
	}

	return table.Save(key, val)
}

func (db *TestBaseDB) Filter(prefix, start []byte) ([][]byte, error) {
	return nil, errors.New("not implemented")
}

func (db *TestBaseDB) Size(prefix []byte) (int64, error) {
	return 0, errors.New("not implemented")
}

func (db *TestBaseDB) StartBatch() database.IBatch {
	return nil
}

func (db *TestBaseDB) EndBatch(batch database.IBatch) error {
	return errors.New("not implemented")
}

func (db *TestBaseDB) Close() error {
	return nil
}

func serialize(v any) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	return json.Marshal(v)
}
