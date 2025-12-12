package db

import (
	cryptorand "crypto/rand"
	"io"
	"log"
	"os"
)

// TestInit Spawns a populated in-memory database for testing purposes
func TestInit() (*Database, error) {
	log.Default().Println("TestInit should only be used for testing")
	options := Options{
		Url:             "file:" + cryptorand.Text(),
		Test:            true,
		MaxIdleConn:     100,
		MaxOpenConn:     100,
		MaxConnLifetime: 0,
		Log:             log.Default(),
	}

	testingFolder := os.Getenv("TEST_FOLDER")
	if testingFolder == "" {
		panic("Missing TEST_FOLDER environment variable")
	}

	database := NewDatabase(options)

	err := database.Connect()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(testingFolder + "schema.sql")
	if err != nil {
		return nil, err
	}

	if schema, err := io.ReadAll(file); err == nil {
		sql := string(schema)
		_, err := database.DB.Exec(sql)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	insert, err := os.Open(testingFolder + "testState.sql")
	if err != nil {
		return nil, err
	}

	ins, err := io.ReadAll(insert)
	if err != nil {
		return nil, err
	}

	_, err = database.DB.Exec(string(ins))
	if err != nil {
		return nil, err
	}

	return database, nil
}
