package tailscale

import (
	"autohost-cli/utils"
	"fmt"
	"os/exec"
	"strings"
)

func InstallTailscale() {
	fmt.Println("🔐 Instalando Tailscale...")
	utils.ExecShell("curl -fsSL https://tailscale.com/install.sh | sh")
	fmt.Println("🔐 Autenticándote con Tailscale...")
	utils.ExecShell("sudo tailscale up")
}

func TailscaleIP() (string, error) {
	out, err := exec.Command("tailscale", "ip", "-4").Output()
	if err != nil {
		return "", err
	}
	ip := strings.TrimSpace(string(out))
	if i := strings.IndexByte(ip, '\n'); i > -1 {
		ip = ip[:i]
	}
	return ip, nil
}
