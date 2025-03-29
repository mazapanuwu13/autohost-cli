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
	Short: "Configura un pipeline de GitHub Actions",
	Long:  `Este comando te ayuda a automatizar la construcción y despliegue de tus proyectos mediante un workflow de GitHub Actions.`,
}

// ciInitCmd crea o actualiza el archivo de workflow en .github/workflows/autohost.yml.
var ciInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa el workflow de GitHub Actions en .github/workflows/autohost.yml",
	Long:  `Crea o actualiza el archivo de workflow para automatizar la build y el deploy de tu aplicación al hacer push a la rama main.`,
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

	fmt.Println("Bienvenido al configurador de GitHub Actions para AutoHost CLI.")
	fmt.Println("Este comando creará o actualizará el archivo .github/workflows/autohost.yml en tu repositorio.")
	fmt.Println("Presiona Enter para omitir campos o usar valores por defecto.\n")

	fmt.Print("Nombre de la imagen Docker (ej: myuser/myapp:latest): ")
	imageName, _ := reader.ReadString('\n')
	imageName = strings.TrimSpace(imageName)
	if imageName == "" {
		imageName = "myuser/myapp:latest"
	}

	fmt.Print("Usuario SSH en el servidor (ej: ubuntu): ")
	sshUser, _ := reader.ReadString('\n')
	sshUser = strings.TrimSpace(sshUser)
	if sshUser == "" {
		sshUser = "ubuntu"
	}

	fmt.Print("Host o IP del servidor (ej: 123.45.67.89): ")
	sshHost, _ := reader.ReadString('\n')
	sshHost = strings.TrimSpace(sshHost)
	if sshHost == "" {
		sshHost = "123.45.67.89"
	}

	fmt.Print("Ruta en el servidor donde se ejecutará docker-compose (ej: /opt/myapp): ")
	remotePath, _ := reader.ReadString('\n')
	remotePath = strings.TrimSpace(remotePath)
	if remotePath == "" {
		remotePath = "/opt/myapp"
	}

	workflowContent := generateWorkflowContent(imageName, sshUser, sshHost, remotePath)

	// Crear el directorio .github/workflows si no existe.
	err := os.MkdirAll(".github/workflows", 0755)
	if err != nil {
		fmt.Println("❌ Error creando el directorio .github/workflows:", err)
		return
	}

	// Crear o sobrescribir el archivo de workflow.
	workflowFile := filepath.Join(".github", "workflows", "autohost.yml")
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
	fmt.Println("Recuerda agregar tu clave privada SSH como secreto en GitHub (SSH_PRIVATE_KEY) para que funcione el deploy.")
	fmt.Println("Más info: https://docs.github.com/en/actions/security-guides/encrypted-secrets")
}

// generateWorkflowContent genera el contenido del workflow de GitHub Actions basado en los parámetros proporcionados.
func generateWorkflowContent(imageName, sshUser, sshHost, remotePath string) string {
	return fmt.Sprintf(`name: AutoHost Deploy

on:
  push:
    branches: [ "main" ]

jobs:
  build-deploy:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Setup SSH
        uses: webfactory/ssh-agent@v0.5.4
        with:
          ssh-private-key: ${{ secrets.SSH_PRIVATE_KEY }}

      - name: Build Docker image
        run: |
          docker build -t %s .
      
      - name: Push Docker image
        run: |
          echo "Skipping push to Docker registry if you don't have credentials"
          # docker push %s

      - name: Deploy to server
        run: |
          ssh -o StrictHostKeyChecking=no %s@%s << EOF
            docker pull %s
            cd %s
            docker compose up -d
          EOF
`, imageName, imageName, sshUser, sshHost, imageName, remotePath)
}
