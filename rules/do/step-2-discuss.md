# Step 2 — DISCUSS / RESEARCH (when incomplete or uncertain)

Runs only when the completeness check found gaps. Ask about those gaps and nothing else — never re-ask what the user already stated.

## Grilling — frontier rounds

Ask in **rounds**. Each round contains only **frontier questions**: questions with no unresolved prerequisite among the questions still open. A question whose answer depends on another open question waits for the next round. This keeps every round answerable in one pass and stops the user from guessing at hypotheticals.

Number the questions continuously across rounds and give each one a recommended answer, so the user can accept the whole round with one word:

```
❓ Q1 — <short title>: <the question, one or two sentences>
➡️ Recommended: <your recommendation and the one-line reason for it>

❓ Q2 — <short title>: <the question>
➡️ Recommended: <your recommendation and the reason>
```

Rules:

- **Never ask about facts you can establish yourself.** Anything discoverable from the codebase, the git history, the project context files, or the environment is the Analyzer's job to find, not the user's job to answer. Ask only about intent, priorities, trade-offs, and decisions that are genuinely the user's to make.
- **The user decides.** A recommendation is a default to accept or override, never a decision already taken. Silence is not acceptance.
- Keep rounds small — a handful of questions, not an interrogation. Stop as soon as the request is complete enough to plan.

## Research delegation

For complex or codebase-dependent uncertainty, delegate to the **`crafter-analyzer`** agent with the user's request, the missing completeness fields, and high-level pointers to relevant areas of the codebase. Do not inject file contents — the Analyzer uses its own Read/Grep/Glob tools. Present the Analyzer's findings to the user to inform the round.

Do not proceed to planning until the request is complete enough to plan. Once it is complete, create the task file per `{CRAFTER_HOME}/rules/task-lifecycle.md` (if not already created) and continue to Step 3.
