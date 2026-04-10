package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// adminUIFS and publicFS hold the embedded assets injected from package main.
var adminUIFS embed.FS
var publicFS embed.FS

// SetEmbeds receives the embedded filesystems from package main.
// Must be called before Execute().
func SetEmbeds(uiFS embed.FS, pubFS embed.FS) {
	adminUIFS = uiFS
	publicFS = pubFS
	// Fail fast with a clear message if the frontend was not built.
	if _, err := adminUIFS.Open("frontend/dist/index.html"); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: frontend/dist/index.html not found in embed.")
		fmt.Fprintln(os.Stderr, "Run 'make build-frontend' before starting the server.")
		os.Exit(1)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "openid-server",
	Short: "OpenID Connect Server",
	Long: `A lightweight OpenID Connect (OIDC) server implementation in Go.
Supports OAuth 2.0 authorization code flow with PKCE and implicit flow.`,
	// Default action is to serve
	Run: func(cmd *cobra.Command, args []string) {
		serveCmd.Run(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().String("host", "0.0.0.0", "server host")
	rootCmd.PersistentFlags().Int("port", 8080, "server port")
	rootCmd.PersistentFlags().String("storage-type", "json", "storage type (json or mongodb)")
	rootCmd.PersistentFlags().String("json-file", "data/openid.json", "JSON storage file path")
	rootCmd.PersistentFlags().String("mongo-uri", "", "MongoDB connection URI")
	rootCmd.PersistentFlags().String("issuer", "", "OpenID issuer URL")
}
