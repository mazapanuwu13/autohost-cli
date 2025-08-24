package app

import (
	"autohost-cli/assets"
	"autohost-cli/utils"
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ComposeService representa un servicio en docker-compose.yml
type ComposeService struct {
	Ports []string `yaml:"ports"`
}

// ComposeFile representa la estructura básica de un docker-compose.yml
type ComposeFile struct {
	Services map[string]ComposeService `yaml:"services"`
}

// PortInfo contiene información sobre los puertos detectados
type PortInfo struct {
	HostPorts []string
	Message   string
}

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

// DetectAppPorts analiza el docker-compose.yml y .env para detectar puertos
func DetectAppPorts(app string) PortInfo {
	appDir := filepath.Join(utils.GetSubdir("apps"), app)
	composePath := filepath.Join(appDir, "docker-compose.yml")
	envPath := filepath.Join(appDir, ".env")

	// Leer archivo docker-compose.yml
	composeData, err := os.ReadFile(composePath)
	if err != nil {
		return PortInfo{Message: "No se pudo leer docker-compose.yml"}
	}

	// Leer variables de entorno del .env
	envVars := make(map[string]string)
	if envData, err := os.ReadFile(envPath); err == nil {
		envVars = parseEnvFile(string(envData))
	}

	// Parsear docker-compose.yml
	var compose ComposeFile
	if err := yaml.Unmarshal(composeData, &compose); err != nil {
		return PortInfo{Message: "No se pudo parsear docker-compose.yml"}
	}

	var hostPorts []string
	for serviceName, service := range compose.Services {
		for _, portMapping := range service.Ports {
			if hostPort := extractHostPort(portMapping, envVars); hostPort != "" {
				hostPorts = append(hostPorts, hostPort)
			}
		}
		_ = serviceName // Avoid unused variable warning if needed
	}

	if len(hostPorts) == 0 {
		return PortInfo{Message: "Sin puertos externos configurados"}
	}

	// Generar mensaje con los puertos encontrados
	if len(hostPorts) == 1 {
		return PortInfo{
			HostPorts: hostPorts,
			Message:   fmt.Sprintf("corriendo en http://localhost:%s", hostPorts[0]),
		}
	} else {
		return PortInfo{
			HostPorts: hostPorts,
			Message:   fmt.Sprintf("corriendo en puertos: %s", strings.Join(hostPorts, ", ")),
		}
	}
}

// parseEnvFile analiza un archivo .env y devuelve un map de variables
func parseEnvFile(content string) map[string]string {
	vars := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			vars[key] = value
		}
	}

	return vars
}

// extractHostPort extrae el puerto del host de un mapeo de puertos como "8080:80" o "${PORT}:80"
func extractHostPort(portMapping string, envVars map[string]string) string {
	// Resolver variables de entorno en el mapeo de puertos
	resolved := resolveEnvVars(portMapping, envVars)

	// Extraer puerto del host de mapeos como "8080:80" o "127.0.0.1:8080:80"
	parts := strings.Split(resolved, ":")
	if len(parts) >= 2 {
		// Si tiene formato IP:HOST_PORT:CONTAINER_PORT, tomar el del medio
		if len(parts) == 3 {
			return parts[1]
		}
		// Si tiene formato HOST_PORT:CONTAINER_PORT, tomar el primero
		if port := strings.TrimSpace(parts[0]); isValidPort(port) {
			return port
		}
	}

	return ""
}

// resolveEnvVars resuelve variables como ${VAR} o $VAR en una cadena
func resolveEnvVars(text string, envVars map[string]string) string {
	// Patrón para ${VAR} y $VAR
	re := regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Z_][A-Z0-9_]*)`)

	return re.ReplaceAllStringFunc(text, func(match string) string {
		var varName string
		if strings.HasPrefix(match, "${") {
			varName = match[2 : len(match)-1] // Remover ${ y }
		} else {
			varName = match[1:] // Remover $
		}

		if value, exists := envVars[varName]; exists {
			return value
		}
		return match // Retornar sin cambios si no se encuentra la variable
	})
}

// isValidPort verifica si una cadena representa un puerto válido
func isValidPort(s string) bool {
	if port, err := strconv.Atoi(s); err == nil {
		return port > 0 && port <= 65535
	}
	return false
}
