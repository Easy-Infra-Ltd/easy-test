package cmd

import (
	"log/slog"
	"testing"

	"github.com/Easy-Infra-Ltd/easy-test/internal/logger"
)

func TestCreateLoggerOptions(t *testing.T) {
	tests := []struct {
		name      string
		verbose   bool
		noColor   bool
		logLevel  string
		colorName string
		expected  LoggerOptions
	}{
		{
			name:      "DefaultOptions",
			verbose:   false,
			noColor:   false,
			logLevel:  "info",
			colorName: "blue",
			expected: LoggerOptions{
				Verbose:   false,
				NoColor:   false,
				Level:     slog.LevelInfo,
				ColorName: "blue",
			},
		},
		{
			name:      "VerboseEnabled",
			verbose:   true,
			noColor:   false,
			logLevel:  "debug",
			colorName: "green",
			expected: LoggerOptions{
				Verbose:   true,
				NoColor:   false,
				Level:     slog.LevelDebug,
				ColorName: "green",
			},
		},
		{
			name:      "NoColorEnabled",
			verbose:   false,
			noColor:   true,
			logLevel:  "error",
			colorName: "red",
			expected: LoggerOptions{
				Verbose:   false,
				NoColor:   true,
				Level:     slog.LevelError,
				ColorName: "red",
			},
		},
		{
			name:      "TraceLevel",
			verbose:   true,
			noColor:   false,
			logLevel:  "trace",
			colorName: "lightBlue",
			expected: LoggerOptions{
				Verbose:   true,
				NoColor:   false,
				Level:     logger.LevelTrace,
				ColorName: "lightBlue",
			},
		},
		{
			name:      "InvalidLevelDefaultsToInfo",
			verbose:   false,
			noColor:   false,
			logLevel:  "invalid",
			colorName: "yellow",
			expected: LoggerOptions{
				Verbose:   false,
				NoColor:   false,
				Level:     slog.LevelInfo,
				ColorName: "yellow",
			},
		},
		{
			name:      "EmptyColorName",
			verbose:   false,
			noColor:   false,
			logLevel:  "warn",
			colorName: "",
			expected: LoggerOptions{
				Verbose:   false,
				NoColor:   false,
				Level:     slog.LevelWarn,
				ColorName: "",
			},
		},
		{
			name:      "AllFlagsEnabled",
			verbose:   true,
			noColor:   true,
			logLevel:  "debug",
			colorName: "magenta",
			expected: LoggerOptions{
				Verbose:   true,
				NoColor:   true,
				Level:     slog.LevelDebug,
				ColorName: "magenta",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateLoggerOptions(tt.verbose, tt.noColor, tt.logLevel, tt.colorName)
			
			if result.Verbose != tt.expected.Verbose {
				t.Errorf("CreateLoggerOptions().Verbose = %v, want %v", result.Verbose, tt.expected.Verbose)
			}
			if result.NoColor != tt.expected.NoColor {
				t.Errorf("CreateLoggerOptions().NoColor = %v, want %v", result.NoColor, tt.expected.NoColor)
			}
			if result.Level != tt.expected.Level {
				t.Errorf("CreateLoggerOptions().Level = %v, want %v", result.Level, tt.expected.Level)
			}
			if result.ColorName != tt.expected.ColorName {
				t.Errorf("CreateLoggerOptions().ColorName = %v, want %v", result.ColorName, tt.expected.ColorName)
			}
		})
	}
}

func TestCreateLogger(t *testing.T) {
	tests := []struct {
		name    string
		opts    LoggerOptions
		wantNil bool
	}{
		{
			name: "ColoredLogger",
			opts: LoggerOptions{
				Verbose:   false,
				NoColor:   false,
				Level:     slog.LevelInfo,
				ColorName: "blue",
			},
			wantNil: false,
		},
		{
			name: "JSONLogger",
			opts: LoggerOptions{
				Verbose:   false,
				NoColor:   true,
				Level:     slog.LevelDebug,
				ColorName: "red",
			},
			wantNil: false,
		},
		{
			name: "TraceLevel",
			opts: LoggerOptions{
				Verbose:   true,
				NoColor:   false,
				Level:     logger.LevelTrace,
				ColorName: "green",
			},
			wantNil: false,
		},
		{
			name: "EmptyColorName",
			opts: LoggerOptions{
				Verbose:   false,
				NoColor:   true,
				Level:     slog.LevelInfo,
				ColorName: "",
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateLogger(tt.opts)
			
			if tt.wantNil && result != nil {
				t.Errorf("CreateLogger() = %v, want nil", result)
			}
			if !tt.wantNil && result == nil {
				t.Errorf("CreateLogger() = nil, want non-nil logger")
			}
		})
	}
}

func TestConfigureProcessLogger(t *testing.T) {
	tests := []struct {
		name        string
		processName string
		area        string
		dryRun      bool
		expectDryPrefix bool
	}{
		{
			name:            "NormalProcess",
			processName:     "TestProcess",
			area:           "TestArea",
			dryRun:         false,
			expectDryPrefix: false,
		},
		{
			name:            "DryRunProcess",
			processName:     "TestProcess",
			area:           "TestArea",
			dryRun:         true,
			expectDryPrefix: true,
		},
		{
			name:            "EmptyProcessName",
			processName:     "",
			area:           "TestArea",
			dryRun:         false,
			expectDryPrefix: false,
		},
		{
			name:            "EmptyArea",
			processName:     "TestProcess",
			area:           "",
			dryRun:         false,
			expectDryPrefix: false,
		},
		{
			name:            "SpecialCharactersInNames",
			processName:     "Test@Process#123",
			area:           "Test$Area%456",
			dryRun:         false,
			expectDryPrefix: false,
		},
		{
			name:            "DryRunWithSpecialCharacters",
			processName:     "Test Process",
			area:           "Test Area",
			dryRun:         true,
			expectDryPrefix: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseLogger := slog.Default()
			
			result := ConfigureProcessLogger(baseLogger, tt.processName, tt.area, tt.dryRun)
			
			if result == nil {
				t.Errorf("ConfigureProcessLogger() returned nil")
			}
			
		})
	}
}

func TestLoggerOptionsEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		expected  slog.Level
	}{
		{
			name:     "CaseSensitiveLevel",
			logLevel: "INFO",
			expected: slog.LevelInfo,
		},
		{
			name:     "NumericLevel",
			logLevel: "123",
			expected: slog.LevelInfo,
		},
		{
			name:     "SpecialCharactersLevel",
			logLevel: "debug!@#",
			expected: slog.LevelInfo,
		},
		{
			name:     "EmptyLevel",
			logLevel: "",
			expected: slog.LevelInfo,
		},
		{
			name:     "WhitespaceLevel",
			logLevel: "  debug  ",
			expected: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := CreateLoggerOptions(false, false, tt.logLevel, "blue")
			if opts.Level != tt.expected {
				t.Errorf("CreateLoggerOptions() with level %q = %v, want %v", tt.logLevel, opts.Level, tt.expected)
			}
		})
	}
}

func BenchmarkCreateLoggerOptions(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CreateLoggerOptions(true, false, "debug", "blue")
	}
}

func BenchmarkConfigureProcessLogger(b *testing.B) {
	baseLogger := slog.Default()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConfigureProcessLogger(baseLogger, "TestProcess", "TestArea", false)
	}
}