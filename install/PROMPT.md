# ssh-skill 安装提示词（Agent 执行版）

> 本文件是**权威安装指令**。Claude Code / Codex 应完整读取并执行本文件。  
> 用户只需让 agent 获取本文件最新内容，无需记住具体安装步骤。  
> 稳定入口（README 短提示词会指向这里）：  
> `https://raw.githubusercontent.com/wjy2001/ssh-skill/master/install/PROMPT.md`

## 目标

把 `ssh-skill` 安装为**全局 Claude skill**，且只下载：

1. `SKILL.md`
2. `bin/ssh-skill` —— **平台无关 launcher**（脚本，负责按当前 OS 选择正确二进制）
3. **当前平台**预编译二进制

**不要**安装或下载：`go/`、docs 源码树、另一平台二进制、全量 git 仓库。

> 为何有 launcher：`SKILL.md` 统一调用 `${CLAUDE_SKILL_DIR}/bin/ssh-skill`，由 launcher 按 `uname -s` 自动选择
> `ssh-skill-linux-amd64` / `ssh-skill-darwin-amd64` / `ssh-skill-windows-amd64.exe`，安装者无需关心当前 OS。

## 硬性约束

1. **禁止** `git clone`
2. **禁止** 下载 zipball / tarball / 全量仓库
3. **禁止** 默认从源码 `go build`（仅当最小下载失败且用户明确同意时才可）
4. **禁止** 修改用户已有 vault：`~/.ssh-skill/`（Windows 同理）
5. 只写全局 skill 目录，不把源码树落盘到用户机器

## 安装落点

| 平台 | 全局技能路径 |
|------|----------------|
| Linux / macOS | `~/.claude/skills/ssh-skill/` |
| Windows | `%USERPROFILE%\.claude\skills\ssh-skill\` |

最终结构：

```text
~/.claude/skills/ssh-skill/
├── SKILL.md
└── bin/
    ├── ssh-skill                      # launcher（平台无关，统一入口）
    ├── ssh-skill-linux-amd64          # Linux x86-64
    ├── ssh-skill-darwin-amd64         # macOS x86-64
    └── ssh-skill-windows-amd64.exe    # Windows x86-64
```

一个平台实际只下载 launcher + **该平台对应**的那一个二进制。

## 推荐执行方式（优先）

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/wjy2001/ssh-skill/master/scripts/install-skill.sh | bash
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/wjy2001/ssh-skill/master/scripts/install-skill.ps1 | iex
```

脚本会：

- 只下载 `SKILL.md` + `bin/ssh-skill`（launcher）+ 当前平台二进制
- 覆盖旧版 skill
- 删除另一平台残留二进制（含旧版 `ssh-skill.exe`）
- 运行 `--version` 验证
- **不**触碰 vault

## 回退方式（脚本不可用时）

仍禁止 clone。用 `raw.githubusercontent.com` 逐文件下载：

- Skill 定义：  
  `https://raw.githubusercontent.com/wjy2001/ssh-skill/master/.claude/skills/ssh-skill/SKILL.md`
- launcher（所有平台都要）：  
  `https://raw.githubusercontent.com/wjy2001/ssh-skill/master/.claude/skills/ssh-skill/bin/ssh-skill`
- Linux x86-64 二进制：  
  `https://raw.githubusercontent.com/wjy2001/ssh-skill/master/.claude/skills/ssh-skill/bin/ssh-skill-linux-amd64`
- macOS x86-64 二进制：  
  `https://raw.githubusercontent.com/wjy2001/ssh-skill/master/.claude/skills/ssh-skill/bin/ssh-skill-darwin-amd64`
- Windows x86-64 二进制：  
  `https://raw.githubusercontent.com/wjy2001/ssh-skill/master/.claude/skills/ssh-skill/bin/ssh-skill-windows-amd64.exe`

下载后写入上述全局 skill 路径，并 `chmod +x`（Unix）后执行 `--version`。
预编译二进制只提供 **amd64**；ARM64 用户需从源码构建（`./scripts/build.sh <os> arm64` 或 `.\scripts\build.ps1`）。

## 验证

安装后必须验证：

```bash
# Linux / macOS / Windows-Git-Bash（launcher 自动选平台二进制）
~/.claude/skills/ssh-skill/bin/ssh-skill --version

# Windows (PowerShell) 不能执行无扩展名的 launcher，直接验证 .exe
%USERPROFILE%\.claude\skills\ssh-skill\bin\ssh-skill-windows-amd64.exe --version
```

## 完成后汇报（简洁中文）

1. 安装路径
2. 版本号输出
3. 下载了哪些文件（应只有 skill + launcher + 当前平台二进制）
4. 下一步：`vault init` / 添加服务器（**先询问用户**，不要擅自改 vault）

可提示用户接下来可以说：

- 「列出已配置服务器」
- 「在 my-server 上执行 uptime」

## 可选：首次配置（仅当用户要求）

若用户明确要求初始化：

1. `vault init`（幂等；已有配置时不要清空）
2. 询问服务器信息（id / host / user / auth-type / password 或 key-path）后再 `add`
3. `test` 验证连通性
4. 不要在聊天中回显明文密码

## 维护说明（给仓库维护者）

- 本文件可随时更新；用户侧 README 短提示词无需改动
- 安装脚本变更时，同步更新本文件中的命令与约束
- 二进制路径变更时，同步更新 raw URL
- `bin/ssh-skill` 固定为 launcher；真实二进制一律以 `ssh-skill-<os>-<arch>[.exe]` 命名，切勿覆盖 launcher