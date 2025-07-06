package mcp

import (
	"encoding/json"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/simulation"
	"github.com/google/uuid"
)

// SimulationRequest represents a request to run a simulation
type SimulationRequest struct {
	Name     string                            `json:"name"`
	Target   simulation.SimulationTargetConfig `json:"target"`
	Cadence  time.Duration                     `json:"cadence"`
	Attempts int                               `json:"attempts"`
	Dry      bool                              `json:"dry,omitempty"`
}

// SimulationResponse represents the response from a simulation
type SimulationResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Status  string    `json:"status"`
	Message string    `json:"message,omitempty"`
	Results []string  `json:"results,omitempty"`
}

// ConfigValidationRequest represents a request to validate configuration
type ConfigValidationRequest struct {
	Config json.RawMessage `json:"config"`
	Type   string          `json:"type"` // "simulation" or "monitor"
}

// ConfigValidationResponse represents the response from config validation
type ConfigValidationResponse struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message"`
}

// LogQueryRequest represents a request to query logs
type LogQueryRequest struct {
	Query     string    `json:"query"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Level     string    `json:"level,omitempty"`
	Limit     int       `json:"limit,omitempty"`
}

// LogQueryResponse represents the response from log query
type LogQueryResponse struct {
	Logs    []LogEntry `json:"logs"`
	Total   int        `json:"total"`
	Message string     `json:"message"`
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Area      string    `json:"area,omitempty"`
	Process   string    `json:"process,omitempty"`
	TraceID   string    `json:"trace_id,omitempty"`
}

// MonitorStatusRequest represents a request to get monitor status
type MonitorStatusRequest struct {
	Name string `json:"name,omitempty"`
}

// MonitorStatusResponse represents the response from monitor status
type MonitorStatusResponse struct {
	Monitors []MonitorInfo `json:"monitors"`
	Message  string        `json:"message"`
}

// MonitorInfo represents information about a monitor
type MonitorInfo struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Targets   []string  `json:"targets"`
	StartTime time.Time `json:"start_time"`
	LastCheck time.Time `json:"last_check,omitempty"`
}

// ConfigTemplateRequest represents a request to generate a config template
type ConfigTemplateRequest struct {
	Type       string            `json:"type"` // "simulation" or "monitor"
	Parameters map[string]string `json:"parameters,omitempty"`
}

// ConfigTemplateResponse represents the response from config template generation
type ConfigTemplateResponse struct {
	Template json.RawMessage `json:"template"`
	Message  string          `json:"message"`
}

// SimulationListResponse represents the response from listing simulations
type SimulationListResponse struct {
	Simulations []SimulationInfo `json:"simulations"`
	Message     string           `json:"message"`
}

// SimulationInfo represents information about a simulation
type SimulationInfo struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Attempts  int        `json:"attempts"`
	Progress  int        `json:"progress"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
