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
│       │   ├── client.go           # SSH 连接管理（Client 包装类型 + bastion 生命周期）
│       │   ├── exec.go             # 远程命令执行（ExitCode 透传 via *ssh.ExitError）
│       │   └── transfer.go         # SFTP 文件传输（进度回调）
│       ├── audit/
│       │   └── audit.go          # JSONL 审计日志写入
│       └── cli/
│           ├── root.go           # 根命令定义
│           ├── add.go            # add 子命令
│           ├── remove.go         # remove 子命令
│           ├── list.go           # list 子命令
│           ├── exec.go           # exec 子命令（解密后 cfg 直传 ssh.Exec）
│           ├── upload.go         # upload 子命令
│           ├── download.go       # download 子命令
│           ├── test.go           # test 子命令
│           ├── vault.go          # vault 子命令组
│           ├── helpers.go        # resolveServer 公共解密工具
│           ├── progress.go       # 传输进度条渲染
│           └── serve.go          # MCP 服务模式（未实现）
├── .claude/skills/ssh-ops/       # Claude Code 技能定义
│   ├── SKILL.md                  # 技能描述、安全规则、工作流
│   └── bin/                      # 编译产物
├── scripts/                      # 跨平台构建脚本
├── docs/                         # 项目文档（你在读的）
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
    │ resolveServer() → 从 vault 加载服务器 + AES-256-GCM 解密密码（in-memory plaintext）
    ▼
ssh/exec.go
    │ 接收已解密的 cfg，不再二次查找 vault
    ▼
ssh/client.go
    │ Connect() → 根据 AuthConfig 建立 SSH 连接（password/key/agent）
    │   返回 *Client 包装类型，持有 bastion 引用以管理生命周期
    ▼
ssh/exec.go
    │ session.Run() → 收集 stdout/stderr/exit_code/duration
    │   errors.As(*ssh.ExitError) 提取真实退出码
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
cli/exec.go → resolveServer() → 解密 cfg
    │
    ▼
ssh/client.go → Connect() 返回 *Client
    │
    ▼
ssh/transfer.go → SFTP 客户端（github.com/pkg/sftp）
    │ progressReader 包装 io.Reader，每 100ms 触发回调
    │ io.CopyBuffer + 256KB 缓冲
    ▼
cli/progress.go → 渲染进度条到 stderr
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
