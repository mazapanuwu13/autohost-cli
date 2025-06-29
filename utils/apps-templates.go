package utils

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed templates/nextcloud/*
var nextcloudFS embed.FS

func InstallNextcloud() error {
	// Leer archivo embebido
	data, err := nextcloudFS.ReadFile("templates/nextcloud/compose.yml")
	if err != nil {
		return fmt.Errorf("error leyendo plantilla embebida: %w", err)
	}

	// Ruta de destino final
	dest := filepath.Join(GetAutohostDir(), "opt", "autohost", "docker", "nextcloud.yml")

	// Crear el directorio si no existe
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("error creando directorio destino: %w", err)
	}

	// Escribir el contenido del archivo embebido al destino
	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("error escribiendo archivo destino: %w", err)
	}

	fmt.Println("✅ Nextcloud instalado correctamente en:", dest)
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
