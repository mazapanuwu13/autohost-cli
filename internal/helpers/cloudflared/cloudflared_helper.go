package cloudflared

import (
	"autohost-cli/utils"
	"fmt"
)

func InstallCloudflared() {
	fmt.Println("🌐 Instalando Cloudflare Tunnel (cloudflared)...")
	utils.ExecShell(`
		curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o cloudflared &&
		chmod +x cloudflared &&
		sudo mv cloudflared /usr/local/bin/
	`)
	fmt.Println("✅ Cloudflare Tunnel instalado.")
	fmt.Println("ℹ️ Ejecuta 'cloudflared tunnel login' para autenticarte.")
}

func ConfigureCloudflareTunnel(domain string) {
	fmt.Println("⚙️ Configurando Cloudflare Tunnel para:", domain)
	utils.ExecShell("cloudflared tunnel create autohost-tunnel")
	utils.ExecShell(fmt.Sprintf("cloudflared tunnel route dns autohost-tunnel %s", domain))
	fmt.Println("✅ Túnel configurado correctamente.")
}
