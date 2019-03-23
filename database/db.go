package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func Open(filename string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	db := DB{conn: conn}
	return &db, db.init()
}

func (d *DB) do(fn func(*sql.Tx) error) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DB) init() error {
	return d.do(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`
			create table if not exists "users" (
				id   integer        primary key autoincrement ,
				name varchar ( 32 ) unique ,
				pass blob           not null
			) ;

			create table if not exists "shares" (
				id   integer        primary key autoincrement ,
				name varchar ( 32 ) unique
			) ;

			create table if not exists "user_has_shares" (
				uid  integer ,
				sid  integer ,
				primary key ( uid, sid )
			) ;
			`)

		return err
	})
}

func (d *DB) Users() Users {
	return Users{db: d}
}
