// Package cli implements the ssh-mcp command-line interface.
// Each subcommand is defined in its own file and registered in root.go.
package cli

import (
	"fmt"
	"os"

	"ssh-mcp/internal/audit"
	"ssh-mcp/internal/config"
	"ssh-mcp/internal/types"
	"ssh-mcp/internal/vault"
)

// App holds the runtime state shared across all subcommands.
type App struct {
	Vault      *types.Vault
	VaultKey   []byte
	VaultPath  string
	AuditLog   *audit.Logger
	ConfigDir  string
}

// Load initializes the application state: config directory, vault key, vault data.
func Load() (*App, error) {
	configDir, err := config.Dir()
	if err != nil {
		return nil, fmt.Errorf("resolve config dir: %w", err)
	}

	keyPath, err := config.VaultKeyPath()
	if err != nil {
		return nil, fmt.Errorf("resolve key path: %w", err)
	}

	vaultKey, err := vault.EnsureKey(keyPath)
	if err != nil {
		return nil, fmt.Errorf("ensure vault key: %w", err)
	}

	vaultPath, err := config.VaultFilePath()
	if err != nil {
		return nil, fmt.Errorf("resolve vault path: %w", err)
	}

	v, err := vault.Load(vaultPath, vaultKey)
	if err != nil {
		return nil, fmt.Errorf("load vault: %w", err)
	}

	auditPath, err := config.AuditLogPath()
	if err != nil {
		return nil, fmt.Errorf("resolve audit path: %w", err)
	}

	auditLog, err := audit.NewLogger(auditPath)
	if err != nil {
		return nil, fmt.Errorf("init audit: %w", err)
	}

	return &App{
		Vault:     v,
		VaultKey:  vaultKey,
		VaultPath: vaultPath,
		AuditLog:  auditLog,
		ConfigDir: configDir,
	}, nil
}

// Save persists the current vault state to disk.
func (a *App) Save() error {
	return vault.Save(a.VaultPath, a.VaultKey, a.Vault)
}

// Run executes the CLI with the given arguments.
// It dispatches to the appropriate subcommand.
func Run(args []string) error {
	if len(args) < 2 {
		printUsage()
		return nil
	}

	cmd := args[1]
	cmdArgs := args[2:]

	switch cmd {
	case "list":
		return cmdList(cmdArgs)
	case "add":
		return cmdAdd(cmdArgs)
	case "remove":
		return cmdRemove(cmdArgs)
	case "exec":
		return cmdExec(cmdArgs)
	case "upload":
		return cmdUpload(cmdArgs)
	case "download":
		return cmdDownload(cmdArgs)
	case "test":
		return cmdTest(cmdArgs)
	case "vault":
		return cmdVault(cmdArgs)
	case "serve":
		return cmdServe(cmdArgs)
	case "--help", "-h", "help":
		printUsage()
		return nil
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func printUsage() {
	fmt.Println(`ssh-mcp — secure SSH remote operations

Usage:
  ssh-mcp list                    List all configured servers
  ssh-mcp add                     Add a new server configuration
  ssh-mcp remove --id <id>        Remove a server configuration
  ssh-mcp exec --server <id> --command <cmd>  Execute a command
  ssh-mcp upload --server <id> --local <path> --remote <path>  Upload a file
  ssh-mcp download --server <id> --remote <path> --local <path>  Download a file
  ssh-mcp test --server <id>      Test SSH connection
  ssh-mcp vault init              Initialize vault key and config
  ssh-mcp serve                   Start MCP server mode (experimental)

Use 'ssh-mcp <command> --help' for detailed options.`)
}
