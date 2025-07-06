package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Easy-Infra-Ltd/easy-test/internal/simulation"
)

func TestParseConfigFile(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		fileContent string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "EmptyFilePath",
			filePath:    "",
			expectError: true,
			errorMsg:    "file path cannot be empty",
		},
		{
			name:        "NonExistentFile",
			filePath:    "nonexistent.json",
			expectError: true,
		},
		{
			name:        "ValidJSONFile",
			filePath:    "valid.json",
			fileContent: `{}`,
			expectError: false,
		},
		{
			name:        "InvalidJSONFile",
			filePath:    "invalid.json",
			fileContent: `{invalid json`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fileContent != "" {
				file, err := os.CreateTemp("", tt.filePath)
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer os.Remove(file.Name())
				
				if _, err := file.WriteString(tt.fileContent); err != nil {
					t.Fatalf("Failed to write to temp file: %v", err)
				}
				file.Close()
				
				tt.filePath = file.Name()
			}

			result, err := ParseConfigFile(tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseConfigFile() expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ParseConfigFile() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ParseConfigFile() unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("ParseConfigFile() returned nil config")
				}
			}
		})
	}
}

func TestParseConfigReader(t *testing.T) {
	tests := []struct {
		name        string
		reader      *strings.Reader
		strict      bool
		expectError bool
		errorMsg    string
		nilReader   bool
	}{
		{
			name:        "NilReader",
			nilReader:   true,
			expectError: true,
			errorMsg:    "reader cannot be nil",
		},
		{
			name:        "ValidJSON",
			reader:      strings.NewReader(`{}`),
			strict:      false,
			expectError: false,
		},
		{
			name:        "InvalidJSON",
			reader:      strings.NewReader(`{invalid`),
			strict:      false,
			expectError: true,
		},
		{
			name:        "StrictModeWithUnknownFields",
			reader:      strings.NewReader(`{"unknownField": "value"}`),
			strict:      true,
			expectError: true,
		},
		{
			name:        "NonStrictModeWithUnknownFields",
			reader:      strings.NewReader(`{"unknownField": "value"}`),
			strict:      false,
			expectError: false,
		},
		{
			name:        "EmptyJSON",
			reader:      strings.NewReader(`{}`),
			strict:      true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reader io.Reader
			if tt.nilReader {
				reader = nil
			} else {
				reader = tt.reader
			}
			result, err := ParseConfigReader(reader, tt.strict)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseConfigReader() expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ParseConfigReader() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ParseConfigReader() unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("ParseConfigReader() returned nil config")
				}
			}
		})
	}
}

func TestValidateConfigData(t *testing.T) {
	tests := []struct {
		name           string
		config         *simulation.SimulationConfig
		expectedErrors int
	}{
		{
			name:           "NilConfig",
			config:         nil,
			expectedErrors: 1,
		},
		{
			name:           "ValidConfig",
			config:         &simulation.SimulationConfig{},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateConfigData(tt.config)
			
			if len(errors) != tt.expectedErrors {
				t.Errorf("ValidateConfigData() returned %d errors, want %d", len(errors), tt.expectedErrors)
			}

			if tt.config == nil && len(errors) > 0 {
				if errors[0].Field != "config" {
					t.Errorf("ValidateConfigData() first error field = %v, want 'config'", errors[0].Field)
				}
			}
		})
	}
}

func TestResolveConfigPath(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flagPath string
		expected string
	}{
		{
			name:     "NoArgsUsesFlag",
			args:     []string{},
			flagPath: "flag.json",
			expected: "flag.json",
		},
		{
			name:     "EmptyArgsUsesFlag",
			args:     []string{""},
			flagPath: "flag.json",
			expected: "flag.json",
		},
		{
			name:     "ArgOverridesFlag",
			args:     []string{"arg.json"},
			flagPath: "flag.json",
			expected: "arg.json",
		},
		{
			name:     "BothEmpty",
			args:     []string{},
			flagPath: "",
			expected: "",
		},
		{
			name:     "MultipleArgsUsesFirst",
			args:     []string{"first.json", "second.json"},
			flagPath: "flag.json",
			expected: "first.json",
		},
		{
			name:     "ArgumentWithSpaces",
			args:     []string{"path with spaces.json"},
			flagPath: "flag.json",
			expected: "path with spaces.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveConfigPath(tt.args, tt.flagPath)
			if result != tt.expected {
				t.Errorf("ResolveConfigPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfigValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		message  string
		expected string
	}{
		{
			name:     "BasicError",
			field:    "testField",
			message:  "test message",
			expected: "validation error in field 'testField': test message",
		},
		{
			name:     "EmptyField",
			field:    "",
			message:  "test message",
			expected: "validation error in field '': test message",
		},
		{
			name:     "EmptyMessage",
			field:    "testField",
			message:  "",
			expected: "validation error in field 'testField': ",
		},
		{
			name:     "SpecialCharacters",
			field:    "field@#$",
			message:  "message with symbols !@#",
			expected: "validation error in field 'field@#$': message with symbols !@#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ConfigValidationError{
				Field:   tt.field,
				Message: tt.message,
			}
			
			result := err.Error()
			if result != tt.expected {
				t.Errorf("ConfigValidationError.Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func BenchmarkParseConfigReader(b *testing.B) {
	reader := strings.NewReader(`{}`)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.Seek(0, 0)
		_, _ = ParseConfigReader(reader, false)
	}
}

func BenchmarkResolveConfigPath(b *testing.B) {
	args := []string{"test.json"}
	flagPath := "flag.json"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ResolveConfigPath(args, flagPath)
	}
}