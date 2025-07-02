package config

import (
	"encoding/json"
	"os"
)

// Config represents the Golem project configuration
type Config struct {
	ProjectName string       `json:"projectName"`
	Version     string       `json:"version"`
	Entry       string       `json:"entry"`
	Output      string       `json:"output"`
	Dev         DevConfig    `json:"dev"`
	Build       BuildConfig  `json:"build"`
	Server      ServerConfig `json:"server"`
	Wasm        WasmConfig   `json:"wasm"`
}

// DevConfig holds development server configuration
type DevConfig struct {
	Port      int      `json:"port"`
	HotReload bool     `json:"hotReload"`
	Watch     []string `json:"watch"`
}

// BuildConfig holds build configuration
type BuildConfig struct {
	Minify    bool   `json:"minify"`
	Target    string `json:"target"`
	Sourcemap bool   `json:"sourcemap"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	GRPC      GRPCConfig `json:"grpc"`
	Functions string     `json:"functions"`
}

// GRPCConfig holds gRPC server configuration
type GRPCConfig struct {
	Port       int  `json:"port"`
	Reflection bool `json:"reflection"`
}

// WasmConfig holds WebAssembly build configuration
type WasmConfig struct {
	OptimizeSize   bool     `json:"optimizeSize"`
	EnableFeatures []string `json:"enableFeatures"`
}

// Load loads the configuration from a JSON file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
