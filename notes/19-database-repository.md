# Chapter 19: Database and the Repository Pattern

## What you will learn

- How `database/sql` works
- The repository pattern: hide storage behind an interface
- A service layer for business rules
- Scanning rows safely
- Transactions
- How to test all of this **without** a real database

## Key ideas

### database/sql in short

`database/sql` is the standard API. You also need a **driver** for your database
(PostgreSQL, MySQL, SQLite). The driver is imported only for its side effect.

```go
import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver
)

db, err := sql.Open("pgx", dsn)
if err != nil { /* ... */ }
defer db.Close()
```

`*sql.DB` is a **connection pool**. Create it once and share it. Do not open one per request.

The main calls:

| Call | Use it for |
|------|-----------|
| `db.ExecContext` | `INSERT`, `UPDATE`, `DELETE` (no rows back) |
| `db.QueryRowContext` | one row, then `.Scan(...)` |
| `db.QueryContext` | many rows, then loop with `rows.Next()` |
| `db.BeginTx` | start a transaction |

Rules that prevent bugs:

- Use **placeholders** (`$1` in PostgreSQL, `?` in MySQL and SQLite). Never build SQL with string concatenation. That is how SQL injection happens.
- Always `defer rows.Close()` and check `rows.Err()` after the loop.
- Always use the `...Context` versions and pass the request context.
- `sql.ErrNoRows` means "no row found". Turn it into your own `ErrNotFound`.

### Repository pattern

```text
handler  ->  service  ->  repository (interface)  ->  PostgreSQL / memory / fake
HTTP         business     data access
             rules
```

- The **repository** knows how to store and load data. It has no business rules.
- The **service** has the rules (validation, who can do what). It depends on the repository **interface**.
- In tests you replace the real repository with an in-memory one.

> [!TIP]
> Define the interface where it is **used** (in the service package), and keep it small.
> Only list the methods the service really calls.

### Reference: a PostgreSQL repository

This code needs a real database and the `pgx` driver, so the exercises below do not run it.
It shows how the in-memory code you write maps to real SQL.

```go
type PostgresRepo struct {
	db *sql.DB
}

func (r *PostgresRepo) Create(ctx context.Context, u *User) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (email, name) VALUES ($1, $2) RETURNING id`,
		u.Email, u.Name,
	).Scan(&u.ID)
	if err != nil {
		// a unique violation on email should become ErrDuplicateEmail
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *PostgresRepo) GetByID(ctx context.Context, id int64) (User, error) {
	var u User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, name FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}
```

---

## Exercises

### Exercise 1: An in-memory repository

Implement `MemoryRepo` so it satisfies `UserRepository`.

Rules:

- IDs start at `1` and go up by one. `Create` sets `u.ID` on the struct you pass in.
- `Create` fails with `ErrDuplicateEmail` if another user has the same email. Ignore upper or lower case.
- `GetByID`, `GetByEmail` and `Delete` return `ErrNotFound` when nothing matches. `GetByEmail` also ignores case.
- `List` returns users ordered by ID, and an empty slice when there are none.
- Every method first checks the context. If `ctx.Err()` is not `nil`, return it.
- The repo must be safe for many goroutines.

**Starter** `ch19/ex01/repo.go`

```go
package repo

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrNotFound       = errors.New("user not found")
	ErrDuplicateEmail = errors.New("email already in use")
)

type User struct {
	ID    int64
	Email string
	Name  string
}

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	List(ctx context.Context) ([]User, error)
	Delete(ctx context.Context, id int64) error
}

type MemoryRepo struct {
	mu     sync.Mutex
	nextID int64
	users  map[int64]User
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{nextID: 1, users: make(map[int64]User)}
}

// TODO: implement Create, GetByID, GetByEmail, List and Delete
```

**Test** `ch19/ex01/repo_test.go`

```go
package repo

import (
	"context"
	"errors"
	"testing"
)

// This line fails to compile until MemoryRepo has all the methods.
var _ UserRepository = (*MemoryRepo)(nil)

func TestCreateAndGet(t *testing.T) {
	ctx := context.Background()
	r := NewMemoryRepo()

	u := &User{Email: "ana@example.com", Name: "Ana"}
	if err := r.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.ID != 1 {
		t.Errorf("ID = %d, want 1 (Create must set it)", u.ID)
	}

	got, err := r.GetByID(ctx, 1)
	if err != nil || got != *u {
		t.Errorf("GetByID = (%+v, %v), want (%+v, nil)", got, err, *u)
	}

	got, err = r.GetByEmail(ctx, "ANA@Example.com")
	if err != nil || got.ID != 1 {
		t.Errorf("GetByEmail (other case) = (%+v, %v), want user 1", got, err)
	}
}

func TestIDsIncrease(t *testing.T) {
	ctx := context.Background()
	r := NewMemoryRepo()
	a := &User{Email: "a@x.com", Name: "A"}
	b := &User{Email: "b@x.com", Name: "B"}
	r.Create(ctx, a)
	r.Create(ctx, b)
	if a.ID != 1 || b.ID != 2 {
		t.Errorf("ids = %d, %d, want 1, 2", a.ID, b.ID)
	}
}

func TestDuplicateEmail(t *testing.T) {
	ctx := context.Background()
	r := NewMemoryRepo()
	r.Create(ctx, &User{Email: "ana@example.com", Name: "Ana"})

	err := r.Create(ctx, &User{Email: "ANA@example.com", Name: "Other"})
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Errorf("err = %v, want ErrDuplicateEmail", err)
	}
	if list, _ := r.List(ctx); len(list) != 1 {
		t.Errorf("a failed Create must not add a user, have %d", len(list))
	}
}

func TestNotFound(t *testing.T) {
	ctx := context.Background()
	r := NewMemoryRepo()

	if _, err := r.GetByID(ctx, 42); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetByID err = %v, want ErrNotFound", err)
	}
	if _, err := r.GetByEmail(ctx, "no@one.com"); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetByEmail err = %v, want ErrNotFound", err)
	}
	if err := r.Delete(ctx, 42); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete err = %v, want ErrNotFound", err)
	}
}

func TestListAndDelete(t *testing.T) {
	ctx := context.Background()
	r := NewMemoryRepo()

	if list, err := r.List(ctx); err != nil || list == nil || len(list) != 0 {
		t.Errorf("empty List = (%#v, %v), want an empty non-nil slice", list, err)
	}

	r.Create(ctx, &User{Email: "a@x.com", Name: "A"})
	r.Create(ctx, &User{Email: "b@x.com", Name: "B"})
	r.Create(ctx, &User{Email: "c@x.com", Name: "C"})

	if err := r.Delete(ctx, 2); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	list, _ := r.List(ctx)
	if len(list) != 2 || list[0].ID != 1 || list[1].ID != 3 {
		t.Errorf("List after delete = %+v, want users 1 and 3", list)
	}

	// the email of a deleted user can be used again
	if err := r.Create(ctx, &User{Email: "b@x.com", Name: "B2"}); err != nil {
		t.Errorf("reusing a deleted email failed: %v", err)
	}
}

func TestCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := NewMemoryRepo()

	if err := r.Create(ctx, &User{Email: "a@x.com"}); !errors.Is(err, context.Canceled) {
		t.Errorf("Create err = %v, want context.Canceled", err)
	}
	if _, err := r.List(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("List err = %v, want context.Canceled", err)
	}
	if _, err := r.GetByID(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Errorf("GetByID err = %v, want context.Canceled", err)
	}
}
```

<details>
<summary>Solution</summary>

```go
package repo

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
)

var (
	ErrNotFound       = errors.New("user not found")
	ErrDuplicateEmail = errors.New("email already in use")
)

type User struct {
	ID    int64
	Email string
	Name  string
}

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	List(ctx context.Context) ([]User, error)
	Delete(ctx context.Context, id int64) error
}

type MemoryRepo struct {
	mu     sync.Mutex
	nextID int64
	users  map[int64]User
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{nextID: 1, users: make(map[int64]User)}
}

func (r *MemoryRepo) Create(ctx context.Context, u *User) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.users {
		if strings.EqualFold(existing.Email, u.Email) {
			return ErrDuplicateEmail
		}
	}
	u.ID = r.nextID
	r.nextID++
	r.users[u.ID] = *u
	return nil
}

func (r *MemoryRepo) GetByID(ctx context.Context, id int64) (User, error) {
	if err := ctx.Err(); err != nil {
		return User{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

func (r *MemoryRepo) GetByEmail(ctx context.Context, email string) (User, error) {
	if err := ctx.Err(); err != nil {
		return User{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return User{}, ErrNotFound
}

func (r *MemoryRepo) List(ctx context.Context) ([]User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *MemoryRepo) Delete(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[id]; !ok {
		return ErrNotFound
	}
	delete(r.users, id)
	return nil
}
```

</details>

### Exercise 2: A service with business rules

The service talks to a repository **interface**. The test uses a fake repository.

`Register(ctx, email, name)`:

1. Trim spaces from both values. Make the email lower case.
1. Empty name gives `ErrNameRequired`.
1. An email without `@`, or that starts or ends with `@`, gives `ErrInvalidEmail`.
1. Call `repo.Create`. If the repo says `ErrDuplicateEmail`, return that error.
1. Any other repo error is returned **wrapped** as `register user: <err>`.
1. On success return the created user (with its ID).

On any validation error the repo must **not** be called.

`Get(ctx, id)`:

- `id <= 0` gives `ErrNotFound` and does not call the repo.
- Otherwise return what the repo returns. Wrap unknown errors with `get user: %w`, but keep `ErrNotFound` working with `errors.Is`.

**Starter** `ch19/ex02/service.go`

```go
package service

import (
	"context"
	"errors"
)

var (
	ErrInvalidEmail   = errors.New("invalid email")
	ErrNameRequired   = errors.New("name is required")
	ErrDuplicateEmail = errors.New("email already in use")
	ErrNotFound       = errors.New("user not found")
)

type User struct {
	ID    int64
	Email string
	Name  string
}

// UserRepository is defined here, where it is used, and is kept small.
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (User, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, email, name string) (User, error) {
	return User{}, nil // TODO
}

func (s *UserService) Get(ctx context.Context, id int64) (User, error) {
	return User{}, nil // TODO
}
```

**Test** `ch19/ex02/service_test.go`

```go
package service

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeRepo struct {
	users       map[int64]User
	nextID      int64
	createCalls int
	getCalls    int
	createErr   error
	getErr      error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{users: make(map[int64]User)}
}

func (f *fakeRepo) Create(ctx context.Context, u *User) error {
	f.createCalls++
	if f.createErr != nil {
		return f.createErr
	}
	f.nextID++
	u.ID = f.nextID
	f.users[u.ID] = *u
	return nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id int64) (User, error) {
	f.getCalls++
	if f.getErr != nil {
		return User{}, f.getErr
	}
	u, ok := f.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

func TestRegisterOK(t *testing.T) {
	repo := newFakeRepo()
	svc := NewUserService(repo)

	u, err := svc.Register(context.Background(), "  Ana@Example.COM ", "  Ana  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != 1 || u.Email != "ana@example.com" || u.Name != "Ana" {
		t.Errorf("got %+v, want {1 ana@example.com Ana}", u)
	}
	if stored := repo.users[1]; stored != u {
		t.Errorf("stored %+v, returned %+v", stored, u)
	}
}

func TestRegisterValidation(t *testing.T) {
	tests := []struct {
		name, email, userName string
		want                  error
	}{
		{"empty name", "ana@example.com", "", ErrNameRequired},
		{"blank name", "ana@example.com", "   ", ErrNameRequired},
		{"empty email", "", "Ana", ErrInvalidEmail},
		{"no at sign", "ana.example.com", "Ana", ErrInvalidEmail},
		{"starts with at", "@example.com", "Ana", ErrInvalidEmail},
		{"ends with at", "ana@", "Ana", ErrInvalidEmail},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeRepo()
			_, err := NewUserService(repo).Register(context.Background(), tc.email, tc.userName)
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
			if repo.createCalls != 0 {
				t.Error("the repository must not be called when validation fails")
			}
		})
	}
}

func TestRegisterDuplicate(t *testing.T) {
	repo := newFakeRepo()
	repo.createErr = ErrDuplicateEmail

	_, err := NewUserService(repo).Register(context.Background(), "ana@example.com", "Ana")
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Errorf("err = %v, want ErrDuplicateEmail", err)
	}
}

func TestRegisterRepoFailure(t *testing.T) {
	dbDown := errors.New("db down")
	repo := newFakeRepo()
	repo.createErr = dbDown

	_, err := NewUserService(repo).Register(context.Background(), "ana@example.com", "Ana")
	if !errors.Is(err, dbDown) {
		t.Errorf("err = %v, should wrap the repo error", err)
	}
	if err == nil || !strings.HasPrefix(err.Error(), "register user: ") {
		t.Errorf("err = %v, want it to start with %q", err, "register user: ")
	}
}

func TestGet(t *testing.T) {
	repo := newFakeRepo()
	svc := NewUserService(repo)
	created, _ := svc.Register(context.Background(), "ana@example.com", "Ana")

	got, err := svc.Get(context.Background(), created.ID)
	if err != nil || got != created {
		t.Errorf("Get = (%+v, %v), want (%+v, nil)", got, err, created)
	}

	if _, err := svc.Get(context.Background(), 99); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(99) err = %v, want ErrNotFound", err)
	}
}

func TestGetInvalidID(t *testing.T) {
	repo := newFakeRepo()
	for _, id := range []int64{0, -5} {
		_, err := NewUserService(repo).Get(context.Background(), id)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Get(%d) err = %v, want ErrNotFound", id, err)
		}
	}
	if repo.getCalls != 0 {
		t.Error("the repository must not be called for an invalid id")
	}
}

func TestGetRepoFailure(t *testing.T) {
	dbDown := errors.New("db down")
	repo := newFakeRepo()
	repo.getErr = dbDown

	_, err := NewUserService(repo).Get(context.Background(), 1)
	if !errors.Is(err, dbDown) {
		t.Errorf("err = %v, should wrap the repo error", err)
	}
}
```

<details>
<summary>Solution</summary>

```go
func (s *UserService) Register(ctx context.Context, email, name string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)

	if name == "" {
		return User{}, ErrNameRequired
	}
	if !strings.Contains(email, "@") ||
		strings.HasPrefix(email, "@") ||
		strings.HasSuffix(email, "@") {
		return User{}, ErrInvalidEmail
	}

	u := User{Email: email, Name: name}
	if err := s.repo.Create(ctx, &u); err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			return User{}, ErrDuplicateEmail
		}
		return User{}, fmt.Errorf("register user: %w", err)
	}
	return u, nil
}

func (s *UserService) Get(ctx context.Context, id int64) (User, error) {
	if id <= 0 {
		return User{}, ErrNotFound
	}
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}
```

Add `"fmt"` and `"strings"` to the imports.

</details>

### Exercise 3: Scan rows safely

`*sql.Rows` has the methods `Next`, `Scan`, `Err` and `Close`. If your function takes a small interface with these methods,
you can test it with a fake. Write `ScanUsers(rows Rows) ([]User, error)`.

- Always close the rows, also when an error happens.
- Loop with `Next`, and read `ID`, `Email`, `Name` in this order with `Scan`.
- If `Scan` fails, return `nil` and an error that wraps the scan error (`scan user: %w`).
- After the loop, check `rows.Err()`. If it is not `nil`, return `nil` and an error that wraps it (`read users: %w`).

**Starter** `ch19/ex03/rows.go`

```go
package rows

type User struct {
	ID    int64
	Email string
	Name  string
}

// Rows is the part of *sql.Rows we need.
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}

func ScanUsers(rows Rows) ([]User, error) {
	return nil, nil // TODO
}
```

**Test** `ch19/ex03/rows_test.go`

```go
package rows

import (
	"errors"
	"testing"
)

type fakeRows struct {
	data      []User
	pos       int // number of times Next returned true so far
	scanErrAt int // 1 based row where Scan fails, 0 means never
	scanErr   error
	iterErr   error // returned by Err()
	closed    bool
}

func (f *fakeRows) Next() bool {
	if f.pos >= len(f.data) {
		return false
	}
	f.pos++
	return true
}

func (f *fakeRows) Scan(dest ...any) error {
	if f.scanErrAt == f.pos {
		return f.scanErr
	}
	u := f.data[f.pos-1]
	*dest[0].(*int64) = u.ID
	*dest[1].(*string) = u.Email
	*dest[2].(*string) = u.Name
	return nil
}

func (f *fakeRows) Err() error   { return f.iterErr }
func (f *fakeRows) Close() error { f.closed = true; return nil }

func TestScanUsers(t *testing.T) {
	rows := &fakeRows{data: []User{
		{1, "a@x.com", "A"},
		{2, "b@x.com", "B"},
		{3, "c@x.com", "C"},
	}}

	got, err := ScanUsers(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 || got[0].Email != "a@x.com" || got[2].Name != "C" || got[1].ID != 2 {
		t.Errorf("got %+v", got)
	}
	if !rows.closed {
		t.Error("rows were not closed")
	}
}

func TestScanUsersEmpty(t *testing.T) {
	rows := &fakeRows{}
	got, err := ScanUsers(rows)
	if err != nil || len(got) != 0 {
		t.Errorf("got (%v, %v), want an empty result and no error", got, err)
	}
	if !rows.closed {
		t.Error("rows were not closed")
	}
}

func TestScanUsersScanError(t *testing.T) {
	boom := errors.New("bad column type")
	rows := &fakeRows{
		data:      []User{{1, "a@x.com", "A"}, {2, "b@x.com", "B"}},
		scanErrAt: 2,
		scanErr:   boom,
	}

	got, err := ScanUsers(rows)
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap the scan error", err)
	}
	if got != nil {
		t.Errorf("got %+v, want nil on error", got)
	}
	if !rows.closed {
		t.Error("rows must be closed even when Scan fails")
	}
}

func TestScanUsersIterationError(t *testing.T) {
	connLost := errors.New("connection lost")
	rows := &fakeRows{
		data:    []User{{1, "a@x.com", "A"}},
		iterErr: connLost,
	}

	got, err := ScanUsers(rows)
	if !errors.Is(err, connLost) {
		t.Errorf("err = %v, want it to wrap rows.Err()", err)
	}
	if got != nil {
		t.Errorf("got %+v, want nil on error", got)
	}
	if !rows.closed {
		t.Error("rows were not closed")
	}
}
```

<details>
<summary>Solution</summary>

```go
package rows

import "fmt"

type User struct {
	ID    int64
	Email string
	Name  string
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}

func ScanUsers(rows Rows) ([]User, error) {
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read users: %w", err)
	}
	return users, nil
}
```

</details>

### Exercise 4: A transaction helper

A transaction must end with exactly one `Commit` or one `Rollback`. Writing this by hand in every function is easy to get wrong,
so write it **once**.

`WithTx(ctx, db, fn)`:

1. Call `db.Begin(ctx)`. If it fails, return an error wrapping it (`begin tx: %w`). Do not call `fn`.
1. Call `fn(tx)`.
1. If `fn` returns an error: call `Rollback`, then return `fn`'s error. If `Rollback` also fails, return **both** (use `errors.Join`) so `errors.Is` finds each one.
1. If `fn` panics: call `Rollback`, then panic again with the same value.
1. If `fn` succeeds: call `Commit`. If it fails, return an error wrapping it (`commit tx: %w`).

**Starter** `ch19/ex04/tx.go`

```go
package tx

import "context"

type Tx interface {
	Commit() error
	Rollback() error
}

type TxBeginner interface {
	Begin(ctx context.Context) (Tx, error)
}

func WithTx(ctx context.Context, db TxBeginner, fn func(tx Tx) error) error {
	return nil // TODO
}
```

**Test** `ch19/ex04/tx_test.go`

```go
package tx

import (
	"context"
	"errors"
	"testing"
)

type fakeTx struct {
	committed   bool
	rolledBack  bool
	commitErr   error
	rollbackErr error
}

func (f *fakeTx) Commit() error   { f.committed = true; return f.commitErr }
func (f *fakeTx) Rollback() error { f.rolledBack = true; return f.rollbackErr }

type fakeDB struct {
	tx       *fakeTx
	beginErr error
}

func (d *fakeDB) Begin(ctx context.Context) (Tx, error) {
	if d.beginErr != nil {
		return nil, d.beginErr
	}
	return d.tx, nil
}

func TestWithTxCommits(t *testing.T) {
	db := &fakeDB{tx: &fakeTx{}}
	called := false

	err := WithTx(context.Background(), db, func(tx Tx) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || !db.tx.committed || db.tx.rolledBack {
		t.Errorf("called=%v committed=%v rolledBack=%v, want true true false",
			called, db.tx.committed, db.tx.rolledBack)
	}
}

func TestWithTxRollsBackOnError(t *testing.T) {
	db := &fakeDB{tx: &fakeTx{}}
	boom := errors.New("insert failed")

	err := WithTx(context.Background(), db, func(tx Tx) error { return boom })

	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want the error from fn", err)
	}
	if db.tx.committed || !db.tx.rolledBack {
		t.Errorf("committed=%v rolledBack=%v, want false true", db.tx.committed, db.tx.rolledBack)
	}
}

func TestWithTxRollbackAlsoFails(t *testing.T) {
	rollbackFailed := errors.New("rollback failed")
	db := &fakeDB{tx: &fakeTx{rollbackErr: rollbackFailed}}
	boom := errors.New("insert failed")

	err := WithTx(context.Background(), db, func(tx Tx) error { return boom })

	if !errors.Is(err, boom) || !errors.Is(err, rollbackFailed) {
		t.Errorf("err = %v, want both the fn error and the rollback error", err)
	}
}

func TestWithTxBeginFails(t *testing.T) {
	noConn := errors.New("no connection")
	db := &fakeDB{beginErr: noConn}
	called := false

	err := WithTx(context.Background(), db, func(tx Tx) error {
		called = true
		return nil
	})

	if !errors.Is(err, noConn) {
		t.Errorf("err = %v, want it to wrap the begin error", err)
	}
	if called {
		t.Error("fn must not run when Begin fails")
	}
}

func TestWithTxCommitFails(t *testing.T) {
	commitFailed := errors.New("commit failed")
	db := &fakeDB{tx: &fakeTx{commitErr: commitFailed}}

	err := WithTx(context.Background(), db, func(tx Tx) error { return nil })
	if !errors.Is(err, commitFailed) {
		t.Errorf("err = %v, want it to wrap the commit error", err)
	}
}

func TestWithTxPanicRollsBackAndRepanics(t *testing.T) {
	db := &fakeDB{tx: &fakeTx{}}

	defer func() {
		r := recover()
		if r != "boom" {
			t.Errorf("recovered %v, want the original panic value %q", r, "boom")
		}
		if db.tx.committed || !db.tx.rolledBack {
			t.Errorf("committed=%v rolledBack=%v, want false true", db.tx.committed, db.tx.rolledBack)
		}
	}()

	WithTx(context.Background(), db, func(tx Tx) error { panic("boom") })
	t.Error("WithTx should have panicked")
}
```

<details>
<summary>Solution</summary>

```go
package tx

import (
	"context"
	"errors"
	"fmt"
)

type Tx interface {
	Commit() error
	Rollback() error
}

type TxBeginner interface {
	Begin(ctx context.Context) (Tx, error)
}

func WithTx(ctx context.Context, db TxBeginner, fn func(tx Tx) error) (err error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, fmt.Errorf("rollback: %w", rbErr))
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
```

</details>

> [!NOTE]
> To use `WithTx` with a real `*sql.DB`, write a small adapter. Return an explicit `nil` on error. If you
> return the `*sql.Tx` directly with a `nil` pointer, the interface value is **not** `nil` (a classic Go trap).
>
> ```go
> type sqlBeginner struct{ db *sql.DB }
>
> func (b sqlBeginner) Begin(ctx context.Context) (Tx, error) {
> 	tx, err := b.db.BeginTx(ctx, nil)
> 	if err != nil {
> 		return nil, err
> 	}
> 	return tx, nil
> }
> ```

## Checkpoint

- [ ] I never build SQL with string concatenation, I use placeholders
- [ ] I can explain the roles of handler, service and repository
- [ ] I test services with a fake repository
- [ ] I always close rows and check `rows.Err()`
- [ ] I know a transaction needs exactly one commit or rollback
