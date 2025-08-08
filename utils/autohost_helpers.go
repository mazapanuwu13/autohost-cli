package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Exec(cmdName string, args ...string) error {
	cmd := exec.Command(cmdName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ExecWithDir(dir string, cmdName string, args ...string) error {
	cmd := exec.Command(cmdName, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ExecShell(command string) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("❌ Error ejecutando comando:", err)
		os.Exit(1)
	}
}

func Confirm(prompt string) bool {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}

func AskOption(prompt string, options []string) string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(prompt)
		for i, opt := range options {
			fmt.Printf("[%d] %s\n", i+1, opt)
		}
		fmt.Print("Elige una opción: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if i, err := strconv.Atoi(input); err == nil && i >= 1 && i <= len(options) {
			return options[i-1]
		}
		fmt.Println("❌ Opción inválida, intenta de nuevo.")
	}
}
