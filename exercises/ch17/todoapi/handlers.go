package todoapi

import (
	"encoding/json"
	"net/http"
	"strconv"
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
	done := r.URL.Query().Get("done")
	todos := a.store.List()
	switch done {
	case "true":
		completedTodos := make([]Todo, 0)
		for _, todo := range todos {
			if todo.Done {
				completedTodos = append(completedTodos, todo)
			}
		}
		writeJSON(w, http.StatusOK, completedTodos)
		return
	case "false":
		incompleteTodos := make([]Todo, 0)
		for _, todo := range todos {
			if !todo.Done {
				incompleteTodos = append(incompleteTodos, todo)
			}
		}
		writeJSON(w, http.StatusOK, incompleteTodos)
		return
	case "":
		writeJSON(w, http.StatusOK, todos)
		return
	default:
		writeError(w, http.StatusBadRequest, "done must be true or false")
	}
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
func (a *API) getTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	todo, err := a.store.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}

	writeJSON(w, http.StatusOK, todo)
}
func (a *API) updateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var todo Todo
	err = json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if strings.TrimSpace(todo.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	todo, err = a.store.Update(id, todo.Title, true)
	if err != nil {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}

	writeJSON(w, http.StatusOK, todo)
}
func (a *API) deleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = a.store.Delete(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
