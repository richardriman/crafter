# Task: Explicit STOP block in review workflow

## Metadata
- **Date:** 2026-03-06
- **Scope:** Small
- **Status:** active

## Request

Rewrite the REVIEW section (step 6d/e equivalent) in `do-workflow.md` so the mandatory wait-for-user gate when findings exist is an unmissable STOP block rather than conditional bullet points that LLMs tend to "optimize" into a single decision. Also update the corresponding REVIEW section in `~/.claude/crafter/rules/do-workflow.md` (same file, the inline skill text references the same rules).

The user's proposed structure:
- d: Present findings → **STOP — ALWAYS wait** (bold imperative, single line). Only zero findings → auto-proceed.
- e: After user responds → Critical/Major → ask fix-or-proceed; only Minor/Suggestion → proceed.

## Plan

**Plan status:** approved

Two files need changes. Two problems to fix:
1. The "wait for user" rule is buried in conditional bullet points — LLMs collapse it. Fix: standalone bold STOP line.
2. The table format requirement is weak — the orchestrator often converts tables to prose/bullets. Fix: make the format requirement an explicit imperative with an example of expected structure.

- [ ] **Step 1 — Rewrite the REVIEW section in `rules/do-workflow.md` (line 26)**

  Replace the current first bullet (line 26) with two clearly separated items:

  **Current text** (line 26): A single long `**MANDATORY: Always wait...**` paragraph that mixes the stop-and-wait rule with the severity-based follow-up logic.

  **New text** — replace the current first two bullets (lines 26-27) with three clearly separated items:

  1. **Output format (imperative):** Reproduce the Reviewer's tables exactly — Diff summary table and Issues found table. Never convert to prose, bullets, or any other format. Include an inline example of the expected table structure so the LLM has a concrete reference.

  2. **STOP gate:** On its own line, the imperative:

     `**STOP — ALWAYS wait for the user's response before proceeding, regardless of severity. Never auto-proceed when findings exist.**`

     Only if there are literally zero findings: proceed automatically.

  3. **After user responds:** Critical/Major issues trigger the "Fix and re-review" or "Proceed anyway" prompt. Only Minor/Suggestion findings → proceed.

  Keep all other bullets in the REVIEW section (lines 28-32) unchanged.

- [ ] **Step 2 — Rewrite Step 6d and 6e in `commands/do.md` (lines 110-117)**

  Replace the current sub-steps d and e with the same two-part structure:

  **Current text** (lines 110-117): Step 6d has the table-format instruction plus a MANDATORY GATE sub-bullet, and step 6e handles user response with severity logic.

  **New text:**

  - **6d.** Present review results to the user. **Output format is mandatory:**
    - Reproduce the Reviewer's **Diff summary** and **Issues found** tables directly — copy the markdown tables as-is.
    - **Never** convert tables to prose, bullet lists, or any other format.
    - After the tables, state the recommendation (must-fix vs. optional).

    Then, on its own line:

    **STOP — ALWAYS wait for the user's response before proceeding, regardless of severity. Never auto-proceed when findings exist.**

    Only if there are zero findings at all: proceed automatically to Step 6a.

  - **6e.** After the user responds:
    - Critical or Major issues → ask "Fix and re-review" (recommended) or "Proceed anyway".
    - No Critical or Major (only Minor/Suggestion) → proceed to Step 6a.

  Keep all surrounding steps (6a-c, 6f) unchanged.

**Files affected:**
- `/Users/ret/dev/ai/crafter/rules/do-workflow.md` — lines 25-32 (REVIEW section)
- `/Users/ret/dev/ai/crafter/commands/do.md` — lines 110-117 (Step 6d-e)

**Alternatives considered:**
- Adding a separate `### REVIEW GATE` heading to make it even more prominent. Rejected because it would break the existing structure where REVIEW is a single section, and would require changes to how the orchestrator references the workflow steps.
- Using ALL CAPS for the entire stop message. Rejected — bold on its own line is sufficient; all-caps reads as shouting and reduces readability of the rest of the section.

**Verification criteria:**
1. Both files have the STOP directive as a standalone bold line, not embedded in sub-bullets or conditional logic.
2. The logic is equivalent: zero findings → auto-proceed; any findings → STOP; after user responds, only Critical/Major trigger fix-or-proceed prompt.
3. Both files express the same logic consistently.
4. No other sections of either file are modified.

**Unknowns / flags:**
None — the scope and intent are clear.
