---
name: "crafter:debug"
description: "Systematic debugging workflow with hypothesis-driven approach"
---

Read and follow all rules from `~/.claude/crafter/rules.md`.

Then read the project context files (if they exist):
- `.planning/PROJECT.md`
- `.planning/ARCHITECTURE.md`
- `.planning/STATE.md`

The problem to debug: $ARGUMENTS

---

## Step 1 — Collect Symptoms

Before jumping to conclusions, gather a complete picture:

- What actually happens?
- What should happen instead?
- Are there error messages, stack traces, or logs? (Ask the user to share them if not provided.)
- Is this reproducible? Under what conditions?
- When did it start? Was anything changed recently?

Do not form a hypothesis until you have a clear symptom picture.

## Step 2 — Formulate a Hypothesis

State your hypothesis explicitly:

> "I believe the issue is caused by X because Y."

Explain your reasoning. If there are multiple plausible hypotheses, list them in order of likelihood.

## Step 3 — Gather Evidence

Test your hypothesis by inspecting code, logs, or configuration. Do not make changes yet — only observe and analyze.

Report what you found and whether it confirms or refutes the hypothesis. If the hypothesis was wrong, form a new one and repeat this step.

## Step 4 — Propose a Fix

Describe the fix clearly:

- What exactly will you change?
- Why will this fix the root cause (not just the symptom)?
- Are there any risks or side effects?

**Wait for explicit user approval before making any changes.**

## Step 5 — Apply Fix

Implement the approved fix.

## Step 6 — Verify

Confirm the original problem is resolved:

- Re-run the scenario that triggered the bug.
- Run relevant tests.
- Check for regressions in related functionality.

Report the outcome clearly.

## Step 7 — Update STATE.md (if relevant)

If this bug was tracked in `.planning/STATE.md` (Known Issues), remove or update the entry. Show the user what changed.
