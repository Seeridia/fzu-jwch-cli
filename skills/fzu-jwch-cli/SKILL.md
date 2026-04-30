---
name: fzu-jwch-cli
description: "Use when working with the FZU JWCH CLI (`fzu-jwch`) for Fuzhou University Academic Affairs Office data: logging in, querying student info, terms, courses, marks, exams, school calendar, JSON output, troubleshooting auth/session issues, or modifying/testing this Go CLI project."
---

# FZU JWCH CLI

## Overview

Use this skill to operate or maintain the `fzu-jwch` Go CLI. The CLI wraps `github.com/west2-online/jwch` and queries Fuzhou University JWCH data from a local terminal.

Prefer running the local project with `go run .` when inside the repository. Use the installed `fzu-jwch` binary only when the user explicitly wants the installed version.

## Safety

- Treat student ID, password, cookies, and config files as secrets.
- Prefer `--password-stdin` for login. Avoid recommending `--password` unless the user explicitly accepts shell history exposure.
- Do not print passwords, raw cookies, or full config JSON in final responses.
- Use `--config <path>` with a temporary test config when testing login flows.
- Remember that this project currently stores the password in the config file with `0600` permissions.

## Login

Use one of these patterns:

```bash
printf '%s' "$FZU_JWCH_PASSWORD" | go run . login --id "$FZU_JWCH_ID" --password-stdin
FZU_JWCH_ID=102400000 FZU_JWCH_PASSWORD='password' go run . login
```

The default config path is `os.UserConfigDir()/fzu-jwch/config.json`. The CLI stores credentials and session data, then automatically refreshes expired sessions unless `--no-auto-login` is set.

If credentials are missing, ask the user to provide them through environment variables or stdin. Do not ask them to paste a password into chat unless there is no other viable path.

## Query Commands

Use these commands for common tasks:

```bash
go run . me
go run . terms
go run . courses --term 2025-2026-1
go run . marks
go run . exams --type cet
go run . exams --type js
go run . exams --type room --term 2025-2026-1
go run . calendar
go run . calendar events --term-id 2025-2026-1
```

Add `--json` when the result should be parsed, transformed, summarized, or consumed by another tool:

```bash
go run . marks --json
```

Use `--timeout <duration>` for slow network calls and `--no-auto-login` when diagnosing session validity without refreshing it.

## Troubleshooting

- `config not found`: run `login` first or pass `--config <path>` pointing at an existing config.
- `missing id` or `missing password`: pass `--id`, set `FZU_JWCH_ID`, use `--password-stdin`, or set `FZU_JWCH_PASSWORD`.
- Session check failure: retry without `--no-auto-login`; the manager will attempt `Login()` and save new session data.
- Exam room query failure due to missing term: pass `exams --type room --term <term>`.
- Network or JWCH server failures may be transient. Retry once before changing code.

## Development Workflow

When modifying this project:

- Command definitions live in `cmd/root.go`.
- Auth and session management live in `internal/auth`.
- The upstream JWCH wrapper interface lives in `internal/client/service.go`.
- Human-readable and JSON output helpers live in `internal/output/output.go`.
- Prefer extending `client.Service` and fake services in tests before wiring new commands to real JWCH calls.

Run verification after changes:

```bash
go test ./...
go vet ./...
go run . --help
go run . login --help
```

The integration test only runs when both `FZU_JWCH_ID` and `FZU_JWCH_PASSWORD` are set.
