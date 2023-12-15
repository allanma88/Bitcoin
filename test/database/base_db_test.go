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

	basedb := &database.BaseDB{Database: db}
	key := "Hello"
	val := "Hello"
	err = basedb.Save([]byte(TestTable), []byte(key), []byte(val))
	if err != nil {
		t.Fatalf("save %s error: %v", key, err)
	}

	var s string
	has, err := basedb.Get([]byte(TestTable), []byte(key), &s)
	if err != nil {
		t.Fatalf("get %s error: %v", key, err)
	}
	if !has {
		t.Fatalf("no %s", key)
	}

	if s != val {
		t.Fatalf("should get %s, but %s", val, s)
	}
}
