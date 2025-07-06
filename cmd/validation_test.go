package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestValidateFilePath(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_file")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	tempDir, err := os.MkdirTemp("", "test_dir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		filePath    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "EmptyPath",
			filePath:    "",
			expectError: true,
			errorMsg:    "file path cannot be empty",
		},
		{
			name:        "NonExistentFile",
			filePath:    "nonexistent_file.txt",
			expectError: true,
			errorMsg:    "does not exist",
		},
		{
			name:        "ValidFile",
			filePath:    tempFile.Name(),
			expectError: false,
		},
		{
			name:        "DirectoryPath",
			filePath:    tempDir,
			expectError: true,
			errorMsg:    "is a directory, not a file",
		},
		{
			name:        "PathWithSpaces",
			filePath:    "file with spaces.txt",
			expectError: true,
			errorMsg:    "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateFilePath() expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateFilePath() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateFilePath() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestFormatValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		errors   []ConfigValidationError
		expected string
		isNil    bool
	}{
		{
			name:   "NoErrors",
			errors: []ConfigValidationError{},
			isNil:  true,
		},
		{
			name: "SingleError",
			errors: []ConfigValidationError{
				{Field: "testField", Message: "test message"},
			},
			expected: "validation error in field 'testField': test message",
		},
		{
			name: "MultipleErrors",
			errors: []ConfigValidationError{
				{Field: "field1", Message: "message1"},
				{Field: "field2", Message: "message2"},
			},
			expected: "found 2 validation errors:\n  1. validation error in field 'field1': message1\n  2. validation error in field 'field2': message2",
		},
		{
			name: "ThreeErrors",
			errors: []ConfigValidationError{
				{Field: "fieldA", Message: "messageA"},
				{Field: "fieldB", Message: "messageB"},
				{Field: "fieldC", Message: "messageC"},
			},
			expected: "found 3 validation errors:\n  1. validation error in field 'fieldA': messageA\n  2. validation error in field 'fieldB': messageB\n  3. validation error in field 'fieldC': messageC",
		},
		{
			name: "ErrorsWithSpecialCharacters",
			errors: []ConfigValidationError{
				{Field: "field@#$", Message: "message with symbols !@#"},
				{Field: "field with spaces", Message: "another message"},
			},
			expected: "found 2 validation errors:\n  1. validation error in field 'field@#$': message with symbols !@#\n  2. validation error in field 'field with spaces': another message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatValidationErrors(tt.errors)

			if tt.isNil {
				if result != nil {
					t.Errorf("FormatValidationErrors() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("FormatValidationErrors() = nil, want error")
				} else if result.Error() != tt.expected {
					t.Errorf("FormatValidationErrors() = %v, want %v", result.Error(), tt.expected)
				}
			}
		})
	}
}

func TestValidatePositiveInt(t *testing.T) {
	tests := []struct {
		name        string
		value       int
		fieldName   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "PositiveValue",
			value:       10,
			fieldName:   "workers",
			expectError: false,
		},
		{
			name:        "ZeroValue",
			value:       0,
			fieldName:   "workers",
			expectError: true,
			errorMsg:    "workers must be greater than 0, got 0",
		},
		{
			name:        "NegativeValue",
			value:       -5,
			fieldName:   "workers",
			expectError: true,
			errorMsg:    "workers must be greater than 0, got -5",
		},
		{
			name:        "MaxIntValue",
			value:       2147483647,
			fieldName:   "maxValue",
			expectError: false,
		},
		{
			name:        "MinNegativeValue",
			value:       -2147483648,
			fieldName:   "minValue",
			expectError: true,
			errorMsg:    "minValue must be greater than 0, got -2147483648",
		},
		{
			name:        "OneValue",
			value:       1,
			fieldName:   "count",
			expectError: false,
		},
		{
			name:        "EmptyFieldName",
			value:       0,
			fieldName:   "",
			expectError: true,
			errorMsg:    " must be greater than 0, got 0",
		},
		{
			name:        "FieldNameWithSpaces",
			value:       -1,
			fieldName:   "field with spaces",
			expectError: true,
			errorMsg:    "field with spaces must be greater than 0, got -1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveInt(tt.value, tt.fieldName)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidatePositiveInt() expected error but got none")
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("ValidatePositiveInt() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePositiveInt() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateNonEmptyString(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		fieldName   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "NonEmptyString",
			value:       "test value",
			fieldName:   "name",
			expectError: false,
		},
		{
			name:        "EmptyString",
			value:       "",
			fieldName:   "name",
			expectError: true,
			errorMsg:    "name cannot be empty",
		},
		{
			name:        "WhitespaceString",
			value:       "   ",
			fieldName:   "description",
			expectError: false, // Whitespace is considered non-empty
		},
		{
			name:        "SingleCharacter",
			value:       "a",
			fieldName:   "initial",
			expectError: false,
		},
		{
			name:        "SpecialCharacters",
			value:       "@#$%^&*()",
			fieldName:   "symbols",
			expectError: false,
		},
		{
			name:        "UnicodeString",
			value:       "こんにちは",
			fieldName:   "greeting",
			expectError: false,
		},
		{
			name:        "EmptyFieldName",
			value:       "",
			fieldName:   "",
			expectError: true,
			errorMsg:    " cannot be empty",
		},
		{
			name:        "LongString",
			value:       strings.Repeat("a", 1000),
			fieldName:   "longField",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNonEmptyString(tt.value, tt.fieldName)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateNonEmptyString() expected error but got none")
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("ValidateNonEmptyString() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateNonEmptyString() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidationEdgeCases(t *testing.T) {
	t.Run("ValidateFilePathPermissions", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test_perms")
		if err != nil {
			t.Skip("Cannot create temp file for permission test")
		}
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		err = ValidateFilePath(tempFile.Name())
		if err != nil {
			t.Logf("File validation with permissions: %v", err)
		}
	})

	t.Run("ValidateFilePathSymlink", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test_target")
		if err != nil {
			t.Skip("Cannot create temp file for symlink test")
		}
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		symlinkPath := tempFile.Name() + "_link"
		err = os.Symlink(tempFile.Name(), symlinkPath)
		if err != nil {
			t.Skip("Cannot create symlink for test")
		}
		defer os.Remove(symlinkPath)

		err = ValidateFilePath(symlinkPath)
		if err != nil {
			t.Errorf("ValidateFilePath() failed for valid symlink: %v", err)
		}
	})
}

func BenchmarkValidateFilePath(b *testing.B) {
	tempFile, err := os.CreateTemp("", "bench_file")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateFilePath(tempFile.Name())
	}
}

func BenchmarkValidatePositiveInt(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidatePositiveInt(10, "workers")
	}
}

func BenchmarkValidateNonEmptyString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateNonEmptyString("test value", "name")
	}
}

func BenchmarkFormatValidationErrors(b *testing.B) {
	errors := []ConfigValidationError{
		{Field: "field1", Message: "message1"},
		{Field: "field2", Message: "message2"},
		{Field: "field3", Message: "message3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatValidationErrors(errors)
	}
}