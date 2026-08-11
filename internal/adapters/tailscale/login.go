package tailscale

import (
	"fmt"
	"os"
	"os/exec"
)

func LoginTailscale() error {
	return LoginHeadscale("https://hs.autohst.dev", "")
}

func LoginHeadscale(loginServer, authKey string) error {
	if loginServer == "" {
		loginServer = "https://hs.autohst.dev"
	}

	fmt.Printf("🔐 Autenticando con Headscale/Tailscale (%s)...\n", loginServer)

	args := []string{"tailscale", "up", "--login-server", loginServer}
	if authKey != "" {
		args = append(args, "--authkey", authKey)
	}

	loginCmd := exec.Command("sudo", args...)
	loginCmd.Stdout = os.Stdout
	loginCmd.Stderr = os.Stderr

	if err := loginCmd.Run(); err != nil {
		fmt.Println("❌ Error al conectar con Headscale/Tailscale:", err)
		return err
	}

	fmt.Println("✅ Conectado a Headscale/Tailscale.")
	return nil
}

