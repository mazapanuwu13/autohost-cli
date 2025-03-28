package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

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

		fmt.Println("🚀 Tu sistema está listo para desplegar servicios.")
		fmt.Println("👉 Próximamente: configuración de túneles y dominios.")
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

// Verifica si Docker está instalado
func dockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// Instala Docker usando el script oficial
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

// Pide confirmación al usuario
func confirm(prompt string) bool {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}
