---
title: ssh-skill 文档索引
description: ssh-skill 的文档导航枢纽，按角色和主题组织所有项目文档的入口
doc_type: reference
last_updated: 2026-07-15
audience: [所有开发者, AI Agent]
---

# ssh-skill 文档索引

安全 SSH 远程操作 CLI 工具 — 为 AI agent（Claude Code）和开发者提供加密凭证存储、目标服务器校验和可审计的命令执行通道。

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
- `README.md`：项目首页 — 安装和快速入门的精简版本

## 按角色推荐阅读路径

### 新用户（5 分钟上手）

1. 先读本页，建立文档地图认知
2. 仓库根目录 [`README.md`](../README.md) — 复制短提示词；agent 读取 [`install/PROMPT.md`](../install/PROMPT.md) 做最小安装（只取 skill + 当前平台二进制）
3. [`getting-started.md`](./getting-started.md) — 安装路径说明、初始化 vault，执行第一条命令
4. [`cli-reference.md`](./cli-reference.md) — 了解所有可用命令
5. [`security.md`](./security.md) — 理解你的凭证如何被保护

### 日常使用者

1. [`cli-reference.md`](./cli-reference.md) — 命令参数速查
2. [`guides.md`](./guides.md) — Claude Code 集成、部署到服务器、排错
3. [`security.md`](./security.md) — 审计日志、密钥轮换

### 维护者 / 贡献者

1. [`architecture.md`](./architecture.md) — 分层依赖、内部包结构
2. [`security.md`](./security.md) — 加密实现细节
3. [`guides.md`](./guides.md) — 构建、测试、发布流程

### AI Agent

1. [`cli-reference.md`](./cli-reference.md) — 命令签名
2. [`architecture.md`](./architecture.md) — 数据流和模块职责
3. [`security.md`](./security.md) — 审计日志、凭证保护

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
