# ssh-skill

Secure SSH remote operations for AI agents (Claude Code).

`ssh-skill` is a single Go binary that provides a secure command channel between AI agents and remote servers. It encrypts credentials locally, validates target servers before execution, and audits `exec` operations for auditability.

## Features

- **Secure credential storage**: AES-256-GCM encrypted vault, auto-generated key
- **Target validation**: Commands can only target pre-configured servers (prevents AI hallucination)
- **Multi-auth support**: Password, SSH key, and SSH agent authentication
- **Audit logging**: Every `exec` is logged to a JSONL audit trail (upload/download/test and config commands are not audited today)
- **CLI-first**: Usable with or without AI — `ssh-skill exec` works directly in your terminal
- **Local only**: All configuration and credentials stored in `~/.ssh-skill/`, never leave your machine

## Installation

> ssh-skill is a Go CLI. The repository ships **pre-built binaries** for Linux and Windows under `.claude/skills/ssh-skill/bin/` — clone and use directly, no build step required. To rebuild from source (e.g. to pin a custom commit or patch), see [Build from source](#build-from-source-optional) below.

### Use the pre-built binary (recommended)

```bash
git clone <your-fork-or-mirror-url> ssh-skill
cd ssh-skill

# Linux / macOS
.claude/skills/ssh-skill/bin/ssh-skill --version

# Windows (PowerShell)
.\.claude\skills\ssh-skill\bin\ssh-skill.exe --version
```

Binaries are checked into the repo under `.claude/skills/ssh-skill/bin/`. No external runtime dependencies; the Go toolchain is only needed if you rebuild from source.

### Build from source (optional)

If you prefer to build locally (e.g. to pin a custom commit or patch):

```bash
# Linux / macOS
./scripts/build.sh

# Windows (PowerShell)
.\scripts\build.ps1
```

The build script compiles `go/cmd/ssh-skill/` into `.claude/skills/ssh-skill/bin/ssh-skill` (or `ssh-skill.exe` on Windows), overwriting the pre-built binary. Requires Go 1.18+.

### Manual build

If you prefer not to use the build script:

```bash
cd go
go build -o ../.claude/skills/ssh-skill/bin/ssh-skill ./cmd/ssh-skill/
```

### Install the Claude Code skill globally

After cloning, copy the skill directory into your global Claude skills folder so any project can use it:

```bash
# Linux / macOS
mkdir -p ~/.claude/skills/ssh-skill
cp -r .claude/skills/ssh-skill/SKILL.md .claude/skills/ssh-skill/bin ~/.claude/skills/ssh-skill/

# Windows (PowerShell)
New-Item -ItemType Directory -Force -Path $env:USERPROFILE\.claude\skills\ssh-skill
Copy-Item .claude\skills\ssh-skill\SKILL.md, .claude\skills\ssh-skill\bin $env:USERPROFILE\.claude\skills\ssh-skill\ -Recurse -Force
```

Verify the binary is on your PATH or referenced by the skill:

```bash
ssh-skill --version
```

## Quick Start

```bash
# 1. Initialize the vault (creates ~/.ssh-skill/)
ssh-skill vault init

# 2. Add a server
ssh-skill add --id my-server --name "My Server" --host 10.0.0.1 --user root --auth-type password --password <password>

# 3. Execute a command
ssh-skill exec --server my-server --command "uptime"

# 4. Upload a file
ssh-skill upload --server my-server --local ./app.tar.gz --remote /tmp/app.tar.gz

# 5. List all servers
ssh-skill list
```

## Claude Code Integration

The `ssh-skill` skill is self-contained under `.claude/skills/ssh-skill/` (binary shipped with the repo). After cloning, Claude Code will automatically use the skill when you ask to perform SSH operations — no build step or installer download required.

## Security Model

- **Threat model**: Defends against passive credential leakage (environment variables, chat history, plaintext config files) and AI hallucination attacks (connecting to unauthorized servers). Does **not** defend against MITM (host key verification currently disabled) or active attacks where an attacker reads the vault key file via Bash. See [`docs/security.md`](./docs/security.md) for the full model.
- **Encryption**: AES-256-GCM. A random 32-byte vault key is generated on first run; Argon2id (time=3, memory=64MB, threads=4) stretches that key with a per-encryption salt. This is not an end-user passphrase KDF.
- **Storage**: All data in `~/.ssh-skill/` with 0600 file permissions.
- **Audit**: Every `exec` is logged with timestamp, server, command, `exit_code`, stdout/stderr lengths, and duration. Other subcommands (upload/download/test/add/remove/list) are not audited today.
- **Process exit codes**: `0` means the client completed the session (the remote command may still have returned non-zero — see audit `exit_code`); `1` means a client, connection, or configuration error.

## Commands

| Command | Description |
|---------|------------|
| `ssh-skill list` | List all configured servers |
| `ssh-skill add` | Add a server configuration |
| `ssh-skill remove --id <id>` | Remove a server configuration |
| `ssh-skill exec --server <id> --command <cmd>` | Execute a command |
| `ssh-skill upload --server <id> --local <p> --remote <p>` | Upload a file |
| `ssh-skill download --server <id> --remote <p> --local <p>` | Download a file |
| `ssh-skill test --server <id>` | Test SSH connection |
| `ssh-skill vault init` | Initialize vault and key |
| `ssh-skill --version, -V` | Print version |

## Project Structure

```
ssh-skill/
├── go/                           # Go module
│   ├── cmd/ssh-skill/main.go       # Entry point
│   └── internal/
│       ├── types/                # Shared data types
│       ├── config/               # Configuration resolution
│       ├── vault/                # AES-256-GCM encryption + key management
│       ├── ssh/                  # SSH client (Client wrapper + bastion lifecycle), exec, transfer
│       ├── audit/                # JSONL audit logging
│       └── cli/                  # CLI subcommands
├── .claude/skills/ssh-skill/       # Claude Code skill (SKILL.md + checked-in bin/)
├── scripts/                      # Cross-platform build scripts (build.sh, build.ps1)
├── docs/                         # Project documentation
└── .harness/                     # Harness CE task management
```

## Requirements

- Go 1.18+
- Target servers must run standard OpenSSH

## License

[Project License]
