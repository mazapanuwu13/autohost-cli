package app

import (
	"autohost-cli/internal/adapters/agentconfig"
	"autohost-cli/internal/ports"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type VPNService struct {
	Tailscale ports.Tailscale
}

func (s *VPNService) Connect(loginServer, authKey string) error {
	if !s.Tailscale.Installed() {
		fmt.Println("⚠️ Tailscale no está instalado. Intentando instalar...")
		if err := s.Tailscale.Install(); err != nil {
			return fmt.Errorf("error al instalar Tailscale: %w", err)
		}
	}

	if err := s.Tailscale.LoginHeadscale(loginServer, authKey); err != nil {
		return err
	}

	ip, err := s.Tailscale.IP()
	if err == nil && ip != "" {
		fmt.Println()
		fmt.Printf("📌 IP Privada Headscale asignada: %s\n", ip)
		s.notifyVPNStatus(ip)
	}

	return nil
}

func (s *VPNService) Disconnect() error {
	if !s.Tailscale.Installed() {
		return fmt.Errorf("Tailscale no está instalado")
	}
	if err := s.Tailscale.Down(); err != nil {
		return err
	}
	s.notifyVPNStatus("")
	return nil
}

func (s *VPNService) Logout() error {
	if !s.Tailscale.Installed() {
		return fmt.Errorf("Tailscale no está instalado")
	}
	if err := s.Tailscale.Logout(); err != nil {
		return err
	}
	s.notifyVPNStatus("")
	return nil
}

func (s *VPNService) notifyVPNStatus(ipVPN string) {
	cfg, err := agentconfig.Load()
	if err != nil || cfg.ApiURL == "" || cfg.ApiToken == "" {
		return // Sin configuración de agente o sin token
	}

	payload := map[string]string{"ip_vpn": ipVPN}
	body, _ := json.Marshal(payload)

	reqURL := fmt.Sprintf("%s/v1/vpn/status", strings.TrimRight(cfg.ApiURL, "/"))
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.ApiToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if ipVPN != "" {
			fmt.Printf("✅ IP Privada (%s) registrada exitosamente en Autohost.\n", ipVPN)
		} else {
			fmt.Println("✅ IP Privada desregistrada de Autohost.")
		}
	}
}
