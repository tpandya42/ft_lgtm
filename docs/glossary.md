# Glossary

| Term | Meaning |
|---|---|
| WASM | WebAssembly — a portable, stack-based binary instruction format. LGTM compiles C/C++ to WASM before executing it. |
| WASI | WebAssembly System Interface — a system interface for WASM modules, providing file, clock, and randomness access through the host. |
| wasm32-wasip1 | The WebAssembly target triple used by wasi-sdk (32-bit WASM, WASI preview 1). |
| sysroot | The directory containing headers and libraries for a cross-compiler target. Apple clang lacks a WASI sysroot, which is why wasi-sdk is required. |
| Fuel | A wasmtime metering mechanism: each executed WASM instruction consumes fuel; exhaustion traps the module. |
| Linear memory | WASM's contiguous, growable byte array, shared between module and host. |
| Trap | An abnormal WASM termination (e.g. fuel exhaustion, out-of-bounds memory), distinct from a clean `proc_exit`. |
| proc_exit | The WASI function a module calls to terminate with an exit code. Code 0 is success; anything else is treated as an execution failure in LGTM. |
| _start | The WASM entry function that wasmtime invokes to begin program execution. |
| CID | Content Identifier — an IPFS concept for content-addressed storage. Reserved in config (`ipfs.*` keys) but not implemented. |
| OTel | OpenTelemetry — the observability framework for traces, metrics, and logs. Reserved in config (`otel.*` keys) but not implemented. |
| LGTM stack | The project's planned observability combination: Loki (logs), Grafana (dashboards), Tempo (traces), Prometheus (metrics). Not implemented. |
| k3s | A lightweight Kubernetes distribution. Installed by `scripts/bootstrap-lgtm-vm.sh` but not yet used to deploy the backend. |
| Sentinel error | A package-level `errors.New` value (e.g. `ErrCompilationFailed`) matched by callers with `errors.Is`. |
