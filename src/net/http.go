package net

import (
	"GoSafe/src/db"
	Store "GoSafe/src/store"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/rs/cors"
)

type Server struct {
	ctx     context.Context
	storage *Store.Store
	auth    auth
}

type request struct {
	userID int
	Object *Store.Object `json:"object"`
	Item   *Store.Item   `json:"Item"`
}

// Returns a populated request, fields might be default initialized
func parseRequest(r *http.Request) (*request, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	req := new(request)
	if len(raw) <= 0 {
		return req, nil
	}

	err = json.Unmarshal(raw, &req)

	if err != nil {
		return nil, err
	}

	return req, nil
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var err error

	defer func() {
		if err != nil {
			log.Default().Println(err)
		}
	}()

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		if cookie, err := r.Cookie("sessionCookie"); err == nil {
			if ok, _ := s.auth.LoggedUserCheck(cookie.Value); ok {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	cookie, err := s.auth.Login(username, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Default().Println(err)
		return
	}

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	if username == "" || password == "" {
		http.Error(w, "username or password is empty", http.StatusBadRequest)
		return
	}

	err = s.auth.Register(username, password)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Default().Println(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	var err error

	defer func() {
		if err != nil {
			log.Default().Println(err)
		}
	}()

	username := r.PathValue("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("sessionCookie")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.auth.Logout(username, cookie.Value)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "sessionCookie",
		MaxAge: -1,
	})

	w.WriteHeader(http.StatusOK)
}

// itemOp http://<location>/item/{itemID} [GET | POST | PUT | DELETE]
func (s *Server) itemOp(w http.ResponseWriter, r *http.Request) {
	var err error

	defer func() {
		if err != nil {
			log.Default().Println(err)
		}
	}()

	req, err := s.requestValidation(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	itemID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil && r.Method != "PUT" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		item, err := s.storage.GetItem(itemID, req.userID)
		if err != nil {
			if errors.Is(err, s.storage.NotFound) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		data, err := json.Marshal(item)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		_, err = w.Write(data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		break
		//update existing resource
	case "POST":
		err := s.storage.UpdateItem(req.Item, itemID, req.userID)
		if err != nil {
			if errors.Is(err, s.storage.NotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		break
	case "DELETE":
		err := s.storage.DeleteItem(itemID, req.userID)
		if err != nil {
			var notFoundError *Store.NotFoundError
			if !errors.As(err, &notFoundError) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		break
		//create resource
	case "PUT":
		err := s.storage.AddItem(req.Item, req.userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		break
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// objectOp http://<location>/object/{objectID} [GET | POST | PUT | DELETE]
func (s *Server) objectOp(w http.ResponseWriter, r *http.Request) {
	req, err := s.requestValidation(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	objectID := r.PathValue("id")

	switch r.Method {
	case "GET":
		obj, err := s.storage.GetObject(objectID, req.userID)
		if err != nil {
			if errors.Is(err, s.storage.NotFound) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		data, err := json.Marshal(obj)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		_, err = w.Write(data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		break
	case "PUT":
		err := s.storage.AddObject(req.Object, req.userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		break
	case "DELETE":
		err := s.storage.DeleteObject(objectID, req.userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		break
	case "POST":
		err := s.storage.UpdateObject(req.Object, objectID, req.userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		break
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (s *Server) serveFront(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "./front/login.html")
		return
	}
	http.FileServer(http.Dir("./front")).ServeHTTP(w, r)
}

func (s *Server) view(w http.ResponseWriter, r *http.Request) {
	req, err := s.requestValidation(r)

	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	view, err := s.storage.GetView(req.userID)
	if err != nil {
		if errors.Is(err, s.storage.NotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(view)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// TODO: dead code
func (s *Server) searchItem(w http.ResponseWriter, r *http.Request) {
	req, err := s.requestValidation(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}
	param, value := r.PathValue("parameter"), r.PathValue("value")
	if param == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if value == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := s.storage.SearchItems(param, value, req.userID)
	if err != nil {
		if errors.Is(err, s.storage.NotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, s.storage.FormatError) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	data, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) Serve() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.serveFront)
	mux.HandleFunc("/register", s.register)
	mux.HandleFunc("/login", s.login)
	mux.HandleFunc("/logout/{username}", s.logout)
	mux.HandleFunc("/object/{id}", s.objectOp)
	mux.HandleFunc("/item/{id}", s.itemOp)
	mux.HandleFunc("/items", s.view)
	mux.HandleFunc("/item/search/{parameter}/{value}", s.searchItem)
	handler := cors.Default().Handler(mux)
	return http.ListenAndServe(":8080", handler)
}

// requestValidation checks the correct parsing of the http body request and if the cookie is valid. req will be populated with a valid request or nil
func (s *Server) requestValidation(r *http.Request) (*request, error) {
	cookie, err := r.Cookie("sessionCookie")

	if err != nil {
		return nil, err
	}

	ok, id := s.auth.LoggedUserCheck(cookie.Value)
	if !ok {
		//TODO: write a different error
		return nil, http.ErrNoCookie
	}

	req, err := parseRequest(r)
	if err != nil {
		return nil, err
	}

	req.userID = id

	if r.Method != "GET" && r.Method != "DELETE" {
		if req.Item == nil && req.Object == nil {
			return nil, errors.New("invalid request")
		}
	}

	return req, nil
}

func NewServer(database *db.Database, ctx context.Context) *Server {
	storage := Store.NewStore(database, ctx)
	auth := NewAuth(database.DB)
	return &Server{
		ctx:     ctx,
		storage: storage,
		auth:    auth,
	}
}
