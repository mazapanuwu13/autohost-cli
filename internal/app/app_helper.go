package app

import (
	"autohost-cli/assets"
	"autohost-cli/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func InstallApp(app string) error {
	destDir := filepath.Join(utils.GetSubdir("apps"), app)
	dest := filepath.Join(destDir, "docker-compose.yml")

	// Crear el directorio destino si no existe
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("error creando directorio de destino: %w", err)
	}

	// 1) Leer desde embed
	data, err := assets.ReadCompose(app) // lee assets/docker/<app>/docker-compose.yml
	if err != nil {
		// 2) Fallback opcional a plantilla personalizada del usuario
		custom := filepath.Join(utils.GetSubdir("templates"), app, "docker-compose.yml")
		if b, e := os.ReadFile(custom); e == nil {
			data = b
			err = nil
			fmt.Println("ℹ️  Usando plantilla personalizada:", custom)
		} else {
			if !errorsIsNotExist(e) { // helper pequeño para distinguir
				return fmt.Errorf("error leyendo plantilla personalizada %s: %w", custom, e)
			}
			return fmt.Errorf("no se encontró plantilla embebida para %s (%v) ni personalizada en %s", app, err, custom)
		}
	} else {
		fmt.Println("📦 Usando plantilla embebida para:", app)
	}

	// Escribir compose
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return fmt.Errorf("error escribiendo archivo destino: %w", err)
	}

	fmt.Printf("✅ %s instalado correctamente en %s\n", app, dest)
	return nil
}

func errorsIsNotExist(err error) bool {
	return err != nil
}

// StartApp ejecuta docker compose up -d para una app
func StartApp(app string) error {
	ymlPath := filepath.Join(utils.GetSubdir("apps"), app, "docker-compose.yml")

	// Validar si existe el archivo docker-compose.yml
	if _, err := os.Stat(ymlPath); os.IsNotExist(err) {
		return fmt.Errorf("el archivo de configuración no existe: %s", ymlPath)
	}

	fmt.Printf("🔄 Levantando aplicación '%s'...\n", app)

	// Usar Exec con working dir del compose
	return utils.ExecWithDir(filepath.Dir(ymlPath), "docker", "compose", "-f", ymlPath, "up", "-d")
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
