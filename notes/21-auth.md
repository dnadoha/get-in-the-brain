# Chapter 21: Authentication

## What you will learn

- Authentication versus authorization
- Storing passwords safely (salt, slow hash, constant time compare)
- Signed tokens with HMAC
- A middleware that protects routes
- Simple role based rules

## Key ideas

### Two different questions

| Term | Question | Failure code |
|------|----------|--------------|
| **Authentication** | Who are you? | `401 Unauthorized` |
| **Authorization** | Are you allowed to do this? | `403 Forbidden` |

### Passwords

Never store a password. Store a **hash** that is made for passwords:

- A **random salt** per password, so two equal passwords give different hashes.
- A **slow** algorithm, so guessing millions of passwords is expensive.
- A **constant time** comparison when you check it.

> [!WARNING]
> In a real project use `bcrypt` or `argon2` from `golang.org/x/crypto`. Do not invent your own.
> Exercise 1 builds a small version from standard library parts **only so you can see the ideas**.
> It is for learning, not for production.

```go
// the real thing (needs: go get golang.org/x/crypto/bcrypt)
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
err = bcrypt.CompareHashAndPassword(hash, []byte(password)) // nil means match
```

### Tokens

After login, the server gives the client a **token**. The client sends it with every request:

```text
Authorization: Bearer <token>
```

A signed token has a payload (who, until when) and a signature made with a secret that only the server knows.
If anyone changes the payload, the signature no longer fits.

```text
base64(payload) . base64(HMAC-SHA256(payload, secret))
```

This is the idea behind JWT. In production, use a well tested JWT library and keep these rules:

- Keep the secret out of source code (use environment config).
- Always set an expiry time.
- Use HTTPS, otherwise tokens can be stolen on the network.
- Give the same error for "unknown user" and "wrong password", so attackers cannot find out which emails exist.

---

## Exercises

### Exercise 1: Password hashing (learning version)

Write two functions.

`HashPassword(password string) (string, error)`:

- An empty password returns `ErrEmptyPassword`.
- Make a random 16 byte salt with `crypto/rand`.
- Digest: `sha256(salt + password)`, then repeat `sha256(previous digest + salt)` until you have done `iterations` rounds in total (the constant is `10000`).
- Return the text `v1$<iterations>$<salt in hex>$<digest in hex>`.

`CheckPassword(password, encoded string) bool`:

- Parse the four parts. Anything malformed returns `false` (no panic).
- Recompute the digest with the stored salt and iterations.
- Compare with `subtle.ConstantTimeCompare`.

**Starter** `ch21/ex01/password.go`

```go
package password

import "errors"

var ErrEmptyPassword = errors.New("password is empty")

const iterations = 10000

func HashPassword(password string) (string, error) {
	return "", nil // TODO
}

func CheckPassword(password, encoded string) bool {
	return false // TODO
}
```

**Test** `ch21/ex01/password_test.go`

```go
package password

import (
	"errors"
	"strings"
	"testing"
)

func TestHashAndCheck(t *testing.T) {
	hash, err := HashPassword("correct horse")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !CheckPassword("correct horse", hash) {
		t.Error("the right password was rejected")
	}
	if CheckPassword("wrong horse", hash) {
		t.Error("a wrong password was accepted")
	}
	if CheckPassword("", hash) {
		t.Error("an empty password was accepted")
	}
}

func TestHashFormat(t *testing.T) {
	hash, _ := HashPassword("secret")
	parts := strings.Split(hash, "$")
	if len(parts) != 4 || parts[0] != "v1" || parts[1] != "10000" {
		t.Errorf("format = %q, want v1$10000$<salt>$<digest>", hash)
	}
	if len(parts[2]) != 32 || len(parts[3]) != 64 {
		t.Errorf("salt hex length %d (want 32), digest hex length %d (want 64)", len(parts[2]), len(parts[3]))
	}
	if strings.Contains(hash, "secret") {
		t.Error("the hash must not contain the password")
	}
}

func TestSaltMakesHashesDifferent(t *testing.T) {
	a, _ := HashPassword("same password")
	b, _ := HashPassword("same password")
	if a == b {
		t.Error("two hashes of the same password must differ (random salt)")
	}
	if !CheckPassword("same password", a) || !CheckPassword("same password", b) {
		t.Error("both hashes should verify")
	}
}

func TestEmptyPassword(t *testing.T) {
	if _, err := HashPassword(""); !errors.Is(err, ErrEmptyPassword) {
		t.Errorf("err = %v, want ErrEmptyPassword", err)
	}
}

func TestMalformedHashes(t *testing.T) {
	bad := []string{
		"",
		"abc",
		"v1$10000$zz$zz",
		"v2$10000$aabb$ccdd",
		"v1$notanumber$aabb$ccdd",
		"v1$0$aabb$ccdd",
		"v1$10000$aabb",
		"v1$10000$aabb$ccdd$extra",
	}
	for _, h := range bad {
		if CheckPassword("anything", h) {
			t.Errorf("CheckPassword accepted the malformed hash %q", h)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package password

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrEmptyPassword = errors.New("password is empty")

const iterations = 10000

func derive(password string, salt []byte, rounds int) []byte {
	sum := sha256.Sum256(append(append([]byte{}, salt...), password...))
	for i := 1; i < rounds; i++ {
		sum = sha256.Sum256(append(sum[:], salt...))
	}
	return sum[:]
}

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	digest := derive(password, salt, iterations)
	return fmt.Sprintf("v1$%d$%s$%s", iterations, hex.EncodeToString(salt), hex.EncodeToString(digest)), nil
}

func CheckPassword(password, encoded string) bool {
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
	got := derive(password, salt, rounds)
	return subtle.ConstantTimeCompare(got, want) == 1
}
```

</details>

### Exercise 2: Signed tokens

Write `Sign` and `Verify`.

`Sign(c Claims, secret []byte) (string, error)`:

1. Turn `c` into JSON.
1. `payload` is the JSON encoded with `base64.RawURLEncoding`.
1. `signature` is `HMAC-SHA256(payload, secret)` encoded the same way.
1. The token is `payload + "." + signature`.

`Verify(token string, secret []byte, now time.Time) (Claims, error)`:

- Not exactly two parts, bad base64, bad JSON, or a signature that does not match: `ErrInvalidToken`.
- Check the **signature first**, before you trust anything in the payload. Use `hmac.Equal`.
- If `now` is at or after `ExpiresAt` (Unix seconds): `ErrExpiredToken`.
- Otherwise return the claims.

The `now` parameter makes the function easy to test, you do not need `time.Sleep`.

**Starter** `ch21/ex02/token.go`

```go
package token

import (
	"errors"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	UserID    int64 `json:"uid"`
	ExpiresAt int64 `json:"exp"` // Unix seconds
}

func Sign(c Claims, secret []byte) (string, error) {
	return "", nil // TODO
}

func Verify(token string, secret []byte, now time.Time) (Claims, error) {
	return Claims{}, nil // TODO
}
```

**Test** `ch21/ex02/token_test.go`

```go
package token

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"
)

var (
	secret = []byte("test-secret")
	now    = time.Unix(1_700_000_000, 0)
)

func validClaims() Claims {
	return Claims{UserID: 7, ExpiresAt: now.Add(time.Hour).Unix()}
}

func TestRoundTrip(t *testing.T) {
	tok, err := Sign(validClaims(), secret)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if strings.Count(tok, ".") != 1 {
		t.Errorf("token %q should have exactly one dot", tok)
	}

	got, err := Verify(tok, secret, now)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if got != validClaims() {
		t.Errorf("claims = %+v, want %+v", got, validClaims())
	}
}

func TestWrongSecret(t *testing.T) {
	tok, _ := Sign(validClaims(), secret)
	if _, err := Verify(tok, []byte("other-secret"), now); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestTamperedPayload(t *testing.T) {
	tok, _ := Sign(validClaims(), secret)
	signature := strings.Split(tok, ".")[1]

	forgedPayload := base64.RawURLEncoding.EncodeToString([]byte(`{"uid":1,"exp":9999999999}`))
	forged := forgedPayload + "." + signature

	if _, err := Verify(forged, secret, now); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken (the signature must cover the payload)", err)
	}
}

func TestMalformedTokens(t *testing.T) {
	for _, tok := range []string{"", "abc", "a.b.c", ".", "!!!.???"} {
		if _, err := Verify(tok, secret, now); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("Verify(%q) err = %v, want ErrInvalidToken", tok, err)
		}
	}
}

func TestExpired(t *testing.T) {
	tok, _ := Sign(validClaims(), secret)

	if _, err := Verify(tok, secret, now.Add(2*time.Hour)); !errors.Is(err, ErrExpiredToken) {
		t.Errorf("err = %v, want ErrExpiredToken", err)
	}
}

func TestExpiryBoundary(t *testing.T) {
	c := Claims{UserID: 1, ExpiresAt: now.Unix()}
	tok, _ := Sign(c, secret)

	if _, err := Verify(tok, secret, now); !errors.Is(err, ErrExpiredToken) {
		t.Errorf("at the exact expiry time err = %v, want ErrExpiredToken", err)
	}
	if _, err := Verify(tok, secret, now.Add(-time.Second)); err != nil {
		t.Errorf("one second before expiry err = %v, want nil", err)
	}
}
```

<details>
<summary>Solution</summary>

```go
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	UserID    int64 `json:"uid"`
	ExpiresAt int64 `json:"exp"`
}

func mac(payload string, secret []byte) []byte {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(payload))
	return m.Sum(nil)
}

func Sign(c Claims, secret []byte) (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(data)
	sig := base64.RawURLEncoding.EncodeToString(mac(payload, secret))
	return payload + "." + sig, nil
}

func Verify(token string, secret []byte, now time.Time) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return Claims{}, ErrInvalidToken
	}

	gotSig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(gotSig, mac(parts[0], secret)) {
		return Claims{}, ErrInvalidToken
	}

	data, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var c Claims
	if err := json.Unmarshal(data, &c); err != nil {
		return Claims{}, ErrInvalidToken
	}

	if now.Unix() >= c.ExpiresAt {
		return Claims{}, ErrExpiredToken
	}
	return c, nil
}
```

</details>

### Exercise 3: Authentication middleware

Now protect routes. This exercise lives in the **same package** as exercise 2.

1. Create the folder `ch21/ex03/` and copy your `token.go` from exercise 2 into it (package `token`).
1. Add `auth.go` below.

`Authenticate(secret []byte, now func() time.Time) func(http.Handler) http.Handler`:

- Read the header `Authorization: Bearer <token>`.
- Missing header, wrong scheme, empty token, or an invalid token: `401` and `{"error":"unauthorized"}`.
- Expired token: `401` and `{"error":"token expired"}`.
- Valid token: put the user id in the request context and call the next handler.
- On any failure the next handler must **not** run.

`UserIDFrom(ctx) (int64, bool)` reads the user id back out of the context.

**Starter** `ch21/ex03/auth.go`

```go
package token

import (
	"context"
	"net/http"
	"time"
)

func Authenticate(secret []byte, now func() time.Time) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next // TODO
	}
}

func UserIDFrom(ctx context.Context) (int64, bool) {
	return 0, false // TODO
}
```

**Test** `ch21/ex03/auth_test.go`

```go
package token

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	testSecret = []byte("middleware-secret")
	fixedNow   = time.Unix(1_700_000_000, 0)
)

func clock() time.Time { return fixedNow }

func makeToken(t *testing.T, secret []byte, expires time.Time) string {
	t.Helper()
	tok, err := Sign(Claims{UserID: 7, ExpiresAt: expires.Unix()}, secret)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	return tok
}

func call(header string) (rec *httptest.ResponseRecorder, reached bool) {
	h := Authenticate(testSecret, clock)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		id, _ := UserIDFrom(r.Context())
		fmt.Fprintf(w, "user %d", id)
	}))

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec, reached
}

func errorOf(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var m map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("body is not JSON: %v (%q)", err, rec.Body.String())
	}
	return m["error"]
}

func TestValidToken(t *testing.T) {
	tok := makeToken(t, testSecret, fixedNow.Add(time.Hour))
	rec, reached := call("Bearer " + tok)

	if !reached || rec.Code != http.StatusOK || rec.Body.String() != "user 7" {
		t.Errorf("reached=%v code=%d body=%q, want true 200 %q", reached, rec.Code, rec.Body.String(), "user 7")
	}
}

func TestRejected(t *testing.T) {
	good := makeToken(t, testSecret, fixedNow.Add(time.Hour))
	wrongSecret := makeToken(t, []byte("other"), fixedNow.Add(time.Hour))

	tests := []struct {
		name   string
		header string
	}{
		{"no header", ""},
		{"wrong scheme", "Basic abc123"},
		{"scheme without token", "Bearer "},
		{"garbage token", "Bearer not-a-token"},
		{"wrong secret", "Bearer " + wrongSecret},
		{"token without scheme", good},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec, reached := call(tc.header)
			if reached {
				t.Error("the handler must not run")
			}
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", rec.Code)
			}
			if msg := errorOf(t, rec); msg != "unauthorized" {
				t.Errorf("error = %q, want %q", msg, "unauthorized")
			}
		})
	}
}

func TestExpiredToken(t *testing.T) {
	tok := makeToken(t, testSecret, fixedNow.Add(-time.Minute))
	rec, reached := call("Bearer " + tok)

	if reached {
		t.Error("the handler must not run")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
	if msg := errorOf(t, rec); msg != "token expired" {
		t.Errorf("error = %q, want %q", msg, "token expired")
	}
}

func TestUserIDFromEmptyContext(t *testing.T) {
	if id, ok := UserIDFrom(context.Background()); ok || id != 0 {
		t.Errorf("got (%d, %v), want (0, false)", id, ok)
	}
}
```

<details>
<summary>Solution (auth.go)</summary>

```go
package token

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type ctxKey struct{}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func Authenticate(secret []byte, now func() time.Time) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || raw == "" {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			claims, err := Verify(raw, secret, now())
			if errors.Is(err, ErrExpiredToken) {
				writeError(w, http.StatusUnauthorized, "token expired")
				return
			}
			if err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := context.WithValue(r.Context(), ctxKey{}, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFrom(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(ctxKey{}).(int64)
	return id, ok
}
```

</details>

### Exercise 4: Authorization rules

Authentication tells you **who**. Authorization decides **what they may do**. Keep these rules in plain functions.
They are easy to read and easy to test.

| Rule | Who |
|------|-----|
| `CanEdit` on an unlocked post | the author, any moderator, any admin |
| `CanEdit` on a locked post | admins only |
| `CanDelete` | the author (even if locked) or any admin. Moderators can **not** delete. |

**Starter** `ch21/ex04/policy.go`

```go
package policy

type Role string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

type User struct {
	ID   int64
	Role Role
}

type Post struct {
	AuthorID int64
	Locked   bool
}

func CanEdit(u User, p Post) bool {
	return false // TODO
}

func CanDelete(u User, p Post) bool {
	return false // TODO
}
```

**Test** `ch21/ex04/policy_test.go`

```go
package policy

import "testing"

var (
	author    = User{ID: 1, Role: RoleUser}
	stranger  = User{ID: 2, Role: RoleUser}
	moderator = User{ID: 3, Role: RoleModerator}
	admin     = User{ID: 4, Role: RoleAdmin}
)

func TestCanEdit(t *testing.T) {
	open := Post{AuthorID: 1}
	locked := Post{AuthorID: 1, Locked: true}

	tests := []struct {
		name string
		user User
		post Post
		want bool
	}{
		{"author edits open post", author, open, true},
		{"stranger edits open post", stranger, open, false},
		{"moderator edits open post", moderator, open, true},
		{"admin edits open post", admin, open, true},
		{"author edits locked post", author, locked, false},
		{"stranger edits locked post", stranger, locked, false},
		{"moderator edits locked post", moderator, locked, false},
		{"admin edits locked post", admin, locked, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanEdit(tc.user, tc.post); got != tc.want {
				t.Errorf("CanEdit = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCanDelete(t *testing.T) {
	open := Post{AuthorID: 1}
	locked := Post{AuthorID: 1, Locked: true}

	tests := []struct {
		name string
		user User
		post Post
		want bool
	}{
		{"author deletes own post", author, open, true},
		{"author deletes own locked post", author, locked, true},
		{"stranger deletes post", stranger, open, false},
		{"moderator deletes post", moderator, open, false},
		{"admin deletes post", admin, open, true},
		{"admin deletes locked post", admin, locked, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanDelete(tc.user, tc.post); got != tc.want {
				t.Errorf("CanDelete = %v, want %v", got, tc.want)
			}
		})
	}
}
```

<details>
<summary>Solution</summary>

```go
func CanEdit(u User, p Post) bool {
	if u.Role == RoleAdmin {
		return true
	}
	if p.Locked {
		return false
	}
	return u.Role == RoleModerator || u.ID == p.AuthorID
}

func CanDelete(u User, p Post) bool {
	return u.Role == RoleAdmin || u.ID == p.AuthorID
}
```

</details>

## Checkpoint

- [ ] I can explain the difference between `401` and `403`
- [ ] I know why passwords need a salt and a slow hash, and that I should use bcrypt or argon2 in real code
- [ ] I understand why a token signature protects the payload
- [ ] My middleware stops the request on failure and never calls `next`
