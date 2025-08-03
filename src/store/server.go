package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type store interface {
	Add(w http.ResponseWriter, r *http.Request) error
	addObject(req request, tx *sql.Tx) error
	addItem(req request, tx *sql.Tx) error

	Update(w http.ResponseWriter, r *http.Request) error
	updateObject(req request, tx *sql.Tx) error
	updateItem(req request, tx *sql.Tx) error

	Delete(w http.ResponseWriter, r *http.Request) error
	deleteObject(req request, tx *sql.Tx) error
	deleteItem(req request, tx *sql.Tx) error

	Close() error
}

type Store struct {
	sessionUser        string
	databaseConnection *sql.DB
	ctx                context.Context
}

type request struct {
	ElementType string `json:"element_type"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Quantity    int    `json:"quantity"`
	Tag         string `json:"tag"`
	Container   string `json:"container"`
}

func parseRequest(r *http.Request) (*request, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	req := new(request)
	err = json.Unmarshal(raw, &req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func NewStore(sessionUser string, db *sql.DB, ctx context.Context) *Store {
	return &Store{
		sessionUser:        sessionUser,
		databaseConnection: db,
		ctx:                ctx,
	}
}

func (s *Store) Add(w http.ResponseWriter, r *http.Request) error {
	err := s.databaseConnection.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	tx, err := s.databaseConnection.BeginTx(s.ctx, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	req, err := parseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	if req.ElementType == "" {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("element_type is required")
	} else {
		if req.ElementType == "Item" {
			err = s.addItem(*req, tx)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return err
			}
		}

		if req.ElementType == "Object" {
			return s.addObject(*req, tx)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return errors.New("element_type not supported")
		}
	}
}

func (s *Store) addObject(req request, tx *sql.Tx) error {
	res, err := tx.Query("SELECT id FROM Objects WHERE id = ?", req.Name)
	if err != nil {
		return err
	}
	if res.Next() {
		return errors.New("object name must be unique")
	}

	_, err = tx.Exec("INSERT INTO Objects (id, name, type, owner_id) VALUES (?,?,?,?)", req.Name, req.ElementType, req.Description, req.Container)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

func (s *Store) addItem(req request, tx *sql.Tx) error {
	_, err := tx.Exec("INSERT INTO Items (name, tags, quantity, owned_by) VALUES (?,?,?,?)", req.Name, req.Tag, req.Description, req.Container)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

func (s *Store) Update(w http.ResponseWriter, r *http.Request) error {
	//TODO implement me
	panic("implement me")
}

func (s *Store) updateObject(req request, tx *sql.Tx) error {
	//TODO implement me
	panic("implement me")
}

func (s *Store) updateItem(req request, tx *sql.Tx) error {
	//TODO implement me
	panic("implement me")
}

func (s *Store) Delete(w http.ResponseWriter, r *http.Request) error {
	//TODO implement me
	panic("implement me")
}

func (s *Store) deleteObject(req request, tx *sql.Tx) error {
	//TODO implement me
	panic("implement me")
}

func (s *Store) deleteItem(req request, tx *sql.Tx) error {
	//TODO implement me
	panic("implement me")
}

func (s *Store) Close() error {
	//TODO implement me
	panic("implement me")
}
