package main

import (
	db "GoSafe/src/db"
	net "GoSafe/src/net"
	"context"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var databasePath string

	if res, ok := os.LookupEnv("DB_PATH"); ok {
		databasePath = res
	} else {
		databasePath = "./GoFind.db"
	}

	options := db.Options{
		Url:             databasePath,
		Test:            false,
		MaxOpenConn:     100,
		MaxIdleConn:     100,
		MaxConnLifetime: 0,
		Log:             log.Default(),
	}

	database := db.NewDatabase(options)
	err := database.Connect()
	if err != nil {
		panic(err)
	}

	server := net.NewServer(database, context.Background())
	err = server.Serve()

	if err != nil {
		log.Fatal(err)
		return
	}
}
