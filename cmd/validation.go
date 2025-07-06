package cmd

import (
	"fmt"
	"os"
)

func ValidateFilePath(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file %q does not exist", filePath)
	}
	if err != nil {
		return fmt.Errorf("cannot access file %q: %w", filePath, err)
	}

	if info.IsDir() {
		return fmt.Errorf("%q is a directory, not a file", filePath)
	}

	return nil
}

func FormatValidationErrors(errors []ConfigValidationError) error {
	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		return errors[0]
	}

	message := fmt.Sprintf("found %d validation errors:", len(errors))
	for i, err := range errors {
		message += fmt.Sprintf("\n  %d. %s", i+1, err.Error())
	}

	return fmt.Errorf("%s", message)
}

func ValidatePositiveInt(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be greater than 0, got %d", fieldName, value)
	}
	return nil
}

func ValidateNonEmptyString(value, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	return nil
}