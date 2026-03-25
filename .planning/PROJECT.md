# Project

> Crafter is a lightweight, human-in-the-loop AI development methodology for Claude Code — a set of conventions, context files, and slash commands that keep the developer in control at every step.

## Stack

- **Language:** Markdown (all commands, rules, agents, templates, documentation)
- **Language:** Go (CLI utility binary — `cli/` directory)
- **Scripting:** Bash (install.sh only)
- **Runtime platform:** Claude Code CLI (custom slash commands)
- **Framework:** cobra (Go CLI only — core project remains pure conventions-and-prompts)

## Conventions

- **Naming:** kebab-case for filenames, UPPER-CASE for planning template files
- **Structure:** top-level directories by function (commands, agents, rules, templates, docs)
- **Commits:** conventional commits format (feat/fix/refactor/docs/chore/test)
- **Language:** all persistent files in English; conversational output matches user's language
- **Source vs install:** When working on Crafter itself, always modify source files in the repository (`agents/`, `commands/`, `rules/`, `templates/`), never the installed copies in `~/.claude/crafter/`.

## Key Decisions

| Date | Decision | Reason |
|---|---|---|
| 2026-02-19 | Orchestrator/agent architecture | Prevents context rot in long conversations — each agent starts with fresh context |
| 2026-02-19 | Two installation modes (global/local) | Global is convenient for solo developers; local is committable and team-shareable |
| 2026-02-19 | Adaptive scope detection in /crafter:do | One command handles everything from one-line fixes to cross-cutting refactors |
| 2026-02-19 | BMAD integration is optional, not required | Keeps Crafter lightweight while allowing multi-perspective analysis when needed |
| 2026-03-24 | Go CLI binary for deterministic utilities | LLMs do JSON CRUD, Jaccard similarity, and atomic writes poorly — a static binary with zero runtime deps handles these reliably |
