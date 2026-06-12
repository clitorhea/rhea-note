package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Salt      []byte `json:"salt"`
	StoreDir  string `json:"store_dir"`
	ServerURL string `json:"server_url"`
	Token     string `json:"token"`
}

func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "secnotes")
}

func ConfigFile() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(ConfigFile())
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFile(), data, 0600)
}
