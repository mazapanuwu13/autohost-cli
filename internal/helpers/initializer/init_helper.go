package initializer

import (
	"autohost-cli/utils"
	"os"
)

func EnsureAutohostDirs() error {
	subdirs := []string{
		"config",
		"templates",
		"apps",
		"logs",
		"state",
		"backups",
	}

	for _, sub := range subdirs {
		if err := os.MkdirAll(utils.GetSubdir(sub), 0755); err != nil {
			return err
		}
	}
	return nil
}
