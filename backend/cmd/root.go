package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

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
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is config/config.toml)")
	rootCmd.PersistentFlags().String("host", "0.0.0.0", "server host")
	rootCmd.PersistentFlags().Int("port", 8080, "server port")
	rootCmd.PersistentFlags().String("storage-type", "json", "storage type (json or mongodb)")
	rootCmd.PersistentFlags().String("json-file", "data.json", "JSON storage file path")
	rootCmd.PersistentFlags().String("mongo-uri", "", "MongoDB connection URI")
	rootCmd.PersistentFlags().String("issuer", "", "OpenID issuer URL")

	// Bind flags to viper
	_ = viper.BindPFlag("server.host", rootCmd.PersistentFlags().Lookup("host"))
	_ = viper.BindPFlag("server.port", rootCmd.PersistentFlags().Lookup("port"))
	_ = viper.BindPFlag("storage.type", rootCmd.PersistentFlags().Lookup("storage-type"))
	_ = viper.BindPFlag("storage.json_file_path", rootCmd.PersistentFlags().Lookup("json-file"))
	_ = viper.BindPFlag("storage.mongo_uri", rootCmd.PersistentFlags().Lookup("mongo-uri"))
	_ = viper.BindPFlag("issuer", rootCmd.PersistentFlags().Lookup("issuer"))
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in config directory first, then current directory
		viper.AddConfigPath("config")
		viper.AddConfigPath(".")
		viper.SetConfigType("toml")
		viper.SetConfigName("config")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("OPENID")
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
