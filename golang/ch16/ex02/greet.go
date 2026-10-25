package greet

import (
	"fmt"
	"net/http"
)

func GreetHandler(w http.ResponseWriter, r *http.Request) {
	var name = r.URL.Query().Get("name")

	if name == "" {
		name = "World"
	}
	fmt.Fprintf(w, "Hello, %s!", name)

}
