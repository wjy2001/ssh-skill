---
title: 快速入门
description: ssh-mcp 的安装、配置和 5 分钟上手教程
doc_type: tutorial
last_updated: 2026-07-06
audience: [新用户, 所有开发者]
---

# 快速入门

5 分钟内完成安装、初始化并执行第一条远程命令。

## 前置条件

- Go 1.25+（仅源码构建需要）
- 目标服务器运行标准 OpenSSH
- 对目标服务器有 SSH 访问权限（密码、密钥或 SSH agent）

## 安装

### 从源码构建

```bash
git clone <repo-url>
cd ssh-skill/go
go build -o ~/bin/ssh-mcp ./cmd/ssh-mcp/
export PATH="$HOME/bin:$PATH"
```

编译产物为单一静态二进制，无运行时依赖。

### 验证安装

```bash
ssh-mcp --help
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
