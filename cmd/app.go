package cmd

import (
	"fmt"

	"autohost-cli/internal/app"
	"autohost-cli/utils"

	"github.com/spf13/cobra"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Gestión de aplicaciones autohospedadas",
}

var appInstallCmd = &cobra.Command{
	Use:   "install [nombre]",
	Short: "Instala una aplicación (por ejemplo: nextcloud, bookstack, etc.)",
	Args:  cobra.ExactArgs(1),
	Run: utils.WithAppName(func(appName string) {

		if err := app.InstallApp(appName); err != nil {
			fmt.Printf("❌ Error al instalar %s: %v\n", appName, err)
			return
		}

		fmt.Printf("✅ %s instalado correctamente. Revisa ~/autohost/docker/compose/%s.yml\n", appName, appName)

		if utils.Confirm(fmt.Sprintf("¿Deseas levantar %s ahora con Docker? [y/N]: ", appName)) {
			if err := app.StartApp(appName); err != nil {
				fmt.Printf("❌ Error al iniciar %s: %v\n", appName, err)
			} else {
				fmt.Printf("🚀 %s está corriendo en http://localhost:8080\n", appName)
			}
		}
	}),
}

var appStartCmd = &cobra.Command{
	Use:   "start [nombre]",
	Short: "Inicia una aplicación",
	Args:  cobra.ExactArgs(1),
	Run: utils.WithAppName(func(appName string) {
		err := app.StartApp(appName)
		if err != nil {
			fmt.Printf("❌ No se pudo iniciar %s: %v\n", appName, err)
		} else {
			fmt.Printf("🚀 %s iniciada correctamente.\n", appName)
		}
	}),
}

var appStopCmd = &cobra.Command{
	Use:   "stop [nombre]",
	Short: "Detiene una aplicación",
	Args:  cobra.ExactArgs(1),
	Run: utils.WithAppName(func(appName string) {
		err := app.StopApp(appName)

		if err != nil {
			fmt.Printf("❌ No se pudo detener %s: %v\n", appName, err)
		} else {
			fmt.Printf("🛑 %s detenida.\n", appName)
		}
	}),
}

var appRemoveCmd = &cobra.Command{
	Use:   "remove [nombre]",
	Short: "Elimina una aplicación",
	Args:  cobra.ExactArgs(1),
	Run: utils.WithAppName(func(appName string) {
		if utils.Confirm(fmt.Sprintf("¿Estás seguro que quieres eliminar %s? [y/N]: ", appName)) {
			err := app.RemoveApp(appName)
			if err != nil {
				fmt.Printf("❌ No se pudo eliminar %s: %v\n", appName, err)
			} else {
				fmt.Printf("🧹 %s eliminada correctamente.\n", appName)
			}
		}
	}),
}

var appStatusCmd = &cobra.Command{
	Use:   "status [nombre]",
	Short: "Muestra el estado de una aplicación",
	Args:  cobra.ExactArgs(1),
	Run: utils.WithAppName(func(appName string) {
		status, err := app.GetAppStatus(appName)
		if err != nil {
			fmt.Printf("❌ No se pudo obtener el estado de %s: %v\n", appName, err)
		} else {
			fmt.Printf("📊 Estado de %s: %s\n", appName, status)
		}
	}),
}

func init() {
	appCmd.AddCommand(appInstallCmd)
	appCmd.AddCommand(appStartCmd)
	appCmd.AddCommand(appStopCmd)
	appCmd.AddCommand(appRemoveCmd)
	appCmd.AddCommand(appStatusCmd)
	rootCmd.AddCommand(appCmd)
}
