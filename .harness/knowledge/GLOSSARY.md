# 术语表

## Vault（凭证保险库）
- **定义**：以 AES-256-GCM 加密存储所有 ServerConfig 的容器，落盘为 `~/.ssh-mcp/servers.json.age`，由 32 字节随机 vault key 解密。
- **所属层**：Repo / Service
- **相关文件**：`go/internal/vault/vault.go`、`go/internal/vault/storage.go`、`go/internal/types/types.go`
- **备注**：密码字段在 vault 内是 hex 编码的密文，仅在 cli 层 `resolveServer()` 调用时解密为 in-memory 明文。

## ServerConfig
- **定义**：单台远程服务器的完整配置（ID、Host、Port、User、Auth、可选 Bastion、Tags），是 vault 内的可持久化实体。
- **所属层**：Types
- **相关文件**：`go/internal/types/types.go`

## AuthMethod
- **定义**：SSH 认证方式枚举，取值 `password` / `key` / `agent`，由 cli 层 flag `--auth-type` 选择。
- **所属层**：Types
- **相关文件**：`go/internal/types/types.go`、`go/internal/ssh/client.go`

## BastionConfig
- **定义**：跳板机（jump host）配置，含独立 Host/Port/User/Auth。`ssh.Client` 包装类型通过 `bastion` 字段持有其生命周期，Close() 时一并关闭。
- **所属层**：Types
- **相关文件**：`go/internal/types/types.go`、`go/internal/ssh/client.go`
- **备注**：类型已定义但 CLI 未暴露添加入口，需手动编辑 vault JSON 后重新加密。

## Client（SSH 客户端包装）
- **定义**：`go/internal/ssh` 中的包装类型，嵌入 `*ssh.Client` 并持有可选 bastion 引用，确保 bastion 连接随主连接 Close() 一起释放，避免 GC finalizer 提前关闭。
- **所属层**：Service
- **相关文件**：`go/internal/ssh/client.go`

## Argon2id
- **定义**：从 vault master key 派生 AES 密钥的 KDF 算法，参数 `time=3, memory=64MB, threads=4`，抗 side-channel 与 GPU 暴力破解。
- **所属层**：Repo
- **相关文件**：`go/internal/vault/vault.go`

## AuditEntry
- **定义**：单条 `exec` 命令的审计记录，以 JSONL 追加写入 `~/.ssh-mcp/audit.log`。字段含 timestamp、server_id、server_host、command、exit_code、stdout_len、stderr_len、duration_ms。
- **所属层**：Types
- **相关文件**：`go/internal/types/types.go`、`go/internal/audit/audit.go`、`go/internal/cli/exec.go`

## 行为契约（B-ID）
- **定义**：Harness CE 任务治理中，每个用户可观察行为或兼容性承诺的唯一标识符，用于在 design.md / implementation.md / PROGRESS.md 之间绑定设计基线与执行证据。
- **所属层**：治理 / 跨层
- **相关文件**：`.harness/templates/tasks/design.md`、`.harness/tools/validate-harness-context.py`

## 设计基线 / 实施计划基线 / 执行证据基线
- **定义**：Harness CE 三份任务状态文件（design.md / implementation.md / PROGRESS.md）各自计算的内容指纹；三者必须一致，否则任务不得标记完成。
- **所属层**：治理
- **相关文件**：`.harness/knowledge/PATTERNS.md`（设计基线派生状态模式）
