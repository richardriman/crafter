# Crafter

A lightweight, human-in-the-loop AI development methodology for Claude Code.

Crafter keeps you in control at every step: plans need your approval, diffs get reviewed before anything is committed, and context files grow alongside your project. Inspired by GSD and BMAD, but designed to be minimal, conversational, and developer-controlled.

## Quick Start

```bash
git clone https://github.com/richardriman/crafter.git
cd crafter
```

**Choose one installation mode:**

```bash
# Global — available in all projects
./install.sh --global

# Local — installed inside your current project (.claude/), committable and team-shareable
cd /your/project
/path/to/crafter/install.sh --local
```

Then open Claude Code in your project and run:

```
/crafter:map-project
```

## Commands

| Command | Description |
|---|---|
| `/crafter:do <task>` | Plan, execute, review, and commit — adaptive to small/medium/large scope |
| `/crafter:debug <problem>` | Systematic, hypothesis-driven debugging workflow |
| `/crafter:status` | Display current project state from `.planning/STATE.md` |
| `/crafter:map-project` | Initialize or update `.planning/` context files from codebase analysis |

## Project Context Files

Crafter maintains three living documents in `.planning/`:

- **PROJECT.md** — stack, dependencies, environment variables, how to run, conventions
- **ARCHITECTURE.md** — directory structure, key patterns, navigation guide
- **STATE.md** — current focus, recent changes, done items, planned work, known issues

These files are read at the start of every Claude Code session (via a small snippet in `CLAUDE.md`) so Claude always has fresh context about your project.

## Orchestrator / Subagent Architecture

Crafter commands run as **orchestrators**: the main context window manages the workflow and communicates with you, while specialized subagents do the actual work in fresh context windows.

This matters because running planning, implementation, verification, and review all in one context leads to context rot, compaction, and hallucinations as the conversation grows. Each subagent starts clean with only the context it needs.

| Subagent | Role | Used in |
|---|---|---|
| **Planner** | Tech lead — proposes the plan | `/crafter:do` PLAN step |
| **Implementer** | Senior developer — implements the approved plan | `/crafter:do` EXECUTE step, `/crafter:debug` fix step |
| **Verifier** | QA engineer — checks criteria, finds regressions | `/crafter:do` VERIFY, `/crafter:debug` verification |
| **Reviewer** | Code reviewer — looks for bugs, security issues, deviations | `/crafter:do` REVIEW step |
| **Analyzer** | Architect-analyst — reads and maps the codebase | `/crafter:map-project`, Large scope research, `/crafter:debug` hypothesis |

Meta-prompts for each role live in `~/.claude/crafter/meta-prompts/`. The orchestrator fills in the `$CONTEXT` placeholder dynamically with only the files each role needs.

### Automatic Parallelization

Claude Code automatically runs independent tool calls in parallel. In practice, this means the orchestrator may spawn **Verify and Review simultaneously** after implementation, since their inputs don't depend on each other. This is expected behavior — not a bug — and it speeds up the workflow without any loss of correctness. If the review-fix loop triggers, both steps are re-run on the updated files anyway.

## Philosophy

Crafter is built on a simple principle: **you are the craftsman, AI is your tool**. The developer stays in control at every decision point — no auto-commits, no silent refactors, no guessing.

Read more in [docs/philosophy.md](docs/philosophy.md).

## BMAD Integration

Crafter works great alongside BMAD party mode for big architectural decisions and brainstorming sessions. It's entirely optional — use it when you need multiple perspectives, skip it when you don't.

Read more in [docs/bmad-integration.md](docs/bmad-integration.md).

