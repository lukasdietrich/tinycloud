package webdav

import (
	"log"
	"net/http"
	"time"
)

func withLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%10s : %-48s  in %3dµs",
			r.Method,
			r.RequestURI,
			time.Since(t0)/time.Microsecond,
		)
	})
}
