---
title: 项目架构
description: ssh-skill 的分层架构、内部包结构和数据流
doc_type: explanation
last_updated: 2026-07-16
audience: [维护者, 贡献者, AI Agent]
---

# 项目架构

本文档描述 `ssh-skill` 的分层架构、内部包职责和数据流。

## 分层结构

文档层名统一为：**UI → CLI → Service → Config → Types**（与包映射见下；不再单独强调 Repo 层）。

```
UI 层：Claude Code（AI agent）—— Bash / skill 调用 ssh-skill CLI
    │
CLI 层（Runtime）：ssh-skill 单一二进制（子命令路由）
    │  命令：list / add / remove / exec / upload / download / test / vault
    ▼
Service 层：
    ├── SSH 连接管理（connect/exec/upload/download）— internal/ssh
    ├── 凭证加密/解密（AES-256-GCM + Argon2id）— internal/vault
    └── 审计日志（JSONL 格式）— internal/audit
    │
Config 层：
    └── SSH_SKILL_CONFIG_DIR 环境变量解析，默认 ~/.ssh-skill/ — internal/config
    │
Types 层：
    └── ServerConfig、AuthConfig、AuthMethod、Vault、ExecResult、AuditEntry — internal/types
```

### 包与层映射

| 层 | 包 | 说明 |
|----|-----|------|
| CLI / Runtime | `internal/cli` | 子命令、路径装配、调用 service |
| Service | `internal/ssh`, `internal/vault`, `internal/audit` | 业务能力；路径由调用方传入 |
| Config | `internal/config` | 配置目录与文件路径解析 |
| Types | `internal/types` | 共享类型 |

## 依赖规则

采用 **Option A（config 为 Runtime 横切）**：

- **CLI 可直接 import `config`**，用于解析 `SSH_SKILL_CONFIG_DIR` 与 vault/audit 路径（横切初始化）。
- **Service 包（`ssh` / `vault` / `audit`）以路径或参数接收依赖**，不 import `config`。
- **“禁止跨层”** 在本项目中的含义是：CLI **不得** 直接 import `golang.org/x/crypto/ssh` 等底层库；应通过 `internal/ssh` 使用。CLI **可以** 使用 `internal/ssh`、`internal/vault`、`internal/audit`、`internal/config`、`internal/types`。
- 共享类型只能存放在 Types 层。
- 实践中的包依赖链：

```text
cli → {ssh, vault, audit, config, types}
ssh / vault / audit → types   （不 import config）
config → （仅标准库 / 环境）
```

## 目录结构

```
ssh-skill/
├── go/                           # Go 模块
│   ├── go.mod                    # 模块定义（module ssh-skill）
│   ├── go.sum                    # 依赖校验锁文件
│   ├── cmd/ssh-skill/
│   │   └── main.go               # 单一入口；非 0 错误时 os.Exit(1)
│   └── internal/
│       ├── types/
│       │   └── types.go          # 共享类型定义
│       ├── config/
│       │   └── config.go         # 配置路径解析
│       ├── vault/
│       │   ├── vault.go          # AES-256-GCM 加解密
│       │   ├── keygen.go         # EnsureKey / 随机密钥
│       │   └── storage.go        # 配置文件读写
│       ├── ssh/
│       │   ├── client.go         # SSH 连接管理（Client 包装 + bastion 生命周期）
│       │   ├── exec.go           # 远程命令执行（从 *ssh.ExitError 提取远程退出码到结果结构体）
│       │   └── transfer.go       # SFTP 文件传输（进度回调）
│       ├── audit/
│       │   └── audit.go          # JSONL 审计日志写入
│       └── cli/
│           ├── root.go           # 根命令定义、Load/Save
│           ├── add.go            # add 子命令
│           ├── remove.go         # remove 子命令
│           ├── list.go           # list 子命令
│           ├── exec.go           # exec 子命令（解密后 cfg 直传 ssh.Exec；写 audit）
│           ├── upload.go         # upload 子命令
│           ├── download.go       # download 子命令
│           ├── test.go           # test 子命令
│           ├── vault.go          # vault 子命令组
│           ├── helpers.go        # resolveServer 公共解密工具
│           ├── progress.go       # 传输进度条渲染
│           └── serve.go          # MCP 服务模式（未实现 stub）
├── .claude/skills/ssh-skill/       # Claude Code 技能定义
│   ├── SKILL.md                  # 技能描述、安全规则、工作流
│   └── bin/                      # 预编译二进制（已签入仓库分发）
├── scripts/                      # 跨平台构建脚本
├── docs/                         # 项目文档（你在读的）
└── .harness/                     # Harness CE 任务管理
    ├── knowledge/                # AI agent 知识库
    ├── agents/                   # Agent 行为规范
    ├── tasks/                    # 任务工作区
    └── templates/                # 文档模板
```

## 数据流

### 命令执行流程（exec）

```
用户 / AI Agent
    │
    │ ssh-skill exec --server X --command "uptime"
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
    │ session.Run() → 收集 stdout/stderr、远程 exit_code、duration
    │   errors.As(*ssh.ExitError) 将远程退出码写入 ExecResult.ExitCode
    │   远程非零退出对 CLI 返回 err=nil（会话本身成功完成）
    ▼
audit/audit.go
    │ 写入 JSONL 记录到 audit.log（含远程 ExitCode）
    ▼
stdout: "14:32:15 up 30 days, ..."
进程退出码：0（客户端成功完成会话，即使远程命令非零）
            1（连接/配置/客户端错误；main 在 Run 返回 error 时 os.Exit(1)）
```

**进程退出码 vs 远程退出码**（方案 B）：

| 情况 | 进程 `os.Exit` | 审计 `exit_code` |
|------|----------------|------------------|
| 远程命令成功 | 0 | 0 |
| 远程命令非零 | **0**（会话完成） | 远程真实码 |
| 连接/配置/客户端失败 | 1 | 可能为 -1 或未写完整会话 |

远程退出码**不会**透传到进程 `os.Exit`；只出现在 `ExecResult` / 审计字段中。

### 文件传输流程

```
ssh-skill upload --server X --local ./app.tar.gz --remote /tmp/
    │
    ▼
cli/upload.go → resolveServer() → 解密 cfg
    │   （download 对应 cli/download.go，同一模式）
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

> 传输入口是 `cli/upload.go` / `cli/download.go`，**不是** `cli/exec.go`。当前 upload/download **不写** audit。

## 外部依赖

| 依赖 | 用途 | 所属层 |
|------|------|--------|
| `golang.org/x/crypto` | SSH 客户端、AES-256-GCM、Argon2id | Service（经 `internal/ssh` / `internal/vault`） |
| `github.com/pkg/sftp` | SFTP 文件传输 | Service（`internal/ssh`） |
| `golang.org/x/sys` | `x/crypto` 的间接依赖 | 间接 |

详见 `.harness/knowledge/DEPENDENCIES.md`。

## 相关文档

- [`security.md`](./security.md) — 加密和凭证存储细节
- [`cli-reference.md`](./cli-reference.md) — CLI 命令参考
- [`.harness/knowledge/ARCHITECTURE.md`](../.harness/knowledge/ARCHITECTURE.md) — AI agent 遵守的架构约束
- [`.harness/knowledge/DEPENDENCIES.md`](../.harness/knowledge/DEPENDENCIES.md) — 依赖策略和升级流程
- 总索引：[`index.md`](./index.md)
