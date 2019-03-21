package tinycloud

import (
	"encoding/json"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

type Users map[string]string

func (u Users) Save(filename string) error {
	os.MkdirAll(filepath.Dir(filename), 0700)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	return enc.Encode(u)
}

func (u *Users) Load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	return json.NewDecoder(f).Decode(u)
}

func (u Users) Put(name, pass string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u[name] = string(hash)
	return nil
}

func (u Users) Exists(name string) bool {
	_, ok := u[name]
	return ok
}

func (u Users) Check(name, pass string) bool {
	hash, ok := u[name]
	if !ok {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass)) == nil
}
