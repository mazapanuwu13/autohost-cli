package utils

import "strings"

// ReplacePlaceholders reemplaza claves tipo {{KEY}} por valores del map.
func ReplacePlaceholders(content string, values map[string]string) string {
	out := content
	for k, v := range values {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}
