# 系统架构约束

## 分层结构

```text
UI → CLI (Runtime) → Service(ssh/vault/audit) → Config → Types
```

说明：Config 是 Runtime 横切关注点，**不是**严格夹在 Service 与 Types 之间的唯一下一层。Service 包（ssh/vault/audit）实际只依赖 Types 与标准库/外部加密库，路径与密钥由 CLI 注入。

### SSH 操作扩展分层

本项目的 ssh-skill CLI 二进制分层如下：

```
UI 层：Claude Code（AI agent）—— Bash 调用 ssh-skill CLI
    │
CLI / Runtime 层：ssh-skill 单一二进制（子命令路由、编排）
    │  命令：list / add / remove / exec / upload / download / test / vault
    │  允许直接 import：config、vault、audit、ssh、types
    ▼
Service 层：
    ├── ssh：连接 / exec / upload / download（import types；使用 golang.org/x/crypto/ssh）
    ├── vault：AES-256-GCM + Argon2id 加解密、密钥与 vault 文件 I/O（import types）
    └── audit：JSONL 审计追加（import types）
    │  不 import config；路径/密钥由 CLI 传入
    ▼
Config 层：
    └── SSH_SKILL_CONFIG_DIR 解析，默认 ~/.ssh-skill/（stdlib only）
    │
Types 层：
    └── ServerConfig、AuthConfig、AuthMethod、Vault、ExecResult、AuditEntry
```

## 依赖规则（以实际代码为准）

**允许的 internal 依赖方向：**

| 包 | 可 import |
|----|-----------|
| `cli` | `config`、`vault`、`audit`、`ssh`、`types` |
| `ssh` | `types`；外部 `golang.org/x/crypto/ssh`、`github.com/pkg/sftp` |
| `vault` | `types`；外部 `golang.org/x/crypto/argon2` |
| `audit` | `types` |
| `config` | 仅标准库 |
| `types` | 叶节点，不依赖本仓库其他 internal 包 |

**禁止：**

- `cli` **不得** import `golang.org/x/crypto/ssh`（SSH 协议细节只在 `internal/ssh`）
- `ssh` / `vault` / `audit` **不得** import `config`（配置目录解析留在 CLI）
- 跨层逆向依赖（例如 `types` → `cli`、`config` → `vault`）
- 共享领域类型放在 Types 以外的包

**与旧 aspirational 规则的关系：**

历史表述 `cli → ssh/vault/audit → config → types`（“每层只依赖相邻下一层”）是理想化分层，**与当前代码不符**。实际采用 **Option A**：

- CLI/Runtime **可以**直接 import `config`
- Service 依赖 Types（及文件系统路径参数），**不**依赖 config 包
- 层名统一：`UI → CLI → Service(ssh/vault/audit) → Config → Types`（Config 为横切，由 CLI 消费）

## 各层职责

### Types
- 定义共享数据结构、Schema、领域类型

### Config
- 管理配置目录解析、环境变量、路径派生（不持有业务状态）

### Service（ssh / vault / audit）
- SSH 连接与传输；凭证加解密与 vault 持久化；审计写入
- 接收 CLI 传入的路径、密钥、领域对象，不自己解析 `SSH_SKILL_CONFIG_DIR`

### CLI / Runtime
- 子命令路由、flag 解析、组装 config 路径与 vault key、调用 Service、用户可见 I/O
- 在连接前解密密码字段（见 `resolveServer`）

### UI
- Claude Code / 用户终端：展示与审批（审批仅在 skill/受控 Bash 路径；裸二进制无审批）

## 变更要求
1. 架构调整前先记录设计意图
2. 实施后同步更新本文档
3. 如果变更会影响审查规则，联动更新 `.harness/agents/REVIEWER.md`
