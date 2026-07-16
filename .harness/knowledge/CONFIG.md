# 配置约定

## 使用原则
- 只记录长期稳定的配置策略，不手工复制完整环境变量清单
- 新增配置项时，同步说明默认值、来源、作用范围和失败行为
- 涉及密钥、令牌、凭据的配置只描述名称和用途，不写真实值

## 配置项

### `SSH_SKILL_CONFIG_DIR`
- **来源**：环境变量
- **默认值**：`~/.ssh-skill/`
- **所属层**：Config
- **影响范围**：决定 vault key、加密 servers 文件、audit log 的存储目录；所有 cli 子命令启动时由 `config.Dir()` 解析。
- **失败行为**：环境变量为空字符串时回退到默认值；`os.UserHomeDir()` 失败时 cli 启动报错退出。
- **相关文件**：`go/internal/config/config.go`
- **设计取舍**：**有意不提供命令行 `--config-dir` flag**，避免每个子命令都要重复解析；用户/CI 通过 export 环境变量一次性设定即可。

### 文件路径派生（由 `SSH_SKILL_CONFIG_DIR` 计算）
- `~/.ssh-skill/.vault-key`：32 字节随机 AES-256 密钥，权限 `0600`
- `~/.ssh-skill/servers.json.age`：AES-256-GCM 加密的服务器配置，权限 `0600`（`.age` 仅为命名约定，非 age 工具格式）
- `~/.ssh-skill/audit.log`：JSONL 审计日志，明文，权限 `0600`，追加写入

### `SSH_AUTH_SOCK`
- **来源**：环境变量（由 ssh-agent 设置）
- **默认值**：无
- **所属层**：Repo（间接，由 `golang.org/x/crypto/ssh/agent` 消费）
- **影响范围**：仅当 ServerConfig 的 `Auth.Method == "agent"` 时被读取；缺失则 `buildAgentAuth()` 返回 `ErrAuthNotConfigured`。
- **失败行为**：未设置或 socket 不可达时，add/exec/test/upload/download 报错退出。
- **相关文件**：`go/internal/ssh/client.go`

### `version`（构建时注入）
- **来源**：`go build -ldflags "-X main.version=v0.1.0"`（CI 通过 git tag 注入）
- **默认值**：`"dev"`（源码直接 `go build` 时的占位值）
- **所属层**：Runtime
- **影响范围**：仅由 `ssh-skill --version` 输出，不影响任何运行时行为。
- **相关文件**：`go/cmd/ssh-skill/main.go`、`go/internal/cli/root.go`

## 变更要求
1. 配置项改名、删除或语义变化时，必须更新本文档
2. 影响运行时行为的配置变更需要补充测试或检查脚本
3. 破坏兼容性的配置变更需要在计划文档中记录迁移方案
