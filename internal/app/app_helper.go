package app

import (
	"autohost-cli/utils"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func InstallApp(app string) error {
	src := filepath.Join(utils.GetSubdir("templates"), app, "docker-compose.yml")
	dest := filepath.Join(utils.GetSubdir("apps"), app, "docker-compose.yml")

	// Crear el directorio destino si no existe
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("error creando directorio de destino: %w", err)
	}

	fmt.Println("📦 Instalando aplicación:", app)
	fmt.Println("📁 Usando archivo fuente:", src)

	// Abrir archivo fuente
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error abriendo archivo fuente: %w", err)
	}
	defer srcFile.Close()

	// Crear archivo destino
	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("error creando archivo destino: %w", err)
	}
	defer destFile.Close()

	// Copiar contenido
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("error copiando contenido: %w", err)
	}

	fmt.Printf("✅ %s instalado correctamente en %s\n", app, dest)
	return nil
}

// StartApp ejecuta docker compose up -d para una app
func StartApp(app string) error {
	ymlPath := filepath.Join(utils.GetSubdir("apps"), app, "docker-compose.yml")

	// Validar si existe el archivo docker-compose.yml
	if _, err := os.Stat(ymlPath); os.IsNotExist(err) {
		return fmt.Errorf("el archivo de configuración no existe: %s", ymlPath)
	}

	cmd := exec.Command("docker", "compose", "-f", ymlPath, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("🔄 Levantando aplicación '%s'...\n", app)
	return cmd.Run()
}

//modificar todas las ficniones de abajo

// StopApp ejecuta docker compose stop para una app
func StopApp(app string) error {
	cmd := exec.Command("docker", "compose", "-f", appComposePath(app), "stop")
	return cmd.Run()
}

// RemoveApp ejecuta docker compose down para una app
func RemoveApp(app string) error {
	cmd := exec.Command("docker", "compose", "-f", appComposePath(app), "down")
	return cmd.Run()
}

// GetAppStatus devuelve si los contenedores están "running", "exited", etc.
func GetAppStatus(app string) (string, error) {
	cmd := exec.Command("docker", "compose", "-f", appComposePath(app), "ps", "--status=running")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if strings.Contains(string(out), "Up") {
		return "en ejecución", nil
	}
	return "detenida", nil
}

// appComposePath devuelve la ruta al archivo docker-compose.yml de la app
func appComposePath(app string) string {
	fmt.Println(app)
	return fmt.Sprintf("%s/docker/compose/%s.yml", utils.GetAutohostDir(), app)
}

// func TemplateExists(appName string) bool {
// 	path := filepath.Join(utils.GetTemplateDir(), appName, "docker-compose.yml")
// 	_, err := os.Stat(path)
// 	return err == nil
// }
