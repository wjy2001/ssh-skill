---
title: 操作指南
description: ssh-mcp 的 Claude Code 集成、部署、CI/CD 和排错指南
doc_type: how-to
last_updated: 2026-07-06
audience: [日常使用者, DevOps, 维护者]
---

# 操作指南

日常使用中的常见任务和问题解决。

## Claude Code 集成

### 安装技能

将 `ssh-ops` 技能安装到你的项目中：

```bash
# 复制技能定义到项目
mkdir -p .claude/skills/ssh-ops/
cp .claude/skills/ssh-ops/SKILL.md .claude/skills/ssh-ops/
```

Claude Code 会在你请求 SSH 操作时自动加载该技能。

### 使用方式

安装后，直接用自然语言请求：

```
> 帮我在 my-server 上检查磁盘使用情况
> 把 app.tar.gz 上传到生产服务器
> 列出所有已配置的服务器
```

Claude Code 会自动调用 `ssh-mcp` CLI，并触发审批对话框。

### 安全规则

技能内置的安全规则：

1. **审批要求**：每个 `ssh-mcp exec` 命令需要用户确认
2. **服务器预配置**：不能连接未通过 `ssh-mcp add` 添加的服务器
3. **凭证不离开本机**：密码 AES-256-GCM 加密存储
4. **审计追踪**：所有命令记录到 `~/.ssh-mcp/audit.log`

### 诊断规则

1. **`ssh-mcp test` 优先**：诊断连接问题时，始终先用 `ssh-mcp test --server <id>`。它使用与 `exec` 完全相同的认证、加密和 SSH 协议栈。
2. **辅助工具仅在 `ssh-mcp` 失败后使用**：`ssh-mcp test` 失败后，才用 `ping`、`nslookup` 等工具缩小原因范围（网络 vs 认证 vs 服务端）。但绝不用它们推翻一次成功的 `ssh-mcp test`。

## 部署

### 本地开发环境

```bash
# Use the build scripts (recommended):
#   Linux/macOS: scripts/build.sh
#   Windows:     scripts/build.ps1
# Or build manually:
cd go
go build -o ../.claude/skills/ssh-ops/bin/ssh-mcp ./cmd/ssh-mcp/
```

### 生产服务器

`ssh-mcp` 是单一静态二进制，部署只需复制文件：

```bash
# 在构建机上
GOOS=linux GOARCH=amd64 go build -o ssh-mcp ./cmd/ssh-mcp/

# 复制到目标服务器
scp ssh-mcp user@server:/usr/local/bin/
ssh user@server chmod +x /usr/local/bin/ssh-mcp
```

### 与 CI/CD 集成

在 CI 环境中使用 `ssh-mcp` 时，注意凭证管理：

```yaml
# GitHub Actions 示例
- name: Deploy via SSH
  run: |
    ssh-mcp vault init
    ssh-mcp add --id prod --name "Production" \
      --host ${{ secrets.SSH_HOST }} \
      --user ${{ secrets.SSH_USER }} \
      --auth-type key \
      --key-path <(echo "${{ secrets.SSH_KEY }}")
    ssh-mcp exec --server prod --command "docker-compose up -d"
```

> 密钥应通过 CI secrets 注入，而非硬编码或存储在仓库中。

### 自定义配置目录

默认配置目录是 `~/.ssh-mcp/`。可通过环境变量覆盖：

```bash
export SSH_MCP_CONFIG_DIR=/opt/ssh-mcp-config
ssh-mcp vault init
```

这在容器环境或多用户共享主机上特别有用。

## 构建与测试

### 构建

```bash
cd go
go build -o ../.claude/skills/ssh-ops/bin/ssh-mcp ./cmd/ssh-mcp/
```

### 运行测试

```bash
cd go
go test ./... -v                   # 所有测试
go test ./internal/vault/... -v    # 仅 vault 测试
go test ./internal/ssh/... -v      # 仅 SSH 测试
go test ./internal/audit/... -v    # 仅审计测试
```

### 代码质量

```bash
cd go
go vet ./...                       # 静态分析
go build -o ../.claude/skills/ssh-ops/bin/ssh-mcp ./cmd/ssh-mcp/  # 编译验证
```

## 排错

### vault init 失败

**症状**：`ssh-mcp vault init` 报错 "vault already exists"

**原因**：`~/.ssh-mcp/` 已存在

**解决**：
```bash
# 如需重置（会丢失已有凭证！）
rm -rf ~/.ssh-mcp/
ssh-mcp vault init
```

### 连接超时

**症状**：`ssh-mcp test --server X` 超时

**排查步骤**：
```bash
# 1. 确认网络连通性
ping <host>

# 2. 确认 SSH 端口开放
nc -zv <host> 22

# 3. 直接用 ssh 测试（排除 ssh-mcp 问题）
ssh -p 22 <user>@<host> echo OK
```

### 认证失败

**症状**：`Permission denied` 或 `authentication failed`

**排查**：
```bash
# 检查配置中存储的认证信息
ssh-mcp list

# 重新添加服务器（密码/密钥路径可能有误）
ssh-mcp remove --id <server>
ssh-mcp add --id <server> ... --auth-type key --key-path ~/.ssh/id_rsa
```

### 找不到服务器

**症状**：`ssh-mcp exec --server X` 报错 "server not found"

**原因**：服务器 ID 输入错误或未添加

**解决**：
```bash
ssh-mcp list    # 查看所有已配置服务器的 ID
```

## 相关文档

- [`getting-started.md`](./getting-started.md) — 安装和首次配置
- [`security.md`](./security.md) — 安全模型和凭证管理
- [`cli-reference.md`](./cli-reference.md) — 完整命令参考
- 总索引：[`index.md`](./index.md)
