package caddy

import (
	"autohost-cli/utils"
	"fmt"
	"os"
	"os/exec"
)

func InstallCaddy() {
	fmt.Println("🚀 Instalando Caddy...")
	utils.ExecShell(`
	sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl &&
		curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg &&
		curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list &&
		sudo apt update && sudo apt install caddy
	`)
	utils.ExecShell("sudo systemctl enable caddy")
	utils.ExecShell("sudo systemctl start caddy")
	fmt.Println("✅ Caddy instalado y activado correctamente.")
}

func CreateCaddyfile() {
	caddyfilePath := "/etc/caddy/Caddyfile"

	if _, err := os.Stat(caddyfilePath); err == nil {
		fmt.Println("📄 Ya existe un Caddyfile, no se modificará.")
		return
	}

	content := `
http://localhost {
	respond \"🚀 AutoHost CLI: Caddy instalado y funcionando\"
}
`
	err := os.WriteFile(caddyfilePath, []byte(content), 0644)
	if err != nil {
		fmt.Println("❌ Error creando Caddyfile:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Caddyfile creado en /etc/caddy/Caddyfile")

	reloadCmd := exec.Command("sudo", "systemctl", "reload", "caddy")
	reloadCmd.Stdout = os.Stdout
	reloadCmd.Stderr = os.Stderr
	if err := reloadCmd.Run(); err != nil {
		fmt.Println("⚠️ No se pudo recargar Caddy automáticamente. Hazlo manualmente con: sudo systemctl reload caddy")
	} else {
		fmt.Println("🔁 Caddy recargado con éxito.")
	}
}
