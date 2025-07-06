package cmd

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/simulation"
)

func TestPrepareSimulationOptions(t *testing.T) {
	tests := []struct {
		name        string
		config      *simulation.SimulationConfig
		dryRun      bool
		workers     int
		timeout     time.Duration
		expectError bool
		errorMsg    string
	}{
		{
			name:        "ValidOptions",
			config:      &simulation.SimulationConfig{},
			dryRun:      false,
			workers:     10,
			timeout:     30 * time.Second,
			expectError: false,
		},
		{
			name:        "NilConfig",
			config:      nil,
			dryRun:      false,
			workers:     10,
			timeout:     30 * time.Second,
			expectError: true,
			errorMsg:    "configuration cannot be nil",
		},
		{
			name:        "ZeroWorkers",
			config:      &simulation.SimulationConfig{},
			dryRun:      false,
			workers:     0,
			timeout:     30 * time.Second,
			expectError: true,
			errorMsg:    "workers must be greater than 0",
		},
		{
			name:        "NegativeWorkers",
			config:      &simulation.SimulationConfig{},
			dryRun:      false,
			workers:     -5,
			timeout:     30 * time.Second,
			expectError: true,
			errorMsg:    "workers must be greater than 0",
		},
		{
			name:        "ZeroTimeout",
			config:      &simulation.SimulationConfig{},
			dryRun:      false,
			workers:     10,
			timeout:     0,
			expectError: true,
			errorMsg:    "timeout must be greater than 0",
		},
		{
			name:        "NegativeTimeout",
			config:      &simulation.SimulationConfig{},
			dryRun:      false,
			workers:     10,
			timeout:     -time.Second,
			expectError: true,
			errorMsg:    "timeout must be greater than 0",
		},
		{
			name:        "DryRunMode",
			config:      &simulation.SimulationConfig{},
			dryRun:      true,
			workers:     5,
			timeout:     15 * time.Second,
			expectError: false,
		},
		{
			name:        "MaxWorkers",
			config:      &simulation.SimulationConfig{},
			dryRun:      false,
			workers:     1000,
			timeout:     5 * time.Minute,
			expectError: false,
		},
		{
			name:        "MinValidWorkers",
			config:      &simulation.SimulationConfig{},
			dryRun:      false,
			workers:     1,
			timeout:     1 * time.Nanosecond,
			expectError: false,
		},
		{
			name:        "LongTimeout",
			config:      &simulation.SimulationConfig{},
			dryRun:      false,
			workers:     10,
			timeout:     24 * time.Hour,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PrepareSimulationOptions(tt.config, tt.dryRun, tt.workers, tt.timeout)

			if tt.expectError {
				if err == nil {
					t.Errorf("PrepareSimulationOptions() expected error but got none")
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("PrepareSimulationOptions() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("PrepareSimulationOptions() unexpected error: %v", err)
				}
				
				if result.Config != tt.config {
					t.Errorf("PrepareSimulationOptions().Config = %v, want %v", result.Config, tt.config)
				}
				if result.DryRun != tt.dryRun {
					t.Errorf("PrepareSimulationOptions().DryRun = %v, want %v", result.DryRun, tt.dryRun)
				}
				if result.Workers != tt.workers {
					t.Errorf("PrepareSimulationOptions().Workers = %v, want %v", result.Workers, tt.workers)
				}
				if result.Timeout != tt.timeout {
					t.Errorf("PrepareSimulationOptions().Timeout = %v, want %v", result.Timeout, tt.timeout)
				}
			}
		})
	}
}

func TestExecuteSimulation(t *testing.T) {
	tests := []struct {
		name        string
		opts        SimulationOptions
		logger      *slog.Logger
		expectError bool
		errorMsg    string
	}{
		{
			name: "NilLogger",
			opts: SimulationOptions{
				Config:  &simulation.SimulationConfig{},
				DryRun:  false,
				Workers: 10,
				Timeout: 30 * time.Second,
			},
			logger:      nil,
			expectError: true,
			errorMsg:    "logger cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExecuteSimulation(tt.opts, tt.logger)

			if tt.expectError {
				if err == nil {
					t.Errorf("ExecuteSimulation() expected error but got none")
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("ExecuteSimulation() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ExecuteSimulation() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSimulationOptionsStruct(t *testing.T) {
	tests := []struct {
		name    string
		config  *simulation.SimulationConfig
		dryRun  bool
		workers int
		timeout time.Duration
	}{
		{
			name:    "DefaultValues",
			config:  &simulation.SimulationConfig{},
			dryRun:  false,
			workers: 1,
			timeout: 1 * time.Second,
		},
		{
			name:    "CustomValues",
			config:  &simulation.SimulationConfig{},
			dryRun:  true,
			workers: 100,
			timeout: 5 * time.Minute,
		},
		{
			name:    "NilConfig",
			config:  nil,
			dryRun:  false,
			workers: 10,
			timeout: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := SimulationOptions{
				Config:  tt.config,
				DryRun:  tt.dryRun,
				Workers: tt.workers,
				Timeout: tt.timeout,
			}

			if opts.Config != tt.config {
				t.Errorf("SimulationOptions.Config = %v, want %v", opts.Config, tt.config)
			}
			if opts.DryRun != tt.dryRun {
				t.Errorf("SimulationOptions.DryRun = %v, want %v", opts.DryRun, tt.dryRun)
			}
			if opts.Workers != tt.workers {
				t.Errorf("SimulationOptions.Workers = %v, want %v", opts.Workers, tt.workers)
			}
			if opts.Timeout != tt.timeout {
				t.Errorf("SimulationOptions.Timeout = %v, want %v", opts.Timeout, tt.timeout)
			}
		})
	}
}

func TestSimulationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		workers     int
		timeout     time.Duration
		expectError bool
	}{
		{
			name:        "MaxInt32Workers",
			workers:     2147483647,
			timeout:     time.Second,
			expectError: false,
		},
		{
			name:        "MinTimeoutNanosecond",
			workers:     1,
			timeout:     1 * time.Nanosecond,
			expectError: false,
		},
		{
			name:        "MaxDurationTimeout",
			workers:     1,
			timeout:     9223372036854775807,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &simulation.SimulationConfig{}
			_, err := PrepareSimulationOptions(config, false, tt.workers, tt.timeout)

			if tt.expectError && err == nil {
				t.Errorf("PrepareSimulationOptions() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("PrepareSimulationOptions() unexpected error: %v", err)
			}
		})
	}
}

func containsString(haystack, needle string) bool {
	return len(needle) == 0 || len(haystack) >= len(needle) && 
		   (haystack == needle || 
		    haystack[:len(needle)] == needle || 
		    haystack[len(haystack)-len(needle):] == needle ||
		    strings.Contains(haystack, needle))
}

func BenchmarkPrepareSimulationOptions(b *testing.B) {
	config := &simulation.SimulationConfig{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = PrepareSimulationOptions(config, false, 10, 30*time.Second)
	}
}

func BenchmarkPrepareSimulationOptionsOnly(b *testing.B) {
	config := &simulation.SimulationConfig{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = PrepareSimulationOptions(config, true, 10, 30*time.Second)
	}
}