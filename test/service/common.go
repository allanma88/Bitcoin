package service

import (
	"Bitcoin/src/database"
	"encoding/json"
)

type TestTable[T any] struct {
	Items map[string][]byte
	Keys  []string
}

func newTestTable[T any]() *TestTable[T] {
	items := make(map[string][]byte)
	keys := make([]string, 0)
	txdb := &TestTable[T]{Items: items, Keys: keys}
	return txdb
}

func (table *TestTable[T]) Save(key []byte, val *T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	table.Items[string(key)] = data
	table.Keys = append(table.Keys, string(key))

	return nil
}

func (table *TestTable[T]) Get(key []byte) (*T, error) {
	data, ok := table.Items[string(key)]
	if ok {
		var val T
		err := json.Unmarshal(data, &val)
		if err != nil {
			return nil, err
		}
		return &val, nil
	}
	return nil, nil
}

func (table *TestTable[T]) Remove(key []byte) error {
	delete(table.Items, string(key))
	return nil
}

func (table *TestTable[T]) Last(n int) ([]*T, error) {
	if n > len(table.Keys) {
		n = len(table.Keys)
	}

	vals := make([]*T, 0, n)
	for i := n; i > 0; i-- {
		key := table.Keys[len(table.Keys)-i]
		data := table.Items[key]

		var val T
		err := json.Unmarshal(data, &val)
		if err != nil {
			return nil, err
		}
		vals = append(vals, &val)
	}

	return vals, nil
}

type TestBaseDB[T any] struct {
	Tables map[string]*TestTable[T]
}

func newTestBaseDB[T any]() database.IBaseDB[T] {
	tables := make(map[string]*TestTable[T])
	txdb := &TestBaseDB[T]{Tables: tables}
	return txdb
}

func (db *TestBaseDB[T]) Save(prefix, key []byte, val *T) error {
	table, ok := db.Tables[string(prefix)]

	if !ok {
		table = newTestTable[T]()
		db.Tables[string(prefix)] = table
	}

	return table.Save(key, val)
}

func (db *TestBaseDB[T]) Get(prefix, key []byte) (*T, error) {
	table, ok := db.Tables[string(prefix)]
	if !ok {
		return nil, nil
	}
	return table.Get(key)
}

func (db *TestBaseDB[T]) Remove(prefix, key []byte) error {
	table, ok := db.Tables[string(prefix)]
	if ok {
		return table.Remove(key)
	}
	return nil
}

func (db *TestBaseDB[T]) Last(prefix []byte, n int) ([]*T, error) {
	table, ok := db.Tables[string(prefix)]
	if ok {
		return table.Last(n)
	}
	return nil, nil
}

func (db *TestBaseDB[T]) Close() error {
	return nil
}
