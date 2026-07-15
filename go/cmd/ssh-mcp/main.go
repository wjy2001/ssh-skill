package main

import (
	"os"

	"ssh-mcp/internal/cli"
)

// version is injected at build time via:
//   go build -ldflags "-X main.version=v0.1.0"
// Default "dev" when built from source without ldflags.
var version = "dev"

func main() {
	cli.SetVersion(version)
	if err := cli.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
