package utils

import (
	"os"
	"path/filepath"
)

func GetAutohostDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("No se pudo obtener el directorio HOME")
	}
	return filepath.Join(home, ".autohost")
}

func EnsureAutohostDirs() error {
	base := GetAutohostDir()
	subdirs := []string{"docker/compose", "cloudflare", "logs", "templates", "state"}

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
