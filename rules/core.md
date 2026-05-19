# Crafter Core Rules

## Language Rules

- **Internal instructions, templates, and commands:** always English
- **Conversation with the user:** match the user's language — auto-detect from their input and respond in kind
- **Persistent files** (`.crafter/*`, saved plans; legacy fallback `.planning/*`): always English
- **Live conversational output** (non-archived responses): use the user's language

## Jargon Confinement

Agents and the orchestrator MUST describe the user's own project, codebase, and domain using domain-neutral language or the user's own terms. Do not import crafter's internal vocabulary into explanations, plans, summaries, or review prose about the user's unrelated domain.

Reserved crafter-internal terms — `gate`, `drift`, `seam` / `split point`, `surface`, `binding`, `escape hatch` — are confined to crafter's own workflow mechanics. All of these terms remain fully legitimate when describing crafter's own mechanics (e.g., `gate`/`gated` for phase gates, `drift` for step drift check / scope drift / beneficial local drift). The prohibition is solely against exporting them onto the user's unrelated domain or codebase.

- **Bad (jargon bleed):** "Doporučuji gatovat panel se záložkami na všech švech a přidat flag gate pro aktivaci" — crafter-internal terms (`gate`, `seam`) verb-ified and projected onto unrelated Gantt app UI elements.
- **Good (domain-neutral):** "Doporučuji skrýt záložkový panel za feature flag a aktivovat ho postupně" — the same intent expressed in the user's own product terms.

## Context File Maintenance

- **PROJECT.md:** Update when the stack, dependencies, or conventions change.
- **ARCHITECTURE.md:** Update when the structure, patterns, or key decisions change. Orchestrators must delegate ARCHITECTURE.md reads and writes to agents.
- **STATE.md:** Updated via post-change steps after each commit.
- Never update context files without showing the user what changed.

## General Principles

- When uncertain, ask — don't guess or assume.
- Plans are written for humans: conversational, clear, and reasoned.
- Show your reasoning — explain why, not just what.
- Respect the existing code style and conventions of the project.

## Karpathy-Inspired Guardrails

- **Think Before Coding:** Surface assumptions explicitly. If multiple interpretations exist, present them instead of picking silently.
- **Simplicity First:** Prefer the smallest change that solves today's requirement. Avoid speculative abstractions and unused flexibility.
- **Surgical Changes:** Every changed line must trace to the approved request. Avoid drive-by refactors and adjacent "improvements."
- **Goal-Driven Execution:** Convert work into verifiable criteria and iterate until each criterion is clearly satisfied.

For `/crafter-do`, these guardrails are expressed as a **Karpathy Contract** in the plan. Each phase and step defines outcome, scope boundary, non-goals, simplicity constraint, drift criteria, verification evidence, and stop conditions. Implementers work inside that contract, Verifiers check drift against it, and Reviewers score the completed phase against it.
