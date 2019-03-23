package webdav

import (
	"net/http"

	"golang.org/x/net/webdav"

	"github.com/lukasdietrich/tinycloud/database"
	"github.com/lukasdietrich/tinycloud/storage"
)

type Config struct {
	Realm    string
	Database *database.DB
	Storage  *storage.Storage
}

func New(config *Config) http.Handler {
	var (
		realm   = config.Realm
		handler http.Handler
	)

	if realm == "" {
		realm = "Restricted Access"
	}

	handler = &webdav.Handler{
		FileSystem: fs{config.Storage},
		LockSystem: webdav.NewMemLS(),
	}

	handler = withLogger(handler)
	handler = withBasicAuth(config.Database.Users(), realm, handler)

	return handler
}
