package database

import (
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	DBPath = "bitcoin"
)

func cleanUp(db *leveldb.DB, path string) {
	db.Close()
	os.RemoveAll(path)
}
