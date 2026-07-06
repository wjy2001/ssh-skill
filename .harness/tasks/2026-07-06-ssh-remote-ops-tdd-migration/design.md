# SSH 远程操作 — TDD 迁移任务

## 任务标识

- **任务名**：ssh-remote-ops-tdd-migration
- **日期**：2026-07-06
- **来源**：Harness CE 升级扫描 — 已完成任务 `2026-07-02-ssh-remote-ops` 缺少 TDD 执行证据
- **关联任务**：`.harness/tasks/2026-07-02-ssh-remote-ops/`

## 快速状态

- **设计基线**：见"设计状态"
- **核心决策**：为已完成任务补 TDD 执行证据基线，不得伪造 RED/GREEN 证据
- **当前阶段**：已阻塞（需人工补充证据）

## 设计状态

### 基线

- **设计基线**：sha256:tbd
- **取代基线**：无

### 流程状态

- **状态**：已阻塞
- **待讨论项**：无
- **阻塞项**：需要人工从已完成测试日志中提取 TDD 执行证据，或批准例外

### 派生状态

- **implementation.md**：有效
- **PROGRESS.md**：有效
- **有效审查报告**：无

### 审批状态

- **审批对象**：design.md + implementation.md + PROGRESS.md
- **实施计划基线**：无
- **执行证据基线**：无
- **审批来源**：harness-ce-update 确定性扫描
- **审批时间**：2026-07-06

## 背景和问题

Harness CE 1.0.0 升级扫描发现：已完成任务 `.harness/tasks/2026-07-02-ssh-remote-ops/` 的执行证据基线为空。

### 已有项（无需补）

- ✅ `design.md` 包含 `行为验收契约`（B-001 到 B-005）
- ✅ `implementation.md` 包含 `TDD 验证矩阵`（B-ID → U-ID 映射）
- ✅ 所有 8 个 U-ID 标记 `done`，测试通过（21/21 PASS，6/6 e2e PASS）
- ✅ `implementation.md` 包含验证结果表和知识沉淀记录

### 缺失项（本任务处理范围）

- ❌ `TDD 执行证据`：虽然测试全部通过，但没有按 B-ID 记录 RED/GREEN 证据
- ❌ `TDD 例外审批`：无（如果部分行为确实无法自动化测试）
- ❌ 执行证据基线为 `无`

## 决策

```text
【判断】✅ 值得做
【洞察】测试已通过但证据未按 B-ID 结构化 — 需要回溯补齐而非伪造
【方案】人工从已有测试日志/代码中提取每个 B-ID 的证据，或对无法回溯的行为批准例外
```

## 行为验收契约（引用）

| B-ID | 行为 | 关联 U-ID | 测试存在？ |
|------|------|-----------|-----------|
| B-001 | vault init 创建密钥 | U3 | 是（11/11 vault tests） |
| B-002 | 添加/列出/删除服务器 | U5 | 是（CLI CRUD 验证） |
| B-003 | SSH 远程执行命令 | U4 | 是（7/7 ssh tests） |
| B-004 | 文件上传/下载 | U4 | 是（ssh tests 含 SFTP） |
| B-005 | 审计日志写入 | U6 | 是（3/3 audit tests） |

## 残余风险

| ID | 风险 | 影响 | 接受理由 | 后续观察 |
|----|------|------|----------|----------|
| R1 | 无法回溯 RED 阶段证据 | TDD 链不完整 | 任务已完成，测试全绿；可降级为例外审批 | 后续任务严格按 TDD 流程 |

## 验收标准

- [ ] `implementation.md` 的 TDD 执行证据区域按 B-ID 补齐（含测试命令、输出摘要、通过状态）
- [ ] 每个 B-ID 标记姿态（test-first/characterization-first/direct-no-behavior）
- [ ] 无法回溯的 B-ID 创建 TDD 例外审批（含理由和批准状态）
- [ ] 执行证据基线更新为非空值
- [ ] `python .harness/tools/validate-harness-context.py` 通过

## 绑定文档包

- 实施文档：`./implementation.md`
- 进度指针：`./PROGRESS.md`
