---
title: 快速入门
description: ssh-skill 的安装、配置和 5 分钟上手教程
doc_type: tutorial
last_updated: 2026-07-16
audience: [新用户, 所有开发者]
---

# 快速入门

5 分钟内完成安装、初始化并执行第一条远程命令。

## 前置条件

- Go 1.18+（构建 ssh-skill 二进制；2022 年 3 月发布，绝大多数环境已具备）
- 目标服务器运行标准 OpenSSH
- 对目标服务器有 SSH 访问权限（密码、密钥或 SSH agent）

## 安装

仓库**自带预编译二进制**（Linux 与 Windows），位于 `.claude/skills/ssh-skill/bin/`。clone 后直接可用，无需构建。如需从源码重新构建（例如钉到自定义 commit 或打补丁），见下方[从源码构建](#从源码构建可选)。

### 使用预编译二进制（推荐）

```bash
git clone <your-fork-or-mirror-url> ssh-skill
cd ssh-skill

# Linux / macOS
.claude/skills/ssh-skill/bin/ssh-skill --version

# Windows (PowerShell)
.\.claude\skills\ssh-skill\bin\ssh-skill.exe --version
```

二进制已签入仓库，无运行时依赖。仅在重新构建时需要 Go 工具链（1.18+）。

### 从源码构建（可选）

如需从源码构建（例如修改代码后重新打包）：

```bash
# Linux / macOS
./scripts/build.sh

# Windows (PowerShell)
.\scripts\build.ps1
```

构建脚本把 `go/cmd/ssh-skill/` 编译到 `.claude/skills/ssh-skill/bin/ssh-skill`（Windows 下为 `ssh-skill.exe`），会覆盖仓库自带的预编译二进制。需要 Go 1.18+。

### 手动构建

如果不想用构建脚本：

```bash
cd go
go build -o ../.claude/skills/ssh-skill/bin/ssh-skill ./cmd/ssh-skill/
```

### 验证安装

```bash
ssh-skill --version
# 或直接调用仓库自带二进制：
.claude/skills/ssh-skill/bin/ssh-skill --version
```

输出版本号即安装成功。

### 安装为 Claude Code 全局技能

clone 后，把技能目录拷贝到全局 Claude skills 文件夹，任意项目即可使用：

```bash
# Linux / macOS
mkdir -p ~/.claude/skills/ssh-skill
cp -r .claude/skills/ssh-skill/SKILL.md .claude/skills/ssh-skill/bin ~/.claude/skills/ssh-skill/

# Windows (PowerShell)
New-Item -ItemType Directory -Force -Path $env:USERPROFILE\.claude\skills\ssh-skill
Copy-Item .claude\skills\ssh-skill\SKILL.md, .claude\skills\ssh-skill\bin $env:USERPROFILE\.claude\skills\ssh-skill\ -Recurse -Force
```

## 首次配置

### 1. 初始化 Vault

```bash
ssh-skill vault init
```

该命令会创建 `~/.ssh-skill/` 目录（权限 0700）、生成随机 32 字节 AES-256 密钥、创建空的加密配置文件。

### 2. 添加服务器

```bash
# 密码认证
ssh-skill add --id my-server --name "生产服务器" --host 10.0.0.1 --user root --auth-type password --password <your-password>

# SSH 密钥认证
ssh-skill add --id dev-box --name "开发机" --host 192.168.1.100 --user dev --auth-type key --key-path ~/.ssh/id_rsa

# SSH Agent 认证
ssh-skill add --id jump-host --name "跳板机" --host jump.example.com --user ops --auth-type agent
```

### 3. 测试连接

```bash
ssh-skill test --server my-server
```

### 4. 执行命令

```bash
ssh-skill exec --server my-server --command "uptime"
ssh-skill exec --server my-server --command "df -h"
```

### 5. 文件传输

```bash
ssh-skill upload --server my-server --local ./app.tar.gz --remote /tmp/app.tar.gz
ssh-skill download --server my-server --remote /var/log/app.log --local ./app.log
```

## 下一步

- [`cli-reference.md`](./cli-reference.md) — 所有命令的完整参数参考
- [`guides.md`](./guides.md) — Claude Code 集成、部署到生产环境
- [`security.md`](./security.md) — 理解凭证加密和安全模型
