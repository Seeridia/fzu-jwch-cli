---
name: fzu-jwch-cli
description: "Use when a user wants an agent to operate the installed FZU JWCH CLI (`fzu-jwch`) to log in or query Fuzhou University academic affairs data such as student info, terms, courses, marks, exams, school calendar, or calendar events. Also use for safe handling of CLI auth/session issues and JSON output."
metadata:
  author: Seeridia
  version: 1.0.0
---

# FZU JWCH CLI

## Purpose

Use this skill to operate the installed `fzu-jwch` command for the user. This is a runtime usage skill for agents, not a source-code development guide.

The CLI queries Fuzhou University JWCH data through `github.com/west2-online/jwch`.

## Availability

Before running queries, check that the command exists:

```bash
command -v fzu-jwch
```

If it is missing, ask the user to install it with:

```bash
curl -fsSL https://raw.githubusercontent.com/Seeridia/fzu-jwch-cli/main/scripts/install.sh | sh
```

After install, the user may need to restart their shell or source the shell profile updated by the installer.

## Safety

- Treat student ID, password, cookies, config files, grades, exams, and personal academic data as private.
- Do not ask the user to paste a password into chat.
- Prefer interactive login with `fzu-jwch login`; it prompts in the terminal and hides the password in a real TTY.
- If non-interactive login is needed, prefer `--password-stdin` or environment variables over `--password`.
- Never print raw cookies, full config JSON, or passwords in final responses.
- When summarizing results, include only the fields the user asked for.

## Login State

Before asking the user to log in, check the saved session:

```bash
fzu-jwch status --json
```

If this command succeeds, do not run `fzu-jwch login`; continue with the requested query. `status` may refresh an expired session using saved credentials.

Only ask the user to log in when `status` fails because no config exists or saved credentials cannot refresh the session.

## Login

For a normal login, run:

```bash
fzu-jwch login
```

The CLI prompts for student ID and password and saves credentials/session data under the user's config directory with `0600` file permissions.

For automation when the user has already provided environment variables in their shell:

```bash
printf '%s' "$FZU_JWCH_PASSWORD" | fzu-jwch login --id "$FZU_JWCH_ID" --password-stdin
```

Do not run commands that echo or expose the password.

## Query Workflow

Prefer `--json` whenever the result will be parsed, filtered, transformed, or summarized by the agent. Use table output only when the user wants to see the raw CLI-style display.

Common commands:

```bash
fzu-jwch status --json
fzu-jwch me --json
fzu-jwch terms --json
fzu-jwch courses --term 202502 --json
fzu-jwch marks --json
fzu-jwch credits --json
fzu-jwch credits --raw --json
fzu-jwch gpa --json
fzu-jwch exams --type cet --json
fzu-jwch exams --type js --json
fzu-jwch exams --type room --term 202502 --json
fzu-jwch rooms --campus qishan --date 2026-05-01 --start 1 --end 2 --json
fzu-jwch calendar --json
fzu-jwch calendar events --term 202502 --json
fzu-jwch week --json
fzu-jwch lectures --json
fzu-jwch plan --json
fzu-jwch notices --page 1 --json
fzu-jwch notices detail --tree-id 1040 --news-id 13769 --json
```

Useful global flags:

```bash
--config <path>       Use a specific config file
--timeout <duration>  Set operation timeout, default 30s
--no-auto-login       Do not refresh expired sessions automatically
--json                Output JSON for query commands
```

## Choosing Terms

- Use `fzu-jwch terms --json` before course queries when the user does not know the exact academic term.
- For course queries, pass the selected 6-digit term with `courses --term <term>`, for example `202502`.
- For exam room queries, `--term` is required and uses the same 6-digit term format.
- For calendar events, use `calendar --json` first if the user does not know the term; pass a listed 6-digit `term` value such as `202502`, not a `2025-2026-1` style label.
- For empty rooms, pass `rooms --campus <campus> --date YYYY-MM-DD --start <1-12> --end <1-12>`. Campus accepts Chinese names or aliases: `qishan`, `jinjiang`, `tongpan`, `quangang`, `yishan`, `xiamen`.
- For notices, use `notices --page <n>` first, then `notices detail --tree-id <id> --news-id <id>` with IDs from the list.

## Troubleshooting

- `config not found`: ask the user to run `fzu-jwch login`, then retry the query.
- Missing ID/password during login: run `fzu-jwch login` interactively.
- Expired session: retry the original command without `--no-auto-login`; the CLI refreshes sessions automatically.
- Slow or transient JWCH failures: retry once with a longer timeout, for example `--timeout 60s`.
- Missing term for exam room query: ask for the term or run `fzu-jwch terms --json`.
- Missing room date/range/campus: ask for those fields; do not invent date or class periods.
- Missing notice IDs: run `fzu-jwch notices --json` and use `wb_tree_id`/`wb_news_id` from the selected notice.

## Response Style

When reporting query results, summarize clearly and avoid dumping large JSON unless the user explicitly asks for raw data. Mention if a command failed and include the actionable next step, but do not expose secrets or config contents.
