# 智能审查清单

## 审查原则
1. 只检查约束和风险，不做主观审美评判
2. 以阻塞问题、行为回归和缺失验证为优先
3. 审查意见应能明确追溯到架构、模式或接口约束

## 检查清单
- [ ] 是否违反分层依赖
- [ ] 是否命中已知反模式
- [ ] 是否遗漏了可复用模式
- [ ] 类型或术语变更是否同步更新 `GLOSSARY.md`
- [ ] 是否需要补充测试、lint 或文档约束
- [ ] `implementation.md` 的知识沉淀和文档防腐记录是否仍有 `待落库`、`待确认` 或 `已阻塞`
- [ ] `design.md`、`implementation.md`、`PROGRESS.md` 的设计基线是否一致
- [ ] `implementation.md`、`PROGRESS.md` 和当前有效 reviews 的实施计划基线是否一致
- [ ] 执行阶段完成时，`implementation.md`、`PROGRESS.md` 和当前有效 reviews 的执行证据基线是否一致
- [ ] `implementation.md` 的执行单元计划是否只包含目标、文件边界、依赖、执行姿态、验证和 review gate，且已纳入实施计划基线
- [ ] `design.md` 是否包含 `行为验收契约`，且每个用户可观察行为或兼容性承诺都有 B-ID
- [ ] `implementation.md` 的 `TDD 验证矩阵` 是否覆盖所有 B-ID，且姿态为 `test-first` / `characterization-first` / `direct-no-behavior`
- [ ] 行为变更是否有 closed RED/GREEN 证据，重构或迁移是否有 characterization 证据
- [ ] `direct-no-behavior` 是否只用于 `no-behavior`；行为或代码运行结果不得用它绕过测试
- [ ] TDD 例外是否位于 `TDD 例外审批`，字段完整、状态已批准，并进入剩余风险
- [ ] 完成前是否运行 `.harness/tools/validate-harness-context.py`
- [ ] `PROGRESS.md` 的每个执行单元状态是否为 `done`，或 grouped review 是否明确列出覆盖单元
- [ ] Critical / Important review 发现是否已关闭；未关闭时不得标记完成
- [ ] `implementation.md` 的文档防腐记录是否和 `design.md` 的文档审查计划一致
- [ ] 当前任务直接引入的文档影响是否已闭合；历史腐烂项是否已明确转交 doc-gardening

## 审查意见格式
- `[约束来源] 建议：内容`
- `[约束来源] 阻塞：内容`
