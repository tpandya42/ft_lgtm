# wasi-sdk (clang / clang++)

## What wasi-sdk provides

wasi-sdk bundles a clang/clang++ cross-compiler targeting `wasm32-wasip1`, plus a WASI sysroot (libc, libc++, headers). The default target is `wasm32-wasip1`, so no `--target` flag is required.

## Why Apple clang is not used

The system clang bundled with Xcode lacks a WASI sysroot and cannot produce `wasm32-wasip1` binaries. wasi-sdk must be installed separately.

## Compiler resolution order (`compilerPath`)

1. `WASI_SDK_PATH` environment variable → `$WASI_SDK_PATH/bin/clang(++)`
2. `~/wasi-sdk/bin/clang(++)`
3. Fallback to bare `clang(++)` on PATH (will fail to build WASI)

## Compilation flags

| Flag | When | Purpose |
|---|---|---|
| `-O2` | Always | Optimize WASM output |
| `-fno-exceptions` | C++ only | wasi-sdk libc++ is built without exceptions; undefined symbols otherwise |
| `-o main.wasm` | Always | Output WASM module path |

## Temporary directory lifecycle

```mermaid
flowchart TD
    Start[CompileToWasm called] --> Mkdir[os.MkdirTemp("lgtm-compile-*")]
    Mkdir --> Write[Write main.c / main.cpp]
    Write --> Clang[Run clang subprocess with timeout]
    Clang --> Read[Read main.wasm]
    Read --> Remove[defer os.RemoveAll(dir)]
```

The temp directory (source + output WASM) is removed with `defer os.RemoveAll(dir)` after compilation, regardless of success or failure.

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `clang: error: unable to execute...` | wasi-sdk not found | Set `WASI_SDK_PATH` or install to `~/wasi-sdk` |
| `undefined reference to __cxa_throw` | C++ compiled without `-fno-exceptions` | Add `-fno-exceptions` for C++ |
| `main.wasm: No such file` | Compilation failed | Check clang stderr, returned with `stage:"compilation"` |

## Links

- wasi-sdk releases: official GitHub repository for the WebAssembly System Interface SDK
- WASI target: `wasm32-wasip1`
