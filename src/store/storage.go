package Store

import (
	"GoSafe/src/db"
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/mattn/go-sqlite3"
)

type store interface {
	GetView(int) (map[string][]*Object, error)

	GetObject(string, int) (*Object, error)
	GetItem(int, int) (*Item, error)

	AddObject(*Object, int) error
	AddItem(*Item, int) error
	addBatchItems([]Item, int) error

	UpdateObject(*Object, string, int) error
	UpdateItem(*Item, int, int) error

	DeleteObject(string, int) error
	DeleteItem(int, int) error

	SearchItems(string, string, int) (map[int]Item, error)
}

type InternalError struct {
	message string
}
type FormatError struct {
	message string
}
type NotFoundError struct {
	message string
}

func (e *InternalError) Error() string {
	return e.message
}
func (e *FormatError) Error() string {
	return e.message
}
func (e *NotFoundError) Error() string {
	return e.message
}

type Store struct {
	database    *db.Database
	ctx         context.Context
	NotFound    *NotFoundError
	Internal    *InternalError
	FormatError *FormatError
}

type ViewItem struct {
	Name string `json:"name"`
	Ref  int    `json:"ref"`
}

type ViewObject struct {
	Name     string `json:"name"`
	Children []any  `json:"children"`
}

func NewStore(db *db.Database, ctx context.Context) *Store {
	return &Store{
		database: db,
		ctx:      ctx,
		NotFound: &NotFoundError{
			message: "request resource not found",
		},
		Internal: &InternalError{
			message: "internal error",
		},
		FormatError: &FormatError{
			message: "format error",
		},
	}
}

// retrieves all the tags for the given item
func (s *Store) readItemsTags(tx *sql.Tx, id int) ([]string, error) {
	res, err := tx.Query("SELECT tags.name FROM Items JOIN tags_items on Items.id = tags_items.item_id JOIN tags on tags.id = tags_items.tag_id WHERE Items.id = ?", id)
	if err != nil {
		return nil, err
	}
	tags := make([]string, 0)
	for res.Next() {
		var tag string
		err = res.Scan(&tag)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return tags, nil
			}
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (s *Store) readObjectTags(tx *sql.Tx, id int) ([]string, error) {
	res, err := tx.Query("SELECT tags.name FROM Objects JOIN tags_objects on Objects.id = tags_objects.object_id JOIN tags on tags.id = tags_objects.tag_id WHERE Objects.id = ?", id)
	if err != nil {
		return nil, err
	}

	tags := make([]string, 0)

	for res.Next() {
		var tag string

		err = res.Scan(&tag)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return tags, nil
			}
			return nil, err
		}

		tags = append(tags, tag)
	}
	return tags, nil
}

func (s *Store) createTag(tag string, tx *sql.Tx) (int64, error) {
	tag = strings.ToLower(tag)
	res := tx.QueryRow("SELECT id FROM Tags WHERE Name = ?", tag)
	var id int64
	err := res.Scan(&id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		res, err := tx.Exec("INSERT INTO Tags (name) VALUES (?)", tag)
		if err != nil {
			return -1, err
		}
		id, err = res.LastInsertId()
		if err != nil {
			return -1, err
		}
		return id, nil
	} else if err != nil {
		return -1, err
	}
	return id, nil
}

// Add the relevant tags to the item, creating them if necessary
func (s *Store) setItemTags(tx *sql.Tx, tags []string, ItemID int) error {
	for _, tag := range tags {
		id, err := s.createTag(tag, tx)
		if err != nil {
			return err
		}
		_, err = tx.Exec("INSERT INTO tags_items (tag_id, Item_id) VALUES (?, ?)", id, ItemID)
		if err != nil {
			if errors.As(err, &sqlite3.ErrConstraint) {
				continue
			}
			return err
		}
	}
	return nil
}

// Add the relevant tags to the object, creating them if necessary
func (s *Store) setObjectTags(tx *sql.Tx, tags []string, objectID int) error {
	for _, tag := range tags {
		id, err := s.createTag(tag, tx)
		if err != nil {
			return err
		}
		_, err = tx.Exec("INSERT INTO tags_objects (tag_id, Object_id) VALUES (?, ?)", id, objectID)
		if err != nil {
			if errors.As(err, &sqlite3.ErrConstraint) {
				continue
			}
			return err
		}
	}
	return nil
}

func (s *Store) GetView(userID int) (*ViewObject, error) {
	view := &ViewObject{
		Name:     "root",
		Children: make([]any, 0),
	}

	tx, err := s.database.DB.BeginTx(s.ctx, nil)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			log.Default().Println(err)
			if tx != nil {
				err := tx.Rollback()
				if err != nil {
					log.Default().Println(err)
				}
			}
		} else {
			err = tx.Commit()
			if err != nil {
				log.Default().Println(err)
			}
		}
	}()

	res, err := tx.Query(
		"SELECT Objects.name, Objects.id, Objects.owner_id, I.name, I.id FROM Objects LEFT JOIN Items I ON Objects.id = I.owned_by WHERE Objects.ownedBy_user = ? ORDER BY Objects.id",
		userID)

	if err != nil {
		return nil, err
	}

	var objName string
	var itemName sql.NullString
	var objID int
	var objOwner, itemID sql.NullInt64
	temp := -1
	root := view

	for res.Next() {
		err := res.Scan(&objName, &objID, &objOwner, &itemName, &itemID)
		if err != nil {
			return nil, err
		}

		if temp == -1 {
			temp = objID
			next := &ViewObject{
				Name:     objName,
				Children: make([]any, 0),
			}
			view = next
		}

		if temp != objID {
			temp = objID
			next := &ViewObject{
				Name:     objName,
				Children: make([]any, 0),
			}
			if !objOwner.Valid {
				root.Children = append(root.Children, next)
			} else {
				view.Children = append(view.Children, next)
			}
			view = next
		}

		if itemName.Valid {
			view.Children = append(view.Children, &ViewItem{
				Name: itemName.String,
				Ref:  int(itemID.Int64),
			})
		}

	}
	return root, nil
}

func (s *Store) AddObject(obj *Object, userID int) error {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return s.Internal
	}

	defer func() {
		if err != nil {
			log.Default().Println(err)
			if tx != nil {
				err := tx.Rollback()
				if err != nil {
					log.Default().Println(err)
				}
			}
		} else {
			err = tx.Commit()
			if err != nil {
				log.Default().Println(err)
			}
		}
	}()

	if obj.Name == "" {
		return s.Internal
	}

	res, err := tx.Query("SELECT id FROM Objects WHERE name = ? and ownedBy_user = ?", obj.Name, userID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return s.Internal
	}

	if res.Next() {
		return s.FormatError
	}

	var owner sql.NullInt64

	if obj.Container == "" {
		owner.Valid = false
	} else {
		owner.Valid = true
		row := tx.QueryRow("SELECT ID FROM Objects WHERE name = ? and ownedBy_user = ?", obj.Container, userID)
		err := row.Scan(&owner.Int64)
		if err != nil {
			return s.FormatError
		}
	}

	objRes, err := tx.Exec("INSERT INTO Objects (name, description, owner_id, ownedBy_user) VALUES (?,?,?,?)",
		obj.Name, obj.Description, owner, userID)

	if err != nil {
		return s.Internal
	}

	objID, err := objRes.LastInsertId()

	if err != nil {
		return s.Internal
	}

	err = s.setObjectTags(tx, obj.Tags, int(objID))

	if err != nil {
		return s.Internal
	}

	if obj.Items != nil {
		err = s.addBatchItems(tx, obj.Items, userID, int(objID))
		if err != nil {
			return s.Internal
		}
	}

	return nil
}

// TODO: error handling sucks
func (s *Store) AddItem(item *Item, userID int) error {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return s.Internal
	}
	defer tx.Commit()

	if item == nil {
		return s.Internal
	}

	var container sql.NullString

	if item.Container != "" {
		var id string
		res := tx.QueryRow("SELECT id FROM Objects WHERE id = ? AND ownedBy_user = ? ",
			item.Container, userID)

		err = res.Scan(&id)
		if err != nil {
			log.Default().Println(err)
			return s.Internal
		}
		container = sql.NullString{
			String: id,
			Valid:  true,
		}
	} else {
		container = sql.NullString{
			String: "",
			Valid:  false,
		}
	}

	res, err := tx.Exec("INSERT INTO Items (name, quantity, description, owned_by, ownedBy_user) VALUES (?,?,?,?,?)",
		item.Name, item.Quantity, item.Description, container, userID)

	if err != nil {
		_ = tx.Rollback()
		log.Default().Println(err)
		return s.Internal
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Default().Println(err)
		return s.Internal
	}

	err = s.setItemTags(tx, item.Tags, int(id))
	if err != nil {
		log.Default().Println(err)
		return s.Internal
	}

	return nil
}

func (s *Store) UpdateObject(object *Object, ObjectID string, userID int) error {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return s.Internal
	}

	defer func() {
		if err != nil {
			log.Default().Println(err)
			if tx != nil {
				err := tx.Rollback()
				if err != nil {
					log.Default().Println(err)
				}
			}
		} else {
			err = tx.Commit()
			if err != nil {
				log.Default().Println(err)
			}
		}
	}()

	var objID int
	row := tx.QueryRow("SELECT id FROM Objects WHERE ownedBy_user = ? and name = ?", userID, ObjectID)

	if err := row.Scan(&objID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.FormatError
		}
		s.Internal.message = err.Error()
		return s.Internal
	}

	if object == nil {
		return s.Internal
	}

	if ObjectID == "" {
		return s.FormatError
	}

	var container sql.NullInt64
	if object.Container != "" {
		res := tx.QueryRow("SELECT id FROM Objects WHERE name = ? AND ownedBy_user = ? ", object.Container, userID)
		err := res.Scan(&container.Int64)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return s.NotFound
			}
			s.Internal.message = err.Error()
			return s.Internal
		}
		container.Valid = true
	} else {
		container = sql.NullInt64{
			Int64: 0,
			Valid: false,
		}
	}

	_, err = tx.Exec("DELETE FROM Items WHERE owned_by = ? AND ownedBy_user = ? ", objID, userID)
	if err != nil {
		s.Internal.message = err.Error()
		return s.Internal
	}

	_, err = tx.Exec(
		"UPDATE Objects SET name = ?, description = ?, owner_id = ?, ownedBy_user = ? WHERE id = ? and ownedBy_user = ? ",
		object.Name, object.Description, container, userID, objID, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.NotFound
		}
		s.Internal.message = err.Error()
		return s.Internal
	}

	if object.Items != nil {
		if err := s.addBatchItems(tx, object.Items, userID, objID); err != nil {
			return err
		}
	}

	_, err = tx.Exec("DELETE FROM tags_objects WHERE object_id = ?", objID)
	if err != nil {
		s.Internal.message = err.Error()
		return s.Internal
	}

	err = s.setObjectTags(tx, object.Tags, objID)
	if err != nil {
		s.Internal.message = err.Error()
		return s.Internal
	}
	return nil
}

func (s *Store) UpdateItem(item *Item, itemID int, userID int) error {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return s.Internal
	}

	defer func() {
		if err != nil {
			log.Default().Println(err)
			if tx != nil {
				err := tx.Rollback()
				if err != nil {
					log.Default().Println(err)
				}
			}
		} else {
			err = tx.Commit()
			if err != nil {
				log.Default().Println(err)
			}
		}
	}()

	if item == nil {
		return s.Internal
	}

	var itemName string

	row := tx.QueryRow("SELECT name FROM ITEMS WHERE ownedBy_user = ? AND id = ?", userID, itemID)
	err = row.Scan(&itemName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.NotFound
		}
		return s.Internal
	}
	_, err = tx.Exec("UPDATE Items SET name = ?, quantity = ?, description = ?, owned_by = ? WHERE ownedBy_user = ? AND id =?",
		item.Name, item.Quantity, item.Description, item.Container, userID, itemID)

	if err != nil {
		return s.Internal
	}

	_, err = tx.Exec("DELETE FROM tags_items WHERE item_id = ?", itemID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	err = s.setItemTags(tx, item.Tags, itemID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteObject(objectID string, userID int) error {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		s.Internal.message = err.Error()
		return s.Internal
	}
	defer func() {
		err = tx.Commit()
		if err != nil {
			log.Default().Println(err)
		}
	}()
	_, err = tx.Exec("DELETE FROM Objects WHERE id = ? and ownedBy_user = ?", objectID, userID)
	if err != nil {
		log.Default().Println(err)
		_ = tx.Rollback()
		return s.Internal
	}

	return nil
}

func (s *Store) DeleteItem(itemID int, userID int) error {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)

	if err != nil {
		log.Default().Println(err)
		return s.Internal
	}

	defer func() {
		err = tx.Commit()
		if err != nil {
			log.Default().Println(err)
		}
	}()

	_, err = tx.Exec("DELETE FROM Items WHERE ownedBy_user = ? and id = ?", userID, itemID)
	if err != nil {
		_ = tx.Rollback()
		log.Default().Println(err)
		return s.Internal
	}
	return nil
}

func (s *Store) addBatchItems(tx *sql.Tx, items []Item, userID int, ownerID int) error {
	if items == nil {
		return s.FormatError
	}

	stmt, err := tx.Prepare(
		"INSERT INTO Items (name, quantity, description, owned_by, ownedBy_user) VALUES (?,?,?,?,?)")
	if err != nil {
		return s.Internal
	}

	defer stmt.Close()

	for _, item := range items {
		res, err := stmt.Exec(item.Name, item.Quantity, item.Description, ownerID, userID)
		if err != nil {
			return s.Internal
		}
		if id, err := res.LastInsertId(); err != nil {
			return s.Internal
		} else {
			item.Id = int(id)
			err = s.setItemTags(tx, item.Tags, int(id))
			if err != nil {
				return s.Internal
			}
		}
	}
	return nil
}

func (s *Store) GetObject(objectID string, userID int) (*Object, error) {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return nil, s.Internal
	}

	defer func() {
		if err != nil {
			log.Default().Println(err)
			if tx != nil {
				err := tx.Rollback()
				if err != nil {
					log.Default().Println(err)
				}
			}
		} else {
			err = tx.Commit()
			if err != nil {
				log.Default().Println(err)
			}
		}
	}()

	if objectID == "" {
		return nil, s.FormatError
	}

	res := tx.QueryRow(
		"SELECT id, description, created_at, owner_id FROM Objects WHERE name = ? and ownedBy_user = ?",
		objectID, userID)

	var obj Object
	var container sql.NullString

	obj.Name = objectID
	var objID int
	if err = res.Scan(&objID, &obj.Description, &obj.CreationDate, &container); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, s.NotFound
		}
		return nil, s.Internal
	}

	if container.Valid {
		obj.Container = container.String
	} else {
		obj.Container = ""
	}

	obj.Tags, err = s.readObjectTags(tx, objID)
	if err != nil {
		return nil, s.Internal
	}
	obj.Items = []Item{}
	itemsFound, err := tx.Query(
		"SELECT id, name, quantity, description FROM Items WHERE owned_by = ? and ownedBy_user = ?",
		objID, userID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, s.Internal
	} else if errors.Is(err, sql.ErrNoRows) {
		return &obj, nil
	}

	for itemsFound.Next() {
		var item Item
		err = itemsFound.Scan(&item.Id, &item.Name, &item.Quantity, &item.Description)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return &obj, nil
			}
			return nil, s.Internal
		}

		item.Tags, err = s.readItemsTags(tx, item.Id)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				item.Tags = []string{}
			}
			return nil, s.Internal
		}
		item.Container = objectID
		obj.Items = append(obj.Items, item)
	}
	return &obj, nil
}

func (s *Store) GetItem(itemID int, userID int) (*Item, error) {
	var item = new(Item)

	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return nil, s.Internal
	}

	defer func() {
		err := tx.Commit()
		if err != nil {
			log.Default().Println(err)
		}
	}()

	var container sql.NullInt64
	var description sql.NullString

	row := tx.QueryRow("SELECT name, description, quantity, owned_by FROM Items WHERE id = ? and ownedBy_user = ?", itemID, userID)

	if err = row.Scan(&item.Name, &description, &item.Quantity, &container); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.NotFound.message = err.Error()
			return nil, s.NotFound
		}
		return nil, s.Internal
	}

	item.Tags, err = s.readItemsTags(tx, item.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			item.Tags = []string{}
		} else {
			log.Default().Println(err)
			return nil, s.Internal
		}
	}

	if container.Valid {
		res := tx.QueryRow("SELECT name FROM Objects WHERE id = ?", container.Int64)
		if err = res.Scan(&item.Container); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.NotFound.message = err.Error()
				return nil, s.NotFound
			}
		}
	} else {
		item.Container = ""
	}

	if description.Valid {
		item.Description = description.String
	} else {
		item.Description = ""
	}

	return item, nil
}

// SearchItems TODO: implement
func (s *Store) SearchItems(param string, value string, userID int) (map[int]Item, error) {
	panic("Implement me")
	return nil, nil
}
