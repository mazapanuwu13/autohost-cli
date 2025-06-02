package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Tunnel string `json:"tunnel"`
	Domain string `json:"domain,omitempty"`
}

func SaveConfig(cfg Config) error {
	path := filepath.Join(GetAutohostDir(), "config.json")

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}
