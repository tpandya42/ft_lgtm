package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/tpandya42/ft_lgtm/internal/api"
	"github.com/tpandya42/ft_lgtm/internal/config"
	"go.uber.org/zap"
)

func main() {
	// Loading Config
	// step 1: start the server on specific port (e.g. 1222)
	cfg, err := config.Load()
	if err != nil {
		// Fallback to basic logger for startup errors
		basicLogger, _ := zap.NewProduction()
		basicLogger.Fatal("config load failed", zap.Error(err))
	}

	// Initialize Logger
	logger := zap.NewNop() // no-op until enabled

	// Creating Echo Instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request",
				zap.String("method", v.Method),
				zap.String("uri", v.URI),
				zap.Int("status", v.Status),
				zap.String("latency", v.Latency.String()),
			)
			return nil
		},
	}))

	// step 2: Define Routes on which user uploads source code
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.POST("/api", func(c *echo.Context) error {
		code := c.FormValue("code")
		lang := c.FormValue("lang")

		// step 3: Try to compile the source code into wasm.
		wasmBinary, err := api.CompileToWasm(c.Request().Context(), lang, code)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
				"stage": "compilation",
			})
		}

		// step 4: try to run wasm binary in wasm runtime (monitor how much memory it takes)
		result, err := api.RunWasm(wasmBinary)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
				"stage": "execution",
			})
		}

		return c.String(http.StatusOK, result)
	})

	// step 5: return output of binary on success else code specific error message

	// Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      e,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server start failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("shutdown failed", zap.Error(err))
	}
}
