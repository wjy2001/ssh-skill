# ssh-mcp

Secure SSH remote operations for AI agents (Claude Code).

`ssh-mcp` is a single Go binary that provides a secure command channel between AI agents and remote servers. It encrypts credentials locally, validates target servers before execution, and logs all commands for auditability.

## Features

- **Secure credential storage**: AES-256-GCM encrypted vault, auto-generated key
- **Target validation**: Commands can only target pre-configured servers (prevents AI hallucination)
- **Multi-auth support**: Password, SSH key, and SSH agent authentication
- **Audit logging**: All commands logged to JSONL audit trail
- **CLI-first**: Usable with or without AI — `ssh-mcp exec` works directly in your terminal
- **Local only**: All configuration and credentials stored in `~/.ssh-mcp/`, never leave your machine

## Installation

```bash
# Clone and build
git clone <repo-url>
cd ssh-skill/go
go build -o ~/bin/ssh-mcp ./cmd/ssh-mcp/

# Or use pre-built binary (place in PATH)
# ssh-mcp is a single static binary, no runtime dependencies
```

## Quick Start

```bash
# 1. Initialize the vault (creates ~/.ssh-mcp/)
ssh-mcp vault init

# 2. Add a server
ssh-mcp add --id my-server --name "My Server" --host 10.0.0.1 --user root --auth-type password --password <password>

# 3. Execute a command
ssh-mcp exec --server my-server --command "uptime"

# 4. Upload a file
ssh-mcp upload --server my-server --local ./app.tar.gz --remote /tmp/app.tar.gz

# 5. List all servers
ssh-mcp list
```

## Claude Code Integration

Add the `ssh-ops` skill to your project or global skills directory:

```bash
# Copy the skill definition
mkdir -p .claude/skills/ssh-ops/
cp path/to/ssh-skill/.claude/skills/ssh-ops/SKILL.md .claude/skills/ssh-ops/
```

Claude Code will automatically use the skill when you ask to perform SSH operations.

## Security Model

- **Threat model**: Defends against passive credential leakage (environment variables, chat history). Does not defend against active attacks (AI reading vault key file via Bash).
- **Encryption**: AES-256-GCM with Argon2id key derivation. Random 32-byte key generated on first run.
- **Storage**: All data in `~/.ssh-mcp/` with 0600 file permissions.
- **Audit**: Every command execution is logged with timestamp, server, command, exit code, and duration.

## Commands

| Command | Description |
|---------|------------|
| `ssh-mcp list` | List all configured servers |
| `ssh-mcp add` | Add a server configuration |
| `ssh-mcp remove --id <id>` | Remove a server configuration |
| `ssh-mcp exec --server <id> --command <cmd>` | Execute a command |
| `ssh-mcp upload --server <id> --local <p> --remote <p>` | Upload a file |
| `ssh-mcp download --server <id> --remote <p> --local <p>` | Download a file |
| `ssh-mcp test --server <id>` | Test SSH connection |
| `ssh-mcp vault init` | Initialize vault and key |
| `ssh-mcp serve` | MCP server mode (experimental) |

## Project Structure

```
ssh-skill/
├── go/                           # Go module
│   ├── cmd/ssh-mcp/main.go       # Entry point
│   └── internal/
│       ├── types/                # Shared data types
│       ├── config/               # Configuration resolution
│       ├── vault/                # AES-256-GCM encryption + key management
│       ├── ssh/                  # SSH client, exec, file transfer
│       ├── audit/                # JSONL audit logging
│       └── cli/                  # CLI subcommands
├── .claude/skills/ssh-ops/       # Claude Code skill definition
├── bin/                          # Build artifacts
└── .harness/                     # Harness CE task management
```

## Requirements

- Go 1.25+
- Target servers must run standard OpenSSH

## License

[Project License]
