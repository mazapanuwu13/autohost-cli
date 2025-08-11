package cmd

import (
	"fmt"

	"autohost-cli/internal/helpers/docker"
	"autohost-cli/utils"

	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Comandos para configurar Docker",
}

var dockerInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Instala Docker y lo configura",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsInitialized() {
			fmt.Println("⚠️ Ejecuta `autohost init` primero.")
			return
		}

		if docker.DockerInstalled() {
			fmt.Println("✅ Docker ya está instalado.")
		} else {
			fmt.Println("🔧 Instalando Docker...")
			docker.InstallDocker()
		}

		if utils.Confirm("¿Agregar usuario al grupo docker? [y/N]: ") {
			docker.AddUserToDockerGroup()
		}

		// Guardar estado
		// err := docker.SaveStatus("docker_installed", true)
		// if err != nil {
		// 	fmt.Println("⚠️ No se pudo guardar el estado:", err)
		// } else {
		// 	fmt.Println("📝 Estado de Docker guardado en ~/.autohost/state/status.json")
		// }
	},
}

func init() {
	dockerCmd.AddCommand(dockerInstallCmd)
	rootCmd.AddCommand(dockerCmd)
}
