# Crafter

A lightweight, human-in-the-loop AI development methodology for Claude Code.

Crafter keeps you in control at every step: plans need your approval, diffs get reviewed before anything is committed, and context files grow alongside your project. Inspired by GSD and BMAD, but designed to be minimal, conversational, and developer-controlled.

## Quick Start

```bash
git clone https://github.com/richardriman/crafter.git
cd crafter
./install.sh --global   # Install commands globally (~/.claude/commands/crafter/)
```

Then, in your project directory:

```bash
/path/to/crafter/install.sh --local    # Set up .planning/ context files
/crafter:map-project                   # Analyze and map your codebase
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

## Philosophy

Crafter is built on a simple principle: **you are the craftsman, AI is your tool**. The developer stays in control at every decision point — no auto-commits, no silent refactors, no guessing.

Read more in [docs/philosophy.md](docs/philosophy.md).

## BMAD Integration

Crafter works great alongside BMAD party mode for big architectural decisions and brainstorming sessions. It's entirely optional — use it when you need multiple perspectives, skip it when you don't.

Read more in [docs/bmad-integration.md](docs/bmad-integration.md).

## License

MIT
