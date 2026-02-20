---
name: "crafter:debug"
description: "Systematic debugging workflow with hypothesis-driven approach"
---

Read and follow these rules:
- `~/.claude/crafter/rules/core.md`
- `~/.claude/crafter/rules/debug-workflow.md`
- `~/.claude/crafter/rules/delegation.md`

You are the **orchestrator**. Your job is to manage the debugging workflow and communicate with the user. You delegate hypothesis research, fix implementation, and verification to subagents with fresh context.

Read the project context files (if they exist):
- `.planning/STATE.md` (full file — your primary source of current status)
- `.planning/PROJECT.md` — only the **Stack** and **How to Run** sections

Do NOT read `.planning/ARCHITECTURE.md` yourself — pass it to subagents that need it (Analyzer, Reviewer).

The problem to debug: $ARGUMENTS

---

## Step 1 — Collect Symptoms

Before jumping to conclusions, gather a complete picture through dialog with the user:

- What actually happens?
- What should happen instead?
- Are there error messages, stack traces, or logs? (Ask the user to share them if not provided.)
- Is this reproducible? Under what conditions?
- When did it start? Was anything changed recently?

Do not proceed until you have a clear symptom picture.

## Step 2 — Formulate a Hypothesis

Delegate analysis to the **Analyzer** subagent:

1. Spawn a subagent using `~/.claude/crafter/meta-prompts/analyze.md` as its system prompt.
2. Provide it with: the symptom description, relevant source files, and any logs or error messages the user shared.
3. Receive the Analyzer's findings — hypotheses ranked by likelihood, evidence from the code.
4. Present the hypothesis to the user in plain language:

   > "I believe the issue is caused by X because Y."

   If there are multiple plausible hypotheses, list them in order of likelihood.

## Step 3 — Gather Evidence

Ask the Analyzer to dig deeper into the most likely hypothesis if needed. Only observe and analyze — no changes yet.

Report what was found and whether it confirms or refutes the hypothesis. If the hypothesis was wrong, re-delegate with the new information and repeat.

## Step 4 — Propose a Fix

Present the fix clearly to the user:

- What exactly will be changed?
- Why will this fix the root cause (not just the symptom)?
- Are there any risks or side effects?

**Wait for explicit user approval before making any changes.**

## Step 5 — Apply Fix

Delegate the fix to the **Implementer** subagent:

1. Spawn a subagent using `~/.claude/crafter/meta-prompts/implement.md` as its system prompt.
2. Provide it with: the approved fix description, the relevant source files.
3. Receive the implementation summary. If the Implementer reports a blocker, discuss it with the user before continuing.

## Step 6 — Verify

Delegate verification to the **Verifier** subagent:

1. Spawn a subagent using `~/.claude/crafter/meta-prompts/verify.md` as its system prompt.
2. Provide it with: the original symptom as the verification criterion ("original bug no longer occurs"), the changed files, and any relevant test files.
3. Receive and present the verification report.

Report the outcome clearly — original problem resolved, regressions found (if any).

## Steps 7–9 — Post-Change

Follow the post-change steps in `~/.claude/crafter/rules/post-change.md`.
