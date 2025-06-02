package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func InstallNextcloud() error {
	compose := `version: '3.8'

services:
  db:
    image: mariadb
    container_name: nextcloud_db
    restart: always
    volumes:
      - db:/var/lib/mysql
    environment:
      MYSQL_ROOT_PASSWORD: example
      MYSQL_DATABASE: nextcloud
      MYSQL_USER: nc_user
      MYSQL_PASSWORD: nc_pass

  app:
    image: nextcloud
    container_name: nextcloud_app
    ports:
      - "8080:80"
    volumes:
      - nextcloud:/var/www/html
    restart: always
    environment:
      MYSQL_PASSWORD: nc_pass
      MYSQL_DATABASE: nextcloud
      MYSQL_USER: nc_user
      MYSQL_HOST: db

volumes:
  db:
  nextcloud:
`

	path := filepath.Join(GetAutohostDir(), "docker", "compose", "nextcloud.yml")
	return os.WriteFile(path, []byte(compose), 0644)
}

func StartApp(app string) error {
	ymlPath := filepath.Join(GetAutohostDir(), "docker", "compose", app+".yml")

	cmd := exec.Command("docker", "compose", "-f", ymlPath, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("🔄 Levantando aplicación con Docker...")
	return cmd.Run()
}
