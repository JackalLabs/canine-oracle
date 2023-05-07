package server

import (
	"os"
	"path"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	nameKey     = []byte("name")
	intervalKey = []byte("interval")
)

func SetUpTestDB() (*leveldb.DB, error) {
	dbPath := path.Join(os.TempDir(), "goleveldb-testdb")
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}

	err = db.Put(nameKey, []byte("jklPrice"), nil)
	if err != nil {
		return nil, err
	}

	err = db.Put(intervalKey, []byte("10"), nil)
	if err != nil {
		return nil, err
	}

	return db, nil
}
