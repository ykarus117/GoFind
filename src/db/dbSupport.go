package db

import (
	cryptorand "crypto/rand"
	_ "embed"
	"log"
	"strings"
)

//go:embed testState.sql
var testState string

// TestInit Spawns a populated in-memory database for testing purposes
func TestInit() (*Database, error) {
	log.Default().Println("TestInit should only be used for testing")

	options := Options{
		Url:             cryptorand.Text(),
		Test:            true,
		MaxIdleConn:     100,
		MaxOpenConn:     100,
		MaxConnLifetime: 0,
		Log:             log.Default(),
	}

	database := NewDatabase(options)

	err := database.Connect()
	if err != nil {
		return nil, err
	}

	for _, row := range strings.Split(string(testState), ";") {
		_, err2 := database.DB.Exec(row)
		if err2 != nil {
			return nil, err2
		}
	}

	return database, nil
}
