package cmd

import (
	"autohost-cli/internal/helpers/caddy"
	"autohost-cli/internal/helpers/cloudflared"
	"autohost-cli/internal/helpers/docker"
	"autohost-cli/internal/helpers/initializer"
	"autohost-cli/internal/helpers/tailscale"
	"autohost-cli/utils"
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// setupCmd representa el comando 'autohost setup'
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configura tu servidor para autohospedar servicios",
	Long: `Este comando instala Docker, Caddy, configura dominios,
		y prepara túneles seguros para desplegar tus apps autohospedadas.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\n🔧 Iniciando configuración del servidor...")

		initializer.EnsureAutohostDirs()

		if !docker.DockerInstalled() {
			if utils.Confirm("⚠️ Docker no está instalado. ¿Deseas instalarlo automáticamente? [y/N]: ") {
				docker.InstallDocker()
			} else {
				fmt.Println("🚫 Instalación cancelada. Instala Docker manualmente y vuelve a ejecutar el setup.")
				return
			}
		} else {
			fmt.Println("✅ Docker ya está instalado.")
		}

		if utils.Confirm("¿Deseas agregar tu usuario al grupo 'docker' para usar Docker sin sudo? [y/N]: ") {
			docker.AddUserToDockerGroup()
		}

		if utils.Confirm("¿Deseas instalar y configurar Caddy como reverse proxy? [y/N]: ") {
			caddy.InstallCaddy()
			caddy.CreateCaddyfile()
		}

		option := utils.AskOption("🔒 ¿Qué tipo de acceso quieres configurar?", []string{"Tailscale (privado)", "Cloudflare Tunnel (público con dominio)"})
		switch option {
		case "Tailscale (privado)":
			tailscale.InstallTailscale()
		case "Cloudflare Tunnel (público con dominio)":
			cloudflared.InstallCloudflared()
			fmt.Print("Introduce el subdominio para el túnel (ej: blog.misitio.com): ")
			reader := bufio.NewReader(os.Stdin)
			domain, _ := reader.ReadString('\n')
			domain = strings.TrimSpace(domain)
			cloudflared.ConfigureCloudflareTunnel(domain)
		}

		fmt.Println("\n✅ Configuración inicial completa.")
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
