package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Easy-Infra-Ltd/easy-test/internal/logger"
	"github.com/google/uuid"
)

// MCPServer represents the MCP server instance
type MCPServer struct {
	logger      *slog.Logger
	tools       map[string]ToolHandler
	simulations map[uuid.UUID]*SimulationTracker
	mu          sync.RWMutex
}

// ToolHandler represents a function that handles MCP tool calls
type ToolHandler func(ctx context.Context, params json.RawMessage) (interface{}, error)

// SimulationTracker tracks running simulations
type SimulationTracker struct {
	ID        uuid.UUID
	Name      string
	Status    string
	StartTime string
	Cancel    context.CancelFunc
	Progress  int
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer() *MCPServer {
	mcpLogger := logger.CreateLoggerFromEnv(nil, "cyan")
	mcpLogger = mcpLogger.With("area", "MCP-Server")

	server := &MCPServer{
		logger:      mcpLogger,
		tools:       make(map[string]ToolHandler),
		simulations: make(map[uuid.UUID]*SimulationTracker),
	}

	server.registerTools()
	return server
}

// registerTools registers all available MCP tools
func (s *MCPServer) registerTools() {
	s.tools["run_simulation"] = s.handleRunSimulation
	s.tools["list_simulations"] = s.handleListSimulations
	s.tools["stop_simulation"] = s.handleStopSimulation
	s.tools["get_simulation_results"] = s.handleGetSimulationResults
	s.tools["validate_config"] = s.handleValidateConfig
	s.tools["generate_config_template"] = s.handleGenerateConfigTemplate
	s.tools["query_logs"] = s.handleQueryLogs
	s.tools["get_log_summary"] = s.handleGetLogSummary
	s.tools["get_monitor_status"] = s.handleGetMonitorStatus
	s.tools["setup_monitor"] = s.handleSetupMonitor
	s.tools["analyze_performance"] = s.handleAnalyzePerformance
}

// GetAvailableTools returns a list of available tools
func (s *MCPServer) GetAvailableTools() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]string, 0, len(s.tools))
	for name := range s.tools {
		tools = append(tools, name)
	}
	return tools
}

// HandleToolCall handles incoming MCP tool calls
func (s *MCPServer) HandleToolCall(ctx context.Context, toolName string, params json.RawMessage) (interface{}, error) {
	s.logger.Info("Handling MCP tool call", "tool", toolName)

	s.mu.RLock()
	handler, exists := s.tools[toolName]
	s.mu.RUnlock()

	if !exists {
		return ErrorResponse{
			Error:   "tool_not_found",
			Code:    404,
			Message: fmt.Sprintf("Tool '%s' not found", toolName),
		}, nil
	}

	result, err := handler(ctx, params)
	if err != nil {
		s.logger.Error("Tool execution failed", "tool", toolName, "error", err)
		return ErrorResponse{
			Error:   "tool_execution_failed",
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	s.logger.Info("Tool call completed successfully", "tool", toolName)
	return result, nil
}

// Start starts the MCP server
func (s *MCPServer) Start(ctx context.Context) error {
	s.logger.Info("Starting MCP server")

	// Here we would typically start the MCP transport (stdio, TCP, etc.)
	// For now, we'll just log that the server is ready
	s.logger.Info("MCP server is ready to handle tool calls")

	<-ctx.Done()
	s.logger.Info("MCP server shutting down")
	return nil
}

// Stop stops the MCP server and cancels all running simulations
func (s *MCPServer) Stop() {
	s.logger.Info("Stopping MCP server")

	s.mu.Lock()
	defer s.mu.Unlock()

	for id, tracker := range s.simulations {
		if tracker.Cancel != nil {
			tracker.Cancel()
		}
		delete(s.simulations, id)
	}
}

// addSimulationTracker adds a simulation tracker
func (s *MCPServer) addSimulationTracker(tracker *SimulationTracker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.simulations[tracker.ID] = tracker
}

// removeSimulationTracker removes a simulation tracker
func (s *MCPServer) removeSimulationTracker(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.simulations, id)
}

// getSimulationTracker gets a simulation tracker
func (s *MCPServer) getSimulationTracker(id uuid.UUID) (*SimulationTracker, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tracker, exists := s.simulations[id]
	return tracker, exists
}

// getAllSimulationTrackers gets all simulation trackers
func (s *MCPServer) getAllSimulationTrackers() []*SimulationTracker {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trackers := make([]*SimulationTracker, 0, len(s.simulations))
	for _, tracker := range s.simulations {
		trackers = append(trackers, tracker)
	}
	return trackers
}
