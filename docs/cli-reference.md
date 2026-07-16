---
title: CLI 命令参考
description: ssh-skill 全部 CLI 子命令、参数和用法参考
doc_type: reference
last_updated: 2026-07-15
audience: [所有开发者, 日常使用者]
---

# CLI 命令参考

`ssh-skill` 是单一二进制，通过子命令路由。所有命令以 `ssh-skill <subcommand> [flags]` 形式调用。

## 全局约定

- `--server`：引用已配置服务器的 ID（通过 `ssh-skill add --id` 指定）
- 所有服务器配置和凭证存储在 `~/.ssh-skill/`（可通过 `SSH_SKILL_CONFIG_DIR` 环境变量覆盖）
- 每个 `exec` 命令自动写入审计日志

## 命令总览

| 命令 | 用途 | 需要 Vault |
|------|------|-----------|
| `vault init` | 初始化加密保险库 | — |
| `add` | 添加服务器配置 | ✅ |
| `remove` | 删除服务器配置 | ✅ |
| `list` | 列出所有已配置服务器 | ✅ |
| `exec` | 在远程服务器执行命令 | ✅ |
| `upload` | 上传文件到远程服务器 | ✅ |
| `download` | 从远程服务器下载文件 | ✅ |
| `test` | 测试 SSH 连接 | ✅ |

---

## vault init

初始化加密保险库，创建 `~/.ssh-skill/` 目录和加密密钥。

```bash
ssh-skill vault init
```

**行为**：
- 幂等：如果 vault 已存在，重新生成空 servers 列表（保留旧密钥）
- 生成 32 字节随机 AES-256 密钥
- 创建空加密配置文件

> 配置目录可通过 `SSH_SKILL_CONFIG_DIR` 环境变量覆盖，**不接受命令行 flag**（这是有意设计，避免每个命令都要重复解析 config-dir flag）。

---

## add

添加服务器配置。密码经 AES-256-GCM 加密后存储。

```bash
ssh-skill add [flags]
```

**必需标志**：

| 标志 | 类型 | 说明 |
|------|------|------|
| `--id` | string | 服务器唯一标识（后续命令引用） |
| `--host` | string | 主机名或 IP 地址 |
| `--auth-type` | string | 认证方式：`password`、`key`、`agent` |

**可选标志**：

| 标志 | 类型 | 默认 | 说明 |
|------|------|------|------|
| `--name` | string | 空 | 人类可读名称 |
| `--port` | int | `22` | SSH 端口 |
| `--user` | string | `root` | SSH 登录用户名 |
| `--password` | string | — | 密码（auth-type=password 时必需，加密后存储） |
| `--key-path` | string | — | 私钥路径（auth-type=key 时使用，支持 `~` 展开） |

> 注：当前实现不支持 `--key-passphrase`、`--bastion-id`、`--config-dir` 等 flag（虽然代码中 `types.BastionConfig` 类型已定义，但 CLI 未暴露添加入口）。如需 bastion，需手动编辑 vault JSON 后重新加密。

---

## remove

删除服务器配置。

```bash
ssh-skill remove --id <server-id>
```

**必需标志**：

| 标志 | 类型 | 说明 |
|------|------|------|
| `--id` | string | 要删除的服务器 ID |

---

## list

列出所有已配置的服务器（不含密码）。

```bash
ssh-skill list [flags]
```

输出表格列：`ID`、`HOST`（`host:port`）、`NAME`、`AUTH`（认证类型；`key` 时可能显示 `key:<path>`）。

---

## exec

在远程服务器执行命令。

审批仅当经 Claude Code skill / Bash 权限策略调用时出现；裸 CLI 二进制直接运行时无审批。

```bash
ssh-skill exec --server <id> --command <cmd>
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

**退出码**（方案 B）：

| 进程退出码 | 含义 |
|-----------|------|
| `0` | 客户端成功完成 SSH 会话（**含**远程命令非零退出） |
| `1` | 客户端/配置/连接/会话建立失败 |

远程命令的退出码写入审计日志 `audit.log` 的 `exit_code` 字段，**不会**透传为进程退出码。不要依赖进程状态判断远程成功与否；也不要把 `255`/`-1` 当作本工具的进程退出码语义。

---

## upload

上传本地文件到远程服务器（SFTP）。

```bash
ssh-skill upload --server <id> --local <path> --remote <path> [flags]
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
ssh-skill download --server <id> --remote <path> --local <path> [flags]
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
ssh-skill test --server <id> [flags]
```

**必需标志**：

| 标志 | 类型 | 说明 |
|------|------|------|
| `--server` | string | 目标服务器 ID |

**输出**：成功时输出类似 `Connection to <host:port> OK — <N>ms` 的消息；失败时输出具体错误原因。


---

## 相关文档

- 总索引：[`index.md`](./index.md)
- [`getting-started.md`](./getting-started.md) — 首次安装和使用教程
- [`security.md`](./security.md) — 加密和凭证存储细节
