package recoverer

import (
	"encoding/json"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Header().Add("Content-Type", "application/json")

					json.NewEncoder(w).Encode(map[string]string{
			"error": "internal server error",
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
