package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"autohost-cli/utils"

	"github.com/spf13/cobra"
)

var cloudflareCmd = &cobra.Command{
	Use:   "cloudflare",
	Short: "Comandos para instalar y configurar Cloudflare Tunnel",
}

// Subcomando: instalar cloudflared
var cloudflareInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Instala Cloudflare Tunnel (cloudflared)",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsInitialized() {
			fmt.Println("⚠️ AutoHost no está inicializado. Ejecuta `autohost init` primero.")
			return
		}

		fmt.Println("🌐 Instalando Cloudflare Tunnel (cloudflared)...")
		installCmd := exec.Command("sh", "-c", `
			curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o cloudflared &&
			chmod +x cloudflared &&
			sudo mv cloudflared /usr/local/bin/
		`)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr

		err := installCmd.Run()
		if err != nil {
			fmt.Println("❌ Error al instalar cloudflared:", err)
		} else {
			fmt.Println("✅ Cloudflare Tunnel instalado con éxito.")
		}
	},
}

// Subcomando: login
var cloudflareLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Inicia sesión con tu cuenta de Cloudflare",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsInitialized() {
			fmt.Println("⚠️ Ejecuta `autohost init` primero.")
			return
		}

		fmt.Println("🔐 Ejecutando 'cloudflared tunnel login'...")
		loginCmd := exec.Command("cloudflared", "tunnel", "login")
		loginCmd.Stdout = os.Stdout
		loginCmd.Stderr = os.Stderr
		err := loginCmd.Run()
		if err != nil {
			fmt.Println("❌ Error al iniciar sesión:", err)
		} else {
			fmt.Println("✅ Sesión iniciada correctamente.")
		}
	},
}

// Subcomando: crear túnel
var cloudflareTunnelCmd = &cobra.Command{
	Use:   "tunnel [dominio]",
	Short: "Crea un túnel en Cloudflare y lo vincula al dominio",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsInitialized() {
			fmt.Println("⚠️ Ejecuta `autohost init` primero.")
			return
		}

		domain := args[0]
		fmt.Printf("⚙️ Creando túnel para %s...\n", domain)

		// Crear el túnel
		createCmd := exec.Command("cloudflared", "tunnel", "create", "autohost-tunnel")
		createCmd.Stdout = os.Stdout
		createCmd.Stderr = os.Stderr
		err := createCmd.Run()
		if err != nil {
			fmt.Println("❌ Error al crear túnel:", err)
			return
		}

		// Mover archivo del túnel
		tunnelFile := filepath.Join(os.Getenv("HOME"), ".cloudflared", "autohost-tunnel.json")
		target := filepath.Join(utils.GetAutohostDir(), "cloudflare", "tunnel.json")

		if err := utils.CopyFile(tunnelFile, target); err != nil {
			fmt.Println("⚠️ No se pudo mover el archivo del túnel:", err)
		}

		// Enlazar túnel al dominio
		routeCmd := exec.Command("cloudflared", "tunnel", "route", "dns", "autohost-tunnel", domain)
		routeCmd.Stdout = os.Stdout
		routeCmd.Stderr = os.Stderr
		err = routeCmd.Run()
		if err != nil {
			fmt.Println("❌ Error al configurar ruta DNS:", err)
		} else {
			fmt.Println("✅ Túnel creado y vinculado al dominio:", domain)
		}

		// Guardar config
		cfg := utils.Config{
			Tunnel: "cloudflare",
			Domain: domain,
		}
		if err := utils.SaveConfig(cfg); err != nil {
			fmt.Println("⚠️ Error al guardar config:", err)
		}

		// utils.SaveStatus("cloudflare_tunnel", true)
		// utils.SaveStatus("cloudflare_domain", domain)
	},
}

func init() {
	cloudflareCmd.AddCommand(cloudflareInstallCmd)
	cloudflareCmd.AddCommand(cloudflareLoginCmd)
	cloudflareCmd.AddCommand(cloudflareTunnelCmd)
	rootCmd.AddCommand(cloudflareCmd)
}
