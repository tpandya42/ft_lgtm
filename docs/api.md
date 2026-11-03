# API reference

Base URL: `http://127.0.0.1:1222` (configurable via `server.port`).

## Endpoints

### `POST /api`

Accepts C or C++ source code, compiles it to WASM, executes it in a Wasmtime sandbox, and returns the program's stdout.

**Request** — `application/json`

```json
{
  "code":     "<source code>",
  "language": "c" | "cpp"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `code` | string | Yes | C or C++ source code |
| `language` | string | Yes | Either `c` or `cpp` |

**Success response** — `200 OK`

```json
{
  "success": true,
  "output": "program stdout text"
}
```

**Error response**

```json
{
  "success": false,
  "error":   "<message>",
  "stage":   "<compilation|execution>"
}
```

The `stage` field is present only for pipeline failures (422). For 400 validation failures it is empty.

**Status codes**

| Code | Meaning | Body |
|---|---|---|
| 200 | Success | `{ success: true, output }` |
| 400 | Invalid request body / empty code / unsupported language | `{ success: false, error }` |
| 422 | Compilation failure/timeout, or execution failure/trap | `{ success: false, error, stage }` |
| 500 | Internal error (temp dir, file I/O, wasmtime setup) | `{ success: false, error }` |

**Examples**

Valid C:

```bash
curl -sS -X POST http://127.0.0.1:1222/api \
  -H 'Content-Type: application/json' \
  -d '{"code":"#include <stdio.h>\nint main(void){ printf(\"hi\\n\"); return 0; }","language":"c"}'
```
```
{"success":true,"output":"hi\n"} HTTP 200
```

Unsupported language:

```bash
curl -sS -X POST http://127.0.0.1:1222/api \
  -H 'Content-Type: application/json' \
  -d '{"code":"int main(void){return 0;}","language":"rust"}'
```
```
{"success":false,"error":"unsupported language: \"rust\" (use \"c\" or \"cpp\")"} HTTP 400
```

Compile error:

```bash
curl -sS -X POST http://127.0.0.1:1222/api \
  -H 'Content-Type: application/json' \
  -d '{"code":"int main(void){ return 0 }","language":"c"}'
```
```
{"success":false,"error":"compilation failed: ...main.c:2:32: error: expected ')'...","stage":"compilation"} HTTP 422
```

### `GET /health`

Returns `{"status":"ok"}` with `200 OK`.

```bash
curl -sS http://127.0.0.1:1222/health
```

## Error envelope semantics

- `success` is always present on error responses (`false`).
- `output` is omitted on errors (`omitempty`).
- `error` is omitted on success (`omitempty`).
- `stage` is `compilation` or `execution` on 422; empty on 400.

Errors leaking from the compiler may contain temporary directory paths; this is acceptable today for local development but should be sanitized before any production exposure (see `docs/security.md`).
