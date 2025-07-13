package cmd

import (
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/Easy-Infra-Ltd/easy-test/internal/logger"
)

var (
	simulationPath string
	dryRun         bool
	workers        int
	timeout        time.Duration
)

var runCmd = &cobra.Command{
	Use:   "run [simulation-file]",
	Short: "Execute a simulation",
	Long: `Run executes a simulation configuration file with optional dry-run mode.

The simulation file should contain a JSON configuration that defines the
test scenarios, endpoints, and parameters for your integration tests.

Examples:
  easy-test run
  easy-test run simulation.json
  easy-test run --dry
  easy-test run --path custom.json --workers 20
  easy-test run --timeout 5m custom-simulation.json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSimulation,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&simulationPath, "path", "p", "simulation.json",
		"path to simulation configuration file")
	runCmd.Flags().BoolVar(&dryRun, "dry", false,
		"dry run without making external requests")
	runCmd.Flags().IntVarP(&workers, "workers", "w", 10,
		"number of concurrent workers")
	runCmd.Flags().DurationVarP(&timeout, "timeout", "t", 30*time.Second,
		"simulation timeout")
}

func runSimulation(cmd *cobra.Command, args []string) error {
	configPath := ResolveConfigPath(args, simulationPath)

	config, err := ParseConfigFile(configPath)
	if err != nil {
		return err
	}

	validationErrors := ValidateConfigData(config)
	if len(validationErrors) > 0 {
		return FormatValidationErrors(validationErrors)
	}

	loggerOpts := CreateLoggerOptions(GetVerbose(), GetNoColor(), GetLogLevel(), "lightGreen")
	baseLogger := CreateLogger(loggerOpts)
	log := ConfigureProcessLogger(baseLogger, "Simulation", "CLI", dryRun)

	simOpts, err := PrepareSimulationOptions(config, dryRun, workers, timeout)
	if err != nil {
		return err
	}

	return ExecuteSimulation(simOpts, log)
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "trace":
		return logger.LevelTrace
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
