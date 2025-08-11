// cmd/init.go
package cmd

import (
	"autohost-cli/internal/helpers/initializer"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa el entorno de AutoHost en ~/.autohost",
	Run: func(cmd *cobra.Command, args []string) {
		err := initializer.EnsureAutohostDirs()
		if err != nil {
			fmt.Println("❌ Error al crear estructura de carpetas:", err)
			os.Exit(1)
		}
		fmt.Println("✅ Entorno de AutoHost creado")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
