package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"technocrat/internal/mcp"
)

var (
	serverPort int
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the MCP server",
	Long: `Start the Technocrat Model Context Protocol (MCP) server.

The server listens for MCP protocol requests and provides tools,
resources, and prompts to connected clients.`,
	RunE: runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Port to listen on")
}

func runServer(cmd *cobra.Command, args []string) error {
	log.Printf("Starting Technocrat MCP Server on port %d...", serverPort)

	server := mcp.NewServer(serverPort)
	if err := server.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
