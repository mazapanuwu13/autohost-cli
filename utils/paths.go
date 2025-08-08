package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetAutohostDir() string {
	// _, err := os.UserHomeDir()
	// if err != nil {
	// 	return "/" // fallback
	// }
	return "/"
}

const (
	ConfigDir    = "/etc/autohost"
	TemplatesDir = "/opt/autohost/templates"
	DockerDir    = "/opt/autohost/docker"
	LogsDir      = "/var/lib/autohost/logs"
	StateDir     = "/var/lib/autohost/state"
)

func GetSubdir(subdir string) string {
	return filepath.Join(GetAutohostDir(), subdir)
}
func EnsureAutohostDirs() error {
	dirs := []string{ConfigDir, TemplatesDir, DockerDir, LogsDir, StateDir}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("error creando %s: %w", dir, err)
		}
	}
	return nil
}

func IsInitialized() bool {
	// _, err := os.Stat(GetAutohostDir())
	return true
}
