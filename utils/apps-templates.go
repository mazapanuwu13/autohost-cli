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
  caddy:
    image: caddy 
    container_name: caddy
    ports:
      - "80:80"
    volumes:
      - /root/.autohost/caddy/Caddyfile:/etc/caddy/Caddyfile
    networks:
      - autohost_net

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
    networks:
      - autohost_net

  app:
    image: nextcloud
    container_name: nextcloud_app
    volumes:
      - nextcloud:/var/www/html
    restart: always
    environment:
      MYSQL_PASSWORD: nc_pass
      MYSQL_DATABASE: nextcloud
      MYSQL_USER: nc_user
      MYSQL_HOST: db
    networks:
      - autohost_net

volumes:
  db:
  nextcloud:

networks:
  autohost_net:
    external: true

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
