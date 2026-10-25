package todoapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

type API struct {
	store *Store
}

func NewAPI(store *Store) *API {
	return &API{store: store}
}

func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", a.listTodos)
	mux.HandleFunc("POST /todos", a.createTodo)
	mux.HandleFunc("GET /todos/{id}", a.getTodo)
	mux.HandleFunc("PUT /todos/{id}", a.updateTodo)
	mux.HandleFunc("DELETE /todos/{id}", a.deleteTodo)
	return mux
}

func notImplemented(w http.ResponseWriter) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (a *API) listTodos(w http.ResponseWriter, r *http.Request) {
	todos := a.store.List()
	writeJSON(w, http.StatusOK, todos)
}

type createTodo struct {
	Title string `json:"title"`
}

func (a *API) createTodo(w http.ResponseWriter, r *http.Request) {
	var todo createTodo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	todo.Title = strings.TrimSpace(todo.Title)
	if todo.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	newTodo := a.store.Create(todo.Title)
	writeJSON(w, http.StatusCreated, newTodo)
}
func (a *API) getTodo(w http.ResponseWriter, r *http.Request)    { notImplemented(w) }
func (a *API) updateTodo(w http.ResponseWriter, r *http.Request) { notImplemented(w) }
func (a *API) deleteTodo(w http.ResponseWriter, r *http.Request) { notImplemented(w) }
