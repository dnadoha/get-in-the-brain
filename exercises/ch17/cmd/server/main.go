package main

import (
	"log"
	"net/http"

	"todoapi"
)

func main() {
	api := todoapi.NewAPI(todoapi.NewStore())
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", api.Routes()))
}
