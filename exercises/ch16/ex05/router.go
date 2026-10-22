package router

import (
	"fmt"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("GET /hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")

		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, "Hello, %s!", name)
	})

	return mux
}
