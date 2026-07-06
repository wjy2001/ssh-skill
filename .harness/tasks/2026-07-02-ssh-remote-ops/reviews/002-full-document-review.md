# 完整文档包审查报告 002

## 审查元数据

- **报告编号**：002-full-document-review
- **审查时间**：2026-07-02T18:30:00+08:00
- **审查范围**：
  - `design.md`（设计事实源）
  - `implementation.md`（实施计划草案）
  - `PROGRESS.md`（任务恢复指针）
  - `reviews/001-derivation-generation.md`（派生生成报告）
  - `AGENTS.md`（项目指令）
  - `.harness/agents/REVIEWER.md`（审查清单）
  - `.harness/knowledge/PATTERNS.md`（模式与反模式）
  - `.harness/knowledge/ARCHITECTURE.md`（架构约束）
- **绑定基线**：
  - 设计基线：sha256:d51afb5fd944a37d982d861049a1b1863a0e673255cb4e02f59a5f99916273d6
  - 实施计划基线：sha256:54082559abcc9a716c83138fc4c52342cfc92dc2c752bdb1898a7e49b7f15379
- **审查者**：完整审查 subagent
- **审查结论**：通过（允许进入 iterate 阶段），附带 3 个建议修复和 3 个 Nice-to-have

---

## 1. 基线一致性

### 1.1 设计基线

| 文档 | 绑定值 | 状态 |
|------|--------|------|
| design.md `## 设计基线` | sha256:d51afb5fd944a37d982d861049a1b1863a0e673255cb4e02f59a5f99916273d6 | 事实源 |
| implementation.md `## 派生状态` | sha256:d51afb5fd944a37d982d861049a1b1863a0e673255cb4e02f59a5f99916273d6 | 一致 |
| PROGRESS.md `## 基线绑定` | sha256:d51afb5fd944a37d982d861049a1b1863a0e673255cb4e02f59a5f99916273d6 | 一致 |

**结论：通过。** 三份文档设计基线完全一致。

### 1.2 实施计划基线

| 文档 | 绑定值 | 状态 |
|------|--------|------|
| implementation.md `## 派生状态` | sha256:54082559abcc9a716c83138fc4c52342cfc92dc2c752bdb1898a7e49b7f15379 | 事实源 |
| PROGRESS.md `## 基线绑定` | sha256:54082559abcc9a716c83138fc4c52342cfc92dc2c752bdb1898a7e49b7f15379 | 一致 |
| 001 报告 实施计划基线说明 | "sha256:待计算" | 信息已过期 |

**结论：通过（附注）。** implementation.md 和 PROGRESS.md 的实施计划基线一致。001 报告中标注的"待计算"已被实际计算的基线值取代，001 报告的该字段已过期但不影响一致性判断——001 报告不是状态绑定文档。

### 1.3 内部一致性小问题

design.md `## 审批状态` 节（第 42 行）仍写 `实施计划基线：无`，但 `## 派生状态` 节（第 35 行）已正确记录了实施计划基线。审批状态节的该字段已过期，应在下一次 plan 更新时同步修正。

---

## 2. 自包含性

### 2.1 design.md 自包含性：通过

包含完整的背景/问题定义、5 个决策（D1-D5）、数据结构定义、行为验收契约（B-001 到 B-005）、验收标准、降级分析。不依赖外部聊天上下文中才能理解的信息。

### 2.2 implementation.md 自包含性：通过

明确声明绑定设计基线，所有执行单元的边界、依赖、验证方式和 review gate 均可独立理解。文件改动计划列出了全部涉及文件及其路径。

### 2.3 PROGRESS.md 恢复指针正确性：通过

- 当前状态准确：派生生成完成，待审查
- 执行单元状态表完整：8 个 U-ID 均为 pending
- 恢复流程完整：7 步恢复 + 5 项检查清单
- 当前指针明确：下一操作为完整审查 → iterate

### 2.4 文档间一致性检查

**发现的矛盾**：

1. **design.md 内部：行为验收契约的入口命名与 D1 决策不一致（Important）**
   - B-001 到 B-004 的"入口"字段引用 MCP 工具名（`ssh_add_server`、`ssh_remove_server`、`ssh_exec`、`ssh_test_connection`），但 D1 已选择 CLI-first 方案。CLI 子命令名为 `ssh-mcp add`、`ssh-mcp remove`、`ssh-mcp exec`、`ssh-mcp test`。
   - B-005 没有明确入口名，不涉及此问题。
   - implementation.md 正确使用了 CLI 子命令名。
   - **建议**：design.md B-001 到 B-004 的入口字段应更新为 CLI 子命令名或同时标注 CLI/MCP 双入口。这不阻塞 iterate——implementation.md 已正确映射。

2. **design.md 内部：降级分析与假设 3 / D4 决策矛盾（Important）**
   - 降级分析表第 589 行："配置文件不存在 → 首次运行自动创建空 vault，提示用户设置主密码"
   - 降级分析表第 592 行："主密码错误 → 返回明确错误，不泄露任何信息"
   - 但假设 3（第 113 行）明确："密钥由 Go 运行时自动生成并落盘……用户无需输入主密码"
   - D4（第 206 行）明确："密钥由 Go 在首次运行时自动生成并落盘"
   - **矛盾**：降级分析假设存在"主密码"概念，但 D4 决定不使用主密码，采用随机密钥自动生成方案。
   - **建议**：修正降级分析表，将"提示用户设置主密码"改为"自动生成 vault key 并落盘 `~/.ssh-mcp/.vault-key`"，将"主密码错误"改为"vault key 损坏或丢失 → 返回明确错误，提示重新初始化 vault"。

3. **design.md 内部：审批状态节过期（Nice-to-have）**
   - 审批状态节（第 42 行）"实施计划基线：无"与派生状态节（第 35 行）矛盾。该节是预留的审批记录模板，尚未被审批流程更新，不影响当前审查。

---

## 3. 实施可行性

### 3.1 U-ID 完整性检查

| U-ID | 目标 | 文件边界 | 依赖 | 验证 | Review Gate | 完备 |
|------|------|---------|------|------|-------------|------|
| U1 | 项目骨架 | 5 个文件，明确 | 无 | 3 项，可验证 | 2 项标准 | 是 |
| U2 | Types + Config | 2 个文件，明确 | U1 | 2 项，可验证 | 2 项标准 | 是 |
| U3 | Vault 层 | 4 个文件，明确 | U2 | 5 项，可验证 | 2 项标准 | 是 |
| U4 | SSH 层 | 4 个文件，明确 | U2 | 3 项，可验证 | 2 项标准 | 是 |
| U5 | CLI 层 | 11 个文件，明确 | U3, U4 | 4 项，可验证 | 2 项标准 | 是 |
| U6 | Audit 层 | 2 个文件，明确 | U2 | 3 项，可验证 | 2 项标准 | 是 |
| U7 | Skill 定义 | 1 个文件，明确 | U5 | 3 项，可验证 | 2 项标准 | 是 |
| U8 | 集成验证 | 6+ 文件范围 | U1-U7 | 6 项，可验证 | 2 项标准 | 是 |

**结论：通过。** 所有 8 个 U-ID 都满足 AGENTS.md 要求的"目标、文件边界、依赖、验证和 review gate"完整性。

### 3.2 文件改动计划覆盖度

对照 design.md 要求的产出：

| design.md 要求 | 文件改动计划覆盖 | 状态 |
|---------------|-----------------|------|
| Go CLI 二进制（CLI-first） | `go/cmd/ssh-mcp/main.go` + `go/internal/cli/*.go` | 覆盖 |
| Types 层 | `go/internal/types/types.go` | 覆盖 |
| Config 层 | `go/internal/config/config.go` | 覆盖 |
| Vault 加密/密钥 | `go/internal/vault/*.go` | 覆盖 |
| SSH 连接/执行/传输 | `go/internal/ssh/*.go` | 覆盖 |
| Audit 日志 | `go/internal/audit/*.go` | 覆盖 |
| Skill 定义 | `.claude/skills/ssh-ops/SKILL.md` | 覆盖 |
| README | `README.md` | 覆盖 |
| MCP 可选模式 | `go/internal/cli/serve.go`（预留） | 覆盖 |
| 知识库更新 | DEPENDENCIES.md, ARCHITECTURE.md, PATTERNS.md | 覆盖 |
| AGENTS.md 补全 | AGENTS.md | 覆盖 |

design.md 第 604 行文档影响列出的 5 项产出全部在文件改动计划中有对应操作。另外 design.md 第 603 行提到 `## 更新：.harness/knowledge/PATTERNS.md` 但文件改动计划中知识库文件列表只列了 DEPENDENCIES.md、ARCHITECTURE.md 和 AGENTS.md，PATTERNS.md 未在文件改动计划中列出但 implementation.md 知识沉淀节（第 338 行）中有 `PATTERNS.md` 沉淀项。这是一个轻微遗漏——文件改动计划应补充 PATTERNS.md 更新行。

**结论：通过（附注）。** PATTERNS.md 更新在知识沉淀表中列出但未在文件改动计划表中列出，建议补充。

### 3.3 TDD 验证矩阵完整性

| design.md B-ID | implementation.md 映射 U-ID | 测试方法 | 覆盖 |
|---------------|---------------------------|---------|------|
| B-001（配置 CRUD） | U2, U3, U5 | Vault CRUD 单元测试 + CLI add/list/remove 集成测试 | 是 |
| B-002（命令执行与校验） | U4, U5 | 目标校验单元测试（mock）+ exec 集成测试 | 是 |
| B-003（凭证加密） | U3 | 加解密往返测试 | 是 |
| B-004（跨会话持久化） | U3, U5 | Vault 文件读写往返测试 | 是 |
| B-005（多认证方式） | U4, U8 | SSH 连接集成测试矩阵（password/key/agent） | 是 |

**结论：通过。** 所有 5 个 B-ID 均映射到至少一个 U-ID 和具体测试方法。

### 3.4 执行姿态正确性

U1（串行）→ U2（串行）→ {U3, U4}（可并行）→ U5（串行，依赖 U3+U4）→ U7（串行，依赖 U5）→ U8（串行，依赖全部）

U6 可与 U3/U4 并行。

并行标注正确：U3 和 U4 都只依赖 U2，互不依赖；U6 只依赖 U2，可与 U3/U4 并行。

**结论：通过。** 依赖链和并行机会标注合理。

---

## 4. 已知反模式检查

### 4.1 软闸门完成：未命中

- 每个 U-ID 有可重复执行的验证命令（`go build`、`go test`、`go vet`）
- U8 有完整的端到端验证流程
- 验证结果表和执行记录表已预留

**结论：通过。**

### 4.2 派生状态冒充事实源：未命中

- design.md 设计基线已在 implementation.md 和 PROGRESS.md 中绑定
- 实施计划基线已在 PROGRESS.md 中绑定
- 基线不一致时的恢复流程（PROGRESS.md 恢复说明第 2 步）已在 PROGRESS.md 中明确

**结论：通过。**

### 4.3 单元 Review 缺失：未命中（预执行阶段不适用）

- PROGRESS.md 为每个 U-ID 预留了 Review 状态列
- implementation.md 为每个 U-ID 定义了 review gate（规格符合 + 质量结论）
- 当前为预执行阶段，所有 U-ID 状态为 pending

**结论：通过（预执行）。** 结构上已为逐单元 review 做好准备。执行阶段需确保每个 U-ID 在标记 done 前通过 review gate。

### 4.4 只报告不防腐：未命中

- implementation.md 知识沉淀节列出了 4 项待落库项，每项有明确的目标位置
- 文档防腐记录节列出了 3 项待检查项

**结论：通过。** 文档防腐计划具体可追踪。

---

## 5. 架构约束

### 5.1 分层遵从性

AGENTS.md 和 ARCHITECTURE.md 定义的分层约束：`Types → Config → Repo → Service → Runtime → UI`（仅允许上层引用下层）。

design.md 架构分层与实现映射：

| 概念层 | design.md 描述 | 实现包 | 引用方向 |
|--------|---------------|--------|---------|
| UI | Claude Code（外部） | 不实现 | 调用 CLI |
| Runtime | ssh-mcp CLI 入口 | `go/cmd/ssh-mcp/main.go` + `go/internal/cli/` | 依赖 Service |
| Service | SSH 管理 / Vault / Audit | `go/internal/ssh/`, `go/internal/vault/`, `go/internal/audit/` | 依赖 Repo |
| Repo | 文件 IO / SSH 客户端 | 内嵌于 Service 包中 | 依赖 Config |
| Config | SSH_MCP_CONFIG_DIR | `go/internal/config/` | 依赖 Types |
| Types | 共享类型 | `go/internal/types/` | 无依赖 |

**关键分析**：
- `go/internal/types/` 不被其他 internal 包引用时会被编译拒绝——实际上各 internal 包会 import types 包，方向正确（上层 import 下层 types）。
- Vault 层（Service）同时处理加密和文件读写（Repo），SSH 层同时处理连接管理（Service）和 SSH 客户端调用（Repo），这是合理的 Go 实践——同一个 internal 包可以包含多个概念层的实现，只要代码内部的调用方向不发生跨层逆向引用。
- 文件改动计划没有出现 CLI 直接引用 config 文件路径的情况——CLI 通过 Vault 的 storage 和 config 包获取配置。

**结论：通过。** 实现计划的分层结构符合架构约束。

### 5.2 文件放置正确性

| 文件 | 概念层 | 放置正确 |
|------|--------|---------|
| `go/internal/types/types.go` | Types | 是 |
| `go/internal/config/config.go` | Config | 是 |
| `go/internal/vault/*.go` | Service + Repo | 是（同一概念领域合并） |
| `go/internal/ssh/*.go` | Service + Repo | 是（同一概念领域合并） |
| `go/internal/audit/*.go` | Service | 是 |
| `go/internal/cli/*.go` | Runtime | 是 |

### 5.3 跨层引用风险

执行 `cx references` 检查实施计划中是否存在潜在的跨层引用：

- CLI 入口（`main.go` + `cli/`）会引用 vault、ssh、audit、config、types 包——CLI 属于 Runtime 层，引用 Service 层（vault/ssh/audit）和 Config 层（config）及 Types 层是合法的。
- 没有发现 CLI 直接操作文件系统或直接调用 `os.Open` 读取 `servers.json.age` 的计划——所有文件操作封装在 vault 包内。

**结论：通过。** 无跨层引用风险。

---

## 6. 行为和验收

### 6.1 B-ID 映射完整性

所有 B-001 到 B-005 均在 TDD 验证矩阵中完整映射（见 3.3 节表格）。每个 B-ID 对应至少一个测试用例。

**结论：通过。**

### 6.2 验收标准 1-7 覆盖检查

| # | 验收标准 | U8 验证覆盖 | 状态 |
|---|---------|-----------|------|
| 1 | `go build` 在 Windows/Linux 编译成功 | U8 验证第 1 项 | 覆盖 |
| 2 | `go test ./...` 全部通过 | U8 验证第 2 项 | 覆盖 |
| 3 | `ssh_list_servers` 可正常调用 | U8 E2E 流程第 3 步 | 覆盖 |
| 4 | 添加服务器 → exec uptime → 正确输出 | U8 E2E 流程第 4 步 | 覆盖 |
| 5 | 不存在 server_id → 错误 | U8 E2E 流程第 5 步 | 覆盖 |
| 6 | 重启 → 列表不丢失 | U8 E2E 流程第 6 步 | 覆盖 |
| 7 | 密码字段为加密状态 | U8 E2E 流程第 7 步 | 覆盖 |

**结论：通过。** 所有 7 条验收标准在 U8 中有对应验证。

### 6.3 降级分析覆盖检查

| 降级场景 | design.md 描述 | 实施计划覆盖 | 状态 |
|---------|---------------|-------------|------|
| 配置文件不存在 | 自动创建空 vault | U3（vault init）+ U5（vault init CLI）| 覆盖（但描述需修正） |
| ~~主密码错误~~ | ~~返回错误~~ | ~~N/A~~ | ⚠️ 场景与 D4 矛盾 |
| SSH 连接超时 | ExecResult{exit_code: -1} | U4（连接超时 30s + 错误处理）| 覆盖 |
| MCP 服务端崩溃 | Claude Code 自动重启 | U5（serve.go 预留 MCP 模式）| 覆盖 |
| 网络断开 | 返回连接错误，不重试 | U4（错误处理，无重试逻辑）| 覆盖 |
| 并发执行 | 连接池复用，线程安全 | U4 review gate（goroutine 安全）| 覆盖 |

**结论：通过（附注）。** "主密码错误"场景与 D4 决策（无主密码）矛盾，应修正降级分析表（见 2.4 节发现 2）。

---

## 7. 合规性

### 7.1 项目目录结构

design.md D5 决定：Go 代码放 `go/` 子目录。implementation.md 文件改动计划中所有 Go 文件路径均以 `go/` 为前缀。符合。

### 7.2 凭证存储方案

design.md D4 决定：AES-256-GCM + 随机密钥自动生成落盘。implementation.md U3 验证项包含"加密后 payload 字节数 > 明文"、"同一数据两次加密结果不同"、"错误密钥解密返回明确错误"、"密钥文件权限 0600"。符合。

### 7.3 命令安全模型

design.md D2 决定：审批模式。implementation.md U7（SKILL.md）要求"包含安全规则说明（审批模式、目标校验）"。目标校验在 U4/U5（exec 命令中校验 server_id）和 U5 review gate 中落实。符合。

---

## 8. 发现汇总

### Critical（阻塞 iterate）

无。

### Important（应修复，但不阻塞 iterate）

1. **[design.md:behavior-contracts-vs-d1]** 行为验收契约 B-001 至 B-004 的"入口"字段使用 MCP 工具命名（`ssh_add_server` 等），但 D1 选择 CLI-first 方案。implementation.md 已正确使用 CLI 子命令名。
   - **建议修复**：将 B-001 到 B-004 入口字段从 MCP 工具名改为 CLI 子命令名（如 `ssh-mcp add`、`ssh-mcp list`），或标注 `CLI: ssh-mcp add / MCP: ssh_add_server` 双入口形式。

2. **[design.md:degradation-vs-d4]** 降级分析表的两行（"配置文件不存在 → 提示用户设置主密码"、"主密码错误 → 返回错误"）与 D4 决策和假设 3 矛盾——D4 采用随机密钥自动生成，不存在"主密码"概念。
   - **建议修复**：修正降级分析表，改为：
     - "配置文件不存在 → 自动生成 vault key 并创建空 vault"
     - "vault key 丢失或损坏 → 返回错误，提示执行 `ssh-mcp vault init` 重新初始化"

3. **[implementation.md:patterns-missing-from-file-plan]** 文件改动计划的知识库文件列表未包含 `.harness/knowledge/PATTERNS.md` 更新行，但知识沉淀节（第 338 行）列出了 PATTERNS.md 沉淀项。
   - **建议修复**：在文件改动计划的知识库文件表中添加 `PATTERNS.md` 更新行。

### Nice-to-have

4. **[design.md:approval-status-stale]** 审批状态节（第 42 行）"实施计划基线：无"与派生状态节（第 35 行）已记录的有效实施计划基线不一致，属模板字段未同步。
5. **[implementation.md:derive-report-reference]** 001 派生生成报告的第 21 行"实施计划基线：待计算"已在 implementation.md 中完成计算值更新，001 报告的该备注已过期，可考虑后续更新或存档。
6. **[design.md:degradation-translation]** 降级分析表第 591 行"主密码错误"与第 589 行"主密码"表述使用了与 D4 不兼容的术语，这两行应同时修正。

---

## 9. 审查结论

```
【评分】🟢 好品味
【结论】允许进入 iterate 阶段
【依据】
  1. 三份核心文档的设计基线和实施计划基线完全一致，无派生状态冒充风险
  2. 8 个执行单元完整定义了目标、文件边界、依赖、验证和 review gate
  3. TDD 验证矩阵完整覆盖 5 个 B-ID
  4. 架构分层符合 Types → Config → Repo → Service → Runtime → UI 约束
  5. 无 Critical 阻塞问题
  6. 3 个 Important 发现均不阻塞执行——implementation.md 已正确遵循 D1-D5 决策
```

### iterate 启动前建议

在 `harness-ce-iterate` 启动前，按优先级：
1. 将 design.md 降级分析表的"主密码"错误修正（Important #2），然后重新计算设计基线并同步 implementation.md/PROGRESS.md
2. 或将这 3 个 Important 发现记录为已知偏差，在 iterate 过程中通过 plan 修订逐步修复，不影响 U1 启动

选择方案 1 更干净（修改量小，只涉及降级分析表 2 行文本），选择方案 2 也可接受（偏差已在本报告中记录，可追溯）。
