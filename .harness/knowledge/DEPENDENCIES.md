# 外部依赖与升级策略

## 使用原则
- 本文记录依赖策略、升级边界和风险，不替代 lockfile 或包管理器清单
- 高风险依赖需要说明替代方案和回滚策略
- 升级依赖时优先小步验证，避免无关的大批量升级

## 依赖清单

### `golang.org/x/crypto`
- **用途**：SSH 客户端连接（password/key/agent 认证）、AES-256-GCM 加解密、Argon2id 密钥派生
- **所属层**：Repo / Service
- **升级策略**：人工审核
- **风险点**：当前锁定 v0.17.0（2023-02，对应 Go 1.18 最低要求）；升级到 v0.21.0+ 会要求 Go 1.20，v0.31.0+ 要求 Go 1.22，v0.53.0+ 要求 Go 1.25。API 稳定性高，breaking change 罕见。
- **替代方案**：无（Go 生态唯一成熟 SSH 和加密实现）
- **验证方式**：`go test ./internal/ssh/...` + `go test ./internal/vault/...`

### `github.com/pkg/sftp`
- **用途**：SFTP 文件上传/下载
- **所属层**：Repo
- **升级策略**：人工审核
- **风险点**：当前锁定 v1.13.5（对应 Go 1.15）；v1.13.10+ 要求 Go 1.23。间接依赖 `golang.org/x/crypto`，需与 x/crypto 版本协同
- **替代方案**：纯 SCP 实现（基于 `x/crypto/ssh`），但 SFTP 更可靠
- **验证方式**：集成测试（in-process SSH server，见 `internal/ssh/exec_test.go`）

### `golang.org/x/sys`
- **用途**：`x/crypto` 的间接依赖
- **所属层**：Repo（间接）
- **升级策略**：跟随 `x/crypto` 自动升级
- **风险点**：低
- **替代方案**：N/A
- **验证方式**：编译通过即可

### Go 运行时
- **用途**：编译和运行 ssh-mcp 二进制
- **所属层**：Runtime
- **升级策略**：跟随 go.mod 声明的最低版本（当前 Go 1.18）
- **风险点**：Go 1.18 于 2022-03 发布，已停止官方支持但生态广泛可用；如需升级 Go 版本，同步升级 `x/crypto` 到匹配版本
- **替代方案**：可尝试降级 `x/crypto` 到 v0.14.0 或更早以支持 Go 1.17，但会丢失 ssh 包安全补丁
- **验证方式**：`go version` + `go build ./cmd/ssh-mcp/`

## 依赖模板

### `[DEPENDENCY_NAME]`
- **用途**：为什么需要这个依赖
- **所属层**：Config / Repo / Service / Runtime / UI
- **升级策略**：自动升级 / 人工审核 / 锁定版本
- **风险点**：兼容性、安全、性能或许可风险
- **替代方案**：可选替代品或移除条件
- **验证方式**：测试、lint、手动检查或 smoke test

## 升级流程
1. 阅读 changelog 或 release notes
2. 评估 API、行为、许可和安全影响
3. 更新 lockfile 后运行相关测试
4. 如升级影响架构或模式，更新 `ARCHITECTURE.md` 或 `PATTERNS.md`
