package server

import (
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	nameKey     = []byte("name")
	apiKey      = []byte("api")
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

	err = db.Put(apiKey, []byte("https://api-osmosis.imperator.co/tokens/v2/price/jkl"), nil)
	if err != nil {
		return nil, err
	}

	err = db.Put(intervalKey, []byte("10"), nil)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestPrepareOracleData(t *testing.T) {
	testDb, err := SetUpTestDB()
	if err != nil {
		t.Error(err)
	}

	apiLink, err := testDb.Get(apiKey, nil)
	if err != nil {
		t.Error("Can't get API Link", err)
	}

	res, err := http.Get(string(apiLink))
	if err != nil {
		t.Error("GET Request Error: ", err)
	}

	defer res.Body.Close()

	_, err = PrepareOracleData(res)
	if err != nil {
		t.Error(err)
	}
}
