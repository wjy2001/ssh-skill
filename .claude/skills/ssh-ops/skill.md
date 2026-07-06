---
name: ssh-ops
description: |
  Secure SSH remote server operations via ssh-mcp CLI. Execute commands, upload/download files,
  and manage server configurations with encrypted credential storage. All commands require
  user approval before execution.
---

# SSH Remote Operations

Use `ssh-mcp` to operate on remote servers via SSH. All server credentials are stored locally
in an encrypted vault (`~/.ssh-mcp/servers.json.age`).

## Safety Rules

1. **Approval required**: Every `ssh-mcp exec` command requires user confirmation via Claude Code's permission dialog.
2. **Server must be pre-configured**: You cannot connect to a server that hasn't been added via `ssh-mcp add`.
3. **Credentials never leave the machine**: All passwords are AES-256-GCM encrypted on disk.
4. **Audit trail**: All commands are logged to `~/.ssh-mcp/audit.log`.

## Diagnostic Rules

1. **`ssh-mcp test` first**: Always start diagnosis with `ssh-mcp test --server <id>`. It exercises the exact same auth, encryption, and SSH stack as `ssh-mcp exec`, so it's the single source of truth for connectivity.
2. **Auxiliary tools only after `ssh-mcp` fails**: If `ssh-mcp test` fails, then use `ping`, `nslookup`, or other shell tools to narrow down the cause (network vs. auth vs. server-side issue). But never use them to *overrule* a successful `ssh-mcp test`.

## Workflow

### First-time setup

```bash
ssh-mcp vault init
```

### Add a server

```bash
# Password authentication
ssh-mcp add --id prod-web --name "Production Web" --host 10.0.1.100 --user deploy --auth-type password --password <password>

# Key authentication
ssh-mcp add --id prod-web --name "Production Web" --host 10.0.1.100 --user deploy --auth-type key --key-path ~/.ssh/prod_ed25519

# SSH agent authentication
ssh-mcp add --id prod-web --name "Production Web" --host 10.0.1.100 --user deploy --auth-type agent
```

### List configured servers

```bash
ssh-mcp list
```

### Execute a command

```bash
ssh-mcp exec --server prod-web --command "uptime"
ssh-mcp exec --server prod-web --command "df -h"
ssh-mcp exec --server prod-web --command "systemctl status nginx" --timeout 10
```

### Upload/download files

```bash
ssh-mcp upload --server prod-web --local ./app.tar.gz --remote /tmp/app.tar.gz
ssh-mcp download --server prod-web --remote /var/log/nginx/access.log --local ./access.log
```

### Test connection

```bash
ssh-mcp test --server prod-web
```

### Remove a server

```bash
ssh-mcp remove --id prod-web
```

## When to Use

- Execute commands on remote Linux servers
- Upload or download files
- Check server status (uptime, disk, memory, services)
- Deploy applications or configurations
- Inspect logs on remote servers

## When NOT to Use

- Interactive terminal sessions (vim, top, htop) — ssh-mcp is for non-interactive commands only
- Servers not pre-configured in the vault
- Operations requiring real-time streaming output

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| `server 'X' not found` | Server not added to vault | Run `ssh-mcp add` first |
| `connection failed` | Wrong host/port or network issue | Verify with `ssh-mcp test` |
| `vault: decryption failed` | Vault key corrupted or changed | Re-initialize with `ssh-mcp vault init` |
| `no valid authentication method` | Missing or wrong auth config | Re-add server with correct auth type |
