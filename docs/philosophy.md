# Philosophy

## Why Crafter Exists

Crafter was born from frustration.

Existing AI development frameworks are either too heavy or too hands-off. GSD has the right instincts — context engineering, task planning, verification criteria — but wraps them in an unwieldy monolith with auto-commits and machine-readable XML plans that feel more like configuring a build system than working with a collaborator. BMAD brings powerful multi-agent collaboration to the table, but also brings enterprise ceremony: 21+ agents, 50+ workflows, and a learning curve that can overshadow the actual work.

Crafter takes the best ideas from both and strips away the overhead. It's a lightweight set of conventions, context files, and Claude Code slash commands designed for a single experienced developer who knows what they want.

---

## Core Principles

### Craftsmanship
You are the craftsman. Claude is your tool. The developer's judgment, taste, and intent drive every decision — Claude executes and advises, it does not decide.

### Human-in-the-loop at every decision point
No auto-commits. No silent refactors. No guessing when the request is ambiguous. Every significant action — plan approval, diff review, commit — requires explicit developer consent.

### Conversational
Plans are written in plain language, for a human reader. Not XML. Not structured task objects. Not pipe-delimited fields. If you can't explain the plan clearly in a few paragraphs, the plan isn't ready yet.

### Adaptive
One command (`/crafter:do`) adapts to the size of the task. A one-line fix and a cross-cutting refactor both go through the same command — the workflow adjusts to match the scope automatically.

### Persistent context
Three living documents in `.planning/` give Claude fresh context at the start of every session without bloating `CLAUDE.md`. They grow as the project evolves and are updated after every significant change.

---

## What We Took from GSD

- **Context files** — PROJECT, ARCHITECTURE, STATE as the foundation of persistent context
- **Verification criteria in planning** — define how you'll know it's done before you start
- **Fresh context per task** — re-read context files at the start of every command

## What We Left Behind from GSD

- Auto-commits
- XML task plans
- Rigid, multi-phase pipeline with no escape hatches
- Excessive ceremony for small tasks

---

## What We Took from BMAD

- **Multi-perspective thinking** — via the BMAD party mode recommendation for big decisions
- **Agent specialization** — the idea that different lenses (PM, Architect, Developer, QA) produce better decisions

## What We Left Behind from BMAD

- 21+ agents
- 50+ workflows
- The assumption that every project needs enterprise-grade orchestration

---

## Target User

Crafter is for an experienced developer who:

- Knows what they want to build
- Values code quality and thoughtful decisions over raw speed
- Prefers control over automation
- Finds existing AI frameworks either too rigid or too opaque
- Wants a collaborator, not an autopilot
