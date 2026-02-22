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
├── meta-prompts/                # Subagent role definitions (system prompts)
│   ├── analyze.md               # Analyzer role
│   ├── implement.md             # Implementer role
│   ├── planner.md               # Planner role
│   ├── review.md                # Reviewer role
│   └── verify.md                # Verifier role
├── rules/                       # Per-concern rule fragments (loaded selectively by commands)
│   ├── core.md                  # Universal rules (language, principles, context maintenance)
│   ├── do-workflow.md           # Standard Change workflow rules
│   ├── debug-workflow.md        # Debug workflow rules
│   ├── delegation.md            # Subagent spawning instruction
│   ├── post-change.md           # Shared post-change steps (docs check, commit, STATE update)
│   ├── task-lifecycle.md        # Task file lifecycle rules (create, update, close)
│   └── update-check.md          # Automatic update checking (24h-cached GitHub Releases API)
├── templates/                   # Templates for .planning/ file initialization
│   ├── ARCHITECTURE.md          # Template for target project's ARCHITECTURE.md
│   ├── PROJECT.md               # Template for target project's PROJECT.md
│   ├── STATE.md                 # Template for target project's STATE.md
│   ├── TASK.md                  # Template for task files (.planning/tasks/)
│   └── claude-md.snippet        # Snippet injected into target project's CLAUDE.md
├── install.sh                   # Installer (local or remote via curl | bash)
├── VERSION                      # Current version identifier
└── README.md                    # Project overview
```

## Navigation — Where to Find What

| What | Where |
|---|---|
| Slash command definitions | `commands/*.md` |
| Universal rules (language, principles) | `rules/core.md` |
| Standard Change workflow rules | `rules/do-workflow.md` |
| Debug workflow rules | `rules/debug-workflow.md` |
| Subagent delegation rules | `rules/delegation.md` |
| Shared post-change steps | `rules/post-change.md` |
| Task file lifecycle rules | `rules/task-lifecycle.md` |
| Update check rules | `rules/update-check.md` |
| Subagent system prompts | `meta-prompts/*.md` |
| Planning file templates | `templates/*.md` |
| Task file template | `templates/TASK.md` |
| CLAUDE.md injection snippet | `templates/claude-md.snippet` |
| Version identifier | `VERSION` |
| Installer | `install.sh` |
| Design philosophy | `docs/philosophy.md` |
| BMAD integration guide | `docs/bmad-integration.md` |

## Key Patterns & Decisions

### Orchestrator / Subagent Model

Commands act as orchestrators: they manage workflow and user communication but never analyze code, implement changes, or review diffs themselves. Work is delegated to five specialized subagent roles (Planner, Implementer, Verifier, Reviewer, Analyzer), each spawned in a fresh context window with only the information it needs.

Each subagent receives its role definition from `meta-prompts/` and a dynamically assembled `$CONTEXT` block with relevant project files.

### Model Selection

Each subagent role has an assigned model tier (opus / sonnet / haiku) based on task complexity. Configuration lives in `rules/delegation.md` alongside other delegation rules. The Analyzer role is adaptive — it uses sonnet by default but upgrades to opus for Large scope tasks.

### Subagent Context Budget

| Subagent | Receives |
|---|---|
| Planner | User request + relevant `.planning/` excerpts + relevant source files |
| Implementer | Approved plan + relevant `.planning/` excerpts + relevant source files |
| Verifier | Verification criteria + changed files + relevant test files |
| Reviewer | Approved plan + changed files + `.planning/ARCHITECTURE.md` (if available) |
| Analyzer | Codebase structure files + package manifests + existing docs + `.planning/` files |

### Role Reference

| Role | Meta-prompt | When used |
|---|---|---|
| **Planner** | `meta-prompts/planner.md` | PLAN step in `/crafter:do` |
| **Implementer** | `meta-prompts/implement.md` | EXECUTE step in `/crafter:do`, review-fix loop in `/crafter:do`, and fix step in `/crafter:debug` |
| **Verifier** | `meta-prompts/verify.md` | VERIFY step in `/crafter:do` and `/crafter:debug` |
| **Reviewer** | `meta-prompts/review.md` | REVIEW step in `/crafter:do` |
| **Analyzer** | `meta-prompts/analyze.md` | `/crafter:map-project`, research phase in Large scope tasks, hypothesis analysis in `/crafter:debug` |

### Human-in-the-Loop Gates

Every significant action requires explicit user approval: plan approval before execution, review-fix loop consent for Critical/Major findings, diff review before commit, commit only on user command.

### Adaptive Scope Detection

`/crafter:do` auto-classifies tasks as Small (1-3 files, direct flow), Medium (multiple files, step-by-step), or Large (research first, then step-by-step).

### Task Lifecycle

Task files in `.planning/tasks/` are created and managed by the orchestrator during workflows. Each task file serves dual purposes: active resume state while work is in progress (allowing sessions to resume after interruption) and a permanent decision record once completed (preserving the plan, outcome, and rationale for future reference).

### Template-Driven .planning/ Initialization

`/crafter:map-project` uses the Analyzer subagent to scan the target codebase, then proposes `.planning/` file contents based on templates. The `claude-md.snippet` uses HTML comment markers for idempotent CLAUDE.md updates.

### Dual Installation Model

`install.sh` supports `--global` (to `~/.claude/`) and `--local` (to `.claude/`). Both use a shared `install_to()` function. The script also supports remote execution via `curl | bash`: when run without a local repo present it auto-detects this mode, downloads the specified (or latest) release tarball from GitHub, and installs from it. An optional `--version` flag selects a specific release.

## Conventions

- Command files are Markdown with YAML frontmatter
- Meta-prompts use a `$CONTEXT` placeholder filled by the orchestrator at delegation time
- `install.sh` uses `set -euo pipefail` for strict error handling
- Rules are split into per-concern fragments; each command loads only the fragments it needs
