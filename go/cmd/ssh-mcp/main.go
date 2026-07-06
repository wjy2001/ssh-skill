package main

import (
	"os"

	"ssh-mcp/internal/cli"
)

var version = "0.1.0"

func main() {
	if err := cli.Run(os.Args); err != nil {
		os.Exit(1)
	}
	_ = version
}
