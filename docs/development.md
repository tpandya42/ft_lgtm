# Development

## Setup

1. Clone the repository.
2. Install Go 1.27.2.
3. Install wasi-sdk 34 at `~/wasi-sdk` or set `WASI_SDK_PATH`.
4. Copy configuration: `cp config.yaml.example config.yaml`.
5. Build: `make build` (produces `bin/lgtm-backend`).
6. Run: `./bin/lgtm-backend` or `make run`.
7. Test: `make test` or `make test-short`.

## Make targets

| Target | Command |
|---|---|
| `make build` | `go build -o ./bin/lgtm-backend ./cmd/server` |
| `make run` | `go run ./cmd/server/main.go` |
| `make test` | `go test ./... -count=1` |
| `make test-short` | `go test ./... -short -count=1` |
| `make coverage` | `go test ./... -coverprofile=coverage.out` + `go tool cover -func=coverage.out` |
| `make format` | `gofmt -w .` |
| `make lint` | `golangci-lint run ./...` |
| `make tidy` | `go mod tidy && go mod verify` |
| `make clean` | `rm -rf ./bin coverage.out` |

## Debugging compile/run failures

1. Confirm wasi-sdk clang exists: `ls ~/wasi-sdk/bin/clang` (or `echo $WASI_SDK_PATH`).
2. Confirm the language matches the file: `c` uses clang/`.c`, `cpp` uses clang++/`.cpp`.
3. Compile errors return clang stderr with `stage:"compilation"`; check the message for syntax or missing-sysroot errors.
4. Fuel exhaustion returns `stage:"execution"` — lower the fuel limit in tests to reproduce quickly (`LGTM_WASM_FUEL_LIMIT=1000`).
5. Output longer than `OutputLimitKB` is truncated; raise `LGTM_WASM_OUTPUT_LIMIT_KB` to see full output.

## Common errors

| Error | Meaning |
|---|---|
| `unsupported language: "rust"` | `language` must be `c` or `cpp` |
| `field "code" must not be empty` | `code` missing or empty |
| `invalid request body` | Body is not valid JSON or `Content-Type` is not `application/json` |
| `compilation failed: ... expected ')'` | Source has a syntax error; clang stderr is included |
| `compilation timeout` | Clang exceeded `wasm.compile_timeout` |
| `execution failed: all fuel consumed` | Infinite loop or fuel-heavy code |
| `port already in use` | Another process is listening on `server.port` |

## Branch and PR conventions

- Work on feature branches off `main`.
- Open a pull request using `.github/pull_request_template.md`.
- Before requesting review: `go build ./...`, `go vet ./...`, `gofmt -w .`, `go test ./... -short`, and update documentation if behavior changed.
- CI runs deps, lint, and build; tests must pass locally before merge.
