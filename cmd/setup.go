package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
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

		ensureAutohostDirs()

		if !dockerInstalled() {
			if confirm("⚠️ Docker no está instalado. ¿Deseas instalarlo automáticamente? [y/N]: ") {
				installDocker()
			} else {
				fmt.Println("🚫 Instalación cancelada. Instala Docker manualmente y vuelve a ejecutar el setup.")
				return
			}
		} else {
			fmt.Println("✅ Docker ya está instalado.")
		}

		if confirm("¿Deseas agregar tu usuario al grupo 'docker' para usar Docker sin sudo? [y/N]: ") {
			addUserToDockerGroup()
		}

		if confirm("¿Deseas instalar y configurar Caddy como reverse proxy? [y/N]: ") {
			installCaddy()
			createCaddyfile()
		}

		option := askOption("🔒 ¿Qué tipo de acceso quieres configurar?", []string{"Tailscale (privado)", "Cloudflare Tunnel (público con dominio)"})
		switch option {
		case "Tailscale (privado)":
			installTailscale()
		case "Cloudflare Tunnel (público con dominio)":
			installCloudflared()
			fmt.Print("Introduce el subdominio para el túnel (ej: blog.misitio.com): ")
			reader := bufio.NewReader(os.Stdin)
			domain, _ := reader.ReadString('\n')
			domain = strings.TrimSpace(domain)
			configureCloudflareTunnel(domain)
		}

		fmt.Println("\n✅ Configuración inicial completa.")
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func dockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func installDocker() {
	fmt.Println("🔄 Instalando Docker...")
	execShell("curl -fsSL https://get.docker.com | sh")
	fmt.Println("✅ Docker instalado con éxito.")
}

func addUserToDockerGroup() {
	user := os.Getenv("SUDO_USER")
	if user == "" {
		user = os.Getenv("USER")
	}
	if user == "" {
		fmt.Println("⚠️ No se pudo determinar el usuario. Saltando este paso.")
		return
	}
	execShell(fmt.Sprintf("sudo usermod -aG docker %s", user))
	fmt.Printf("✅ Usuario '%s' agregado al grupo 'docker'.\n", user)
}

func installCaddy() {
	fmt.Println("🚀 Instalando Caddy...")
	execShell(`
	sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl &&
		curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg &&
		curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list &&
		sudo apt update && sudo apt install caddy
	`)
	execShell("sudo systemctl enable caddy")
	execShell("sudo systemctl start caddy")
	fmt.Println("✅ Caddy instalado y activado correctamente.")
}

func createCaddyfile() {
	caddyfilePath := "/etc/caddy/Caddyfile"

	if _, err := os.Stat(caddyfilePath); err == nil {
		fmt.Println("📄 Ya existe un Caddyfile, no se modificará.")
		return
	}

	content := `
http://localhost {
	respond \"🚀 AutoHost CLI: Caddy instalado y funcionando\"
}
`
	err := os.WriteFile(caddyfilePath, []byte(content), 0644)
	if err != nil {
		fmt.Println("❌ Error creando Caddyfile:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Caddyfile creado en /etc/caddy/Caddyfile")

	reloadCmd := exec.Command("sudo", "systemctl", "reload", "caddy")
	reloadCmd.Stdout = os.Stdout
	reloadCmd.Stderr = os.Stderr
	if err := reloadCmd.Run(); err != nil {
		fmt.Println("⚠️ No se pudo recargar Caddy automáticamente. Hazlo manualmente con: sudo systemctl reload caddy")
	} else {
		fmt.Println("🔁 Caddy recargado con éxito.")
	}
}

func installTailscale() {
	fmt.Println("🔐 Instalando Tailscale...")
	execShell("curl -fsSL https://tailscale.com/install.sh | sh")
	fmt.Println("🔐 Autenticándote con Tailscale...")
	execShell("sudo tailscale up")
}

func installCloudflared() {
	fmt.Println("🌐 Instalando Cloudflare Tunnel (cloudflared)...")
	execShell(`
		curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o cloudflared &&
		chmod +x cloudflared &&
		sudo mv cloudflared /usr/local/bin/
	`)
	fmt.Println("✅ Cloudflare Tunnel instalado.")
	fmt.Println("ℹ️ Ejecuta 'cloudflared tunnel login' para autenticarte.")
}

func configureCloudflareTunnel(domain string) {
	fmt.Println("⚙️ Configurando Cloudflare Tunnel para:", domain)
	execShell("cloudflared tunnel create autohost-tunnel")
	execShell(fmt.Sprintf("cloudflared tunnel route dns autohost-tunnel %s", domain))
	fmt.Println("✅ Túnel configurado correctamente.")
}

func confirm(prompt string) bool {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}

func askOption(prompt string, options []string) string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(prompt)
		for i, opt := range options {
			fmt.Printf("[%d] %s\n", i+1, opt)
		}
		fmt.Print("Elige una opción: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if i, err := strconv.Atoi(input); err == nil && i >= 1 && i <= len(options) {
			return options[i-1]
		}
		fmt.Println("❌ Opción inválida, intenta de nuevo.")
	}
}

func execShell(command string) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("❌ Error ejecutando comando:", err)
		os.Exit(1)
	}
}

func ensureAutohostDirs() {
	dirs := []string{
		"/etc/autohost",
		"/opt/autohost/docker",
		"/opt/autohost/templates",
		"/var/lib/autohost/logs",
		"/var/lib/autohost/state",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("❌ Error creando %s: %v\n", dir, err)
			os.Exit(1)
		}
	}
	fmt.Println("📁 Estructura de carpetas FHS creada.")
}
