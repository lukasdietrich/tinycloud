package tinycloud

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

type ctxUser struct{}

type Config struct {
	Users  Users
	Folder string
	Realm  string
}

func New(config *Config) http.Handler {
	handler := webdav.Handler{
		FileSystem: fs{afero.NewBasePathFs(afero.NewOsFs(), config.Folder)},
		LockSystem: webdav.NewMemLS(),
	}

	for name, _ := range config.Users {
		os.MkdirAll(filepath.Join(config.Folder, name), 0700)
	}

	if config.Realm == "" {
		config.Realm = "Restricted Access"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || !config.Users.Check(user, pass) {
			w.Header().Set(
				"WWW-Authenticate",
				"Basic realm="+strconv.Quote(config.Realm),
			)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), ctxUser{}, user))
		w2 := &writer{ResponseWriter: w, time: time.Now()}

		handler.ServeHTTP(w2, r)

		log.Printf("%10s (%03d) %5dµs %s",
			r.Method,
			w2.code,
			time.Now().Sub(w2.time)/time.Microsecond,
			r.RequestURI,
		)
	})
}

type writer struct {
	code int
	time time.Time
	http.ResponseWriter
}

func (w *writer) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}
