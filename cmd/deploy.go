package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [template]",
	Short: "Despliega un servicio basado en una plantilla",
	Long: `Este comando despliega un servicio basado en una plantilla predefinida.
Se copia la plantilla en un directorio de servicios y se levanta con docker-compose.
Ejemplo: autohost deploy ghost-blog`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		fmt.Println("🚀 Desplegando servicio:", templateName)

		// Directorio de origen de la plantilla (se espera que exista en ./templates/<template>)
		srcDir := filepath.Join("templates", templateName)

		// Directorio de destino: ~/.autohost/services/<template>
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("❌ No se pudo obtener el directorio home:", err)
			return
		}
		targetDir := filepath.Join(home, ".autohost", "services", templateName)

		// Verificar que la plantilla exista
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			fmt.Println("❌ La plantilla no existe:", srcDir)
			return
		}

		// Copiar la plantilla al directorio de servicios
		err = copyDir(srcDir, targetDir)
		if err != nil {
			fmt.Println("❌ Error al copiar la plantilla:", err)
			return
		}
		fmt.Println("✅ Plantilla copiada a:", targetDir)

		// Ejecutar "docker-compose up -d" en el directorio de destino
		fmt.Println("🔄 Iniciando docker-compose...")
		cmdCompose := exec.Command("docker", "compose", "up", "-d")
		cmdCompose.Dir = targetDir
		cmdCompose.Stdout = os.Stdout
		cmdCompose.Stderr = os.Stderr
		err = cmdCompose.Run()
		if err != nil {
			fmt.Println("❌ Error al ejecutar docker compose:", err)
			return
		}
		fmt.Println("🚀 Servicio desplegado exitosamente en:", targetDir)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

// copyDir copia recursivamente un directorio de src a dst.
func copyDir(src string, dst string) error {
	// Crear el directorio destino si no existe
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile copia un archivo de src a dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Sync()
}
