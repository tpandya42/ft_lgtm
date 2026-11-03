package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	wasmtime "github.com/bytecodealliance/wasmtime-go/v49"
	"github.com/labstack/echo/v5"
	"github.com/tpandya42/ft_lgtm/internal/config"
)

var (
	ErrCompilationTimeout  = errors.New("compilation timeout")
	ErrCompilationFailed   = errors.New("compilation failed")
	ErrExecutionFailed     = errors.New("execution failed")
	ErrUnsupportedLanguage = errors.New("unsupported language")
)

type Handler struct {
	Wasm config.WasmConfig
}

func NewHandler(wasm config.WasmConfig) *Handler {
	return &Handler{Wasm: wasm}
}

func (h *Handler) Execute(c *echo.Context) error {
	var req ExecuteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ExecuteResponse{
			Error: "invalid request body: " + err.Error(),
		})
	}

	if req.Code == "" {
		return c.JSON(http.StatusBadRequest, ExecuteResponse{
			Error: `field "code" must not be empty`,
		})
	}

	wasmBinary, err := CompileToWasm(c.Request().Context(), req.Language, req.Code, h.Wasm.CompileTimeout)
	if err != nil {
		return writePipelineError(c, err, "compilation")
	}

	result, err := RunWasm(wasmBinary, h.Wasm)
	if err != nil {
		return writePipelineError(c, err, "execution")
	}
	return c.JSON(http.StatusOK, ExecuteResponse{
		Success: true,
		Output:  result,
	})
}

func writePipelineError(c *echo.Context, err error, stage string) error {
	status := errorStatus(err)
	res := ExecuteResponse{Error: err.Error()}
	if status != http.StatusBadRequest {
		res.Stage = stage
	}
	return c.JSON(status, res)
}

func errorStatus(err error) int {
	switch {
	case errors.Is(err, ErrUnsupportedLanguage):
		return http.StatusBadRequest
	case errors.Is(err, ErrCompilationFailed),
		errors.Is(err, ErrCompilationTimeout),
		errors.Is(err, ErrExecutionFailed):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func compilerFor(lang string) (string, string, error) {
	switch lang {
	case "c":
		return "clang", ".c", nil
	case "cpp":
		return "clang++", ".cpp", nil
	default:
		return "", "", fmt.Errorf("%w: %q (use \"c\" or \"cpp\")", ErrUnsupportedLanguage, lang)
	}
}

func compilerPath(name string) string {
	if sdk := os.Getenv("WASI_SDK_PATH"); sdk != "" {
		return filepath.Join(sdk, "bin", name)
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidate := filepath.Join(home, "wasi-sdk", "bin", name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return name
}

func CompileToWasm(ctx context.Context, lang string, code string, timeout time.Duration) ([]byte, error) {
	compiler, ext, err := compilerFor(lang)
	if err != nil {
		return nil, err
	}

	dir, err := os.MkdirTemp("", "lgtm-compile-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	sourcePath := filepath.Join(dir, "main"+ext)
	outputPath := filepath.Join(dir, "main.wasm")

	if err := os.WriteFile(sourcePath, []byte(code), 0600); err != nil {
		return nil, err
	}

	compileCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := []string{"-O2", "-o", outputPath, sourcePath}
	if lang == "cpp" {
		args = append([]string{"-fno-exceptions"}, args...)
	}

	cmd := exec.CommandContext(compileCtx, compilerPath(compiler), args...)
	cmd.Dir = dir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(compileCtx.Err(), context.DeadlineExceeded) {
			return nil, ErrCompilationTimeout
		}
		return nil, fmt.Errorf("%w: %s", ErrCompilationFailed, stderr.String())
	}

	wasm, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("read wasm output: %w", err)
	}

	if len(wasm) == 0 {
		return nil, errors.New("empty wasm output")
	}

	return wasm, nil
}

func RunWasm(wasmByte []byte, wasmCfg config.WasmConfig) (string, error) {
	dir, err := os.MkdirTemp("", "lgtm-run-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	stdoutPath := filepath.Join(dir, "stdout.txt")
	stderrPath := filepath.Join(dir, "stderr.txt")

	engineConfig := wasmtime.NewConfig()
	engineConfig.SetConsumeFuel(true)
	engine := wasmtime.NewEngineWithConfig(engineConfig)

	store := wasmtime.NewStore(engine)
	memoryBytes := int64(wasmCfg.MemoryLimitMB) * 1024 * 1024
	store.Limiter(memoryBytes, -1, -1, -1, -1)
	if err := store.SetFuel(wasmCfg.FuelLimit); err != nil {
		return "", err
	}

	wasi := wasmtime.NewWasiConfig()
	wasi.SetArgv([]string{"program"})
	if err := wasi.SetStdoutFile(stdoutPath); err != nil {
		return "", err
	}
	if err := wasi.SetStderrFile(stderrPath); err != nil {
		return "", err
	}
	store.SetWasi(wasi)

	module, err := wasmtime.NewModule(engine, wasmByte)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrExecutionFailed, err)
	}

	linker := wasmtime.NewLinker(engine)
	if err := linker.DefineWasi(); err != nil {
		return "", err
	}

	instance, err := linker.Instantiate(store, module)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrExecutionFailed, err)
	}

	start := instance.GetFunc(store, "_start")
	if start == nil {
		return "", errors.New("module has no _start export")
	}

	if _, err := start.Call(store); err != nil && !isCleanExit(err) {
		return "", fmt.Errorf("%w: %v", ErrExecutionFailed, err)
	}

	stdout, err := os.ReadFile(stdoutPath)
	if err != nil {
		return "", err
	}

	outputLimitBytes := wasmCfg.OutputLimitKB * 1024
	if len(stdout) > outputLimitBytes {
		stdout = stdout[:outputLimitBytes]
	}

	return string(stdout), nil
}

func isCleanExit(err error) bool {
	var wasmErr *wasmtime.Error
	if errors.As(err, &wasmErr) {
		status, ok := wasmErr.ExitStatus()
		return ok && status == 0
	}
	return false
}
