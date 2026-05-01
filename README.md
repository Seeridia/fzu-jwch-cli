# FZU JWCH CLI

`fzu-jwch` 是一个使用 Go 编写的命令行工具，用于通过 [`github.com/west2-online/jwch`](https://github.com/west2-online/jwch) 查询福州大学教务处数据。

100% Codex 完成

## 安装

推荐使用安装脚本：

```bash
curl -fsSL https://raw.githubusercontent.com/Seeridia/fzu-jwch-cli/main/scripts/install.sh | sh
```

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
fzu-jwch courses --term 202502
fzu-jwch marks
fzu-jwch credits
fzu-jwch credits --raw
fzu-jwch gpa
fzu-jwch exams --type cet
fzu-jwch exams --type js
fzu-jwch exams --type room --term 202502
fzu-jwch rooms --campus qishan --date 2026-05-01 --start 1 --end 2
fzu-jwch calendar
fzu-jwch calendar events --term 202502
fzu-jwch week
fzu-jwch lectures
fzu-jwch plan
fzu-jwch notices --page 1
fzu-jwch notices detail --tree-id 1040 --news-id 13769
```

空教室查询的 `--campus` 支持 `qishan`、`jinjiang`、`tongpan`、`quangang`、`yishan`、`xiamen`，也支持对应中文校区名；`--start` 和 `--end` 为 1 到 12 的节次。

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
