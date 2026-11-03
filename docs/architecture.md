# Architecture

LGTM is a small, single-service Go application that accepts source code over HTTP, compiles it to WebAssembly (WASM), executes it inside a Wasmtime sandbox, and returns the output as JSON.

## System context

```mermaid
flowchart LR
    Client((Client / frontend)) -->|HTTP POST /api| Backend[Go backend on port 1222]
    Backend --> Clang[(wasi-sdk clang/clang++)]
    Backend --> Wasmtime[(wasmtime-go runtime)]
    Wasmtime --> WASI[WASI syscalls via files]
    Clang --> Temp[(temp dir: main.c / main.cpp)]
```

## Components

| Component | Path | Responsibility |
|---|---|---|
| Entry point | `cmd/server/main.go` | Config load, Echo bootstrap, graceful shutdown via `net/http.Server` |
| Handler | `internal/api/handler.go` | JSON contract, compile pipeline, execute pipeline |
| DTOs | `internal/api/dto.go` | `ExecuteRequest`, `ExecuteResponse` |
| Configuration | `internal/config/config.go` | viper loader with env prefix `LGTM` |
| Logging | `internal/observability/logger.go` | zap factory (not wired into main yet; main uses `zap.NewNop()`) |

## Request lifecycle

```mermaid
sequenceDiagram
    participant C as Client
    participant E as Echo server
    participant H as Handler.Execute
    participant Cl as Clang (wasi-sdk)
    participant W as wasmtime-go

    C->>E: POST /api JSON {code, language}
    E->>H: Execute(c)
    H->>H: validate (code non-empty, language c|cpp)
    H->>Cl: CompileToWasm (temp dir, -O2, -fno-exceptions for C++)
    Cl-->>H: main.wasm or compile error
    H->>W: RunWasm (fuel, memory, WASI files)
    W-->>H: stdout or trap / exit error
    H-->>E: JSON {success, output} or {error, stage}
    E-->>C: HTTP response
```

## Failure modes and status codes

| Failure | Status | `stage` field |
|---|---|---|
| Malformed JSON | 400 | empty |
| Empty `code` | 400 | empty |
| Unsupported `language` | 400 | empty |
| Compilation failure / timeout | 422 | `compilation` |
| Execution trap / non-zero exit | 422 | `execution` |
| Internal error (temp dir, file I/O) | 500 | empty |

Sentinel errors (`ErrCompilationFailed`, `ErrCompilationTimeout`, `ErrExecutionFailed`, `ErrUnsupportedLanguage`) are wrapped with `%w` so the HTTP layer classifies them with `errors.Is` rather than string matching.

## Current vs target topology

Today the backend runs as a single binary on the developer machine (`localhost:1222`). The target topology (not yet implemented) deploys it into k3s on the Ubuntu VM behind an ingress with domain `lgtm.local`, and adds IPFS storage and the observability stack. Those components are intentionally absent from this document until implemented.
