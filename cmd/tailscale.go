package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"autohost-cli/internal/infra"
	"autohost-cli/utils"

	"github.com/spf13/cobra"
)

var tailscaleCmd = &cobra.Command{
	Use:   "tailscale",
	Short: "Comandos para instalar, autenticar y gestionar Tailscale",
}

// Subcomando: install
var tailscaleInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Instala Tailscale en el sistema",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsInitialized() {
			fmt.Println("⚠️ AutoHost no está inicializado. Ejecuta `autohost init` primero.")
			return
		}

		fmt.Println("📦 Instalando Tailscale...")

		installCmd := exec.Command("sh", "-c", `
		curl -fsSL https://tailscale.com/install.sh | sh
		`)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		installCmd.Run()

		fmt.Println("✅ Tailscale instalado. Ahora ejecuta `autohost tailscale login` para autenticarte.")
	},
}

// Subcomando: login
var tailscaleLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Autentica y conecta Tailscale al daemon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔐 Autenticando con Tailscale...")

		loginCmd := exec.Command("sudo", "tailscale", "up")
		loginCmd.Stdout = os.Stdout
		loginCmd.Stderr = os.Stderr

		if err := loginCmd.Run(); err != nil {
			fmt.Println("❌ Error al conectar con Tailscale:", err)
			return
		}

		fmt.Println("✅ Conectado a Tailscale.")
	},
}

// Subcomando: logout
var tailscaleLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Cierra sesión de Tailscale",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔌 Cerrando sesión de Tailscale...")

		logoutCmd := exec.Command("sudo", "tailscale", "logout")
		logoutCmd.Stdout = os.Stdout
		logoutCmd.Stderr = os.Stderr
		logoutCmd.Run()

		fmt.Println("✅ Sesión cerrada.")
	},
}

var tailscaleSplitDnsCmd = &cobra.Command{
	Use:   "split-dns",
	Short: "Configura Split DNS para Tailscale (vía Terraform)",
	Long: `Aplica Split DNS en tu tailnet usando Terraform y el provider oficial de Tailscale.
Requiere TAILSCALE_API_KEY (y opcional TAILSCALE_TAILNET) en el entorno.

Ejemplo:
  autohost tailscale split-dns \
    --domain maza-server \
    --nameservers 100.112.92.90 \
    --search-paths maza-server \
    --tailnet tu-org.ts.net`,
	RunE: func(cmd *cobra.Command, args []string) error {
		domain, _ := cmd.Flags().GetString("domain")
		nsStr, _ := cmd.Flags().GetString("nameservers")
		searchStr, _ := cmd.Flags().GetString("search-paths")
		tailnet, _ := cmd.Flags().GetString("tailnet")

		if domain == "" || nsStr == "" {
			return fmt.Errorf("flags requeridas: --domain y --nameservers (separados por coma si son varios)")
		}

		nameservers := splitAndTrim(nsStr)
		searchPaths := splitAndTrim(searchStr)

		fmt.Println("⚙️  Configurando Split DNS con Terraform...")
		err := infra.ConfigureSplitDNSWithTerraform(infra.SplitDNSOpts{
			Tailnet:      tailnet,
			Domain:       domain,
			Nameservers:  nameservers,
			SearchPaths:  searchPaths,
			APIKeyEnvVar: "TAILSCALE_API_KEY",
		})
		if err != nil {
			return err
		}
		fmt.Println("✅ Split DNS aplicado.")
		return nil
	},
}

// Subcomando: status
var tailscaleStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Muestra el estado actual de Tailscale",
	Run: func(cmd *cobra.Command, args []string) {
		statusCmd := exec.Command("sudo", "tailscale", "status")
		statusCmd.Stdout = os.Stdout
		statusCmd.Stderr = os.Stderr
		statusCmd.Run()
	},
}

func init() {
	tailscaleCmd.AddCommand(tailscaleInstallCmd)
	tailscaleCmd.AddCommand(tailscaleLoginCmd)
	tailscaleCmd.AddCommand(tailscaleLogoutCmd)
	tailscaleCmd.AddCommand(tailscaleStatusCmd)
	tailscaleCmd.AddCommand(tailscaleSplitDnsCmd)
	tailscaleSplitDnsCmd.Flags().String("domain", "", "Dominio a resolver vía Split DNS (ej. maza-server)")
	tailscaleSplitDnsCmd.Flags().String("nameservers", "", "Lista de resolvers (coma-separados), ej. 100.112.92.90,1.1.1.1")
	tailscaleSplitDnsCmd.Flags().String("search-paths", "", "(Opcional) dominios de búsqueda, coma-separados")
	tailscaleSplitDnsCmd.Flags().String("tailnet", "", "(Opcional) tailnet; si no se indica usa TAILSCALE_TAILNET o '-'")

	rootCmd.AddCommand(tailscaleCmd)
}

func splitAndTrim(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}
