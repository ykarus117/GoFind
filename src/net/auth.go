package net

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type auth interface {
	LoggedUserCheck(sessionIDs string) (bool, int)
	Register(username string, password string) error
	Login(username string, password string) (*http.Cookie, error)
	Logout(username string, sessionID string) error
	Delete(username string, sessionID string, password string) error
	passwordCheck(username string, password string) bool
}

type session struct {
	username string
	userID   int
	begin    time.Time
	end      time.Time
}

type Auth struct {
	db          *sql.DB
	loggedUsers map[string]session
	lock        sync.RWMutex
}

func NewAuth(db *sql.DB) *Auth {
	return &Auth{
		db:          db,
		loggedUsers: make(map[string]session),
		lock:        sync.RWMutex{},
	}
}

// LoggedUserCheck returns true if and only if the user is logged in, presents a valid sessionID and the session is not expired
// returns the user unique ID or -1
func (a *Auth) LoggedUserCheck(sessionID string) (bool, int) {
	a.lock.RLock()
	userSession, ok := a.loggedUsers[sessionID]
	a.lock.RUnlock()

	if !ok {
		return false, -1
	}

	if time.Until(userSession.end) <= 0 {
		a.lock.Lock()
		delete(a.loggedUsers, sessionID)
		a.lock.Unlock()
		return false, -1
	}
	return true, userSession.userID
}

func (a *Auth) passwordCheck(username string, password string) bool {
	var salt []byte
	var hashPwd []byte
	pwdHash := sha256.New()

	res := a.db.QueryRow("SELECT salt, pwd FROM Users WHERE username = ?", username)

	err := res.Scan(&salt, &hashPwd)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		log.Println(err)
		return false
	}

	_, err = pwdHash.Write([]byte(password))

	if err != nil {
		log.Println(err)
		return false
	}

	_, err = pwdHash.Write(salt)
	if err != nil {
		log.Println(err)
		return false
	}

	if subtle.ConstantTimeCompare(hashPwd, pwdHash.Sum(nil)) != 1 {
		return false
	}

	return true
}

func (a *Auth) Delete(username string, sessionID string, password string) error {
	if a.passwordCheck(username, password) {
		a.lock.Lock()
		//necessary? can't hurt I guess
		if a.loggedUsers[sessionID].username == username {
			delete(a.loggedUsers, username)
			a.lock.Unlock()

			tx, err := a.db.Begin()
			if err != nil {
				return err
			}
			_, err = tx.Exec("DELETE FROM Users WHERE username in ( SELECT username FROM Users WHERE username = ? LIMIT 1)", username)
			if err != nil {
				err := tx.Rollback()
				if err != nil {
					return err
				}
			}
			err = tx.Commit()
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("not in logged users")
}

func (a *Auth) Login(username string, password string) (*http.Cookie, error) {
	if !a.passwordCheck(username, password) {
		return nil, errors.New("invalid password")
	}

	res := a.db.QueryRow("SELECT ID FROM Users where username = ?", username)
	var id int

	if err := res.Scan(&id); err != nil {
		return nil, errors.New("ugh, how?")
	}

	sessionID := cryptorand.Text()
	a.lock.Lock()
	a.loggedUsers[sessionID] = session{
		username: username,
		userID:   id,
		begin:    time.Now(),
		end:      time.Now().AddDate(0, 3, 0),
	}

	a.lock.Unlock()

	//TODO: export cookie name globally
	cookie := &http.Cookie{
		Value:    sessionID,
		Name:     "sessionCookie",
		HttpOnly: true,
		Secure:   true,
		Expires:  a.loggedUsers[username].end,
	}

	return cookie, nil
}

func (a *Auth) Register(username string, password string) error {
	res := a.db.QueryRow("SELECT username FROM Users WHERE username = ?", username)
	var retrievedUsername string
	err := res.Scan(&retrievedUsername)

	// scan nil return signals that a row was returned
	if err == nil {
		return errors.New("user already exists")
	}

	if len(username) < 3 {
		return errors.New("username too short. Minimum 3 characters")
	}

	if len(password) < 8 {
		return errors.New("password too short. Minimum 8 characters")
	}

	pwdHash := sha256.New()
	salt := make([]byte, 16)

	_, err = cryptorand.Read(salt)
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

	tx, err := a.db.Begin()
	if err != nil {
		return err
	}

	userRow, err := tx.Exec("INSERT INTO Users (username, pwd, salt) VALUES (?, ?, ?)", username, pwdHash.Sum(nil), salt)

	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		return err
	}

	insertId, err := userRow.LastInsertId()

	if err != nil {
		return err
	}

	Objectview := fmt.Sprintf("CREATE VIEW user%d_objects AS SELECT * FROM Objects WHERE ownedBy_user = %d", insertId, insertId)
	Itemview := fmt.Sprintf("CREATE VIEW user%d_items AS SELECT * FROM ITEMS WHERE ownedBy_user = %d", insertId, insertId)

	_, err = tx.Exec(Objectview)
	if err != nil {
		return err
	}

	_, err = tx.Exec(Itemview)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (a *Auth) Logout(username string, sessionID string) error {
	a.lock.RLock()
	_, ok := a.loggedUsers[sessionID]
	a.lock.RUnlock()

	if !ok {
		return errors.New("user not found")
	}

	a.lock.Lock()
	defer a.lock.Unlock()
	delete(a.loggedUsers, sessionID)

	return nil
}
