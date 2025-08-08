package utils

import (
	"os"
	"path/filepath"
)

func GetAutohostDir() string {
	if custom := os.Getenv("AUTOHOST_DIR"); custom != "" {
		return custom
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "./.autohost"
	}
	return filepath.Join(home, ".autohost")
}

func GetSubdir(subdir string) string {
	return filepath.Join(GetAutohostDir(), subdir)
}

func IsInitialized() bool {
	_, err := os.Stat(GetAutohostDir())
	return !os.IsNotExist(err)
}
