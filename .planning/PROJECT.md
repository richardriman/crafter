# Project

> Crafter is a lightweight, human-in-the-loop AI development methodology for Claude Code — a set of conventions, context files, and slash commands that keep the developer in control at every step.

## Stack

- **Language:** Markdown (all commands, rules, meta-prompts, templates, documentation)
- **Scripting:** Bash (install.sh only)
- **Runtime platform:** Claude Code CLI (custom slash commands)
- **Framework:** None — pure conventions-and-prompts, no runtime dependencies

## Key Dependencies

None. This project has no package manifests and no external dependencies. It relies solely on:
- A Bash shell (for `install.sh`)
- Claude Code CLI (for slash command execution)

## Environment Variables

None.

## How to Run

```bash
# Install globally (available in all projects)
./install.sh --global

# Install locally in a specific project
cd /your/project
/path/to/crafter/install.sh --local

# Then in Claude Code, initialize project context:
/crafter:map-project
```

There is no build step, no test suite, and no development server.

## Conventions

- **Naming:** kebab-case for filenames, UPPER-CASE for planning template files
- **Structure:** top-level directories by function (commands, meta-prompts, rules, templates, docs)
- **Commits:** conventional commits format (feat/fix/refactor/docs/chore/test)
- **Language:** all persistent files in English; conversational output matches user's language

## Key Decisions

| Date | Decision | Reason |
|---|---|---|
| 2026-02-19 | Orchestrator/subagent architecture | Prevents context rot in long conversations — each subagent starts with fresh context |
| 2026-02-19 | Two installation modes (global/local) | Global is convenient for solo developers; local is committable and team-shareable |
| 2026-02-19 | Adaptive scope detection in /crafter:do | One command handles everything from one-line fixes to cross-cutting refactors |
| 2026-02-19 | BMAD integration is optional, not required | Keeps Crafter lightweight while allowing multi-perspective analysis when needed |
