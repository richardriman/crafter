# Crafter Core Rules

## Language Rules

- **Internal instructions, templates, and commands:** always English
- **Conversation with the user:** match the user's language — auto-detect from their input and respond in kind
- **Persistent files** (`.planning/*`, saved plans): always English
- **Live conversational output** (non-archived responses): use the user's language

## Context File Maintenance

- **PROJECT.md:** Update when the stack, dependencies, or conventions change.
- **ARCHITECTURE.md:** Update when the structure, patterns, or key decisions change.
- **STATE.md:** Updated via post-change steps after each commit.
- Never update context files without showing the user what changed.

## General Principles

- When uncertain, ask — don't guess or assume.
- Plans are written for humans: conversational, clear, and reasoned.
- Show your reasoning — explain why, not just what.
- Respect the existing code style and conventions of the project.
- For architectural decisions with significant tradeoffs, consider a BMAD party mode session (see `docs/bmad-integration.md`).
