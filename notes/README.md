# Golang Backend Development: Exercises and Tests

A beginner friendly path from your first `Hello, Go` to a small REST API with tests.
Each chapter has short notes and **incremental exercises**. Every exercise comes with a
starter file, a test file, and a hidden solution.

> [!NOTE]
> I could not open the video list of the playlist (79 videos), so the chapters follow the
> usual beginner to backend path for Go. If you paste the video titles, I can match the
> chapters to them one by one.

## Before you start

You need Go 1.22 or newer.

```bash
go version
```

Create one module for the whole course:

```bash
mkdir golearn && cd golearn
go mod init golearn
```

Every exercise lives in its own folder, for example `ch04/ex02/`.
Create the folder, paste the **Starter** file and the **Test** file, then run:

```bash
go test ./ch04/ex02/
```

## How each exercise works

1. Read the task.
1. Copy the starter file and the test file into the folder shown.
1. Run `go test`. The test fails at first. That is normal.
1. Edit the starter file until the test passes.
1. Only then open the solution and compare.

Useful commands you will use all the time:

| Command | What it does |
|---------|--------------|
| `go test ./...` | Run every test in the module |
| `go test -v ./ch04/ex02/` | Run one folder with details |
| `go test -race ./...` | Run tests and look for data races |
| `go test -cover ./...` | Show test coverage |
| `go vet ./...` | Find common mistakes |
| `gofmt -w .` | Format all files |

## Chapters

| # | Chapter | You will practice |
|---|---------|-------------------|
| 01 | [Getting started](./01-getting-started.md) | Modules, `go run`, your first test |
| 02 | [Variables, types and constants](./02-variables-types-constants.md) | Types, conversion, `iota` |
| 03 | [Control flow](./03-control-flow.md) | `if`, `switch`, `for` |
| 04 | [Functions](./04-functions.md) | Multiple returns, variadic, closures, `defer` |
| 05 | [Arrays, slices and maps](./05-collections.md) | Slices, maps, copying |
| 06 | [Strings and runes](./06-strings-runes.md) | Unicode, `strings`, `strings.Builder` |
| 07 | [Structs and methods](./07-structs-methods.md) | Types with behavior, embedding |
| 08 | [Pointers](./08-pointers.md) | Pointers, pointer receivers, linked list |
| 09 | [Interfaces](./09-interfaces.md) | Interfaces, type switch, fakes |
| 10 | [Error handling](./10-errors.md) | Custom errors, wrapping, `errors.Is/As` |
| 11 | [Packages and testing](./11-packages-testing.md) | Table tests, subtests, benchmarks |
| 12 | [Generics](./12-generics.md) | Type parameters and constraints |
| 13 | [Goroutines and channels](./13-goroutines-channels.md) | Concurrency basics, worker pool |
| 14 | [Sync, select and context](./14-sync-select-context.md) | Mutex, timeouts, cancellation |
| 15 | [JSON and files](./15-json-files.md) | `encoding/json`, `io.Reader`, temp files |
| 16 | [HTTP basics](./16-http-basics.md) | Handlers, `httptest`, status codes |
| 17 | [REST API with CRUD](./17-rest-crud.md) | Routing, in-memory store, validation |
| 18 | [Middleware](./18-middleware.md) | Logging, recover, auth, chaining |
| 19 | [Database and repository pattern](./19-database-repository.md) | Repository, service layer, `database/sql` |
| 20 | [Config and logging](./20-config-logging.md) | Environment config, `log/slog` |
| 21 | [Authentication](./21-auth.md) | Password hashing, signed tokens |
| 22 | [Graceful shutdown and rate limit](./22-shutdown-ratelimit.md) | `http.Server`, signals, token bucket |
| 23 | [Final project: Notes API](./23-final-project.md) | Put everything together |

## A note on the tests

The code in these chapters uses only the Go standard library, so nothing needs to be
downloaded. I wrote the tests carefully, but I was not able to run them in my environment.
If a test looks wrong to you, trust the task text and tell me which one failed. I will fix it.
