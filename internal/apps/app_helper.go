package apps

import (
	"fmt"
	"os/exec"
	"strings"
)

// StartApp ejecuta docker compose up -d para una app
func StartApp(app string) error {
	cmd := exec.Command("docker", "compose", "-f", (app), "up", "-d")
	return cmd.Run()
}

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
	return fmt.Sprintf("%s/.autohost/docker/compose/%s.yml", getHomeDir(), app)
}

// getHomeDir obtiene el directorio home del usuario
func getHomeDir() string {
	home, err := exec.Command("sh", "-c", "echo $HOME").Output()
	if err != nil {
		return "/root"
	}
	return strings.TrimSpace(string(home))
}
