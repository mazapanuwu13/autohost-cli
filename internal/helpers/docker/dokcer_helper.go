package docker

import (
	"autohost-cli/utils"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

func DockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func runningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	// opcional: variable para forzar
	return os.Getenv("AUTOHOST_IN_CONTAINER") == "true"
}

func dockerAvailable() bool { return exec.Command("docker", "version").Run() == nil }

type osRelease struct {
	ID     string
	IDLike string
}

func readOSRelease() osRelease {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return osRelease{}
	}
	defer f.Close()
	kv := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		k := parts[0]
		v := strings.Trim(parts[1], `"'`)
		kv[k] = v
	}
	return osRelease{ID: kv["ID"], IDLike: kv["ID_LIKE"]}
}

func ensureCurl() error {
	osr := readOSRelease()
	id := osr.ID + " " + osr.IDLike

	switch {
	case strings.Contains(id, "debian") || strings.Contains(id, "ubuntu"):
		return utils.ExecShell(`sudo apt-get update -y && sudo apt-get install -y curl ca-certificates && sudo update-ca-certificates`)
	case strings.Contains(id, "rhel") || strings.Contains(id, "centos") || strings.Contains(id, "rocky") || strings.Contains(id, "almalinux"):
		return utils.ExecShell(`sudo yum install -y curl ca-certificates || sudo dnf install -y curl ca-certificates`)
	case strings.Contains(id, "fedora"):
		return utils.ExecShell(`sudo dnf install -y curl ca-certificates`)
	case strings.Contains(id, "amzn"): // Amazon Linux
		return utils.ExecShell(`sudo yum install -y curl ca-certificates || sudo dnf install -y curl ca-certificates`)
	case strings.Contains(id, "alpine"):
		return utils.ExecShell(`sudo apk add --no-cache curl ca-certificates && sudo update-ca-certificates`)
	case strings.Contains(id, "suse") || strings.Contains(id, "sles") || strings.Contains(id, "opensuse"):
		return utils.ExecShell(`sudo zypper --non-interactive install -y curl ca-certificates`)
	default:
		// mejor intentar y que falle claro
		return utils.Exec("which", "curl")
	}
}

func systemctlAvailable() bool { return exec.Command("which", "systemctl").Run() == nil }

func InstallDocker() {
	if runningInContainer() {
		fmt.Println("⚠️  Detecté contenedor. No instalo Docker aquí. Usa el socket del host o dind para pruebas.")
		return
	}
	if dockerAvailable() {
		fmt.Println("✅ Docker ya está instalado.")
		return
	}
	fmt.Println("🔄 Instalando Docker...")

	// Asegura curl
	if err := ensureCurl(); err != nil {
		panic("❌ No pude instalar/ubicar curl: " + err.Error())
	}

	// Script oficial SIN pipe ciego
	if err := utils.ExecShell(`
set -e
tmp="$(mktemp)"
curl -fsSL https://get.docker.com -o "$tmp"
sh "$tmp"
rm -f "$tmp"
`); err != nil {
		panic("❌ Error ejecutando el instalador de Docker: " + err.Error())
	}

	// Arrancar/enable del daemon (si hay systemd)
	if systemctlAvailable() {
		_ = utils.Exec("sudo", "systemctl", "enable", "--now", "docker")
	} else {
		// fallback best-effort
		_ = utils.Exec("sudo", "service", "docker", "start")
	}

	// Verificar CLI + daemon
	if err := exec.Command("docker", "--version").Run(); err != nil {
		panic("❌ Docker CLI no quedó instalado correctamente.")
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		fmt.Println("⚠️  Docker instalado, pero el daemon no responde aún. Revisa el servicio o reinicia el host.")
	} else {
		fmt.Println("✅ Docker instalado y en ejecución.")
	}
}

func AddUserToDockerGroup() {
	// Si eres root en servidor, agrega al usuario “real” si existe.
	// En contenedor o siendo root sin usuario objetivo, omite.
	if runningInContainer() {
		fmt.Println("⚠️  En contenedor no modifico grupos. Omite este paso.")
		return
	}
	current, _ := user.Current()
	uid0 := current != nil && current.Uid == "0"

	// Detecta usuario adecuado:
	u := os.Getenv("SUDO_USER")
	if u == "" && !uid0 && current != nil {
		u = current.Username
	}
	if u == "" || u == "root" {
		fmt.Println("ℹ️  Saltando: no hay usuario no-root claro para agregar a 'docker'.")
		return
	}

	// Crea grupo si falta y agrega usuario
	if err := utils.ExecShell(`getent group docker >/dev/null 2>&1 || sudo groupadd docker`); err != nil {
		fmt.Println("⚠️  No pude crear/verificar grupo docker:", err)
	}
	if err := utils.Exec("sudo", "usermod", "-aG", "docker", u); err != nil {
		fmt.Printf("⚠️  No pude agregar el usuario '%s' al grupo docker: %v\n", u, err)
		return
	}
	fmt.Printf("✅ Usuario '%s' agregado al grupo 'docker'. Cierra sesión y vuelve a entrar para aplicar cambios.\n", u)
}
