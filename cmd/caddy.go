package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"autohost-cli/utils"

	"github.com/spf13/cobra"
)

var caddyCmd = &cobra.Command{
	Use:   "caddy",
	Short: "Comandos para instalar y administrar el servidor Caddy",
}

var caddyInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Instala el servidor Caddy y prepara su configuración",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsInitialized() {
			fmt.Println("⚠️ AutoHost no está inicializado. Ejecuta `autohost init` primero.")
			return
		}

		homeDir, _ := os.UserHomeDir()
		caddyDir := filepath.Join(homeDir, ".autohost", "caddy")
		caddyfilePath := filepath.Join(caddyDir, "Caddyfile")

		err := os.MkdirAll(caddyDir, 0755)
		if err != nil {
			fmt.Println("❌ No se pudo crear el directorio de configuración de Caddy:", err)
			return
		}

		fmt.Println("📦 Instalando Caddy...")

		installScript := `
		sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl &&
		curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg &&
		curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list &&
		sudo apt update &&
		sudo apt install caddy
	`

		installCmd := exec.Command("bash", "-c", installScript)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr

		if err := installCmd.Run(); err != nil {
			fmt.Println("❌ Error al instalar Caddy:", err)
			return
		}

		// Crear Caddyfile si no existe
		if _, err := os.Stat(caddyfilePath); os.IsNotExist(err) {
			base := `# Archivo de configuración de Caddy para AutoHost
# Ejemplo:
# plex.localhost {
#     reverse_proxy 127.0.0.1:32400
# }
`
			os.WriteFile(caddyfilePath, []byte(base), 0644)
		}

		fmt.Println("✅ Caddy instalado y configurado. Puedes editar tu archivo en:")
		fmt.Println("   ", caddyfilePath)
	},
}

// Flags
var (
	serviceName string
	servicePort int
	serviceHost string
)

var caddyAddServiceCmd = &cobra.Command{
	Use:   "add-service",
	Short: "Agrega un nuevo servicio al archivo Caddyfile",
	Run: func(cmd *cobra.Command, args []string) {
		homeDir := utils.GetAutohostDir()
		caddyfilePath := filepath.Join(homeDir, ".autohost", "caddy", "Caddyfile")

		block := fmt.Sprintf(`

%s {
    reverse_proxy 127.0.0.1:%d
}
`, serviceHost, servicePort)

		contentBytes, err := os.ReadFile(caddyfilePath)
		if err != nil {
			fmt.Println("❌ No se pudo leer el archivo Caddyfile:", err)
			return
		}
		content := string(contentBytes)

		if strings.Contains(content, serviceHost) {
			fmt.Printf("⚠️ Ya existe una entrada para %s en el Caddyfile.\n", serviceHost)
			return
		}

		file, err := os.OpenFile(caddyfilePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("❌ No se pudo abrir el archivo Caddyfile:", err)
			return
		}
		defer file.Close()

		_, err = file.WriteString(block)
		if err != nil {
			fmt.Println("❌ No se pudo escribir en el archivo Caddyfile:", err)
			return
		}

		fmt.Printf("✅ Servicio '%s' agregado exitosamente a Caddyfile.\n", serviceName)
	},
}

var caddyStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia Caddy con el archivo de configuración de AutoHost",
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, _ := os.UserHomeDir()
		caddyfilePath := filepath.Join(homeDir, ".autohost", "caddy", "Caddyfile")

		fmt.Println("🚀 Iniciando servidor Caddy...")
		startCmd := exec.Command("caddy", "run", "--config", caddyfilePath)
		startCmd.Stdout = os.Stdout
		startCmd.Stderr = os.Stderr
		err := startCmd.Run()
		if err != nil {
			fmt.Println("❌ Error al iniciar Caddy:", err)
		} else {
			fmt.Println("✅ Caddy iniciado correctamente.")
		}
	},
}

func init() {
	// Flags para add-service
	caddyAddServiceCmd.Flags().StringVar(&serviceName, "name", "", "Nombre del servicio (ej: plex)")
	caddyAddServiceCmd.Flags().IntVar(&servicePort, "port", 0, "Puerto local del servicio")
	caddyAddServiceCmd.Flags().StringVar(&serviceHost, "host", "", "Host local (ej: plex.localhost)")
	caddyAddServiceCmd.MarkFlagRequired("name")
	caddyAddServiceCmd.MarkFlagRequired("port")
	caddyAddServiceCmd.MarkFlagRequired("host")

	// Agregar comandos a caddy
	caddyCmd.AddCommand(caddyInstallCmd)
	caddyCmd.AddCommand(caddyAddServiceCmd)
	caddyCmd.AddCommand(caddyStartCmd)

	// Agregar grupo caddy al root
	rootCmd.AddCommand(caddyCmd)
}
