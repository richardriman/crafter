# Architecture

## Structure

```
crafter/
├── skills/                      # Canonical Crafter workflow definitions (skills-first source)
│   ├── crafter-debug/
│   │   └── SKILL.md             # crafter-debug — debugging orchestrator
│   ├── crafter-do/
│   │   └── SKILL.md             # crafter-do — adaptive change workflow orchestrator
│   ├── crafter-map-project/
│   │   └── SKILL.md             # crafter-map-project — project context initialization
│   └── crafter-status/
│       └── SKILL.md             # crafter-status — display current state
├── commands/                    # Compatibility slash-command wrappers
│   ├── debug.md                 # /crafter:debug -> routes to crafter-debug skill
│   ├── do.md                    # /crafter:do -> routes to crafter-do skill
│   ├── map-project.md           # /crafter:map-project -> routes to crafter-map-project skill
│   └── status.md                # /crafter:status -> routes to crafter-status skill
├── docs/                        # Supplementary documentation
│   └── philosophy.md            # Design philosophy and principles
├── hooks/                       # Claude Code lifecycle hooks
│   └── crafter-check-update.js  # SessionStart hook — automatic update check (24h-cached GitHub Releases API)
├── cli/                         # Go CLI binary source (crafter utility tool)
│   ├── main.go                  # Entry point
│   ├── cmd/                     # Cobra command definitions
│   ├── internal/skillbook/      # Skillbook logic (types, store, jaccard, format)
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
    └── commands/
        └── release.md           # /crafter:release — internal release preparation (not distributed)
```

## Key Patterns & Decisions

### Orchestrator / Agent Model

Commands act as orchestrators: they manage workflow and user communication but never analyze code, implement changes, or review diffs themselves. Work is delegated to five specialized agents (Planner, Implementer, Verifier, Reviewer, Analyzer), each spawned in a fresh context window with only the information it needs.

Agents are defined as native Claude Code agents in `agents/` and are invoked by name (e.g., `crafter-planner`).

### Skills-first with Command Wrappers

Crafter's canonical workflow logic lives in `skills/crafter-*/SKILL.md`. The `commands/` directory is kept as a compatibility layer for slash-command entry points and routes into skills.

### Model Selection

Each agent role has an assigned model tier (opus / sonnet) based on task complexity. Configuration lives in `rules/delegation.md` alongside other delegation rules. The Analyzer role is adaptive — it uses sonnet by default but upgrades to opus for Large scope tasks.

### Agent Roles and Context

Agent role definitions, model tiers, and context budgets are specified in `rules/delegation.md` and `agents/*.md`.

### Human-in-the-Loop Gates

Every significant action requires explicit user approval: plan approval before execution, review-fix loop consent for Critical/Major findings, diff review before commit, commit only on user command.

### Adaptive Scope Detection

`/crafter:do` auto-classifies tasks as Small (1-3 files, direct flow), Medium (multiple files, step-by-step), or Large (research first, then step-by-step).

### Task Lifecycle

Task files in `.crafter/tasks/` serve dual purposes: active resume state while work is in progress and a permanent decision record once completed.

### Template-Driven .crafter/ Initialization

`/crafter:map-project` uses the Analyzer to scan the target codebase and propose `.crafter/` file contents based on templates.

### Dual Installation Model

`install.sh` supports `--global` (to `~/.claude/`) and `--local` (to `.claude/`) via a shared `install_to()` function, and also supports remote execution via `curl | bash` with optional `--version` selection. Installer deploys both `skills/crafter-*/SKILL.md` and compatibility `commands/crafter/*.md`.

### Crafter CLI — Utility Binary

A Go CLI binary (`crafter`) provides deterministic utilities that LLMs handle poorly. The binary is a utility tool, NOT orchestration — orchestration stays in markdown prompts. The CLI is invoked via Bash by the orchestrator.

Current subcommands:
- `crafter skillbook get` — read skillbook, filter/sort skills, format as markdown, increment appliedCount
- `crafter skillbook add` — add observation with Jaccard dedup and confidence promotion
- `crafter skillbook init` — create empty skillbook

Distribution: cross-compiled for darwin-arm64, darwin-amd64, linux-amd64, linux-arm64. Binaries attached to GitHub releases. `install.sh` downloads the correct binary to `~/.claude/crafter/bin/crafter`.

### Skillbook — Project-Level Learning

The skillbook system lets agents learn from experience across sessions. After each task, the orchestrator reflects on what happened and captures observations via `crafter skillbook add`. Before spawning an agent, the orchestrator calls `crafter skillbook get` and appends the output to the agent's task prompt.

Key mechanics: Jaccard keyword-overlap deduplication (threshold 0.6), three confidence tiers (low/medium/high) with promotion on repeated observations, top-10 skill selection sorted by confidence then usage count, atomic file writes.

The skillbook file (`.planning/skillbook.json`) is project-level — agents learn project-specific patterns, not general knowledge.

## Conventions

- Skill files are Markdown with YAML frontmatter (`skills/*/SKILL.md`); command files are wrappers for compatibility
- Agent files define the role, constraints, and output format for each agent. The orchestrator spawns agents via the Task tool with a task description; agents explore the codebase themselves using their Read/Grep/Glob tools.
- `install.sh` uses `set -euo pipefail` for strict error handling
- Rules are split into per-concern fragments; each command loads only the fragments it needs
