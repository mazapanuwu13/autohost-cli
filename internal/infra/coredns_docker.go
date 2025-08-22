// internal/infra/coredns_docker.go
package infra

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	coreDNSContainer = "coredns-autohost"
	coreDNSImage     = "coredns/coredns:latest"
)

// InstallAndRunCoreDNSWithDocker asegura Docker, genera/actualiza el Corefile para la zona
// y (re)levanta un contenedor de CoreDNS con --network host.
// - zone: apex interno (ej: "maza-server")
// - fqdn: host completo dentro de la zona (ej: "app.maza-server")
// - tailIP: IP de Tailscale del host donde corre CoreDNS (ej: "100.x.y.z")
//
// Devuelve la ruta del Corefile y error (si aplica).
func InstallAndRunCoreDNSWithDocker(zone, fqdn, tailIP string) (string, error) {
	if strings.TrimSpace(zone) == "" || strings.TrimSpace(fqdn) == "" || strings.TrimSpace(tailIP) == "" {
		return "", errors.New("zone, fqdn y tailIP son requeridos")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		return "", fmt.Errorf("docker no está instalado en PATH: %w", err)
	}

	// Preparar directorio y Corefile
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".autohost", "coredns")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	corefilePath := filepath.Join(dir, "Corefile")

	// Si no existe, crear uno base con la zona + bloque global "."
	created := false
	if _, err := os.Stat(corefilePath); os.IsNotExist(err) {
		base := `# CoreDNS (Docker) para AutoHost

{{.Zone}}:53 {
    # Limita el binding a la IP de Tailscale para evitar conflictos en :53
    bind {{.TailIP}}

    log
    errors
    hosts {
        {{.TailIP}} {{.FQDN}}
        fallthrough
    }
}

. {
    log
    errors
    forward . /etc/resolv.conf
}
`
		if err := renderToFile(base, corefilePath, map[string]string{
			"Zone": zone, "TailIP": tailIP, "FQDN": fqdn,
		}); err != nil {
			return "", fmt.Errorf("no pude escribir Corefile inicial: %w", err)
		}
		created = true
	} else {
		// Si ya existe, asegurar que tenga la zona y el mapping FQDN → tailIP
		if _, err := AddDomainToCorefileDocker(zone, fqdn, tailIP); err != nil {
			return "", fmt.Errorf("no pude actualizar Corefile existente: %w", err)
		}
	}

	// Levantar/reciclar contenedor
	exists, running, _ := containerState(coreDNSContainer)

	switch {
	case !exists:
		// run nuevo
		if err := runCoreDNSContainer(corefilePath); err != nil {
			return "", err
		}
	case exists && !running:
		// start
		if err := dockerStart(coreDNSContainer); err != nil {
			return "", err
		}
	default: // running
		// reinicia si acabamos de crear el corefile o si se actualizó (AddDomain... ya lo hace en EnsureDomainAndReload)
		if created {
			if err := dockerRestart(coreDNSContainer); err != nil {
				return "", err
			}
		}
	}

	return corefilePath, nil
}

// AddDomainToCorefileDocker garantiza que el Corefile tenga:
// - un bloque de zona para `zone` (con bind a tailIP)
// - dentro del bloque "hosts", una línea "tailIP fqdn" (idempotente; si hay otra IP para ese fqdn, la reemplaza)
// Devuelve `changed=true` si el archivo fue modificado.
func AddDomainToCorefileDocker(zone, fqdn, tailIP string) (changed bool, err error) {
	if strings.TrimSpace(zone) == "" || strings.TrimSpace(fqdn) == "" || strings.TrimSpace(tailIP) == "" {
		return false, errors.New("zone, fqdn y tailIP son requeridos")
	}

	corefilePath, err := corefilePath()
	if err != nil {
		return false, err
	}
	b, err := os.ReadFile(corefilePath)
	if err != nil {
		return false, fmt.Errorf("no pude leer Corefile en %s: %w", corefilePath, err)
	}
	content := string(b)

	// Asegurar bloque de la zona (y línea bind tailIP)
	var zoneChanged bool
	content, zoneChanged = ensureZoneBlock(content, zone, tailIP)
	changed = changed || zoneChanged

	// Insertar/actualizar la entrada "tailIP fqdn" dentro de hosts{}
	newContent, hostChanged := ensureHostMapping(content, zone, fqdn, tailIP)
	if hostChanged {
		content = newContent
		changed = true
	}

	// Escribir si cambió
	if changed {
		tmp := corefilePath + ".tmp"
		if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
			return false, err
		}
		if err := os.Rename(tmp, corefilePath); err != nil {
			return false, err
		}
	}
	return changed, nil
}

// RestartCoreDNSDocker reinicia el contenedor de CoreDNS para aplicar cambios del Corefile.
func RestartCoreDNSDocker() error {
	exists, _, _ := containerState(coreDNSContainer)
	if !exists {
		return fmt.Errorf("el contenedor %q no existe; inicia CoreDNS primero", coreDNSContainer)
	}
	return dockerRestart(coreDNSContainer)
}

// EnsureDomainAndReload agrega/actualiza el FQDN y si hubo cambios, reinicia CoreDNS.
func EnsureDomainAndReload(zone, fqdn, tailIP string) error {
	changed, err := AddDomainToCorefileDocker(zone, fqdn, tailIP)
	if err != nil {
		return err
	}
	if changed {
		if err := RestartCoreDNSDocker(); err != nil {
			return fmt.Errorf("reinicio CoreDNS falló: %w", err)
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// Helpers de contenedor Docker
// -----------------------------------------------------------------------------

func runCoreDNSContainer(corefilePath string) error {
	cmd := exec.Command(
		"docker", "run", "-d",
		"--name", coreDNSContainer,
		"--restart", "unless-stopped",
		"--network", "host",
		"-v", corefilePath+":/Corefile:ro",
		coreDNSImage, "-conf", "/Corefile",
	)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("no se pudo iniciar el contenedor CoreDNS: %w", err)
	}
	return nil
}

func dockerRestart(name string) error {
	cmd := exec.Command("docker", "restart", name)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func dockerStart(name string) error {
	cmd := exec.Command("docker", "start", name)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func containerState(name string) (exists bool, running bool, err error) {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", name)
	out, e := cmd.Output()
	if e != nil {
		// no existe o error de docker
		return false, false, nil
	}
	s := strings.TrimSpace(string(out))
	switch s {
	case "true":
		return true, true, nil
	case "false":
		return true, false, nil
	default:
		return true, false, nil
	}
}

// -----------------------------------------------------------------------------
// Helpers de Corefile (parsing/edición string-based simple)
// -----------------------------------------------------------------------------

func corefilePath() (string, error) {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".autohost", "coredns")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", fmt.Errorf("no existe %s; inicia CoreDNS con InstallAndRunCoreDNSWithDocker primero", dir)
	}
	return filepath.Join(dir, "Corefile"), nil
}

// ensureZoneBlock garantiza un bloque para la zona con bind a tailIP y hosts{}.
// Si la zona no existe, lo inserta antes del bloque '.' (o lo crea si tampoco existe '.').
func ensureZoneBlock(content, zone, tailIP string) (string, bool) {
	changed := false
	zoneHeader := zone + ":53 {"

	if strings.Contains(content, zoneHeader) {
		// Asegurar "bind tailIP" dentro del bloque de la zona
		zStart := strings.Index(content, zoneHeader)
		zEnd := findMatchingBrace(content, zStart)
		if zEnd == -1 {
			return content, false
		}
		zBlock := content[zStart:zEnd]
		if !strings.Contains(zBlock, "bind "+tailIP) {
			// Insertar/actualizar bind (si hay otro bind, lo reemplazamos a lo bruto)
			if strings.Contains(zBlock, "bind ") {
				zBlock = replaceBindLine(zBlock, tailIP)
			} else {
				// Insertar después de la apertura "{"
				openIdx := strings.Index(zBlock, "{")
				if openIdx != -1 {
					openIdx++
					zBlock = zBlock[:openIdx] + "\n    bind " + tailIP + zBlock[openIdx:]
				}
			}
			content = content[:zStart] + zBlock + content[zEnd:]
			changed = true
		}
		// Asegurar que exista hosts{}
		if !strings.Contains(zBlock, "hosts {") {
			openIdx := strings.Index(zBlock, "{")
			if openIdx != -1 {
				openIdx++
				toInsert := `
    hosts {
        # entries managed by autohost
        fallthrough
    }`
				newZ := zBlock[:openIdx] + toInsert + zBlock[openIdx:]
				content = content[:zStart] + newZ + content[zEnd:]
				changed = true
			}
		}
		return content, changed
	}

	// Si no existe la zona, construir bloque de zona con bind+hosts
	zoneBlock := fmt.Sprintf(`
%s
    bind %s
    log
    errors
    hosts {
        # entries managed by autohost
        fallthrough
    }
}
`, zoneHeader, tailIP)

	// Intentar insertar antes del bloque global "."
	dotHeader := "\n. {"
	idx := strings.Index(content, dotHeader)
	if idx == -1 && !strings.Contains(content, "\n. {\n") && !strings.Contains(content, "\n.\n{") && !strings.Contains(content, ". {\n") {
		// No hay bloque ".", lo creamos completo abajo
		dotBlock := `
. {
    log
    errors
    forward . /etc/resolv.conf
}
`
		return strings.TrimSpace(content) + "\n\n" + strings.TrimSpace(zoneBlock) + "\n\n" + strings.TrimSpace(dotBlock) + "\n", true
	}

	if idx == -1 {
		// fallback: insertar al inicio
		return strings.TrimSpace(zoneBlock) + "\n\n" + content, true
	}

	// Insertar la zona justo antes del bloque "."
	out := content[:idx] + "\n" + strings.TrimSpace(zoneBlock) + "\n" + content[idx:]
	return out, true
}

// ensureHostMapping añade/actualiza "tailIP fqdn" en el bloque hosts{} de la zona.
func ensureHostMapping(content, zone, fqdn, tailIP string) (string, bool) {
	zStart := strings.Index(content, zone+":53 {")
	if zStart == -1 {
		return content, false
	}
	zEnd := findMatchingBrace(content, zStart)
	if zEnd == -1 {
		return content, false
	}

	zBlock := content[zStart:zEnd]
	hStartRel := strings.Index(zBlock, "hosts {")
	if hStartRel == -1 {
		// insertar hosts{} después de "{"
		openIdx := strings.Index(zBlock, "{")
		if openIdx == -1 {
			return content, false
		}
		openIdx++
		toInsert := fmt.Sprintf(`
    hosts {
        %s %s
        fallthrough
    }`, tailIP, fqdn)
		newZ := zBlock[:openIdx] + toInsert + zBlock[openIdx:]
		return content[:zStart] + newZ + content[zEnd:], true
	}

	// localizar cierre de hosts{}
	hStart := zStart + hStartRel
	hEnd := findMatchingBrace(content, hStart+len("hosts "))
	if hEnd == -1 {
		return content, false
	}
	hBlock := content[hStart:hEnd]

	lines := strings.Split(hBlock, "\n")
	found := false
	for i, ln := range lines {
		trim := strings.TrimSpace(ln)
		// match "<ip> <fqdn>"
		if strings.HasSuffix(trim, " "+fqdn) {
			found = true
			if !strings.HasPrefix(trim, tailIP+" ") { // IP distinta -> reemplazar
				lines[i] = replaceLineKeepingIndent(ln, tailIP+" "+fqdn)
			}
			break
		}
	}
	if !found {
		// insertar antes de "fallthrough" si existe
		inserted := false
		for i, ln := range lines {
			if strings.Contains(ln, "fallthrough") {
				indent := leadingSpaces(ln)
				lines = append(lines[:i], append([]string{indent + tailIP + " " + fqdn}, lines[i:]...)...)
				inserted = true
				break
			}
		}
		if !inserted {
			lines = append(lines, "        "+tailIP+" "+fqdn)
		}
	}

	newH := strings.Join(lines, "\n")
	newContent := content[:hStart] + newH + content[hEnd:]
	return newContent, true
}

// findMatchingBrace encuentra la posición después del '}' que cierra el bloque
// que inicia en 'startIdxHeader'. Se busca el primer '{' después de ese header
// y se balancean llaves hasta cerrarlo.
func findMatchingBrace(s string, startIdxHeader int) int {
	open := strings.Index(s[startIdxHeader:], "{")
	if open == -1 {
		return -1
	}
	i := startIdxHeader + open + 1
	level := 1
	for i < len(s) {
		switch s[i] {
		case '{':
			level++
		case '}':
			level--
			if level == 0 {
				return i + 1 // índice después de la llave de cierre
			}
		}
		i++
	}
	return -1
}

func replaceBindLine(block, tailIP string) string {
	lines := strings.Split(block, "\n")
	for i, ln := range lines {
		trim := strings.TrimSpace(ln)
		if strings.HasPrefix(trim, "bind ") {
			lines[i] = replaceLineKeepingIndent(ln, "bind "+tailIP)
		}
	}
	return strings.Join(lines, "\n")
}

func leadingSpaces(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return s[:i]
}

func replaceLineKeepingIndent(oldLine, newCore string) string {
	return leadingSpaces(oldLine) + newCore
}

// -----------------------------------------------------------------------------
// Utils
// -----------------------------------------------------------------------------

func renderToFile(tmpl, outPath string, data any) error {
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, data)
}
