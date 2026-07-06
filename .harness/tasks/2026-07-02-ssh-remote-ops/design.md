# 远程 SSH 服务器操作系统 — 设计文档

## 任务标识

- **slug**：`ssh-remote-ops`
- **日期**：2026-07-02
- **来源请求**：用户希望 AI 能直接操作远程服务器，通过 SSH 连接执行命令。需要 Go 二进制做命令转发防止误传，支持多种认证方式，服务器配置本地持久化。

## 快速状态

- **核心结论**：CLI-first 架构——Go 二进制独立可执行（`ssh-mcp exec/list/add/remove/vault`），Skill 通过 Bash 调用。审批模式保护命令安全，AES-256-GCM 加密 JSON 文件存储凭证，随机密钥自动生成落盘。
- **关键决策**：D1～D5 全部关闭，3 个开放问题全部关闭。
- **流程状态**：`方向已确认` — 进入派生生成阶段。
- **详细状态**：见 `## 设计状态`。

---

## 设计状态

### 基线

- **设计基线**：sha256:8dc437c862725de6fedbf36e2bf499dcb18899b6a372d10e25fc44fe4bcfea90
- **取代基线**：sha256:d51afb5fd944a37d982d861049a1b1863a0e673255cb4e02f59a5f99916273d6（术语修正：入口命名 MCP→CLI、降级分析表去主密码）

### 流程状态

- **状态**：已审核
- **待讨论项**：无
- **阻塞项**：无

### 派生状态

- **implementation.md**：草案（有效，绑定设计基线 sha256:d51afb5f）
- **PROGRESS.md**：有效
- **实施计划基线**：sha256:54082559abcc9a716c83138fc4c52342cfc92dc2c752bdb1898a7e49b7f15379
- **执行证据基线**：无
- **有效审查报告**：reviews/001-derivation-generation.md

### 审批状态

- **审批对象**：design.md + implementation.md + PROGRESS.md + 当前有效 reviews 摘要
- **实施计划基线**：sha256:54082559abcc9a716c83138fc4c52342cfc92dc2c752bdb1898a7e49b7f15379
- **执行证据基线**：无
- **审批来源**：用户（对话确认）
- **审批时间**：2026-07-02 18:35

---

## 背景和问题

用户当前通过终端手动 SSH 到服务器执行运维命令。用户希望 Claude Code（AI）能够：

1. 通过 SSH 直接操作远程服务器（执行命令、上传/下载文件）
2. 使用 Go 编译的二进制程序作为中间层，拦截和转发命令——防止 AI 把命令发错服务器
3. 支持多种 SSH 登录方式（密钥、密码、凭证文件）
4. 服务器连接配置只需配置一次，之后可直接复用
5. 所有敏感配置（IP、账号、密钥路径）只存储在本地，绝不外传

本质问题：**如何在 AI agent 和远程服务器之间建立一个安全、可审计、支持持久化配置的命令通道。**

---

## 问题定义审计

### 目标

1. AI 能通过 SSH 在指定服务器上执行 shell 命令并获取输出
2. AI 能上传/下载文件到指定服务器
3. 命令在执行前经过 Go 中间层校验（目标服务器匹配、命令白名单/黑名单）
4. 服务器连接信息本地持久化，支持增删改查
5. 支持密码、SSH 密钥、SSH agent 三种认证

### 成功指标

| # | 指标 | 验证方式 |
|---|------|---------|
| S1 | AI 能够 `ssh-exec --server myserver "ls -la"` 并得到正确输出 | 手动端到端测试 |
| S2 | 发往错误服务器的命令被 Go 中间层拦截并拒绝 | 单元测试 + 集成测试 |
| S3 | 服务器配置在重启 Claude Code 后仍然可用 | 持久化测试 |
| S4 | 配置文件不包含明文密码（密码需加密存储） | 静态分析配置文件 |
| S5 | 支持 ED25519/RSA 密钥、密码、agent 三种认证 | 集成测试矩阵 |

### 约束

- C1：Go 二进制必须能编译为 Windows/Linux/macOS 三个平台的独立可执行文件
- C2：配置文件只存储在用户本地文件系统，不上传任何远程服务
- C3：密码必须加密存储（AES-256-GCM + Argon2id 密钥派生）
- C4：命令转发层必须在命令执行前校验目标服务器身份
- C5：不依赖外部数据库或网络服务

### 根因假设

- 用户的操作对象是 Linux 服务器（SSH 服务端为标准 OpenSSH）
- 用户的主要使用场景是 Claude Code（本地 CLI），而非 Web UI
- 服务器数量在 10～100 台量级，不需要企业级集群管理

---

## 事实、假设和待确认问题

### 事实（有证据支持）

1. **Claude Code 支持 MCP 服务器**：通过 stdio 传输协议，Claude Code 可以启动 MCP 服务器进程并调用其工具。官方 Go SDK（`modelcontextprotocol/go-sdk` v1.6.x）成熟可用。[来源：pkg.go.dev 文档]
2. **Go SSH 标准库成熟**：`golang.org/x/crypto/ssh` 原生支持 password、publickey、keyboard-interactive、ssh-agent 四种认证方式。[来源：Go 标准库文档]
3. **MCP 与 Skill 互补**：MCP 是"连接层"（外部能力），Skill 是"知识层"（操作流程）。MCP 更适合管理持久连接和暴露工具，Skill 更适合定义安全策略和操作工作流。[来源：Claude Code 文档 + skywork.ai 对比分析]
4. **stdio 传输延迟 < 5ms**：本地 MCP 服务器通过 stdin/stdout 通信，延迟极低，适合命令转发场景。[来源：Claude Code MCP 文档]
5. **AES-256-GCM + Argon2id 是凭证加密的行业标准**：多个 Go SSH 管理工具（sshman、sectool）采用此方案。[来源：pkg.go.dev 多个项目]
6. **Claude Code MCP 工具定义支持 JSON Schema**：Go SDK 通过 struct tag `jsonschema:"..."` 自动生成工具参数 schema，支持强类型参数验证。[来源：go-sdk README]

### 假设（中置信度，已确认）

1. 用户接受在本地编译/下载一个 Go 二进制文件（约 10-20MB）
2. 用户不要求实时交互式终端（如 vim、top），只需要执行命令并获取输出
3. 密钥由 Go 运行时自动生成并落盘（`~/.ssh-mcp/.vault-key`），用户无需输入主密码
4. 威胁模型：防被动泄露（env/聊天记录），不防 AI 主动 `cat` 密钥文件

### 待确认问题（全部已关闭 ✅）

1. **命令安全策略**→ 审批模式（D2 已决定）
2. **多服务器并行**→ v1 只做单服务器，字段 `string`。不做 `--servers s1,s2,s3` 并行。
3. **审计日志**→ 做。JSONL 格式，`~/.ssh-mcp/audit.log`，记录时间、服务器、命令、退出码、耗时。
4. **密钥管理**→ Go 首次运行自动生成 32 字节随机密钥，落盘 `~/.ssh-mcp/.vault-key`（chmod 0600）。不用环境变量、不用交互输入、不用主密码。

---

## 待讨论项

### D1：架构方案 — 纯 MCP vs 纯 Skill vs MCP+Skill 混合（已关闭 ✅）

**最终选择**：CLI-first（方案 B）。Go 二进制作为独立 CLI 工具，Skill 通过 Bash 调用子命令。同时保留 `ssh-mcp serve` 作为 MCP 模式的可选入口。

**背景**：需要在 Claude Code 中暴露 SSH 操作能力。

**选项**：

| | 纯 Skill | 纯 MCP | MCP + Skill 混合 ⭐推荐 |
|---|---|---|---|
| **实现方式** | Skill 调用本地 ssh 命令或 Go 二进制 | Go MCP 服务端暴露工具 | MCP 提供连接/工具，Skill 提供安全策略/工作流 |
| **连接持久化** | ❌ 每次调用重新连接 | ✅ 服务端维护连接池 | ✅ 同纯 MCP |
| **工具暴露** | 通过 Bash 调用二进制 | 通过 MCP 协议暴露为工具 | ✅ 同纯 MCP |
| **安全策略** | 在 Skill 的 SKILL.md 中描述 | 在 Go 二进制中硬编码 | Skill 定义策略，MCP 执行 |
| **凭证管理** | 需要 Skill 中处理 | MCP 服务端管理 | ✅ 同纯 MCP |
| **复杂度** | 低 — 一个 SKILL.md + 一个 Go 二进制 | 中 — 需要 MCP 服务端开发 | 中高 — 两部分都需要 |
| **可测试性** | Go 二进制可独立测试 | MCP 服务端可独立测试 | ✅ 两者都可独立测试 |
| **灵活性** | 低 — 难以扩展 | 中 — 工具可增减 | 高 — 策略和工具独立演进 |
| **推荐理由** | — | — | MCP 管理"连接和操作"，Skill 管理"如何安全使用"；两者职责清晰，独立演进 |

**推荐**：MCP + Skill 混合。MCP 服务端（Go 二进制）作为持久化进程管理 SSH 连接池、凭证存储和工具暴露；Skill 作为操作手册定义安全策略、命令审批规则和使用模式。

**代价**：需要同时开发 Go MCP 服务端和 Skill 定义文件，工作量约为纯 MCP 的 1.3 倍。

**待用户选择**：[ ] 纯 MCP / [ ] 纯 Skill / [ ] MCP+Skill 混合

---

### D2：命令安全模型 — 白名单 vs 黑名单 vs 审批模式（已关闭 ✅）

**最终选择**：审批模式。所有命令执行前由用户在 Claude Code 的权限弹窗中确认，不做自动化的黑名单/白名单过滤。

**背景**：用户要求"防止命令被误传到错误的服务器"。这涉及两层安全：

1. **目标校验**：确保 AI 指定的目标服务器确实存在且已配置（防止 AI 幻觉出服务器名）
2. **命令校验**：确保命令内容安全（防止危险命令）

**选项**：

| | 白名单模式 | 黑名单模式 | 审批模式 ⭐推荐 |
|---|---|---|---|
| **机制** | 只允许执行预定义的命令列表 | 禁止匹配危险模式（如 `rm -rf`）的命令 | 所有命令执行前需要用户确认 |
| **安全性** | 高 — 攻击面极小 | 低 — 无法穷举危险模式 | 最高 — 人在回路 |
| **可用性** | 低 — 每次新命令需更新白名单 | 高 — 大部分命令可直接执行 | 中 — 每次需要点击确认 |
| **误拦率** | 高 — 合法命令常被拦 | 低 — 但危险命令可能漏网 | 无 — 用户自己判断 |
| **实现** | 预设命令列表 + 参数正则 | 危险模式正则匹配 | 利用 MCP 的 `permission` 机制 |

**推荐**：**分层防护**——第一层：目标服务器必须存在于本地配置中（硬校验，Go 中间层执行）；第二层：默认审批模式（利用 Claude Code MCP 的权限确认机制），用户可配置常用命令为"允许"以减少审批频率；第三层：可选的命令模式黑名单（如阻止 `rm -rf /`、`mkfs`）作为最后防线。

**待用户选择**：[ ] 纯白名单 / [ ] 纯黑名单 / [ ] 审批模式 / [ ] 分层防护（推荐）

---

### D3：Go 项目结构 — 单一二进制 vs 客户端-服务端分离（已关闭 ✅）

**最终选择**：单一二进制。一个 `ssh-mcp` 文件包含所有功能，通过子命令区分模式。

**背景**：Go 二进制需要作为 MCP 服务端运行。

**选项**：

| | 单一二进制 ⭐推荐 | 客户端-服务端分离 |
|---|---|---|
| **结构** | 一个 `ssh-mcp` 二进制 | `ssh-mcp-server`（守护进程）+ `ssh-mcp`（CLI 客户端） |
| **部署** | 简单 — 只需一个文件 | 复杂 — 需要管理守护进程生命周期 |
| **MCP 集成** | 直接通过 stdio 运行 | 需要先启动 server，再配置 HTTP transport |
| **连接复用** | Claude Code 生命周期内自然复用 | 跨 Claude Code 会话可复用 |
| **复杂度** | 低 | 高 |
| **维护** | 简单 | 复杂 |

**推荐**：单一二进制 stdio 模式。Claude Code 启动时自动拉起 MCP 服务端，退出时自动回收。MCP 服务端在内存中维护 SSH 连接池，实现会话内连接复用（`ssh-mcp` 启动后到 Claude Code 退出前）。如果将来需要跨会话连接复用，可以在不改变 Go 代码结构的情况下切换为 HTTP transport。

**待用户选择**：[ ] 单一二进制 stdio / [ ] 客户端-服务端分离

---

### D4：凭证存储方案 — 加密文件 vs 系统密钥链 vs SSH config 复用（已关闭 ✅）

**最终选择**：加密 JSON 文件。`~/.ssh-mcp/servers.json.age`，AES-256-GCM 加密，密钥由 Go 在首次运行时自动生成并落盘到 `~/.ssh-mcp/.vault-key`。

**背景**：需要本地存储服务器连接信息（主机、端口、用户、认证方式）。

**选项**：

| | 加密 JSON 文件 ⭐推荐 | 系统密钥链 | 复用 ~/.ssh/config |
|---|---|---|---|
| **密码存储** | AES-256-GCM 加密 | 由 OS 密钥链保证 | 不支持（ssh config 不存密码） |
| **跨平台** | ✅ 天然跨平台 | ❌ Windows/macOS/Linux API 不同 | ✅ 但功能受限 |
| **复杂度** | 中 — 需要实现加解密 | 高 — 三个平台三套 API | 低 — 直接读取 |
| **安全性** | 高 — 主密码 + Argon2id + AES-GCM | 最高 — OS 级保护 | 中 — 密钥路径明文 |
| **可移植性** | 高 — 一个加密文件 | 低 — 绑定 OS | 高 |
| **Go 依赖** | `golang.org/x/crypto`（已有） | 需要 `keyring` 第三方库 | 无额外依赖 |

**推荐**：加密 JSON 文件（主方案）+ `~/.ssh/config` 读取（辅助）。主配置文件 `~/.ssh-mcp/servers.json.age`（AES-256-GCM 加密），存储完整连接信息（含加密后的密码）。同时支持读取 `~/.ssh/config` 中的 Host 定义作为快速导入源（仅导入 Host、Hostname、User、Port、IdentityFile，不含密码）。

文件格式：
```
[16B salt][12B nonce][AES-256-GCM 加密的 JSON payload]
```

JSON payload 结构：
```json
{
  "version": 1,
  "servers": [
    {
      "id": "prod-web-01",
      "name": "生产 Web 服务器",
      "host": "10.0.1.100",
      "port": 22,
      "user": "deploy",
      "auth": {
        "type": "password",
        "encrypted_password": "base64..."
      }
    },
    {
      "id": "staging-db",
      "name": "预发布数据库",
      "host": "10.0.2.50",
      "port": 22,
      "user": "admin",
      "auth": {
        "type": "key",
        "key_path": "~/.ssh/staging_rsa",
        "passphrase_encrypted": "base64..."
      }
    }
  ]
}
```

**待用户选择**：[ ] 加密 JSON 文件 / [ ] 系统密钥链 / [ ] 复用 ~/.ssh/config / [ ] 组合方案（推荐）

---

### D5：项目目录和命名（已关闭 ✅）

**最终选择**：子目录隔离。Go 代码放在 `go/` 子目录中，与 `.claude/skills/`、`.harness/` 平级。

**背景**：需要确定 Go 项目根目录位置和 Skill 目录命名。

**选项**：

| | 项目根即 Go 模块根 ⭐推荐 | `go/` 子目录放 Go 代码 | `cmd/` 下单一入口 |
|---|---|---|---|
| **结构** | `D:\project\github\ssh-skill/` 下直接放 `go.mod`、`main.go` | Go 代码放 `go/` 子目录 | 标准 Go 项目布局 |
| **skill 位置** | `.claude/skills/ssh-ops/SKILL.md` | 同左 | 同左 |
| **推荐理由** | 项目本身就是一个 Go 模块 + Skill 定义，不需要额外嵌套 | — | 这是 Go 社区标准布局 |

**最终选择**：子目录隔离。目录结构如下：

```
ssh-skill/                        # 项目根
├── go/                           # Go 模块根
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── ssh-mcp/
│   │       └── main.go           # 入口
│   └── internal/
│       ├── vault/                # 凭证加密/解密 + 密钥管理
│       ├── ssh/                  # SSH 连接/执行/文件传输
│       ├── config/               # 服务器配置管理（CRUD）
│       └── audit/                # 审计日志写入
├── .claude/
│   └── skills/
│       └── ssh-ops/
│           └── SKILL.md          # Skill 定义
├── bin/                          # 编译产物（.gitignore）
├── README.md
├── AGENTS.md
├── CLAUDE.md
└── .harness/
```

---

## 残余风险

1. **vault key 文件被 AI 读取**：AI 有 Bash 权限，可以 `cat ~/.ssh-mcp/.vault-key`。缓解：威胁模型已明确——防被动泄露不防主动攻击。如需要可后续加文件访问审计。
2. **配置文件丢失导致无法连接任何服务器**：缓解：提示用户备份 `~/.ssh-mcp/` 目录。
3. **Windows 平台路径问题**：缓解：Go 的 `os.UserHomeDir()` + `filepath.Join` 自动处理。

---

## 候选解释和反证

### 为什么不直接用 Claude Code 的 Bash 工具 + `ssh` 命令？

- **问题**：Claude Code 可以在 Bash 中直接执行 `ssh user@host "command"`。为什么需要 Go 中间层？
- **反证**：
  1. 无目标校验：AI 可能把命令发到错误的 IP（幻觉）
  2. 无法管理凭证：密码在命令中明文传递或在聊天历史中暴露
  3. 无连接复用：每次 Bash 调用都是全新 SSH 连接
  4. 无审计能力：无法追踪 AI 执行了什么命令
- **结论**：Go 中间层提供了 Bash 直接调用无法实现的**目标校验、凭证隔离、连接复用和审计能力**。

### 为什么不使用 Ansible/SSH 工具？

- **问题**：已有成熟的运维工具，为什么重复造轮子？
- **反证**：Ansible 面向人类运维工程师设计，不面向 AI agent——它没有 MCP 接口、没有 command forwarding safety layer、也没有针对 AI 幻觉的防护机制。本方案解决的是"AI → 服务器"这个独特链路的问题，而非替代人类运维工具。
- **结论**：本方案是 AI agent 和服务器之间的适配层，与 Ansible 等工具处于不同的抽象层级。

---

## 勘查结论

### 技术栈确认

| 组件 | 选型 | 依据 |
|------|------|------|
| SSH 库 | `golang.org/x/crypto/ssh` | Go 官方维护，支持全部认证方式 |
| MCP SDK | `github.com/modelcontextprotocol/go-sdk` v1.6.x | 官方 Go SDK，支持 stdio 传输 |
| 加密 | AES-256-GCM + Argon2id（`golang.org/x/crypto`） | 行业标准，无第三方依赖 |
| 配置文件 | 加密 JSON（自定义格式） | 最小依赖原则 |
| Skill 引擎 | Claude Code 内置 Skill 系统 | 项目已有 Skill 基础设施 |

### 架构分层（遵循项目分层约束）

```
UI 层：Claude Code（AI agent）
    │  Bash 调用 ssh-mcp CLI
    │  Skill 文件定义用法和安全规则
    ▼
CLI 层：ssh-mcp 单一二进制（用户可直接在终端使用）
    │  子命令：exec / upload / download / list / add / remove / test / vault / serve
    │  启动时自动加载 vault key，解密 servers.json.age
Service 层：
    ├── SSH 连接管理（connect/exec/upload/download）
    ├── 凭证加密/解密（AES-256-GCM，密钥从 ~/.ssh-mcp/.vault-key 读取）
    └── 审计日志（JSONL，~/.ssh-mcp/audit.log）
Repo 层：
    ├── 配置文件读写（~/.ssh-mcp/servers.json.age）
    ├── Vault key 管理（~/.ssh-mcp/.vault-key，自动生成）
    └── SSH 客户端（golang.org/x/crypto/ssh）
Config 层：
    └── 环境变量 SSH_MCP_CONFIG_DIR（覆盖默认 ~/.ssh-mcp/）
Types 层：
    └── ServerConfig、AuthMethod、ExecResult、AuditEntry 等共享类型
```

### 证据范围和置信度

| 结论 | 置信度 | 证据 |
|------|--------|------|
| MCP 官方 Go SDK 成熟可用 | 高 | pkg.go.dev 文档 + v1.6.x 发布记录 |
| `x/crypto/ssh` 支持全部认证方式 | 高 | Go 标准库文档 |
| MCP+Skill 混合架构是最佳方案 | 高 | Claude Code 官方文档 + skywork.ai 对比分析 |
| AES-256-GCM+Argon2id 是安全的凭证存储方案 | 高 | 多个 Go SSH 工具项目的实践 |
| stdio 传输适合本地 MCP 服务端 | 高 | Claude Code 官方推荐 stdio 用于本地服务端 |
| 设计基线一致性要求来自项目规范 | 高 | AGENTS.md + HARNSESS CE 规范 |

---

## 决策

```
【判断】✅ 值得做
【洞察】AI agent 和远程服务器之间存在一个独特的安全缺口——没有标准化的、
       带安全校验的命令通道。Bash 直接 SSH 无法防止 AI 幻觉导致的误操作。
【方案】CLI-first 单一 Go 二进制（ssh-mcp），子命令模式暴露 exec/list/add/remove/vault。
       Skill 通过 Bash 调用 CLI。凭证 AES-256-GCM 加密存储，随机密钥自动生成落盘。
       审批模式保护命令安全，JSONL 审计日志可追溯。
```

---

## 数据结构

### 核心类型

```go
// Types 层：共享类型定义

// AuthMethod 认证方式
type AuthMethod string
const (
    AuthPassword           AuthMethod = "password"
    AuthKey                AuthMethod = "key"
    AuthAgent              AuthMethod = "agent"
)

// ServerConfig 服务器连接配置（持久化实体）
type ServerConfig struct {
    ID       string     `json:"id"`        // 唯一标识，如 "prod-web-01"
    Name     string     `json:"name"`      // 可读名称
    Host     string     `json:"host"`      // IP 或域名
    Port     int        `json:"port"`      // 默认 22
    User     string     `json:"user"`      // SSH 用户名
    Auth     AuthConfig `json:"auth"`      // 认证配置
    Bastion  *BastionConfig `json:"bastion,omitempty"` // 跳板机（可选）
    Tags     []string   `json:"tags,omitempty"` // 分组标签
}

// AuthConfig 认证配置（只有一种生效）
type AuthConfig struct {
    Method              AuthMethod `json:"method"`
    EncryptedPassword   string     `json:"encrypted_password,omitempty"`   // 加密后的密码
    PrivateKeyPath      string     `json:"private_key_path,omitempty"`     // 密钥文件路径
    EncryptedPassphrase string     `json:"encrypted_passphrase,omitempty"` // 密钥的加密密码
}

// BastionConfig 跳板机配置
type BastionConfig struct {
    Host    string     `json:"host"`
    Port    int        `json:"port"`
    User    string     `json:"user"`
    Auth    AuthConfig `json:"auth"`
}

// Vault 加密存储容器
type Vault struct {
    Version int            `json:"version"`
    Servers []ServerConfig `json:"servers"`
}

// ExecResult 命令执行结果（MCP 工具返回值）
type ExecResult struct {
    ServerID string `json:"server_id"`
    Command  string `json:"command"`
    Stdout   string `json:"stdout"`
    Stderr   string `json:"stderr"`
    ExitCode int    `json:"exit_code"`
    Duration int64  `json:"duration_ms"`
}

// FileTransferResult 文件传输结果
type FileTransferResult struct {
    ServerID string `json:"server_id"`
    Path     string `json:"path"`
    Size     int64  `json:"size_bytes"`
    Duration int64  `json:"duration_ms"`
}
```

### CLI 子命令定义（既是 CLI 入口，也可被 Skill 通过 Bash 调用）

| 子命令 | 参数 | 输出 | 说明 |
|--------|------|------|------|
| `ssh-mcp list` | 无 | `[]ServerConfig`（密码脱敏） | 列出已配置的服务器 |
| `ssh-mcp add` | `--id --name --host --port --user --auth-type [--password --key-path]` | 成功/失败 | 添加服务器配置 |
| `ssh-mcp remove` | `--id` | 成功/失败 | 删除服务器配置 |
| `ssh-mcp exec` | `--server --command [--timeout]` | `ExecResult{stdout, stderr, exit_code, duration}` | 执行命令 |
| `ssh-mcp upload` | `--server --local --remote` | `FileTransferResult` | 上传文件 |
| `ssh-mcp download` | `--server --remote --local` | `FileTransferResult` | 下载文件 |
| `ssh-mcp test` | `--server` | 连接状态 + 延迟 | 测试 SSH 连接 |
| `ssh-mcp vault init` | 无 | 成功/失败 | 初始化 vault key 和空配置文件 |
| `ssh-mcp serve` | 无 | MCP 服务端运行中 | （可选）启动 MCP stdio 模式 |

### 审计日志条目

```go
// AuditEntry 单条审计记录
type AuditEntry struct {
    Timestamp    string `json:"timestamp"`     // RFC3339
    ServerID     string `json:"server_id"`
    ServerHost   string `json:"server_host"`
    Command      string `json:"command"`       // 完整命令
    ExitCode     int    `json:"exit_code"`
    StdoutLen    int    `json:"stdout_len"`
    StderrLen    int    `json:"stderr_len"`
    DurationMs   int64  `json:"duration_ms"`
}
```

存储位置：`~/.ssh-mcp/audit.log`，JSONL 格式（每行一个 JSON 对象）。

### 安全保证

- 每个子命令执行前校验 `--server` 值是否存在于本地配置中（硬校验）
- 所有命令通过 Claude Code 的 Bash 权限弹窗由用户审批
- 密钥文件和凭证文件 chmod 0600，目录 chmod 0700

---

## 行为验收契约

### B-001：服务器配置生命周期管理
- **类型**：CRUD
- **入口**：`ssh-mcp add`、`ssh-mcp remove`、`ssh-mcp list` CLI 子命令
- **行为**：
  - GIVEN 用户添加了服务器 "prod-web" 的配置
  - WHEN AI 调用 `ssh_list_servers`
  - THEN 返回的列表包含 "prod-web"（密码字段脱敏为 `***`）
  - WHEN AI 调用 `ssh_remove_server --id "prod-web"`
  - THEN 该服务器从列表中消失，且配置文件更新
- **风险等级**：高（涉及凭证管理）
- **自动化要求**：单元测试 + 集成测试

### B-002：命令执行与目标校验
- **类型**：核心功能
- **入口**：`ssh-mcp exec` CLI 子命令
- **行为**：
  - GIVEN 已配置服务器 "prod-web"
  - WHEN AI 调用 `ssh_exec --server_id "prod-web" --command "uptime"`
  - THEN 返回 `ExecResult{stdout: "...", exit_code: 0}`
  - GIVEN 未配置的服务器 ID "ghost-server"
  - WHEN AI 调用 `ssh_exec --server_id "ghost-server" --command "ls"`
  - THEN 返回错误："server 'ghost-server' not found in local config"
- **风险等级**：高（防止误操作）
- **自动化要求**：单元测试（mock SSH） + 集成测试（真实 SSH）

### B-003：凭证加密存储
- **类型**：安全
- **入口**：`ssh-mcp add` CLI 子命令（写入时自动加密）
- **行为**：
  - GIVEN 用户添加密码认证服务器
  - WHEN 配置文件写入磁盘
  - THEN 密码字段以 AES-256-GCM 加密形式存储，明文不出现
  - WHEN 未解锁 vault 时直接读取配置文件
  - THEN 无法获取明文密码
- **风险等级**：关键
- **自动化要求**：单元测试验证加密/解密往返

### B-004：跨会话配置持久化
- **类型**：数据持久化
- **入口**：配置文件读写
- **行为**：
  - GIVEN 前一个 Claude Code 会话中添加了 3 台服务器
  - WHEN 新会话启动
  - THEN `ssh_list_servers` 返回相同的 3 台服务器
- **风险等级**：中
- **自动化要求**：集成测试验证文件读写

### B-005：多认证方式支持
- **类型**：功能
- **入口**：`ssh-mcp exec`、`ssh-mcp test` CLI 子命令
- **行为**：
  - GIVEN 服务器 A 使用密码认证
  - AND 服务器 B 使用 ED25519 密钥认证
  - AND 服务器 C 使用 SSH agent 认证
  - WHEN 分别连接三台服务器
  - THEN 均成功建立 SSH 连接
- **风险等级**：中
- **自动化要求**：集成测试（本地 SSH 容器）

---

## API / 入口影响

- **CLI 子命令**：9 个子命令（list/add/remove/exec/upload/download/test/vault/serve），独立可执行
- **Skill 接口**：Skill 通过 Bash 调用 `ssh-mcp` 子命令，`SKILL.md` 描述用法和安全规则
- **MCP 模式（可选）**：`ssh-mcp serve` 启动 stdio MCP 服务端，预留
- **配置文件**：`~/.ssh-mcp/servers.json.age`（加密 JSON），`~/.ssh-mcp/.vault-key`（随机密钥）
- **审计日志**：`~/.ssh-mcp/audit.log`（JSONL）
- **环境变量**：`SSH_MCP_CONFIG_DIR`（覆盖默认配置目录，可选）

---

## 兼容性和迁移策略

- **N/A**：新项目，无现有用户或 API 需要兼容
- **唯一注意**：如果用户已有 `~/.ssh/config` 配置，应提供导入命令（非破坏性读取），不在 v1 中实现。

---

## 边界和降级分析

| 场景 | 降级行为 |
|------|---------|
| 配置文件不存在 | 首次运行自动创建空 vault 和随机密钥文件 |
| vault key 文件丢失 | 无法解密已有凭证，需重新初始化 vault 和添加服务器 |
| SSH 连接超时 | 返回 `ExecResult{exit_code: -1, stderr: "connection timeout"}` |
| 进程崩溃 | 无持久连接，下次调用重新创建连接（CLI 无状态） |
| 网络断开 | 返回连接错误，不重试（由 AI 决定是否重试） |

---

## 文档影响

- **新建**：`.claude/skills/ssh-ops/SKILL.md` — 操作工作流
- **新建**：`README.md` — 项目说明、安装步骤、快速开始
- **更新**：`.harness/knowledge/DEPENDENCIES.md` — 添加 Go 依赖
- **更新**：`.harness/knowledge/ARCHITECTURE.md` — 添加 SSH 操作层说明
- **更新**：`AGENTS.md` — 补全 `## 环境` 和 `## 验证` 段

---

## 文档审查计划

- `design.md`：由 `reviews/001-derivation-generation.md` 和 `reviews/002-full-document-review.md` 覆盖
- `implementation.md`：在 derivation generation 阶段生成草案
- `PROGRESS.md`：在 derivation generation 阶段生成草案
- 项目知识文档：在任务完成前检查

---

## 验收标准

1. `go build ./cmd/ssh-mcp/` 在 Windows/Linux 上成功编译
2. `go test ./...` 所有测试通过
3. Claude Code 配置 MCP 服务端后，`ssh_list_servers` 可正常调用
4. 添加一台测试服务器 → 执行 `uptime` → 返回正确输出
5. 对不存在的 server_id 执行命令 → 返回错误
6. 重启 Claude Code → 服务器列表不丢失
7. 检查配置文件 → 密码字段为加密状态

---

## TDD 验证要求

`implementation.md` 必须包含 `## TDD 验证矩阵`，每个 B-ID 至少映射到一个 U-ID。

| B-ID | 说明 | 验证级别 |
|------|------|---------|
| B-001 | 服务器 CRUD | 单元测试 + 集成测试 |
| B-002 | 命令执行与目标校验 | 单元测试 + 集成测试 |
| B-003 | 凭证加密存储 | 单元测试 |
| B-004 | 跨会话持久化 | 集成测试 |
| B-005 | 多认证方式 | 集成测试 |

---

## 执行边界

`harness-ce-iterate` 可以：
- 创建 Go 模块、源码文件、测试文件
- 安装 Go 依赖
- 编译二进制
- 创建 Skill 定义文件（SKILL.md）
- 创建 MCP 配置文件示例
- 创建 README.md
- 更新 `.harness/knowledge/*` 知识文件
- 更新 `AGENTS.md` 的环境和验证段

`harness-ce-iterate` 不可以：
- 修改 `design.md` 正文（需回到 `harness-ce-plan`）
- 连接真实生产服务器进行测试
- 将任何配置或凭证上传到远程

---

## 风险和阻塞项

| 风险 | 等级 | 缓解 |
|------|------|------|
| Go 模块下载被墙（中国大陆网络） | 中 | 设置 `GOPROXY=https://goproxy.cn` |
| `x/crypto/ssh` 对非标准 SSH 实现的兼容性 | 低 | 只在标准 OpenSSH 服务端测试 |
| Windows 平台 SSH agent 不可用 | 低 | Agent 模式在 Windows 上降级为不可用 |

---

## 设计审核门禁

只有绑定文档包已审核后，`harness-ce-iterate` 才能执行。
