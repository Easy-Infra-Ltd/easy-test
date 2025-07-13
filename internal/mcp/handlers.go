package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/api"
	"github.com/Easy-Infra-Ltd/easy-test/internal/monitor"
	"github.com/Easy-Infra-Ltd/easy-test/internal/simulation"
	"github.com/google/uuid"
)

// handleRunSimulation handles the run_simulation MCP tool
func (s *MCPServer) handleRunSimulation(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req SimulationRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("failed to parse simulation request: %w", err)
	}

	// Validate request
	if req.Name == "" {
		return ErrorResponse{
			Error:   "invalid_request",
			Code:    400,
			Message: "simulation name is required",
		}, nil
	}

	if req.Attempts <= 0 {
		req.Attempts = 1
	}

	if req.Cadence <= 0 {
		req.Cadence = 1 * time.Second
	}

	// Create simulation config
	simConfig := &simulation.SimulationConfig{
		Name:     req.Name,
		Target:   req.Target,
		Cadence:  req.Cadence,
		Attempts: req.Attempts,
	}

	// Create simulation
	sim := simulation.NewSimulationFromConfig(simConfig, req.Dry)

	// Create simulation tracker
	simID := uuid.New()
	ctx, cancel := context.WithCancel(ctx)
	tracker := &SimulationTracker{
		ID:        simID,
		Name:      req.Name,
		Status:    "running",
		StartTime: time.Now().Format(time.RFC3339),
		Cancel:    cancel,
		Progress:  0,
	}

	s.addSimulationTracker(tracker)

	// Start simulation in goroutine
	go func() {
		defer func() {
			tracker.Status = "completed"
			if r := recover(); r != nil {
				tracker.Status = "failed"
				s.logger.Error("Simulation panicked", "simulation", req.Name, "error", r)
			}
		}()

		select {
		case <-ctx.Done():
			tracker.Status = "cancelled"
			return
		default:
			clients := sim.Start()
			tracker.Progress = 100
			s.logger.Info("Simulation completed", "simulation", req.Name, "clients", len(clients))
		}
	}()

	return SimulationResponse{
		ID:      simID,
		Name:    req.Name,
		Status:  "running",
		Message: fmt.Sprintf("Simulation '%s' started successfully", req.Name),
	}, nil
}

// handleListSimulations handles the list_simulations MCP tool
func (s *MCPServer) handleListSimulations(ctx context.Context, params json.RawMessage) (interface{}, error) {
	trackers := s.getAllSimulationTrackers()

	simulations := make([]SimulationInfo, 0, len(trackers))
	for _, tracker := range trackers {
		startTime, _ := time.Parse(time.RFC3339, tracker.StartTime)

		simInfo := SimulationInfo{
			ID:        tracker.ID,
			Name:      tracker.Name,
			Status:    tracker.Status,
			StartTime: startTime,
			Progress:  tracker.Progress,
		}

		if tracker.Status == "completed" || tracker.Status == "failed" {
			endTime := time.Now()
			simInfo.EndTime = &endTime
		}

		simulations = append(simulations, simInfo)
	}

	return SimulationListResponse{
		Simulations: simulations,
		Message:     fmt.Sprintf("Found %d simulations", len(simulations)),
	}, nil
}

// handleStopSimulation handles the stop_simulation MCP tool
func (s *MCPServer) handleStopSimulation(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("failed to parse stop simulation request: %w", err)
	}

	simID, err := uuid.Parse(req.ID)
	if err != nil {
		return ErrorResponse{
			Error:   "invalid_request",
			Code:    400,
			Message: "invalid simulation ID format",
		}, nil
	}

	tracker, exists := s.getSimulationTracker(simID)
	if !exists {
		return ErrorResponse{
			Error:   "simulation_not_found",
			Code:    404,
			Message: fmt.Sprintf("simulation with ID %s not found", req.ID),
		}, nil
	}

	if tracker.Status != "running" {
		return ErrorResponse{
			Error:   "simulation_not_running",
			Code:    400,
			Message: fmt.Sprintf("simulation %s is not running (status: %s)", req.ID, tracker.Status),
		}, nil
	}

	if tracker.Cancel != nil {
		tracker.Cancel()
	}
	tracker.Status = "cancelled"

	return SimulationResponse{
		ID:      simID,
		Name:    tracker.Name,
		Status:  "cancelled",
		Message: fmt.Sprintf("Simulation '%s' stopped successfully", tracker.Name),
	}, nil
}

// handleGetSimulationResults handles the get_simulation_results MCP tool
func (s *MCPServer) handleGetSimulationResults(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("failed to parse get simulation results request: %w", err)
	}

	simID, err := uuid.Parse(req.ID)
	if err != nil {
		return ErrorResponse{
			Error:   "invalid_request",
			Code:    400,
			Message: "invalid simulation ID format",
		}, nil
	}

	tracker, exists := s.getSimulationTracker(simID)
	if !exists {
		return ErrorResponse{
			Error:   "simulation_not_found",
			Code:    404,
			Message: fmt.Sprintf("simulation with ID %s not found", req.ID),
		}, nil
	}

	// For now, return basic simulation info
	// In a real implementation, this would gather actual test results
	results := []string{
		fmt.Sprintf("Simulation '%s' status: %s", tracker.Name, tracker.Status),
		fmt.Sprintf("Progress: %d%%", tracker.Progress),
		fmt.Sprintf("Started: %s", tracker.StartTime),
	}

	return SimulationResponse{
		ID:      simID,
		Name:    tracker.Name,
		Status:  tracker.Status,
		Results: results,
		Message: fmt.Sprintf("Retrieved results for simulation '%s'", tracker.Name),
	}, nil
}

// handleValidateConfig handles the validate_config MCP tool
func (s *MCPServer) handleValidateConfig(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req ConfigValidationRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("failed to parse config validation request: %w", err)
	}

	errors := []string{}
	valid := true

	switch req.Type {
	case "simulation":
		var simConfig simulation.SimulationConfig
		if err := json.Unmarshal(req.Config, &simConfig); err != nil {
			errors = append(errors, fmt.Sprintf("invalid JSON: %v", err))
			valid = false
		} else {
			// Basic validation
			if simConfig.Name == "" {
				errors = append(errors, "simulation name is required")
				valid = false
			}
			if simConfig.Attempts <= 0 {
				errors = append(errors, "attempts must be greater than 0")
				valid = false
			}
			if simConfig.Target.Count <= 0 {
				errors = append(errors, "target count must be greater than 0")
				valid = false
			}
			if simConfig.Target.Client.Url == "" {
				errors = append(errors, "target client URL is required")
				valid = false
			}
		}
	default:
		errors = append(errors, fmt.Sprintf("unsupported config type: %s", req.Type))
		valid = false
	}

	message := "Configuration is valid"
	if !valid {
		message = "Configuration validation failed"
	}

	return ConfigValidationResponse{
		Valid:   valid,
		Errors:  errors,
		Message: message,
	}, nil
}

// handleGenerateConfigTemplate handles the generate_config_template MCP tool
func (s *MCPServer) handleGenerateConfigTemplate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req ConfigTemplateRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("failed to parse config template request: %w", err)
	}

	var template interface{}
	var message string

	switch req.Type {
	case "simulation":
		template = simulation.SimulationConfig{
			Name:     "example-simulation",
			Cadence:  5 * time.Second,
			Attempts: 10,
			Target: simulation.SimulationTargetConfig{
				Count: 3,
				Client: &api.ClientConfig{
					Url:         "https://example.com/api",
					ContentType: "application/json",
				},
				Monitor: &monitor.MonitorConfig{
					Name: "example-monitor",
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
		message = "Generated simulation configuration template"
	default:
		return ErrorResponse{
			Error:   "unsupported_type",
			Code:    400,
			Message: fmt.Sprintf("unsupported config type: %s", req.Type),
		}, nil
	}

	templateJSON, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template: %w", err)
	}

	return ConfigTemplateResponse{
		Template: json.RawMessage(templateJSON),
		Message:  message,
	}, nil
}

// Placeholder handlers for remaining tools
func (s *MCPServer) handleQueryLogs(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return LogQueryResponse{
		Logs:    []LogEntry{},
		Total:   0,
		Message: "Log querying not yet implemented",
	}, nil
}

func (s *MCPServer) handleGetLogSummary(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return LogQueryResponse{
		Logs:    []LogEntry{},
		Total:   0,
		Message: "Log summary not yet implemented",
	}, nil
}

func (s *MCPServer) handleGetMonitorStatus(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return MonitorStatusResponse{
		Monitors: []MonitorInfo{},
		Message:  "Monitor status not yet implemented",
	}, nil
}

func (s *MCPServer) handleSetupMonitor(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return MonitorStatusResponse{
		Monitors: []MonitorInfo{},
		Message:  "Monitor setup not yet implemented",
	}, nil
}

func (s *MCPServer) handleAnalyzePerformance(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"message": "Performance analysis not yet implemented",
	}, nil
}
