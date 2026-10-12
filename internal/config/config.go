/*
configuration flow (viper)

1. CLI Flags (--port=8080)
2. Environment Variables
3. Config file
4. Defaults in code
*/
package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Wasm   WasmConfig   `mapstructure:"wasm"`
	IPFS   IPFSConfig   `mapstructure:"ipfs"`
	Otel   OtelConfig   `mapstructure:"otel"`
}

type ServerConfig struct {
	Port	int 	`mapstructure:"port"`
	ReadTimeout		time.Duration	`mapstructure:"read_timeout"`
	WriteTimeout	time.Duration	`mapstructure:"write_timeout"`
	IdleTimeout		time.Duration	`mapstructure:"idle_timeout"`
}

type WasmConfig struct {
	MemoryLimitMB 	int		`mapstructure:"memory_limit_mb"`
	FuelLimit		uint64  `mapstructure:"fuel_limit"`
	Timeout			time.Duration	`mapstructure:"timeout"`
	OutputLimitKB	int				`mapstructure:"output_limit_kb"`
}

type IPFSConfig struct {
	APIAddress string `mapstructure:"api_address"`
	GatewayURL string `mapstructure:"gateway_url"`
}

type OtelConfig struct {
	ServiceName      string `mapstructure:"service_name"`
	ExporterEndpoint string `mapstructure:"exporter_endpoint"`
	Enabled          bool   `mapstructure:"enabled"`
}

func Load() (*Config, error) {
	v := viper.New()

	// Config File
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// Env vars: LGTM_SERVER_PORT, LGTM_WASM_MEMORY_LIMIT_MB, etc.
	v.SetEnvPrefix("LGTM")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults
	v.SetDefault("server.port", 1222)
	v.SetDefault("server.read_timeout", "10s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "60s")
	v.SetDefault("wasm.memory_limit_mb", 10)
	v.SetDefault("wasm.fuel_limit", 5000000)
	v.SetDefault("wasm.timeout", "5s")
	v.SetDefault("wasm.output_limit_kb", 10)
	v.SetDefault("ipfs.api_address", "/ip4/127.0.0.1/tcp/5001")
	v.SetDefault("ipfs.gateway_url", "http://ipfs.lgtm.local")
	v.SetDefault("otel.service_name", "lgtm-backend")
	v.SetDefault("otel.exporter_endpoint", "http://otel-collector:4317")
	v.SetDefault("otel.enabled", false)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	return &cfg, v.Unmarshal(&cfg)
}
