# Go

## Toolchain

- Module: `github.com/tpandya42/ft_lgtm`
- Go version: `1.27.2` (darwin/arm64)
- Dependency manager: native Go modules (`go.mod`, `go.sum`)

## Package layout

```
cmd/server/   # application entry point (main package)
internal/api/ # HTTP handler, DTOs, compile/execute pipeline, tests
internal/config/   # viper configuration loader
internal/observability/  # zap logger factory
```

`cmd/` contains only `main`; all library code lives under `internal/`, which cannot be imported by external modules — appropriate for an application (not a library) project.

## Core idioms used

| Idiom | Where | Purpose |
|---|---|---|
| Struct tags | `config.go`, `dto.go` | mapstructure decoding, JSON serialization |
| Error wrapping with `%w` | `handler.go` | Enables `errors.Is` classification of sentinel errors |
| `errors.Is` / `errors.As` | `handler.go` | Status mapping and wasmtime trap inspection |
| `context.Context` | `handler.go` | Compile deadline bound to the HTTP request context |
| `defer` cleanup | `handler.go` | Temp dir removal, cancel functions |
| `zap.NewNop()` | `main.go` | Disabled structured logger until wiring is added |

## Concurrency

The only goroutine beyond the main thread is `server.ListenAndServe()`. Shutdown is driven by `signal.NotifyContext` and `server.Shutdown(ctx)` — no shared mutable state, no locks required.

## Toolchain commands

| Command | Purpose |
|---|---|
| `go build ./...` | Compile all packages |
| `go vet ./...` | Static analysis |
| `gofmt -w .` | Formatting |
| `go test ./...` | Run tests |
| `go mod tidy` | Sync `go.mod`/`go.sum` |
| `go mod verify` | Verify downloaded modules |

CI runs `go mod tidy` + `git diff` to ensure `go.mod`/`go.sum` are always in sync.
