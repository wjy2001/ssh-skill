---
title: 快速入门
description: ssh-mcp 的安装、配置和 5 分钟上手教程
doc_type: tutorial
last_updated: 2026-07-15
audience: [新用户, 所有开发者]
---

# 快速入门

5 分钟内完成安装、初始化并执行第一条远程命令。

## 前置条件

- Go 1.18+（构建 ssh-mcp 二进制；2022 年 3 月发布，绝大多数环境已具备）
- 目标服务器运行标准 OpenSSH
- 对目标服务器有 SSH 访问权限（密码、密钥或 SSH agent）

## 安装

ssh-mcp **不提供预编译二进制下载**。所有使用者从源码构建——这让分发透明、可审计、可钉到任意 commit。

### 从源码构建（推荐）

```bash
git clone <your-fork-or-mirror-url> ssh-skill
cd ssh-skill

# Linux / macOS
./scripts/build.sh

# Windows (PowerShell)
.\scripts\build.ps1
```

构建脚本把 `go/cmd/ssh-mcp/` 编译到 `.claude/skills/ssh-ops/bin/ssh-mcp`（Windows 下为 `ssh-mcp.exe`），无需任何额外依赖。

### 手动构建

如果不想用构建脚本：

```bash
cd go
go build -o ../.claude/skills/ssh-ops/bin/ssh-mcp ./cmd/ssh-mcp/
```

### 验证安装

```bash
ssh-mcp --version
# 或直接调用构建产物：
.claude/skills/ssh-ops/bin/ssh-mcp --version
```

输出版本号即构建成功。

### 安装为 Claude Code 全局技能

构建后，把技能目录拷贝到全局 Claude skills 文件夹，任意项目即可使用：

```bash
# Linux / macOS
mkdir -p ~/.claude/skills/ssh-ops
cp -r .claude/skills/ssh-ops/SKILL.md .claude/skills/ssh-ops/bin ~/.claude/skills/ssh-ops/

# Windows (PowerShell)
New-Item -ItemType Directory -Force -Path $env:USERPROFILE\.claude\skills\ssh-ops
Copy-Item .claude\skills\ssh-ops\SKILL.md, .claude\skills\ssh-ops\bin $env:USERPROFILE\.claude\skills\ssh-ops\ -Recurse -Force
```

## 首次配置

### 1. 初始化 Vault

```bash
ssh-mcp vault init
```

该命令会创建 `~/.ssh-mcp/` 目录（权限 0700）、生成随机 32 字节 AES-256 密钥、创建空的加密配置文件。

### 2. 添加服务器

```bash
# 密码认证
ssh-mcp add --id my-server --name "生产服务器" --host 10.0.0.1 --user root --auth-type password --password <your-password>

# SSH 密钥认证
ssh-mcp add --id dev-box --name "开发机" --host 192.168.1.100 --user dev --auth-type key --key-path ~/.ssh/id_rsa

# SSH Agent 认证
ssh-mcp add --id jump-host --name "跳板机" --host jump.example.com --user ops --auth-type agent
```

### 3. 测试连接

```bash
ssh-mcp test --server my-server
```

### 4. 执行命令

```bash
ssh-mcp exec --server my-server --command "uptime"
ssh-mcp exec --server my-server --command "df -h"
```

### 5. 文件传输

```bash
ssh-mcp upload --server my-server --local ./app.tar.gz --remote /tmp/app.tar.gz
ssh-mcp download --server my-server --remote /var/log/app.log --local ./app.log
```

## 下一步

- [`cli-reference.md`](./cli-reference.md) — 所有命令的完整参数参考
- [`guides.md`](./guides.md) — Claude Code 集成、部署到生产环境
- [`security.md`](./security.md) — 理解凭证加密和安全模型
