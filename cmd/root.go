package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	tokenPath    = "token-path"
	accessorPath = "accessor-path"
)

var rootCmd = &cobra.Command{
	Use:   "sidevault",
	Short: "Tools for handling Vault token management in Kubernetes.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func stringFlag(cmd *cobra.Command, name, value, usage string) {
	cmd.PersistentFlags().String(name, value, usage)
	viper.BindPFlag(name, cmd.PersistentFlags().Lookup(name))
}

func intFlag(cmd *cobra.Command, name string, value int, usage string) {
	cmd.PersistentFlags().Int(name, value, usage)
	viper.BindPFlag(name, cmd.PersistentFlags().Lookup(name))
}

func init() {
	cobra.OnInitialize(initConfig)

	stringFlag(
		rootCmd,
		tokenPath,
		"/var/run/secrets/vaultproject.io/.vault-token",
		"File system path to the Vault token.",
	)

	stringFlag(
		rootCmd,
		accessorPath,
		"/var/run/secrets/vaultproject.io/.vault-accessor",
		"File system path to the Vault token accessor.",
	)
}

func initConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}
