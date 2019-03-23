package database

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type Users struct {
	db *DB
}

func (u Users) Exists(name string) (bool, error) {
	var count int

	return count > 0, u.db.do(func(tx *sql.Tx) error {
		return tx.QueryRow(
			`
			select count(*)
			from "users"
			where "name" = ? ;
			`, name).Scan(&count)
	})
}

func (u Users) Check(name, pass string) (bool, error) {
	var hash []byte

	err := u.db.do(func(tx *sql.Tx) error {
		return tx.QueryRow(
			`
			select "pass"
			from "users"
			where "name" = ? ;
			`, name).Scan(&hash)
	})

	if err != nil {
		return false, err
	}

	return bcrypt.CompareHashAndPassword(hash, []byte(pass)) == nil, nil
}

func (u Users) Add(name, pass string) error {
	hash, err := hashPassword(pass)
	if err != nil {
		return err
	}

	return u.db.do(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`
			insert into "users"
			( "name", "pass" )
			values
			( ?, ? ) ;
			`, name, hash)

		return err
	})
}

func (u Users) Update(name, pass string) error {
	hash, err := hashPassword(pass)
	if err != nil {
		return err
	}

	return u.db.do(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`
			update "users"
			set "pass" = ?
			where "name" = ? ;
			`, hash, name)

		return err
	})
}

func (u Users) Delete(name string) error {
	return u.db.do(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`
			delete from "users"
			where "name" = ? ;
			`, name)

		return err
	})
}

func (u Users) List() ([]string, error) {
	var list []string

	return list, u.db.do(func(tx *sql.Tx) error {
		rows, err := tx.Query(
			`
			select "name"
			from "users"
			order by "name" asc ;
			`)

		if err != nil {
			return err
		}

		defer rows.Close()

		var name string

		for rows.Next() {
			if err := rows.Scan(&name); err != nil {
				return err
			}

			list = append(list, name)
		}

		return nil
	})
}

func hashPassword(pass string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
}
