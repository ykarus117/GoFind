package main

import (
	"GoSafe/src/Store"
	"GoSafe/src/db"
	"context"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	database := db.NewDatabase(db.Options{
		Url:             "./testingDatabase/organizer.db",
		Test:            true,
		MaxOpenConn:     100,
		MaxIdleConn:     100,
		MaxConnLifetime: 0,
	})

	err := database.Connect()
	if err != nil {
		panic(err)
	}

	storage := Store.NewStore(database, context.Background())

	http.HandleFunc("GET /item/{id}", func(h http.ResponseWriter, r *http.Request) {
		log.Println(storage.AddItem(h, r))
	})
	http.HandleFunc("PUT /item", func(h http.ResponseWriter, r *http.Request) {
		log.Println(storage.UpdateItem(h, r))
	})
	http.HandleFunc("DELETE /item/{id}", func(h http.ResponseWriter, r *http.Request) {
		log.Println(storage.DeleteItem(h, r))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
