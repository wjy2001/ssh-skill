# 实施文档

## 关联

- 设计文档：./design.md
- 进度指针：./PROGRESS.md
- 审查报告：./reviews/

## 派生状态

- **绑定设计基线**：sha256:138126660dcf50849c30e606384f7980534439ff0f6830c91a9f42a4672a02cb
- **实施计划基线**：sha256:cb2137b2f494ff16f6a598cb3172223d33033a84cd56ad320ad9018a23dfaca4
- **执行证据基线**：无
- **派生状态**：有效
- **生成来源**：确定性迁移器
- **生成者**：harness-ce-update
- **生成时间**：2026-07-06 15:31
- **取代原因**：无

## 实施计划

- [x] 扫描 Harness CE 托管文件和用户知识文件
- [ ] LLM 读取 current / baseline / candidates，生成最小语义补丁
- [ ] 人工确认 LLM 补丁或候选文件
- [ ] 应用必要补丁并重新运行 harness-ce-update

## 文件改动计划

| 文件 | 动作（新增/修改/删除） | 原因 | 对应设计章节 |
|------|------------------------|------|--------------|
| candidates/ | 新增/修改 | 保存需要人工确认的迁移候选 | LLM 语义迁移边界 |

## 验证计划

| 验证 | 命令/检查 | 预期结果 | 覆盖范围 |
|------|-----------|----------|----------|
| 升级复扫 | harness-ce-update --apply | 无未确认升级差异 | Harness CE 托管文件 |

## 执行单元计划

| 单元 | 目标 | 文件边界 | 依赖 | 执行姿态 | 验证 | review gate |
|------|------|----------|------|----------|------|-------------|
| U1 | 审查并处理迁移候选 | candidates/ 与受影响托管文件 | 无 | direct-no-behavior | 重新运行 harness-ce-update | 规格符合 + 质量通过 |

## 执行配置

- 测试：开启
- 评审：开启
- 沉淀：开启
- 文档防腐：开启

## 执行记录

- [x] [U1] 扫描 Harness CE 托管文件和用户知识文件
- [ ] [U1] LLM 读取 current / baseline / candidates，生成最小语义补丁
- [ ] [U1] 人工确认 LLM 补丁或候选文件
- [ ] [U1] 应用必要补丁并重新运行 harness-ce-update

## 知识沉淀

| 类型（模式/反模式/术语/决策/无） | 目标文档 | 沉淀内容 | 状态（无需/待落库/已落库） | 证据 |
|----------------------------------|----------|----------|---------------------------|------|
| 决策 | .harness/harness-ce.json | Harness CE 升级状态由 migrator 管理 | 已落库 | .harness/harness-ce.json |

## 文档防腐记录

| 类型 | 文件 | 状态 | 处理动作 | 证据 |
|------|------|------|----------|------|
| user | AGENTS.md | 待确认 | 新模板与项目知识不同步；禁止自动覆盖 | D:\project\github\ssh-skill\.harness\tasks\2026-07-06-harness-ce-upgrade\candidates\AGENTS.md.new |
| user | .harness/knowledge/ARCHITECTURE.md | 待确认 | 新模板与项目知识不同步；禁止自动覆盖 | D:\project\github\ssh-skill\.harness\tasks\2026-07-06-harness-ce-upgrade\candidates\.harness__knowledge__ARCHITECTURE.md.new |
| user | .harness/knowledge/PATTERNS.md | 待确认 | 新模板与项目知识不同步；禁止自动覆盖 | D:\project\github\ssh-skill\.harness\tasks\2026-07-06-harness-ce-upgrade\candidates\.harness__knowledge__PATTERNS.md.new |
| user | .harness/knowledge/DEPENDENCIES.md | 待确认 | 新模板与项目知识不同步；禁止自动覆盖 | D:\project\github\ssh-skill\.harness\tasks\2026-07-06-harness-ce-upgrade\candidates\.harness__knowledge__DEPENDENCIES.md.new |

## LLM 语义迁移记录

| 文件 | 动作 | 理由 | 状态 | 证据 |
|------|------|------|------|------|
| `AGENTS.md` | 最小编辑（意图路由 + 执行计划段落） | 新增 B-ID、TDD 验证矩阵、test-first/characterization-first/direct-no-behavior、validator 工具要求；保留项目环境/验证段 | ✅ 已应用 | 两句 Edit |
| `.harness/templates/tasks/design.md` | 替换为 candidate | 内容完全相同，仅建立 baseline | ✅ 已应用 | cp candidate |
| `.harness/templates/tasks/implementation.md` | 替换为 candidate | 模板注释新增 TDD 基线说明行 | ✅ 已应用 | cp candidate |
| `.harness/templates/tasks/PROGRESS.md` | 替换为 candidate | 模板新增 B-ID、TDD 状态、TDD gate 列 | ✅ 已应用 | cp candidate |
| `.harness/agents/ORCHESTRATOR.md` | 替换为 candidate | 执行单元包新增行为ID、TDD gate 字段 | ✅ 已应用 | cp candidate |
| `.harness/agents/REVIEWER.md` | 替换为 candidate | 检查清单新增 6 条 TDD 相关项 | ✅ 已应用 | cp candidate |
| `.harness/agents/TESTER.md` | 替换为 candidate | 完整 TDD 重写（B-ID、RED/GREEN、例外审批） | ✅ 已应用 | cp candidate |
| `.harness/knowledge/ARCHITECTURE.md` | 不修改 | 当前含 SSH 专项架构，覆盖模板且更详细 | ✅ 无需操作 | — |
| `.harness/knowledge/PATTERNS.md` | 追加 2 个模式 | 追加"行为契约驱动完成门禁"和"用 direct 绕过行为测试"反模式 | ✅ 已应用 | 两句 Edit |
| `.harness/knowledge/DEPENDENCIES.md` | 不修改 | 当前含 Go 依赖清单，覆盖模板且更详细 | ✅ 无需操作 | — |

## TDD 迁移任务

- **发现问题**：已完成任务 `2026-07-02-ssh-remote-ops` 的执行证据基线为 `无`，缺少 TDD 执行证据和 TDD 例外审批
- **迁移任务**：`.harness/tasks/2026-07-06-ssh-remote-ops-tdd-migration/`
- **状态**：已阻塞（需人工补齐证据或批准例外）
- **不阻塞**：当前 Harness CE 升级任务本身
