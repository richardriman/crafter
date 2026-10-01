# Architecture

## Structure

```
crafter/
├── skills/                      # Canonical Crafter workflow definitions (skills-first source)
│   ├── crafter-buffer/
│   │   └── SKILL.md             # crafter-buffer — append UAT or Gap entries to the per-run NDJSON buffer
│   ├── crafter-debug/
│   │   └── SKILL.md             # crafter-debug — debugging orchestrator
│   ├── crafter-do/
│   │   └── SKILL.md             # crafter-do — adaptive change workflow orchestrator
│   ├── crafter-map-project/
│   │   └── SKILL.md             # crafter-map-project — project context initialization
│   └── crafter-status/
│       └── SKILL.md             # crafter-status — display current state
├── docs/                        # Supplementary documentation
│   └── philosophy.md            # Design philosophy and principles
├── cli/                         # Go CLI binary source (crafter utility tool)
│   ├── main.go                  # Entry point
│   ├── cmd/                     # Cobra command definitions
│   │   ├── buffer.go            # `crafter buffer` parent command
│   │   ├── buffer_gap.go        # `crafter buffer gap` — append Gap entry to gaps-buffer.jsonl
│   │   ├── buffer_uat.go        # `crafter buffer uat` — append UAT entry to uat-buffer.jsonl
│   │   ├── pr_body.go           # `crafter pr-body` — render PR body sections from buffers + task file
│   │   ├── statusline.go        # `crafter statusline` — render the full status panel (plan │ model │ vcs │ ctx │ cost)
│   │   ├── check_update.go      # `crafter check-update` — SessionStart hook: print update notice + spawn background refresh
│   │   ├── install.go           # `crafter install` — installer-machinery parent command
│   │   ├── install_hook.go      # `crafter install hook` — register SessionStart hook into settings.json
│   │   ├── install_statusline.go # `crafter install statusline` — reconcile statusLine into settings.json
│   │   └── update.go            # `crafter update`
│   ├── internal/buffer/         # Buffer logic (types, store with O_APPEND atomic write, format)
│   ├── internal/claudesettings/ # settings.json load/mutate/save + statusLine reconcile logic
│   ├── internal/prbody/         # PR body renderer (reads NDJSON buffers + task file, emits markdown sections)
│   ├── internal/statusline/     # Statusline logic (task resolve, plan parse, per-section panel render)
│   ├── Makefile                 # Cross-compilation targets
│   ├── go.mod                   # Go module definition
│   └── go.sum                   # Dependency checksums
├── agents/                      # Native Claude Code agent definitions
│   ├── crafter-analyzer.md      # Analyzer agent
│   ├── crafter-checker.md       # Checker agent — drift + code review in one fresh-context pass (full and delta modes)
│   ├── crafter-implementer.md   # Implementer agent
│   └── crafter-planner.md       # Planner agent
├── rules/                       # Per-concern rule fragments (loaded selectively by commands)
│   ├── core.md                  # Universal rules (language, principles, context maintenance)
│   ├── do-workflow.md           # Standard Change workflow rules
│   ├── debug-workflow.md        # Debug workflow rules
│   ├── delegation.md            # Agent spawning instruction
│   ├── post-change.md           # Shared post-change steps (docs check, commit, STATE update)
│   └── task-lifecycle.md        # Task file lifecycle rules (create, update, close)
├── templates/                   # Templates for .crafter/ file initialization
│   ├── ARCHITECTURE.md          # Template for target project's ARCHITECTURE.md
│   ├── PROJECT.md               # Template for target project's PROJECT.md
│   ├── STATE.md                 # Template for target project's STATE.md
│   └── TASK.md                  # Template for task files (.crafter/tasks/)
├── tests/                       # Test suite
│   └── test_install.sh          # Pure-Bash tests for install.sh (zero external dependencies)
├── install.sh                   # Installer (local or remote via curl | bash)
├── VERSION                      # Current version identifier
├── README.md                    # Project overview
└── .claude/                     # Project-local Claude Code config (internal)
    └── skills/
        └── crafter-release/
            └── SKILL.md         # /crafter-release — internal release preparation (not distributed)
```

## Key Patterns & Decisions

### Orchestrator / Agent Model

Skills act as orchestrators: they manage workflow and user communication but never analyze code, implement changes, or review diffs themselves. Work is delegated to four specialized agents (Planner, Implementer, Checker, Analyzer), each spawned in a fresh context window with only the information it needs.

Agents are defined as native Claude Code agents in `agents/` and are invoked by name (e.g., `crafter-planner`). Startup glue (extension discovery, resume detection, scope classification) is run inline by the orchestrator rather than delegated, and Small-scope plans are written inline too.

### Skills-first

Crafter's canonical workflow logic lives in `skills/crafter-*/SKILL.md`.

### Model Selection

All four agents run on `opus` (Opus 5.5); effort is `high` for Planner and Checker and `medium` for Implementer and Analyzer. The table and rationale live in `rules/delegation.md` § Model Configuration, mirrored by each agent's frontmatter. The same file gates the caveman directive to English-only conversations (§ Skill Directives, item 1); for other languages the Checker's hard output-format limits (`agents/crafter-checker.md` § Output format) keep reports terse.

### Agent Roles and Context

Agent role definitions, model tiers, and context budgets are specified in `rules/delegation.md` and `agents/*.md`. Every agent definition carries the same `## Ending your turn` section, which names the unwanted early stops and allows ending only when the assignment is complete or blocked.

### Human-in-the-Loop Gates

Plan approval is the one unconditional gate: execution never starts without it. After the check pass, commits are triggered automatically — the default is auto-commit once no Critical or Major findings remain, and explicit user approval is required only for the manual-verification exception (the plan states that verification needs manual testing). Suggestion findings do not gate anything: each is recorded as a `Decision (Tech Debt — auto-recorded)` entry and the run continues. Minor findings stop the run for a user decision — fix all, pick numbers, or defer all; chosen ones enter the fix loop, deferred ones are recorded as `Decision (Tech Debt — user-deferred)` (under `--auto` they are auto-recorded without a stop). Critical or Major findings, and harmful drift, stop the run and trigger a mandatory fix loop with a 5-iteration cap before the commit can proceed. Critical and Major are gated on reachability: the Checker must name a reachable trigger and record one of three states — verified reachable (severity stands), verified unreachable with the checked callers named (at most Minor), or cannot determine (severity kept). The orchestrator never downgrades a finding; a Critical or Major that names neither a trigger nor a state is sent back to the Checker as a `delta pass` with `reachability check on finding #N` — a Checker-only pass, at most one per finding, which does not increment the fix-loop iteration count and leaves the 5-iteration cap untouched.

Two independent flags modify the flow. `--ext` (default off) enables extension-skill discovery and the pre-spawn extension checks; without it no discovery scan runs at all. `--auto` (default off) runs the full Plan → Execute → Check → PR cycle without interactive pauses, retaining four hard gates (initial clarification, plan approval, fix-loop cap reached, ad-hoc escape hatch); everything else is handled automatically. `--auto` enforces the green-commit invariant: if the fix loop cannot bring the work to green within budget, the run exits with state rather than committing. The `--fast` flag was removed and is now rejected with an error.

### Execution Contracts and the Check Pass

`/crafter-do` plans work as execution contracts. Small and Medium scope produce a flat step list under one contract for the whole task; Large groups steps into vertical phases with one contract per phase. A contract defines: outcome, scope boundary, non-goals, seams, verification evidence, and stop conditions — there is no per-step contract. Execution follows the same units: the whole task in one Implementer spawn for Small/Medium, one whole phase per spawn for Large, with the Implementer running the relevant tests and reporting them as evidence. Each unit ends with one `crafter-checker` pass covering drift and code review together; the fix loop re-checks in delta mode over the files the fix changed, widening back to a full pass when the fix reaches outside the delta. Step 6b (Summary and Commit) closes the unit before moving to the next phase or to Steps 7–9.

### Adaptive Scope Detection

`/crafter-do` first checks whether the request is complete enough to plan, then auto-classifies tasks as Small (1-3 files, isolated), Medium (multiple files/cross-cutting), or Large (incomplete, architectural, many files, or unfamiliar). Incomplete tasks go through targeted discussion and/or research before planning.

### Task Lifecycle

Task files in `.crafter/tasks/` serve dual purposes: active resume state while work is in progress and a permanent decision record once completed.

### Template-Driven .crafter/ Initialization

`/crafter-map-project` uses the Analyzer to scan the target codebase and propose `.crafter/` file contents based on templates.

### Dual Installation Model

`install.sh` supports `--global` (to `~/.claude/`) and `--local` (to `.claude/`) via a shared `install_to()` function, and also supports remote execution via `curl | bash` with optional `--version` selection. Installer deploys `skills/crafter-*/SKILL.md`.

The statusline is wired by default on every install (both `--global` and `--local`). There is no opt-in flag; `--with-statusline` has been removed and the installer hard-errors if it is passed. The statusline reconcile is implemented in `crafter install statusline` via a three-rung decision tree: **absent** → set automatically; **ours** (already a Crafter command) → idempotent update only if the binary path differs, otherwise no-op; **foreign** (any other statusLine value) → on a real terminal the installer prompts to overwrite; on yes the original file is backed up to `settings.json.bak` and the old command is printed for recovery, then overwritten; on no, or when non-interactive (`curl | bash`, CI, no TTY), the foreign value is left untouched and manual-merge guidance is printed. All `settings.json` mutation is performed by the Go `crafter` binary — the installer does not need `node` to edit settings.

### Crafter CLI — Utility Binary

A Go CLI binary (`crafter`) provides deterministic utilities that LLMs handle poorly. The binary is a utility tool, NOT orchestration — orchestration stays in markdown prompts. The CLI is invoked via Bash by the orchestrator.

Current subcommands:
- `crafter buffer uat` — append a UAT entry (NDJSON line) to `<run-dir>/uat-buffer.jsonl`, creating the file with a marker line if missing
- `crafter buffer gap` — append a Gap entry (NDJSON line) to `<run-dir>/gaps-buffer.jsonl`, creating the file with a marker line if missing
- `crafter update` — fetch and run the official installer to update global or local Crafter installations
- `crafter pr-body` — read per-run NDJSON buffers and task file, render `## Manual QA Plan`, `## Known Gaps`, and `## Decisions` sections for the PR body
- `crafter statusline` — render the full status panel for Claude Code's status bar: up to five sections joined by ` │ ` in the order `plan │ model │ vcs │ ctx │ cost`. **plan** is the plan position (active task on the current branch → full plan-progress segment e.g. `Phase 2/3 · 7/12 [█████░░░░░] 58%`, else the cascade `✓ done` / `N active elsewhere`, else dropped); **model** is `display_name` + capacity + `(effort)` e.g. `Opus 5.5 1M (high)`; **vcs** is the group `<project> ⎇ <branch> +N/-N` (branch icon configurable via `CRAFTER_STATUSLINE_BRANCH_ICON`, default `⎇`); **ctx** is a progress bar + `%` from `context_window.used_percentage`; **cost** is `$X.XX` from `cost.total_cost_usd`. Each section degrades independently and is omitted when it has no data; always exits 0 and never collapses to empty merely because no task is active
- `crafter check-update` — SessionStart hook command: reads the installed VERSION and a 4h-cached result file (`~/.claude/cache/crafter-update-check.json`), prints an update notice when one is available, then spawns a detached background refresh that queries GitHub `releases/latest` and rewrites the cache. Silent-fail and non-blocking; registered by `install.sh` as `"<crafter-bin>" check-update`
- `crafter install hook` — register the Crafter SessionStart hook into a `settings.json` (idempotent: no-op when the command is already present; preserves all pre-existing hook entries verbatim)
- `crafter install statusline` — reconcile the Crafter statusLine into a `settings.json` using the three-rung decision tree: **absent** → set; **ours, identical** → no-op; **ours, differs** → update; **foreign** → act on `--on-foreign keep|overwrite` (the TTY prompt and fallback live in `install.sh`; this subcommand never reads a terminal)

Run-directory lifecycle (`.crafter/run/<task-id>/`) — canonical wording in `rules/do-workflow.md → ### Run directory lifecycle`.

Distribution: cross-compiled for darwin-arm64, darwin-amd64, linux-amd64, linux-arm64. Binaries attached to GitHub releases. `install.sh` downloads the correct binary to `~/.claude/crafter/bin/crafter` and links global installs to `~/.local/bin/crafter` for shell usage.

### PR Composer — `--auto` End-of-Task PR Creation

Under `--auto`, Step 9b (defined in `skills/crafter-do/SKILL.md → ## Step 9b`) closes the run by opening a GitHub PR. The orchestrator composes a baseline body (Summary + Test plan) from the task file's `## Plan → Approach` and `## Outcome` sections, then invokes `crafter pr-body --run-dir .crafter/run/<task-id>/ --task-file …` to render three appended sections (`## Manual QA Plan`, `## Known Gaps`, `## Decisions`) from the per-run NDJSON buffers. The two parts are concatenated and passed to `gh pr create`. On success the run directory is deleted; on failure the run directory is preserved and the ad-hoc escape hatch is triggered. This mirrors the GH#16 buffer pattern — deterministic rendering is delegated to the Go binary, not inlined as LLM prose.

### Skill Adaptation — Caveman, Ponytail, and Report Language

When the external `caveman` or `ponytail` skills are active in a session, the orchestrator detects them via marker files (`$HOME/.claude/.caveman-active` / `.ponytail-active`) at startup, because subagents run in fresh contexts and never receive those skills' own SessionStart injection. Detection and the human-facing caveman-lite policy live in `rules/core.md`; a single pre-spawn propagation rule in `rules/delegation.md` appends the appropriate directive to every agent's task prompt. Caveman mode is audience-driven: lite for orchestrator prose to the user, full for agent reasoning and returned reports. Ponytail (YAGNI / shortest-working-diff discipline) is scoped to `crafter-implementer` and `crafter-planner` only — the two roles that author code or plans. The caveman directive lives entirely in that pre-spawn propagation rule — no agent file carries a `## Behavior under caveman` section; only `crafter-implementer` and `crafter-planner` carry `## Behavior under ponytail`. Independent of both markers and emitted on every spawn, the orchestrator also names the user's conversation language in the agent's prompt (`rules/delegation.md` → `## Report Language (always on)`): free-text report prose comes back in that language, while code, identifiers, file paths, required headings, table columns, status-line formats, persistent files, and buffer-bound deviation/classification text stay English.

### Agent Memory — Project-Level Learning

Project-level learning uses native Claude Code subagent memory instead of a custom store. Each `agents/crafter-*.md` file declares `memory: project` in its frontmatter, which gives the agent a project-scoped memory file at `.claude/agent-memory/<agent>/MEMORY.md`, auto-loaded when the agent is spawned. Agents curate their own file; the curation rules live in each agent prompt. The orchestrator has no role in reading or writing memory.

## Conventions

- Skill files are Markdown with YAML frontmatter (`skills/*/SKILL.md`)
- Extension skills must satisfy the Crafter skill contract — see [`docs/skill-contract.md`](../docs/skill-contract.md)
- Agent files define the role, constraints, and output format for each agent. The orchestrator spawns agents via the Task tool with a task description; agents explore the codebase themselves using their Read/Grep/Glob tools.
- `install.sh` uses `set -euo pipefail` for strict error handling
- Rules are split into per-concern fragments; each skill loads only the fragments it needs
