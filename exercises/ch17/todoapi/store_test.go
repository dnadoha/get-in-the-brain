package todoapi

import (
	"errors"
	"sync"
	"testing"
)

func TestStoreCreate(t *testing.T) {
	s := NewStore()
	a := s.Create("Buy milk")
	b := s.Create("Walk dog")

	if a.ID != 1 || b.ID != 2 {
		t.Errorf("ids = %d, %d, want 1, 2", a.ID, b.ID)
	}
	if a.Title != "Buy milk" || a.Done {
		t.Errorf("got %+v, want title Buy milk and Done=false", a)
	}
}

func TestStoreList(t *testing.T) {
	s := NewStore()
	if got := s.List(); got == nil || len(got) != 0 {
		t.Errorf("empty store: List = %#v, want an empty non-nil slice", got)
	}

	s.Create("a")
	s.Create("b")
	s.Create("c")
	got := s.List()
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	for i, todo := range got {
		if todo.ID != i+1 {
			t.Errorf("List()[%d].ID = %d, want %d (ordered by id)", i, todo.ID, i+1)
		}
	}
}

func TestStoreGet(t *testing.T) {
	s := NewStore()
	created := s.Create("Read")

	got, err := s.Get(created.ID)
	if err != nil || got != created {
		t.Errorf("Get = (%+v, %v), want (%+v, nil)", got, err, created)
	}

	if _, err := s.Get(99); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(99) err = %v, want ErrNotFound", err)
	}
}

func TestStoreUpdate(t *testing.T) {
	s := NewStore()
	created := s.Create("Old")

	got, err := s.Update(created.ID, "New", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID || got.Title != "New" || !got.Done {
		t.Errorf("Update returned %+v", got)
	}

	again, _ := s.Get(created.ID)
	if again != got {
		t.Errorf("update was not saved: %+v", again)
	}

	if _, err := s.Update(99, "x", false); !errors.Is(err, ErrNotFound) {
		t.Errorf("Update(99) err = %v, want ErrNotFound", err)
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	created := s.Create("Temp")

	if err := s.Delete(created.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := s.Get(created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("after Delete, Get err = %v, want ErrNotFound", err)
	}
	if err := s.Delete(created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("second Delete err = %v, want ErrNotFound", err)
	}
}

func TestStoreIDsAreNotReused(t *testing.T) {
	s := NewStore()
	s.Create("one")
	two := s.Create("two")
	s.Delete(two.ID)

	three := s.Create("three")
	if three.ID != 3 {
		t.Errorf("id = %d, want 3 (ids must not be reused)", three.ID)
	}
}

func TestStoreConcurrent(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Create("x")
			s.List()
		}()
	}
	wg.Wait()

	todos := s.List()
	if len(todos) != 50 {
		t.Fatalf("len = %d, want 50", len(todos))
	}
	seen := make(map[int]bool)
	for _, todo := range todos {
		if seen[todo.ID] {
			t.Errorf("duplicate id %d", todo.ID)
		}
		seen[todo.ID] = true
	}
}
