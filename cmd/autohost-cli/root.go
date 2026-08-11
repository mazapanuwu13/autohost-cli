package cli

import (
	"autohost-cli/cmd/autohost-cli/agent"
	"autohost-cli/cmd/autohost-cli/app"
	"autohost-cli/cmd/autohost-cli/cc"
	"autohost-cli/cmd/autohost-cli/enroll"
	"autohost-cli/cmd/autohost-cli/install"
	"autohost-cli/cmd/autohost-cli/setup"
	"autohost-cli/cmd/autohost-cli/up"
	"autohost-cli/cmd/autohost-cli/vpn"
	"autohost-cli/internal/adapters/catalog"
	"autohost-cli/internal/adapters/docker"
	"autohost-cli/internal/adapters/installed"
	"autohost-cli/internal/adapters/tailscale"
	appSvc "autohost-cli/internal/app"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Inyectado en build time por goreleaser
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

const banner = `
    █████╗ ██╗   ██╗████████╗██████╗ ██╗  ██╗██████╗ ███████╗████████╗       ██████╗ ██╗     ██╗
    ██╔══██╗██║   ██║╚══██╔══╝██╔══██╗██║  ██║██╔══██╗██╔════╝╚══██╔══╝      ██╔════╝ ██║     ██║
    ███████║██║   ██║   ██║   ██║  ██║███████║██║  ██║███████╗   ██║   █████╗██║      ██║     ██║
    ██╔══██║██║   ██║   ██║   ██║  ██║██╔══██║██║  ██║╚════██║   ██║   ╚════╝██║      ██║     ██║
    ██║  ██║╚██████╔╝   ██║   ██████╔╝██║  ██║██████╔╝███████║   ██║         ╚██████╗ ███████╗██║
    ╚═╝  ╚═╝ ╚═════╝    ╚═╝   ╚═════╝ ╚═╝  ╚═╝╚═════╝ ╚══════╝   ╚═╝          ╚═════╝ ╚══════╝╚═╝
`

// ANSI color codes
const (
	colorCyan  = "\033[36m"
	colorGray  = "\033[90m"
	colorWhite = "\033[97m"
	colorReset = "\033[0m"
)

func printBanner() {
	fmt.Print(colorCyan + banner + colorReset)
	fmt.Printf("    %sv%s%s  %s• %s • %s\n", colorWhite, Version, colorReset, colorGray, Date, colorReset)
	fmt.Printf("    %sCLI para autohosting con Docker/Tailscale/Cloudflare/Caddy%s\n\n", colorGray, colorReset)
	fmt.Printf("    Usa %sautohost --help%s para ver los comandos disponibles.\n\n", colorWhite, colorReset)
}

var rootCmd = &cobra.Command{
	Use:     "autohost",
	Short:   "CLI para autohosting con Docker/Tailscale/Cloudflare/Caddy",
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		printBanner()
		cmd.Usage()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Composition root: all services are constructed and injected here.
	dockerAdapter := docker.New()

	appService := &appSvc.AppService{
		Docker:    dockerAdapter,
		Catalog:   catalog.New(),
		Installed: installed.New(),
	}

	rootCmd.AddCommand(agent.AgentCmd(&appSvc.AgentService{}))
	rootCmd.AddCommand(enroll.EnrollCmd(&appSvc.EnrollService{}))
	rootCmd.AddCommand(up.UpCmd(&appSvc.UpService{}))
	rootCmd.AddCommand(cc.CCCmd(&appSvc.CCService{}))
	rootCmd.AddCommand(app.AppCmd(appService))
	rootCmd.AddCommand(install.InstallCmd(appService))
	rootCmd.AddCommand(setup.SetupCmd(&appSvc.SetupService{
		Docker:    dockerAdapter,
		Tailscale: tailscale.New(),
	}))
	rootCmd.AddCommand(vpn.VPNCmd(&appSvc.VPNService{
		Tailscale: tailscale.New(),
	}))
}
