# Chapter 20: Config and Logging

## What you will learn

- Reading configuration from environment variables
- Making config code testable
- Structured logging with `log/slog`
- Log levels, attributes and JSON output
- Keeping secrets out of your logs

## Key ideas

### Configuration from the environment

The 12 factor way: keep config out of the code, put it in environment variables.

```go
port := os.Getenv("PORT") // "" if not set

v, ok := os.LookupEnv("PORT") // ok tells you if it was set at all
```

Rules of thumb:

- Give **safe defaults** for things like the port.
- **Fail early** at startup if something required is missing or invalid. A crash at startup is far better than a strange bug at midnight.
- Report **all** problems at once, not one at a time.
- Never hard code secrets in source code.

### Make it testable

Do not call `os.Getenv` all over your program. Read the environment in one place and pass in
the lookup function:

```go
func Load(getenv func(string) string) (Config, error)

cfg, err := Load(os.Getenv)               // in main
cfg, err := Load(func(k string) string {  // in a test
	return map[string]string{"PORT": "9000"}[k]
})
```

### Structured logging with slog

`log/slog` writes logs as key value pairs. Machines can search them, humans can read them.

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

logger.Info("user created", "user_id", 42, "plan", "free")
// {"time":"...","level":"INFO","msg":"user created","user_id":42,"plan":"free"}
```

| Level | Use for |
|-------|---------|
| `Debug` | Details for developers, off in production |
| `Info` | Normal events: a request, a job finished |
| `Warn` | Something odd, but handled |
| `Error` | Something failed and needs attention |

Useful tricks:

```go
reqLogger := logger.With("request_id", id)  // adds fields to every line
logger.Info("hello", slog.Group("user", "id", 1, "name", "Ana"))
```

> [!WARNING]
> Never log passwords, tokens, API keys or full credit card numbers.
> Exercise 4 shows a safety net.

---

## Exercises

### Exercise 1: Load config

Write `Load(getenv func(string) string) (Config, error)`.

| Variable | Field | Rule |
|----------|-------|------|
| `PORT` | `Port` | Default `8080`. Must be a number from 1 to 65535. |
| `APP_ENV` | `Env` | Default `"development"`. Must be `development`, `staging` or `production`. |
| `DATABASE_URL` | `DatabaseURL` | **Required**. |
| `READ_TIMEOUT` | `ReadTimeout` | Default `5s`. Use `time.ParseDuration`. Must be greater than 0. |
| `DEBUG` | `Debug` | Default `false`. Use `strconv.ParseBool`. |

- An unset or empty variable uses the default (except `DATABASE_URL`).
- Every error message must contain the name of the variable.
- Collect **all** problems and return them together with `errors.Join`.

**Starter** `ch20/ex01/config.go`

```go
package config

import "time"

type Config struct {
	Port        int
	Env         string
	DatabaseURL string
	ReadTimeout time.Duration
	Debug       bool
}

func Load(getenv func(string) string) (Config, error) {
	return Config{}, nil // TODO
}
```

**Test** `ch20/ex01/config_test.go`

```go
package config

import (
	"strings"
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(key string) string { return m[key] }
}

func TestDefaults(t *testing.T) {
	cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://db/app"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8080 || cfg.Env != "development" || cfg.ReadTimeout != 5*time.Second || cfg.Debug {
		t.Errorf("wrong defaults: %+v", cfg)
	}
	if cfg.DatabaseURL != "postgres://db/app" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
}

func TestOverrides(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"PORT":         "9000",
		"APP_ENV":      "production",
		"DATABASE_URL": "postgres://prod/app",
		"READ_TIMEOUT": "250ms",
		"DEBUG":        "true",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 9000 || cfg.Env != "production" || cfg.ReadTimeout != 250*time.Millisecond || !cfg.Debug {
		t.Errorf("overrides not applied: %+v", cfg)
	}
}

func TestInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		key  string
		val  string
	}{
		{"port not a number", "PORT", "abc"},
		{"port too big", "PORT", "70000"},
		{"port zero", "PORT", "0"},
		{"unknown env", "APP_ENV", "prod"},
		{"bad duration", "READ_TIMEOUT", "soon"},
		{"zero duration", "READ_TIMEOUT", "0s"},
		{"negative duration", "READ_TIMEOUT", "-5s"},
		{"bad bool", "DEBUG", "maybe"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(env(map[string]string{
				"DATABASE_URL": "postgres://db/app",
				tc.key:         tc.val,
			}))
			if err == nil {
				t.Fatalf("expected an error for %s=%q", tc.key, tc.val)
			}
			if !strings.Contains(err.Error(), tc.key) {
				t.Errorf("error %q should mention %s", err, tc.key)
			}
		})
	}
}

func TestMissingDatabaseURL(t *testing.T) {
	_, err := Load(env(nil))
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("err = %v, want an error mentioning DATABASE_URL", err)
	}
}

func TestAllProblemsAreReported(t *testing.T) {
	_, err := Load(env(map[string]string{"PORT": "abc", "DEBUG": "nope"}))
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, name := range []string{"PORT", "DEBUG", "DATABASE_URL"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q should mention %s", err, name)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package config

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Config struct {
	Port        int
	Env         string
	DatabaseURL string
	ReadTimeout time.Duration
	Debug       bool
}

func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Port:        8080,
		Env:         "development",
		ReadTimeout: 5 * time.Second,
	}
	var errs []error

	if v := getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 || p > 65535 {
			errs = append(errs, fmt.Errorf("PORT must be a number from 1 to 65535, got %q", v))
		} else {
			cfg.Port = p
		}
	}

	if v := getenv("APP_ENV"); v != "" {
		switch v {
		case "development", "staging", "production":
			cfg.Env = v
		default:
			errs = append(errs, fmt.Errorf("APP_ENV must be development, staging or production, got %q", v))
		}
	}

	cfg.DatabaseURL = getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}

	if v := getenv("READ_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			errs = append(errs, fmt.Errorf("READ_TIMEOUT must be a positive duration like 5s, got %q", v))
		} else {
			cfg.ReadTimeout = d
		}
	}

	if v := getenv("DEBUG"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("DEBUG must be true or false, got %q", v))
		} else {
			cfg.Debug = b
		}
	}

	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
```

</details>

### Exercise 2: Build a logger

- `ParseLevel(s string) (slog.Level, error)` accepts `debug`, `info`, `warn`, `error` in any letter case. An empty string means `info`. Anything else is an error.
- `NewLogger(w io.Writer, format string, level slog.Level) (*slog.Logger, error)` returns a logger that writes to `w`. `format` is `"json"` or `"text"` (any letter case). Other values are an error. Messages below `level` must be dropped.

**Starter** `ch20/ex02/logger.go`

```go
package logger

import (
	"io"
	"log/slog"
)

func ParseLevel(s string) (slog.Level, error) {
	return slog.LevelInfo, nil // TODO
}

func NewLogger(w io.Writer, format string, level slog.Level) (*slog.Logger, error) {
	return nil, nil // TODO
}
```

**Test** `ch20/ex02/logger_test.go`

```go
package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		in   string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"Warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"", slog.LevelInfo},
	}
	for _, tc := range tests {
		got, err := ParseLevel(tc.in)
		if err != nil || got != tc.want {
			t.Errorf("ParseLevel(%q) = (%v, %v), want (%v, nil)", tc.in, got, err, tc.want)
		}
	}

	if _, err := ParseLevel("verbose"); err == nil {
		t.Error("ParseLevel(verbose) should return an error")
	}
}

func TestJSONLogger(t *testing.T) {
	var buf bytes.Buffer
	l, err := NewLogger(&buf, "json", slog.LevelInfo)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	l.Info("hello", "user_id", 42)

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("output is not JSON: %v (%q)", err, buf.String())
	}
	if m["msg"] != "hello" || m["level"] != "INFO" || m["user_id"] != float64(42) {
		t.Errorf("unexpected log line: %v", m)
	}
}

func TestTextLogger(t *testing.T) {
	var buf bytes.Buffer
	l, err := NewLogger(&buf, "TEXT", slog.LevelInfo)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	l.Info("hello", "user_id", 42)

	out := buf.String()
	if !strings.Contains(out, "msg=hello") || !strings.Contains(out, "user_id=42") {
		t.Errorf("unexpected text output: %q", out)
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l, _ := NewLogger(&buf, "json", slog.LevelWarn)

	l.Debug("hidden")
	l.Info("hidden too")
	if buf.Len() != 0 {
		t.Errorf("messages below the level were written: %q", buf.String())
	}

	l.Warn("shown")
	if !strings.Contains(buf.String(), "shown") {
		t.Error("a warning should have been written")
	}
}

func TestUnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	if _, err := NewLogger(&buf, "xml", slog.LevelInfo); err == nil {
		t.Error("expected an error for an unknown format")
	}
}
```

<details>
<summary>Solution</summary>

```go
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

func ParseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	}
	return slog.LevelInfo, fmt.Errorf("unknown log level %q", s)
}

func NewLogger(w io.Writer, format string, level slog.Level) (*slog.Logger, error) {
	opts := &slog.HandlerOptions{Level: level}
	switch strings.ToLower(format) {
	case "json":
		return slog.New(slog.NewJSONHandler(w, opts)), nil
	case "text":
		return slog.New(slog.NewTextHandler(w, opts)), nil
	}
	return nil, fmt.Errorf("unknown log format %q", format)
}
```

</details>

### Exercise 3: Request logging middleware

Write `RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler`.

After each request, write **one** log line with the message `"request"` and these attributes:

| Key | Value |
|-----|-------|
| `method` | HTTP method |
| `path` | URL path |
| `status` | Response status (default `200`) |
| `duration_ms` | How long it took, in milliseconds |

The level depends on the status: `5xx` is `Error`, `4xx` is `Warn`, everything else is `Info`.

The `statusRecorder` from chapter 18 is already in the starter.

**Starter** `ch20/ex03/reqlog.go`

```go
package reqlog

import (
	"log/slog"
	"net/http"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next // TODO
	}
}
```

**Test** `ch20/ex03/reqlog_test.go`

```go
package reqlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func logLine(t *testing.T, status int, setStatus bool) map[string]any {
	t.Helper()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	h := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if setStatus {
			w.WriteHeader(status)
		}
		fmt.Fprint(w, "body")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items", nil))

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 || lines[0] == "" {
		t.Fatalf("want exactly one log line, got %d: %q", len(lines), buf.String())
	}
	if rec.Body.String() != "body" {
		t.Errorf("response body changed: %q", rec.Body.String())
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &m); err != nil {
		t.Fatalf("log line is not JSON: %v", err)
	}
	return m
}

func TestRequestLoggerFields(t *testing.T) {
	m := logLine(t, http.StatusCreated, true)

	if m["msg"] != "request" || m["method"] != "GET" || m["path"] != "/items" {
		t.Errorf("wrong fields: %v", m)
	}
	if m["status"] != float64(201) {
		t.Errorf("status = %v, want 201", m["status"])
	}
	if _, ok := m["duration_ms"]; !ok {
		t.Error("duration_ms is missing")
	}
	if m["level"] != "INFO" {
		t.Errorf("level = %v, want INFO", m["level"])
	}
}

func TestRequestLoggerDefaultStatus(t *testing.T) {
	m := logLine(t, 0, false)
	if m["status"] != float64(200) {
		t.Errorf("status = %v, want 200", m["status"])
	}
}

func TestRequestLoggerLevels(t *testing.T) {
	tests := []struct {
		status int
		level  string
	}{
		{200, "INFO"},
		{302, "INFO"},
		{404, "WARN"},
		{429, "WARN"},
		{500, "ERROR"},
		{503, "ERROR"},
	}
	for _, tc := range tests {
		m := logLine(t, tc.status, true)
		if m["level"] != tc.level {
			t.Errorf("status %d: level = %v, want %s", tc.status, m["level"], tc.level)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package reqlog

import (
	"log/slog"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			level := slog.LevelInfo
			switch {
			case rec.status >= 500:
				level = slog.LevelError
			case rec.status >= 400:
				level = slog.LevelWarn
			}

			logger.Log(r.Context(), level, "request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
```

</details>

### Exercise 4: Redact secrets

Write `RedactSensitive(groups []string, a slog.Attr) slog.Attr`. It is a `ReplaceAttr` function for a slog handler.

- If the attribute key is `password`, `token` or `authorization` (any letter case), keep the key and replace the value with the string `"[REDACTED]"`.
- Return all other attributes unchanged.
- It must also work for attributes inside groups.

**Starter** `ch20/ex04/redact.go`

```go
package redact

import "log/slog"

func RedactSensitive(groups []string, a slog.Attr) slog.Attr {
	return a // TODO
}
```

**Test** `ch20/ex04/redact_test.go`

```go
package redact

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func newLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{
		ReplaceAttr: RedactSensitive,
	}))
}

func TestRedactsSecrets(t *testing.T) {
	var buf bytes.Buffer
	newLogger(&buf).Info("login",
		"user", "ana",
		"password", "hunter2",
		"Token", "abc123",
		"attempts", 3,
	)

	out := buf.String()
	if strings.Contains(out, "hunter2") || strings.Contains(out, "abc123") {
		t.Fatalf("secrets leaked into the log: %s", out)
	}

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["password"] != "[REDACTED]" || m["Token"] != "[REDACTED]" {
		t.Errorf("secret keys should stay with value [REDACTED]: %v", m)
	}
	if m["user"] != "ana" || m["attempts"] != float64(3) {
		t.Errorf("normal values must not change: %v", m)
	}
	if m["msg"] != "login" || m["level"] != "INFO" {
		t.Errorf("built in fields must not change: %v", m)
	}
}

func TestRedactsInsideGroups(t *testing.T) {
	var buf bytes.Buffer
	newLogger(&buf).Info("call",
		slog.Group("request",
			"path", "/me",
			"authorization", "Bearer xyz-secret",
		),
	)

	out := buf.String()
	if strings.Contains(out, "xyz-secret") {
		t.Fatalf("secret leaked from a group: %s", out)
	}
	if !strings.Contains(out, "/me") {
		t.Errorf("the normal value in the group was lost: %s", out)
	}
}
```

<details>
<summary>Solution</summary>

```go
package redact

import (
	"log/slog"
	"strings"
)

var sensitive = map[string]bool{
	"password":      true,
	"token":         true,
	"authorization": true,
}

func RedactSensitive(groups []string, a slog.Attr) slog.Attr {
	if sensitive[strings.ToLower(a.Key)] {
		return slog.String(a.Key, "[REDACTED]")
	}
	return a
}
```

</details>

## Checkpoint

- [ ] I read the environment in one place and pass values down
- [ ] My program fails at startup when config is wrong, and reports every problem
- [ ] I log with key value pairs and choose the right level
- [ ] I never log secrets
