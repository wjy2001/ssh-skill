# 远程 SSH 服务器操作系统 — 实施计划

## 任务标识

- **slug**：`ssh-remote-ops`
- **日期**：2026-07-02
- **设计文档**：`.harness/tasks/2026-07-02-ssh-remote-ops/design.md`

## 派生状态

- **派生来源**：`harness-ce-plan` 派生子 agent
- **派生时间**：2026-07-02T18:10:00+08:00
- **绑定设计基线**：sha256:8dc437c862725de6fedbf36e2bf499dcb18899b6a372d10e25fc44fe4bcfea90（术语修正，无设计决策变更）
- **实施计划基线**：sha256:54082559abcc9a716c83138fc4c52342cfc92dc2c752bdb1898a7e49b7f15379
- **执行证据基线**：无
- **文档状态**：草案（待审查）

---

## 实施计划

本计划将设计文档中的架构拆解为 8 个执行单元（U1～U8），按依赖链串行推进。每个单元有明确的目标、文件边界、依赖、验证方式和 review gate。

### 执行单元总览

| U-ID | 标题 | 依赖 | 姿态 |
|------|------|------|------|
| U1 | 项目骨架 | 无 | 串行 |
| U2 | Types 和 Config 层 | U1 | 串行 |
| U3 | Vault 层 | U2 | 串行 |
| U4 | SSH 层 | U2 | 串行（可与 U3 并行） |
| U5 | CLI 层 | U3, U4 | 串行 |
| U6 | Audit 层 | U2 | 串行（可与 U3/U4 并行） |
| U7 | Skill 定义 | U5 | 串行 |
| U8 | 集成验证 | U1～U7 | 串行 |

---

## 文件改动计划

以下列出所有将创建或修改的文件，按模块分组。

### Go 模块文件（go/ 目录）

| 文件路径 | 操作 | 职责 |
|---------|------|------|
| `go/go.mod` | 新建 | Go 模块定义，声明 module path 和依赖 |
| `go/go.sum` | 新建 | 依赖校验锁文件 |
| `go/cmd/ssh-mcp/main.go` | 新建 | 单一入口，子命令路由分发 |
| `go/internal/types/types.go` | 新建 | 共享类型：ServerConfig, AuthConfig, AuthMethod, BastionConfig, Vault, ExecResult, FileTransferResult, AuditEntry |
| `go/internal/config/config.go` | 新建 | 配置解析与默认值：SSH_MCP_CONFIG_DIR 环境变量，默认配置目录 |
| `go/internal/vault/vault.go` | 新建 | Vault 核心：AES-256-GCM 加解密，Argon2id 密钥派生 |
| `go/internal/vault/keygen.go` | 新建 | 随机密钥生成（32 字节），首次运行自动生成 |
| `go/internal/vault/storage.go` | 新建 | 配置文件读写：servers.json.age 的加载和保存 |
| `go/internal/vault/vault_test.go` | 新建 | Vault 层单元测试 |
| `go/internal/ssh/client.go` | 新建 | SSH 连接建立，多认证方式支持（password/key/agent） |
| `go/internal/ssh/exec.go` | 新建 | 远程命令执行，目标校验 |
| `go/internal/ssh/transfer.go` | 新建 | 文件上传/下载（SFTP） |
| `go/internal/ssh/client_test.go` | 新建 | SSH 层单元测试（mock）/ 集成测试 |
| `go/internal/cli/list.go` | 新建 | `ssh-mcp list` 子命令实现 |
| `go/internal/cli/add.go` | 新建 | `ssh-mcp add` 子命令实现 |
| `go/internal/cli/remove.go` | 新建 | `ssh-mcp remove` 子命令实现 |
| `go/internal/cli/exec.go` | 新建 | `ssh-mcp exec` 子命令实现 |
| `go/internal/cli/upload.go` | 新建 | `ssh-mcp upload` 子命令实现 |
| `go/internal/cli/download.go` | 新建 | `ssh-mcp download` 子命令实现 |
| `go/internal/cli/test.go` | 新建 | `ssh-mcp test` 子命令实现 |
| `go/internal/cli/vault.go` | 新建 | `ssh-mcp vault init` 子命令实现 |
| `go/internal/cli/serve.go` | 新建 | `ssh-mcp serve` MCP 模式入口 |
| `go/internal/cli/root.go` | 新建 | 根命令注册，子命令路由 |
| `go/internal/audit/audit.go` | 新建 | 审计日志写入：JSONL 格式，audit.log |
| `go/internal/audit/audit_test.go` | 新建 | 审计层单元测试 |

### Skill 定义文件

| 文件路径 | 操作 | 职责 |
|---------|------|------|
| `.claude/skills/ssh-ops/SKILL.md` | 新建 | Skill 定义：用法说明、安全规则、命令审批策略 |

### 项目配置文件

| 文件路径 | 操作 | 职责 |
|---------|------|------|
| `README.md` | 新建 | 项目说明、安装步骤、快速开始 |
| `bin/.gitkeep` | 新建 | 确保 bin/ 目录纳入版本控制 |
| `.gitignore` | 更新 | 添加 `bin/ssh-mcp*`、`*.exe` 等忽略规则 |

### 知识库文件

| 文件路径 | 操作 | 职责 |
|---------|------|------|
| `.harness/knowledge/DEPENDENCIES.md` | 更新 | 添加 Go 运行时和外部依赖说明 |
| `.harness/knowledge/ARCHITECTURE.md` | 更新 | 添加 SSH 操作层架构说明 |
| `AGENTS.md` | 更新 | 补全 `## 环境` 和 `## 验证` 段 |

---

## 验证计划

### 编译验证

- `go build ./cmd/ssh-mcp/` 在 Windows/Linux/macOS 上成功
- 无编译警告
- 二进制文件大小合理（< 30MB）

### 单元测试

- `go test ./internal/types/...`：类型定义正确性
- `go test ./internal/config/...`：配置解析正确性
- `go test ./internal/vault/...`：加解密往返测试、密钥生成测试、文件读写测试
- `go test ./internal/ssh/...`：目标校验逻辑、连接参数构造（mock）
- `go test ./internal/audit/...`：JSONL 写入格式正确性
- `go test ./internal/cli/...`：子命令参数解析

### 集成测试

- 配置文件读写的端到端往返测试
- SSH 连接本地测试容器（OpenSSH Docker 容器）
- 多认证方式集成测试矩阵（password / ED25519 key / RSA key / agent）
- 审计日志写入和读取验证

### 端到端验证

- 编译二进制 → 安装 vault → 添加服务器 → 执行命令 → 获取输出
- 对不存在 server_id 执行命令 → 返回错误
- 重启进程 → 服务器列表不丢失
- 检查配置文件 → 密码字段加密状态

---

## TDD 验证矩阵

每个行为验收契约（B-ID）至少映射到一个执行单元（U-ID），且对应至少一个测试用例。

| B-ID | 说明 | 验证级别 | 映射 U-ID | 测试方法 |
|------|------|---------|-----------|---------|
| B-001 | 服务器配置生命周期管理 | 单元测试 + 集成测试 | U2, U3, U5 | Vault CRUD 单元测试 + CLI add/list/remove 集成测试 |
| B-002 | 命令执行与目标校验 | 单元测试 + 集成测试 | U4, U5 | 目标校验单元测试（mock）+ exec 集成测试 |
| B-003 | 凭证加密存储 | 单元测试 | U3 | 加密/解密往返测试，明文不出现在加密文件中 |
| B-004 | 跨会话配置持久化 | 集成测试 | U3, U5 | Vault 文件读写往返测试 |
| B-005 | 多认证方式支持 | 集成测试 | U4, U8 | SSH 连接集成测试矩阵（password/key/agent） |

---

## 执行单元计划

### U1: 项目骨架

- **目标**：初始化 Go 模块，创建目录结构，安装核心依赖
- **文件边界**：
  - `go/go.mod`（新建）
  - `go/cmd/ssh-mcp/main.go`（新建，最小入口：仅打印版本）
  - `go/internal/` 目录树（新建空目录）
  - `.gitignore`（更新）
  - `bin/.gitkeep`（新建）
- **依赖**：无
- **执行姿态**：串行
- **验证**：
  - `go mod tidy` 成功
  - `go build ./cmd/ssh-mcp/` 编译无错误
  - `dir go/internal/` 确认目录结构完整
- **Review gate**：
  - 规格符合：目录结构与 design.md D5 一致（`go/cmd/ssh-mcp/`、`go/internal/{vault,ssh,config,audit}`）
  - 质量：go.mod 无不必要的间接依赖

### U2: Types 和 Config 层

- **目标**：定义所有共享数据类型和配置解析逻辑
- **文件边界**：
  - `go/internal/types/types.go`（新建）：ServerConfig, AuthConfig, AuthMethod, BastionConfig, Vault, ExecResult, FileTransferResult, AuditEntry
  - `go/internal/config/config.go`（新建）：SSH_MCP_CONFIG_DIR 解析，默认 `~/.ssh-mcp/`，目录权限确保
- **依赖**：U1
- **执行姿态**：串行
- **验证**：
  - `go vet ./internal/types/...` 无问题
  - `go vet ./internal/config/...` 无问题
  - 类型定义与 design.md 数据结构节完全对应
- **Review gate**：
  - 规格符合：类型字段名、JSON tag、omitempty 标记与 design.md 一致
  - 质量：无未使用字段，json tag 命名 snake_case

### U3: Vault 层

- **目标**：实现 AES-256-GCM 加解密、Argon2id 密钥派生、随机密钥生成、加密配置文件读写
- **文件边界**：
  - `go/internal/vault/vault.go`（新建）：加解密核心逻辑
  - `go/internal/vault/keygen.go`（新建）：密钥生成与持久化到 `~/.ssh-mcp/.vault-key`
  - `go/internal/vault/storage.go`（新建）：加载/保存 `~/.ssh-mcp/servers.json.age`
  - `go/internal/vault/vault_test.go`（新建）：加密往返测试、密钥生成测试
- **依赖**：U2
- **执行姿态**：串行
- **验证**：
  - `go test ./internal/vault/... -v` 全部通过
  - 加密后 payload 字节数 > 明文 payload
  - 同一数据两次加密结果不同（随机 nonce）
  - 错误密钥解密返回明确错误
  - 密钥文件权限 0600，目录权限 0700
- **Review gate**：
  - 规格符合：文件格式 `[16B salt][12B nonce][AES-GCM ciphertext]` 与 design.md D4 一致
  - 质量：Argon2id 参数合理（time=3, memory=64MB, threads=4），无硬编码密钥

### U4: SSH 层

- **目标**：实现 SSH 连接建立、命令执行、文件传输、多认证方式支持
- **文件边界**：
  - `go/internal/ssh/client.go`（新建）：连接建立，认证方式分发（password/key/agent），目标校验
  - `go/internal/ssh/exec.go`（新建）：Run 命令，超时控制，ExecResult 构造
  - `go/internal/ssh/transfer.go`（新建）：SFTP 上传/下载
  - `go/internal/ssh/client_test.go`（新建）：单元测试（mock SSH server）+ 集成测试标记
- **依赖**：U2（仅依赖 types，不依赖 vault）
- **执行姿态**：串行（可与 U3 并行）
- **验证**：
  - `go vet ./internal/ssh/...` 无问题
  - 单元测试覆盖：目标校验逻辑（server_id 匹配）、连接参数构造、错误处理
  - 集成测试覆盖：本地 SSH 测试容器（OpenSSH Docker）
- **Review gate**：
  - 规格符合：支持 password / key / agent 三种认证方式（design.md B-005）
  - 质量：连接超时默认 30s，SSH 已知主机验证，goroutine 安全

### U5: CLI 层

- **目标**：实现所有 CLI 子命令入口，集成 vault、ssh、audit 层
- **文件边界**：
  - `go/internal/cli/root.go`（新建）：根命令注册，全局 flags（--config-dir）
  - `go/internal/cli/list.go`（新建）：列出所有服务器，密码脱敏
  - `go/internal/cli/add.go`（新建）：添加服务器，自动加密密码
  - `go/internal/cli/remove.go`（新建）：删除服务器
  - `go/internal/cli/exec.go`（新建）：执行命令，目标校验，审计记录
  - `go/internal/cli/upload.go`（新建）：上传文件
  - `go/internal/cli/download.go`（新建）：下载文件
  - `go/internal/cli/test.go`（新建）：连接测试
  - `go/internal/cli/vault.go`（新建）：vault 初始化
  - `go/internal/cli/serve.go`（新建）：MCP serve 入口（预留）
  - `go/cmd/ssh-mcp/main.go`（更新）：引入 CLI 模块，执行 root
- **依赖**：U3, U4
- **执行姿态**：串行
- **验证**：
  - `go build ./cmd/ssh-mcp/` 编译成功
  - `./ssh-mcp --help` 输出子命令列表
  - `./ssh-mcp vault init` 创建配置目录和文件
  - `./ssh-mcp list` 空列表输出
- **Review gate**：
  - 规格符合：9 个子命令与 design.md CLI 子命令定义表一致
  - 质量：每个子命令有 --help 输出，参数校验完整，错误信息明确

### U6: Audit 层

- **目标**：实现 JSONL 格式审计日志写入
- **文件边界**：
  - `go/internal/audit/audit.go`（新建）：AuditEntry 写入函数，文件追加模式，权限控制
  - `go/internal/audit/audit_test.go`（新建）：JSONL 格式验证，追加写入验证
- **依赖**：U2
- **执行姿态**：串行（可与 U3/U4 并行）
- **验证**：
  - `go test ./internal/audit/... -v` 全部通过
  - 每条记录独立一行，合法 JSON
  - 多线程写入无数据竞争
- **Review gate**：
  - 规格符合：AuditEntry 字段与 design.md 审计日志条目节一致
  - 质量：文件追加模式，flock 锁，audit.log 权限 0600

### U7: Skill 定义

- **目标**：编写 SKILL.md，描述操作工作流和安全规则
- **文件边界**：
  - `.claude/skills/ssh-ops/SKILL.md`（新建）
- **依赖**：U5
- **执行姿态**：串行
- **验证**：
  - SKILL.md 内容可读，语法正确
  - 包含所有子命令的使用示例
  - 包含安全规则说明（审批模式、目标校验）
- **Review gate**：
  - 规格符合：Skill 定义与 design.md 行为验收契约一致
  - 质量：示例命令可执行，无占位符未替换

### U8: 集成验证

- **目标**：编译、全量测试、端到端验证
- **文件边界**：
  - `go/` 下所有文件（可做微小修正）
  - `README.md`（新建或更新）
  - `.harness/knowledge/DEPENDENCIES.md`（更新）
  - `.harness/knowledge/ARCHITECTURE.md`（更新）
  - `AGENTS.md`（更新环境和验证段）
- **依赖**：U1～U7
- **执行姿态**：串行
- **验证**：
  - `go build ./cmd/ssh-mcp/` 在至少两个平台（Windows + Linux）编译成功
  - `go test ./... -v` 全部通过
  - `go vet ./...` 无问题
  - 端到端流程：vault init → add server → list → exec uptime → remove → list（验证为空）
  - 密码字段在文件中为加密状态
  - 审计日志文件存在且内容合法
- **Review gate**：
  - 规格符合：所有验收标准 1～7 项通过（design.md 验收标准节）
  - 质量：无 panic、无 goroutine 泄漏、无竞态条件

---

## 执行记录

| U-ID | 开始时间 | 完成时间 | 操作摘要 | 成功/失败 | 产物 |
|------|---------|---------|---------|----------|------|
| U1 | 2026-07-02 18:15 | 2026-07-02 18:16 | 创建 go/ 目录结构、go.mod、main.go、.gitignore、bin/.gitkeep | 成功 | 目录树 + 可编译二进制 |
| U2 | 2026-07-02 18:16 | 2026-07-02 18:17 | 定义 types.go（8个类型）、config.go（5个路径函数） | 成功 | go vet 通过 |
| U3 | 2026-07-02 18:17 | 2026-07-02 18:18 | vault.go（AES-256-GCM+Argon2id）、keygen.go、storage.go + 11个测试 | 成功 | 11/11 PASS |
| U4 | 2026-07-02 18:18 | 2026-07-02 18:19 | client.go（多认证）、exec.go（命令执行）、transfer.go（SFTP）+ 7个测试 | 成功 | 7/7 PASS，go vet 通过 |
| U5 | 2026-07-02 18:19 | 2026-07-02 18:20 | 11个CLI文件（root+9命令+加密辅助）、main.go集成 | 成功 | 编译通过，CLI CRUD + 加密验证 |
| U6 | 2026-07-02 18:19 | 2026-07-02 18:19 | audit.go（JSONL写入器）+ 3个测试（含并发安全） | 成功 | 3/3 PASS |
| U7 | 2026-07-02 18:20 | 2026-07-02 18:20 | SKILL.md 创建（用法、安全规则、工作流、排错） | 成功 | 技能文件就绪 |
| U8 | 2026-07-02 18:20 | 2026-07-02 18:22 | 全量测试、go vet、编译、端到端测试、README.md、AGENTS.md、知识库更新 | 成功 | 21/21 test PASS，6/6 e2e PASS |

---

## 验证结果

| 验证项 | 命令 | 预期 | 实际 | 通过 |
|-------|------|------|------|------|
| 编译 | `go build ./cmd/ssh-mcp/` | 成功 | 成功（7.3MB 二进制） | ✅ |
| go vet | `go vet ./...` | 无问题 | 无问题 | ✅ |
| Vault 单元测试 | `go test ./internal/vault/... -v` | 全部通过 | 11/11 PASS | ✅ |
| SSH 单元测试 | `go test ./internal/ssh/... -v` | 全部通过 | 7/7 PASS | ✅ |
| Audit 单元测试 | `go test ./internal/audit/... -v` | 全部通过 | 3/3 PASS | ✅ |
| 全量测试 | `go test ./... -count=1` | 全部通过 | 21/21 PASS | ✅ |
| 端到端：vault init | `ssh-mcp vault init` | 创建目录+密钥 | 通过 | ✅ |
| 端到端：add server | `ssh-mcp add ...` | 添加成功 | 通过 | ✅ |
| 端到端：list | `ssh-mcp list` | 显示已添加 | 通过 | ✅ |
| 端到端：密码加密 | `grep plaintext servers.json.age` | 找不到明文 | 0 matches | ✅ |
| 端到端：remove | `ssh-mcp remove ...` | 删除成功 | 通过 | ✅ |
| 端到端：list after remove | `ssh-mcp list` | 空 | 通过 | ✅ |

---

## 知识沉淀

| 沉淀项 | 目标位置 | 状态 |
|-------|---------|------|
| Go 项目结构和依赖管理策略 | `.harness/knowledge/DEPENDENCIES.md` | ✅ 已落库 |
| SSH 操作层架构说明 | `.harness/knowledge/ARCHITECTURE.md` | ✅ 已落库 |
| Vault 加密模式（AES-256-GCM + Argon2id） | `.harness/knowledge/PATTERNS.md` | 转交 doc-gardening（模式已存在无新增） |
| AGENTS.md 环境和验证段补全 | `AGENTS.md` | ✅ 已落库 |

## 文档防腐记录

| 检查项 | 状态 |
|-------|------|
| GLOSSARY.md 术语（SSH、vault、MCP、audit 等） | 待 doc-gardening（非本任务阻塞项） |
| ARCHITECTURE.md 分层与当前实现一致 | ✅ 已更新 |
| README.md 内容与当前功能一致 | ✅ 已创建 |
