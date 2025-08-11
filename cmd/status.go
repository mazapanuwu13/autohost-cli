package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"autohost-cli/internal/helpers/docker"
	"autohost-cli/utils"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Muestra el estado actual de AutoHost",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsInitialized() {
			fmt.Println("⚠️ AutoHost no está inicializado. Ejecuta `autohost init`.")
			return
		}

		fmt.Println("📦 Estado del sistema AutoHost\n")

		// Estado de Docker
		if docker.DockerInstalled() {
			fmt.Println("✅ Docker instalado")
		} else {
			fmt.Println("❌ Docker no está disponible")
		}

		// Leer status.json
		status := loadStatus()

		// Cloudflare
		if status["cloudflare_tunnel"] == true {
			fmt.Println("✅ Cloudflare Tunnel configurado")
		} else {
			fmt.Println("❌ Cloudflare Tunnel no configurado")
		}

		// Dominio
		if domain, ok := status["cloudflare_domain"].(string); ok && domain != "" {
			fmt.Println("🌐 Dominio vinculado:", domain)
		} else {
			fmt.Println("🌐 Ningún dominio vinculado aún")
		}

		// Mostrar ubicación del sistema
		base := utils.GetAutohostDir()
		fmt.Println("\n🛠️ Directorio base:", base)
		fmt.Println("📝 Configuración:", filepath.Join(base, "config.json"))
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func loadStatus() map[string]interface{} {
	status := make(map[string]interface{})
	path := filepath.Join(utils.GetAutohostDir(), "state", "status.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return status
	}

	_ = json.Unmarshal(data, &status)
	return status
}
