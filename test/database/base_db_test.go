package database

import (
	"Bitcoin/src/database"
	"fmt"
	"os"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

func Test_Get(t *testing.T) {
	db, err := leveldb.OpenFile(DBPath, nil)
	if err != nil {
		t.Fatalf("open %s error: %v", DBPath, err)
	}
	defer cleanUp(db, DBPath)

	basedb := &database.BaseDB[string]{Database: db}
	key := "Hello"
	val := "Hello"
	err = basedb.Save([]byte(database.TxTable), []byte(key), &val)
	if err != nil {
		t.Fatalf("save %s error: %v", key, err)
	}

	data, err := basedb.Get([]byte(database.TxTable), []byte(key))
	if err != nil {
		t.Fatalf("get %s error: %v", key, err)
	}

	if *data != val {
		t.Fatalf("should get %s, but %s", val, *data)
	}
}

func Test_Lasts(t *testing.T) {
	db, err := leveldb.OpenFile(DBPath, nil)
	if err != nil {
		t.Fatalf("open %s error: %v", DBPath, err)
	}
	defer cleanUp(db, DBPath)

	basedb := &database.BaseDB[string]{Database: db}

	n := 5
	keys := make([]string, n)
	vals := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("Hello%d", i)
		vals[i] = fmt.Sprintf("Hello%d", i)

		err = basedb.Save([]byte(database.TxTable), []byte(keys[i]), &vals[i])
		if err != nil {
			t.Fatalf("save %s error: %v", keys[i], err)
		}
	}

	lastVals, err := basedb.Last([]byte(database.TxTable), n)
	if err != nil {
		t.Fatalf("last error: %v", err)
	}
	if len(lastVals) != 5 {
		t.Fatalf("should get %d values, but actually %d", n, len(vals))
	}

	for i := 0; i < n; i++ {
		if vals[i] != *lastVals[i] {
			t.Fatalf("should get %s, but %s", vals[i], *lastVals[i])
		}
	}
}

func cleanUp(db *leveldb.DB, path string) {
	db.Close()
	os.RemoveAll(path)
}
