package tailscale

import (
	"fmt"
	"os"
	"os/exec"
)

func DownTailscale() error {
	fmt.Println("🔌 Desconectando interfaz VPN (tailscale down)...")

	downCmd := exec.Command("sudo", "tailscale", "down")
	downCmd.Stdout = os.Stdout
	downCmd.Stderr = os.Stderr

	if err := downCmd.Run(); err != nil {
		fmt.Println("❌ Error al desconectar VPN:", err)
		return err
	}

	fmt.Println("✅ VPN desconectada temporalmente (credenciales conservadas).")
	return nil
}
