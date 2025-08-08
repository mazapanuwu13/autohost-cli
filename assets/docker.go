package assets

import (
	"embed"
	"io/fs"
	"path/filepath"
)

//go:embed docker/**/*
var dockerFS embed.FS

// ReadCompose lee assets/docker/<app>/docker-compose.yml
func ReadCompose(app string) ([]byte, error) {
	return fs.ReadFile(dockerFS, filepath.Join("docker", app, "docker-compose.yml"))
}

func ReadEnvExample(app string) ([]byte, error) {
	return fs.ReadFile(dockerFS, filepath.Join("docker", app, ".env.example"))
}

// ListApps devuelve todas las apps que tienen plantilla
func ListApps() ([]string, error) {
	entries, err := fs.ReadDir(dockerFS, "docker")
	if err != nil {
		return nil, err
	}
	var apps []string
	for _, e := range entries {
		if e.IsDir() {
			apps = append(apps, e.Name())
		}
	}
	return apps, nil
}
