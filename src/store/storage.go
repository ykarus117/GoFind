package Store

import (
	"GoSafe/src/db"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mattn/go-sqlite3"
)

type store interface {
	GetView(int) (*ViewObject, error)

	GetObject(string, int) (*Object, error)
	GetItem(int, int) (*Item, error)

	AddObject(*Object, int) error
	AddItem(*Item, int) error

	UpdateObject(*Object, string, int) error
	UpdateItem(*Item, int, int) error

	DeleteObject(string, int) error
	DeleteItem(int, int) error
}

type Store struct {
	database *db.Database
	ctx      context.Context
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

func (s *Store) addBatchItems(tx *sql.Tx, items []Item, userID int, ownerID int) error {
	if items == nil {
		return errors.New("items is nil")
	}

	stmt, err := tx.Prepare(
		"INSERT INTO Items (name, quantity, description, owned_by, ownedBy_user) VALUES (?,?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, item := range items {
		res, err := stmt.Exec(item.Name, item.Quantity, item.Description, ownerID, userID)
		if err != nil {
			return err
		}
		if id, err := res.LastInsertId(); err != nil {
			return err
		} else {
			item.Id = int(id)
			err = s.setItemTags(tx, item.Tags, int(id))
			if err != nil {
				return err
			}
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
		return nil, &InternalError{
			Err:     err,
			Message: fmt.Sprintf("Begin transaction (user %v)", userID),
		}
	}

	res, err := tx.Query(
		"SELECT Objects.name, Objects.id, Objects.owner_id, I.name, I.id FROM Objects LEFT JOIN Items I ON Objects.id = I.owned_by WHERE Objects.ownedBy_user = ? ORDER BY Objects.id",
		userID)

	if err != nil {
		tx.Rollback()
		return nil, &InternalError{
			Err:     err,
			Message: fmt.Sprintf("Main view select query (user: %v)", userID),
		}
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
			tx.Rollback()
			return nil, &InternalError{
				Err:     err,
				Message: fmt.Sprintf("Main view scan (user %v)", userID),
			}
		}

		if temp != objID || temp < 0 {
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

	res, err = tx.Query("SELECT name, id FROM ITEMS WHERE ownedBy_user = ? AND owned_by IS NULL", userID)

	if err != nil {
		tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return root, nil
		}
		return nil, err
	}

	for res.Next() {
		err = res.Scan(&itemName, &itemID)
		if err != nil {
			tx.Rollback()
			return nil, &InternalError{
				Err:     err,
				Message: fmt.Sprintf("Loose item view scan (user: %v)", userID),
			}
		}

		root.Children = append(root.Children, &ViewItem{
			Name: itemName.String,
			Ref:  int(itemID.Int64),
		})
	}

	tx.Commit()
	return root, nil
}

func (s *Store) AddObject(obj *Object, userID int) error {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("AddObject transaction (user %v)", userID),
		}
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
		tx.Commit()
	}()

	if obj.Name == "" {
		return &FormatError{
			Message: fmt.Sprintf("Name is required (user %v, object %v)", userID, obj),
		}
	}

	res, err := tx.Query("SELECT id FROM Objects WHERE name = ? and ownedBy_user = ?", obj.Name, userID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("AddObject select query (user %v)", userID),
		}
	}

	if res.Next() {
		return &FormatError{
			Message: fmt.Sprintf("Object already exists (user %v, object %v)", userID, obj.Name),
		}
	}

	var owner sql.NullInt64

	if obj.Container == "" {
		owner.Valid = false
	} else {
		owner.Valid = true
		row := tx.QueryRow("SELECT ID FROM Objects WHERE name = ? and ownedBy_user = ?", obj.Container, userID)
		err := row.Scan(&owner.Int64)
		if err != nil {
			return &InternalError{
				Err:     err,
				Message: fmt.Sprintf("AddObject owner select query (user %v, object %v)", userID, obj),
			}
		}
	}

	objRes, err := tx.Exec("INSERT INTO Objects (name, description, owner_id, ownedBy_user) VALUES (?,?,?,?)",
		obj.Name, obj.Description, owner, userID)

	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("AddObject update query (user %v, object %v)", userID, obj),
		}
	}

	objID, err := objRes.LastInsertId()

	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("AddObject new object id return query (user %v, object %v)", userID, obj),
		}
	}

	err = s.setObjectTags(tx, obj.Tags, int(objID))

	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("AddObject set tags (user %v, object %v)", userID, obj),
		}
	}

	if obj.Items != nil {
		err = s.addBatchItems(tx, obj.Items, userID, int(objID))
		if err != nil {
			return &InternalError{
				Err:     err,
				Message: fmt.Sprintf("AddObject add batch items (user %v object %v)", userID, obj),
			}
		}
	}

	return nil
}

func (s *Store) AddItem(item *Item, userID int) error {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("AddItem transaction (user %v)", userID),
		}
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
		tx.Commit()
	}()

	if item == nil {
		return &InternalError{
			Message: "AddItem nil item",
		}
	}

	var container sql.NullString

	if item.Container != "" {
		var id string
		res := tx.QueryRow("SELECT id FROM Objects WHERE name = ? AND ownedBy_user = ? ",
			item.Container, userID)

		err = res.Scan(&id)
		if err != nil {
			return &InternalError{
				Err:     err,
				Message: fmt.Sprintf("AddItem select container id (user %v, item %v)", userID, item),
			}
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
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("AddItem create (user %v, item %v)", userID, item),
		}
	}

	id, err := res.LastInsertId()
	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("created item id retrieve (user %v, item %v)", userID, item),
		}
	}

	err = s.setItemTags(tx, item.Tags, int(id))
	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("AddItem set tags (user %v, item %v)", userID, item),
		}
	}

	return nil
}

func (s *Store) UpdateObject(object *Object, ObjectID string, userID int) error {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("UpdateObject transaction (user %v, object %v)", userID, ObjectID),
		}
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
		tx.Commit()
	}()

	var objID int
	row := tx.QueryRow("SELECT id FROM Objects WHERE ownedBy_user = ? and name = ?", userID, ObjectID)

	if err := row.Scan(&objID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &NotFoundError{
				Err:     err,
				Message: fmt.Sprintf("UpdateObject Requested object (user %v, object %v) not found", userID, ObjectID),
			}
		}
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("UpdateObject select query (user %v, object %v)", userID, ObjectID),
		}
	}

	if object == nil {
		return &InternalError{
			Message: "UpdateObject nil object",
		}
	}

	if ObjectID == "" {
		return &FormatError{
			Message: "UpdateObject empty ObjectID",
		}
	}

	var container sql.NullInt64
	if object.Container != "" {
		res := tx.QueryRow("SELECT id FROM Objects WHERE name = ? AND ownedBy_user = ? ", object.Container, userID)
		err := res.Scan(&container.Int64)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return &NotFoundError{
					Err:     err,
					Message: fmt.Sprintf("UpdateObject requested object container not found (user %v, object %v)", userID, object.Container),
				}
			}
			return &InternalError{
				Err:     err,
				Message: fmt.Sprintf("UpdateObject container select query (user %v, object %v)", userID, object),
			}
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
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("UpdateObject delete objects' items (user %v, object %v)", userID, ObjectID),
		}
	}

	_, err = tx.Exec(
		"UPDATE Objects SET name = ?, description = ?, owner_id = ?, ownedBy_user = ? WHERE id = ? and ownedBy_user = ? ",
		object.Name, object.Description, container, userID, objID, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &NotFoundError{
				Err:     err,
				Message: fmt.Sprintf("UpdateObject (user %v, object %v)", userID, ObjectID),
			}
		}
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("UpdateObject update object (user %v, object %v)", userID, ObjectID),
		}
	}

	if object.Items != nil {
		if err := s.addBatchItems(tx, object.Items, userID, objID); err != nil {
			return &InternalError{
				Err:     err,
				Message: fmt.Sprintf("UpdateObject add batch items (user %v, object %v)", userID, object.Items),
			}
		}
	}

	_, err = tx.Exec("DELETE FROM tags_objects WHERE object_id = ?", objID)
	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("UpdateObject delete tags object (user %v, tags %v)", userID, object.Tags),
		}
	}

	err = s.setObjectTags(tx, object.Tags, objID)
	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("UpdateObject set tags object (user %v, tags %v)", userID, object.Tags),
		}
	}
	return nil
}

func (s *Store) UpdateItem(item *Item, itemID int, userID int) error {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("UpdateItem transaction (user %v, item %v)", userID, itemID),
		}
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
		tx.Commit()
	}()

	if item == nil {
		return &InternalError{
			Message: fmt.Sprintf("UpdateItem nil item"),
		}
	}

	var itemName string

	row := tx.QueryRow("SELECT name FROM Items WHERE ownedBy_user = ? AND id = ?", userID, itemID)
	err = row.Scan(&itemName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &NotFoundError{
				Err:     err,
				Message: fmt.Sprintf("UpdateItem item (user %v, item %v)", userID, itemID),
			}
		}
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("UpdateItem select query (user %v, item %v)", userID, itemID),
		}
	}

	var objectId sql.NullInt64
	row = tx.QueryRow("SELECT id FROM Objects WHERE ownedBy_user = ? AND name = ?", userID, item.Container)
	err = row.Scan(&objectId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &NotFoundError{
				Err:     err,
				Message: fmt.Sprintf("Object %v id not found (user %v)", item.Container, userID),
			}
		}
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("UpdateItem select object id (user %v, item %v)", userID, itemID),
		}
	}

	_, err = tx.Exec("UPDATE Items SET name = ?, quantity = ?, description = ?, owned_by = ? WHERE ownedBy_user = ? AND id =?",
		item.Name, item.Quantity, item.Description, objectId.Int64, userID, itemID)

	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("UpdateItem update item (user %v, item %v)", userID, itemID),
		}
	}

	_, err = tx.Exec("DELETE FROM tags_items WHERE item_id = ?", itemID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("UpdateItem delete tags item (user %v, item %v)", userID, itemID),
		}
	}

	err = s.setItemTags(tx, item.Tags, itemID)
	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("UpdateItem set tags item (user %v, item %v)", userID, itemID),
		}
	}

	return nil
}

func (s *Store) DeleteObject(objectID string, userID int) error {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("DeleteObject transaction (user %v, object %v)", userID, objectID),
		}
	}

	_, err = tx.Exec("DELETE FROM Objects WHERE name = ? AND ownedBy_user = ?", objectID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &NotFoundError{
				Err:     err,
				Message: fmt.Sprintf("DeleteObject Requested object (user %v, object %v)", userID, objectID),
			}
		}
		tx.Rollback()
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("DeleteObject delete object (user %v, object %v)", userID, objectID),
		}
	}
	tx.Commit()
	return nil
}

func (s *Store) DeleteItem(itemID int, userID int) error {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)

	if err != nil {
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("DeleteItem transaction (user %v, item %v)", userID, itemID),
		}
	}

	_, err = tx.Exec("DELETE FROM Items WHERE ownedBy_user = ? and id = ?", userID, itemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &NotFoundError{
				Err:     err,
				Message: fmt.Sprintf("DeleteItem item (user %v, item %v)", userID, itemID),
			}
		}
		tx.Rollback()
		return &InternalError{
			Err:     err,
			Message: fmt.Sprintf("DeleteItem delete item (user %v, item %v)", userID, itemID),
		}
	}
	tx.Commit()
	return nil
}

func (s *Store) GetObject(objectID string, userID int) (*Object, error) {
	tx, err := s.database.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return nil, &InternalError{
			Err:     err,
			Message: fmt.Sprintf("GetObject transaction (user %v, object %v)", userID, objectID),
		}
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
		tx.Commit()
	}()

	if objectID == "" {
		return nil, &FormatError{
			Message: fmt.Sprintf("ObjectId is required (user %v)", userID),
		}
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
			return nil, &NotFoundError{
				Err:     err,
				Message: fmt.Sprintf("GetObject (user %v, object %v)", userID, objectID),
			}
		}
		return nil, &InternalError{
			Err:     err,
			Message: fmt.Sprintf("GetObject select query (user %v, object %v)", userID, objectID),
		}
	}

	if container.Valid {
		obj.Container = container.String
	} else {
		obj.Container = ""
	}

	obj.Tags, err = s.readObjectTags(tx, objID)
	if err != nil {
		return nil, &InternalError{
			Err:     err,
			Message: fmt.Sprintf("GetObject read tags object (user %v, object %v)", userID, objectID),
		}
	}
	obj.Items = []Item{}
	itemsFound, err := tx.Query(
		"SELECT id, name, quantity, description FROM Items WHERE owned_by = ? and ownedBy_user = ?",
		objID, userID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, &InternalError{
			Err:     err,
			Message: fmt.Sprintf("GetObject read item (user %v, object %v)", userID, objectID),
		}
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
			return nil, &InternalError{
				Err:     err,
				Message: fmt.Sprintf("GetObject read item (user %v, object %v)", userID, objectID),
			}
		}

		item.Tags, err = s.readItemsTags(tx, item.Id)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				item.Tags = []string{}
			}
			return nil, &InternalError{
				Err:     err,
				Message: fmt.Sprintf("GetObject read item tags (user %v, item %v)", userID, item.Id),
			}
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
		return nil, &InternalError{
			Err:     err,
			Message: fmt.Sprintf("GetItem transaction (user %v, item %v)", userID, itemID),
		}
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
		tx.Commit()
	}()

	// just for convenience
	item.Id = itemID

	var container sql.NullInt64
	var description sql.NullString

	row := tx.QueryRow("SELECT name, description, quantity, owned_by FROM Items WHERE id = ? and ownedBy_user = ?", itemID, userID)

	if err = row.Scan(&item.Name, &description, &item.Quantity, &container); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NotFoundError{
				Message: fmt.Sprintf("GetItem item (user %v, item %v)", userID, itemID),
			}
		}
		return nil, &InternalError{
			Err:     err,
			Message: fmt.Sprintf("GetItem select query (user %v, item %v)", userID, itemID),
		}
	}

	item.Tags, err = s.readItemsTags(tx, item.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			item.Tags = []string{}
		} else {
			return nil, &InternalError{
				Err:     err,
				Message: fmt.Sprintf("GetItem read item tags (user %v item %v)", userID, itemID),
			}
		}
	}

	if container.Valid {
		res := tx.QueryRow("SELECT name FROM Objects WHERE id = ?", container.Int64)
		if err = res.Scan(&item.Container); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, &NotFoundError{
					Message: fmt.Sprintf("GetItem (user %v item %v)", userID, itemID),
				}
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
