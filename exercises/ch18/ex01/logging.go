package logging

import (
	"fmt"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

type statusRec struct {
	http.ResponseWriter
	status int
}

func (s *statusRec) WriteHeader(code int) {
	s.status = code

	s.ResponseWriter.WriteHeader(code)
}

func Logging(log func(string)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRec{
				w,
				http.StatusOK,
			}
			next.ServeHTTP(rec, r)
			log(fmt.Sprintf("%s %s %d", r.Method, r.URL.Path, rec.status))
		})
	}
}
