package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/api"
	"github.com/Easy-Infra-Ltd/easy-test/internal/monitor"
	"github.com/Easy-Infra-Ltd/easy-test/internal/simulation"
)

func TestNewMCPServer(t *testing.T) {
	server := NewMCPServer()

	if server == nil {
		t.Fatal("NewMCPServer() returned nil")
	}

	if server.logger == nil {
		t.Error("MCP server logger is nil")
	}

	if server.tools == nil {
		t.Error("MCP server tools map is nil")
	}

	if server.simulations == nil {
		t.Error("MCP server simulations map is nil")
	}
}

func TestGetAvailableTools(t *testing.T) {
	server := NewMCPServer()
	tools := server.GetAvailableTools()

	expectedTools := []string{
		"run_simulation",
		"list_simulations",
		"stop_simulation",
		"get_simulation_results",
		"validate_config",
		"generate_config_template",
		"query_logs",
		"get_log_summary",
		"get_monitor_status",
		"setup_monitor",
		"analyze_performance",
	}

	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}

	toolsMap := make(map[string]bool)
	for _, tool := range tools {
		toolsMap[tool] = true
	}

	for _, expectedTool := range expectedTools {
		if !toolsMap[expectedTool] {
			t.Errorf("Expected tool %s not found in available tools", expectedTool)
		}
	}
}

func TestHandleToolCallUnknownTool(t *testing.T) {
	server := NewMCPServer()
	ctx := context.Background()

	result, err := server.HandleToolCall(ctx, "unknown_tool", json.RawMessage(`{}`))

	if err != nil {
		t.Errorf("HandleToolCall should not return error for unknown tool, got: %v", err)
	}

	errorResp, ok := result.(ErrorResponse)
	if !ok {
		t.Errorf("Expected ErrorResponse, got %T", result)
	}

	if errorResp.Error != "tool_not_found" {
		t.Errorf("Expected error 'tool_not_found', got %s", errorResp.Error)
	}
}

func TestHandleValidateConfig(t *testing.T) {
	server := NewMCPServer()
	ctx := context.Background()

	// Test valid simulation config
	validConfig := simulation.SimulationConfig{
		Name:     "test-simulation",
		Cadence:  5 * time.Second,
		Attempts: 3,
		Target: simulation.SimulationTargetConfig{
			Count: 2,
			Client: &api.ClientConfig{
				Url:         "https://example.com/api",
				ContentType: "application/json",
			},
			Monitor: &monitor.MonitorConfig{
				Name: "test-monitor",
				MonitorTargets: []*monitor.MonitorTargetConfig{
					{
						Client: &api.ClientConfig{
							Url:         "https://example.com/health",
							ContentType: "application/json",
						},
						Freq:             5 * time.Second,
						Retries:          3,
						ExpectedResponse: map[string]any{"status": "ok"},
					},
				},
			},
		},
	}

	configJSON, _ := json.Marshal(validConfig)

	requestParams := ConfigValidationRequest{
		Config: json.RawMessage(configJSON),
		Type:   "simulation",
	}

	paramsJSON, _ := json.Marshal(requestParams)

	result, err := server.HandleToolCall(ctx, "validate_config", json.RawMessage(paramsJSON))

	if err != nil {
		t.Errorf("HandleToolCall should not return error for valid config, got: %v", err)
	}

	validationResp, ok := result.(ConfigValidationResponse)
	if !ok {
		t.Errorf("Expected ConfigValidationResponse, got %T", result)
	}

	if !validationResp.Valid {
		t.Errorf("Expected valid config, got invalid with errors: %v", validationResp.Errors)
	}
}

func TestHandleValidateConfigInvalid(t *testing.T) {
	server := NewMCPServer()
	ctx := context.Background()

	// Test invalid simulation config (missing name)
	invalidConfig := simulation.SimulationConfig{
		Name:     "", // Invalid: empty name
		Cadence:  5 * time.Second,
		Attempts: 0, // Invalid: zero attempts
		Target: simulation.SimulationTargetConfig{
			Count: 0, // Invalid: zero count
			Client: &api.ClientConfig{
				Url:         "", // Invalid: empty URL
				ContentType: "application/json",
			},
		},
	}

	configJSON, _ := json.Marshal(invalidConfig)

	requestParams := ConfigValidationRequest{
		Config: json.RawMessage(configJSON),
		Type:   "simulation",
	}

	paramsJSON, _ := json.Marshal(requestParams)

	result, err := server.HandleToolCall(ctx, "validate_config", json.RawMessage(paramsJSON))

	if err != nil {
		t.Errorf("HandleToolCall should not return error for invalid config, got: %v", err)
	}

	validationResp, ok := result.(ConfigValidationResponse)
	if !ok {
		t.Errorf("Expected ConfigValidationResponse, got %T", result)
	}

	if validationResp.Valid {
		t.Error("Expected invalid config, but got valid")
	}

	if len(validationResp.Errors) == 0 {
		t.Error("Expected validation errors, but got none")
	}
}

func TestHandleGenerateConfigTemplate(t *testing.T) {
	server := NewMCPServer()
	ctx := context.Background()

	requestParams := ConfigTemplateRequest{
		Type: "simulation",
	}

	paramsJSON, _ := json.Marshal(requestParams)

	result, err := server.HandleToolCall(ctx, "generate_config_template", json.RawMessage(paramsJSON))

	if err != nil {
		t.Errorf("HandleToolCall should not return error for template generation, got: %v", err)
	}

	templateResp, ok := result.(ConfigTemplateResponse)
	if !ok {
		t.Errorf("Expected ConfigTemplateResponse, got %T", result)
	}

	if len(templateResp.Template) == 0 {
		t.Error("Expected template JSON, but got empty")
	}

	// Verify template is valid JSON
	var template simulation.SimulationConfig
	if err := json.Unmarshal(templateResp.Template, &template); err != nil {
		t.Errorf("Template is not valid JSON: %v", err)
	}

	if template.Name != "example-simulation" {
		t.Errorf("Expected template name 'example-simulation', got %s", template.Name)
	}
}

func TestHandleListSimulations(t *testing.T) {
	server := NewMCPServer()
	ctx := context.Background()

	result, err := server.HandleToolCall(ctx, "list_simulations", json.RawMessage(`{}`))

	if err != nil {
		t.Errorf("HandleToolCall should not return error for list simulations, got: %v", err)
	}

	listResp, ok := result.(SimulationListResponse)
	if !ok {
		t.Errorf("Expected SimulationListResponse, got %T", result)
	}

	if len(listResp.Simulations) != 0 {
		t.Error("Expected empty simulations list for new server")
	}
}

func TestSimulationTracker(t *testing.T) {
	server := NewMCPServer()

	// Test adding simulation tracker
	tracker := &SimulationTracker{
		ID:     [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Name:   "test-simulation",
		Status: "running",
	}

	server.addSimulationTracker(tracker)

	// Test getting simulation tracker
	retrieved, exists := server.getSimulationTracker(tracker.ID)
	if !exists {
		t.Error("Expected simulation tracker to exist")
	}

	if retrieved.Name != tracker.Name {
		t.Errorf("Expected name %s, got %s", tracker.Name, retrieved.Name)
	}

	// Test getting all simulation trackers
	all := server.getAllSimulationTrackers()
	if len(all) != 1 {
		t.Errorf("Expected 1 simulation tracker, got %d", len(all))
	}

	// Test removing simulation tracker
	server.removeSimulationTracker(tracker.ID)

	_, exists = server.getSimulationTracker(tracker.ID)
	if exists {
		t.Error("Expected simulation tracker to be removed")
	}
}
