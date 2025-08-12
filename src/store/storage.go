package Store

import (
	"GoSafe/src/Item"
	"GoSafe/src/Object"
	"GoSafe/src/db"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type store interface {
	AddObject(w http.ResponseWriter, r *http.Request) error
	AddItem(w http.ResponseWriter, r *http.Request) error
	addBatchItems([]Item.Item, int) error

	UpdateObject(w http.ResponseWriter, r *http.Request) error
	UpdateItem(w http.ResponseWriter, r *http.Request) error

	DeleteObject(w http.ResponseWriter, r *http.Request) error
	DeleteItem(w http.ResponseWriter, r *http.Request) error

	requestValidation(r *http.Request) (*request, error)
}

type Store struct {
	a        *Auth
	database *db.Database
	ctx      context.Context
}

type request struct {
	Object *Object.Object `json:"object"`
	Item   *Item.Item     `json:"Item"`
	userid int
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

func (s *Store) addObject(obj *Object.Object, userID int) error {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return err
	}

	res, err := tx.Query("SELECT id FROM Objects WHERE id = ?", obj.Name)
	if err != nil {
		return err
	}

	if res.Next() {
		return errors.New("objects name must be unique")
	}

	tags := strings.Join(obj.Tags, ",")

	_, err = s.database.DB.Exec("INSERT INTO Objects (id, name, type, owner_id, ownedBy_user) VALUES (?,?,?,?,?)", obj.Name, obj.Name, tags, obj.Container, userID)

	if err != nil {
		err = tx.Rollback()
		if err != nil {
			return err
		}
		return err
	}

	if obj.Items != nil {
		err = s.addBatchItems(obj.Items, userID)
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
	}
	return nil
}

func NewStore(db *db.Database, ctx context.Context) *Store {
	return &Store{
		database: db,
		a:        NewAuth(db.DB),
		ctx:      ctx,
	}
}

// RequestValidation checks the correct parsing of the http body request and if the cookie is valid. req will be populated with a valid request or nil
func (s *Store) requestValidation(r *http.Request) (*request, error) {
	cookie, err := r.Cookie("sessionCookie")

	if err != nil {
		return nil, err
	}

	if len(cookie.Unparsed) <= 0 {
		return nil, errors.New("invalid cookie")
	}

	user := cookie.Unparsed[0]
	ok, id := s.a.LoggedUserCheck(user, cookie.Value)
	if !ok {
		return nil, errors.New("invalid cookie")
	}

	req, err := parseRequest(r)
	if err != nil {
		return nil, err
	}

	req.userid = id
	return req, nil
}

func (s *Store) AddObject(w http.ResponseWriter, r *http.Request) error {
	req, err := s.requestValidation(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	if req.Object == nil {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("object is required")
	}

	err = s.addObject(req.Object, req.userid)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}
	return nil
}

func (s *Store) AddItem(w http.ResponseWriter, r *http.Request) error {
	req, err := s.requestValidation(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	item := req.Item

	if item == nil {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("item is required")
	}

	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	var tags string

	if len(item.Tags) > 0 {
		tags = strings.Join(item.Tags, ",")
	}

	_, err = tx.Exec("INSERT INTO Items (name, tags, quantity, owned_by) VALUES (?,?,?,?)", item.Name, tags, item.Description, item.Container)

	if err != nil {
		err = tx.Rollback()
		if err != nil {
			return err
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return nil
}

func (s *Store) UpdateObject(w http.ResponseWriter, r *http.Request) error {
	req, err := parseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	obj := req.Object
	if obj == nil {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("item is required")
	}

	//if the object does not exist it is created
	if res := s.database.DB.QueryRow("SELECT id FROM Objects WHERE id = ?", obj.Name); res.Scan() != nil {
		err = s.addObject(obj, req.userid)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
	} else {
		tx, err := s.database.DB.BeginTx(s.ctx, nil)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		_, err = tx.Exec("UPDATE Objects SET name = ?, description = ?, owner_id = ?, ownedBy_user = ?  WHERE id = ?", obj.Name, obj.Description, obj.Container, req.userid, req.userid)

		if err != nil {
			err = tx.Rollback()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return err
			}
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		err = tx.Commit()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
	}

	return nil
}

func (s *Store) UpdateItem(w http.ResponseWriter, r *http.Request) error {
	//TODO implement me
	panic("implement me")
}

func (s *Store) DeleteObject(w http.ResponseWriter, r *http.Request) error {
	//TODO implement me
	panic("implement me")
}

func (s *Store) DeleteItem(w http.ResponseWriter, r *http.Request) error {
	//TODO implement me
	panic("implement me")
}

func (s *Store) addBatchItems(items []Item.Item, userID int) error {
	if items == nil {
		return errors.New("items must not be nil")
	}

	stmt, err := s.database.DB.Prepare("INSERT INTO Items (name, tags, quantity, ownedBy_user, owned_by) VALUES (?,?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, item := range items {
		tags := strings.Join(item.Tags, ",")
		_, err := stmt.Exec(item.Name, tags, item.Quantity, item.Container, userID)
		if err != nil {
			return err
		}
	}
	return nil
}
