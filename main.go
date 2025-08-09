/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"autohost-cli/cmd"
)

func main() {
	// err := utils.LoadURLsConfig("config/urls.toml")
	// if err != nil {
	// 	log.Fatalf("❌ No se pudo cargar config de URLs: %v", err)
	// }
	cmd.Execute()
}
