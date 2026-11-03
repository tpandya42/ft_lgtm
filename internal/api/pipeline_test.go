package api

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCompileToWasmUnsupportedLanguage(t *testing.T) {
	_, err := CompileToWasm(context.Background(), "rust", "fn main() {}", time.Second)
	if !errors.Is(err, ErrUnsupportedLanguage) {
		t.Fatalf("error = %v, want ErrUnsupportedLanguage", err)
	}
}

func TestCompileTimeout(t *testing.T) {
	requireWasiSDK(t)
	_, err := CompileToWasm(context.Background(), "c", "int main(void) { return 0; }", time.Nanosecond)
	if !errors.Is(err, ErrCompilationTimeout) {
		t.Fatalf("error = %v, want ErrCompilationTimeout", err)
	}
}

func TestRunWasmExitCodes(t *testing.T) {
	requirePipeline(t)

	clean, err := CompileToWasm(context.Background(), "c", "int main(void) { return 0; }", 15*time.Second)
	if err != nil {
		t.Fatalf("compile clean exit: %v", err)
	}
	if _, err := RunWasm(clean, testConfig()); err != nil {
		t.Fatalf("run clean exit: %v", err)
	}

	failing, err := CompileToWasm(context.Background(), "c", "int main(void) { return 3; }", 15*time.Second)
	if err != nil {
		t.Fatalf("compile non-zero exit: %v", err)
	}
	if _, err := RunWasm(failing, testConfig()); !errors.Is(err, ErrExecutionFailed) {
		t.Fatalf("error = %v, want ErrExecutionFailed", err)
	}
}
