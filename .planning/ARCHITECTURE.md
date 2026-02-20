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
├── rules/
│   └── rules.md                 # Core rules governing all workflows
├── templates/                   # Templates for .planning/ file initialization
│   ├── ARCHITECTURE.md          # Template for target project's ARCHITECTURE.md
│   ├── PROJECT.md               # Template for target project's PROJECT.md
│   ├── STATE.md                 # Template for target project's STATE.md
│   └── claude-md.snippet        # Snippet injected into target project's CLAUDE.md
├── install.sh                   # Installer (--global or --local)
└── README.md                    # Project overview
```

## Navigation — Where to Find What

| What | Where |
|---|---|
| Slash command definitions | `commands/*.md` |
| Core behavioral rules | `rules/rules.md` |
| Subagent system prompts | `meta-prompts/*.md` |
| Planning file templates | `templates/*.md` |
| CLAUDE.md injection snippet | `templates/claude-md.snippet` |
| Installer | `install.sh` |
| Design philosophy | `docs/philosophy.md` |
| BMAD integration guide | `docs/bmad-integration.md` |

## Key Patterns & Decisions

### Orchestrator / Subagent Model

Commands act as orchestrators: they manage workflow and user communication but never analyze code, implement changes, or review diffs themselves. Work is delegated to five specialized subagent roles (Planner, Implementer, Verifier, Reviewer, Analyzer), each spawned in a fresh context window with only the information it needs.

Each subagent receives its role definition from `meta-prompts/` and a dynamically assembled `$CONTEXT` block with relevant project files.

### Human-in-the-Loop Gates

Every significant action requires explicit user approval: plan approval before execution, diff review before commit, commit only on user command.

### Adaptive Scope Detection

`/crafter:do` auto-classifies tasks as Small (1-3 files, direct flow), Medium (multiple files, step-by-step), or Large (research first, then step-by-step).

### Template-Driven .planning/ Initialization

`/crafter:map-project` uses the Analyzer subagent to scan the target codebase, then proposes `.planning/` file contents based on templates. The `claude-md.snippet` uses HTML comment markers for idempotent CLAUDE.md updates.

### Dual Installation Model

`install.sh` supports `--global` (to `~/.claude/`) and `--local` (to `.claude/`). Both use a shared `install_to()` function.

## Conventions

- Command files are Markdown with YAML frontmatter
- Meta-prompts use a `$CONTEXT` placeholder filled by the orchestrator at delegation time
- `install.sh` uses `set -euo pipefail` for strict error handling
