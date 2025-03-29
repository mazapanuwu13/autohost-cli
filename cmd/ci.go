package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// ciCmd es el comando principal para la configuración de CI/CD con GitHub Actions.
var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Configura un pipeline de GitHub Actions para despliegue local",
	Long:  `Este comando te ayuda a automatizar la construcción y despliegue de tus proyectos en tu servidor en casa mediante un workflow self-hosted.`,
}

// ciInitCmd crea o actualiza el archivo de workflow en .github/workflows/deploy.yml.
var ciInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa el workflow de GitHub Actions en .github/workflows/deploy.yml",
	Long:  `Crea o actualiza el archivo de workflow para automatizar el deploy usando un runner self-hosted en tu servidor local.`,
	Run: func(cmd *cobra.Command, args []string) {
		setupGitHubActionsWorkflow()
	},
}

func init() {
	ciCmd.AddCommand(ciInitCmd)
	rootCmd.AddCommand(ciCmd)
}

// setupGitHubActionsWorkflow interactúa con el usuario para obtener parámetros y genera el workflow.
func setupGitHubActionsWorkflow() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Bienvenido al configurador de GitHub Actions para AutoHost CLI (Runner self-hosted).")
	fmt.Println("Este comando creará o actualizará el archivo .github/workflows/deploy.yml en tu repositorio.")
	fmt.Println("Presiona Enter para usar valores por defecto.\n")

	fmt.Print("Nombre de la rama que disparará el deploy (por defecto: main): ")
	branch, _ := reader.ReadString('\n')
	branch = strings.TrimSpace(branch)
	if branch == "" {
		branch = "main"
	}

	workflowContent := generateWorkflowContent(branch)

	// Crear el directorio .github/workflows si no existe.
	err := os.MkdirAll(".github/workflows", 0755)
	if err != nil {
		fmt.Println("❌ Error creando el directorio .github/workflows:", err)
		return
	}

	// Crear o sobrescribir el archivo de workflow (en este ejemplo lo llamamos deploy.yml).
	workflowFile := filepath.Join(".github", "workflows", "deploy.yml")
	file, err := os.Create(workflowFile)
	if err != nil {
		fmt.Println("❌ Error creando el archivo de workflow:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(workflowContent)
	if err != nil {
		fmt.Println("❌ Error escribiendo en el archivo de workflow:", err)
		return
	}

	fmt.Printf("✅ Archivo de workflow creado/actualizado: %s\n", workflowFile)
	fmt.Println("Recuerda configurar tu runner self-hosted en el servidor donde se ejecutará el deploy.")
}

// generateWorkflowContent genera el contenido del workflow de GitHub Actions para un runner self-hosted.
func generateWorkflowContent(branch string) string {
	return fmt.Sprintf(`name: Deploy to Home Server

on:
  push:
    branches: [ "%s" ]

jobs:
  deploy:
    runs-on: self-hosted

    steps:
      - name: Checkout repository
        uses: actions/checkout@v3

      - name: Stop existing container (if running)
        run: |
          docker compose down || true

      - name: Build and start container
        run: |
          docker compose up --build -d
`, branch)
}
