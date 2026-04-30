# FZU JWCH CLI

`fzu-jwch` 是一个使用 Go 编写的命令行工具，用于通过 [`github.com/west2-online/jwch`](https://github.com/west2-online/jwch) 查询福州大学教务处数据。

100% Codex 完成

## 安装

```bash
go install github.com/seeridia/fzu-jwch-cli@latest
```

本地开发时可以直接运行：

```bash
go run . --help
```

## 登录

账号、密码和会话数据会保存在 `os.UserConfigDir()/fzu-jwch/config.json`，文件权限为 `0600`。

```bash
fzu-jwch login --id 102400000 --password 'your-password'
printf '%s' 'your-password' | fzu-jwch login --id 102400000 --password-stdin
FZU_JWCH_ID=102400000 FZU_JWCH_PASSWORD='your-password' fzu-jwch login
```

## 命令

```bash
fzu-jwch me
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
go run . --help
go run . login --help
```

集成测试默认会跳过；只有同时设置 `FZU_JWCH_ID` 和 `FZU_JWCH_PASSWORD` 时才会运行。
