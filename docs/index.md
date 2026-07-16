---
title: ssh-skill 文档索引
description: ssh-skill 的文档导航枢纽，按角色和主题组织所有项目文档的入口
doc_type: reference
last_updated: 2026-07-15
audience: [所有开发者, AI Agent]
---

# ssh-skill 文档索引

安全 SSH 远程操作 CLI 工具 — 为 AI agent（Claude Code）和开发者提供加密凭证存储、目标服务器校验和可审计的命令执行通道。

> 项目使用 [Harness CE](https://github.com/anthropics/harness-ce) 管理任务工作流和知识库。`.harness/` 下的文档、规则和任务状态是项目治理的事实源，`docs/` 是面向人类开发者的可导航知识库。

## 文档目录树

```text
docs/
├── index.md              # 你在这里
├── getting-started.md    # 安装、配置、5 分钟上手
├── cli-reference.md      # 全部 CLI 子命令和参数参考
├── security.md           # 安全模型、加密方案、威胁分析
├── architecture.md       # 项目分层架构和数据流
└── guides.md             # Claude Code 集成、部署、排错指南
```

## 文档治理边界

- `docs/`：面向人类开发者的项目知识 — CLI 参考、架构说明、操作指南
- `.harness/knowledge/`：AI agent 可消费的结构化知识 — 架构约束、模式、依赖策略
- `.harness/tasks/`：任务工作区 — 设计文档、实施计划、进度指针（阶段性产物）
- `README.md`：项目首页 — 安装和快速入门的精简版本
- `AGENTS.md`：AI agent 入口 — 工具链、工作流程、验证命令

## 按角色推荐阅读路径

### 新用户（5 分钟上手）

1. 先读本页，建立文档地图认知
2. 仓库根目录 [`README.md`](../README.md) — 复制「一键安装提示词」，让 Claude Code / Codex 自动安装（无需先 clone）
3. [`getting-started.md`](./getting-started.md) — 安装路径说明、初始化 vault，执行第一条命令
4. [`cli-reference.md`](./cli-reference.md) — 了解所有可用命令
5. [`security.md`](./security.md) — 理解你的凭证如何被保护

### 日常使用者

1. [`cli-reference.md`](./cli-reference.md) — 命令参数速查
2. [`guides.md`](./guides.md) — Claude Code 集成、部署到服务器、排错
3. [`security.md`](./security.md) — 审计日志、密钥轮换

### 维护者 / 贡献者

1. [`architecture.md`](./architecture.md) — 分层依赖、内部包结构
2. `.harness/knowledge/ARCHITECTURE.md` — AI agent 遵守的架构约束
3. `.harness/knowledge/PATTERNS.md` — 已验证模式与反模式
4. [`security.md`](./security.md) — 加密实现细节
5. [`guides.md`](./guides.md) — 构建、测试、发布流程

### AI Agent

1. `AGENTS.md` — 环境验证和任务工作流
2. `.harness/knowledge/ARCHITECTURE.md` — 分层约束
3. `.harness/knowledge/DEPENDENCIES.md` — Go 依赖清单和升级策略
4. [`cli-reference.md`](./cli-reference.md) — 命令签名
5. [`architecture.md`](./architecture.md) — 数据流和模块职责

## 规划、设计与审查入口

项目使用 Harness CE 管理任务生命周期。以下说明各入口的使用时机：

### 何时看 `.harness/tasks/`

- 需要了解某个功能的设计决策 → 查找对应 `design.md`
- 需要了解实现方式和验证结果 → 查找对应 `implementation.md`
- 需要恢复中断的任务 → 读取对应 `PROGRESS.md`
- 需要审查文档包是否一致 → 检查 `reviews/`

### 何时看 `.harness/knowledge/`

- 修改架构 → 先读 `ARCHITECTURE.md` 确认分层约束
- 引入新模式 → 检查 `PATTERNS.md` 是否已有记录
- 升级依赖 → 参考 `DEPENDENCIES.md` 的升级策略

## 关键概念速查

| 概念 | 位置 |
| --- | --- |
| Vault（凭证保险库） | [`security.md#vault-加密存储`](./security.md#vault-加密存储) |
| AES-256-GCM | [`security.md#加密方案`](./security.md#加密方案) |
| Argon2id 密钥派生 | [`security.md#加密方案`](./security.md#加密方案) |
| 审计日志（JSONL） | [`security.md#审计日志`](./security.md#审计日志) |
| 分层架构（UI → CLI → Service → Config → Types） | [`architecture.md#分层结构`](./architecture.md#分层结构) |
| MCP Server 模式 | [`cli-reference.md#serve`](./cli-reference.md#serve) |
| 目标服务器校验 | [`security.md#目标校验`](./security.md#目标校验) |

## 文档约定

- **CLI 参考**：以 `ssh-skill --help` 输出为权威来源，本文档提供中文阅读层和补充说明
- **架构文档**：描述"是什么、为什么这样分层"，不重复代码细节
- **操作指南**：面向任务，回答"怎么做"，提供可复制执行的命令
- **安全文档**：记录威胁模型、加密方案和设计决策，回答"为什么安全"
- **Harness CE 任务文档**：为阶段性产物，保留日期标识，不作为长期规范文档
