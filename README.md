# Crafter

A lightweight, human-in-the-loop AI development methodology for Claude Code.

Crafter keeps you in control at every step: plans need your approval, diffs get reviewed before anything is committed, and context files grow alongside your project. Inspired by GSD, but designed to be minimal, conversational, and developer-controlled.

## Quick Start

**Global install (default) — available in all projects:**

```bash
curl -fsSL https://raw.githubusercontent.com/richardriman/crafter/main/install.sh | bash
```

**Additional options:**

```bash
# Local install — inside your current project (.claude/), committable and team-shareable
curl -fsSL https://raw.githubusercontent.com/richardriman/crafter/main/install.sh | bash -s -- --local

# Install a specific version
curl -fsSL https://raw.githubusercontent.com/richardriman/crafter/main/install.sh | bash -s -- --version 0.1.0
```

Then open Claude Code in your project and run:

```
/crafter-map-project
```

Crafter is now **skills-first** in source (`skills/crafter-*/SKILL.md`). Installer deploys skills to `.claude/skills/crafter-*/SKILL.md`.

**Alternative (for contributors working on Crafter itself):**

```bash
git clone https://github.com/richardriman/crafter.git
cd crafter
./install.sh
```

## CLI Binary

Global install now links the CLI command to:

```bash
~/.local/bin/crafter
```

If `crafter` is not found in your shell, add `~/.local/bin` to `PATH`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

When running `./install.sh` from a local clone, installer now falls back to
building `cli` from source when `cli/bin/crafter` is missing.

Update commands:

```bash
# Update global install to latest
crafter update

# Update global install to a specific release
crafter update --version 0.8.1

# Update local project install
crafter update --local
```

### Statusline

`crafter statusline` renders Crafter state as a full status panel for Claude Code's native status bar. The panel is a single line of up to five sections joined by ` │ ` (a space-padded `│`), in the order `plan │ model │ vcs │ ctx │ cost`. Each section degrades independently and is simply omitted when it has no data, so the panel always renders whatever it can:

```
Phase 2/3 · 7/12 [█████░░░░░] 58% │ Opus 4.8 1M (high) │ crafter ⎇ feat/statusline +18/-4 │ [████░░░░░░] 43% │ $0.42
```

**plan** — the plan position. When an active task is on the current branch it shows the full plan-progress segment (`Phase 2/3 · 7/12 [█████░░░░░] 58%`); before an approved plan exists it shows the edge states `planning` (plan not written yet) or `plan: awaiting approval` (written but not approved). When no task is active on the current branch the section falls back through `✓ done` (a completed task on this branch) and `N active elsewhere` (active tasks on other branches; the count is not pluralized, so a single task renders `1 active elsewhere`). The section is dropped when none of these apply.

**model** — `display_name` + the abbreviated context-window capacity (`1M`, `200k`) + the effort level in parentheses, e.g. `Opus 4.8 1M (high)`. The capacity is dropped when unknown and the `(level)` suffix is dropped when there is no effort level.

**vcs** — a group of `<project> ⎇ <branch> +N/-N`: the project name (basename of `workspace.project_dir`, dim grey), the branch icon and branch name, and the green/red added/removed line counts. Each part appears only when its data is present. The branch icon defaults to `⎇` (U+2387) and is configurable via the `CRAFTER_STATUSLINE_BRANCH_ICON` environment variable.

**ctx** — a progress bar plus percentage from `context_window.used_percentage` (e.g. `[████░░░░░░] 43%`), using the same bar style as the plan section. Omitted when the percentage is null.

**cost** — the session cost from `cost.total_cost_usd`, formatted `$X.XX`. Omitted when zero or absent.

The command never breaks the status bar: it always exits 0 and produces no output on any error (no `.crafter/` directory, detached HEAD, unreadable files). It does not collapse to empty just because there is no task — any section with data still renders.

**Known limitation:** the resolver matches only the standard `- **Work branch:**` metadata field. A task file using a non-standard field (e.g. `- **Branch:**`) is not counted toward `N active elsewhere`. This is intentional — the resolver is strict to the single documented field.

To wire it up, pass `--with-statusline` to the installer:

```bash
curl -fsSL https://raw.githubusercontent.com/richardriman/crafter/main/install.sh | bash -s -- --with-statusline
```

The installer applies a three-rung decision tree to your `settings.json`:

- **absent** — no `statusLine` key exists: the installer sets it automatically to `{ "type": "command", "command": "<crafter-bin> statusline" }`.
- **ours** — the key already holds a Crafter statusline command: the installer updates it only if the binary path changed (e.g. after a move); an identical entry is a no-op.
- **foreign** — any other value is present (another tool's statusline, including a previous GSD one): on a real terminal the installer **prompts** whether to overwrite. On **yes**, the original `settings.json` is backed up to `settings.json.bak` and the old command is printed to the terminal so it is recoverable, then the Crafter statusline is written. On **no**, or when running non-interactively (`curl | bash`, CI, no TTY), the installer leaves the foreign value untouched and prints a ready-to-paste composite wrapper so you can merge both statuslines manually.

The installer no longer needs `node` to edit `settings.json` — all JSON mutation is performed by the Go `crafter` binary.

## Skills

| Command | Description |
|---|---|
| `/crafter-do <task>` | Plan, execute, review, and commit — adaptive to small/medium/large scope |
| `/crafter-debug <problem>` | Systematic, hypothesis-driven debugging workflow |
| `/crafter-status` | Display current project state from `.crafter/STATE.md` (with `.planning` fallback) |
| `/crafter-map-project` | Initialize or update `.crafter/` context files from codebase analysis |

`/crafter-do` enforces Karpathy-inspired guardrails across planning, implementation, verification, and review: **Think Before Coding**, **Simplicity First**, **Surgical Changes**, and **Goal-Driven Execution**. Plans are vertical execution contracts with step-level drift checks and phase-level review.

## Project Context Files

Crafter maintains three living documents in `.crafter/`:

- **PROJECT.md** — stack, dependencies, environment variables, how to run, conventions
- **ARCHITECTURE.md** — directory structure, key patterns, navigation guide
- **STATE.md** — current focus, recent changes, done items, planned work, known issues

These files are loaded on-demand by Crafter skills (for example `/crafter-do`, `/crafter-debug`, `/crafter-status`, `/crafter-map-project`) when needed.

If an existing project still uses legacy `.planning/`, Crafter can run with fallback and proactively offers migration to `.crafter/` using `git mv`.

## Orchestrator / Agent Architecture

Crafter commands run as **orchestrators**: the main context window manages the workflow and communicates with you, while specialized agents do the actual work in fresh context windows.

This matters because running planning, implementation, verification, and review all in one context leads to context rot, compaction, and hallucinations as the conversation grows. Each agent starts clean with only the context it needs.

| Agent | Role | Used in |
|---|---|---|
| **Planner** | Tech lead — writes the execution contract | `/crafter-do` PLAN step |
| **Implementer** | Senior developer — implements the current approved step | `/crafter-do` EXECUTE step, `/crafter-debug` fix step |
| **Verifier** | QA engineer — checks step drift, criteria, and regressions | `/crafter-do` VERIFY, `/crafter-debug` verification |
| **Reviewer** | Code reviewer — looks for bugs, security issues, unapproved deviations | `/crafter-do` REVIEW step |
| **Analyzer** | Architect-analyst — reads and maps the codebase | `/crafter-map-project`, Large scope research, `/crafter-debug` hypothesis |

Agents for each role are defined as native Claude Code agents in `~/.claude/agents/`. The orchestrator spawns agents by name and provides each one with only the context it needs.

### Vertical Verification

`/crafter-do` executes one step at a time. After each step, the Verifier performs a lightweight drift check against the approved contract. Once all steps in a phase pass, phase verification runs, then full review checks the coherent phase diff. High-risk steps can still trigger immediate review when needed.

## Philosophy

Crafter is built on a simple principle: **you are the craftsman, AI is your tool**. The developer stays in control at every decision point — no auto-commits, no silent refactors, no guessing.

Read more in [docs/philosophy.md](docs/philosophy.md).
