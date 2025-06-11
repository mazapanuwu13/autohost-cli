package cmd

import (
	"fmt"
	"os"
	"os/exec"

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
	rootCmd.AddCommand(tailscaleCmd)
}
