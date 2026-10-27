package logging

import (
	"fmt"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func Logging(log func(string)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			log(fmt.Sprintf("%s %s %d", r.Method, r.URL.Path, r.Response.StatusCode))
		})
	}
}
