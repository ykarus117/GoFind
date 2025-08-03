package store

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
)

type auth interface {
	Register(username string, password string) error
	Login(username string, password string) (int, error)
	Logout(username string, cookie http.Cookie)
}

type Auth struct {
	db          *sql.DB
	loggedUsers map[string]int
	lock        sync.Mutex
}

func NewAuth(ctx context.Context, db *sql.DB) *Auth {
	return &Auth{
		db:          db,
		loggedUsers: make(map[string]int),
		lock:        sync.Mutex{},
	}
}

func (a *Auth) Login(username string, password string) (int, error) {
	if a.db.Ping() != nil {
		return 0, fmt.Errorf("db ping fail")
	}

	if _, ok := a.loggedUsers[username]; ok {
		return -1, nil
	}

	var salt string
	var hashpwd string

	tx, err := a.db.Begin()
	if err != nil {
		return 0, err
	}

	userHash := sha256.New()
	pwdHash := sha256.New()

	_, err = userHash.Write([]byte(username))
	if err != nil {
		return 0, err
	}

	res, err := tx.Query("SELECT salt, pwd FROM Users WHERE username = ?", userHash.Sum(nil))
	if err != nil {
		return 0, err
	}

	err = res.Scan(hashpwd, salt)

	if err != nil {
		return 0, err
	}
	_, err = pwdHash.Write([]byte(password))
	if err != nil {
		return 0, err
	}

	if subtle.ConstantTimeCompare([]byte(hashpwd), pwdHash.Sum(nil)) != 1 {
		return 0, fmt.Errorf("invalid password")
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	sessionToken := rand.Int()
	a.loggedUsers[username] = sessionToken

	return sessionToken, nil
}

func (a *Auth) Register(username string, password string) error {
	if a.db.Ping() != nil {
		return fmt.Errorf("db ping fail")
	}
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}

	userHash := sha256.New()
	pwdHash := sha256.New()
	salt := make([]byte, 16)

	_, err = cryptorand.Read(salt)
	if err != nil {
		return err
	}

	_, err = userHash.Write([]byte(username))
	if err != nil {
		return err
	}

	_, err = pwdHash.Write([]byte(password))
	if err != nil {
		return err
	}

	_, err = pwdHash.Write(salt)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO Users (username, pwd, salt) VALUES (?, ?, ?)", username, password, salt)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (a *Auth) Logout(username string, cookie http.Cookie) {
	a.lock.Lock()
	defer a.lock.Unlock()
}
