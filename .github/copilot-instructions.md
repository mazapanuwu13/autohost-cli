# GitHub Copilot Instructions — autohost-cli (Go 1.23.0)

## Contexto del proyecto
CLI para automatizar *self-hosting* en Linux (Raspberry Pi / Debian/Ubuntu). Componentes clave: Docker, Caddy, CoreDNS, Tailscale, Cloudflare Tunnel.  
Árbol relevante:
- `cmd/` (comandos cobra)
- `internal/helpers/{app,caddy,cloudflared,docker,initializer,setup,tailscale}`
- `internal/infra/{coredns_docker,terraform_splitdns}`
- `assets/docker/*` (plantillas compose/apps)
- `utils/`, `config/`

## Objetivo para Copilot
- Cuando haya issues con la etiqueta **`autohost-task`**, proponer un PR que:
  1) implemente el fix o feature,  
  2) incluya pruebas (si aplica),  
  3) actualice docs y mensajes de ayuda,  
  4) mantenga idempotencia y detección segura de entorno,  
  5) no rompa flujos existentes.

---

## Reglas de código (Go 1.23.0)
- Versiones: **Go 1.23.0**.
- Formato: `gofmt` / `goimports` obligatorio.
- Lint: `go vet` y `staticcheck` (si disponible).
- Estructura: funciones pequeñas, errores envueltos con contexto (`fmt.Errorf("context: %w", err)`).
- Logs: usar emojis breves en UX del CLI, pero logs técnicos deben ser claros y sin ruido. No exponer secretos.
- Sistemas de archivos: preferir rutas en `$HOME/.autohost` (crearlas si faltan). Permitir override por env:
  - `AUTOHOST_HOME` (default: `~/.autohost`)
  - `AUTOHOST_CADDY_DIR` (default: `/etc/caddy` si existe y hay systemd; si no, `~/.autohost/caddy`)
  - `AUTOHOST_COREDNS_DIR` (default: `~/.autohost/coredns`)
- Concurrencia: evitar, salvo necesidad. Si se usa, usar context con timeout/cancel.

---

## Estilo de CLI
- Cobra: comandos con `Use`, `Short`, `Long`, `Example`.
- Mensajes al usuario:
  - Claros, con acción siguiente (“ejecuta…”, “reinicia sesión”).
  - Nunca fallar silenciosamente. Mostrar causa y recomendación.
- Flags:
  - Elegir buen default. Validar input.
  - **No** pedir datos que se puedan auto-descubrir (IP, rutas, etc.).
- Idempotencia:
  - Repetir comandos no debe romper estado (crear si no existe, actualizar si cambió, *no duplicar*).

---

## Sistemas/Permisos
- Detectar root: `os.Geteuid() == 0`.
- Para tareas que requieren privilegios (`apt`, mover archivos a `/etc/*`, `systemctl`):
  - Si no es root:
    - Intentar con `sudo` (verificar binario presente).
    - Si no hay `sudo`, dar instrucciones claras.
- Grupo `docker`:
  - Tras instalación, ejecutar `usermod -aG docker $USER` (si no está ya).
  - Mensajear: “cierra sesión y vuelve a entrar (o reboot) para aplicar el grupo”.

---

