package cli

import (
	"fmt"
)

func cmdServe(args []string) error {
	fmt.Println("MCP server mode is not yet implemented.")
	fmt.Println("Use the CLI subcommands directly, or configure as an MCP server with:")
	fmt.Println("  claude mcp add ssh-mcp -- /path/to/ssh-mcp serve")
	return nil
}
