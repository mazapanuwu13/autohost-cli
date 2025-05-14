package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Config representa la configuración del usuario para AutoHost CLI.
type Config struct {
	Tunnel string `json:"tunnel"`
	Domain string `json:"domain,omitempty"`
}

// setupCmd representa el comando 'autohost setup'
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configura tu servidor para autohospedar servicios",
	Long: `Este comando instala Docker, configura dominios, 
y prepara túneles seguros para desplegar tus apps autohospedadas.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔧 Iniciando configuración...")

		// Verifica si Docker está instalado
		if dockerInstalled() {
			fmt.Println("✅ Docker ya está instalado")
		} else {
			fmt.Println("⚠️ Docker no está instalado")
			if confirm("¿Deseas instalar Docker automáticamente? [y/N]: ") {
				installDocker()
			} else {
				fmt.Println("🚫 Instalación cancelada. Instala Docker manualmente y vuelve a ejecutar el setup.")
				return
			}
		}

		// Preguntar si se desea agregar permisos al usuario para Docker
		if confirm("¿Deseas agregar tu usuario al grupo 'docker' para usar Docker sin sudo? [y/N]: ") {
			addUserToDockerGroup()
		}

		// Elegir el tipo de túnel seguro
		fmt.Println("🔒 ¿Qué tipo de acceso quieres configurar?")
		fmt.Println("[1] Tailscale (privado)")
		fmt.Println("[2] Cloudflare Tunnel (público con dominio)")
		fmt.Print("Elige una opción [1/2]: ")
		reader := bufio.NewReader(os.Stdin)
		option, _ := reader.ReadString('\n')
		option = strings.TrimSpace(option)

		var config Config

		switch option {
		case "1":
			installTailscale()
			config.Tunnel = "tailscale"
		case "2":
			installCloudflared()
			config.Tunnel = "cloudflare"
			// Pedir subdominio para Cloudflare Tunnel
			fmt.Print("Introduce el subdominio para el túnel (ej: blog.misitio.com): ")
			domain, _ := reader.ReadString('\n')
			domain = strings.TrimSpace(domain)
			config.Domain = domain

			// Configurar automáticamente el túnel
			configureCloudflareTunnel(domain)
		default:
			fmt.Println("❌ Opción inválida. Abortando configuración de túnel.")
			return
		}

		// Guardar configuración en ~/.autohost/config.json
		err := saveConfig(config)
		if err != nil {
			fmt.Println("❌ Error guardando configuración:", err)
		} else {
			fmt.Println("✅ Configuración guardada en ~/.autohost/config.json")
		}

		fmt.Println("✅ Configuración inicial completa.")
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

// Verifica si Docker está instalado.
func dockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// Instala Docker usando el script oficial.
func installDocker() {
	fmt.Println("🔄 Instalando Docker...")

	cmd := exec.Command("sh", "-c", "curl -fsSL https://get.docker.com | sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("❌ Error al instalar Docker:", err)
		os.Exit(1)
	} else {
		fmt.Println("✅ Docker instalado con éxito.")
	}
}

// Añade al usuario actual al grupo 'docker' para no usar sudo
func addUserToDockerGroup() {
	// Intentar determinar el usuario real (si se usó sudo)
	user := os.Getenv("SUDO_USER")
	if user == "" {
		// Si no se usó sudo, tomar la variable USER
		user = os.Getenv("USER")
	}
	if user == "" {
		fmt.Println("⚠️ No se pudo determinar el usuario para agregar al grupo 'docker'. Saltando este paso.")
		return
	}

	fmt.Printf("👤 Agregando al usuario '%s' al grupo 'docker'...\n", user)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("sudo usermod -aG docker %s", user))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("❌ Error al ejecutar usermod:", err)
		return
	}
	fmt.Printf("✅ Usuario '%s' agregado al grupo 'docker'. ", user)
	fmt.Println("Es posible que debas cerrar y volver a iniciar sesión para que surta efecto.")
}

// Instala Tailscale.
func installTailscale() {
	fmt.Println("🔐 Instalando Tailscale...")

	cmd := exec.Command("sh", "-c", "curl -fsSL https://tailscale.com/install.sh | sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("❌ Error al instalar Tailscale:", err)
		os.Exit(1)
	} else {
		// fmt.Println("✅ Tailscale instalado con éxito.")
		// fmt.Println("ℹ️ Ejecuta 'sudo tailscale up' para autenticarte con tu cuenta.")
		fmt.Println("🔐 Autenticándote con Tailscale...")
		cmdLogin := exec.Command("sudo", "tailscale", "up")
		cmdLogin.Stdout = os.Stdout
		cmdLogin.Stderr = os.Stderr
		err := cmdLogin.Run()
		if err != nil {
			fmt.Println("❌ Error al ejecutar 'tailscale up':", err)
			fmt.Println("ℹ️ Puedes ejecutarlo manualmente con: sudo tailscale up")
		} else {
			fmt.Println("✅ Tailscale conectado correctamente.")
		}

	}
}

// Instala Cloudflare Tunnel (cloudflared).
func installCloudflared() {
	fmt.Println("🌐 Instalando Cloudflare Tunnel (cloudflared)...")

	cmd := exec.Command("sh", "-c", `
		curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o cloudflared &&
		chmod +x cloudflared &&
		sudo mv cloudflared /usr/local/bin/
	`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("❌ Error al instalar Cloudflare Tunnel:", err)
		os.Exit(1)
	} else {
		fmt.Println("✅ Cloudflare Tunnel instalado con éxito.")
		fmt.Println("ℹ️ Ejecuta 'cloudflared tunnel login' para autenticarte con tu cuenta de Cloudflare.")
	}
}

// Configura automáticamente Cloudflare Tunnel para el dominio proporcionado.
func configureCloudflareTunnel(domain string) {
	fmt.Println("⚙️ Configurando Cloudflare Tunnel para el dominio:", domain)
	// Intenta crear el túnel llamado 'autohost-tunnel'
	cmd := exec.Command("sh", "-c", "cloudflared tunnel create autohost-tunnel")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("❌ Error al crear el túnel. Es posible que ya exista o que necesites crearlo manualmente.")
	} else {
		// Configurar la ruta DNS para el túnel
		routeCmd := exec.Command("sh", "-c", fmt.Sprintf("cloudflared tunnel route dns autohost-tunnel %s", domain))
		routeCmd.Stdout = os.Stdout
		routeCmd.Stderr = os.Stderr
		err = routeCmd.Run()
		if err != nil {
			fmt.Println("❌ Error al configurar la ruta DNS:", err)
		} else {
			fmt.Println("✅ Túnel configurado con éxito.")
		}
	}
}

// Guarda la configuración en ~/.autohost/config.json.
func saveConfig(cfg Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".autohost")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}
	configFile := filepath.Join(configDir, "config.json")

	file, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Stat()
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

// Pide confirmación al usuario.
func confirm(prompt string) bool {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}
