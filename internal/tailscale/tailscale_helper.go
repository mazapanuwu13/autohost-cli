package tailscale

import (
	"autohost-cli/utils"
	"fmt"
)

func InstallTailscale() {
	fmt.Println("🔐 Instalando Tailscale...")
	utils.ExecShell("curl -fsSL https://tailscale.com/install.sh | sh")
	fmt.Println("🔐 Autenticándote con Tailscale...")
	utils.ExecShell("sudo tailscale up")
}
