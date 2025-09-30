package Store

import (
	"GoSafe/src/db"
	"context"
	"log"
	"net/http"
)

type Server struct {
	ctx        context.Context
	storage    store
	auth       auth
	permFolder string
}

func NewServer(database *db.Database, auth *Auth, ctx context.Context, permFolder string) *Server {
	storage := NewStore(database, ctx)

	http.HandleFunc("POST /item/{id}", func(h http.ResponseWriter, r *http.Request) {
		log.Println(storage.AddItem(h, r))
	})
	http.HandleFunc("DELETE /item/{id}", func(h http.ResponseWriter, r *http.Request) {
		log.Println(storage.DeleteItem(h, r))
	})
	http.HandleFunc("PUT /item", func(h http.ResponseWriter, r *http.Request) {
		log.Println(storage.UpdateItem(h, r))
	})
	http.HandleFunc("GET /item/{id}", func(h http.ResponseWriter, r *http.Request) {
		panic("implement me")
	})

	http.HandleFunc("POST /object", func(h http.ResponseWriter, r *http.Request) {
		log.Println(storage.AddObject(h, r))
	})
	http.HandleFunc("DELETE /object/{id}", func(h http.ResponseWriter, r *http.Request) {
		log.Println(storage.DeleteObject(h, r))
	})
	http.HandleFunc("PUT /object", func(h http.ResponseWriter, r *http.Request) {
		log.Println(storage.UpdateObject(h, r))
	})
	http.HandleFunc("GET /object/{id}", func(h http.ResponseWriter, r *http.Request) {
		panic("implement me")
	})

	return &Server{
		ctx:        ctx,
		storage:    storage,
		auth:       auth,
		permFolder: permFolder,
	}
}
