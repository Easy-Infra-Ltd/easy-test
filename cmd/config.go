package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Easy-Infra-Ltd/easy-test/internal/simulation"
)

type ConfigValidationError struct {
	Field   string
	Message string
}

func (e ConfigValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

func ParseConfigFile(filePath string) (*simulation.SimulationConfig, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open configuration file %q: %w", filePath, err)
	}
	defer file.Close()

	return ParseConfigReader(file, false)
}

func ParseConfigReader(reader io.Reader, strict bool) (*simulation.SimulationConfig, error) {
	if reader == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}

	var config simulation.SimulationConfig
	decoder := json.NewDecoder(reader)
	
	if strict {
		decoder.DisallowUnknownFields()
	}
	
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON configuration: %w", err)
	}

	return &config, nil
}

func ValidateConfigData(config *simulation.SimulationConfig) []ConfigValidationError {
	var errors []ConfigValidationError

	if config == nil {
		errors = append(errors, ConfigValidationError{
			Field:   "config",
			Message: "configuration cannot be nil",
		})
		return errors
	}

	return errors
}

func ResolveConfigPath(args []string, flagPath string) string {
	if len(args) > 0 && args[0] != "" {
		return args[0]
	}
	return flagPath
}