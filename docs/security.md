---
title: 安全模型
description: ssh-skill 的安全设计、威胁模型、加密方案和审计机制
doc_type: explanation
last_updated: 2026-07-15
audience: [所有开发者, 安全审计者]
---

# 安全模型

本文档描述 `ssh-skill` 的安全设计：它防什么、不防什么，以及加密和审计的实现方式。

## 威胁模型

### 防御范围

`ssh-skill` 设计用于防御**被动凭证泄露**：

- 凭证出现在聊天历史、日志文件或环境变量中
- 配置文件被非授权用户读取
- AI agent 在执行时无意中暴露凭证

### 不防御范围

以下场景超出设计范围：

- **主动 MITM 攻击**：当前实现使用 `ssh.InsecureIgnoreHostKey()`，**不验证远程主机密钥**。攻击者若能拦截网络流量，可冒充目标服务器。计划在后续版本中引入 known_hosts 持久化与首次信任模型。
- **主动攻击**：攻击者通过 Bash 命令直接读取 vault key 文件（`~/.ssh-skill/.vault-key`）
- **内存转储**：运行中的进程内存被 dump
- **内核级攻击**：rootkit、内核模块
- **物理访问**：攻击者拥有机器的物理访问权

> 如果你需要防御主动攻击或 MITM，应在操作系统层面加固（SELinux、文件完整性监控、主机密钥钉扎），而非依赖应用层加密。当前版本聚焦"被动凭证泄露"和"AI 幻觉连接未授权服务器"两类威胁。

## Vault 加密存储

### 存储结构

```
~/.ssh-skill/
├── .vault-key         # 32 字节随机 AES-256 密钥（权限 0600）
├── servers.json.age   # 加密的服务器配置（AES-256-GCM）
└── audit.log          # JSONL 审计日志（明文，权限 0600）
```

- 目录权限：`0700`
- 密钥文件和配置文件权限：`0600`
- 配置文件格式：JSON → AES-256-GCM 加密 → 二进制（`.age` 后缀仅命名约定）

### 加密方案

```
随机 32 字节 vault key（.vault-key，CSPRNG）
        │
        ▼  每次 Encrypt：Argon2id(vault key, 随机 salt)
派生密钥 ──► AES-256-GCM
        ├── Nonce：每次加密随机生成（12 字节），前置到密文
        ├── 认证标签：16 字节 GCM 认证标签
        └── Salt：每次加密随机，与密文一起存储
```

**要点**：
- Argon2id 作用在**随机 32 字节 vault key** 上，**不是**用户登录口令的 KDF。
- 审计日志为明文，**不**经 Argon2id / AES 加密。
- 参数：`time=3, memory=64MB, threads=4`（对应 `vault.go` 中 `argonTime=3, argonMemory=64*1024, argonThreads=4`）。

**密钥生命周期**：
1. 首次 `vault init`：`crypto/rand` 生成 32 字节密钥 → 写入 `.vault-key`
2. 每次加密 `servers.json.age`：读取 vault key → Argon2id(key, 随机 salt) 派生 → AES-256-GCM
3. 每次解密：从密文取 salt → 同样派生 → AES-256-GCM 解密
4. 没有密钥轮换机制（按需可手动删除 `.vault-key` 和 `servers.json.age` 重建）

## 目标校验

`ssh-skill` 的核心安全机制之一：**AI agent 只能连接预配置的服务器**。

### 校验流程

```
AI agent 发起: ssh-skill exec --server X --command "rm -rf /"
                    │
                    ▼
           查找 X 是否在 servers.json.age 中
                    │
          ┌─────────┴─────────┐
          ▼                   ▼
       找到了                未找到
          │                   │
          ▼                   ▼
     解密凭证 → SSH 连接   拒绝执行 + 记录审计
```

**为什么重要**：防止 AI 幻觉出新主机名并发起连接，也防止 prompt injection 试图让 agent 连接未授权服务器。

## 审计日志

**当前仅 `exec` 写入审计日志**。`upload` / `download` / `test` / `add` / `remove` / `list` 等命令**不**写 audit。

`exec` 成功或失败后，记录追加到 `~/.ssh-skill/audit.log`（JSONL 格式，每行一条记录）。

### 日志条目结构

```json
{
  "timestamp": "2026-07-06T15:30:00Z",
  "server_id": "my-server",
  "server_host": "10.0.0.1:22",
  "command": "uptime",
  "exit_code": 0,
  "stdout_len": 41,
  "stderr_len": 0,
  "duration_ms": 1234
}
```

字段对应 `go/internal/types/types.go` 中的 `AuditEntry` 结构体。`stdout_len`/`stderr_len` 记录长度而非完整内容，避免审计日志膨胀；如需完整输出，由调用方另行收集。

**特性**：
- 追加写入，不覆盖历史记录
- 明文存储（审计日志不加密，便于 `grep` 和外部工具消费）
- 并发安全（多个 `ssh-skill exec` 可同时写入）
- 日志文件无自动轮换或大小限制（可通过外部 logrotate 管理）

## 凭证处理规则

| 场景 | 行为 |
|------|------|
| `ssh-skill add --password <pwd>` | 密码经 vault 加密后存入 `servers.json.age`。`--password` 出现在 **argv** 上，**会**进入 shell 历史与进程列表（`ps`）；工具不阻止该泄露，由用户自行选择交互输入或密钥认证 |
| `ssh-skill list` | 列：`ID`、`HOST`（`host:port`）、`NAME`、`AUTH` — **不显示密码** |
| 环境变量 | **不**从环境变量读取密码（防止 `.env` 文件或 CI 变量泄露） |
| 进程内存 | 解密后明文在进程生命周期内存在于内存；**无**显式内存清零（scrub）；进程退出后由 OS 回收，**不**保证「立即释放」 |

## 安全检查清单

- [ ] `~/.ssh-skill/` 权限为 0700
- [ ] `~/.ssh-skill/.vault-key` 权限为 0600
- [ ] `~/.ssh-skill/servers.json.age` 权限为 0600
- [ ] 不在版本控制中提交 `.vault-key` 或 `servers.json.age`
- [ ] 定期检查 `audit.log` 中是否有异常命令
- [ ] 定期备份 `.vault-key`（丢失后无法解密已存储凭证）

## 相关文档

- [`cli-reference.md`](./cli-reference.md) — 命令参数参考
- [`getting-started.md`](./getting-started.md) — 首次初始化 vault
- 总索引：[`index.md`](./index.md)
