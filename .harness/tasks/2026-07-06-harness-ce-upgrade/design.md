# 设计文档

## 任务标识

- **任务名**：harness-ce-upgrade
- **日期**：2026-07-06

## 设计状态

### 基线

- **设计基线**：sha256:138126660dcf50849c30e606384f7980534439ff0f6830c91a9f42a4672a02cb
- **取代基线**：无

### 流程状态

- **状态**：已阻塞
- **待讨论项**：无
- **阻塞项**：存在需要人工确认的升级差异

### 派生状态

- **implementation.md**：有效
- **PROGRESS.md**：有效
- **有效审查报告**：无

### 审批状态

- **审批对象**：design.md + implementation.md + PROGRESS.md + 当前有效 reviews 摘要
- **实施计划基线**：sha256:cb2137b2f494ff16f6a598cb3172223d33033a84cd56ad320ad9018a23dfaca4
- **执行证据基线**：无
- **审批来源**：确定性迁移器
- **审批时间**：2026-07-06 15:31

## 决策

`	ext
【判断】✅ 值得做
【洞察】流程模板可以迁移，项目知识不能静默覆盖
【方案】确定性迁移器产出候选 -> LLM 做最小语义补丁 -> 人工确认后闭合状态
`

## 文档影响

- **需要更新**：见 PROGRESS.md 的迁移记录
- **依据**：Harness CE 1.0.0 模板与目标项目存在差异
- **防腐检查**：托管文件 baseline、项目内 skill 退场、knowledge drift

## LLM 语义迁移边界

1. 以 .harness/harness-ce.json、baseline 和 candidates 为事实源
2. managed / semi-managed 文件可做最小三方语义合并
3. user 文件不得整体替换，只能追加或最小修改明确属于 Harness CE 的长期流程内容
4. 意图不明确时保持 待确认 或 已阻塞

## 执行状态影响

- **implementation.md**：有效
- **PROGRESS.md**：有效
- **原因**：迁移器生成同一设计基线的派生状态
- **执行限制**：这是升级迁移候选文档包，未经过 harness-ce-plan 完整审查，不得直接交给 harness-ce-iterate 执行
