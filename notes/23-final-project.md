# Chapter 23: Final Project, a Notes API

You now have all the pieces. In this project you build a small but real backend:
**users can register, log in, and manage their own private notes.**

You will use: structs, maps, slices, errors, JSON, `net/http` routing, validation, a mutex,
password hashing, signed tokens, auth middleware, query parameters and pagination.

The tests only talk to your API over HTTP (with `httptest`). You are free to design the inside
any way you like: one file, many files, a separate store type, anything.

## The API

| Method | Path | Auth | Success | What it does |
|--------|------|------|---------|--------------|
| `POST` | `/register` | no | `201` | Create an account |
| `POST` | `/login` | no | `200` | Get a token |
| `POST` | `/notes` | yes | `201` | Create a note |
| `GET` | `/notes` | yes | `200` | List my notes (search and pages) |
| `GET` | `/notes/{id}` | yes | `200` | Read one of my notes |
| `DELETE` | `/notes/{id}` | yes | `204` | Delete one of my notes |

All responses with a body are JSON. Errors always look like `{"error": "message"}`.

## Setup

Create the folder `ch23/notesapi/` and add these files. They are **given**, you do not change them.
`security.go` is a compact copy of what you built in chapter 21.

**Starter** `ch23/notesapi/helpers.go`

```go
package notesapi

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// decodeJSON reads the request body into v. On failure it writes
// a 400 response itself and returns false.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return false
	}
	return true
}
```

**Starter** `ch23/notesapi/security.go`

```go
package notesapi

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

const hashRounds = 10000

func derive(password string, salt []byte, rounds int) []byte {
	sum := sha256.Sum256(append(append([]byte{}, salt...), password...))
	for i := 1; i < rounds; i++ {
		sum = sha256.Sum256(append(sum[:], salt...))
	}
	return sum[:]
}

// hashPassword is the learning version from chapter 21.
// Use bcrypt or argon2 in real projects.
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	digest := derive(password, salt, hashRounds)
	return fmt.Sprintf("v1$%d$%s$%s", hashRounds, hex.EncodeToString(salt), hex.EncodeToString(digest)), nil
}

func checkPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "v1" {
		return false
	}
	rounds, err := strconv.Atoi(parts[1])
	if err != nil || rounds < 1 {
		return false
	}
	salt, err := hex.DecodeString(parts[2])
	if err != nil {
		return false
	}
	want, err := hex.DecodeString(parts[3])
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(derive(password, salt, rounds), want) == 1
}

type claims struct {
	UserID    int64 `json:"uid"`
	ExpiresAt int64 `json:"exp"`
}

func mac(payload string, secret []byte) []byte {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(payload))
	return m.Sum(nil)
}

func signToken(c claims, secret []byte) (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(data)
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac(payload, secret)), nil
}

func verifyToken(token string, secret []byte, now time.Time) (claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return claims{}, ErrInvalidToken
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(sig, mac(parts[0], secret)) {
		return claims{}, ErrInvalidToken
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return claims{}, ErrInvalidToken
	}
	var c claims
	if err := json.Unmarshal(data, &c); err != nil {
		return claims{}, ErrInvalidToken
	}
	if now.Unix() >= c.ExpiresAt {
		return claims{}, ErrExpiredToken
	}
	return c, nil
}
```

**Starter** `ch23/notesapi/server.go` (this is **your** file. Everything starts as "not implemented".)

```go
package notesapi

import (
	"net/http"
	"time"
)

type Config struct {
	Secret []byte           // key for signing tokens
	Now    func() time.Time // clock, so tests can move time
}

type server struct {
	cfg Config
	// TODO: add your own state here (users, notes, a mutex ...)
}

type authedHandler func(w http.ResponseWriter, r *http.Request, userID int64)

func NewServer(cfg Config) http.Handler {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	s := &server{cfg: cfg}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", s.register)
	mux.HandleFunc("POST /login", s.login)
	mux.HandleFunc("POST /notes", s.requireAuth(s.createNote))
	mux.HandleFunc("GET /notes", s.requireAuth(s.listNotes))
	mux.HandleFunc("GET /notes/{id}", s.requireAuth(s.getNote))
	mux.HandleFunc("DELETE /notes/{id}", s.requireAuth(s.deleteNote))
	return mux
}

func notImplemented(w http.ResponseWriter) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (s *server) requireAuth(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { notImplemented(w) }
}

func (s *server) register(w http.ResponseWriter, r *http.Request) { notImplemented(w) }
func (s *server) login(w http.ResponseWriter, r *http.Request)    { notImplemented(w) }

func (s *server) createNote(w http.ResponseWriter, r *http.Request, userID int64) { notImplemented(w) }
func (s *server) listNotes(w http.ResponseWriter, r *http.Request, userID int64)  { notImplemented(w) }
func (s *server) getNote(w http.ResponseWriter, r *http.Request, userID int64)    { notImplemented(w) }
func (s *server) deleteNote(w http.ResponseWriter, r *http.Request, userID int64) { notImplemented(w) }
```

**Test** `ch23/notesapi/helpers_test.go` (shared by all milestones, copy this first)

```go
package notesapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var testSecret = []byte("test-secret")

type clock struct{ t time.Time }

func (c *clock) Now() time.Time          { return c.t }
func (c *clock) Advance(d time.Duration) { c.t = c.t.Add(d) }

func newTestServer() (http.Handler, *clock) {
	c := &clock{t: time.Unix(1_700_000_000, 0)}
	return NewServer(Config{Secret: testSecret, Now: c.Now}), c
}

func js(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func send(h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func jsonBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("response is not a JSON object: %v (%q)", err, rec.Body.String())
	}
	return m
}

func errMsg(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	msg, _ := jsonBody(t, rec)["error"].(string)
	return msg
}

func mustRegister(t *testing.T, h http.Handler, email, password string) {
	t.Helper()
	rec := send(h, http.MethodPost, "/register", "", js(map[string]string{"email": email, "password": password}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("register %s: status = %d, want 201 (%s)", email, rec.Code, rec.Body.String())
	}
}

func mustLogin(t *testing.T, h http.Handler, email, password string) string {
	t.Helper()
	rec := send(h, http.MethodPost, "/login", "", js(map[string]string{"email": email, "password": password}))
	if rec.Code != http.StatusOK {
		t.Fatalf("login %s: status = %d, want 200 (%s)", email, rec.Code, rec.Body.String())
	}
	token, _ := jsonBody(t, rec)["token"].(string)
	if token == "" {
		t.Fatalf("login %s: no token in %s", email, rec.Body.String())
	}
	return token
}

// signedUp registers and logs in a user, and returns the token.
func signedUp(t *testing.T, h http.Handler, email string) string {
	t.Helper()
	mustRegister(t, h, email, "password123")
	return mustLogin(t, h, email, "password123")
}

func mustCreateNote(t *testing.T, h http.Handler, token, title, body string) int {
	t.Helper()
	rec := send(h, http.MethodPost, "/notes", token, js(map[string]string{"title": title, "body": body}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create note %q: status = %d, want 201 (%s)", title, rec.Code, rec.Body.String())
	}
	id, _ := jsonBody(t, rec)["id"].(float64)
	return int(id)
}
```

---

## Milestone 1: Accounts

### `POST /register`

Body: `{"email": "...", "password": "..."}`

| Situation | Response |
|-----------|----------|
| Body is not valid JSON | `400` `{"error":"invalid JSON"}` |
| Email (after trim and lower case) has no `@` | `400` `{"error":"invalid email"}` |
| Password shorter than 8 characters | `400` `{"error":"password must be at least 8 characters"}` |
| Email already registered (any letter case) | `409` `{"error":"email already registered"}` |
| Success | `201` `{"id": 1, "email": "ana@example.com"}` |

The email is **trimmed and lower cased** before it is stored. Ids start at 1.
The response must never contain the password or its hash. Store only the hash from `hashPassword`.

### `POST /login`

Body: `{"email": "...", "password": "..."}`

| Situation | Response |
|-----------|----------|
| Body is not valid JSON | `400` `{"error":"invalid JSON"}` |
| Unknown email **or** wrong password | `401` `{"error":"invalid credentials"}` (the **same** message for both) |
| Success | `200` `{"token": "<token>"}` |

The token is made with `signToken`. It is valid for **1 hour**: `ExpiresAt = cfg.Now().Add(time.Hour).Unix()`.
Email matching ignores letter case.

**Test** `ch23/notesapi/m1_accounts_test.go`

```go
package notesapi

import (
	"net/http"
	"strings"
	"testing"
)

func TestM1_Register(t *testing.T) {
	h, _ := newTestServer()
	rec := send(h, http.MethodPost, "/register", "", js(map[string]string{
		"email": "  Ana@Example.COM ", "password": "password123",
	}))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (%s)", rec.Code, rec.Body.String())
	}
	body := jsonBody(t, rec)
	if body["id"] != float64(1) || body["email"] != "ana@example.com" {
		t.Errorf("body = %v, want id 1 and email ana@example.com", body)
	}
	if len(body) != 2 {
		t.Errorf("body has extra fields: %v (never return the password or its hash)", body)
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Errorf("the response mentions the password: %s", rec.Body.String())
	}
}

func TestM1_RegisterAssignsIncreasingIDs(t *testing.T) {
	h, _ := newTestServer()
	mustRegister(t, h, "a@example.com", "password123")

	rec := send(h, http.MethodPost, "/register", "", js(map[string]string{
		"email": "b@example.com", "password": "password123",
	}))
	if jsonBody(t, rec)["id"] != float64(2) {
		t.Errorf("second user id = %v, want 2", jsonBody(t, rec)["id"])
	}
}

func TestM1_RegisterValidation(t *testing.T) {
	h, _ := newTestServer()

	tests := []struct {
		name string
		body string
		want string
	}{
		{"invalid JSON", `{"email":`, "invalid JSON"},
		{"no at sign", js(map[string]string{"email": "nope", "password": "password123"}), "invalid email"},
		{"empty email", js(map[string]string{"email": "", "password": "password123"}), "invalid email"},
		{"short password", js(map[string]string{"email": "a@example.com", "password": "short"}), "password must be at least 8 characters"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := send(h, http.MethodPost, "/register", "", tc.body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
			if msg := errMsg(t, rec); msg != tc.want {
				t.Errorf("error = %q, want %q", msg, tc.want)
			}
		})
	}
}

func TestM1_RegisterDuplicate(t *testing.T) {
	h, _ := newTestServer()
	mustRegister(t, h, "ana@example.com", "password123")

	rec := send(h, http.MethodPost, "/register", "", js(map[string]string{
		"email": "ANA@example.com", "password": "another-password",
	}))
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rec.Code)
	}
	if msg := errMsg(t, rec); msg != "email already registered" {
		t.Errorf("error = %q, want %q", msg, "email already registered")
	}
}

func TestM1_Login(t *testing.T) {
	h, _ := newTestServer()
	mustRegister(t, h, "ana@example.com", "password123")

	token := mustLogin(t, h, "ana@example.com", "password123")
	if strings.Count(token, ".") != 1 {
		t.Errorf("token %q does not look like a signed token", token)
	}

	// the email can be written in any letter case
	mustLogin(t, h, "  ANA@Example.com", "password123")
}

func TestM1_LoginFailures(t *testing.T) {
	h, _ := newTestServer()
	mustRegister(t, h, "ana@example.com", "password123")

	tests := []struct {
		name string
		body string
	}{
		{"wrong password", js(map[string]string{"email": "ana@example.com", "password": "wrong-password"})},
		{"unknown email", js(map[string]string{"email": "nobody@example.com", "password": "password123"})},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := send(h, http.MethodPost, "/login", "", tc.body)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", rec.Code)
			}
			if msg := errMsg(t, rec); msg != "invalid credentials" {
				t.Errorf("error = %q, want %q (the same message for both cases)", msg, "invalid credentials")
			}
		})
	}
}

func TestM1_LoginInvalidJSON(t *testing.T) {
	h, _ := newTestServer()
	rec := send(h, http.MethodPost, "/login", "", `{oops`)
	if rec.Code != http.StatusBadRequest || errMsg(t, rec) != "invalid JSON" {
		t.Errorf("got %d %q, want 400 invalid JSON", rec.Code, errMsg(t, rec))
	}
}
```

---

## Milestone 2: Private notes

Implement `requireAuth` and the three note endpoints `createNote`, `getNote`, `deleteNote`.

### Authentication (`requireAuth`)

Read `Authorization: Bearer <token>` and check it with `verifyToken(token, cfg.Secret, cfg.Now())`.

| Situation | Response |
|-----------|----------|
| Header missing, wrong scheme, empty or invalid token | `401` `{"error":"unauthorized"}` |
| Token expired | `401` `{"error":"token expired"}` |
| Valid | call the handler with the user id from the token |

### Notes

A note is `{"id": 1, "title": "...", "body": "..."}`. The JSON has **only** these three fields (no owner).
Note ids are one counter for the whole server, starting at 1.

| Request | Rules |
|---------|-------|
| `POST /notes` | Body `{"title": "...", "body": "..."}`. The title is trimmed and must not be empty. Invalid JSON gives `400 invalid JSON`, empty title gives `400 title is required`. Success is `201` and the note. |
| `GET /notes/{id}` | `400 invalid id` if the id is not a number. `404 note not found` if it does not exist **or belongs to someone else**. Otherwise `200` and the note. |
| `DELETE /notes/{id}` | Same checks as `GET`. Success is `204` with an empty body. |

> [!NOTE]
> Returning `404` (not `403`) for another user's note is on purpose. It does not tell strangers that the note exists.

**Test** `ch23/notesapi/m2_notes_test.go`

```go
package notesapi

import (
	"net/http"
	"testing"
	"time"
)

func TestM2_AuthRequired(t *testing.T) {
	h, _ := newTestServer()

	routes := []struct{ method, path, body string }{
		{http.MethodPost, "/notes", `{"title":"x"}`},
		{http.MethodGet, "/notes", ""},
		{http.MethodGet, "/notes/1", ""},
		{http.MethodDelete, "/notes/1", ""},
	}
	for _, rt := range routes {
		for _, token := range []string{"", "garbage"} {
			rec := send(h, rt.method, rt.path, token, rt.body)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("%s %s (token %q): status = %d, want 401", rt.method, rt.path, token, rec.Code)
				continue
			}
			if msg := errMsg(t, rec); msg != "unauthorized" {
				t.Errorf("%s %s (token %q): error = %q, want %q", rt.method, rt.path, token, msg, "unauthorized")
			}
		}
	}
}

func TestM2_TokenFromAnotherSecretIsRejected(t *testing.T) {
	h, c := newTestServer()
	forged, err := signToken(claims{UserID: 1, ExpiresAt: c.Now().Add(time.Hour).Unix()}, []byte("not-the-secret"))
	if err != nil {
		t.Fatal(err)
	}
	if rec := send(h, http.MethodGet, "/notes", forged, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestM2_TokenExpires(t *testing.T) {
	h, clock := newTestServer()
	token := signedUp(t, h, "ana@example.com")

	clock.Advance(59 * time.Minute)
	if rec := send(h, http.MethodGet, "/notes", token, ""); rec.Code != http.StatusOK {
		t.Fatalf("after 59 minutes: status = %d, want 200", rec.Code)
	}

	clock.Advance(2 * time.Minute) // now 61 minutes after login
	rec := send(h, http.MethodGet, "/notes", token, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("after 61 minutes: status = %d, want 401", rec.Code)
	}
	if msg := errMsg(t, rec); msg != "token expired" {
		t.Errorf("error = %q, want %q", msg, "token expired")
	}
}

func TestM2_CreateAndGet(t *testing.T) {
	h, _ := newTestServer()
	token := signedUp(t, h, "ana@example.com")

	rec := send(h, http.MethodPost, "/notes", token, js(map[string]string{"title": "  Shopping ", "body": "milk"}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: status = %d, want 201 (%s)", rec.Code, rec.Body.String())
	}
	created := jsonBody(t, rec)
	if created["id"] != float64(1) || created["title"] != "Shopping" || created["body"] != "milk" {
		t.Errorf("created = %v, want id 1, title Shopping, body milk", created)
	}
	if len(created) != 3 {
		t.Errorf("a note must have exactly id, title and body, got %v", created)
	}

	rec = send(h, http.MethodGet, "/notes/1", token, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get: status = %d, want 200", rec.Code)
	}
	got := jsonBody(t, rec)
	if got["id"] != float64(1) || got["title"] != "Shopping" || got["body"] != "milk" {
		t.Errorf("get = %v", got)
	}
}

func TestM2_CreateValidation(t *testing.T) {
	h, _ := newTestServer()
	token := signedUp(t, h, "ana@example.com")

	rec := send(h, http.MethodPost, "/notes", token, `{"title":`)
	if rec.Code != http.StatusBadRequest || errMsg(t, rec) != "invalid JSON" {
		t.Errorf("bad JSON: got %d %q", rec.Code, errMsg(t, rec))
	}

	for _, body := range []string{`{}`, `{"title":"   ","body":"x"}`} {
		rec := send(h, http.MethodPost, "/notes", token, body)
		if rec.Code != http.StatusBadRequest || errMsg(t, rec) != "title is required" {
			t.Errorf("body %s: got %d %q, want 400 title is required", body, rec.Code, errMsg(t, rec))
		}
	}
}

func TestM2_NotesArePrivate(t *testing.T) {
	h, _ := newTestServer()
	ana := signedUp(t, h, "ana@example.com")
	bob := signedUp(t, h, "bob@example.com")

	id := mustCreateNote(t, h, ana, "Ana's secret", "dear diary")
	path := "/notes/" + js(id)

	rec := send(h, http.MethodGet, path, bob, "")
	if rec.Code != http.StatusNotFound || errMsg(t, rec) != "note not found" {
		t.Errorf("bob reading ana's note: got %d %q, want 404 note not found", rec.Code, errMsg(t, rec))
	}

	rec = send(h, http.MethodDelete, path, bob, "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("bob deleting ana's note: status = %d, want 404", rec.Code)
	}

	if rec := send(h, http.MethodGet, path, ana, ""); rec.Code != http.StatusOK {
		t.Errorf("ana lost her note: status = %d, want 200", rec.Code)
	}
}

func TestM2_Delete(t *testing.T) {
	h, _ := newTestServer()
	token := signedUp(t, h, "ana@example.com")
	id := mustCreateNote(t, h, token, "Temp", "x")
	path := "/notes/" + js(id)

	rec := send(h, http.MethodDelete, path, token, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", rec.Body.String())
	}

	if rec := send(h, http.MethodGet, path, token, ""); rec.Code != http.StatusNotFound {
		t.Errorf("get after delete: status = %d, want 404", rec.Code)
	}
	if rec := send(h, http.MethodDelete, path, token, ""); rec.Code != http.StatusNotFound {
		t.Errorf("second delete: status = %d, want 404", rec.Code)
	}
}

func TestM2_InvalidAndUnknownID(t *testing.T) {
	h, _ := newTestServer()
	token := signedUp(t, h, "ana@example.com")

	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		rec := send(h, method, "/notes/abc", token, "")
		if rec.Code != http.StatusBadRequest || errMsg(t, rec) != "invalid id" {
			t.Errorf("%s /notes/abc: got %d %q, want 400 invalid id", method, rec.Code, errMsg(t, rec))
		}

		rec = send(h, method, "/notes/999", token, "")
		if rec.Code != http.StatusNotFound || errMsg(t, rec) != "note not found" {
			t.Errorf("%s /notes/999: got %d %q, want 404 note not found", method, rec.Code, errMsg(t, rec))
		}
	}
}
```

---

## Milestone 3: List, search and pages

### `GET /notes`

Returns only the notes of the logged in user, ordered by id (oldest first).

Query parameters:

| Parameter | Default | Rule | Error |
|-----------|---------|------|-------|
| `limit` | `10` | whole number from 1 to 100 | `400` `{"error":"invalid limit"}` |
| `offset` | `0` | whole number, 0 or more | `400` `{"error":"invalid offset"}` |
| `q` | none | keep notes whose title **or** body contains `q`, ignoring letter case | none |

Response `200`:

```json
{
  "items": [ {"id": 1, "title": "...", "body": "..."} ],
  "total": 25
}
```

- `total` is the number of notes that match `q`, **before** `limit` and `offset` are applied.
- `items` is `[]` (never `null`) when there is nothing to show, for example when `offset` is past the end.

**Test** `ch23/notesapi/m3_list_test.go`

```go
package notesapi

import (
	"fmt"
	"net/http"
	"testing"
)

func list(t *testing.T, h http.Handler, token, query string) (items []map[string]any, total int) {
	t.Helper()
	rec := send(h, http.MethodGet, "/notes"+query, token, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /notes%s: status = %d, want 200 (%s)", query, rec.Code, rec.Body.String())
	}
	body := jsonBody(t, rec)

	raw, ok := body["items"].([]any)
	if !ok {
		t.Fatalf("GET /notes%s: \"items\" must be a JSON array (use [] not null), got %v", query, body["items"])
	}
	for _, it := range raw {
		items = append(items, it.(map[string]any))
	}
	n, _ := body["total"].(float64)
	return items, int(n)
}

func titles(items []map[string]any) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i], _ = it["title"].(string)
	}
	return out
}

func TestM3_EmptyList(t *testing.T) {
	h, _ := newTestServer()
	token := signedUp(t, h, "ana@example.com")

	items, total := list(t, h, token, "")
	if len(items) != 0 || total != 0 {
		t.Errorf("got %d items and total %d, want 0 and 0", len(items), total)
	}
}

func TestM3_OnlyOwnNotes(t *testing.T) {
	h, _ := newTestServer()
	ana := signedUp(t, h, "ana@example.com")
	bob := signedUp(t, h, "bob@example.com")

	mustCreateNote(t, h, ana, "ana 1", "")
	mustCreateNote(t, h, bob, "bob 1", "")
	mustCreateNote(t, h, ana, "ana 2", "")

	items, total := list(t, h, ana, "")
	if total != 2 || len(items) != 2 || items[0]["title"] != "ana 1" || items[1]["title"] != "ana 2" {
		t.Errorf("ana sees %v (total %d), want ana 1 and ana 2", titles(items), total)
	}

	items, total = list(t, h, bob, "")
	if total != 1 || len(items) != 1 || items[0]["title"] != "bob 1" {
		t.Errorf("bob sees %v (total %d), want only bob 1", titles(items), total)
	}
}

func TestM3_Pagination(t *testing.T) {
	h, _ := newTestServer()
	token := signedUp(t, h, "ana@example.com")
	for i := 1; i <= 25; i++ {
		mustCreateNote(t, h, token, fmt.Sprintf("note %02d", i), "")
	}

	t.Run("default limit is 10", func(t *testing.T) {
		items, total := list(t, h, token, "")
		if len(items) != 10 || total != 25 || items[0]["title"] != "note 01" || items[9]["title"] != "note 10" {
			t.Errorf("got %d items, total %d, titles %v", len(items), total, titles(items))
		}
	})

	t.Run("last page is shorter", func(t *testing.T) {
		items, total := list(t, h, token, "?limit=10&offset=20")
		if len(items) != 5 || total != 25 || items[0]["title"] != "note 21" || items[4]["title"] != "note 25" {
			t.Errorf("got %d items, total %d, titles %v", len(items), total, titles(items))
		}
	})

	t.Run("middle page", func(t *testing.T) {
		items, _ := list(t, h, token, "?limit=5&offset=5")
		got := titles(items)
		want := []string{"note 06", "note 07", "note 08", "note 09", "note 10"}
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("offset past the end", func(t *testing.T) {
		items, total := list(t, h, token, "?offset=100")
		if len(items) != 0 || total != 25 {
			t.Errorf("got %d items, total %d, want 0 and 25", len(items), total)
		}
	})

	t.Run("largest limit", func(t *testing.T) {
		items, _ := list(t, h, token, "?limit=100")
		if len(items) != 25 {
			t.Errorf("got %d items, want 25", len(items))
		}
	})
}

func TestM3_Search(t *testing.T) {
	h, _ := newTestServer()
	token := signedUp(t, h, "ana@example.com")
	mustCreateNote(t, h, token, "Groceries", "milk and eggs")
	mustCreateNote(t, h, token, "Work", "finish REPORT")
	mustCreateNote(t, h, token, "Ideas", "send report to boss")

	t.Run("matches body, any letter case", func(t *testing.T) {
		items, total := list(t, h, token, "?q=report")
		if total != 2 || fmt.Sprint(titles(items)) != "[Work Ideas]" {
			t.Errorf("got %v (total %d), want [Work Ideas]", titles(items), total)
		}
	})

	t.Run("matches title", func(t *testing.T) {
		items, total := list(t, h, token, "?q=GROCER")
		if total != 1 || fmt.Sprint(titles(items)) != "[Groceries]" {
			t.Errorf("got %v (total %d), want [Groceries]", titles(items), total)
		}
	})

	t.Run("no match", func(t *testing.T) {
		items, total := list(t, h, token, "?q=zzz")
		if len(items) != 0 || total != 0 {
			t.Errorf("got %d items, total %d, want 0 and 0", len(items), total)
		}
	})

	t.Run("total ignores the page size", func(t *testing.T) {
		items, total := list(t, h, token, "?q=report&limit=1")
		if len(items) != 1 || total != 2 {
			t.Errorf("got %d items, total %d, want 1 and 2", len(items), total)
		}
	})
}

func TestM3_InvalidParameters(t *testing.T) {
	h, _ := newTestServer()
	token := signedUp(t, h, "ana@example.com")

	tests := []struct {
		query string
		want  string
	}{
		{"?limit=0", "invalid limit"},
		{"?limit=101", "invalid limit"},
		{"?limit=abc", "invalid limit"},
		{"?offset=-1", "invalid offset"},
		{"?offset=x", "invalid offset"},
	}
	for _, tc := range tests {
		rec := send(h, http.MethodGet, "/notes"+tc.query, token, "")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", tc.query, rec.Code)
			continue
		}
		if msg := errMsg(t, rec); msg != tc.want {
			t.Errorf("%s: error = %q, want %q", tc.query, msg, tc.want)
		}
	}
}
```

---

## Working tips

- Go milestone by milestone: `go test -run 'TestM1' -v ./ch23/notesapi/`, then `TestM2`, then `TestM3`.
- Run `go test -race ./ch23/notesapi/` at the end. Your state is shared by many requests, so it needs a mutex.
- Keep the HTTP code thin. If a handler grows long, move logic into small functions or a store type.
- Write a helper that parses the id from the path, and one that reads the JSON body. You will need them more than once.

<details>
<summary>Reference solution (server.go). Try on your own first.</summary>

```go
package notesapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Secret []byte
	Now    func() time.Time
}

type user struct {
	id           int64
	email        string
	passwordHash string
}

type note struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	ownerID int64
}

type server struct {
	cfg Config

	mu         sync.Mutex
	users      map[string]*user // by lower case email
	nextUserID int64
	notes      []note // always ordered by id
	nextNoteID int64

	dummyHash string // used so unknown emails take as long as wrong passwords
}

type authedHandler func(w http.ResponseWriter, r *http.Request, userID int64)

func NewServer(cfg Config) http.Handler {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	dummy, _ := hashPassword("dummy-password")
	s := &server{
		cfg:       cfg,
		users:     make(map[string]*user),
		dummyHash: dummy,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", s.register)
	mux.HandleFunc("POST /login", s.login)
	mux.HandleFunc("POST /notes", s.requireAuth(s.createNote))
	mux.HandleFunc("GET /notes", s.requireAuth(s.listNotes))
	mux.HandleFunc("GET /notes/{id}", s.requireAuth(s.getNote))
	mux.HandleFunc("DELETE /notes/{id}", s.requireAuth(s.deleteNote))
	return mux
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ---- accounts ----

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *server) register(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decodeJSON(w, r, &in) {
		return
	}
	email := normalizeEmail(in.Email)
	if !strings.Contains(email, "@") {
		writeError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if len(in.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hash, err := hashPassword(in.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[email]; exists {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}
	s.nextUserID++
	u := &user{id: s.nextUserID, email: email, passwordHash: hash}
	s.users[email] = u

	writeJSON(w, http.StatusCreated, map[string]any{"id": u.id, "email": u.email})
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decodeJSON(w, r, &in) {
		return
	}

	s.mu.Lock()
	u, found := s.users[normalizeEmail(in.Email)]
	s.mu.Unlock()

	hash := s.dummyHash
	if found {
		hash = u.passwordHash
	}
	if !checkPassword(in.Password, hash) || !found {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := signToken(claims{
		UserID:    u.id,
		ExpiresAt: s.cfg.Now().Add(time.Hour).Unix(),
	}, s.cfg.Secret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// ---- authentication ----

func (s *server) requireAuth(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || raw == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		c, err := verifyToken(raw, s.cfg.Secret, s.cfg.Now())
		if errors.Is(err, ErrExpiredToken) {
			writeError(w, http.StatusUnauthorized, "token expired")
			return
		}
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r, c.UserID)
	}
}

// ---- notes ----

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

// indexOf returns the position of the note with this id owned by userID, or -1.
// The caller must hold s.mu.
func (s *server) indexOf(id, userID int64) int {
	for i, n := range s.notes {
		if n.ID == id && n.ownerID == userID {
			return i
		}
	}
	return -1
}

func (s *server) createNote(w http.ResponseWriter, r *http.Request, userID int64) {
	var in struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextNoteID++
	n := note{ID: s.nextNoteID, Title: title, Body: in.Body, ownerID: userID}
	s.notes = append(s.notes, n)
	writeJSON(w, http.StatusCreated, n)
}

func (s *server) getNote(w http.ResponseWriter, r *http.Request, userID int64) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.indexOf(id, userID)
	if i < 0 {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}
	writeJSON(w, http.StatusOK, s.notes[i])
}

func (s *server) deleteNote(w http.ResponseWriter, r *http.Request, userID int64) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.indexOf(id, userID)
	if i < 0 {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}
	s.notes = append(s.notes[:i], s.notes[i+1:]...)
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) listNotes(w http.ResponseWriter, r *http.Request, userID int64) {
	q := r.URL.Query()

	limit := 10
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 100 {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = n
	}

	offset := 0
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			writeError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = n
	}

	search := strings.ToLower(q.Get("q"))

	s.mu.Lock()
	matches := make([]note, 0)
	for _, n := range s.notes {
		if n.ownerID != userID {
			continue
		}
		if search != "" &&
			!strings.Contains(strings.ToLower(n.Title), search) &&
			!strings.Contains(strings.ToLower(n.Body), search) {
			continue
		}
		matches = append(matches, n)
	}
	s.mu.Unlock()

	total := len(matches)
	start := min(offset, total)
	end := min(start+limit, total)

	writeJSON(w, http.StatusOK, struct {
		Items []note `json:"items"`
		Total int    `json:"total"`
	}{Items: matches[start:end], Total: total})
}
```

> [!NOTE]
> `matches[start:end]` is a slice of a non nil slice, so it encodes as `[]` and not `null` even when it is empty.

</details>

## Run it for real

Add `ch23/cmd/notes/main.go`:

```go
package main

import (
	"log"
	"net/http"
	"os"

	"golearn/ch23/notesapi"
)

func main() {
	secret := os.Getenv("TOKEN_SECRET")
	if len(secret) < 16 {
		log.Fatal("TOKEN_SECRET must be set and at least 16 characters long")
	}

	handler := notesapi.NewServer(notesapi.Config{Secret: []byte(secret)})
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
```

```bash
export TOKEN_SECRET="change-me-to-something-long-and-random"
go run ./ch23/cmd/notes

curl -X POST localhost:8080/register -d '{"email":"ana@example.com","password":"password123"}'
curl -X POST localhost:8080/login    -d '{"email":"ana@example.com","password":"password123"}'

TOKEN="paste the token here"
curl -X POST localhost:8080/notes -H "Authorization: Bearer $TOKEN" -d '{"title":"Hello","body":"My first note"}'
curl "localhost:8080/notes?q=hello" -H "Authorization: Bearer $TOKEN"
```

## Stretch goals

If you want more, try these on your own. Write the test first.

1. Move the data into a real database with the repository pattern from chapter 19.
1. Add `PUT /notes/{id}` to edit a note.
1. Add the middleware from chapters 18, 20 and 22: request id, request logging, recovery, rate limiting.
1. Add graceful shutdown to `main`.
1. Replace the learning password hash with `bcrypt` from `golang.org/x/crypto`.
1. Add a `Dockerfile` and run the API in a container.

## Checkpoint

- [ ] All tests of the three milestones pass, also with `-race`
- [ ] I can explain why `404` is better than `403` for someone else's note
- [ ] I can explain why login gives the same error for an unknown email and a wrong password
- [ ] I could add a new endpoint with a test first
