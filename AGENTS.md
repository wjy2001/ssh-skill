# AGENTS.md

## 你是本项目的 AI 开发伙伴

你的目标是：在系统约束内，将我的意图转化为高质量、可验证的工程产出。

## 项目结构约定

- 本文件（`AGENTS.md`）是项目指令入口，Claude Code/Codex 自动加载；`CLAUDE.md` 可通过 `@AGENTS.md` 引用。
- Harness CE 执行流程由插件全局 skills 承载：`harness-ce-plan` 产出/校准预执行文档包，`harness-ce-iterate` 只按已审核文档包执行；本项目只保存知识库和任务状态。
- 任务工作区在 `.harness/tasks/<YYYY-MM-DD>-<slug>/`，每个任务目录含 `design.md`、`implementation.md`、`PROGRESS.md`，并可包含 `reviews/` 和 `archived/`。
- 项目级知识：
  - `.harness/knowledge/ARCHITECTURE.md`：架构约束与分层依赖
  - `.harness/knowledge/PATTERNS.md`：已验证模式与反模式
  - `.harness/knowledge/GLOSSARY.md`：术语表
  - `.harness/knowledge/CONFIG.md`：配置策略
  - `.harness/knowledge/DEPENDENCIES.md`：依赖策略
- 编排与审查规则：
  - `.harness/agents/ORCHESTRATOR.md`：编排与委托策略
  - `.harness/agents/REVIEWER.md`：审查清单
  - `.harness/agents/TESTER.md`：测试行为规范
  - `.harness/contacts/REVIEWERS.yaml`：审查者配置

## 环境

> 以下内容由项目填写，声明本项目的具体环境。agent 每次任务开始前按此验证。

每次任务开始前，先验证以下环境可用：

- **包管理器**：Go modules（`go mod tidy`）
- **运行时版本**：Go 1.25+（参考 `go/go.mod`）
- **锁定文件**：`go/go.sum`
- **安装命令**：`cd go && go mod download`

如果安装失败，先修复环境再继续任务。如果以上项目未填写，询问项目负责人补全。

## 验证

> 以下内容由项目填写，声明本项目的验证命令。agent 每次改动后必须执行。

每次改动后，必须运行以下验证命令。结果记录在任务 `implementation.md` 的验证结果表中：

- **测试**：`cd go && go test ./... -v`
- **类型检查**：`cd go && go vet ./...`
- **Lint**：`cd go && go vet ./...`（Go 标准 vet 替代 lint）
- **完整验证**：`cd go && go vet ./... && go test ./... -v && go build -o ../.claude/skills/ssh-ops/bin/ssh-mcp ./cmd/ssh-mcp/`

如果验证失败，修复后重新验证，通过后才能标记任务完成。如果以上项目未填写，询问项目负责人补全。

## 你的工作流程

1. 收到任务后，先读取 `AGENTS.md`，再读取相关 `.harness/knowledge/*` 和 `.harness/agents/ORCHESTRATOR.md`。
2. 需要设计时使用 `harness-ce-plan`，维护当前任务文档包：`design.md`、`implementation.md` 草案、`PROGRESS.md` 恢复指针和 `reviews/`。
3. 只有绑定文档包已审核后，才使用 `harness-ce-iterate` 执行实现。
4. 主智能体先做任务地图、共享上下文和 ownership 划分。
5. 当平台允许 subagent 时，默认将边界清晰的工作包委托出去；小任务也可以委托。
6. subagent 只回传结构化结果，主智能体负责集成、验证、审查和沉淀判断。

## 意图路由

“文档”不是一个统一目标。先按上下文收敛写入范围：

- 用户要求调整计划、修改设计、调整 plan 产物、调整刚生成的规划文档，或在 planning 上下文说“调整相关文档”：使用 `harness-ce-plan`，只改当前任务文档包。
- 用户要求修正执行记录、修正进度、更新任务状态：只维护当前任务 `implementation.md` / `PROGRESS.md`，不得改 `design.md` 正文；如果需要改变设计正文，回到 `harness-ce-plan`。
- 用户明确要求项目文档、全部文档、知识库、文档防腐或 doc-gardening：才进入项目级文档维护。

`design.md` 是任务设计事实源。`implementation.md` 是实施计划、验证计划、TDD 计划和执行证据；`PROGRESS.md` 是任务恢复指针，不承载完整设计或审查报告。三者都必须绑定同一设计基线。`design.md` 必须包含 `行为验收契约`，每个用户可观察行为或兼容性承诺都有 B-ID。`implementation.md` 还必须记录实施计划基线和执行证据基线：实施计划基线覆盖预执行计划、文件改动计划、验证计划、TDD 验证矩阵和执行单元计划；执行证据基线覆盖执行配置、执行记录、TDD 执行证据、TDD 例外审批、TDD 测试评审、TDD 独立校验、验证结果、执行评审、偏差、知识沉淀、文档防腐和最终状态。设计基线、实施计划基线或执行证据基线不一致时不得继续执行或标记完成。

## 架构约束

- UI → Runtime → Service → Repo → Config → Types
- 上层可引用下层，逆向或跨层引用 CI 自动拒绝。
- 共享类型只能在 Types 层定义，暴露 API 的 Schema。

## 代码模式

- 按需查阅 `.harness/knowledge/PATTERNS.md`。
- 新增通用模式必须补充进该文件。
- 废弃模式标记 `@deprecated` 并说明替代方案。

## 执行计划

- 任务隔离在 `.harness/tasks/<YYYY-MM-DD>-<slug>/` 目录。
- `harness-ce-plan` 创建/校准预执行文档包，不执行实现。
- `harness-ce-plan` 必须把验收标准拆成 B-ID 行为契约，并在 `implementation.md` 中生成 `TDD 验证矩阵`。
- `harness-ce-plan` 每次修改 `design.md` 正文后必须更新设计基线，并在基线变化时把状态重置为 `待讨论` 或 `待完整审查`。
- `harness-ce-plan` 必须主动处理待讨论项：阻塞点、关键风险、关键假设和必须用户选择的方案分支都要逐项确认；低风险项只进入残余风险。
- `harness-ce-plan` 在方向确认后使用派生生成 subagent 生成 `implementation.md` / `PROGRESS.md` 草案，并用新的 subagent 完整审查文档包；完整报告写入 `reviews/`。
- `harness-ce-iterate` 必须读取已审核文档包后才执行，并在 `implementation.md` 中追加执行证据，更新 `PROGRESS.md` 恢复指针。
- `harness-ce-iterate` 启动前必须确认 `design.md`、`implementation.md`、`PROGRESS.md` 和当前有效 `reviews/` 的设计基线与实施计划基线一致；不一致时回到 `harness-ce-plan` 重建派生状态。
- `harness-ce-iterate` 执行时按 B-ID 和执行单元推进：`test-first` 必须 RED 再 GREEN，`characterization-first` 必须改前 GREEN 再改后 GREEN，`direct-no-behavior` 只允许 `no-behavior`。每个单元计划必须有目标、文件边界、依赖、验证和 review gate；每个 gate 通过后**立即同步更新** `PROGRESS.md`（`## 当前指针`、`## 执行单元状态`、`## 最近动作`），不得延迟到单元结束或任务完成才批量写入。Critical / Important review 发现未关闭时，不得把单元标记为 `done`。
- 任务完成前必须运行 `python .harness/tools/validate-harness-context.py`；该 validator 只校验证据结构和基线，不重新运行项目测试命令。
- 任务完成后，检查并更新项目级知识文档以保持项目描述最新。

## 项目文档

- [`docs/index.md`](./docs/index.md) — 文档导航枢纽（按角色、主题的入口地图）
- [`docs/getting-started.md`](./docs/getting-started.md) — 安装和 5 分钟上手教程
- [`docs/cli-reference.md`](./docs/cli-reference.md) — 全部 CLI 子命令和参数参考
- [`docs/security.md`](./docs/security.md) — 安全模型、加密方案、威胁分析
- [`docs/architecture.md`](./docs/architecture.md) — 分层架构和数据流
- [`docs/guides.md`](./docs/guides.md) — Claude Code 集成、部署、排错

## 文档维护

- 必查文档清单：
  - `AGENTS.md`、`CLAUDE.md`
  - `.harness/knowledge/*.md`
  - `.harness/agents/*.md`
  - `.harness/contacts/REVIEWERS.yaml`
  - `.harness/templates/tasks/*.md`
  - `README.md`
  - `docs/**/*.md`（存在时）
- 新增依赖或运行时版本变化时，更新本文件的 `## 环境` 段。
- 修改架构后更新 `ARCHITECTURE.md`。
- 发现模式时，写进 `PATTERNS.md`。
- 创建新计划时，检查 `GLOSSARY.md` 术语是否已定义。
- 修改配置或依赖策略时，分别更新 `CONFIG.md` 或 `DEPENDENCIES.md`。
- 每个任务完成前，`implementation.md` 的“知识沉淀”和“文档防腐记录”不能留下当前任务范围内的 `待落库`、`待确认` 或 `已阻塞`。
- 历史腐烂项可记录为 `已转交 doc-gardening`，不得把当前主任务变成泛文档整理任务。
