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

> ssh-mcp is a Go CLI. There are **no pre-built binary downloads** — build from source. This keeps the distribution transparent, auditable, and lets you pin to any commit. Requires Go 1.18+ (released March 2022).

### Build from source (recommended)

```bash
git clone <your-fork-or-mirror-url> ssh-skill
cd ssh-skill

# Linux / macOS
./scripts/build.sh

# Windows (PowerShell)
.\scripts\build.ps1
```

The build script compiles `go/cmd/ssh-mcp/` into `.claude/skills/ssh-ops/bin/ssh-mcp` (or `ssh-mcp.exe` on Windows). No external dependencies beyond the Go toolchain.

### Manual build

If you prefer not to use the build script:

```bash
cd go
go build -o ../.claude/skills/ssh-ops/bin/ssh-mcp ./cmd/ssh-mcp/
```

### Install the Claude Code skill globally

After building, copy the skill directory into your global Claude skills folder so any project can use it:

```bash
# Linux / macOS
mkdir -p ~/.claude/skills/ssh-ops
cp -r .claude/skills/ssh-ops/SKILL.md .claude/skills/ssh-ops/bin ~/.claude/skills/ssh-ops/

# Windows (PowerShell)
New-Item -ItemType Directory -Force -Path $env:USERPROFILE\.claude\skills\ssh-ops
Copy-Item .claude\skills\ssh-ops\SKILL.md, .claude\skills\ssh-ops\bin $env:USERPROFILE\.claude\skills\ssh-ops\ -Recurse -Force
```

Verify the binary is on your PATH or referenced by the skill:

```bash
ssh-mcp --version
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

The `ssh-ops` skill is self-contained under `.claude/skills/ssh-ops/` (binary built locally). After running the build script, Claude Code will automatically use the skill when you ask to perform SSH operations — no installer download required.

## Security Model

- **Threat model**: Defends against passive credential leakage (environment variables, chat history, plaintext config files) and AI hallucination attacks (connecting to unauthorized servers). Does **not** defend against MITM (host key verification currently disabled) or active attacks where an attacker reads the vault key file via Bash. See [`docs/security.md`](./docs/security.md) for the full model.
- **Encryption**: AES-256-GCM with Argon2id key derivation (time=3, memory=64MB, threads=4). Random 32-byte key generated on first run.
- **Storage**: All data in `~/.ssh-mcp/` with 0600 file permissions.
- **Audit**: Every command execution is logged with timestamp, server, command, exit code, stdout/stderr length, and duration.

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
| `ssh-mcp --version, -V` | Print version |
| `ssh-mcp serve` | MCP server mode (not yet implemented) |

## Project Structure

```
ssh-skill/
├── go/                           # Go module
│   ├── cmd/ssh-mcp/main.go       # Entry point
│   └── internal/
│       ├── types/                # Shared data types
│       ├── config/               # Configuration resolution
│       ├── vault/                # AES-256-GCM encryption + key management
│       ├── ssh/                  # SSH client (Client wrapper + bastion lifecycle), exec, transfer
│       ├── audit/                # JSONL audit logging
│       └── cli/                  # CLI subcommands
├── .claude/skills/ssh-ops/       # Claude Code skill (SKILL.md + bin/ build output)
├── scripts/                      # Cross-platform build scripts (build.sh, build.ps1)
├── docs/                         # Project documentation
└── .harness/                     # Harness CE task management
```

## Requirements

- Go 1.18+
- Target servers must run standard OpenSSH

## License

[Project License]
