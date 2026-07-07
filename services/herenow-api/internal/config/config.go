// Package config loads and persists local herenow configuration under
// $HERENOW_HOME (default ~/.herenow).
package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	herenowv1 "github.com/agarwalvivek29/here.now/packages/schema/generated/go/herenow/v1"
)

type Config struct {
	Addr    string `json:"addr"`
	BaseURL string `json:"base_url"`
	DataDir string `json:"data_dir"`
	Token   string `json:"token"`
	Sub     string `json:"sub"`
	Email   string `json:"email"`
}

func (c Config) Identity() *herenowv1.Identity {
	return &herenowv1.Identity{Sub: c.Sub, Email: c.Email}
}

func Dir() string {
	if d := os.Getenv("HERENOW_HOME"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".herenow")
}

func Path() string { return filepath.Join(Dir(), "config.json") }

func Default() Config {
	return Config{
		Addr:    ":8080",
		BaseURL: "http://localhost:8080",
		DataDir: filepath.Join(Dir(), "data"),
	}
}

func Load() (Config, error) {
	c := Default()
	b, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return c, err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return c, err
	}
	return c, nil
}

func Save(c Config) error {
	if err := os.MkdirAll(Dir(), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), b, 0o600)
}
