# ssh-skill

Secure SSH remote operations for AI agents (Claude Code / Codex).

`ssh-skill` is a single Go binary that provides a secure command channel between AI agents and remote servers. It encrypts credentials locally, validates target servers before execution, and audits `exec` operations for auditability.

## Features

- **Secure credential storage**: AES-256-GCM encrypted vault, auto-generated key
- **Target validation**: Commands can only target pre-configured servers (prevents AI hallucination)
- **Multi-auth support**: Password, SSH key, and SSH agent authentication
- **Audit logging**: Every `exec` is logged to a JSONL audit trail (upload/download/test and config commands are not audited today)
- **CLI-first**: Usable with or without AI — `ssh-skill exec` works directly in your terminal
- **Local only**: All configuration and credentials stored in `~/.ssh-skill/`, never leave your machine

## Install with Claude Code / Codex (recommended)

**你不需要先手动 clone 本仓库，也不应下载全量源码。**
安装只应拿到：`SKILL.md` + **当前平台**预编译二进制。
不要拉取 `go/`、`.harness/`、docs 源码树，也不要 `git clone` 全仓库。

打开本 README → 复制下面提示词 → 粘贴到 Claude Code / Codex → 让 agent 自动最小安装。

### 一键安装提示词（复制即用）

```text
请帮我安装 GitHub 上的 ssh-skill（安全 SSH 远程操作技能 / CLI）。

仓库：https://github.com/wjy2001/ssh-skill
安装脚本（优先）：
- Linux / macOS:
  curl -fsSL https://raw.githubusercontent.com/wjy2001/ssh-skill/master/scripts/install-skill.sh | bash
- Windows (PowerShell):
  irm https://raw.githubusercontent.com/wjy2001/ssh-skill/master/scripts/install-skill.ps1 | iex

硬性要求（最小下载）：
1. 禁止 git clone / 禁止下载 zipball / 禁止拉取全量仓库
2. 只允许下载安装所需文件：
   - SKILL.md
   - 当前平台二进制（Linux/macOS: bin/ssh-skill，Windows: bin/ssh-skill.exe）
3. 安装到全局 Claude skills 目录：
   - Linux / macOS: ~/.claude/skills/ssh-skill/
   - Windows: %USERPROFILE%\.claude\skills\ssh-skill\
4. 覆盖旧版 skill；删除另一平台残留二进制
5. 运行 --version 验证
6. 不要修改用户已有的 ~/.ssh-skill/ vault 数据
7. 不要默认要求安装 Go，不要从源码构建（除非最小下载失败且用户明确同意）

若脚本不可用，再回退到 raw.githubusercontent.com 逐文件下载（仍禁止 clone）：
- https://raw.githubusercontent.com/wjy2001/ssh-skill/master/.claude/skills/ssh-skill/SKILL.md
- https://raw.githubusercontent.com/wjy2001/ssh-skill/master/.claude/skills/ssh-skill/bin/<platform-binary>

完成后用简洁中文汇报：安装路径、版本号、下一步 vault init / 添加服务器。
告诉我：现在可以直接说「列出已配置服务器」或「在 my-server 上执行 uptime」。
```

### 为什么这样更省流量

| 方式 | 大约下载 | 是否暴露全仓库 |
|------|----------|----------------|
| `git clone` 全量 | 全仓库（源码 + 双平台二进制 + 文档 + harness） | 是 |
| 最小安装脚本 | `SKILL.md` + **1 个**平台二进制（约 6MB） | 否（只取 skill 文件） |

仓库对公众仍可读；这里约束的是 **agent 安装路径只取 skill 与二进制**，不把源码树装进用户机器。

### 安装后首次配置提示词（可选）

安装完成后，若要继续让 agent 初始化并添加服务器，可再发：

```text
ssh-skill 已安装。请帮我完成首次配置：
1. 执行 vault init（幂等，不要清空已有配置）
2. 询问我服务器信息（id / host / user / auth-type / password 或 key-path）后再 add
3. 用 test 验证连通性
4. 不要在聊天中回显明文密码
```

### Agent 安装后的标准落点

| 平台 | 全局技能路径 |
|------|----------------|
| Linux / macOS | `~/.claude/skills/ssh-skill/` |
| Windows | `%USERPROFILE%\.claude\skills\ssh-skill\` |

目录结构应为：

```text
~/.claude/skills/ssh-skill/
├── SKILL.md
└── bin/
    ├── ssh-skill          # Linux / macOS
    └── ssh-skill.exe      # Windows
```

## Manual Installation

若你更想自己装，而不是让 agent 装：

### Use the pre-built binary

```bash
git clone https://github.com/wjy2001/ssh-skill.git
cd ssh-skill

# Linux / macOS
.claude/skills/ssh-skill/bin/ssh-skill --version

# Windows (PowerShell)
.\.claude\skills\ssh-skill\bin\ssh-skill.exe --version
```

### Install the skill globally

```bash
# Linux / macOS
mkdir -p ~/.claude/skills/ssh-skill
cp -r .claude/skills/ssh-skill/SKILL.md .claude/skills/ssh-skill/bin ~/.claude/skills/ssh-skill/

# Windows (PowerShell)
New-Item -ItemType Directory -Force -Path $env:USERPROFILE\.claude\skills\ssh-skill
Copy-Item .claude\skills\ssh-skill\SKILL.md, .claude\skills\ssh-skill\bin $env:USERPROFILE\.claude\skills\ssh-skill\ -Recurse -Force
```

### Build from source (optional)

仅在需要改代码或预编译二进制不可用时：

```bash
# Linux / macOS
./scripts/build.sh

# Windows (PowerShell)
.\scripts\build.ps1
```

或手动：

```bash
cd go
go build -o ../.claude/skills/ssh-skill/bin/ssh-skill ./cmd/ssh-skill/
```

需要 Go 1.18+。

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

## Claude Code / Codex Integration

推荐分发方式：**用户复制 README 安装提示词 / 运行最小安装脚本 → 只下载 skill + 当前平台二进制 → 装到全局 skills 目录**。

安装完成后，直接用自然语言即可，例如：

```text
帮我在 my-server 上检查磁盘使用情况
把 app.tar.gz 上传到生产服务器
列出所有已配置的服务器
```

技能本体位于仓库 `.claude/skills/ssh-skill/`（`SKILL.md` + 预编译 `bin/`）。全局安装后，Claude Code / 兼容 skill 机制的 agent 在处理 SSH 任务时应优先走该技能。

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
