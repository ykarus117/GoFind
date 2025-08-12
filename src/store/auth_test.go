package Store

import (
	"GoSafe/src/db"
	"database/sql"
	"fmt"
	"maps"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var loggedInUsers = map[string]session{
	// testing user for registration and login
	"user0": {
		sessionID: "userSessionID",
		begin:     time.Now(),
		end:       time.Now().AddDate(0, 0, 1),
	},
	// valid user
	"user1": {
		sessionID: "user1SessionID",
		begin:     time.Now(),
		end:       time.Now().AddDate(0, 0, 1),
	},
	// valid user, expired session
	"user2": {
		sessionID: "user2SessionID",
		begin:     time.Now(),
		end:       time.Now().AddDate(0, 0, -1),
	},
	// valid user
	"user3": {
		sessionID: "user3SessionID",
		begin:     time.Now(),
		end:       time.Now().AddDate(0, 0, 3),
	},
}

func TestAuth_Login(t *testing.T) {
	testDB, err := db.TestInit()

	if err != nil {
		panic(err)
	}

	type fields struct {
		db          *sql.DB
		loggedUsers map[string]session
		lock        sync.RWMutex
	}
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantErr       bool
		expectedUsers map[string]session
	}{
		{
			name: "Existing user login",
			fields: fields{
				db:          testDB.DB,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username: "user0",
				password: "foobar123",
			},
			wantErr:       false,
			expectedUsers: loggedInUsers,
		},
		{
			name: "Non-existing user login",
			fields: fields{
				db:          testDB.DB,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username: "NoUser",
				password: "foobar",
			},
			wantErr:       true,
			expectedUsers: loggedInUsers,
		},
		{
			name: "Existing user wrong password",
			fields: fields{
				db:          testDB.DB,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username: "user0",
				password: "wrongPassword",
			},
			wantErr:       true,
			expectedUsers: loggedInUsers,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Auth{
				db:          tt.fields.db,
				loggedUsers: tt.fields.loggedUsers,
				lock:        tt.fields.lock,
			}

			_, err := a.Login(tt.args.username, tt.args.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !maps.Equal(a.loggedUsers, tt.expectedUsers) {
				t.Errorf("Login() loggedUsers = %v, want %v", a.loggedUsers, tt.expectedUsers)
				return
			}
		})
	}
}

func TestAuth_Logout(t *testing.T) {

	type fields struct {
		db          *sql.DB
		loggedUsers map[string]session
		lock        sync.RWMutex
	}
	type args struct {
		username  string
		sessionID string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantErr       bool
		expectedUsers map[string]session
	}{
		{
			name: "Existing and valid user logout",
			fields: fields{
				db:          nil,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username:  "user3",
				sessionID: loggedInUsers["user3"].sessionID,
			},
			wantErr: false,
			expectedUsers: func() map[string]session {
				expected := make(map[string]session)
				maps.Copy(expected, loggedInUsers)
				delete(expected, "user1")
				return expected
			}(),
		},
		{
			name: "Non-existing user logout",
			fields: fields{
				db:          nil,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username:  "invalidUser",
				sessionID: "invalidUserTestingID",
			},
			wantErr:       true,
			expectedUsers: loggedInUsers,
		},
		{
			name: "Existing user wrong sessionID",
			fields: fields{
				db:          nil,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username:  "user1",
				sessionID: "invalidID",
			},
			wantErr:       true,
			expectedUsers: loggedInUsers,
		},
		{
			name: "Valid User expired logout",
			fields: fields{
				db:          nil,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username:  "user2",
				sessionID: loggedInUsers["user2"].sessionID,
			},
			wantErr:       false,
			expectedUsers: loggedInUsers,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Auth{
				db:          tt.fields.db,
				loggedUsers: tt.fields.loggedUsers,
				lock:        tt.fields.lock,
			}
			if err := a.Logout(tt.args.username, tt.args.sessionID); (err != nil) != tt.wantErr {
				t.Errorf("Logout() error = %v, wantErr %v", err, tt.wantErr)

				if !maps.Equal(a.loggedUsers, tt.expectedUsers) {
					t.Errorf("Logout() loggedUsers = %v, want %v", a.loggedUsers, tt.expectedUsers)
				}
			}
		})
	}
}

func TestAuth_Register(t *testing.T) {
	testDB, err := db.TestInit()
	if err != nil {
		panic(err)
	}

	type fields struct {
		db          *sql.DB
		loggedUsers map[string]session
		lock        sync.RWMutex
	}
	type args struct {
		username string
		password string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "User Creation",
			fields: fields{
				db:          testDB.DB,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username: fmt.Sprintf("testingUser-%d-%d-%d", rand.Int()%100, rand.Int()%100, rand.Int()%100),
				password: fmt.Sprintf("foobar_%d", rand.Int()),
			},
			wantErr: false,
		},
		{
			name: "Existing user creation",
			fields: fields{
				db:          testDB.DB,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username: "user0",
				password: "wrongPassword",
			},
			wantErr: true,
		},
		{
			name: "Empty user creation",
			fields: fields{
				db:          testDB.DB,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username: "",
				password: "foobar",
			},
			wantErr: true,
		},
		{
			name: "Empty password creation",
			fields: fields{
				db:          testDB.DB,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username: "user",
				password: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid password format (only lowercase letters)",
			fields: fields{
				db:          testDB.DB,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username: "user",
				password: "testingpassword",
			},
			wantErr: true,
		},
		{
			name: "Invalid password format (only numbers)",
			fields: fields{
				db:          testDB.DB,
				loggedUsers: make(map[string]session),
				lock:        sync.RWMutex{},
			},
			args: args{
				username: "user",
				password: "12345678",
			},
			wantErr: true,
		},
		{
			name: "Invalid password format (too short)",
			fields: fields{
				db:          testDB.DB,
				loggedUsers: make(map[string]session),
				lock:        sync.RWMutex{},
			},
			args: args{
				username: "user",
				password: "test123",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Auth{
				db:          tt.fields.db,
				loggedUsers: tt.fields.loggedUsers,
				lock:        tt.fields.lock,
			}
			if err := a.Register(tt.args.username, tt.args.password); (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuth_LoggedUserCheck(t *testing.T) {
	type fields struct {
		db          *sql.DB
		loggedUsers map[string]session
		lock        sync.RWMutex
	}
	type args struct {
		username  string
		sessionID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Logged in user",
			fields: fields{
				db:          nil,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username:  "user1",
				sessionID: loggedInUsers["user1"].sessionID,
			},
			want: true,
		},
		{
			name: "Expired session user",
			fields: fields{
				db:          nil,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username:  "user2",
				sessionID: loggedInUsers["user2"].sessionID,
			},
			want: false,
		},
		{
			name: "Not logged in User",
			fields: fields{
				db:          nil,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username:  "user10",
				sessionID: "missingUser",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Auth{
				db:          tt.fields.db,
				loggedUsers: tt.fields.loggedUsers,
				lock:        tt.fields.lock,
			}
			if got, _ := a.LoggedUserCheck(tt.args.username, tt.args.sessionID); got != tt.want {
				t.Errorf("IsUserLogged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuth_Delete(t *testing.T) {
	testDB, err := db.TestInit()
	if err != nil {
		panic(err)
	}

	type fields struct {
		db          *sql.DB
		loggedUsers map[string]session
		lock        sync.RWMutex
	}
	type args struct {
		username  string
		sessionID string
		password  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Existing and logged in user",
			fields: fields{
				db:          testDB.DB,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username:  "user0",
				sessionID: loggedInUsers["user0"].sessionID,
				password:  "foobar123",
			},
			wantErr: false,
		},
		{
			name: "Non existing user",
			fields: fields{
				db:          testDB.DB,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username:  "nonExistingUser",
				sessionID: "invalidSessionID",
				password:  "foobar",
			},
			wantErr: true,
		},
		{
			name: "Existing, not logged in user",
			fields: fields{
				db: testDB.DB,
				loggedUsers: map[string]session{
					"user1": session{
						sessionID: "user1",
						begin:     time.Now(),
						end:       time.Now().Add(time.Hour),
					},
					"user2": session{
						sessionID: "user2",
						begin:     time.Now(),
						end:       time.Now().Add(time.Hour),
					},
				},
				lock: sync.RWMutex{},
			},
			args: args{
				username:  "user0",
				sessionID: "userSessionID",
				password:  "foobar123",
			},
			wantErr: true,
		},
		{
			name: "Existing, logged in user with wrong password",
			fields: fields{
				db:          testDB.DB,
				loggedUsers: loggedInUsers,
				lock:        sync.RWMutex{},
			},
			args: args{
				username:  "user0",
				sessionID: loggedInUsers["user0"].sessionID,
				password:  "wrongPassword",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Auth{
				db:          tt.fields.db,
				loggedUsers: tt.fields.loggedUsers,
				lock:        tt.fields.lock,
			}
			if err := a.Delete(tt.args.username, tt.args.sessionID, tt.args.password); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
