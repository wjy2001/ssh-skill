---
title: 项目架构
description: ssh-mcp 的分层架构、内部包结构和数据流
doc_type: explanation
last_updated: 2026-07-06
audience: [维护者, 贡献者, AI Agent]
---

# 项目架构

本文档描述 `ssh-mcp` 的分层架构、内部包职责和数据流。

## 分层结构

```
UI 层：Claude Code（AI agent）—— Bash 调用 ssh-mcp CLI
    │
CLI 层：ssh-mcp 单一二进制（子命令路由）
    │  命令：list / add / remove / exec / upload / download / test / vault / serve
    ▼
Service 层：
    ├── SSH 连接管理（connect/exec/upload/download）
    ├── 凭证加密/解密（AES-256-GCM + Argon2id）
    └── 审计日志（JSONL 格式）
    │
Repo 层：
    ├── 配置文件读写（~/.ssh-mcp/servers.json.age）
    ├── Vault key 管理（~/.ssh-mcp/.vault-key）
    └── SSH 客户端（golang.org/x/crypto/ssh）
    │
Config 层：
    └── SSH_MCP_CONFIG_DIR 环境变量解析，默认 ~/.ssh-mcp/
    │
Types 层：
    └── ServerConfig、AuthConfig、AuthMethod、Vault、ExecResult、AuditEntry
```

## 依赖规则

- 每一层只能依赖相邻的下一层
- 禁止跨层引用（如 CLI 直接引用 SSH 客户端库）
- 共享类型只能存放在 Types 层
- `internal` 包之间只能向下依赖：`cli → ssh/vault/audit → config → types`

## 目录结构

```
ssh-skill/
├── go/                           # Go 模块
│   ├── go.mod                    # 模块定义（module ssh-mcp）
│   ├── go.sum                    # 依赖校验锁文件
│   ├── cmd/ssh-mcp/
│   │   └── main.go               # 单一入口，子命令路由分发
│   └── internal/
│       ├── types/
│       │   └── types.go          # 共享类型定义
│       ├── config/
│       │   └── config.go         # 配置路径解析
│       ├── vault/
│       │   ├── vault.go          # AES-256-GCM 加解密
│       │   ├── keygen.go         # 密钥生成（Argon2id）
│       │   └── storage.go        # 配置文件读写
│       ├── ssh/
│       │   ├── client.go         # SSH 连接管理（多认证）
│       │   ├── exec.go           # 远程命令执行
│       │   └── transfer.go       # SFTP 文件传输
│       ├── audit/
│       │   └── audit.go          # JSONL 审计日志写入
│       └── cli/
│           ├── root.go           # 根命令定义
│           ├── add.go            # add 子命令
│           ├── remove.go         # remove 子命令
│           ├── list.go           # list 子命令
│           ├── exec.go           # exec 子命令
│           ├── upload.go         # upload 子命令
│           ├── download.go       # download 子命令
│           ├── test.go           # test 子命令
│           ├── vault.go          # vault 子命令组
│           └── serve.go          # MCP 服务模式
├── .claude/skills/ssh-ops/       # Claude Code 技能定义
│   └── SKILL.md                  # 技能描述、安全规则、工作流
├── docs/                         # 项目文档（你在读的）
├── bin/                          # 构建产物
└── .harness/                     # Harness CE 任务管理
    ├── knowledge/                # AI agent 知识库
    ├── agents/                   # Agent 行为规范
    ├── tasks/                    # 任务工作区
    └── templates/                # 文档模板
```

## 数据流

### 命令执行流程

```
用户 / AI Agent
    │
    │ ssh-mcp exec --server X --command "uptime"
    ▼
cli/exec.go
    │ 解析参数、从 vault 加载服务器配置
    ▼
vault/vault.go
    │ 读取 .vault-key → 解密 servers.json.age → 返回 ServerConfig
    ▼
ssh/client.go
    │ 根据 AuthConfig 建立 SSH 连接（password/key/agent）
    ▼
ssh/exec.go
    │ 执行命令 → 收集 stdout/stderr/exit_code/duration
    ▼
audit/audit.go
    │ 写入 JSONL 记录到 audit.log
    ▼
stdout: "14:32:15 up 30 days, ..."
exit code: 0
```

### 文件传输流程

```
ssh-mcp upload --server X --local ./app.tar.gz --remote /tmp/
    │
    ▼
ssh/client.go → SSH 连接
    │
    ▼
ssh/transfer.go → SFTP 客户端（github.com/pkg/sftp）
    │ 打开本地文件 → 创建远程文件 → 流式复制
    ▼
audit/audit.go → 记录传输操作
```

## 外部依赖

| 依赖 | 用途 | 所属层 |
|------|------|--------|
| `golang.org/x/crypto` | SSH 客户端、AES-256-GCM、Argon2id | Repo / Service |
| `github.com/pkg/sftp` | SFTP 文件传输 | Repo |
| `golang.org/x/sys` | `x/crypto` 的间接依赖 | Repo（间接） |

详见 `.harness/knowledge/DEPENDENCIES.md`。

## 相关文档

- [`security.md`](./security.md) — 加密和凭证存储细节
- [`cli-reference.md`](./cli-reference.md) — CLI 命令参考
- [`.harness/knowledge/ARCHITECTURE.md`](../.harness/knowledge/ARCHITECTURE.md) — AI agent 遵守的架构约束
- [`.harness/knowledge/DEPENDENCIES.md`](../.harness/knowledge/DEPENDENCIES.md) — 依赖策略和升级流程
- 总索引：[`index.md`](./index.md)
