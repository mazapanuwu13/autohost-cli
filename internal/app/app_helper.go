package app

import (
	"autohost-cli/assets"
	"autohost-cli/utils"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func InstallApp(app string) error {
	appDir := filepath.Join(utils.GetSubdir("apps"), app)
	composePath := filepath.Join(appDir, "docker-compose.yml")
	envPath := filepath.Join(appDir, ".env")

	// Crear el directorio destino
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return fmt.Errorf("error creando directorio de destino: %w", err)
	}

	// === 1) Compose: embebido con fallback opcional ===
	data, err := assets.ReadCompose(app) // assets/docker/<app>/docker-compose.yml
	if err != nil {
		// Fallback: plantilla personalizada del usuario (opcional)
		custom := filepath.Join(utils.GetSubdir("templates"), app, "docker-compose.yml")
		if b, e := os.ReadFile(custom); e == nil {
			data = b
			fmt.Println("ℹ️  Usando plantilla personalizada:", custom)
		} else {
			if !errors.Is(e, os.ErrNotExist) {
				return fmt.Errorf("error leyendo plantilla personalizada %s: %w", custom, e)
			}
			return fmt.Errorf("no se encontró plantilla embebida para %s (%v) ni personalizada en %s", app, err, custom)
		}
	} else {
		fmt.Println("📦 Usando plantilla embebida para:", app)
	}

	if err := os.WriteFile(composePath, data, 0o644); err != nil {
		return fmt.Errorf("error escribiendo docker-compose.yml: %w", err)
	}

	// === 2) .env: crear desde .env.example si no existe ===
	if _, err := os.Stat(envPath); errors.Is(err, os.ErrNotExist) {
		// Intentar leer .env.example embebido
		if example, e := assets.ReadEnvExample(app); e == nil {
			values := map[string]string{}

			// Genera APP_KEY solo si el ejemplo lo pide
			if strings.Contains(string(example), "{{APP_KEY}}") {
				if key, genErr := utils.GenerateLaravelAppKey(); genErr == nil {
					values["APP_KEY"] = key
				} else {
					return fmt.Errorf("no se pudo generar APP_KEY: %w", genErr)
				}
			}

			// Si quieres agregar más placeholders globales, hazlo aquí:
			// p.ej. PUERTOS ALEATORIOS, PASSWORDS, ETC.
			// if strings.Contains(string(example), "{{MYSQL_PASSWORD}}") {
			//     values["MYSQL_PASSWORD"] = utils.GeneratePassword(20)
			// }

			final := utils.ReplacePlaceholders(string(example), values)
			if writeErr := os.WriteFile(envPath, []byte(final), 0o600); writeErr != nil {
				return fmt.Errorf("error escribiendo .env: %w", writeErr)
			}
			fmt.Println("✅ .env generado desde .env.example")
		} else if errors.Is(e, os.ErrNotExist) {
			// Si la app no trae .env.example, crea uno vacío
			if writeErr := os.WriteFile(envPath, []byte("# .env generado por autohost\n"), 0o600); writeErr != nil {
				return fmt.Errorf("error creando .env vacío: %w", writeErr)
			}
			fmt.Println("ℹ️  Sin .env.example embebido; se creó .env vacío.")
		} else {
			return fmt.Errorf("error leyendo .env.example embebido: %w", e)
		}
	} else {
		fmt.Println("ℹ️  .env ya existe; no se sobrescribe.")
	}

	fmt.Printf("✅ %s instalado correctamente en %s\n", app, appDir)
	return nil
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
