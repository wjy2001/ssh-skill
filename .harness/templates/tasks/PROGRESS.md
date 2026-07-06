# 任务进度

<!--
  PROGRESS.md 是任务恢复指针，不是设计文档、实施方案或审查报告。
  它只回答：现在到哪了、接手后先做什么、什么不能做。
  建议保持在 50-100 行内。
-->

## 当前指针

- **任务**：[slug]
- **阶段**：待讨论 / 方向已确认 / 派生生成中 / 派生文档已生成 / 待完整审查 / 已审核 / 执行中 / 已完成 / 已阻塞
- **绑定设计基线**：sha256:<from-design>
- **绑定实施计划基线**：sha256:<from-plan> / 无
- **绑定执行证据基线**：sha256:<from-evidence> / 无
- **当前执行单元**：无 / U1 / U2 / ...
- **当前行为**：无 / B1 / B2 / ...
- **TDD 状态**：未开始 / RED / GREEN / refactor / exception-pending / closed
- **执行引擎**：未开始 / inline / serial-subagent / parallel-subagent
- **最后更新时间**：YYYY-MM-DD HH:MM
- **下一步**：[下一步动作]
- **禁止事项**：[例如：不得执行实现 / 不得使用过期 implementation.md / 不得跳过完整审查]

## 当前阻塞

- 无 / D1 / D2

## 执行单元状态

| 单元 | 状态 | 当前阻塞 | 最近验证 | TDD gate | Review gate |
|------|------|----------|----------|----------|-------------|
| - | pending / running / implemented / reviewed / done / blocked / needs-plan | 无 / ID | 无 / 命令摘要 | 未开始 / RED / GREEN / closed / exception-pending | 未开始 / 通过 / 未关闭 |

## 最近动作

<!-- 只保留最近 3-5 条，完整过程看 design.md / implementation.md / reviews/。 -->

- YYYY-MM-DD HH:MM - [动作摘要]

## 交接说明

- **先读**：`design.md`、`implementation.md`、当前有效 `reviews/*.md`
- **必须检查**：设计基线、实施计划基线、执行证据基线、设计状态、待讨论项、有效审查报告、执行单元状态、当前 B-ID、TDD gate
- **不要做**：[明确禁止事项]

## 参考文件

- 设计文档：`./design.md`
- 实施文档：`./implementation.md`
- 审查报告：`./reviews/`
- 旧实施归档：`./archived/`
