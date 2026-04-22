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

## Skills

| Command | Description |
|---|---|
| `/crafter-do <task>` | Plan, execute, review, and commit — adaptive to small/medium/large scope |
| `/crafter-debug <problem>` | Systematic, hypothesis-driven debugging workflow |
| `/crafter-status` | Display current project state from `.crafter/STATE.md` (with `.planning` fallback) |
| `/crafter-map-project` | Initialize or update `.crafter/` context files from codebase analysis |

`/crafter-do` enforces Karpathy-inspired guardrails across planning, implementation, and review: **Think Before Coding**, **Simplicity First**, **Surgical Changes**, and **Goal-Driven Execution**.

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
| **Planner** | Tech lead — proposes the plan | `/crafter-do` PLAN step |
| **Implementer** | Senior developer — implements the approved plan | `/crafter-do` EXECUTE step, `/crafter-debug` fix step |
| **Verifier** | QA engineer — checks criteria, finds regressions | `/crafter-do` VERIFY, `/crafter-debug` verification |
| **Reviewer** | Code reviewer — looks for bugs, security issues, deviations | `/crafter-do` REVIEW step |
| **Analyzer** | Architect-analyst — reads and maps the codebase | `/crafter-map-project`, Large scope research, `/crafter-debug` hypothesis |

Agents for each role are defined as native Claude Code agents in `~/.claude/agents/`. The orchestrator spawns agents by name and provides each one with only the context it needs.

### Automatic Parallelization

Claude Code automatically runs independent tool calls in parallel. In practice, this means the orchestrator may spawn **Verify and Review simultaneously** after implementation, since their inputs don't depend on each other. This is expected behavior — not a bug — and it speeds up the workflow without any loss of correctness. If the review-fix loop triggers, both steps are re-run on the updated files anyway.

## Philosophy

Crafter is built on a simple principle: **you are the craftsman, AI is your tool**. The developer stays in control at every decision point — no auto-commits, no silent refactors, no guessing.

Read more in [docs/philosophy.md](docs/philosophy.md).
