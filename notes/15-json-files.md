# Chapter 15: JSON and Files

## What you will learn

- Turning structs into JSON and back
- Struct tags: rename, `omitempty`, skip
- Working with `io.Reader` and `io.Writer`
- Reading and writing files
- Testing file code with `t.TempDir()`

## Key ideas

### Struct tags

Only **exported** fields (capital letter) are included. Tags control the JSON name.

```go
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"` // left out when empty
	Password string `json:"-"`               // never included
}
```

### Marshal and Unmarshal

```go
data, err := json.Marshal(user)      // struct to []byte
err = json.Unmarshal(data, &user)    // []byte to struct, needs a pointer
```

JSON numbers become `float64` when you decode into `any`. Decode into a struct when you can.

### Optional values

A pointer field can tell the difference between "missing" and "zero".

```go
type Profile struct {
	Name string `json:"name"`
	Age  *int   `json:"age"` // nil when missing or null
}
```

### Readers and writers

`io.Reader` and `io.Writer` are everywhere: files, HTTP bodies, buffers, strings.
Functions that take a reader or writer are easy to test, because you can pass
`strings.NewReader("...")` or a `bytes.Buffer` instead of a real file.

```go
dec := json.NewDecoder(r)            // read JSON from a reader
dec.DisallowUnknownFields()          // error on fields you did not expect
err := dec.Decode(&v)

err = json.NewEncoder(w).Encode(v)   // write JSON to a writer, adds a newline
```

### Files

```go
data, err := os.ReadFile("a.json")                // read the whole file
err = os.WriteFile("a.json", data, 0o644)         // write the whole file
```

A missing file gives an error you can check with `errors.Is(err, os.ErrNotExist)`.

For big files, read line by line:

```go
sc := bufio.NewScanner(f)
for sc.Scan() {
	line := sc.Text()
}
if err := sc.Err(); err != nil { /* ... */ }
```

---

## Exercises

### Exercise 1: Struct tags

Define `User` with these fields **in this order** and these JSON rules:

| Field | Type | JSON |
|-------|------|------|
| `ID` | `int` | `"id"` |
| `Name` | `string` | `"name"` |
| `Email` | `string` | `"email"`, left out when empty |
| `Password` | `string` | never in JSON |

Then write `ToJSON(u User) ([]byte, error)` and `FromJSON(data []byte) (User, error)`.

**Starter** `ch15/ex01/user.go`

```go
package user

type User struct {
	// TODO
}

func ToJSON(u User) ([]byte, error) {
	return nil, nil // TODO
}

func FromJSON(data []byte) (User, error) {
	return User{}, nil // TODO
}
```

**Test** `ch15/ex01/user_test.go`

```go
package user

import "testing"

func TestToJSON(t *testing.T) {
	data, err := ToJSON(User{ID: 1, Name: "Ana", Password: "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := string(data), `{"id":1,"name":"Ana"}`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestToJSONWithEmail(t *testing.T) {
	data, err := ToJSON(User{ID: 2, Name: "Bo", Email: "bo@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"id":2,"name":"Bo","email":"bo@example.com"}`
	if string(data) != want {
		t.Errorf("got %s, want %s", data, want)
	}
}

func TestFromJSON(t *testing.T) {
	in := `{"id":3,"name":"Cy","email":"c@x.io","password":"hacked"}`
	u, err := FromJSON([]byte(in))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != 3 || u.Name != "Cy" || u.Email != "c@x.io" {
		t.Errorf("got %+v", u)
	}
	if u.Password != "" {
		t.Errorf("Password should never be read from JSON, got %q", u.Password)
	}
}

func TestFromJSONInvalid(t *testing.T) {
	if _, err := FromJSON([]byte(`{"id": `)); err == nil {
		t.Error("expected an error for invalid JSON")
	}
}
```

<details>
<summary>Solution</summary>

```go
package user

import "encoding/json"

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	Password string `json:"-"`
}

func ToJSON(u User) ([]byte, error) {
	return json.Marshal(u)
}

func FromJSON(data []byte) (User, error) {
	var u User
	if err := json.Unmarshal(data, &u); err != nil {
		return User{}, err
	}
	return u, nil
}
```

</details>

### Exercise 2: Parse an order from a reader

The types are given. Write:

- `(o Order) Total() int` is the sum of `Qty * PriceCents` of all items.
- `ParseOrder(r io.Reader) (Order, error)` decodes JSON from the reader. It must **reject unknown fields**.

**Starter** `ch15/ex02/order.go`

```go
package order

import "io"

type Item struct {
	Name       string `json:"name"`
	Qty        int    `json:"qty"`
	PriceCents int    `json:"price_cents"`
}

type Order struct {
	ID    int    `json:"id"`
	Items []Item `json:"items"`
}

func (o Order) Total() int {
	return 0 // TODO
}

func ParseOrder(r io.Reader) (Order, error) {
	return Order{}, nil // TODO
}
```

**Test** `ch15/ex02/order_test.go`

```go
package order

import (
	"strings"
	"testing"
)

const validOrder = `{"id":7,"items":[
	{"name":"pen","qty":3,"price_cents":150},
	{"name":"book","qty":1,"price_cents":1200}
]}`

func TestParseOrder(t *testing.T) {
	o, err := ParseOrder(strings.NewReader(validOrder))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.ID != 7 || len(o.Items) != 2 {
		t.Fatalf("got %+v", o)
	}
	if o.Items[0].Name != "pen" || o.Items[1].PriceCents != 1200 {
		t.Errorf("items parsed wrong: %+v", o.Items)
	}
}

func TestTotal(t *testing.T) {
	o, err := ParseOrder(strings.NewReader(validOrder))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := o.Total(); got != 1650 {
		t.Errorf("Total = %d, want 1650", got)
	}
	if got := (Order{}).Total(); got != 0 {
		t.Errorf("Total of empty order = %d, want 0", got)
	}
}

func TestParseOrderUnknownField(t *testing.T) {
	_, err := ParseOrder(strings.NewReader(`{"id":1,"items":[],"coupon":"FREE"}`))
	if err == nil {
		t.Error("expected an error for an unknown field")
	}
}

func TestParseOrderInvalid(t *testing.T) {
	if _, err := ParseOrder(strings.NewReader(`not json`)); err == nil {
		t.Error("expected an error for invalid JSON")
	}
}
```

<details>
<summary>Solution</summary>

```go
package order

import (
	"encoding/json"
	"io"
)

type Item struct {
	Name       string `json:"name"`
	Qty        int    `json:"qty"`
	PriceCents int    `json:"price_cents"`
}

type Order struct {
	ID    int    `json:"id"`
	Items []Item `json:"items"`
}

func (o Order) Total() int {
	total := 0
	for _, it := range o.Items {
		total += it.Qty * it.PriceCents
	}
	return total
}

func ParseOrder(r io.Reader) (Order, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o Order
	if err := dec.Decode(&o); err != nil {
		return Order{}, err
	}
	return o, nil
}
```

</details>

### Exercise 3: Read text from any reader

- `CountLines(r io.Reader) (int, error)` counts the lines.
- `WordFrequency(r io.Reader) (map[string]int, error)` counts words. Make words lower case and remove punctuation at the start and end of each word. Skip empty words.

> [!TIP]
> `bufio.NewScanner(r)` reads lines. Call `sc.Split(bufio.ScanWords)` to read words.
> `strings.TrimFunc` can remove characters that are not letters or digits.

**Starter** `ch15/ex03/text.go`

```go
package text

import "io"

func CountLines(r io.Reader) (int, error) {
	return 0, nil // TODO
}

func WordFrequency(r io.Reader) (map[string]int, error) {
	return nil, nil // TODO
}
```

**Test** `ch15/ex03/text_test.go`

```go
package text

import (
	"reflect"
	"strings"
	"testing"
)

func TestCountLines(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"one line", "hello", 1},
		{"three lines", "a\nb\nc", 3},
		{"trailing newline", "a\nb\n", 2},
		{"blank line counts", "a\n\nb", 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CountLines(strings.NewReader(tc.in))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("CountLines = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestWordFrequency(t *testing.T) {
	got, err := WordFrequency(strings.NewReader("Go is fun. Go, go!\nIs it? \"Yes\""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]int{"go": 3, "is": 2, "fun": 1, "it": 1, "yes": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestWordFrequencyEmpty(t *testing.T) {
	got, err := WordFrequency(strings.NewReader("  ...  !!! "))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want an empty result", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package text

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

func CountLines(r io.Reader) (int, error) {
	sc := bufio.NewScanner(r)
	n := 0
	for sc.Scan() {
		n++
	}
	return n, sc.Err()
}

func WordFrequency(r io.Reader) (map[string]int, error) {
	sc := bufio.NewScanner(r)
	sc.Split(bufio.ScanWords)

	counts := make(map[string]int)
	for sc.Scan() {
		w := strings.TrimFunc(strings.ToLower(sc.Text()), func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})
		if w != "" {
			counts[w]++
		}
	}
	return counts, sc.Err()
}
```

</details>

### Exercise 4: Save and load JSON files

- `SaveJSON(path string, v any) error` writes `v` as indented JSON (2 spaces) with a final newline. Use permission `0o644`.
- `LoadJSON(path string, v any) error` reads the file and decodes it into `v` (a pointer). If reading fails, **wrap** the error with `%w` so `errors.Is(err, os.ErrNotExist)` still works.

**Starter** `ch15/ex04/store.go`

```go
package store

func SaveJSON(path string, v any) error {
	return nil // TODO
}

func LoadJSON(path string, v any) error {
	return nil // TODO
}
```

**Test** `ch15/ex04/store_test.go`

```go
package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type settings struct {
	Name string
	Port int
	Tags []string
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	in := settings{Name: "api", Port: 8080, Tags: []string{"a", "b"}}

	if err := SaveJSON(path, in); err != nil {
		t.Fatalf("SaveJSON: %v", err)
	}

	var out settings
	if err := LoadJSON(path, &out); err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}
	if out.Name != in.Name || out.Port != in.Port || len(out.Tags) != 2 || out.Tags[1] != "b" {
		t.Errorf("round trip gave %+v, want %+v", out, in)
	}
}

func TestSaveIsIndented(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.json")
	if err := SaveJSON(path, settings{Name: "x"}); err != nil {
		t.Fatalf("SaveJSON: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "\n  \"Name\": \"x\"") {
		t.Errorf("file is not indented with 2 spaces:\n%s", text)
	}
	if !strings.HasSuffix(text, "\n") {
		t.Error("file should end with a newline")
	}
}

func TestLoadMissingFile(t *testing.T) {
	var out settings
	err := LoadJSON(filepath.Join(t.TempDir(), "nope.json"), &out)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("err = %v, want it to wrap os.ErrNotExist", err)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{oops"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out settings
	if err := LoadJSON(path, &out); err == nil {
		t.Error("expected an error for invalid JSON")
	}
}
```

<details>
<summary>Solution</summary>

```go
package store

import (
	"encoding/json"
	"fmt"
	"os"
)

func SaveJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func LoadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("load %s: %w", path, err)
	}
	return json.Unmarshal(data, v)
}
```

</details>

### Exercise 5: Write JSON to any writer

Write `WriteJSON(w io.Writer, v any) error` with a `json.Encoder`.

- A map `{"a": 1}` must produce exactly `{"a":1}` followed by a newline.
- Errors from the writer or from encoding must be returned.

This is the same pattern you will use to send JSON in HTTP responses.

**Starter** `ch15/ex05/write.go`

```go
package write

import "io"

func WriteJSON(w io.Writer, v any) error {
	return nil // TODO
}
```

**Test** `ch15/ex05/write_test.go`

```go
package write

import (
	"bytes"
	"errors"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, map[string]int{"a": 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := buf.String(), "{\"a\":1}\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

type brokenWriter struct{}

func (brokenWriter) Write(p []byte) (int, error) {
	return 0, errors.New("disk full")
}

func TestWriteJSONWriterError(t *testing.T) {
	if err := WriteJSON(brokenWriter{}, map[string]int{"a": 1}); err == nil {
		t.Error("expected the writer error to be returned")
	}
}

func TestWriteJSONEncodeError(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, make(chan int)); err == nil {
		t.Error("channels cannot be encoded, expected an error")
	}
}
```

<details>
<summary>Solution</summary>

```go
package write

import (
	"encoding/json"
	"io"
)

func WriteJSON(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}
```

</details>

## Checkpoint

- [ ] I can control JSON names with struct tags
- [ ] I prefer functions that take `io.Reader` or `io.Writer`
- [ ] I use `t.TempDir()` for tests that need files
