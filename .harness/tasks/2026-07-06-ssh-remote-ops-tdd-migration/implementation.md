# SSH 远程操作 — TDD 迁移实施文档

## 关联

- 设计文档：`./design.md`
- 进度指针：`./PROGRESS.md`
- 审查报告：`./reviews/`

## 派生状态

- **绑定设计基线**：sha256:<from-design>
- **实施计划基线**：sha256:<from-plan>
- **执行证据基线**：无
- **派生状态**：有效
- **生成来源**：harness-ce-update LLM 语义审查
- **生成者**：main agent
- **生成时间**：2026-07-06

## 实施计划

本任务不是执行实现，而是回溯补齐已通过测试的 TDD 执行证据。所有测试已在 ssh-remote-ops 任务中通过。

## 文件改动计划

| 文件 | 动作 | 原因 |
|------|------|------|
| `.harness/tasks/2026-07-02-ssh-remote-ops/implementation.md` | 修改 | 追加 TDD 执行证据区域 |
| `.harness/tasks/2026-07-02-ssh-remote-ops/PROGRESS.md` | 修改 | 更新执行证据基线 |

## 验证计划

| 验证 | 命令/检查 | 预期结果 |
|------|-----------|----------|
| Go 测试通过 | `cd go && go test ./... -v` | 21/21 PASS |
| 上下文校验 | `python .harness/tools/validate-harness-context.py` | 通过 |

## 执行单元计划

| 单元 | 目标 | 姿态 | 文件边界 |
|------|------|------|----------|
| U1 | 补齐 B-001 到 B-005 的 TDD 执行证据 | direct-no-behavior（纯文档回溯，不改变代码行为） | `.harness/tasks/2026-07-02-ssh-remote-ops/implementation.md` |
| U2 | 运行 validator 确认证据基线闭合 | direct-no-behavior | `.harness/tools/validate-harness-context.py` |

## TDD 验证矩阵

| B-ID | 行为 | U-ID | 姿态 |
|------|------|------|------|
| B-001 | vault init | U1 | characterization-first（测试先于文档补全存在） |
| B-002 | server CRUD | U1 | characterization-first |
| B-003 | SSH exec | U1 | characterization-first |
| B-004 | file transfer | U1 | characterization-first |
| B-005 | audit log | U1 | characterization-first |

## 执行记录

| U-ID | 操作摘要 | 状态 |
|------|---------|------|
| U1 | 待执行 — 补齐 TDD 执行证据 | pending |
| U2 | 待执行 — 运行 validator | pending |

## 知识沉淀

| 类型 | 目标文档 | 沉淀内容 | 状态 |
|------|----------|----------|------|
| 无 | - | 本任务是迁移补齐，不产出新模式 | 无需 |
