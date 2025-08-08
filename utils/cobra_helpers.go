package utils

import "github.com/spf13/cobra"

func WithAppName(fn func(appName string)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		appName := args[0]
		fn(appName)
	}
}
