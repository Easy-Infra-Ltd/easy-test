package cmd

import (
	"log/slog"
	"testing"

	"github.com/Easy-Infra-Ltd/easy-test/internal/logger"
	"github.com/Easy-Infra-Ltd/easy-test/internal/simulation"
)

func TestGetVerbose(t *testing.T) {
	tests := []struct {
		name     string
		setValue bool
		expected bool
	}{
		{
			name:     "VerboseTrue",
			setValue: true,
			expected: true,
		},
		{
			name:     "VerboseFalse",
			setValue: false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verbose = tt.setValue

			result := GetVerbose()
			if result != tt.expected {
				t.Errorf("GetVerbose() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		setValue string
		expected string
	}{
		{
			name:     "LogLevelInfo",
			setValue: "info",
			expected: "info",
		},
		{
			name:     "LogLevelDebug",
			setValue: "debug",
			expected: "debug",
		},
		{
			name:     "LogLevelTrace",
			setValue: "trace",
			expected: "trace",
		},
		{
			name:     "LogLevelWarn",
			setValue: "warn",
			expected: "warn",
		},
		{
			name:     "LogLevelError",
			setValue: "error",
			expected: "error",
		},
		{
			name:     "EmptyString",
			setValue: "",
			expected: "",
		},
		{
			name:     "InvalidLevel",
			setValue: "invalid",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logLevel = tt.setValue

			result := GetLogLevel()
			if result != tt.expected {
				t.Errorf("GetLogLevel() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetNoColor(t *testing.T) {
	tests := []struct {
		name     string
		setValue bool
		expected bool
	}{
		{
			name:     "NoColorTrue",
			setValue: true,
			expected: true,
		},
		{
			name:     "NoColorFalse",
			setValue: false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noColor = tt.setValue

			result := GetNoColor()
			if result != tt.expected {
				t.Errorf("GetNoColor() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetConfigFile(t *testing.T) {
	tests := []struct {
		name     string
		setValue string
		expected string
	}{
		{
			name:     "ConfigFileSet",
			setValue: "/path/to/config.yaml",
			expected: "/path/to/config.yaml",
		},
		{
			name:     "EmptyConfigFile",
			setValue: "",
			expected: "",
		},
		{
			name:     "RelativePath",
			setValue: "./config.json",
			expected: "./config.json",
		},
		{
			name:     "HomeDirectory",
			setValue: "~/.easy-test.yaml",
			expected: "~/.easy-test.yaml",
		},
		{
			name:     "SpecialCharacters",
			setValue: "/path/with spaces/config-file.yaml",
			expected: "/path/with spaces/config-file.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgFile = tt.setValue

			result := GetConfigFile()
			if result != tt.expected {
				t.Errorf("GetConfigFile() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{
			name:     "TraceLevel",
			input:    "trace",
			expected: logger.LevelTrace,
		},
		{
			name:     "DebugLevel",
			input:    "debug",
			expected: slog.LevelDebug,
		},
		{
			name:     "InfoLevel",
			input:    "info",
			expected: slog.LevelInfo,
		},
		{
			name:     "WarnLevel",
			input:    "warn",
			expected: slog.LevelWarn,
		},
		{
			name:     "ErrorLevel",
			input:    "error",
			expected: slog.LevelError,
		},
		{
			name:     "DefaultToInfo",
			input:    "invalid",
			expected: slog.LevelInfo,
		},
		{
			name:     "EmptyStringDefaultToInfo",
			input:    "",
			expected: slog.LevelInfo,
		},
		{
			name:     "CaseSensitive",
			input:    "INFO",
			expected: slog.LevelInfo, // Should default to info since it's case sensitive
		},
		{
			name:     "NumericString",
			input:    "123",
			expected: slog.LevelInfo,
		},
		{
			name:     "SpecialCharacters",
			input:    "debug!@#",
			expected: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateSimulationConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *simulation.SimulationConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "NilConfig",
			config:      nil,
			expectError: true,
			errorMsg:    "configuration cannot be nil",
		},
		{
			name:        "ValidConfig",
			config:      &simulation.SimulationConfig{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSimulationConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateSimulationConfig() expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("validateSimulationConfig() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateSimulationConfig() unexpected error: %v", err)
				}
			}
		})
	}
}

func BenchmarkGetVerbose(b *testing.B) {
	verbose = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetVerbose()
	}
}

func BenchmarkGetLogLevel(b *testing.B) {
	logLevel = "debug"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetLogLevel()
	}
}

func BenchmarkParseLogLevel(b *testing.B) {
	levels := []string{"trace", "debug", "info", "warn", "error", "invalid"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		level := levels[i%len(levels)]
		_ = parseLogLevel(level)
	}
}

func TestConcurrentAccess(t *testing.T) {
	done := make(chan bool, 100)
	verbose = true
	logLevel = "debug"
	noColor = false
	cfgFile = "/test/config.yaml"

	for i := 0; i < 100; i++ {
		go func() {
			_ = GetVerbose()
			_ = GetLogLevel()
			_ = GetNoColor()
			_ = GetConfigFile()
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	if GetVerbose() != true {
		t.Error("Concurrent access changed verbose value")
	}
	if GetLogLevel() != "debug" {
		t.Error("Concurrent access changed logLevel value")
	}
	if GetNoColor() != false {
		t.Error("Concurrent access changed noColor value")
	}
	if GetConfigFile() != "/test/config.yaml" {
		t.Error("Concurrent access changed cfgFile value")
	}
}
