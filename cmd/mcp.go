package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Easy-Infra-Ltd/easy-test/internal/mcp"
	"github.com/spf13/cobra"
)

var (
	mcpPort      int
	mcpTransport string
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP (Model Context Protocol) server",
	Long: `Start the MCP server to expose Easy-Test functionality through the Model Context Protocol.

This allows AI assistants and other MCP clients to interact with Easy-Test programmatically,
including running simulations, validating configurations, querying logs, and monitoring services.

The MCP server exposes the following tools:
- run_simulation: Execute test simulations
- list_simulations: List running and completed simulations
- stop_simulation: Stop running simulations
- get_simulation_results: Get simulation results
- validate_config: Validate simulation configurations
- generate_config_template: Generate configuration templates
- query_logs: Query and search logs
- get_log_summary: Get log summaries
- get_monitor_status: Get monitoring status
- setup_monitor: Setup monitoring for targets
- analyze_performance: Analyze performance metrics

Examples:
  easy-test mcp                    # Start MCP server on default port
  easy-test mcp --port 8080        # Start MCP server on port 8080
  easy-test mcp --transport stdio  # Start MCP server using stdio transport`,
	Run: runMCPServer,
}

func init() {
	rootCmd.AddCommand(mcpCmd)

	// MCP-specific flags
	mcpCmd.Flags().IntVarP(&mcpPort, "port", "p", 3000,
		"port to run the MCP server on (TCP transport only)")
	mcpCmd.Flags().StringVarP(&mcpTransport, "transport", "t", "stdio",
		"transport protocol (stdio or tcp)")
}

func runMCPServer(cmd *cobra.Command, args []string) {
	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create and start MCP server
	server := mcp.NewMCPServer()

	// Start server in goroutine
	go func() {
		if err := server.Start(ctx); err != nil {
			cmd.Printf("MCP server error: %v\n", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	cmd.Println("Shutting down MCP server...")

	// Stop the server
	server.Stop()
	cancel()

	cmd.Println("MCP server stopped")
}
