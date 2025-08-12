package db

import (
	"database/sql"
	"log"
	"time"
)

type Database struct {
	url         string
	DB          *sql.DB
	log         *log.Logger
	maxOpenConn int
	maxIdleConn int
	maxConnLife int
}
type Options struct {
	Url             string
	Test            bool
	MaxOpenConn     int
	MaxIdleConn     int
	MaxConnLifetime int
	Log             *log.Logger
}

func NewDatabase(Options Options) *Database {
	var url string

	// In-memory database for testing, using a :memory: database will incur in errors as transactions open with .Begin() seem to open a new connection
	// and sqlite supports only one connection per :memory: database
	if Options.Test {
		url = Options.Url + "?mode=memory&cache=shared"
	} else {
		url = Options.Url + "?journal=WAL&timeout=2500"
	}

	return &Database{
		url:         url,
		DB:          nil,
		log:         Options.Log,
		maxOpenConn: Options.MaxOpenConn,
		maxIdleConn: Options.MaxIdleConn,
		maxConnLife: Options.MaxConnLifetime,
	}
}

func (db *Database) Connect() error {
	d, err := sql.Open("sqlite3", db.url)

	if err != nil {
		return err
	}

	err = d.Ping()
	if err != nil {
		return err
	}

	db.DB = d
	db.DB.SetMaxIdleConns(db.maxIdleConn)
	db.DB.SetMaxOpenConns(db.maxOpenConn)
	db.DB.SetConnMaxLifetime(time.Second * time.Duration(db.maxConnLife))
	return nil
}
