package cmd

import (
	"fmt"
	"log"

	"technocrat/internal/mcp"

	"github.com/spf13/cobra"
)

var (
	serverPort  int
	serverStdio bool
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the MCP server",
	Long: `Start the Technocrat Model Context Protocol (MCP) server.

The server can run in two modes:
- HTTP mode (default): Listens on a port for HTTP requests (for Amazon Q, VS Code)  
- stdio mode: Uses stdin/stdout for communication (for Claude Desktop, Cursor, Windsurf)

The server provides tools, resources, and prompts to connected clients.`,
	RunE: runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Port to listen on (HTTP mode)")
	serverCmd.Flags().BoolVar(&serverStdio, "stdio", false, "Use stdio transport (for Claude Desktop)")
}

func runServer(cmd *cobra.Command, args []string) error {
	if serverStdio {
		log.Printf("Starting Technocrat MCP Server in stdio mode...")
		server := mcp.NewStdioServer()
		if err := server.Start(); err != nil {
			return fmt.Errorf("failed to start stdio server: %w", err)
		}
	} else {
		log.Printf("Starting Technocrat MCP Server on port %d...", serverPort)
		server := mcp.NewServer(serverPort)
		if err := server.Start(); err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}
	}

	return nil
}
