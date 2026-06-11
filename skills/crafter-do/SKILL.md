---
name: "crafter-do"
description: "Perform a change using Crafter workflow (adaptive: small/medium/large scope)"
fast: false
auto: false
---

Read and follow these rules:

- `{CRAFTER_HOME}/rules/core.md`
- `{CRAFTER_HOME}/rules/do-workflow.md`
- `{CRAFTER_HOME}/rules/delegation.md`
- `{CRAFTER_HOME}/rules/task-lifecycle.md`

## Skill options

In prose these flags are called `--fast` and `--auto`; in frontmatter they are set as `fast: true` and `auto: true`.

### `--fast` (default: off)

Set `fast: true` in this skill's frontmatter to enable silence-as-approval for phase summaries.

**Trade-off — speed vs. explicitness:**

- **With `--fast` on:** after the review loop closes clean and remaining Minor/Suggestion findings exist, the orchestrator presents the Phase Summary but treats user silence as implicit approval. Each remaining Minor/Suggestion finding is automatically recorded as a tech-debt entry in the task file's `## Decisions` section (format: `Decision (Tech Debt — auto-recorded): <severity> — <description>`), and the commit proceeds without waiting for an explicit "approved" response. Phases ship faster at the cost of reduced visibility into deferred findings.
- **Without `--fast` (default):** the orchestrator waits for an explicit affirmative response from the user before committing. Silence is never treated as approval. This is the safe, deliberate path — choose it when explicitness matters more than speed.

The `--fast` flag does **not** bypass the manual-verification exception: if a phase or step explicitly mentions non-automatable testing (UI, external integration), explicit user confirmation is always required regardless of this flag.

`--fast` and `--auto` are mutually exclusive — passing both produces a clear error and the workflow stops. See `### --auto (default: off)` below for unattended-orchestration semantics.

See Step 6b for the approval path that consumes this flag.

### `--auto` (default: off)

Set `auto: true` in this skill's frontmatter to enable fully unattended orchestration (Symphony, CI bots, or any non-interactive context).

After plan approval, the run executes Plan → Execute → Verify → Review → PR end-to-end without stopping for anything that does not threaten green commits. Phase summaries are not surfaced to the user; commits proceed automatically once a phase is green.

**`--auto` is mutually exclusive with `--fast`.** Passing both produces a clear parser-level error and the workflow does not start.

**Green-commit invariant:** `--auto` MUST never produce a non-green commit. If the auto-fix loop cannot bring the phase back to green within budget, the run exits with state and hands off to the orchestrator without committing. See `rules/do-workflow.md` → `### --auto (unattended orchestration)` for the full invariant statement.

**Four retained gates** (each is an exit + handoff via the task file, NOT an interactive pause). See `rules/do-workflow.md` → `#### Retained gates` for full descriptions:

- **Initial clarification** — Analyzer cannot understand the ticket well enough to produce a plan.
- **Plan approval** — PLAN.md is ready and awaiting human approval before execution begins.
- **Green-commit cap reached** — Critical/Major fix loop exhausted its iteration budget with findings still present.
- **Ad-hoc escape hatch** — genuinely blocked by something outside the fix-loop (missing auth/secret, hard contradiction, infrastructure outage, irrecoverable agent blocker).

Everything else (Critical/Major findings the auto-fix loop can clear within budget, manual-verification exception, Minor/Suggestion findings, Karpathy FLAGs, non-blocking drift, all phase-summary approval gates) is handled automatically — see `rules/do-workflow.md` → `#### Removed gates`.

See Step 6b for the approval-path branch that consumes this flag.

---

You are the **orchestrator**. Your job is to manage the workflow, communicate with the user, and delegate work to agents. You do not analyze code, implement changes, or review diffs yourself — you pass context to the right agent and relay results back to the user.

The user's raw input is: $ARGUMENTS

---

## Flag Validation (before anything else)

**Fully orchestrator-side — do NOT delegate.** Run this check inline:

`--auto` and `--fast` are mutually exclusive. If both flags are active (`auto: true` AND `fast: true` in frontmatter, or equivalent invocation context), produce a clear error and stop immediately — do not proceed to project resolution or any other step:

> Error: `--auto` and `--fast` are mutually exclusive — pass at most one. `--auto` strictly supersedes `--fast` per `rules/do-workflow.md` → `### --auto`.

## Project Resolution (before anything else)

**Fully orchestrator-side — do NOT delegate.** Run this procedure inline, then set `PROJECT_PATH` and `CRAFTER_DIR` before continuing:

1. **Check for `--project <path>` in `$ARGUMENTS`.** If present, extract the path as `PROJECT_PATH` and strip `--project <path>` from the remaining arguments. Verify the directory exists — if not, tell the user and stop. Use the remaining arguments as the effective `$ARGUMENTS` for all subsequent steps.

2. **If no `--project` was specified**, check whether `.crafter/` exists at the current working directory.
   - If yes: set `PROJECT_PATH` to `.`.
   - If no: scan one level deep for non-hidden directories containing `.crafter/` (skip names starting with `.`).
      - **Exactly one found:** use it as `PROJECT_PATH`. Inform the user and mention the `--project` shortcut.
      - **Multiple found:** list them and ask the user which one to use — **wait for the user's response** before continuing.
      - **None found:** repeat the scan using legacy `.planning/` paths. If still none found, set `PROJECT_PATH` to `.`.

3. **Resolve `CRAFTER_DIR` inside `PROJECT_PATH`.**
   - If `{PROJECT_PATH}/.crafter/` exists: set `CRAFTER_DIR` to `.crafter`.
   - Else if `{PROJECT_PATH}/.planning/` exists: set `CRAFTER_DIR` to `.planning`; proactively offer migration (`git -C {PROJECT_PATH} mv .planning .crafter`); ask the user and wait for a response; if approved and succeeds set `CRAFTER_DIR` to `.crafter`, otherwise continue with `.planning`.
   - Else: set `CRAFTER_DIR` to `.crafter`.

Use `{PROJECT_PATH}/{CRAFTER_DIR}` as the base for all context paths throughout the entire workflow.

---

Read the project context files for basic orientation (if they exist):
- `{PROJECT_PATH}/{CRAFTER_DIR}/STATE.md` (full file — your primary source of current status)
- `{PROJECT_PATH}/{CRAFTER_DIR}/PROJECT.md` — only the **Stack** and **How to Run** sections

Do NOT read `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` yourself — pass it to agents that need it (Planner, Reviewer).

---

## Extension Skills

Spawn the **`crafter-step-runner`** agent. Pass: step id `extension-skills`, and `{PROJECT_PATH}`. The agent internally reads `{CRAFTER_HOME}/rules/do/extension-skills.md`, performs the discovery scan across all three priority locations — (1) project-local (`{PROJECT_PATH}/.claude/crafter/skills/`), (2) parent-project (walk up parent directories for the first `../.claude/crafter/skills/` found), and (3) global (`{CRAFTER_HOME}/skills/`) — and returns a structured summary listing any compatible extension skills found (name, location, `When-Applies` clause) and confirming the supplemental-only invariant.

Act on the returned summary: record the discovered extension skills (if any) for use as supplemental context in Steps 1, 4, and 6. Do not invoke extension skills yourself — pass their names and capabilities to the relevant agent when delegating those steps.

---

## Workflow Master Plan (navigation map)

Use this section to route the entire workflow without loading any step module into context. Each row names the step, its purpose, and where it goes next.

### Full step sequence (in order)

| Step | Purpose | Routes to |
|------|---------|-----------|
| Flag Validation | Reject invalid flag combos (e.g. `--auto` + `--fast`), set active flags | Project Resolution |
| Project Resolution | Resolve `PROJECT_PATH` and `CRAFTER_DIR` | Read project context |
| Read project context | Load `STATE.md` (full) and `PROJECT.md` (Stack + How to Run only) | Extension-skill discovery |
| Extension-skill discovery | Discover and load any extension skill; apply supplemental-only rule | Step 0 |
| **Step 0** — Resume Detection | Detect active task file; determine resume entry point (see below) | Step 1 (new run or plan pending), Step 3 (draft plan), or Step 4 (approved plan) |
| **Step 1** — Completeness & Scope | Lightweight completeness check; classify scope (Small/Medium/Large) | Step 2 (if gaps remain) or Step 3 (if complete enough to plan) |
| **Step 2** — Discuss / Research | Resolve gaps via clarifying questions or `crafter-analyzer` delegation | Step 3 (once complete enough to plan) |
| **Step 3** — Plan + Approval Gate | Delegate planning to `crafter-planner`; present plan summary; await explicit user approval; change status to `approved` | Step 4 |
| **Step 4** — Execute (one step at a time) | Delegate one plan step to `crafter-implementer`; after each step → Step 5; after all steps in a phase pass → Step 5a then Step 6 | Step 5 (after each step) |
| **Step 5** — Step Drift Check | Delegate to `crafter-verifier` (mode: step drift check); handle recommended action (see routing chains below) | Step 4 next step (continue / record & continue / fix & re-check) or Step 3 (replan) |
| **Step 5a** — Phase Verification | Delegate to `crafter-verifier` (mode: phase verification); if fails → discuss + re-delegate or adjust plan | Step 6 (on pass) |
| **Step 6** — Review + Fix Loop | Delegate to `crafter-reviewer`; run fix loop for Critical/Major (up to 5 iterations); re-runs Step 5a then Step 6 per iteration | Step 6b (on clean) |
| **Step 6b** — Phase Summary + Commit | Choose approval path (see flag branching below); commit on approval | Step 6a (Medium/Large, non-last phase) or Steps 7–9 (Small scope or last phase) |
| **Step 6a** — Session Break | Medium/Large only: suggest `/clear` + re-invoke for next phase; Step 0 resumes at next unchecked step | Step 4 (next phase), or Steps 7–9 (plan complete) |
| **Steps 7–9** — Post-Change | Docs check, consolidated end-of-task commit, `STATE.md` update, task-file completion, session wrap-up | Step 9b (`--auto` only) or session wrap-up |
| **Step 9b** — PR Composition | `--auto` only, after Steps 7–9: compose PR body, open PR via `gh pr create`, print PR URL | Session wrap-up |

### Scope branching

| Scope | Step 6a behavior |
|-------|-----------------|
| **Small** | Skip Step 6a entirely — go straight from Step 6b to Steps 7–9 |
| **Medium / Large** | Run Step 6a between phases; suggest `/clear` + re-invoke |

### Flag branching

**Step 6b approval paths** (first matching path applies for non-`--auto`):

1. **`--auto`:** commit automatically under the green-commit invariant; record auto-fixed and tech-debt `Decision` entries; no interactive pause.
2. **Zero findings + no manual-verification exception:** auto-approve — present one-line notice and commit immediately.
3. **`--fast` + Minor/Suggestion findings remain:** silence-as-approval — present Phase Summary, treat next user turn as approval, record each deferred finding as a `Decision (Tech Debt — auto-recorded)` entry, then commit. Manual-verification exception still requires explicit confirmation.
4. **Otherwise (default):** explicit approval — present Phase Summary and wait for affirmative response; silence is never approval.

**Step 9b (`--auto` only):** runs ONLY when `auto: true`. Non-`--auto` runs never execute Step 9b.

### Resume entry points (from Step 0)

| Task-file plan status | Resume at |
|-----------------------|-----------|
| Plan section still `_(pending)_` | **Step 1** (scope / completeness check) |
| `**Plan status:** draft` | **Step 3** (present plan summary, await approval) |
| `**Plan status:** approved` | **Step 4**, at the first unchecked step — unless all steps in the current phase are checked and a phase verification / review gate is pending, in which case resume at that gate |

### High-risk routing chains

- **Step 5 drift → replan:** Step 5 Verifier recommends `replan` → return to **Step 3** with the new discovery.
- **Step 6 fix loop:** Critical/Major found → re-delegate fix to `crafter-implementer` → re-run **Step 5a** (phase re-verification) → re-run **Step 6** (review from top) → increment iteration count → if 5 iterations exhausted with findings still present → present options (manual override / accept-without-commit / replan-and-abort) or exit with state under `--auto`; if Verifier in fix-loop iteration recommends `replan` → return to **Step 3**.
- **Step 6b → Step 6a (Medium/Large, non-last phase):** after commit, run Step 6a session break; Step 0 resumes at next unchecked step or pending gate when re-invoked.
- **Step 6a → next Step 4 or Steps 7–9:** if the phase is complete and plan is complete → proceed to **Steps 7–9**; otherwise → Step 4 (next phase's first step).

---

## Step 0 — Resume Detection

Spawn the **`crafter-step-runner`** agent. Pass: step id `step-0-resume`, `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/` path, the effective `$ARGUMENTS` (after `--project` extraction), and the current branch name. The agent internally reads `{CRAFTER_HOME}/rules/do/step-0-resume.md` and `{CRAFTER_HOME}/rules/task-lifecycle.md`, searches for active task files (using the resume-intent word list and the `^- \*\*Status:\*\* active$|^\*\*Status:\*\* active$` grep pattern — two alternatives, the second handles task files whose `Status:` line is not a list item), applies the branch-sanity and main/master guards, and returns a structured summary: resume-status (`new-run` / `resume-pending` / `resume-draft` / `resume-approved`), active task file path (if any), plan status, next unchecked step (if resuming), branch mismatch details (if any), and any branch/guard question that requires the user's response.

Act on the returned summary:
- If the summary includes a **branch mismatch or guard question**: stop and ask the user as described; wait for their instruction before continuing.
- `new-run` or `resume-pending` (plan section still `_(pending)_`): continue to **Step 1**.
- `resume-draft` (`**Plan status:** draft`): skip to **Step 3** (present plan summary, await approval).
- `resume-approved` (`**Plan status:** approved`): skip to **Step 4** at the first unchecked step (or the pending phase gate if all steps in the current phase are checked).

## Step 1 — Completeness and scope

Spawn the **`crafter-step-runner`** agent. Pass: step id `step-1-scope`, the effective `$ARGUMENTS`, the `STATE.md` and `PROJECT.md` excerpts already in context, the list of discovered extension skills (from the Extension Skills step), and `{PROJECT_PATH}/{CRAFTER_DIR}`. The agent internally reads `{CRAFTER_HOME}/rules/do/step-1-scope.md`, runs the lightweight completeness check, classifies scope (Small/Medium/Large), applies the extension-skill supplemental-only check, and returns a structured summary: completeness verdict, scope classification, missing fields (if any), and whether the request is complete enough to plan.

Act on the returned summary:
- If **missing fields** exist (request not complete enough to plan): continue to **Step 2**.
- If **complete enough to plan**: create the task file per `{CRAFTER_HOME}/rules/task-lifecycle.md` (respecting the main/master guard — use the approved topic branch, not `main/master`), then continue to **Step 3**.

## Step 2 — DISCUSS / RESEARCH (when incomplete or uncertain)

Spawn the **`crafter-analyzer`** agent. Pass: the effective `$ARGUMENTS`, the missing completeness fields identified in Step 1, and high-level pointers to relevant areas of the codebase. Do not inject file contents — the Analyzer uses its own Read/Grep/Glob tools. The agent internally reads `{CRAFTER_HOME}/rules/do/step-2-discuss.md`, resolves gaps via targeted clarifying questions or codebase exploration, and returns a structured summary of findings and any remaining open questions.

Act on the returned summary: present the Analyzer's findings to the user to inform the discussion. Do not proceed to planning until the request is complete enough to plan. Once complete, create the task file per `{CRAFTER_HOME}/rules/task-lifecycle.md` and continue to **Step 3**.

## Step 3 — PLAN

Spawn the **`crafter-planner`** agent. Pass: the complete user request, the completeness/refinement notes, the task file path, high-level pointers to relevant modules or areas of code, and a mention of `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` if it exists. Do not inject file contents — the Planner uses its own Read/Grep/Glob tools. The agent internally reads `{CRAFTER_HOME}/rules/do/step-3-plan.md`, writes the full plan directly to the task file, and returns a structured summary covering: Approach, Phases/steps, Assumptions, Karpathy Contract, Verification criteria, and Risks.

**Orchestrator-only residue (NOT delegated):**

1. Present the Planner's structured summary to the user.
2. **Wait for explicit user approval before proceeding.** Silence is not approval.
3. If the user requests changes, re-spawn the Planner with the revised request and the same task file path; repeat until approved.
4. Once the user approves, use the **Edit tool directly** to change `**Plan status:** draft` to `**Plan status:** approved` in the task file's `## Plan` section.
5. Continue to **Step 4**. If the approved plan contains phases, execute one step at a time.

## Step 4 — EXECUTE

**Orchestrator-only pre-check (NOT delegated):** Before delegating, check whether any extension skill discovered at startup has a `When-Applies` clause matching the current step. If any match, include their names and capabilities in the context provided to the Implementer as supplemental domain-specialist context. Extension skills cannot replace the Implementer as writer or decision-maker for any step.

Spawn the **`crafter-implementer`** agent. Pass: the current step contract, phase context, relevant areas, non-goals, drift criteria, verification evidence, accepted deviations, stop conditions, and the names/capabilities of any matching extension skills. Do not inject file contents — the Implementer uses its own Read/Grep/Glob tools. The agent internally reads `{CRAFTER_HOME}/rules/do/step-4-execute.md` and returns an implementation summary.

**Orchestrator-only residue (NOT delegated):**

- If the agent reports a **blocker**: stop and discuss it with the user before continuing.
- After each step: run **Step 5** (drift check).
- After all steps in a phase pass drift checks: run **Step 5a** (phase verification) then **Step 6** (phase review).

## Step 5 — STEP DRIFT CHECK

Spawn the **`crafter-verifier`** agent. Pass: mode `step drift check`, the current step contract, phase context, non-goals, implementer summary, accepted deviations, changed files, and permission to inspect relevant `git diff` output. Include the reminder in the task prompt: "Write your verification report as plain text in your response. Do not create any files." Do not inject file contents — the Verifier reads and explores files itself. The agent internally reads `{CRAFTER_HOME}/rules/do/step-5-drift.md` and returns a verification report with a recommended action.

**Orchestrator-only residue (NOT delegated):** Present the report to the user clearly and handle the recommended action:

- **continue:** check off the completed step in the task file and continue.
- **record decision and continue:** append a `Decision (Orchestrator Accepted)` entry to the task file's `## Decisions` section and continue.
- **fix current step:** re-delegate the current step to the `crafter-implementer` agent before continuing.
- **ask user:** stop and ask the user whether to accept the drift, revise scope, or replan; wait for the user's response. If accepted, append a `Decision (User Accepted)` entry.
- **replan:** return to **Step 3** with the new discovery.

## Step 5a — PHASE VERIFICATION

Spawn the **`crafter-verifier`** agent. Pass: mode `phase verification`, the approved phase contract, phase verification criteria, accepted deviations, and the list of changed files. Include the reminder in the task prompt: "Write your verification report as plain text in your response. Do not create any files." The agent internally reads `{CRAFTER_HOME}/rules/do/step-5a-phase-verification.md` and returns a verification report.

**Orchestrator-only residue (NOT delegated):** Present the verification report. If phase verification fails, discuss the result with the user and decide whether to re-delegate to the Implementer, adjust the plan, or re-run a specific step drift check.

## Step 6 — REVIEW

**Orchestrator-only pre-check (NOT delegated):** Before delegating, check whether any extension skill discovered at startup has a `When-Applies` clause matching the current phase. If any match, include their names and capabilities in the context provided to the Reviewer as supplemental review context. Extension skill findings are advisory only and cannot replace the Reviewer's report or verdict.

Spawn the **`crafter-reviewer`** agent. Pass: the approved phase contract, accepted deviations, the list of changed files, and a mention of `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` if available. The agent internally reads `{CRAFTER_HOME}/rules/do/step-6-review.md` and returns a review report. Initialize the fix-loop **iteration count at 0** before the first review.

**Orchestrator-only residue (NOT delegated):** Handle the review output as follows:

a. Reproduce the Reviewer's output verbatim:
   - Copy the **Diff summary** and **Issues found** tables as-is.
   - Copy the **Karpathy scorecard** table as-is.
   - Copy the **Contract deviations** section as-is.
   - Never convert tables to prose, bullet lists, or any other format.
   - After the tables, state the recommendation (must-fix vs. optional).

b. **STOP — ALWAYS wait for the user's response before proceeding, regardless of severity. Never auto-proceed when findings exist.**

   Only if there are zero findings at all: proceed directly to **Step 6b** (auto-approve path) without waiting.

c. After the user responds:
   - If there are **no Critical or Major issues** (only Minor/Suggestion or none): proceed to **Step 6b**.
   - If there are **Critical or Major issues**: on user acknowledgement, enter the fix loop — there is no "Proceed anyway" choice for those severities. Go to sub-step (d).

d. Fix loop for Critical/Major issues:
   1. Check the iteration count. If 5 iterations have already been completed, do NOT start a 6th. Present all remaining Critical/Major findings and ask the user to choose one of:
      - **(a) manual override** — authorize manual iteration beyond the cap; re-enter the fix loop only on explicit user instruction.
      - **(b) accept-without-commit** — accept unresolved findings and proceed without committing this phase; record a Decision entry noting the unresolved findings and that the green-commit invariant is deliberately broken for this phase.
      - **(c) replan-and-abort** — abandon the current phase and return to planning.
      Under `--auto`, do NOT present the (a)/(b)/(c) choice — exit with state per `rules/do-workflow.md` → `### --auto (unattended orchestration)`.
      Do not continue to sub-step (d.2) until the user has chosen (non-`--auto`).
   2. Spawn the `crafter-implementer` agent. Pass: the list of Critical/Major issues (severity, file, line, description), the approved phase contract, and accepted deviations. The Implementer reads files itself.
   3. Receive the fix summary. If the Implementer reports a blocker, stop and discuss with the user.
   4. Re-run **Step 5a (PHASE VERIFICATION)** on the newly changed files.
   5. Increment the iteration count, then re-run **Step 6 (REVIEW)** from the top (go back to sub-step above).

e. After review completes, record any notable decisions in the task file's `## Decisions` section per `{CRAFTER_HOME}/rules/task-lifecycle.md`.

## Step 6b — Phase Summary and Auto-Commit

**Fully orchestrator-side — do NOT delegate.** After the review loop closes clean (no Critical or Major findings remain), choose the first approval path that applies:

#### `--auto` branch (runs before paths 1–3)

When `--auto` is set (`auto: true` in frontmatter):

1. Append `Decision (Auto-Fixed): <severity> — <description>` entries for any Critical/Major findings the fix loop cleared.
2. Append `Decision (Tech Debt — auto-recorded): <severity> — <description>` entries for any remaining Minor/Suggestion findings.
3. Record any manual-verification requirements as UAT buffer entries via the `crafter-buffer` skill.
4. Commit automatically per `{CRAFTER_HOME}/rules/post-change.md` under the green-commit invariant (see `rules/do-workflow.md` → `### --auto (unattended orchestration)`).

When `--auto` is **not** set, fall through to paths (1)–(3):

#### (1) Auto-approve on clean summary

Conditions: zero remaining findings of any severity in the final review state.

**Exception — manual verification:** if the phase plan or any of its steps explicitly requires manual testing (UI interaction, external integration, non-automatable scenarios), always wait for explicit user confirmation even on a fully clean summary — this exception overrides auto-approve and is not bypassed by `--fast`.

When auto-approve applies: present a one-line notice ("Phase clean — committing automatically.") and proceed directly to the commit per `{CRAFTER_HOME}/rules/post-change.md`.

#### (2) Silence-as-approval (`--fast`)

Conditions: `--fast` flag active AND remaining Minor/Suggestion findings exist.

Present the Phase Summary and wait for the user's next turn; if that turn does not raise concerns, treat it as implicit approval. Record each remaining Minor/Suggestion finding as `Decision (Tech Debt — auto-recorded): <severity> — <description>`. Then commit per `{CRAFTER_HOME}/rules/post-change.md`. The manual-verification exception from path (1) still applies.

#### (3) Explicit approval (default)

Conditions: remaining Minor/Suggestion findings exist AND `--fast` is not set.

Present the Phase Summary and wait for an affirmative response. **Silence does not count as approval.** Do not proceed until the user explicitly confirms (e.g., "approved", "looks good", "proceed").

#### Commit

On approval (any path), run the commit per `{CRAFTER_HOME}/rules/post-change.md`. After committing, continue to **Step 6a** (session break, Medium/Large scope) or **Steps 7–9** (last phase or Small scope).

## Step 6a — Session Break (Medium/Large scope only)

**Fully orchestrator-side — do NOT delegate.** Skip this step for Small scope — proceed directly to Steps 7–9.

Apply these routing rules:

1. If this was the **last step in the current phase**: proceed to **Step 5a** (Phase Verification) and **Step 6** (Review).
2. If this was the **last step in the entire plan** and phase verification/review are complete: proceed directly to **Steps 7–9**.
3. Otherwise: suggest the user run `/clear` and re-invoke `/crafter-do` to continue with the next step in a fresh context. If the user prefers to continue without clearing, go back to **Step 4 (EXECUTE)** for the next plan step.

The resume detection in Step 0 will pick up the active task file and continue from the next unchecked step or pending phase gate.

## Steps 7–9 — Post-Change

The final per-phase commit has already landed via Step 6b. These steps cover end-of-task follow-up work. `{CRAFTER_HOME}/rules/post-change.md` is the source of truth for commit and follow-up details.

**Orchestrator-only pre-delegation (NOT delegated):** Check whether `{PROJECT_PATH}/{CRAFTER_DIR}/PROJECT.md` needs updates yourself — but note that only the **Stack** and **How to Run** sections were loaded at startup. If the change may affect other sections of PROJECT.md (e.g., Overview, Architecture, Decisions), read the full file before deciding. For `ARCHITECTURE.md`, spawn the **`crafter-implementer`** agent and ask it to check whether `ARCHITECTURE.md` needs updates given what was changed in this task; pass the task summary and changed files. Receive the Implementer's recommendation.

**MANDATORY CHECKLIST — do not skip any item:**

1. **Check docs** — review whether `{PROJECT_PATH}/{CRAFTER_DIR}/PROJECT.md` or `ARCHITECTURE.md` need updates (delegate ARCHITECTURE.md check to Implementer as described above).
2. **Consolidated end-of-task commit** — if any PROJECT.md/ARCHITECTURE.md updates, a skillbook entry, or STATE.md changes exist, bundle them into one single consolidated commit per `{CRAFTER_HOME}/rules/post-change.md`; if none of those updates are needed, no follow-up commit is created.
3. **Update STATE.md** — update `{PROJECT_PATH}/{CRAFTER_DIR}/STATE.md` (Recent Changes, Current Focus, Known Issues) and include this update in the consolidated commit.
4. **Complete the task file** — set Status to `completed`, fill in the `## Outcome` section, check off remaining plan steps (file is in `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/`).
5. **Suggest session wrap-up** — if there is more to do, suggest the user run `/clear` and start the next task with `/crafter-do` or `/crafter-debug`.

**Do not end the conversation until all 5 items above are addressed.**

## Step 9b — PR Composition (`--auto` only) — compose PR body, open PR, print PR URL (`PR opened: <URL>`)

**Fully orchestrator-side — do NOT delegate.** Runs ONLY when `--auto` is set (`auto: true` in frontmatter), and ONLY after Steps 7–9 complete. Non-`--auto` runs never execute this step.

1. **Compose the baseline body** from the task file's `## Plan → Approach` paragraph and `## Outcome` section (already in context from Steps 7–9). If `## Outcome` is empty, Steps 7–9 did not complete correctly — exit via the Ad-hoc escape hatch (`rules/do-workflow.md → #### Ad-hoc escape hatch`) rather than proceeding. Structure:
   ```
   ## Summary

   <1–3 sentences derived from ## Plan → Approach and ## Outcome>

   ## Test plan

   - <acceptance criterion 1 from issue ACs>
   - <acceptance criterion 2 from issue ACs>
   ...
   ```
   Ensure the baseline body ends with exactly one trailing newline so the seam with the appended `crafter pr-body` sections renders cleanly.

2. **Invoke the rendering subcommand:**
   ```sh
   crafter pr-body --run-dir .crafter/run/<task-id>/ --task-file {PROJECT_PATH}/{CRAFTER_DIR}/tasks/<task-id>.md
   ```
   This produces the appended sections (`## Manual QA Plan`, `## Known Gaps`, `## Decisions`); empty sections are omitted.

3. **Concatenate** the baseline body and subcommand output to form the full PR body.

4. **Derive the PR title:** use `git log -1 --format='%s'`; fall back to the task file's H1 if that command fails or returns empty.

5. **Open the PR:**
   ```sh
   TITLE=$(git log -1 --format='%s')
   printf '%s' "<full-body>" | gh pr create --title "$TITLE" --body-file -
   ```

**The only push in the `--auto` flow is the one embedded in `gh pr create` — the orchestrator does NOT run `git push` separately at any point. `rules/post-change.md` forbids a standalone `git push`.**

On **failure** of `gh pr create`: record a `Decision (Auto-Recorded): PR creation failed — <error>` entry in the task file's `## Decisions` section; do NOT run the cleanup hook (preserve the run directory for retry/debug); exit via the Ad-hoc escape hatch (`rules/do-workflow.md → #### Ad-hoc escape hatch`).

On **success**: print the PR URL as a one-line notice (`PR opened: <URL>`); run the cleanup hook (`rm -rf .crafter/run/<task-id>/`); proceed to the session wrap-up (Step 7–9 item 5).
