package webdav

import (
	"context"
	"net/http"
	"strconv"

	"github.com/lukasdietrich/tinycloud/database"
)

type ctxUser struct{}

func withBasicAuth(users database.Users, realm string, next http.Handler) http.Handler {
	var (
		realmKey   = "WWW-Authenticate"
		realmValue = "Basic realm=" + strconv.Quote(realm)

		unauthorizedCode  = http.StatusUnauthorized
		unauthorizedText  = []byte(http.StatusText(unauthorizedCode))
		internalErrorCode = http.StatusInternalServerError
		internalErrorText = []byte(http.StatusText(internalErrorCode))
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if ok {
			correct, err := users.Check(user, pass)

			if err != nil {
				w.WriteHeader(internalErrorCode)
				w.Write(internalErrorText)

				return
			}

			if correct {
				ctx := context.WithValue(r.Context(), ctxUser{}, user)
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}
		}

		w.Header().Set(realmKey, realmValue)
		w.WriteHeader(unauthorizedCode)
		w.Write(unauthorizedText)
	})
}
