# Philosophy

## Why Crafter Exists

Crafter was born from frustration.

Existing AI development frameworks are either too heavy or too hands-off. GSD has the right instincts — context engineering, task planning, verification criteria — but wraps them in an unwieldy monolith with auto-commits and machine-readable XML plans that feel more like configuring a build system than working with a collaborator.

Crafter takes the best ideas from both and strips away the overhead. It's a lightweight set of conventions, context files, and skills designed for a single experienced developer who knows what they want.

---

## Core Principles

### Craftsmanship
You are the craftsman. Claude is your tool. The developer's judgment, taste, and intent drive every decision — Claude executes and advises, it does not decide.

### Human-in-the-loop at every decision point
No auto-commits. No silent refactors. No guessing when the request is ambiguous. Every significant action — plan approval, diff review, commit — requires explicit developer consent.

### Conversational
Plans are written in plain language, for a human reader. Not XML. Not structured task objects. Not pipe-delimited fields. If you can't explain the plan clearly in a few paragraphs, the plan isn't ready yet.

### Adaptive
One command (`/crafter-do`) adapts to the size of the task. A one-line fix and a cross-cutting refactor both go through the same command — the workflow adjusts to match the scope automatically.

### Persistent context
Three living documents in `.crafter/` give Crafter workflows persistent project context without relying on global session preloads. They grow as the project evolves and are updated after every significant change.

---

## Orchestrator / Agent Architecture

Crafter commands run as orchestrators: the main context window manages the workflow and communicates with the developer, while specialized agents do the actual work in fresh, isolated context windows.

This matters because running planning, implementation, verification, and review all in one context leads to context rot, compaction, and hallucinations as the conversation grows. Each agent starts clean with only the context it needs.

Five roles cover the full workflow:

- **Planner** — proposes the implementation plan
- **Implementer** — executes the approved plan
- **Verifier** — checks verification criteria and looks for regressions
- **Reviewer** — reviews the diff for bugs, security issues, and plan deviations
- **Analyzer** — reads and maps the codebase for research and architecture work

Verify and Review can run in parallel after implementation, since their inputs are independent. This is expected behavior — not a shortcut — and it speeds up the workflow without sacrificing correctness.

---

## Karpathy-Inspired Guardrails

Every change passes through four checkpoints before it is considered done:

- **Think Before Coding** — surface assumptions explicitly; if multiple interpretations exist, present them instead of picking silently
- **Simplicity First** — prefer the smallest change that solves today's requirement; avoid speculative abstractions
- **Surgical Changes** — every changed line must trace to the approved request; no drive-by refactors
- **Goal-Driven Execution** — convert work into verifiable criteria and iterate until each criterion is satisfied

These apply across planning, implementation, and review — not just at one stage.

---

## What We Took from GSD

- **Context files** — PROJECT, ARCHITECTURE, STATE as the foundation of persistent context
- **Verification criteria in planning** — define how you'll know it's done before you start
- **Fresh context per task** — re-read context files at the start of every command
- **Agent specialization** — different roles for planning, execution, verification, and review

## What We Left Behind from GSD

- Auto-commits
- XML task plans
- Rigid, multi-phase pipeline with no escape hatches
- Excessive ceremony for small tasks

---

## Target User

Crafter is for an experienced developer who:

- Knows what they want to build
- Values code quality and thoughtful decisions over raw speed
- Prefers control over automation
- Finds existing AI frameworks either too rigid or too opaque
- Wants a collaborator, not an autopilot
