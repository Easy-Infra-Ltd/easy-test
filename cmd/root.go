package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	verbose  bool
	logLevel string
	noColor  bool
)

var rootCmd = &cobra.Command{
	Use:   "easy-test",
	Short: "Easy testing framework for services and integrations",
	Long: `Easy-Test provides simulation capabilities with monitoring, 
logging, and OpenTelemetry support for testing services and integrations.

This tool allows you to run complex test simulations, validate configurations,
and monitor your services with detailed logging and observability features.`,
	Run: func(cmd *cobra.Command, args []string) {
		runCmd.Run(cmd, args)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.easy-test.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"verbose output")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info",
		"log level (trace|debug|info|warn|error)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false,
		"disable colored output")

	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}

func GetVerbose() bool {
	return verbose
}

func GetLogLevel() string {
	return logLevel
}

func GetNoColor() bool {
	return noColor
}

func GetConfigFile() string {
	return cfgFile
}
