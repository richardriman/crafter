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
├── hooks/                       # Claude Code lifecycle hooks
│   └── crafter-check-update.js  # SessionStart hook — automatic update check (24h-cached GitHub Releases API)
├── cli/                         # Go CLI binary source (crafter utility tool)
│   ├── main.go                  # Entry point
│   ├── cmd/                     # Cobra command definitions
│   │   ├── buffer.go            # `crafter buffer` parent command
│   │   ├── buffer_gap.go        # `crafter buffer gap` — append Gap entry to gaps-buffer.jsonl
│   │   ├── buffer_uat.go        # `crafter buffer uat` — append UAT entry to uat-buffer.jsonl
│   │   ├── pr_body.go           # `crafter pr-body` — render PR body sections from buffers + task file
│   │   ├── skillbook.go         # `crafter skillbook` parent command
│   │   ├── skillbook_add.go     # `crafter skillbook add`
│   │   ├── skillbook_get.go     # `crafter skillbook get`
│   │   ├── skillbook_init.go    # `crafter skillbook init`
│   │   ├── statusline.go        # `crafter statusline` — render plan position as a status-bar segment
│   │   └── update.go            # `crafter update`
│   ├── internal/buffer/         # Buffer logic (types, store with O_APPEND atomic write, format)
│   ├── internal/prbody/         # PR body renderer (reads NDJSON buffers + task file, emits markdown sections)
│   ├── internal/skillbook/      # Skillbook logic (types, store, jaccard, format)
│   ├── internal/statusline/     # Statusline logic (task resolve, plan parse, segment render)
│   ├── Makefile                 # Cross-compilation targets
│   ├── go.mod                   # Go module definition
│   └── go.sum                   # Dependency checksums
├── agents/                      # Native Claude Code agent definitions
│   ├── crafter-analyzer.md      # Analyzer agent
│   ├── crafter-implementer.md   # Implementer agent
│   ├── crafter-planner.md       # Planner agent
│   ├── crafter-reviewer.md      # Reviewer agent
│   └── crafter-verifier.md      # Verifier agent
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

Skills act as orchestrators: they manage workflow and user communication but never analyze code, implement changes, or review diffs themselves. Work is delegated to five specialized agents (Planner, Implementer, Verifier, Reviewer, Analyzer), each spawned in a fresh context window with only the information it needs.

Agents are defined as native Claude Code agents in `agents/` and are invoked by name (e.g., `crafter-planner`).

### Skills-first

Crafter's canonical workflow logic lives in `skills/crafter-*/SKILL.md`.

### Model Selection

Each agent role has an assigned model tier (opus / sonnet) based on task complexity. Configuration lives in `rules/delegation.md` alongside other delegation rules. The Analyzer role is adaptive — it uses sonnet by default but upgrades to opus for Large scope tasks.

### Agent Roles and Context

Agent role definitions, model tiers, and context budgets are specified in `rules/delegation.md` and `agents/*.md`.

### Human-in-the-Loop Gates

Every significant action requires user approval: plan approval before execution, diff review before commit. Phase-close commits are no longer gated on an explicit per-commit command — instead they are triggered automatically once phase verification passes and a clean review summary is produced. Approval follows one of three paths: (1) auto-approve when the review summary is clean and no user intervention is needed, (2) silence-approve when the `--fast` flag is set (silence is treated as approval), or (3) explicit user approval for any other case. Critical or Major review findings trigger a mandatory fix-loop with a 5-iteration cap before the commit can proceed.

A fourth mode, `--auto` (unattended orchestration), runs the full Plan → Execute → Verify → Review → PR cycle without interactive pauses. It retains four hard gates (initial clarification, plan approval, fix-loop cap reached, ad-hoc escape hatch); everything else is handled automatically. `--auto` enforces the green-commit invariant: if the fix loop cannot bring a phase to green within budget, the run exits with state rather than committing. `--auto` and `--fast` are mutually exclusive.

### Vertical Planning and Drift Checks

`/crafter-do` plans work as vertical execution contracts. Each phase and step defines a Karpathy Contract: outcome, scope boundary, non-goals, simplicity constraint, drift criteria, verification evidence, and stop conditions. The Implementer works one step at a time, the Verifier runs step drift checks before the next step, and full Review runs after phase verification unless a high-risk step requires immediate review. The phase-close flow now includes Step 6b (Phase Summary and Auto-Commit) as a standard step before moving to the next phase or to Steps 7–9.

### Adaptive Scope Detection

`/crafter-do` first checks whether the request is complete enough to plan, then auto-classifies tasks as Small (1-3 files, isolated), Medium (multiple files/cross-cutting), or Large (incomplete, architectural, many files, or unfamiliar). Incomplete tasks go through targeted discussion and/or research before planning.

### Task Lifecycle

Task files in `.crafter/tasks/` serve dual purposes: active resume state while work is in progress and a permanent decision record once completed.

### Template-Driven .crafter/ Initialization

`/crafter-map-project` uses the Analyzer to scan the target codebase and propose `.crafter/` file contents based on templates.

### Dual Installation Model

`install.sh` supports `--global` (to `~/.claude/`) and `--local` (to `.claude/`) via a shared `install_to()` function, and also supports remote execution via `curl | bash` with optional `--version` selection. Installer deploys `skills/crafter-*/SKILL.md`.

### Crafter CLI — Utility Binary

A Go CLI binary (`crafter`) provides deterministic utilities that LLMs handle poorly. The binary is a utility tool, NOT orchestration — orchestration stays in markdown prompts. The CLI is invoked via Bash by the orchestrator.

Current subcommands:
- `crafter buffer uat` — append a UAT entry (NDJSON line) to `<run-dir>/uat-buffer.jsonl`, creating the file with a marker line if missing
- `crafter buffer gap` — append a Gap entry (NDJSON line) to `<run-dir>/gaps-buffer.jsonl`, creating the file with a marker line if missing
- `crafter skillbook get` — read skillbook, filter/sort skills, format as markdown, increment appliedCount
- `crafter skillbook add` — add observation with Jaccard dedup and confidence promotion
- `crafter skillbook init` — create empty skillbook
- `crafter update` — fetch and run the official installer to update global or local Crafter installations
- `crafter pr-body` — read per-run NDJSON buffers and task file, render `## Manual QA Plan`, `## Known Gaps`, and `## Decisions` sections for the PR body
- `crafter statusline` — read Claude Code's stdin JSON payload and render the active task's plan position as a single composable status-bar segment (e.g. `crafter · Phase 2/3 · 7/12 [█████░░░░░] 58%`); silent when not a Crafter project or no active task

Run-directory lifecycle (`.crafter/run/<task-id>/`) — canonical wording in `rules/do-workflow.md → ### Run directory lifecycle`.

Distribution: cross-compiled for darwin-arm64, darwin-amd64, linux-amd64, linux-arm64. Binaries attached to GitHub releases. `install.sh` downloads the correct binary to `~/.claude/crafter/bin/crafter` and links global installs to `~/.local/bin/crafter` for shell usage.

### PR Composer — `--auto` End-of-Task PR Creation

Under `--auto`, Step 9b (defined in `skills/crafter-do/SKILL.md → ## Step 9b`) closes the run by opening a GitHub PR. The orchestrator composes a baseline body (Summary + Test plan) from the task file's `## Plan → Approach` and `## Outcome` sections, then invokes `crafter pr-body --run-dir .crafter/run/<task-id>/ --task-file …` to render three appended sections (`## Manual QA Plan`, `## Known Gaps`, `## Decisions`) from the per-run NDJSON buffers. The two parts are concatenated and passed to `gh pr create`. On success the run directory is deleted; on failure the run directory is preserved and the ad-hoc escape hatch is triggered. This mirrors the GH#16 buffer pattern — deterministic rendering is delegated to the Go binary, not inlined as LLM prose.

### Skillbook — Project-Level Learning

The skillbook system lets agents learn from experience across sessions. After each task, the orchestrator reflects on what happened and captures observations via `crafter skillbook add`. Before spawning an agent, the orchestrator calls `crafter skillbook get` and appends the output to the agent's task prompt.

Key mechanics: Jaccard keyword-overlap deduplication (threshold 0.6), three confidence tiers (low/medium/high) with promotion on repeated observations, top-10 skill selection sorted by confidence then usage count, atomic file writes.

The skillbook file (`{PROJECT_PATH}/{CRAFTER_DIR}/skillbook.json`, with `.crafter` preferred and `.planning` as legacy fallback) is project-level — agents learn project-specific patterns, not general knowledge.

## Conventions

- Skill files are Markdown with YAML frontmatter (`skills/*/SKILL.md`)
- Extension skills must satisfy the Crafter skill contract — see [`docs/skill-contract.md`](../docs/skill-contract.md)
- Agent files define the role, constraints, and output format for each agent. The orchestrator spawns agents via the Task tool with a task description; agents explore the codebase themselves using their Read/Grep/Glob tools.
- `install.sh` uses `set -euo pipefail` for strict error handling
- Rules are split into per-concern fragments; each skill loads only the fragments it needs
