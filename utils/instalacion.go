package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Verifica si Docker está instalado
func DockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// Instala Docker usando el script oficial
func InstallDocker() {
	cmd := exec.Command("sh", "-c", "curl -fsSL https://get.docker.com | sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("❌ Error al instalar Docker: " + err.Error())
	}
	fmt.Println("✅ Docker instalado con éxito.")
}

// Añade al usuario actual al grupo docker
func AddUserToDockerGroup() {
	user := os.Getenv("SUDO_USER")
	if user == "" {
		user = os.Getenv("USER")
	}
	if user == "" {
		fmt.Println("⚠️ No se pudo determinar el usuario.")
		return
	}
	cmd := exec.Command("sudo", "usermod", "-aG", "docker", user)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	fmt.Printf("✅ Usuario '%s' agregado al grupo docker.\n", user)
}

// Guarda estado en ~/.autohost/state/status.json
func SaveStatus(key string, value interface{}) error {
	status := make(map[string]interface{})
	path := filepath.Join(GetAutohostDir(), "state", "status.json")

	// Leer estado actual si existe
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &status)
	}

	status[key] = value
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Pide confirmación al usuario.
func Confirm(prompt string) bool {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}
