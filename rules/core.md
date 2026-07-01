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

## Skill Detection: Caveman and Ponytail

At orchestrator startup, check for two marker files:

- `$HOME/.claude/.caveman-active` — exists only when the caveman skill is active; its content is the configured level (`lite`, `full`, or `ultra`).
- `$HOME/.claude/.ponytail-active` — exists only when the ponytail skill is active; its content is the configured level (`lite`, `full`, or `ultra`).

Record each skill's active state and level as workflow context available at every delegation point. When a marker is absent, the skill is inactive and **behavior is byte-for-byte unchanged** — no mention of the skills, no compression, no discipline changes.

### Caveman — communication compression discipline

Caveman removes filler words, pleasantries, and hedging while keeping ALL technical substance (code, file paths, identifiers, numbers, structured tables, and required output formats) verbatim and intact. Compression operates within the user's language — filler/hedging are dropped in whatever language the user uses; language-specific mechanics like "removes articles" apply only where the language has them.

- **lite:** light-touch compression — removes filler and hedging, still conversational.
- **full:** aggressive compression — drops prose scaffolding, maximizes density, suited to agent-to-agent output.

**Selection is audience-based, not marker-level-based.** The marker's `lite`/`full`/`ultra` value is presence-detection only for caveman — it is NOT a ceiling; an `ultra` marker value has no distinct caveman behavior and collapses to the same audience-based lite/full selection. Audience always wins:

- Human-facing prose → caveman-lite.
- Reasoning, inter-agent summaries, and pure agent-facing output → caveman-full.

#### Human-facing caveman-lite policy (orchestrator)

When caveman is active, the orchestrator applies caveman-lite to all of its own conversational output directed at the user. The following are **always excluded** from compression:

(a) **Verbatim-relayed content** — Reviewer diff sections, issue lists, scorecard tables, and Verifier reports the orchestrator reproduces without modification; relay them exactly as produced.
(b) **Human-in-the-loop gate prompts** — plan-approval questions and any other HITL gate where compression would reduce clarity or leave the user unable to make an informed decision.
(c) **Safety-critical content (Auto-Clarity)** — security warnings, irreversible-action confirmations, and multi-step sequences that must be written in full to be safely actionable.
(d) **Commits, PR titles/bodies, and release notes** — remain in neutral human project voice per `CLAUDE.md`; no compression applied.
(e) **Persistent-file English** — `.crafter/*`, saved plans, and task files are always English regardless of caveman state, per the Language Rules above.

**Jargon Confinement is unchanged by caveman.** The orchestrator must not import crafter-internal vocabulary onto the user's domain regardless of caveman level.

### Ponytail — YAGNI / the-ladder / shortest-working-diff code discipline

Ponytail enforces YAGNI, the-ladder, and shortest-working-diff discipline on code authoring and planning: prefer deletion over addition, stop at the first rung that holds, and never add speculative abstractions.

The marker's level (`lite`, `full`, or `ultra`) is **passed through** to the agents that receive the directive — no audience override; the user's configured intensity is honored.

**Ponytail applies only to `crafter-implementer` and `crafter-planner`.** It does not apply to the reviewer, verifier, analyzer, or step-runner.

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
