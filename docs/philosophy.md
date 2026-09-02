# Philosophy

## Why Crafter Exists

Crafter was born from frustration.

Existing AI development frameworks are either too heavy or too hands-off. GSD has the right instincts — context engineering, task planning, verification criteria — but wraps them in an unwieldy monolith that commits without a human-approved plan or a passed check, and in machine-readable XML plans that feel more like configuring a build system than working with a collaborator.

Crafter takes the best ideas from both and strips away the overhead. It's a lightweight set of conventions, context files, and skills designed for a single experienced developer who knows what they want.

---

## Core Principles

### Craftsmanship
You are the craftsman. Claude is your tool. The developer's judgment, taste, and intent drive every decision — Claude executes and advises, it does not decide.

### Human-in-the-loop at the decision points that matter
No silent refactors. No guessing when the request is ambiguous. Plan approval is always a human gate, Critical and Major findings stop the flow until they are resolved, and work that only a human can verify waits for explicit consent. Once the check pass closes clean — or with nothing worse than Minor/Suggestion findings recorded as tech-debt decisions — the commit lands automatically, because there is nothing left for the developer to decide.

### Conversational
Plans are written in plain language, for a human reader. Not XML. Not structured task objects. Not pipe-delimited fields. If you can't explain the plan clearly in a few paragraphs, the plan isn't ready yet.

### Vertical execution contracts
Plans describe outcomes, boundaries, and verification evidence — not line-by-line implementation recipes. Small and Medium tasks get one contract for the whole task; Large work is organized as vertical phases with one contract per phase. Each execution unit is implemented in a single pass and then checked as a whole, which keeps the check focused on a coherent unit of work rather than on every small implementation step.

### Adaptive
One command (`/crafter-do`) adapts to the size of the task. A one-line fix and a cross-cutting refactor both go through the same command — the workflow adjusts to match the scope automatically.

### Persistent context
Three living documents in `.crafter/` give Crafter workflows persistent project context without relying on global session preloads. They grow as the project evolves and are updated after every significant change.

---

## Orchestrator / Agent Architecture

Crafter commands run as orchestrators: the main context window manages the workflow and communicates with the developer, while specialized agents do the actual work in fresh, isolated context windows.

This matters because running planning, implementation, and checking all in one context leads to context rot, compaction, and hallucinations as the conversation grows. Each agent starts clean with only the context it needs.

Four roles cover the full workflow:

- **Planner** — proposes the implementation plan
- **Implementer** — implements the approved contract it was handed (the whole task, or a whole phase)
- **Checker** — checks drift against the contract and reviews the code for bugs, security issues, and unapproved deviations — in one pass
- **Analyzer** — reads and maps the codebase for research and architecture work

Each execution unit ends with one Checker pass: the whole task for Small and Medium, each phase for Large. Drift and code review are covered together in a single fresh-context pass, so there is no separate verification stage. When findings need fixing, the follow-up passes are narrowed to the files the fix touched.

---

## Change Guardrails

Every change passes through four checkpoints before it is considered done:

- **Think Before Coding** — surface assumptions explicitly; if multiple interpretations exist, present them instead of picking silently
- **Simplicity First** — prefer the smallest change that solves today's requirement; avoid speculative abstractions
- **Surgical Changes** — every changed line must trace to the approved request; no drive-by refactors
- **Goal-Driven Execution** — convert work into verifiable criteria and iterate until each criterion is satisfied

In `/crafter-do`, these guardrails are carried by the plan's contract: outcome, scope boundary, non-goals, seams, verification evidence, and stop conditions — defined once for the whole task (Small/Medium) or once per phase (Large).

These apply across planning, implementation, and checking — not just at one stage.

---

## What We Took from GSD

- **Context files** — PROJECT, ARCHITECTURE, STATE as the foundation of persistent context
- **Verification criteria in planning** — define how you'll know it's done before you start
- **Fresh context per task** — re-read context files at the start of every command
- **Agent specialization** — different roles for planning, execution, and checking

## What We Left Behind from GSD

- Commits without a human-approved plan or a passed check
- XML task plans
- Rigid, multi-phase pipeline with no escape hatches
- Excessive ceremony for small tasks

---

## Target User

Crafter is for an experienced developer who:

- Knows what they want to build
- Values code quality and thoughtful decisions over raw speed
- Wants control at the decision points and automation for everything else
- Finds existing AI frameworks either too rigid or too opaque
- Wants a collaborator on the decisions, not an autopilot on the judgment calls
