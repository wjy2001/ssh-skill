# 代码模式与反模式

## ✅ 已验证模式

### 流程闸门闭环

- **何时使用**：设计 AI agent 工作流时，避免流程只停在“执行完成”的口头状态。
- **结构**：`加载项目契约 -> 制定设计 -> 审核 design.md -> 执行变更 -> 记录可重复验证 -> 按规则审查 -> 修复阻塞问题 -> 沉淀为可审查变更`
- **示例**：全局 `harness-ce-plan`、全局 `harness-ce-iterate`
- **来源**：把设计评审、测试、评审、沉淀作为全局 skill-first 工作流。
- **版本**：1.1.0

### 文档防腐闭环

- **何时使用**：任务会改变架构、配置、依赖、接口、术语或长期工程约束时。
- **结构**：`识别文档影响 -> 同步更新知识库 -> 检查引用和术语 -> 写入 implementation.md 防腐状态 -> reviewer 阻止未闭合状态`
- **示例**：`.harness/tasks/*/PROGRESS.md`、`.github/workflows/doc-gardening.yml`
- **来源**：将文档保鲜从“定时提醒”升级为可追踪任务状态和 PR 闭环。
- **版本**：1.1.0

### 设计基线派生状态

- **何时使用**：任务有 `design.md`、`implementation.md`、`PROGRESS.md` 三份状态文件时。
- **结构**：`design.md 计算设计基线 -> implementation.md 计算实施计划基线 -> iterate 追加执行证据基线 -> PROGRESS.md 绑定三类基线 -> 不一致则重建或取代派生状态`
- **示例**：`.harness/tasks/*/design.md` 的 `## 设计状态`、`.harness/tasks/*/implementation.md` 的 `## 派生状态`、`.harness/tasks/*/PROGRESS.md` 的 `## 当前指针`
- **来源**：防止 plan 修订后旧执行记录继续冒充当前状态。
- **版本**：1.1.0

### 执行单元 Review Gate

- **何时使用**：执行已审核文档包，尤其是多文件、多阶段或可并行的任务。
- **结构**：`执行单元 U-ID -> bounded handoff packet -> 单元验证 -> task review(规格符合 + 质量结论) -> fix/re-review -> final review`
- **示例**：`.harness/tasks/*/implementation.md` 的 `## 执行单元计划`、`## 执行评审`，以及 `PROGRESS.md` 的 `## 执行单元状态`
- **来源**：迁移 Compound Engineering / Superpowers 的执行单元、review gate 和 ledger 思路，但保留 Harness CE 的 design baseline 数据结构。
- **版本**：1.0.0

### 行为契约驱动完成门禁

- **何时使用**：任务完成状态需要被机器校验，而不是靠 agent 自述。
- **结构**：`B-ID 行为验收契约 -> TDD 验证矩阵 -> RED/GREEN 执行证据 -> TDD 例外审批 -> 本地 validator`
- **示例**：`.harness/tasks/*/design.md` 的 `## 行为验收契约`、`.harness/tasks/*/implementation.md` 的 `## TDD 验证矩阵` / `## TDD 执行证据`，以及 `.harness/tools/validate-harness-context.py`
- **来源**：把完成判断从"文档账本闭合"升级为"行为契约和 TDD 证据闭合"。
- **版本**：1.0.0

### 模式名：[简短描述]

- **何时使用**：适用场景
- **结构**：代码骨架或设计图
- **示例**：项目中的实际用例（可引用文件路径）
- **来源**：从哪个问题/需求沉淀而来
- **版本**：语义化版本（该模式发生重大变化时递增）

---

## ⚠️ 反模式（危险区）

### 反模式名：软闸门完成

- **表现**：只要求“自检”“建议沉淀”或“目录存在”，没有留下可重复验证证据、阻塞判断或可审查变更。
- **错误示例**：实现后直接输出最终结果，没有记录验证命令和结果；安装后只检查目录存在，漏掉全局 `harness-ce-plan` / `harness-ce-iterate` 是否可用；未审核 `design.md` 就开始执行。
- **为什么危险**：流程看似完整，但质量、兼容性和知识复利都不可复查。
- **正确替代**：使用“流程闸门闭环”模式。
- **检测方法**：review 检查是否存在验证摘要、阻塞问题处理记录、具体文件验收和沉淀落库变更。

### 反模式名：只报告不防腐

- **表现**：定时任务只打印过期文档或坏引用，不创建任务、不修复、不阻止完成。
- **错误示例**：CI 输出 “Harness knowledge files older than 30 days” 后仍然成功结束，没有任何 `PROGRESS.md` 或 PR。
- **为什么危险**：文档腐烂会被重复发现，但没有所有者、状态和关闭条件。
- **正确替代**：使用“文档防腐闭环”模式。
- **检测方法**：检查是否存在 doc-gardening 任务、PR、状态表和 reviewer gate。

### 反模式名：派生状态冒充事实源

- **表现**：`design.md` 已被修改，但 `implementation.md` / `PROGRESS.md` 仍按旧设计继续执行或标记完成。
- **错误示例**：用户要求“调整相关文档”后只改了 `design.md`，旧 `PROGRESS.md` 仍保留 `已完成`，没有基线失效记录。
- **为什么危险**：执行记录和设计事实源分裂，review 无法判断当前任务到底按哪个设计完成。
- **正确替代**：使用“设计基线派生状态”模式。
- **检测方法**：检查三份任务文档的设计基线是否一致，且设计状态是否已审核。

### 反模式名：单元 Review 缺失

- **表现**：多执行单元任务只做最终汇总 review，没有逐单元记录规格符合、质量结论和 Critical/Important 关闭状态。
- **错误示例**：U1-U4 都修改了不同契约文件，但 `reviews/` 只写一份“整体看起来可以”的最终报告。
- **为什么危险**：后续恢复任务时无法知道哪个单元真正通过、哪个单元只是被最终总结带过。
- **正确替代**：使用“执行单元 Review Gate”模式。
- **检测方法**：review 检查每个 U-ID 是否有独立 review，或 grouped review 是否显式列出覆盖 U-ID 和阻塞项状态。

### 反模式名：用 direct 绕过行为测试

- **表现**：行为变更、兼容性承诺或代码运行结果变化被标成 `direct-no-behavior`，没有 RED/GREEN 或 characterization 证据。
- **错误示例**：修改 validator、workflow 或脚本逻辑后，只在验证结果表写"已检查"，没有 B-ID、TDD 矩阵和执行证据。
- **为什么危险**：任务看起来完成，但用户可观察行为没有被测试锁住，换模型或恢复上下文后容易把假完成当事实。
- **正确替代**：使用"行为契约驱动完成门禁"模式。
- **检测方法**：运行 `.harness/tools/validate-harness-context.py`；review 检查 `direct-no-behavior` 是否只对应 `no-behavior`。

### 反模式名：[简短描述]

- **表现**：出现什么问题
- **错误示例**：不该怎么写的代码
- **为什么危险**：导致的根本后果
- **正确替代**：指向对应的已验证模式
- **检测方法**：是否存在 CI/测试可以阻止此反模式
