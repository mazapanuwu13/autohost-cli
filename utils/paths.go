package utils

import (
	"os"
	"path/filepath"
)

func GetAutohostDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback muy simple, no recomendado para producción
		return "./autohost"
	}
	return filepath.Join(home, "go", "src", "github.com", "mazapanuwu13", "autohost-cli")
}

func EnsureAutohostDirs() error {
	base := GetAutohostDir()

	subdirs := []string{"etc/autohost", "/opt/autohost/docker", "/opt/autohost/templates", "/var/lib/autohost/logs", "/var/lib/autohost/state"}

	for _, sub := range subdirs {
		if err := os.MkdirAll(filepath.Join(base, sub), 0755); err != nil {
			return err
		}
	}
	return nil
}

func IsInitialized() bool {
	_, err := os.Stat(GetAutohostDir())
	return !os.IsNotExist(err)
}
