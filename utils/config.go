package utils

import (
	"encoding/json"
	"fmt"
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

func ConfigureCaddy(app, domain string) error {
	caddyfilePath := "/opt/autohost/Caddyfile" // o ~/.autohost/...
	entry := fmt.Sprintf(`
%s {
	reverse_proxy %s:80
}
`, domain, app)

	// Aquí lo puedes mejorar para no duplicar entradas
	f, err := os.OpenFile(caddyfilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(entry)
	return err
}
