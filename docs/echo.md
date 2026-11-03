# Echo v5

The HTTP server is built with **Echo v5** (`github.com/labstack/echo/v5`), chosen for its small surface area, JSON-first binding, and middleware ecosystem.

## Server bootstrap

`main.go` creates `echo.New()` and registers:

- `middleware.Recover()` — panics are recovered and turned into 500s instead of crashing the process.
- `middleware.RequestLoggerWithConfig(...)` — structured per-request log line (method, URI, status, latency) through the zap factory.

Routes:

- `GET /health` — inline handler returning `{"status":"ok"}`.
- `POST /api` — `api.NewHandler(cfg.Wasm).Execute`.

## Handler contract

Echo v5 handler signature:

```go
func(c *echo.Context) error
```

This differs from Echo v4, which returned `void`. Echo v5 also removes `HideBanner`, `HidePort`, and `e.Shutdown`; graceful shutdown is implemented with the standard library's `http.Server` and `server.Shutdown(ctx)`.

## Body binding

`c.Bind(&req)` decodes the request body based on the `Content-Type` header:

- `application/json` → JSON decoding into `ExecuteRequest`.
- `application/x-www-form-urlencoded` → form decoding (legacy path, not part of the current contract).

Bind errors produce a 400 with `"invalid request body: ..."`.

## JSON responses

All handlers respond with `c.JSON(status, payload)`. Success returns `ExecuteResponse{Success: true, Output: ...}`; failures return `ExecuteResponse{Success: false, Error: ..., Stage: ...}`.

## Graceful shutdown

```go
server := &http.Server{
    Addr:         ":" + port,
    Handler:      e,
    ReadTimeout:  cfg.Server.ReadTimeout,
    WriteTimeout: cfg.Server.WriteTimeout,
    IdleTimeout:  cfg.Server.IdleTimeout,
}

go func() { server.ListenAndServe() }()

<-ctx.Done() // SIGINT / SIGTERM
server.Shutdown(shutdownCtx)
```

This allows in-flight requests to complete before the process exits.

## v4 → v5 differences encountered during development

| v4 | v5 | Impact |
|---|---|---|
| `func(echo.Context) error` | `func(*echo.Context) error` | Pointer receiver required |
| `e.HideBanner`, `e.HidePort` | removed | Banner/port display is always on |
| `e.Shutdown(ctx)` | `server.Shutdown(ctx)` | Use stdlib `http.Server` instead |
| `e.Logger` | middleware-driven | Request logger configured explicitly |
