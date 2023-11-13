package database

import (
	"Bitcoin/src/database"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	TestTable = "test"
)

func Test_BaseDB_Get(t *testing.T) {
	db, err := leveldb.OpenFile(DBPath, nil)
	if err != nil {
		t.Fatalf("open %s error: %v", DBPath, err)
	}
	defer cleanUp(db, DBPath)

	basedb := &database.BaseDB[string]{Database: db}
	key := "Hello"
	val := "Hello"
	err = basedb.Save([]byte(TestTable), []byte(key), &val)
	if err != nil {
		t.Fatalf("save %s error: %v", key, err)
	}

	data, err := basedb.Get([]byte(TestTable), []byte(key))
	if err != nil {
		t.Fatalf("get %s error: %v", key, err)
	}

	if *data != val {
		t.Fatalf("should get %s, but %s", val, *data)
	}
}
