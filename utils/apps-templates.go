package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func InstallNextcloud() error {
	// Obtener el directorio home del usuario
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("no se pudo obtener el directorio home del usuario: %w", err)
	}

	// Ruta fuente y destino construidas de forma segura
	src := filepath.Join(homeDir, "go", "src", "github.com", "mazapanuwu13", "autohost-cli", "templates", "nextcloud", "compose.yml")
	dest := filepath.Join(GetAutohostDir(), "docker", "compose", "nextcloud.yml")

	// Asegúrate de que el directorio de destino exista
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("error creando directorio de destino: %w", err)
	}

	fmt.Println("Usando archivo fuente:", src)

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

	fmt.Println("✅ Nextcloud instalado correctamente.")
	return nil
}

func StartApp(app string) error {
	ymlPath := filepath.Join(GetAutohostDir(), "docker", "compose", app+".yml")

	cmd := exec.Command("docker", "compose", "-f", ymlPath, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("🔄 Levantando aplicación con Docker...")
	return cmd.Run()
}
