package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

var (
	domain     string
	serviceURL string
	tunnelName string
)

// secureExposeCmd representa el comando "secure-expose"
var secureExposeCmd = &cobra.Command{
	Use:   "secure-expose",
	Short: "Expone un servicio local con Cloudflare Tunnel",
	Run: func(cmd *cobra.Command, args []string) {
		checkCloudflared()
		createTunnel(tunnelName)
		createConfig(tunnelName, domain, serviceURL)
		routeDNS(tunnelName, domain)
		fmt.Println("\n✅ ¡Túnel configurado con éxito!")
		fmt.Println("Ejecuta esto para iniciarlo:")
		fmt.Printf("cloudflared tunnel run %s\n", tunnelName)
	},
}

func init() {
	rootCmd.AddCommand(secureExposeCmd)

	secureExposeCmd.Flags().StringVar(&domain, "domain", "", "Subdominio (ej: app.midominio.com)")
	secureExposeCmd.Flags().StringVar(&serviceURL, "service", "", "Servicio local (ej: http://localhost:3000)")
	secureExposeCmd.Flags().StringVar(&tunnelName, "tunnel-name", "", "Nombre del túnel")

	secureExposeCmd.MarkFlagRequired("domain")
	secureExposeCmd.MarkFlagRequired("service")
	secureExposeCmd.MarkFlagRequired("tunnel-name")
}

// ---- Funciones auxiliares ----

func checkCloudflared() {
	_, err := exec.LookPath("cloudflared")
	if err != nil {
		fmt.Println("❌ cloudflared no está instalado.")
		os.Exit(1)
	}
}

func createTunnel(name string) {
	cmd := exec.Command("cloudflared", "tunnel", "create", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func createConfig(name, domain, service string) {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".cloudflared")
	configPath := filepath.Join(configDir, "config.yml")

	tmpl := `tunnel: {{.TunnelName}}
credentials-file: {{.CredentialsPath}}

ingress:
  - hostname: {{.Domain}}
    service: {{.ServiceURL}}
  - service: http_status:404
`

	credsFile := filepath.Join(configDir, fmt.Sprintf("%s.json", name))
	data := map[string]string{
		"TunnelName":      name,
		"CredentialsPath": credsFile,
		"Domain":          domain,
		"ServiceURL":      service,
	}

	t, err := template.New("config").Parse(tmpl)
	if err != nil {
		fmt.Println("❌ Error generando config.yml:", err)
		os.Exit(1)
	}

	f, _ := os.Create(configPath)
	defer f.Close()
	t.Execute(f, data)
}

func routeDNS(name, domain string) {
	cmd := exec.Command("cloudflared", "tunnel", "route", "dns", name, domain)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
