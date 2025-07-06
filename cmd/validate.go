package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Easy-Infra-Ltd/easy-test/internal/simulation"
)

var validateCmd = &cobra.Command{
	Use:   "validate <simulation-file>",
	Short: "Validate simulation configuration",
	Long: `Validate checks the syntax and structure of a simulation configuration file
without executing the simulation.

This command will:
- Parse the JSON configuration file
- Validate the structure matches expected schema
- Check for required fields and valid values
- Report any errors or warnings

Examples:
  easy-test validate simulation.json
  easy-test validate --verbose config/test-simulation.json`,
	Args: cobra.ExactArgs(1),
	RunE: validateConfig,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func validateConfig(cmd *cobra.Command, args []string) error {
	configPath := args[0]

	if err := ValidateFilePath(configPath); err != nil {
		return err
	}

	loggerOpts := CreateLoggerOptions(GetVerbose(), GetNoColor(), GetLogLevel(), "blue")
	baseLogger := CreateLogger(loggerOpts)
	log := ConfigureProcessLogger(baseLogger, "Validation", "CLI", false)

	if GetVerbose() {
		log.Info("Starting validation", "file", configPath)
	}

	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open configuration file: %w", err)
	}
	defer file.Close()

	config, err := ParseConfigReader(file, true) // strict mode
	if err != nil {
		return err
	}

	validationErrors := ValidateConfigData(config)
	if len(validationErrors) > 0 {
		return FormatValidationErrors(validationErrors)
	}

	fmt.Printf("✓ Configuration file %q is valid\n", configPath)
	
	if GetVerbose() {
		log.Info("Configuration validated successfully",
			"file", configPath,
			"config", fmt.Sprintf("%+v", config))
	}

	return nil
}

func validateSimulationConfig(config *simulation.SimulationConfig) error {
	
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	

	return nil
}