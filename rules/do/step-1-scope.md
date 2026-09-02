# Completeness and scope (inline startup procedure 2)

**Fully orchestrator-side — do NOT delegate.** This is a judgement call over context you already hold; make it yourself. **Skipped entirely** when resume detection returned `resume-draft` or `resume-approved` — the plan already exists and the task file's `**Scope:**` field carries the classification.

## Completeness — synthesize, don't interview

Synthesize the completeness check from what is already available: the effective request, the project context files already in context, and anything you can find in the codebase yourself. Do not ask the user for facts you can look up, and do not re-ask for anything they already stated.

A request is complete enough to plan when these are clear: goal, scope, non-goals, acceptance criteria, constraints, risks, and validation strategy. For trivial requests this is a one-sentence assessment (e.g., "Complete: the requested one-line behavior and its verification are explicit."). For non-trivial requests, name the missing pieces explicitly — those, and only those, become the questions for Step 2.

If the effective request contains a clear, actionable request (not just resume-intent words), never ask "What do you want to do?" or similar — the user already told you.

## Scope

Classify the scope from the project context, the completeness check, and the request:

- **Small** — touches 1–3 files, intent is clear, change is isolated
- **Medium** — touches multiple files, intent is clear, change is cross-cutting
- **Large** — incomplete/vague request, architectural impact, many files, or unfamiliar territory

The classification is load-bearing beyond planning:

| Scope | Plan | Execution unit |
|---|---|---|
| Small | inline, 3–6 sentences in the task file — no Planner spawn | the whole task in one Implementer spawn |
| Medium | `crafter-planner`, flat step list, one contract | the whole task in one Implementer spawn |
| Large | `crafter-planner`, vertical phases, one contract per phase | one phase per Implementer spawn |

When scope is genuinely ambiguous, ask the user rather than guessing.

**Extension skill check — only when `--ext` is active.** Without the flag, skip this paragraph entirely. With it, check the extension skills discovered at startup (see `{CRAFTER_HOME}/rules/do/extension-skills.md`); if any skill's `When-Applies` matches the request, record their names and capabilities and pass them as supplemental context to the Analyzer in Step 2 or the Planner in Step 3. Extension skills may contribute domain-specific completeness criteria; they cannot replace the orchestrator's scope classification.

The task file is created only once the request is **complete enough to plan** — i.e. on entry to Step 3. If gaps remain, continue to Step 2 without creating it; Step 2 creates it once the gaps are closed. If the request is already complete enough to plan, create it now per `{CRAFTER_HOME}/rules/task-lifecycle.md`, recording the scope in the `**Scope:**` metadata field and respecting the main/master guard (use the approved topic branch, not `main/master`), then continue to Step 3.
