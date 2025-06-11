package cmd

import (
	"fmt"

	"autohost-cli/internal/apps"
	"autohost-cli/utils"

	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "app",
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

var appsRunCmd = &cobra.Command{
	Use:   "run [nombre]",
	Short: "Inicia una aplicación",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app := args[0]
		err := utils.StartApp(app)
		if err != nil {
			fmt.Printf("❌ No se pudo iniciar %s: %v\n", app, err)
		} else {
			fmt.Printf("🚀 %s iniciada correctamente.\n", app)
		}
	},
}

var appsStopCmd = &cobra.Command{
	Use:   "stop [nombre]",
	Short: "Detiene una aplicación",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app := args[0]
		err := apps.StopApp(app)

		if err != nil {
			fmt.Printf("❌ No se pudo detener %s: %v\n", app, err)
		} else {
			fmt.Printf("🛑 %s detenida.\n", app)
		}
	},
}

var appsRemoveCmd = &cobra.Command{
	Use:   "remove [nombre]",
	Short: "Elimina una aplicación",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app := args[0]
		if utils.Confirm(fmt.Sprintf("¿Estás seguro que quieres eliminar %s? [y/N]: ", app)) {
			err := apps.RemoveApp(app)
			if err != nil {
				fmt.Printf("❌ No se pudo eliminar %s: %v\n", app, err)
			} else {
				fmt.Printf("🧹 %s eliminada correctamente.\n", app)
			}
		}
	},
}

var appsStatusCmd = &cobra.Command{
	Use:   "status [nombre]",
	Short: "Muestra el estado de una aplicación",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app := args[0]
		status, err := apps.GetAppStatus(app)
		if err != nil {
			fmt.Printf("❌ No se pudo obtener el estado de %s: %v\n", app, err)
		} else {
			fmt.Printf("📊 Estado de %s: %s\n", app, status)
		}
	},
}

func init() {
	appsCmd.AddCommand(appsInstallCmd)
	appsCmd.AddCommand(appsRunCmd)
	appsCmd.AddCommand(appsStopCmd)
	appsCmd.AddCommand(appsRemoveCmd)
	appsCmd.AddCommand(appsStatusCmd)
	rootCmd.AddCommand(appsCmd)
}
