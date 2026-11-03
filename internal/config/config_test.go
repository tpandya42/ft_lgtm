package config

import (
	"testing"
	"time"
)

func TestLoadDefaultsAndEnvOverrides(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("LGTM_WASM_FUEL_LIMIT", "12345")
	t.Setenv("LGTM_WASM_OUTPUT_LIMIT_KB", "1")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != 1222 {
		t.Fatalf("port = %d, want 1222", cfg.Server.Port)
	}
	if cfg.Wasm.FuelLimit != 12345 {
		t.Fatalf("fuel = %d, want 12345", cfg.Wasm.FuelLimit)
	}
	if cfg.Wasm.OutputLimitKB != 1 {
		t.Fatalf("output_limit_kb = %d, want 1", cfg.Wasm.OutputLimitKB)
	}
	if cfg.Wasm.CompileTimeout != 15*time.Second {
		t.Fatalf("compile_timeout = %v, want 15s", cfg.Wasm.CompileTimeout)
	}
	if cfg.Otel.Enabled {
		t.Fatal("otel.enabled = true, want false")
	}
}
