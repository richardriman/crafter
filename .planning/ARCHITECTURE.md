# Architecture

## Structure

```
crafter/
├── commands/                    # Claude Code slash command definitions
│   ├── debug.md                 # /crafter:debug — debugging orchestrator
│   ├── do.md                    # /crafter:do — adaptive change workflow orchestrator
│   ├── map-project.md           # /crafter:map-project — project context initialization
│   └── status.md                # /crafter:status — display current state
├── docs/                        # Supplementary documentation
│   ├── bmad-integration.md      # BMAD party mode integration guide
│   └── philosophy.md            # Design philosophy and principles
├── hooks/                       # Claude Code lifecycle hooks
│   └── crafter-check-update.js  # SessionStart hook — automatic update check (24h-cached GitHub Releases API)
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
├── templates/                   # Templates for .planning/ file initialization
│   ├── ARCHITECTURE.md          # Template for target project's ARCHITECTURE.md
│   ├── PROJECT.md               # Template for target project's PROJECT.md
│   ├── STATE.md                 # Template for target project's STATE.md
│   ├── TASK.md                  # Template for task files (.planning/tasks/)
│   └── claude-md.snippet        # Snippet injected into target project's CLAUDE.md
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

### Model Selection

Each agent role has an assigned model tier (opus / sonnet / haiku) based on task complexity. Configuration lives in `rules/delegation.md` alongside other delegation rules. The Analyzer role is adaptive — it uses sonnet by default but upgrades to opus for Large scope tasks.

### Agent Roles and Context

Agent role definitions, model tiers, and context budgets are specified in `rules/delegation.md` and `agents/*.md`.

### Human-in-the-Loop Gates

Every significant action requires explicit user approval: plan approval before execution, review-fix loop consent for Critical/Major findings, diff review before commit, commit only on user command.

### Adaptive Scope Detection

`/crafter:do` auto-classifies tasks as Small (1-3 files, direct flow), Medium (multiple files, step-by-step), or Large (research first, then step-by-step).

### Task Lifecycle

Task files in `.planning/tasks/` serve dual purposes: active resume state while work is in progress and a permanent decision record once completed.

### Template-Driven .planning/ Initialization

`/crafter:map-project` uses the Analyzer to scan the target codebase and propose `.planning/` file contents based on templates; `claude-md.snippet` uses HTML comment markers for idempotent CLAUDE.md updates.

### Dual Installation Model

`install.sh` supports `--global` (to `~/.claude/`) and `--local` (to `.claude/`) via a shared `install_to()` function, and also supports remote execution via `curl | bash` with optional `--version` selection.

## Conventions

- Command files are Markdown with YAML frontmatter
- Agent files define the role, constraints, and output format for each agent. The orchestrator spawns agents via the Task tool with a task description; agents explore the codebase themselves using their Read/Grep/Glob tools.
- `install.sh` uses `set -euo pipefail` for strict error handling
- Rules are split into per-concern fragments; each command loads only the fragments it needs
