package Store

import (
	"GoSafe/src/db"
	"context"
	"reflect"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestStore_AddItem(t *testing.T) {
	testDB, err := db.TestInit()
	if err != nil {
		panic(err)
	}
	type fields struct {
		database *db.Database
		ctx      context.Context
	}
	type args struct {
		item   *Item
		userID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Add new Item",
			fields: fields{
				database: testDB,
				ctx:      context.Background(),
			},
			args: args{
				item: &Item{
					Name:        "ItemTest1",
					Description: "I am ItemTest1",
					Quantity:    20,
					Container:   "",
					Tags: []string{
						"tag1",
						"testTag",
					},
				},
				userID: 0,
			},
			wantErr: false,
		},
		{
			name: "Add item with existing name",
			fields: fields{
				database: testDB,
				ctx:      context.Background(),
			},
			args: args{
				item: &Item{
					Name:        "Item One",
					Description: "I am Item One",
					Quantity:    15,
					Container:   "",
					Tags: []string{
						"tag1",
					},
				},
				userID: 0,
			},
			wantErr: false,
		},
		{
			name: "Null item",
			fields: fields{
				database: testDB,
				ctx:      context.Background(),
			},
			args: args{
				item:   nil,
				userID: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStore(tt.fields.database, tt.fields.ctx)
			if err := s.AddItem(tt.args.item, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("AddItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStore_AddObject(t *testing.T) {
	testDB, err := db.TestInit()
	if err != nil {
		panic(err)
	}
	type fields struct {
		database *db.Database
		ctx      context.Context
	}
	type args struct {
		obj    *Object
		userID int
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantErr        bool
		getObjectError bool
		want           Object
	}{
		// TODO: Add test cases.
		{
			name: "Add new valid Object",
			fields: fields{
				database: testDB,
				ctx:      context.Background(),
			},
			args: args{
				obj: &Object{
					Name:        "ObjectTest1",
					Description: "I am ObjectTest1",
					Tags: []string{
						"tag1",
						"testTag",
					},
					Items: []Item{},
				},
				userID: 0,
			},
			wantErr:        false,
			getObjectError: false,
			want: Object{
				Name:        "ObjectTest1",
				Description: "I am ObjectTest1",
				Tags: []string{
					"tag1",
					// tags ara enforced lower case
					"testtag",
				},
				Items: []Item{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStore(tt.fields.database, tt.fields.ctx)

			if err := s.AddObject(tt.args.obj, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("AddObject() error = %v, wantErr %v", err, tt.wantErr)
			}

			readObj, err := s.GetObject(tt.args.obj.Name, tt.args.userID)

			if (err != nil) != tt.getObjectError {
				t.Errorf("GetObject() error = %v, getObjectError %v", err, tt.getObjectError)
			}

			readObj.CreationDate = time.Now().UTC()
			tt.want.CreationDate = readObj.CreationDate

			if readObj != nil && !reflect.DeepEqual(*readObj, tt.want) {
				t.Errorf("AddObject() got:\n %v \n want:\n %v", *readObj, tt.want)
			}

		})
	}
}

func TestStore_DeleteItem(t *testing.T) {
	type fields struct {
		database *db.Database
		ctx      context.Context
	}
	type args struct {
		itemID int
		userID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				database: tt.fields.database,
				ctx:      tt.fields.ctx,
			}
			if err := s.DeleteItem(tt.args.itemID, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("DeleteItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStore_DeleteObject(t *testing.T) {
	type fields struct {
		database *db.Database
		ctx      context.Context
	}
	type args struct {
		objectID string
		userID   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				database: tt.fields.database,
				ctx:      tt.fields.ctx,
			}
			if err := s.DeleteObject(tt.args.objectID, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("DeleteObject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStore_GetItem(t *testing.T) {
	testDB, err := db.TestInit()
	if err != nil {
		t.Errorf("%s", err)
	}
	type fields struct {
		database *db.Database
		ctx      context.Context
	}
	type args struct {
		itemID int
		userID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Item
		wantErr bool
	}{
		{
			name: "Existing Item (Item id: 0)",
			fields: fields{
				database: testDB,
				ctx:      context.Background(),
			},
			args: args{
				itemID: 0,
				userID: 0,
			},
			want: &Item{
				Name:        "Item One",
				Description: "",
				Quantity:    1,
				Container:   "obj1",
				Tags:        []string{"tag2", "tag3"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStore(tt.fields.database, context.Background())
			got, err := s.GetItem(tt.args.itemID, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetItem() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStore_GetObject(t *testing.T) {
	testDB, err := db.TestInit()
	if err != nil {
		t.Fatal(err)
	}
	type fields struct {
		database *db.Database
		ctx      context.Context
	}
	type args struct {
		objectID string
		userID   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Object
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Existing Object",
			fields: fields{
				database: testDB,
				ctx:      context.Background(),
			},
			args: args{
				objectID: "obj1",
				userID:   0,
			},
			want: &Object{
				Name:        "obj1",
				Description: "Im object 1!",
				Container:   "",
				Tags:        []string{},
				Items: []Item{{
					Name:      "Item One",
					Quantity:  1,
					Container: "obj1",
					Tags: []string{
						"tag2",
						"tag3",
					},
				}},
			},
			wantErr: false,
		},
		{
			name: "Non-existing Object",
			fields: fields{
				database: testDB,
				ctx:      context.Background(),
			},
			args: args{
				objectID: "obj65",
				userID:   0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Existing object, user but wrong match (user does not own the object)",
			fields: fields{
				database: testDB,
				ctx:      context.Background(),
			},
			args: args{
				objectID: "obj9",
				userID:   0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Existing object non existing user",
			fields: fields{
				database: testDB,
				ctx:      context.Background(),
			},
			args: args{
				objectID: "obj1",
				userID:   1,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStore(tt.fields.database, context.Background())
			got, err := s.GetObject(tt.args.objectID, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil {
				tt.want.CreationDate = time.Now().UTC()
				got.CreationDate = tt.want.CreationDate
			}

			if got != nil && !reflect.DeepEqual(*got, *tt.want) {
				t.Errorf("GetObject() Got = %v \n Want = %v", got, tt.want)
			}
		})
	}
}

func TestStore_GetView(t *testing.T) {

	testDB, err := db.TestInit()
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		database *db.Database
		ctx      context.Context
	}
	type args struct {
		userID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Existing View",
			fields: fields{
				database: testDB,
				ctx:      context.Background(),
			},
			args: args{
				userID: 0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				database: tt.fields.database,
				ctx:      tt.fields.ctx,
			}
			got, err := s.GetView(tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetView() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			println(got)
		})
	}
}

func TestStore_UpdateItem(t *testing.T) {
	testDB, err := db.TestInit()
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		database *db.Database
		ctx      context.Context
	}
	type args struct {
		item   *Item
		itemID int
		userID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    Item
	}{
		// TODO: Add test cases.
		{
			name: "Valid field update, existing Item",
			fields: fields{
				database: testDB,
				ctx:      context.Background(),
			},
			args: args{
				item: &Item{
					Name:        "Item One",
					Quantity:    1,
					Description: "Im item 1!",
					Container:   "obj1",
					Tags:        []string{"tag1"},
				},
				itemID: 0,
				userID: 0,
			},
			wantErr: false,
			want: Item{
				Name:        "Item One",
				Quantity:    1,
				Description: "Im item 1!",
				Container:   "obj1",
				Tags:        []string{"tag1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				database: tt.fields.database,
				ctx:      tt.fields.ctx,
			}
			if err := s.UpdateItem(tt.args.item, tt.args.itemID, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("UpdateItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			var item *Item

			if item, err = s.GetItem(tt.args.itemID, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("GetItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.want, *item) {
				t.Errorf("Want item %s, got item %s", tt.want, item)
			}
		})
	}
}

func TestStore_UpdateObject(t *testing.T) {

	testDB, err := db.TestInit()
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		database *db.Database
		ctx      context.Context
	}
	type args struct {
		object   *Object
		objectID string
		userID   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    Object
	}{
		{
			name: "Valid update",
			fields: fields{
				database: testDB,
				ctx:      context.Background(),
			},
			args: args{
				objectID: "obj1",
				object: &Object{
					Name:        "Object Two",
					Description: "I've become object 2",
					Container:   "obj2",
					Items: []Item{
						{
							Name:        "Item One",
							Quantity:    1,
							Description: "Im item 1!",
							Container:   "Object Two",
							Tags:        []string{"tag1"},
						},
					},
					Tags: []string{"tag1"},
				},
				userID: 0,
			},
			wantErr: false,
			want: Object{
				Name:        "Object Two",
				Description: "I've become object 2",
				Container:   "obj2",
				Items: []Item{
					{
						Name:        "Item One",
						Quantity:    1,
						Description: "Im item 1!",
						Container:   "Object Two",
						Tags:        []string{"tag1"},
					},
				},
				Tags: []string{"tag1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStore(tt.fields.database, tt.fields.ctx)
			if err := s.UpdateObject(tt.args.object, tt.args.objectID, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("UpdateObject() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if res, err := s.GetObject(tt.args.object.Name, tt.args.userID); err != nil {
					t.Errorf("GetObject() error = %v", err)
				} else {
					res.CreationDate = time.Now().UTC()
					tt.want.CreationDate = res.CreationDate
					if !reflect.DeepEqual(*res, tt.want) {
						t.Errorf("UpdateObject() got = %v, want = %v", res, tt.want)
					}
				}
			}
		})
	}
}
