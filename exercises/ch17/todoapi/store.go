package todoapi

import (
	"errors"
	"sort"
	"sync"
)

var ErrNotFound = errors.New("todo not found")

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type Store struct {
	mu     sync.Mutex
	nextID int
	todos  map[int]Todo
}

func NewStore() *Store {
	return &Store{nextID: 1, todos: make(map[int]Todo)}
}

// Create adds a todo with the next id and Done set to false.
func (s *Store) Create(title string) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo := Todo{ID: s.nextID, Title: title, Done: false}
	s.todos[s.nextID] = todo
	s.nextID++
	return todo
}

// List returns all todos ordered by ID. It never returns nil.
func (s *Store) List() []Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	todos := make([]Todo, 0, s.nextID-1)
	for _, todo := range s.todos {
		todos = append(todos, todo)
	}

	sort.Slice(todos, func(i, j int) bool {
		return todos[i].ID < todos[j].ID
	})
	return todos
}

// Get returns the todo with this id, or ErrNotFound.
func (s *Store) Get(id int) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, ok := s.todos[id]
	if ok {
		return todo, nil
	}
	return Todo{}, ErrNotFound
}

// Update changes title and done of an existing todo, or returns ErrNotFound.
func (s *Store) Update(id int, title string, done bool) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, ok := s.todos[id]
	if ok {
		todo.Title = title
		todo.Done = done
		s.todos[id] = todo
		return todo, nil
	}
	return Todo{}, ErrNotFound
}

// Delete removes a todo, or returns ErrNotFound.
func (s *Store) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.todos[id]
	if ok {
		delete(s.todos, id)
		return nil
	}
	return ErrNotFound
}
