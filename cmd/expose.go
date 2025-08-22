// cmd/expose.go
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"autohost-cli/internal/infra"
	"autohost-cli/utils"

	"github.com/spf13/cobra"
)

var (
	// selector
	provider string // tailscale|cloudflare

	// tailscale
	subdomain string // ej: app.maza-server  (FQDN dentro de tu zona interna)
	port      int    // ej: 3000
	withCaddy bool   // genera vhost y reload

	// opcional override para Terraform
	tailnet string

	// cloudflare (tu flujo existente; opcional)
	domain     string
	serviceURL string
	tunnelName string
)

var exposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "Expone una app por Tailscale (split-DNS + CoreDNS en Docker) o Cloudflare Tunnel",
	RunE: func(cmd *cobra.Command, args []string) error {
		if provider == "" {
			provider = "tailscale"
		}
		switch strings.ToLower(provider) {
		case "tailscale":
			if err := require(utils.IsInitialized(), "⚠️ AutoHost no está inicializado. Ejecuta `autohost init` primero."); err != nil {
				return err
			}
			if err := require(subdomain != "", "--subdomain es requerido (ej: app.maza-server)"); err != nil {
				return err
			}
			if err := require(port > 0, "--port es requerido (ej: 3000)"); err != nil {
				return err
			}
			return exposeWithTailscale(subdomain, port, withCaddy, tailnet)

		// case "cloudflare":
		// 	// valida y llama a exposeWithCloudflare(...)
		// 	return exposeWithCloudflare(tunnelName, domain, serviceURL)

		default:
			return fmt.Errorf("provider inválido: %s (usa tailscale|cloudflare)", provider)
		}
	},
}

func init() {
	rootCmd.AddCommand(exposeCmd)

	exposeCmd.Flags().StringVar(&provider, "provider", "tailscale", "Proveedor: tailscale|cloudflare")

	// Tailscale
	exposeCmd.Flags().StringVar(&subdomain, "subdomain", "", "Subdominio interno FQDN (ej: app.maza-server)")
	exposeCmd.Flags().IntVar(&port, "port", 0, "Puerto local donde corre la app (ej: 3000)")
	exposeCmd.Flags().BoolVar(&withCaddy, "with-caddy", true, "Generar vhost en Caddy y recargar")
	exposeCmd.Flags().StringVar(&tailnet, "tailnet", "", "(Opcional) tailnet para Terraform (si se omite, se usa TAILSCALE_TAILNET o '-')")

	// Cloudflare (compat con tu código actual; opcional)
	exposeCmd.Flags().StringVar(&domain, "domain", "", "Dominio FQDN (ej: app.midominio.com)")
	exposeCmd.Flags().StringVar(&serviceURL, "service", "", "Servicio local (ej: http://localhost:3000)")
	exposeCmd.Flags().StringVar(&tunnelName, "tunnel-name", "", "Nombre del túnel")
}

// --------------------- TAILSCALE FLOW ---------------------
func exposeWithTailscale(fqdn string, port int, setupCaddy bool, tailnet string) error {
	fmt.Println("🔗 Proveedor: Tailscale (Split-DNS + CoreDNS en Docker)")

	// 0) binarios imprescindibles
	if err := checkBinary("tailscale"); err != nil {
		return err
	}

	// 1) IP tailscale local (este host será el nameserver)
	tailIP, err := tailscaleIP()
	if err != nil || tailIP == "" {
		return fmt.Errorf("no pude obtener IP de tailscale (¿logueado?): %v", err)
	}
	fmt.Printf("🛰️  IP tailnet local: %s\n", tailIP)

	// 2) dividir host y apex (zona)
	host, zone := splitHostZone(fqdn)
	if zone == "" || host == "" {
		return fmt.Errorf("subdomain inválido: %s (esperado: host.zona, p.ej. app.maza-server)", fqdn)
	}
	fmt.Printf("🌐 Zona: %s | Host: %s\n", zone, host)

	// 3) CoreDNS (Docker): asegurar contenedor y Corefile base
	corefilePath, err := infra.InstallAndRunCoreDNSWithDocker(zone, fqdn, tailIP)
	if err != nil {
		return fmt.Errorf("CoreDNS (Docker): %w", err)
	}
	fmt.Println("🧩 CoreDNS (Docker) listo. Corefile:", corefilePath)

	// 4) Añadir/actualizar FQDN en Corefile y reiniciar contenedor si cambió
	if err := infra.EnsureDomainAndReload(zone, fqdn, tailIP); err != nil {
		return fmt.Errorf("CoreDNS update/reload: %w", err)
	}
	fmt.Println("✅ CoreDNS actualizado (si fue necesario).")

	// 5) Terraform Split-DNS: la zona la resuelve ESTE nameserver (tailIP)
	fmt.Println("⚙️  Aplicando Split DNS (Terraform) en el tailnet…")
	if err := infra.ConfigureSplitDNSWithTerraform(infra.SplitDNSOpts{
		Tailnet:      tailnet,          // si vacío, tu función usa TAILSCALE_TAILNET o '-'
		Domain:       zone,             // apex
		Nameservers:  []string{tailIP}, // este nodo responde la zona
		SearchPaths:  []string{zone},   // para resolver "host" corto
		APIKeyEnvVar: "TAILSCALE_API_KEY",
	}); err != nil {
		return err
	}
	fmt.Println("✅ Split DNS aplicado en la tailnet.")

	// 6) (opcional) Caddy: fqdn → localhost:port
	if setupCaddy {
		if err := ensureCaddySite(fqdn, port); err != nil {
			fmt.Println("⚠️  No se pudo escribir/reload Caddy:", err)
		} else {
			fmt.Println("✅ Caddy site configurado y recargado.")
		}
	} else {
		fmt.Println("ℹ️  Omitido Caddy (usa --with-caddy para generarlo).")
	}

	fmt.Printf("\n🎯 Listo. %s resolverá a %s en tu tailnet y proxyeará a localhost:%d (si Caddy está habilitado)\n", fqdn, tailIP, port)
	fmt.Printf("   Corefile: %s\n", corefilePath)
	return nil
}

// --------------------- HELPERS ---------------------

func tailscaleIP() (string, error) {
	out, err := exec.Command("tailscale", "ip", "-4").Output()
	if err != nil {
		return "", err
	}
	ip := strings.TrimSpace(string(out))
	if i := strings.IndexByte(ip, '\n'); i > -1 {
		ip = ip[:i]
	}
	return ip, nil
}

func splitHostZone(fqdn string) (host, zone string) {
	s := strings.TrimSpace(fqdn)
	if s == "" {
		return "", ""
	}
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return "", ""
	}
	host = parts[0]
	zone = strings.Join(parts[1:], ".")
	return
}

func ensureCaddySite(fqdn string, port int) error {
	if err := checkBinary("caddy"); err != nil {
		return err
	}
	home, _ := os.UserHomeDir()
	sitesDir := filepath.Join(home, ".autohost", "caddy", "sites")
	_ = os.MkdirAll(sitesDir, 0o755)

	sitePath := filepath.Join(sitesDir, safeName(fqdn)+".caddy")
	siteT := `{{.Host}} {
	encode zstd gzip
	reverse_proxy localhost:{{.Port}}
}
`
	if err := renderToFile(siteT, sitePath, map[string]any{
		"Host": fqdn, "Port": port,
	}); err != nil {
		return err
	}

	// Asegura import en Caddyfile maestro
	caddyfile := "/etc/caddy/Caddyfile"
	importLine := "\n# autohost import\nimport " + sitesDir + "/*.caddy\n"
	_ = ensureLineInFile(caddyfile, importLine)

	// reload caddy
	_ = exec.Command("systemctl", "reload", "caddy").Run()
	return nil
}

func ensureLineInFile(path, line string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return os.WriteFile(path, []byte(line+"\n"), 0o644)
	}
	if !strings.Contains(string(b), line) {
		return os.WriteFile(path, append(b, []byte("\n"+line+"\n")...), 0o644)
	}
	return nil
}

func renderToFile(tmpl, outPath string, data any) error {
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, data)
}

func checkBinary(bin string) error {
	_, err := exec.LookPath(bin)
	if err != nil {
		return fmt.Errorf("❌ %s no está instalado", bin)
	}
	return nil
}

func safeName(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, ":", "_")
	return s
}

func require(ok bool, msg string) error {
	if !ok {
		return errors.New(msg)
	}
	return nil
}
