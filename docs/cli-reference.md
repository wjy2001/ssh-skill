---
title: CLI 命令参考
description: ssh-mcp 全部 CLI 子命令、参数和用法参考
doc_type: reference
last_updated: 2026-07-06
audience: [所有开发者, 日常使用者]
---

# CLI 命令参考

`ssh-mcp` 是单一二进制，通过子命令路由。所有命令以 `ssh-mcp <subcommand> [flags]` 形式调用。

## 全局约定

- `--server` / `-s`：引用已配置服务器的 ID（通过 `ssh-mcp add --id` 指定）
- 所有服务器配置和凭证存储在 `~/.ssh-mcp/`（可通过 `SSH_MCP_CONFIG_DIR` 环境变量覆盖）
- 每个 `exec` 命令自动写入审计日志

## 命令总览

| 命令 | 用途 | 需要 Vault |
|------|------|-----------|
| `vault init` | 初始化加密保险库 | — |
| `vault status` | 查看 vault 状态 | ✅ |
| `add` | 添加服务器配置 | ✅ |
| `remove` | 删除服务器配置 | ✅ |
| `list` | 列出所有已配置服务器 | ✅ |
| `exec` | 在远程服务器执行命令 | ✅ |
| `upload` | 上传文件到远程服务器 | ✅ |
| `download` | 从远程服务器下载文件 | ✅ |
| `test` | 测试 SSH 连接 | ✅ |
| `serve` | 启动 MCP 服务模式（实验性） | ✅ |

---

## vault init

初始化加密保险库，创建 `~/.ssh-mcp/` 目录和加密密钥。

```bash
ssh-mcp vault init [flags]
```

**行为**：
- 如果 vault 已存在，提示错误（不会覆盖）
- 生成 32 字节随机 AES-256 密钥
- 创建空加密配置文件

**标志**：

| 标志 | 类型 | 默认 | 说明 |
|------|------|------|------|
| `--config-dir` | string | `~/.ssh-mcp/` | 自定义配置目录 |

---

## vault status

查看 vault 状态，包括密钥存在性、服务器数量等。

```bash
ssh-mcp vault status [flags]
```

---

## add

添加服务器配置。密码经 AES-256-GCM 加密后存储。

```bash
ssh-mcp add [flags]
```

**必需标志**：

| 标志 | 类型 | 说明 |
|------|------|------|
| `--id` | string | 服务器唯一标识（后续命令引用） |
| `--name` | string | 人类可读名称 |
| `--host` | string | 主机名或 IP 地址 |
| `--user` | string | SSH 登录用户名 |
| `--auth-type` | string | 认证方式：`password`、`key`、`agent` |

**可选标志**：

| 标志 | 类型 | 默认 | 说明 |
|------|------|------|------|
| `--port` | int | `22` | SSH 端口 |
| `--password` | string | — | 密码（auth-type=password 时必需） |
| `--key-path` | string | `~/.ssh/id_rsa` | 私钥路径（auth-type=key 时使用） |
| `--key-passphrase` | string | — | 私钥密码（如有） |
| `--bastion-id` | string | — | 跳板机服务器 ID |
| `--config-dir` | string | `~/.ssh-mcp/` | 自定义配置目录 |

---

## remove

删除服务器配置。

```bash
ssh-mcp remove --id <server-id> [flags]
```

**必需标志**：

| 标志 | 类型 | 说明 |
|------|------|------|
| `--id` | string | 要删除的服务器 ID |

---

## list

列出所有已配置的服务器（不含密码）。

```bash
ssh-mcp list [flags]
```

输出表格：ID、名称、主机、端口、用户、认证类型。

---

## exec

在远程服务器执行命令。**此命令会触发 Claude Code 审批对话框。**

```bash
ssh-mcp exec --server <id> --command <cmd> [flags]
```

**必需标志**：

| 标志 | 类型 | 说明 |
|------|------|------|
| `--server` | string | 目标服务器 ID |
| `--command` | string | 要执行的命令 |

**可选标志**：

| 标志 | 类型 | 默认 | 说明 |
|------|------|------|------|
| `--timeout` | int | `30` | 命令超时（秒） |
| `--config-dir` | string | `~/.ssh-mcp/` | 自定义配置目录 |

**退出码**：远程命令的退出码透传。`255` 表示连接或执行错误。

---

## upload

上传本地文件到远程服务器（SFTP）。

```bash
ssh-mcp upload --server <id> --local <path> --remote <path> [flags]
```

**必需标志**：

| 标志 | 类型 | 说明 |
|------|------|------|
| `--server` | string | 目标服务器 ID |
| `--local` | string | 本地文件路径 |
| `--remote` | string | 远程目标路径 |

---

## download

从远程服务器下载文件（SFTP）。

```bash
ssh-mcp download --server <id> --remote <path> --local <path> [flags]
```

**必需标志**：

| 标志 | 类型 | 说明 |
|------|------|------|
| `--server` | string | 目标服务器 ID |
| `--remote` | string | 远程文件路径 |
| `--local` | string | 本地目标路径 |

---

## test

测试与服务器的 SSH 连接。验证认证、加密和网络连通性。

```bash
ssh-mcp test --server <id> [flags]
```

**必需标志**：

| 标志 | 类型 | 说明 |
|------|------|------|
| `--server` | string | 目标服务器 ID |

**输出**：成功时输出 "OK"，失败时输出具体错误原因。

---

## serve

启动 MCP（Model Context Protocol）服务模式（实验性）。

```bash
ssh-mcp serve [flags]
```

以 MCP 协议对外暴露工具调用接口，允许 AI agent 通过标准协议调用 SSH 操作。

**可选标志**：

| 标志 | 类型 | 默认 | 说明 |
|------|------|------|------|
| `--config-dir` | string | `~/.ssh-mcp/` | 自定义配置目录 |

---

## 相关文档

- 总索引：[`index.md`](./index.md)
- [`getting-started.md`](./getting-started.md) — 首次安装和使用教程
- [`security.md`](./security.md) — 加密和凭证存储细节
