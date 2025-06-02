package cmd

import (
	"fmt"

	"autohost-cli/utils"

	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Gestión de aplicaciones autohospedadas",
}

var appsInstallCmd = &cobra.Command{
	Use:   "install [nombre]",
	Short: "Instala una aplicación (por ahora: nextcloud)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsInitialized() {
			fmt.Println("⚠️ Ejecuta `autohost init` primero.")
			return
		}

		app := args[0]

		switch app {
		case "nextcloud":
			err := utils.InstallNextcloud()
			if err != nil {
				fmt.Println("❌ Error al instalar Nextcloud:", err)
				return
			}
			fmt.Println("✅ Nextcloud instalado. Revisa ~/.autohost/docker/compose/nextcloud.yml")

			if utils.Confirm("¿Deseas levantar la aplicación ahora con Docker? [y/N]: ") {
				err := utils.StartApp("nextcloud")
				if err != nil {
					fmt.Println("❌ Error al iniciar Nextcloud:", err)
				} else {
					fmt.Println("🚀 Nextcloud está corriendo en http://localhost:8080")
				}
			}

		default:
			fmt.Println("❌ Aplicación no soportada aún:", app)
		}
	},
}

func init() {
	appsCmd.AddCommand(appsInstallCmd)
	rootCmd.AddCommand(appsCmd)
}
