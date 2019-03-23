package storage

import (
	"os"
	"path"

	"github.com/spf13/afero"
)

type Base string

const (
	Users  = Base("users")
	Shares = Base("shares")
)

const (
	fileMode = os.FileMode(0700)
)

type Storage struct {
	afero.Fs
}

func New(basePath string) (*Storage, error) {
	s := Storage{Fs: afero.NewBasePathFs(afero.NewOsFs(), basePath)}
	return &s, s.init()
}

func (s *Storage) init() error {
	for _, base := range [...]Base{Users, Shares} {
		if err := s.MkdirAll(string(base), fileMode); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) Resolve(base Base, name string, file string) string {
	return path.Join(string(base), name, file)
}
