# 系统架构约束

## 分层结构

```text
Types -> Config -> Repo -> Service -> Runtime -> UI
```

### SSH 操作扩展分层

本项目的 ssh-mcp CLI 二进制遵循同一分层约束：

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
- 禁止跨层引用，例如 `UI` 直接引用 `Repo`
- 共享类型只能存放在 `Types`
- CLI 层（相当于 Runtime）依赖 Service 层，Service 依赖 Repo 和 Config
- internal 包之间只能向下依赖：cli → ssh/vault/audit → config → types

## 各层职责
### Types
- 定义共享数据结构、Schema、领域类型

### Config
- 管理配置读取、环境变量、静态开关

### Repo
- 管理外部资源访问，如数据库、文件、HTTP 客户端

### Service
- 组合 Repo 和领域逻辑，产出可复用业务能力

### Runtime
- 组装运行时流程、任务编排、依赖注入

### UI
- 承担展示、交互和用户输入输出

## 变更要求
1. 架构调整前先记录设计意图
2. 实施后同步更新本文档
3. 如果变更会影响审查规则，联动更新 `.harness/agents/REVIEWER.md`
