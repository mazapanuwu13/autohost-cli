// Package agentconfig manages the agent configuration file at /etc/autohost/config.yaml.
package agentconfig

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// AgentConfig holds the agent's runtime configuration.
type AgentConfig struct {
	ApiToken     string
	RefreshToken string
	ApiURL       string
	NodeID       string
	GRPCAddress  string
}

const configPath = "/etc/autohost/config.yaml"

// Load reads the agent config from /etc/autohost/config.yaml.
func Load() (*AgentConfig, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		if out, sudoErr := exec.Command("sudo", "cat", configPath).Output(); sudoErr == nil && len(out) > 0 {
			content = out
		} else {
			return nil, fmt.Errorf("no se pudo leer %s: %w", configPath, err)
		}
	}
	var raw struct {
		APIURL     string `yaml:"api_url"`
		AgentToken string `yaml:"agent_token"`
	}
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return nil, fmt.Errorf("config inválido: %w", err)
	}
	return &AgentConfig{ApiURL: raw.APIURL, ApiToken: raw.AgentToken}, nil
}

// Save updates api_url and agent_token in /etc/autohost/config.yaml, preserving
// all other content. Uses sudo when the process is not running as root.
func Save(cfg AgentConfig) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("archivo de configuración no existe: %s. Por favor, ejecuta 'autohost agent install' primero", configPath)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsPermission(err) {
			return fmt.Errorf("error leyendo configuración: %w", err)
		}
		// File is owned by root — read it via sudo.
		out, sudoErr := exec.Command("sudo", "cat", configPath).Output()
		if sudoErr != nil {
			return fmt.Errorf("error leyendo configuración: %w", err)
		}
		content = out
	}

	updated := string(content)
	updated = regexp.MustCompile(`(?m)^agent_token:.*$`).
		ReplaceAllString(updated, fmt.Sprintf(`agent_token: "%s"`, cfg.ApiToken))
	if cfg.RefreshToken != "" {
		updated = regexp.MustCompile(`(?m)^refresh_token:.*$`).
			ReplaceAllString(updated, fmt.Sprintf(`refresh_token: "%s"`, cfg.RefreshToken))
	}
	if cfg.ApiURL != "" {
		updated = regexp.MustCompile(`(?m)^api_url:.*$`).
			ReplaceAllString(updated, fmt.Sprintf(`api_url: "%s"`, cfg.ApiURL))

		// Derive ws_url from api_url.
		if wsURL, err := deriveWSURL(cfg.ApiURL); err == nil {
			updated = regexp.MustCompile(`(?m)^ws_url:.*$`).
				ReplaceAllString(updated, fmt.Sprintf(`ws_url: "%s"`, wsURL))
		}
	}
	// Use grpc_address returned by the server; fall back to deriving it from api_url.
	grpcAddr := CleanGRPCAddress(cfg.GRPCAddress)
	if grpcAddr == "" && cfg.ApiURL != "" {
		if derived, err := deriveGRPCAddress(cfg.ApiURL); err == nil {
			grpcAddr = derived
		}
	}
	if grpcAddr != "" {
		updated = regexp.MustCompile(`(?m)^grpc_address:.*$`).
			ReplaceAllString(updated, fmt.Sprintf(`grpc_address: "%s"`, grpcAddr))
	}
	if cfg.NodeID != "" {
		updated = regexp.MustCompile(`(?m)^node_id:.*$`).
			ReplaceAllString(updated, fmt.Sprintf(`node_id: "%s"`, cfg.NodeID))
	}

	if os.Geteuid() == 0 {
		return os.WriteFile(configPath, []byte(updated), 0600)
	}

	// Write to a temp file first, then copy with sudo.
	tmp, err := os.CreateTemp("", "autohost-config-*.yaml")
	if err != nil {
		return fmt.Errorf("error creando archivo temporal: %w", err)
	}
	defer os.Remove(tmp.Name())

	if err := os.WriteFile(tmp.Name(), []byte(updated), 0600); err != nil {
		return fmt.Errorf("error escribiendo archivo temporal: %w", err)
	}
	tmp.Close()

	cmd := exec.Command("sudo", "cp", tmp.Name(), configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error copiando archivo con sudo: %w", err)
	}
	return nil
}

// deriveWSURL converts an HTTP API URL to a WebSocket URL with /ws path.
func deriveWSURL(apiURL string) (string, error) {
	u, err := url.Parse(apiURL)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	default:
		u.Scheme = "ws"
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/ws"
	return u.String(), nil
}

// CleanGRPCAddress normalizes a gRPC address string to ensure it is in host:port format
// without http:// or https:// schemes.
func CleanGRPCAddress(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	hasHTTPS := strings.HasPrefix(raw, "https://")
	hasHTTP := strings.HasPrefix(raw, "http://")
	clean := strings.TrimPrefix(raw, "https://")
	clean = strings.TrimPrefix(clean, "http://")
	clean = strings.TrimRight(clean, "/")

	if strings.HasPrefix(clean, ":") {
		return clean
	}

	if _, _, err := net.SplitHostPort(clean); err == nil {
		return clean
	}

	if hasHTTPS {
		return clean + ":443"
	}
	if hasHTTP {
		return clean + ":9090"
	}
	if strings.Contains(clean, ".") {
		return clean + ":443"
	}
	return clean + ":9090"
}

// deriveGRPCAddress extracts host from the API URL and determines appropriate port.
func deriveGRPCAddress(apiURL string) (string, error) {
	u, err := url.Parse(apiURL)
	if err != nil {
		return "", err
	}
	host := u.Hostname()
	if u.Scheme == "https" {
		return host + ":443", nil
	}
	return host + ":9090", nil
}
