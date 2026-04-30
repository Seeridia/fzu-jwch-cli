# FZU JWCH CLI

`fzu-jwch` 是一个使用 Go 编写的命令行工具，用于通过 [`github.com/west2-online/jwch`](https://github.com/west2-online/jwch) 查询福州大学教务处数据。

100% Codex 完成

## 安装

推荐使用安装脚本：

```bash
curl -fsSL https://raw.githubusercontent.com/Seeridia/fzu-jwch-cli/main/scripts/install.sh | sh
```

安装脚本会下载适合当前系统的 release 二进制文件，默认安装到 `~/.local/bin/fzu-jwch`。如果 `~/.local/bin` 不在 `PATH` 中，脚本会把它写入当前 shell 的配置文件。

## Skill

如果希望让 Agent 更好地调用这个 CLI，可以安装仓库内置的 skill：

```bash
npx skills add https://github.com/Seeridia/fzu-jwch-cli
```

## 登录

```bash
fzu-jwch status
fzu-jwch login
```

`fzu-jwch status` 用于检查已保存的登录状态；如果会话过期，CLI 会使用已保存的凭据自动刷新。直接运行 `fzu-jwch login` 时，CLI 会提示输入学号和密码；密码在终端中不会回显。

## 命令

```bash
fzu-jwch me
fzu-jwch status
fzu-jwch terms
fzu-jwch courses --term 2025-2026-1
fzu-jwch marks
fzu-jwch exams --type cet
fzu-jwch exams --type js
fzu-jwch exams --type room --term 2025-2026-1
fzu-jwch calendar
fzu-jwch calendar events --term-id 2025-2026-1
```

所有查询命令都支持 `--json` 输出：

```bash
fzu-jwch marks --json
```

全局参数：

```bash
--config <path>       使用自定义配置文件
--no-auto-login       不自动刷新过期会话
--timeout <duration>  操作超时时间，默认：30s
--json                查询命令以 JSON 格式输出
```

## 测试

```bash
go test ./...
go vet ./...
go run ./cmd/fzu-jwch --help
go run ./cmd/fzu-jwch login --help
```

集成测试默认会跳过；只有同时设置 `FZU_JWCH_ID` 和 `FZU_JWCH_PASSWORD` 时才会运行。
