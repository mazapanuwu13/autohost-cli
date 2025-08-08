package docker

import (
	"autohost-cli/utils"
	"fmt"
	"os"
	"os/exec"
)

func DockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func InstallDocker() {
	fmt.Println("🔄 Instalando Docker...")
	utils.ExecShell("curl -fsSL https://get.docker.com | sh")
	fmt.Println("✅ Docker instalado con éxito.")
}

func AddUserToDockerGroup() {
	user := os.Getenv("SUDO_USER")
	if user == "" {
		user = os.Getenv("USER")
	}
	if user == "" {
		fmt.Println("⚠️ No se pudo determinar el usuario. Saltando este paso.")
		return
	}
	utils.ExecShell(fmt.Sprintf("sudo usermod -aG docker %s", user))
	fmt.Printf("✅ Usuario '%s' agregado al grupo 'docker'.\n", user)
}
