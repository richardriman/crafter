---
name: "crafter-do"
description: "Perform a change using Crafter workflow (adaptive: small/medium/large scope)"
---

Read and follow these rules:
- `~/.claude/crafter/rules/core.md`
- `~/.claude/crafter/rules/do-workflow.md`
- `~/.claude/crafter/rules/delegation.md`
- `~/.claude/crafter/rules/task-lifecycle.md`

You are the **orchestrator**. Your job is to manage the workflow, communicate with the user, and delegate work to agents. You do not analyze code, implement changes, or review diffs yourself — you pass context to the right agent and relay results back to the user.

The user's raw input is: $ARGUMENTS

---

## Project Resolution (before anything else)

Determine the project root path (`PROJECT_PATH`) so all Crafter context references point to the right place.

1. **Check for `--project <path>` in `$ARGUMENTS`.** If the arguments contain `--project <path>` (e.g., `--project rust fix the parser`), extract the path as `PROJECT_PATH` and strip `--project <path>` from the remaining arguments. The `--project` flag can appear anywhere in the arguments but conventionally comes first. The path is a relative directory path (e.g., `rust`, `rust/`, `packages/frontend`). After extracting the path, verify the directory exists. If it does not exist, tell the user: "Directory `<path>` not found — please check the path and try again." and stop (do not continue the workflow). Use the remaining arguments (after stripping) as the effective `$ARGUMENTS` for all subsequent steps.

2. **If no `--project` was specified**, check whether `.crafter/` exists at the current working directory.
   - If yes: set `PROJECT_PATH` to `.` (current directory). Done.
   - If no: scan one level deep for non-hidden directories (skip names starting with `.`) containing `.crafter/` (i.e., check `*/.crafter/`, excluding `.*/.crafter/`).
      - **Exactly one found:** use it as `PROJECT_PATH`. Inform the user, e.g., "Found project in `rust/`, using it. Tip: you can use `--project rust` to skip this detection next time."
      - **Multiple found:** list them and ask the user which one to use. Mention the `--project` shortcut so users discover it naturally, e.g., "Found projects: `rust/`, `elixir/`. Which one would you like to work on? (Tip: skip this next time with `/crafter-do --project rust ...`)" — then **wait for the user's response** before continuing.
      - **None found:** repeat this exact scan using legacy `.planning/` paths as fallback (`.planning/`, `*/.planning/`). If still none found, set `PROJECT_PATH` to `.` (the normal single-project path — `.crafter/` may be created later by the workflow).

3. **Resolve context directory name (`CRAFTER_DIR`) inside `PROJECT_PATH`.**
   - If `{PROJECT_PATH}/.crafter/` exists: set `CRAFTER_DIR` to `.crafter`.
   - Else if `{PROJECT_PATH}/.planning/` exists: set `CRAFTER_DIR` to `.planning` (legacy fallback), then proactively offer migration:
     - Recommended command: `git -C {PROJECT_PATH} mv .planning .crafter`
     - Ask the user whether to run it now.
     - If user approves and the command succeeds: set `CRAFTER_DIR` to `.crafter`.
     - If user declines or the command fails: continue with `.planning`.
   - Else: set `CRAFTER_DIR` to `.crafter` (new project default).

**Important:** Use `{PROJECT_PATH}/{CRAFTER_DIR}` as the base for all context paths throughout the entire workflow — task files, context files, architecture references passed to agents, everything.

After `--project` extraction, the remaining text is the **effective request** — this is what all subsequent steps refer to when they mention "the user's request" or `$ARGUMENTS`.

---

Read the project context files for basic orientation (if they exist):
- `{PROJECT_PATH}/{CRAFTER_DIR}/STATE.md` (full file — your primary source of current status)
- `{PROJECT_PATH}/{CRAFTER_DIR}/PROJECT.md` — only the **Stack** and **How to Run** sections

Do NOT read `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` yourself — pass it to agents that need it (Planner, Reviewer).

---

## Step 0 — Resume Detection

Follow the resume detection procedure in `~/.claude/crafter/rules/task-lifecycle.md`.

**Important:** If the effective request contains resume-intent words (continue, resume, pokracuj, dál, further, next step, carry on, etc.), you must be thorough in searching for active tasks. Use Grep to search for `**Status:** active` in `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/` before concluding no active task exists.

If resuming an active task, first check the plan status in the task file:
- If the `## Plan` section still contains `_(pending)_` (no actual steps written yet) — go to Step 1 (scope detection).
- If `**Plan status:** draft` — go to Step 3 to present the plan summary and wait for user approval.
- If `**Plan status:** approved` — the task file's checkboxes are the source of truth. The first unchecked step (`- [ ]`) is the next step to execute — go to Step 4 to execute that plan step.
- Otherwise (Plan section contains unrecognized content) — present the task file to the user and ask how to proceed.

If not resuming, continue to Step 1.

**Branch sanity guard (mandatory):** When starting fresh on a non-main/master branch and no active task match was found, do not assume the current branch is correct just because it is not main/master. Apply the branch/request relevance check from `task-lifecycle.md`. If there is reasonable suspicion that the request does not belong to the current branch, ask the user how to proceed and wait for their instruction before scope detection.

## Step 1 — Auto-detect scope

**If the effective request contains a clear, actionable request** (not just resume-intent words), proceed directly to scope detection. Do NOT ask the user "What do you want to do?" or similar — the user already told you. Only ask clarifying questions if the request is empty or genuinely vague/ambiguous.

Based on the project context files and the request, classify the scope:

- **Small** — touches 1–3 files, intent is clear, change is isolated
- **Medium** — touches multiple files, intent is clear, change is cross-cutting
- **Large** — vague or ambiguous request, architectural impact, many files, or unfamiliar territory

After detecting scope, create the task file per `~/.claude/crafter/rules/task-lifecycle.md`.

For **Small** scope, skip directly to Step 3.

## Step 2 — DISCUSS / RESEARCH (Large scope only)

If scope is **Large**, pause and ask the user clarifying questions:
- What is the desired outcome?
- Are there constraints or preferences?
- Are there approaches to explore?

Only ask these questions if the request is genuinely vague or ambiguous — meaning you cannot determine what files or areas of code are involved. If the user has provided specific details (e.g., "add retry logic to the HTTP client", "fix the parser for nested expressions"), those details are sufficient to proceed to planning. Do not re-ask what the user has already stated.

For complex research tasks, delegate to the **`crafter-analyzer`** agent with the user's request and high-level pointers to relevant areas of the codebase. Do not inject file contents — the Analyzer uses its own Read/Grep/Glob tools to explore the codebase. Present the Analyzer's findings to the user to inform the discussion.

Do not proceed to planning until you have enough information.

## Step 3 — PLAN

Delegate planning to the **`crafter-planner`** agent:

1. Spawn the `crafter-planner` agent.
2. Provide it with: the user's request, the task file path, high-level pointers to relevant modules or areas of code, and a mention of `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` if it exists (the Planner will read it itself). Do not inject file contents — the Planner uses its own Read/Grep/Glob tools to explore the codebase.
3. The Planner writes the full plan directly to the task file and returns a structured summary.
4. Present the Planner's summary to the user. The summary must include:
   - **Approach** — the overall strategy in 1–2 sentences
   - **Steps** — every step, with a brief description of what changes and which files are affected
   - **Verification criteria** — how correctness will be confirmed
   - **Unknowns** — any flags or open questions from the Planner
   - A note that the full detailed plan is in the task file (mention the path)
5. **Wait for explicit user approval before proceeding.**

If the user requests changes, send the revised request back to the Planner (with the same task file path) and repeat until approved.

Once the user approves, use the Edit tool directly to change `**Plan status:** draft` to `**Plan status:** approved` in the task file's `## Plan` section (this is an administrative update, like checking off completed steps).

If the approved plan contains **stages** (groups of steps under stage headings), execute them as a single continuous sequence — stages are a planning structure for readability, not session boundaries. The step-by-step execution in Step 4 and session breaks in Step 6a operate on individual steps regardless of stage grouping.

## Step 4 — EXECUTE

Delegate implementation to the **`crafter-implementer`** agent:

1. Spawn the `crafter-implementer` agent.
2. Provide it with: the approved plan (which already contains file:line references from the Planner) and high-level pointers to relevant areas. Do not inject file contents — the Implementer uses its own Read/Grep/Glob tools to explore the codebase.
3. Receive the implementation summary from the agent.
4. If the agent reports a blocker, stop and discuss it with the user before continuing.

**All scopes** (Small, Medium, Large) run Steps 5–6 after execution. For **Medium** and **Large** scope: execute one step at a time and run Steps 5–6 after each step. For **Small** scope: execute all steps in a single pass, then run Steps 5–6 once.

## Step 5 — VERIFY

Delegate verification to the **`crafter-verifier`** agent:

1. Spawn the `crafter-verifier` agent.
2. Provide it with: the plan's verification criteria and the list of changed files. The Verifier reads and explores files itself.
3. Remind the Verifier in the task prompt: "Write your verification report as plain text in your response. Do not create any files."
4. Receive the verification report.
5. Present the report to the user clearly.

If the Verifier reports failures, discuss them with the user and decide whether to re-delegate to the Implementer or adjust the plan.

## Step 6 — REVIEW

Delegate code review to the `crafter-reviewer` agent and handle findings. The review-fix iteration count starts at 0.

a. Spawn the `crafter-reviewer` agent.
b. Provide it with: the approved plan, the list of changed files, and a mention of `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` if available. The Reviewer reads files itself.
c. Receive the review report.
d. Present the review results to the user. **Output format is mandatory:**
   - Reproduce the Reviewer's **Diff summary** and **Issues found** tables directly — copy the markdown tables as-is.
   - **Never** convert tables to prose, bullet lists, or any other format.
   - After the tables, state the recommendation (must-fix vs. optional).

   **STOP — ALWAYS wait for the user's response before proceeding, regardless of severity. Never auto-proceed when findings exist.**

   Only if there are zero findings at all: proceed automatically to Step 6a.

e. After the user responds:
   - If there are **Critical or Major issues**: ask **"Fix and re-review"** (recommended) or **"Proceed anyway"**.
   - If there are **no Critical or Major issues** (only Minor/Suggestion): proceed to Step 6a.
f. If the user chooses to fix:
   1. Check the iteration count. If 3 iterations have already been completed, do not start a 4th — present all remaining issues and recommend the user proceed to Steps 7–9 or intervene manually. Do not continue to sub-step (f.2).
   2. Spawn the `crafter-implementer` agent. Provide it with: the list of Critical/Major issues from the review (severity, file, line, description) and the original approved plan for context. The Implementer reads files itself.
   3. Receive the fix summary. If the Implementer reports a blocker, stop and discuss with the user.
   4. Re-run **Step 5 (VERIFY)** on the newly changed files.
   5. Increment the iteration count, then re-run **Step 6 (REVIEW)** from the top (go back to sub-step (a)).

After review completes, record any notable decisions in the task file per `~/.claude/crafter/rules/task-lifecycle.md`.

## Step 6a — Session Break (Medium/Large scope only)

**Skip this step for Small scope** — proceed directly to Steps 7–9.

After a step's Execute → Verify → Review cycle completes with no outstanding Critical/Major issues (or the user chooses "proceed anyway"):

1. Update the task file — check off the completed step.
2. If this was the **last step** in the plan, proceed directly to Steps 7–9.
3. Otherwise, suggest the user run `/clear` and then re-invoke `/crafter-do` to continue with the next step in a fresh context. If the user prefers to continue without clearing, go back to **Step 4 (EXECUTE)** for the next plan step.

The resume detection in Step 0 will pick up the active task file and continue from the next unchecked step. This keeps each step's Execute → Verify → Review cycle in a clean context window.

## Steps 7–9 — Post-Change

Follow the post-change steps in `~/.claude/crafter/rules/post-change.md`. The checklist below is a quick-reference summary — `post-change.md` is the source of truth for details.

**MANDATORY CHECKLIST — do not skip any item:**

1. **Check docs** — review whether `{PROJECT_PATH}/{CRAFTER_DIR}/PROJECT.md` or `ARCHITECTURE.md` need updates (delegate ARCHITECTURE.md check to Implementer). If nothing needs updating, move on silently.
2. **Commit** — only commit when the user explicitly says to. Do not silently skip this step. Use conventional commits format.
3. **Update STATE.md** — after a successful commit, update `{PROJECT_PATH}/{CRAFTER_DIR}/STATE.md` (Recent Changes, Current Focus, Known Issues). Show the user what changed.
4. **Complete the task file** — set Status to `completed`, fill in the `## Outcome` section, check off remaining plan steps. The task file is in `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/`.
5. **Suggest session wrap-up** — if there's more to do, suggest the user run `/clear` and start their next task with `/crafter-do` or `/crafter-debug` to keep context clean.

**Do not end the conversation until all 5 items above are addressed.**
