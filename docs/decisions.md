# Architecture decision records

Decisions already made in the codebase, recorded for future maintainers.

## ADR-001: C and C++ as the first supported languages

**Status:** Accepted

**Context:** The 42 LGTM subject targets a code playground. The team chose C/C++ because clang targets `wasm32-wasip1` cleanly through wasi-sdk, and both languages are common in the 42 curriculum.

**Decision:** Support `c` (clang, `.c`) and `cpp` (clang++, `.cpp`) only. All other languages return 400.

**Consequences:** Simple toolchain, well-understood sandboxing. Adding Rust, Go, or Zig later requires their own WASI toolchains and exception/ABI handling.

## ADR-002: wasmtime-go over wasmer-go

**Status:** Accepted

**Context:** Early prototypes considered wasmer-go. wasmtime-go v49 was chosen for its active maintenance, fuel metering, memory limiter, and clean `*wasmtime.Error` exit-status inspection.

**Decision:** Use `github.com/bytecodealliance/wasmtime-go/v49`.

**Consequences:** Wasmtime's embedder API is lower-level than some alternatives; trap vs clean-exit distinction requires explicit handling (`isCleanExit`).

## ADR-003: Echo v5

**Status:** Accepted

**Context:** A minimal Go HTTP framework was needed for the JSON API.

**Decision:** Use `github.com/labstack/echo/v5`.

**Consequences:** v5 handler signature returns `error`; graceful shutdown uses the stdlib `http.Server` rather than `e.Shutdown`. No banner/port hiding options.

## ADR-004: viper for configuration

**Status:** Accepted

**Decision:** Use `github.com/spf13/viper` with env prefix `LGTM` and `.` → `_` key replacer.

**Consequences:** Env overrides work (`LGTM_WASM_FUEL_LIMIT`), defaults are centralized, and `config.yaml` is gitignored with `config.yaml.example` committed. CLI flags are not yet implemented.

## ADR-005: Handler struct over inline closures

**Status:** Accepted

**Decision:** `internal/api/handler.go` exposes `Handler` with `NewHandler(cfg.Wasm)` and an `Execute` method. Config is injected at construction.

**Consequences:** Tests can construct isolated handlers with custom limits; main.go stays thin. Config fields flow through to compile/execute limits.

## ADR-006: HTTP status semantics

**Status:** Accepted

**Decision:** 400 for invalid request shape, 422 for user-code failures (compile/execution), 500 only for internal errors. Sentinel errors wrapped with `%w` and classified via `errors.Is`.

**Consequences:** Clients can distinguish "you sent something invalid" from "your code failed to compile/run".

## ADR-007: Sandbox limits (fuel, memory, output)

**Status:** Accepted

**Decision:** Fuel = 5,000,000, memory = 10 MiB, output = 10 KB, compile timeout = 15 s, all configurable.

**Consequences:** Untrusted code is bounded. A wall-clock execution timeout is not yet implemented (see `docs/security.md`).

## ADR-008: zap deferred

**Status:** Accepted

**Decision:** `internal/observability/logger.go` exists; `main.go` uses `zap.NewNop()` until structured logging is wired.

**Consequences:** No log output today beyond the no-op; Echo request logger middleware is configured but discards entries.

## ADR-009: External wasi-sdk toolchain

**Status:** Accepted

**Decision:** clang/clang++ come from wasi-sdk (default target `wasm32-wasip1`), resolved via `WASI_SDK_PATH` → `~/wasi-sdk` → PATH.

**Consequences:** Apple clang cannot be used (no WASI sysroot). C++ requires `-fno-exceptions`.

## ADR-010: JSON API replaces form values

**Status:** Accepted

**Decision:** `POST /api` accepts `application/json` with fields `code` and `language`.

**Consequences:** Cleaner contract for a frontend client; form-encoded requests are no longer supported.

## ADR-011: Tests before IPFS

**Status:** Accepted

**Decision:** The test suite (unit, contract, integration) was implemented before IPFS, OTel, or k3s work begins.

**Consequences:** Refactoring for future features is covered by tests; CI gap (no test job) remains known.
