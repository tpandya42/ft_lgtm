package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/tpandya42/ft_lgtm/internal/config"
)

func testConfig() config.WasmConfig {
	return config.WasmConfig{
		MemoryLimitMB:  10,
		FuelLimit:      5_000_000,
		Timeout:        5 * time.Second,
		OutputLimitKB:  10,
		CompileTimeout: 15 * time.Second,
	}
}

func newTestEcho(cfg config.WasmConfig) *echo.Echo {
	e := echo.New()
	e.POST("/api", NewHandler(cfg).Execute)
	return e
}

func postJSON(t *testing.T, e *echo.Echo, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func jsonBody(t *testing.T, req ExecuteRequest) string {
	t.Helper()
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return string(b)
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder) ExecuteResponse {
	t.Helper()
	var resp ExecuteResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}
	return resp
}

func requireWasiSDK(t *testing.T) {
	t.Helper()
	p := compilerPath("clang")
	if !filepath.IsAbs(p) {
		t.Skip("wasi-sdk not found: set WASI_SDK_PATH or install to ~/wasi-sdk")
	}
	if _, err := os.Stat(p); err != nil {
		t.Skipf("wasi-sdk clang unavailable: %v", err)
	}
}

func requirePipeline(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping compile/run test in -short mode")
	}
	requireWasiSDK(t)
}

func TestErrorStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"unsupported language", fmt.Errorf("%w: %q", ErrUnsupportedLanguage, "rust"), http.StatusBadRequest},
		{"compilation failed", fmt.Errorf("%w: boom", ErrCompilationFailed), http.StatusUnprocessableEntity},
		{"compilation timeout", ErrCompilationTimeout, http.StatusUnprocessableEntity},
		{"execution failed", fmt.Errorf("%w: boom", ErrExecutionFailed), http.StatusUnprocessableEntity},
		{"internal", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errorStatus(tt.err); got != tt.want {
				t.Fatalf("errorStatus() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCompilerFor(t *testing.T) {
	tests := []struct {
		lang     string
		compiler string
		ext      string
	}{
		{"c", "clang", ".c"},
		{"cpp", "clang++", ".cpp"},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			compiler, ext, err := compilerFor(tt.lang)
			if err != nil {
				t.Fatalf("compilerFor(%q): %v", tt.lang, err)
			}
			if compiler != tt.compiler || ext != tt.ext {
				t.Fatalf("got (%q, %q), want (%q, %q)", compiler, ext, tt.compiler, tt.ext)
			}
		})
	}

	if _, _, err := compilerFor("rust"); !errors.Is(err, ErrUnsupportedLanguage) {
		t.Fatalf("compilerFor(rust) error = %v, want ErrUnsupportedLanguage", err)
	}
}

func TestExecuteValidation(t *testing.T) {
	e := newTestEcho(testConfig())

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantErr    string
	}{
		{"malformed json", `{bad`, http.StatusBadRequest, "invalid request body"},
		{"empty body", ``, http.StatusBadRequest, "code"},
		{"missing code", `{"language":"c"}`, http.StatusBadRequest, "code"},
		{"unsupported language", `{"code":"int main(void){return 0;}","language":"rust"}`, http.StatusBadRequest, "unsupported language"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := postJSON(t, e, tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			resp := decodeResponse(t, rec)
			if resp.Success {
				t.Fatal("success = true, want false")
			}
			if !strings.Contains(resp.Error, tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", resp.Error, tt.wantErr)
			}
			if resp.Stage != "" {
				t.Fatalf("stage = %q, want empty for 400 responses", resp.Stage)
			}
		})
	}
}

func TestExecutePipeline(t *testing.T) {
	const (
		helloC   = "#include <stdio.h>\nint main(void) { printf(\"Hello from WASM!\\n\"); return 0; }\n"
		helloCpp = "#include <iostream>\nint main() { std::cout << \"hi from c++\" << std::endl; return 0; }\n"
		badC     = "int main(void) { return 0 }\n"
		spinC    = "int main(void) { for (;;) {} }\n"
		exit3C   = "int main(void) { return 3; }\n"
		bigC     = "#include <stdio.h>\nint main(void) { for (int i = 0; i < 2000; i++) putchar('A'); return 0; }\n"
	)

	tests := []struct {
		name       string
		req        ExecuteRequest
		mutate     func(*config.WasmConfig)
		wantStatus int
		wantOutput string
		wantStage  string
		wantErr    string
		check      func(*testing.T, ExecuteResponse)
	}{
		{"c hello", ExecuteRequest{Code: helloC, Language: "c"}, nil, http.StatusOK, "Hello from WASM!", "", "", nil},
		{"cpp hello", ExecuteRequest{Code: helloCpp, Language: "cpp"}, nil, http.StatusOK, "hi from c++", "", "", nil},
		{"compile error", ExecuteRequest{Code: badC, Language: "c"}, nil, http.StatusUnprocessableEntity, "", "compilation", "compilation failed", nil},
		{"fuel exhausted", ExecuteRequest{Code: spinC, Language: "c"}, func(c *config.WasmConfig) { c.FuelLimit = 1000 }, http.StatusUnprocessableEntity, "", "execution", "execution failed", nil},
		{"non-zero exit", ExecuteRequest{Code: exit3C, Language: "c"}, nil, http.StatusUnprocessableEntity, "", "execution", "execution failed", nil},
		{"output truncation", ExecuteRequest{Code: bigC, Language: "c"}, func(c *config.WasmConfig) { c.OutputLimitKB = 1 }, http.StatusOK, "AAAA", "", "", func(t *testing.T, resp ExecuteResponse) {
			if len(resp.Output) != 1024 {
				t.Fatalf("output length = %d, want 1024", len(resp.Output))
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requirePipeline(t)
			cfg := testConfig()
			if tt.mutate != nil {
				tt.mutate(&cfg)
			}
			rec := postJSON(t, newTestEcho(cfg), jsonBody(t, tt.req))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			resp := decodeResponse(t, rec)
			if tt.wantOutput != "" && !strings.Contains(resp.Output, tt.wantOutput) {
				t.Fatalf("output = %q, want substring %q", resp.Output, tt.wantOutput)
			}
			if resp.Stage != tt.wantStage {
				t.Fatalf("stage = %q, want %q", resp.Stage, tt.wantStage)
			}
			if tt.wantErr != "" && !strings.Contains(resp.Error, tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", resp.Error, tt.wantErr)
			}
			if tt.wantStatus == http.StatusOK && !resp.Success {
				t.Fatal("success = false, want true")
			}
			if tt.check != nil {
				tt.check(t, resp)
			}
		})
	}
}
