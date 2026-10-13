# Chapter 06: Strings and Runes

## What you will learn

- Strings are read only bytes
- Runes and Unicode
- Useful functions in `strings` and `unicode`
- Building text fast with `strings.Builder`

## Key ideas

### Bytes and runes

A Go string is a list of **bytes** in UTF-8. One character can use more than one byte.

```go
s := "héllo"
fmt.Println(len(s))         // 6, bytes
fmt.Println(len([]rune(s))) // 5, characters
```

A `rune` is one Unicode character (an alias for `int32`).

```go
for i, r := range "héllo" {
	fmt.Println(i, string(r)) // range gives runes, i is the byte position
}
```

> [!WARNING]
> `s[i]` gives you a **byte**, not a character. Use `[]rune(s)` when you need characters.

### Handy helpers

```go
strings.ToUpper("go")              // "GO"
strings.Contains("golang", "lang") // true
strings.Split("a,b,c", ",")        // ["a" "b" "c"]
strings.Fields("  a   b ")         // ["a" "b"], splits on any spaces
strings.Join([]string{"a", "b"}, "-") // "a-b"
unicode.IsLetter('x')              // true
```

### strings.Builder

Strings cannot change, so `s += "x"` in a loop creates many copies. Use a builder.

```go
var b strings.Builder
for i := 0; i < 3; i++ {
	b.WriteString("go")
}
fmt.Println(b.String()) // "gogogo"
```

---

## Exercises

### Exercise 1: Reverse a string

Write `Reverse(s string) string`. It must work with Unicode.

**Starter** `ch06/ex01/reverse.go`

```go
package reverse

func Reverse(s string) string {
	return "" // TODO
}
```

**Test** `ch06/ex01/reverse_test.go`

```go
package reverse

import "testing"

func TestReverse(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", ""},
		{"a", "a"},
		{"hello", "olleh"},
		{"héllo", "olléh"},
		{"日本語", "語本日"},
	}
	for _, tc := range tests {
		if got := Reverse(tc.in); got != tc.want {
			t.Errorf("Reverse(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package reverse

func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
```

</details>

### Exercise 2: Palindrome

Write `IsPalindrome(s string) bool`. Ignore upper or lower case. Ignore everything that is not a letter or a digit.

> [!TIP]
> `unicode.IsLetter`, `unicode.IsDigit` and `unicode.ToLower` work on a single rune.

**Starter** `ch06/ex02/palindrome.go`

```go
package palindrome

func IsPalindrome(s string) bool {
	return false // TODO
}
```

**Test** `ch06/ex02/palindrome_test.go`

```go
package palindrome

import "testing"

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"", true},
		{"a", true},
		{"racecar", true},
		{"Racecar", true},
		{"A man, a plan, a canal: Panama", true},
		{"No lemon, no melon", true},
		{"hello", false},
		{"ab", false},
		{"12321", true},
		{"12345", false},
	}
	for _, tc := range tests {
		if got := IsPalindrome(tc.in); got != tc.want {
			t.Errorf("IsPalindrome(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package palindrome

import "unicode"

func IsPalindrome(s string) bool {
	var clean []rune
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			clean = append(clean, unicode.ToLower(r))
		}
	}
	for i, j := 0, len(clean)-1; i < j; i, j = i+1, j-1 {
		if clean[i] != clean[j] {
			return false
		}
	}
	return true
}
```

</details>

### Exercise 3: Title case

Write `TitleCase(s string) string`.

- Split the text into words (any amount of spaces between them).
- Make the first letter of each word upper case and the rest lower case.
- Join the words with exactly one space.

**Starter** `ch06/ex03/title.go`

```go
package title

func TitleCase(s string) string {
	return "" // TODO
}
```

**Test** `ch06/ex03/title_test.go`

```go
package title

import "testing"

func TestTitleCase(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", ""},
		{"go", "Go"},
		{"hello world", "Hello World"},
		{"  hello   wORLD  ", "Hello World"},
		{"the quick BROWN fox", "The Quick Brown Fox"},
		{"élan vital", "Élan Vital"},
	}
	for _, tc := range tests {
		if got := TitleCase(tc.in); got != tc.want {
			t.Errorf("TitleCase(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package title

import (
	"strings"
	"unicode"
)

func TitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		r := []rune(strings.ToLower(w))
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}
```

</details>

### Exercise 4: Caesar cipher

Write `Caesar(s string, shift int) string`.

- Move every letter `shift` places forward in the alphabet. After `z` comes `a` again.
- Keep upper and lower case.
- Leave all other characters (spaces, digits, punctuation) as they are.
- The shift can be negative or bigger than 26.
- Use a `strings.Builder`.

**Starter** `ch06/ex04/caesar.go`

```go
package caesar

func Caesar(s string, shift int) string {
	return "" // TODO
}
```

**Test** `ch06/ex04/caesar_test.go`

```go
package caesar

import "testing"

func TestCaesar(t *testing.T) {
	tests := []struct {
		in    string
		shift int
		want  string
	}{
		{"abc", 1, "bcd"},
		{"xyz", 3, "abc"},
		{"Hello, World!", 13, "Uryyb, Jbeyq!"},
		{"bcd", -1, "abc"},
		{"abc", -1, "zab"},
		{"abc", 27, "bcd"},
		{"abc", 0, "abc"},
		{"Go 1.22", 5, "Lt 1.22"},
	}
	for _, tc := range tests {
		if got := Caesar(tc.in, tc.shift); got != tc.want {
			t.Errorf("Caesar(%q, %d) = %q, want %q", tc.in, tc.shift, got, tc.want)
		}
	}
}

func TestCaesarRoundTrip(t *testing.T) {
	msg := "The Quick Brown Fox, 123!"
	if got := Caesar(Caesar(msg, 7), -7); got != msg {
		t.Errorf("round trip = %q, want %q", got, msg)
	}
}
```

<details>
<summary>Solution</summary>

```go
package caesar

import "strings"

func Caesar(s string, shift int) string {
	shift = ((shift % 26) + 26) % 26
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune('a' + (r-'a'+rune(shift))%26)
		case r >= 'A' && r <= 'Z':
			b.WriteRune('A' + (r-'A'+rune(shift))%26)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
```

</details>

## Checkpoint

- [ ] I know `len(s)` counts bytes, not characters
- [ ] I can loop over the characters of a string with `range`
- [ ] I use `strings.Builder` when I build text in a loop
