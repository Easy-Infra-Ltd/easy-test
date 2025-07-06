package logger

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv/v1.21.0"
)

func TestIsHandledKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"ProcessKey", ProcessKey, true},
		{"AreaKey", AreaKey, true},
		{"slog.LevelKey", slog.LevelKey, true},
		{"slog.MessageKey", slog.MessageKey, true},
		{"slog.TimeKey", slog.TimeKey, true},
		{"OTELTraceIDKey", OTELTraceIDKey, true},
		{"OTELSpanIDKey", OTELSpanIDKey, true},
		{"OTELTraceFlagsKey", OTELTraceFlagsKey, true},
		{"OTELServiceNameKey", OTELServiceNameKey, true},
		{"OTELServiceVersionKey", OTELServiceVersionKey, true},
		{"OTELResourceKey", OTELResourceKey, true},
		{"OTELScopeNameKey", OTELScopeNameKey, true},
		{"OTELScopeVersionKey", OTELScopeVersionKey, true},
		{"OTELTimestampKey", OTELTimestampKey, true},
		{"OTELSeverityTextKey", OTELSeverityTextKey, true},
		{"OTELSeverityNumberKey", OTELSeverityNumberKey, true},
		{"OTELBodyKey", OTELBodyKey, true},
		{"OTELAttributesKey", OTELAttributesKey, true},
		{"UnhandledKey", "custom_key", false},
		{"EmptyString", "", false},
		{"NilLike", "nil", false},
		{"SimilarKey", "process_key", false},
		{"CaseVariation", "PROCESS", false},
		{"SpecialChars", "process@key", false},
		{"Numeric", "123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHandledKey(tt.key)
			if result != tt.expected {
				t.Errorf("isHandledKey(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestIsHandledKeyConcurrency(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 1000

	var wg sync.WaitGroup
	results := make(chan bool, numGoroutines*numIterations)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range numIterations {
				result := isHandledKey(ProcessKey)
				results <- result
			}
		}()
	}

	wg.Wait()
	close(results)

	for result := range results {
		if !result {
			t.Error("Expected ProcessKey to be handled, got false")
		}
	}
}

func TestStringifyAttrs(t *testing.T) {
	tests := []struct {
		name     string
		attrs    map[string]any
		expected string
	}{
		{
			name:     "EmptyMap",
			attrs:    map[string]any{},
			expected: "",
		},
		{
			name: "OnlyHandledKeys",
			attrs: map[string]any{
				ProcessKey:      "test-process",
				AreaKey:         "test-area",
				slog.LevelKey:   "INFO",
				slog.MessageKey: "test message",
			},
			expected: "",
		},
		{
			name: "SingleUnhandledKey",
			attrs: map[string]any{
				"custom_key": "custom_value",
			},
			expected: "custom_key=custom_value",
		},
		{
			name: "MultipleUnhandledKeys",
			attrs: map[string]any{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			expected: "key1=value1key2=value2key3=value3",
		},
		{
			name: "MixedHandledAndUnhandled",
			attrs: map[string]any{
				ProcessKey:   "test-process",
				"custom_key": "custom_value",
				AreaKey:      "test-area",
				"another":    "value",
			},
			expected: "another=valuecustom_key=custom_value",
		},
		{
			name: "DifferentValueTypes",
			attrs: map[string]any{
				"string_val": "test",
				"int_val":    42,
				"float_val":  3.14,
				"bool_val":   true,
			},
			expected: "bool_val=truefloat_val=3.14int_val=42string_val=test",
		},
		{
			name: "ComplexValueTypes",
			attrs: map[string]any{
				"slice_val": []string{"a", "b", "c"},
				"map_val":   map[string]string{"nested": "value"},
			},
			expected: "map_val=map[nested:value]slice_val=[a b c]",
		},
		{
			name: "EmptyValues",
			attrs: map[string]any{
				"empty_string": "",
				"zero_int":     0,
				"nil_val":      nil,
			},
			expected: "empty_string=nil_val=<nil>zero_int=0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringifyAttrs(tt.attrs)
			if result != tt.expected {
				t.Errorf("stringifyAttrs() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestColouriser(t *testing.T) {
	tests := []struct {
		name       string
		colourCode int
		value      string
		expected   string
	}{
		{
			name:       "RedColour",
			colourCode: red,
			value:      "test",
			expected:   "\033[31mtest\033[0m",
		},
		{
			name:       "BlueColour",
			colourCode: blue,
			value:      "hello world",
			expected:   "\033[34mhello world\033[0m",
		},
		{
			name:       "EmptyString",
			colourCode: green,
			value:      "",
			expected:   "\033[32m\033[0m",
		},
		{
			name:       "SpecialCharacters",
			colourCode: yellow,
			value:      "test@#$%",
			expected:   "\033[33mtest@#$%\033[0m",
		},
		{
			name:       "LightGray",
			colourCode: lightGray,
			value:      "trace message",
			expected:   "\033[37mtrace message\033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Colouriser(tt.colourCode, tt.value)
			if result != tt.expected {
				t.Errorf("Colouriser(%d, %q) = %q, want %q", tt.colourCode, tt.value, result, tt.expected)
			}
		})
	}
}

func TestGetAreaColour(t *testing.T) {
	areaColours = map[string]int{}
	areaColoursIdx = 0

	tests := []struct {
		name string
		area string
	}{
		{"FirstArea", "database"},
		{"SecondArea", "api"},
		{"ThirdArea", "auth"},
		{"DuplicateArea", "database"},
	}

	colours := make(map[string]int)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colour := getAreaColour(tt.area)

			found := slices.Contains(allColours, colour)
			if !found {
				t.Errorf("getAreaColour(%q) returned invalid colour %d", tt.area, colour)
			}

			if existingColour, exists := colours[tt.area]; exists {
				if colour != existingColour {
					t.Errorf("getAreaColour(%q) returned different colour %d, expected %d", tt.area, colour, existingColour)
				}
			} else {
				colours[tt.area] = colour
			}
		})
	}
}

func TestGetAreaColourConcurrency(t *testing.T) {
	areaColours = map[string]int{}
	areaColoursIdx = 0

	const numGoroutines = 50
	const numAreas = 10

	var wg sync.WaitGroup
	results := make(chan struct {
		area   string
		colour int
	}, numGoroutines*numAreas)

	for i := range numGoroutines {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := range numAreas {
				area := fmt.Sprintf("area_%d", j)
				colour := getAreaColour(area)
				results <- struct {
					area   string
					colour int
				}{area, colour}
			}
		}(i)
	}

	wg.Wait()
	close(results)

	areaColourMap := make(map[string]int)
	for result := range results {
		if existingColour, exists := areaColourMap[result.area]; exists {
			if result.colour != existingColour {
				t.Errorf("Concurrent access to getAreaColour returned inconsistent results for area %s", result.area)
			}
		} else {
			areaColourMap[result.area] = result.colour
		}
	}
}

func TestMapSlogLevelToOTELSeverity(t *testing.T) {
	tests := []struct {
		name     string
		level    slog.Level
		expected int
	}{
		{"LevelTrace", LevelTrace, 1},
		{"LevelDebug", slog.LevelDebug, 5},
		{"LevelInfo", slog.LevelInfo, 9},
		{"LevelWarn", slog.LevelWarn, 13},
		{"LevelError", slog.LevelError, 17},
		{"LevelFatal", LevelFatal, 21},
		{"CustomLevel", slog.Level(100), 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapSlogLevelToOTELSeverity(tt.level)
			if result != tt.expected {
				t.Errorf("mapSlogLevelToOTELSeverity(%v) = %d, want %d", tt.level, result, tt.expected)
			}
		})
	}
}

func TestPrettyLine(t *testing.T) {
	noColourFunc := func(code int, value string) string { return value }

	tests := []struct {
		name        string
		data        map[string]any
		expectError bool
		contains    []string
	}{
		{
			name:        "MissingProcess",
			data:        map[string]any{},
			expectError: true,
		},
		{
			name: "MissingArea",
			data: map[string]any{
				ProcessKey: "test-process",
			},
			expectError: true,
		},
		{
			name: "MissingLevel",
			data: map[string]any{
				ProcessKey: "test-process",
				AreaKey:    "test-area",
			},
			expectError: true,
		},
		{
			name: "MissingMessage",
			data: map[string]any{
				ProcessKey:    "test-process",
				AreaKey:       "test-area",
				slog.LevelKey: "INFO",
			},
			expectError: true,
		},
		{
			name: "ValidBasicLog",
			data: map[string]any{
				ProcessKey:      "test-process",
				AreaKey:         "test-area",
				slog.LevelKey:   "INFO",
				slog.MessageKey: "test message",
			},
			expectError: false,
			contains:    []string{"test-process", "test-area", "INFO", "test message"},
		},
		{
			name: "TraceLevel",
			data: map[string]any{
				ProcessKey:      "test-process",
				AreaKey:         "test-area",
				slog.LevelKey:   "DEBUG-4",
				slog.MessageKey: "trace message",
			},
			expectError: false,
			contains:    []string{"test-process", "test-area", "TRACE", "trace message"},
		},
		{
			name: "WithOTELTraceID",
			data: map[string]any{
				ProcessKey:      "test-process",
				AreaKey:         "test-area",
				slog.LevelKey:   "INFO",
				slog.MessageKey: "test message",
				OTELTraceIDKey:  "abcd1234567890abcd1234567890abcd",
			},
			expectError: false,
			contains:    []string{"test-process", "test-area", "INFO", "test message", "abcd1234"},
		},
		{
			name: "WithServiceName",
			data: map[string]any{
				ProcessKey:         "test-process",
				AreaKey:            "test-area",
				slog.LevelKey:      "INFO",
				slog.MessageKey:    "test message",
				OTELServiceNameKey: "my-service",
			},
			expectError: false,
			contains:    []string{"test-process", "test-area", "INFO", "test message", "service=my-service"},
		},
		{
			name: "WithCustomAttributes",
			data: map[string]any{
				ProcessKey:      "test-process",
				AreaKey:         "test-area",
				slog.LevelKey:   "INFO",
				slog.MessageKey: "test message",
				"user_id":       "12345",
				"request_id":    "req-abc",
			},
			expectError: false,
			contains:    []string{"test-process", "test-area", "INFO", "test message", "request_id=req-abc", "user_id=12345"},
		},
		{
			name: "InvalidProcessType",
			data: map[string]any{
				ProcessKey:      123,
				AreaKey:         "test-area",
				slog.LevelKey:   "INFO",
				slog.MessageKey: "test message",
			},
			expectError: true,
		},
		{
			name: "InvalidAreaType",
			data: map[string]any{
				ProcessKey:      "test-process",
				AreaKey:         []string{"invalid"},
				slog.LevelKey:   "INFO",
				slog.MessageKey: "test message",
			},
			expectError: true,
		},
		{
			name: "InvalidLevelType",
			data: map[string]any{
				ProcessKey:      "test-process",
				AreaKey:         "test-area",
				slog.LevelKey:   123,
				slog.MessageKey: "test message",
			},
			expectError: true,
		},
		{
			name: "InvalidMessageType",
			data: map[string]any{
				ProcessKey:      "test-process",
				AreaKey:         "test-area",
				slog.LevelKey:   "INFO",
				slog.MessageKey: map[string]string{"invalid": "message"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PrettyLine(tt.data, noColourFunc)

			if tt.expectError {
				if err == nil {
					t.Errorf("PrettyLine() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("PrettyLine() unexpected error: %v", err)
				return
			}

			for _, expectedContent := range tt.contains {
				if !strings.Contains(result, expectedContent) {
					t.Errorf("PrettyLine() result %q does not contain expected content %q", result, expectedContent)
				}
			}
		})
	}
}

func TestHandlerOptions(t *testing.T) {
	buf := &bytes.Buffer{}

	tests := []struct {
		name    string
		options []Option
		check   func(*Handler) error
	}{
		{
			name:    "WithTimestamp",
			options: []Option{WithTimestamp()},
			check: func(h *Handler) error {
				if !h.timestamp {
					return fmt.Errorf("timestamp should be enabled")
				}
				return nil
			},
		},
		{
			name:    "WithDestinationWriter",
			options: []Option{WithDestinationWriter(buf)},
			check: func(h *Handler) error {
				if h.writer != buf {
					return fmt.Errorf("writer should be set to provided buffer")
				}
				return nil
			},
		},
		{
			name:    "WithColour",
			options: []Option{WithColour()},
			check: func(h *Handler) error {
				if !h.colourise {
					return fmt.Errorf("colourise should be enabled")
				}
				return nil
			},
		},
		{
			name:    "WithOutputEmptyAttrs",
			options: []Option{WithOutputEmptyAttrs()},
			check: func(h *Handler) error {
				if !h.outputEmptyAttrs {
					return fmt.Errorf("outputEmptyAttrs should be enabled")
				}
				return nil
			},
		},
		{
			name:    "WithOpenTelemetry",
			options: []Option{WithOpenTelemetry()},
			check: func(h *Handler) error {
				if !h.enableOTEL {
					return fmt.Errorf("enableOTEL should be enabled")
				}
				return nil
			},
		},
		{
			name:    "WithResource",
			options: []Option{WithResource(NewDefaultResource())},
			check: func(h *Handler) error {
				if h.resource == nil {
					return fmt.Errorf("resource should be set")
				}
				return nil
			},
		},
		{
			name:    "WithInstrumentationScope",
			options: []Option{WithInstrumentationScope("test-scope", "1.0.0")},
			check: func(h *Handler) error {
				if h.instrumentationScope == nil {
					return fmt.Errorf("instrumentationScope should be set")
				}
				if h.instrumentationScope.Name != "test-scope" {
					return fmt.Errorf("instrumentationScope name should be 'test-scope', got %s", h.instrumentationScope.Name)
				}
				if h.instrumentationScope.Version != "1.0.0" {
					return fmt.Errorf("instrumentationScope version should be '1.0.0', got %s", h.instrumentationScope.Version)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(nil, tt.options...)
			if err := tt.check(handler); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestNewDefaultResource(t *testing.T) {
	if err := os.Setenv("OTEL_SERVICE_NAME", "test-service"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("OTEL_SERVICE_VERSION", "2.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("DEPLOYMENT_ENVIRONMENT", "test"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Unsetenv("OTEL_SERVICE_NAME"); err != nil {
			t.Fatal(err)
		}
		if err := os.Unsetenv("OTEL_SERVICE_VERSION"); err != nil {
			t.Fatal(err)
		}
		if err := os.Unsetenv("DEPLOYMENT_ENVIRONMENT"); err != nil {
			t.Fatal(err)
		}
	}()

	resource := NewDefaultResource()
	if resource == nil {
		t.Fatal("NewDefaultResource() should not return nil")
	}

	attrs := resource.Attributes()
	attrMap := make(map[attribute.Key]attribute.Value)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
	}

	if serviceName, exists := attrMap[semconv.ServiceNameKey]; !exists || serviceName.AsString() != "test-service" {
		t.Errorf("Expected service name 'test-service', got %v", serviceName)
	}

	if serviceVersion, exists := attrMap[semconv.ServiceVersionKey]; !exists || serviceVersion.AsString() != "2.0.0" {
		t.Errorf("Expected service version '2.0.0', got %v", serviceVersion)
	}
}

func TestHandlerGetServiceInfo(t *testing.T) {
	tests := []struct {
		name                   string
		envServiceName         string
		envServiceVersion      string
		resource               *resource.Resource
		expectedServiceName    string
		expectedServiceVersion string
	}{
		{
			name:                   "FromEnvironment",
			envServiceName:         "env-service",
			envServiceVersion:      "env-version",
			resource:               nil,
			expectedServiceName:    "env-service",
			expectedServiceVersion: "env-version",
		},
		{
			name:                   "DefaultValues",
			envServiceName:         "",
			envServiceVersion:      "",
			resource:               nil,
			expectedServiceName:    "unknown_service",
			expectedServiceVersion: "unknown",
		},
		{
			name:              "FromResource",
			envServiceName:    "",
			envServiceVersion: "",
			resource: resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName("resource-service"),
				semconv.ServiceVersion("resource-version"),
			),
			expectedServiceName:    "resource-service",
			expectedServiceVersion: "resource-version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envServiceName != "" {
				if err := os.Setenv("OTEL_SERVICE_NAME", tt.envServiceName); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Unsetenv("OTEL_SERVICE_NAME"); err != nil {
					t.Fatal(err)
				}
			}
			if tt.envServiceVersion != "" {
				if err := os.Setenv("OTEL_SERVICE_VERSION", tt.envServiceVersion); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Unsetenv("OTEL_SERVICE_VERSION"); err != nil {
					t.Fatal(err)
				}
			}

			defer func() {
				if err := os.Unsetenv("OTEL_SERVICE_NAME"); err != nil {
					t.Fatal(err)
				}
				if err := os.Unsetenv("OTEL_SERVICE_VERSION"); err != nil {
					t.Fatal(err)
				}
			}()

			handler := &Handler{resource: tt.resource}
			serviceName, serviceVersion := handler.getServiceInfo()

			if serviceName != tt.expectedServiceName {
				t.Errorf("Expected service name %q, got %q", tt.expectedServiceName, serviceName)
			}
			if serviceVersion != tt.expectedServiceVersion {
				t.Errorf("Expected service version %q, got %q", tt.expectedServiceVersion, serviceVersion)
			}
		})
	}
}

func TestProcessColourFunctions(t *testing.T) {
	testColour := "lightBlue"
	setProcessColour(testColour)

	result := getProcessColour("any-process")
	expected := lightBlue

	if result != expected {
		t.Errorf("Expected process colour %d, got %d", expected, result)
	}
}

func TestCreateLoggerSink(t *testing.T) {
	if err := os.Unsetenv("DEBUG_LOG"); err != nil {
		t.Fatal(err)
	}
	sink := CreateLoggerSink()
	if sink != os.Stderr {
		t.Error("Expected os.Stderr when DEBUG_LOG is not set")
	}
}

func TestTrace(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := New(&slog.HandlerOptions{Level: LevelTrace}, WithDestinationWriter(buf))
	logger := slog.New(handler)

	logger = logger.With(ProcessKey, "test-process", AreaKey, "test-area")

	Trace(logger, "test trace message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test trace message") {
		t.Error("Trace output should contain the message")
	}
}

func BenchmarkIsHandledKey(b *testing.B) {
	keys := []string{
		ProcessKey,
		AreaKey,
		"custom_key",
		slog.LevelKey,
		"another_custom",
		OTELTraceIDKey,
		"unhandled_key",
	}

	b.ResetTimer()
	i := 0
	for b.Loop() {
		key := keys[i%len(keys)]
		_ = isHandledKey(key)
		i++
	}
}

func BenchmarkStringifyAttrs(b *testing.B) {
	attrs := map[string]any{
		"key1":     "value1",
		"key2":     42,
		"key3":     3.14,
		ProcessKey: "test-process",
		"key4":     true,
		AreaKey:    "test-area",
	}

	b.ResetTimer()
	for b.Loop() {
		_ = stringifyAttrs(attrs)
	}
}

func BenchmarkGetAreaColour(b *testing.B) {
	areas := []string{
		"database",
		"api",
		"auth",
		"cache",
		"queue",
	}

	b.ResetTimer()
	i := 0
	for b.Loop() {
		area := areas[i%len(areas)]
		_ = getAreaColour(area)
		i++
	}
}
