package db

import (
	"database/sql"
	_ "embed"
	"log"
	"os"
	"strings"
	"time"
)

//go:embed schema.sql
var embeddedSchema string

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
		url = "file:" + Options.Url + "?cache=shared&mode=memory"
	} else {
		// If no database exist create one
		if _, err := os.Stat(Options.Url); err != nil {
			file, err := os.OpenFile("GoFind.db", os.O_CREATE, 0755)
			if err != nil {
				panic(err)
			}
			file.Close()
		}
		// Remember to build with 'sqlite_foreign_keys' to enable fk for all connections
		url = "file:" + Options.Url + "?journal=WAL&timeout=2500?"
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

	res := d.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='Users' ")

	var tmp string

	//Database has no schema, create it
	if err := res.Scan(&tmp); err != nil {
		db.log.Println("Setting up database schema...")
		lines := strings.Split(embeddedSchema, ";")
		for _, line := range lines {
			_, err = d.Exec(line)
			if err != nil {
				return err
			}
		}
	}

	db.DB = d
	db.DB.SetMaxIdleConns(db.maxIdleConn)
	db.DB.SetMaxOpenConns(db.maxOpenConn)
	db.DB.SetConnMaxLifetime(time.Second * time.Duration(db.maxConnLife))
	return nil
}
