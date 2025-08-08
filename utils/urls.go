package utils

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml"
)

type CaddyConfig struct {
	GPGKey string `toml:"gpg_key"`
	Repo   string `toml:"repo"`
}

type TailscaleConfig struct {
	InstallScript string `toml:"install_script"`
}

type DockerConfig struct {
	InstallScript string `toml:"install_script"`
}

type CloudflareConfig struct {
	DownloadURL string `toml:"download_url"`
}

type DownloadConfig struct {
	Caddy      CaddyConfig      `toml:"caddy"`
	Tailscale  TailscaleConfig  `toml:"tailscale"`
	Docker     DockerConfig     `toml:"docker"`
	Cloudflare CloudflareConfig `toml:"cloudflare"`
}

var DownloadURLs DownloadConfig

func LoadURLsConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error leyendo urls config: %w", err)
	}

	if err := toml.Unmarshal(data, &DownloadURLs); err != nil {
		return fmt.Errorf("error parseando urls config: %w", err)
	}
	return nil
}
