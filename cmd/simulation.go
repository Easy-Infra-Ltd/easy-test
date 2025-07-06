package cmd

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/simulation"
)

type SimulationOptions struct {
	Config  *simulation.SimulationConfig
	DryRun  bool
	Workers int
	Timeout time.Duration
}

func PrepareSimulationOptions(config *simulation.SimulationConfig, dryRun bool, workers int, timeout time.Duration) (SimulationOptions, error) {
	if config == nil {
		return SimulationOptions{}, fmt.Errorf("configuration cannot be nil")
	}

	if workers <= 0 {
		return SimulationOptions{}, fmt.Errorf("workers must be greater than 0, got %d", workers)
	}

	if timeout <= 0 {
		return SimulationOptions{}, fmt.Errorf("timeout must be greater than 0, got %v", timeout)
	}

	return SimulationOptions{
		Config:  config,
		DryRun:  dryRun,
		Workers: workers,
		Timeout: timeout,
	}, nil
}

func ExecuteSimulation(opts SimulationOptions, logger *slog.Logger) error {
	if logger == nil {
		return fmt.Errorf("logger cannot be nil")
	}

	if GetVerbose() {
		logger.Info("Starting simulation execution",
			"dryRun", opts.DryRun,
			"workers", opts.Workers,
			"timeout", opts.Timeout)
	}

	sim := simulation.NewSimulationFromConfig(opts.Config, opts.DryRun)
	sim.Start()

	logger.Info("Simulation completed successfully")
	return nil
}
