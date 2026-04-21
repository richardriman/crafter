# Task: Improve Plan Summary Detail

## Metadata
- **Status:** completed
- **Scope:** Medium
- **Created:** 2026-02-27

## Request

Three related improvements to the /crafter:do workflow:

1. **Planner writes full plan to task file:** The crafter-planner agent should write the full detailed plan directly into the task file's Plan section (the planner already receives task file path context). The orchestrator then presents a summary in conversation, but the user can read full details in the task file.

2. **More detailed plan summary in conversation:** The orchestrator's conversational summary of the plan should include more detail than currently — overall approach, all steps with what they do and which files they touch, and verification criteria.

3. **Fix review rule compliance:** The orchestrator sometimes ignores review rules — it should always stop and wait for user response when there are ANY findings (including Minor and Suggestion), and it should consistently use table format for review output.

## Plan

- [x] Step 1: Update planner agent — add Edit tool, task file writing instructions, structured summary output
- [x] Step 2: Update Step 3 (PLAN) in do.md — pass task file path, specify summary content, remove redundant write
- [x] Step 3: Update task-lifecycle.md — reflect that planner writes plan directly
- [x] Step 4: Strengthen review rules in do.md and do-workflow.md — MANDATORY gate, table format

### Step 1 — Update planner agent to write plans to the task file

**File: `~/.claude/crafter/agents/crafter-planner.md`**

a. **Line 4 (tools frontmatter)** — Add `Edit` to the tools list:
   `tools: Read, Grep, Glob, Bash` → `tools: Read, Edit, Grep, Glob, Bash`

b. **After line 13 (end of Context section)** — Add task file writing instructions:
   If the orchestrator provides a task file path, write the full plan into that file's `## Plan` section after producing the plan. Use the Edit tool to replace the existing content (typically `_(pending)_` or a previous draft) under the `## Plan` heading with the complete plan. Use checkboxes (`- [ ]`) for each plan step if multi-step. Do not modify any other section of the task file.

c. **Lines 44-46 (Output format section)** — Add structured summary instructions:
   After writing to the task file, also return a **structured summary** for conversation display:
   - **Approach** — 1-2 sentences on overall strategy
   - **Steps** — each step with brief description of what changes and which files
   - **Verification** — the verification criteria
   - **Unknowns** — any flags or open questions
   This summary is what the orchestrator shows in conversation. The full plan lives in the task file.

### Step 2 — Update Step 3 (PLAN) in command file

**File: `~/.claude/crafter/commands/do.md`**

a. **Lines 56-68** — Rewrite Step 3:
   - Sub-step 2: add "the task file path" to what's provided to the planner
   - Sub-step 3: change to "The Planner writes the full plan directly to the task file and returns a structured summary"
   - Sub-step 4: replace vague "Present the plan to the user clearly" with explicit list: Approach, Steps (with files), Verification criteria, Unknowns, note about full plan in task file
   - Remove/replace "After plan approval, update the task file with the approved plan" — plan is already in the task file from the planner

### Step 3 — Update task-lifecycle.md

**File: `~/.claude/crafter/rules/task-lifecycle.md`**

a. **Line 38** — Change:
   `- **After plan approval:** Write the approved plan to the Plan section. Use checkboxes...`
   → `- **After planning:** The Planner agent writes the full plan directly to the Plan section (with checkboxes for each step). No additional write is needed from the orchestrator after approval.`

### Step 4 — Strengthen review rules

**File: `~/.claude/crafter/commands/do.md`**

a. **Line 100 (Step 6, sub-step d)** — Replace with emphatic version:
   Present review using Reviewer's table format — reproduce Diff summary and Issues found tables directly. Do not convert to prose/bullets.
   **MANDATORY GATE — If ANY findings** (including Minor/Suggestion): **STOP and wait**. Do not proceed. Minor/Suggestion are informational (don't trigger fix loop) but user must see and acknowledge them.
   Only if zero findings: proceed automatically to Step 6a.

**File: `~/.claude/crafter/rules/do-workflow.md`**

b. **Lines 26-31 (REVIEW section)** — Add emphasis and format rule:
   - **MANDATORY: Always wait for the user's response when findings exist** — MUST present and STOP. Do not proceed automatically. Do not skip. The ONLY automatic proceed is zero findings.
   - **Always use table format for review output** — reproduce Reviewer's tables directly. Do not convert to prose or bullet lists.

### Verification Criteria

1. Planner agent has `Edit` in tools frontmatter
2. Planner agent has task file writing instructions in Context section
3. Planner agent returns structured summary (Approach, Steps, Verification, Unknowns)
4. Step 3 of do.md passes task file path to planner
5. Step 3 of do.md specifies summary content (approach, steps+files, verification, unknowns, task file ref)
6. Step 3 of do.md removes redundant task file write after approval
7. task-lifecycle.md reflects planner writes plan directly
8. Review gate uses MANDATORY/STOP emphasis in do.md
9. Table format specified in both do.md and do-workflow.md
10. Only four files modified

## Decisions

_(none yet)_

## Outcome

Commit 7722b9f — Planner writes full plan to task file with draft/approved lifecycle, orchestrator presents structured summary, review gate enforced as MANDATORY with table format, resume detection handles all plan states.
