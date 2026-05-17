---
name: "crafter-do"
description: "Perform a change using Crafter workflow (adaptive: small/medium/large scope)"
fast: false
auto: false
---

Read and follow these rules:
- `~/.claude/crafter/rules/core.md`
- `~/.claude/crafter/rules/do-workflow.md`
- `~/.claude/crafter/rules/delegation.md`
- `~/.claude/crafter/rules/task-lifecycle.md`
- `~/.claude/crafter/rules/do/flag-validation.md`
- `~/.claude/crafter/rules/do/project-resolution.md`
- `~/.claude/crafter/rules/do/extension-skills.md`

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

Apply the flag-validation procedure in `~/.claude/crafter/rules/do/flag-validation.md`.

## Project Resolution (before anything else)

Apply the project-resolution procedure in `~/.claude/crafter/rules/do/project-resolution.md` and set `PROJECT_PATH` and `CRAFTER_DIR` before continuing.

---

Read the project context files for basic orientation (if they exist):
- `{PROJECT_PATH}/{CRAFTER_DIR}/STATE.md` (full file — your primary source of current status)
- `{PROJECT_PATH}/{CRAFTER_DIR}/PROJECT.md` — only the **Stack** and **How to Run** sections

Do NOT read `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` yourself — pass it to agents that need it (Planner, Reviewer).

---

## Extension Skills

Apply the extension-skill discovery and supplemental-only rules in `~/.claude/crafter/rules/do/extension-skills.md`.

---

## Step 0 — Resume Detection

Follow the resume detection procedure in `~/.claude/crafter/rules/task-lifecycle.md`.

**Important:** If the effective request contains resume-intent words (continue, resume, pokracuj, dál, further, next step, carry on, etc.), you must be thorough in searching for active tasks. Use Grep to search for `**Status:** active` in `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/` before concluding no active task exists.

If resuming an active task, first check the plan status in the task file:
- If the task file contains `**Work branch:** <branch>` and `<branch>` differs from the current branch, do not resume silently. Tell the user the expected branch and ask whether to switch branches, continue anyway, or start fresh.
- If the `## Plan` section still contains `_(pending)_` (no actual steps written yet) — go to Step 1 (scope detection).
- If `**Plan status:** draft` — go to Step 3 to present the plan summary and wait for user approval.
- If `**Plan status:** approved` — the task file's checkboxes are the source of truth. The first unchecked step (`- [ ]`) is the next step to execute — go to Step 4 to execute that plan step.
- Otherwise (Plan section contains unrecognized content) — present the task file to the user and ask how to proceed.

If not resuming, continue to Step 1.

**Branch sanity guard (mandatory):** When starting fresh on a non-main/master branch and no active task match was found, do not assume the current branch is correct just because it is not main/master. Apply the branch/request relevance check from `task-lifecycle.md`. If there is reasonable suspicion that the request does not belong to the current branch, ask the user how to proceed and wait for their instruction before scope detection.

**Main/master guard (mandatory):** When starting fresh on `main` or `master` and no active task match was found, do not plan or create a task file on that branch by default. Derive a suitable topic branch proposal from the request (choose an appropriate conventional prefix like `fix/`, `feature/`, `refactor/`, `docs/`, or `chore/`), present it to the user, and ask whether to create/switch to it. Only continue after the user explicitly accepts the topic branch or explicitly chooses to stay on `main/master` anyway.

## Step 1 — Completeness and scope

**If the effective request contains a clear, actionable request** (not just resume-intent words), do not ask the user "What do you want to do?" or similar — the user already told you. Instead, run a lightweight completeness check.

A request is complete enough to plan when these are clear: goal, scope, non-goals, acceptance criteria, constraints, risks, and validation strategy. For trivial requests, this can be a one-sentence assessment (e.g., "Completeness check passed because the requested one-line behavior and verification are explicit."). For non-trivial requests, identify missing pieces explicitly.

Based on the project context files, completeness check, and request, classify the scope:

- **Small** — touches 1–3 files, intent is clear, change is isolated
- **Medium** — touches multiple files, intent is clear, change is cross-cutting
- **Large** — incomplete/vague request, architectural impact, many files, or unfamiliar territory

**Extension skill check (supplemental only).** Before finalising the scope classification, check for compatible extension skills discovered at startup (see `~/.claude/crafter/rules/do/extension-skills.md`). If any skill's `When-Applies` matches the request, record their names and capabilities. Pass this list as supplemental context when delegating to the Analyzer in Step 2 or when building plan context in Step 3, so those agents can consult the extension skills as domain specialists. Extension skills may contribute domain-specific completeness criteria; they cannot replace the orchestrator's scope classification or scope-gate decision. See `rules/do-workflow.md` → `### Extension-skill supplemental-only invariant`.

If the request is complete enough to plan, create the task file per `~/.claude/crafter/rules/task-lifecycle.md` and continue to Step 3. Respect the main/master guard first — fresh task files should normally capture the approved topic branch, not `main/master`.

## Step 2 — DISCUSS / RESEARCH (when incomplete or uncertain)

If the completeness check finds gaps, pause and ask targeted clarifying questions. Ask only for missing information — do not re-ask what the user has already stated.

For complex or codebase-dependent uncertainty, delegate to the **`crafter-analyzer`** agent with the user's request, the missing completeness fields, and high-level pointers to relevant areas of the codebase. Do not inject file contents — the Analyzer uses its own Read/Grep/Glob tools to explore the codebase. Present the Analyzer's findings to the user to inform the discussion.

Do not proceed to planning until the request is complete enough to plan. Once it is complete, create the task file per `~/.claude/crafter/rules/task-lifecycle.md` and continue to Step 3.

## Step 3 — PLAN

Delegate planning to the **`crafter-planner`** agent:

1. Spawn the `crafter-planner` agent.
2. Provide it with: the complete user request, the completeness/refinement notes, the task file path, high-level pointers to relevant modules or areas of code, and a mention of `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` if it exists (the Planner will read it itself). Do not inject file contents — the Planner uses its own Read/Grep/Glob tools to explore the codebase.
3. The Planner writes the full plan directly to the task file and returns a structured summary.
4. Present the Planner's summary to the user. The summary must include:
   - **Approach** — the overall strategy in 1–2 sentences
   - **Phases / steps** — every phase and step, with the outcome and relevant areas
   - **Assumptions** — explicit assumptions or competing interpretations the Planner identified
   - **Karpathy Contract** — scope boundaries, non-goals, drift checks, and stop conditions
   - **Verification criteria** — step drift checks and phase verification criteria
   - **Risks / unknowns** — any flags or open questions from the Planner
   - A note that the full detailed plan is in the task file (mention the path)
5. **Wait for explicit user approval before proceeding.**

If the user requests changes, send the revised request back to the Planner (with the same task file path) and repeat until approved.

Once the user approves, use the Edit tool directly to change `**Plan status:** draft` to `**Plan status:** approved` in the task file's `## Plan` section (this is an administrative update, like checking off completed steps).

If the approved plan contains **phases** (groups of steps under phase headings), execute one step at a time. Phase boundaries determine when phase verification and full review run.

## Step 4 — EXECUTE

**Extension skill check (supplemental only).** Before delegating, check for compatible extension skills discovered at startup (see `~/.claude/crafter/rules/do/extension-skills.md`) whose `When-Applies` matches the current step. If any match, include their names and capabilities in the context provided to the `crafter-implementer` agent so it can consult them as domain specialists during implementation. Extension skills cannot replace the `crafter-implementer` as the writer or decision-maker for any step. See `rules/do-workflow.md` → `### Extension-skill supplemental-only invariant`.

Delegate implementation to the **`crafter-implementer`** agent:

1. Spawn the `crafter-implementer` agent.
2. Provide it with: the current step contract, phase context, relevant areas, non-goals, drift criteria, verification evidence, accepted deviations, and stop conditions. Do not inject file contents — the Implementer uses its own Read/Grep/Glob tools to explore the codebase.
3. Receive the implementation summary from the agent.
4. If the agent reports a blocker, stop and discuss it with the user before continuing.

**All scopes** execute one step at a time. For Small scope this is usually one phase with one or a few steps. After each step, run Step 5 (drift check). After all steps in a phase pass drift checks, run Step 5a (phase verification) and Step 6 (phase review).

## Step 5 — STEP DRIFT CHECK

Delegate verification to the **`crafter-verifier`** agent:

1. Spawn the `crafter-verifier` agent.
2. Provide it with: mode `step drift check`, the current step contract, phase context, non-goals, implementer summary, accepted deviations, changed files, and permission to inspect relevant `git diff` output. The Verifier reads and explores files itself.
3. Remind the Verifier in the task prompt: "Write your verification report as plain text in your response. Do not create any files."
4. Receive the verification report.
5. Present the report to the user clearly.

Handle the Verifier's recommended action:

- **continue:** check off the completed step and continue.
- **record decision and continue:** if the drift is local, beneficial, and does not affect scope or later steps, append a `Decision (Orchestrator Accepted)` entry to the task file and continue.
- **fix current step:** re-delegate the current step to the Implementer before continuing.
- **ask user:** stop and ask the user whether to accept the drift, revise scope, or replan. If accepted, append a `Decision (User Accepted)` entry.
- **replan:** return to Step 3 with the new discovery.

## Step 5a — PHASE VERIFICATION

When all steps in the current phase have passed drift checks, delegate phase verification to the **`crafter-verifier`** agent:

1. Spawn the `crafter-verifier` agent.
2. Provide it with: mode `phase verification`, the approved phase contract, phase verification criteria, accepted deviations, and the list of changed files. The Verifier reads and explores files itself.
3. Remind the Verifier in the task prompt: "Write your verification report as plain text in your response. Do not create any files."
4. Receive and present the verification report.

If phase verification fails, discuss the result with the user and decide whether to re-delegate to the Implementer, adjust the plan, or re-run a specific step drift check.

## Step 6 — REVIEW

**Extension skill check (supplemental only).** Before delegating, check for compatible extension skills discovered at startup (see `~/.claude/crafter/rules/do/extension-skills.md`) whose `When-Applies` matches the current phase. If any match, include their names and capabilities in the context provided to the `crafter-reviewer` agent as supplemental review context, so it can factor in domain-specific review criteria. Extension skill findings are advisory only; they cannot replace the `crafter-reviewer` report or its verdict. See `rules/do-workflow.md` → `### Extension-skill supplemental-only invariant`.

After phase verification passes, delegate code review to the `crafter-reviewer` agent and handle findings. The review-fix iteration count starts at 0. Run review after an individual step only when the step is high-risk: security/auth, data migration, public API, architecture, concurrency, destructive behavior, or a verifier concern.

a. Spawn the `crafter-reviewer` agent.
b. Provide it with: the approved phase contract, accepted deviations, the list of changed files, and a mention of `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` if available. The Reviewer reads files itself.
c. Receive the review report.
d. Present the review results to the user. **Output format is mandatory:**
   - Reproduce the Reviewer's **Diff summary** and **Issues found** tables directly — copy the markdown tables as-is.
   - Reproduce the Reviewer's **Karpathy scorecard** table directly — copy the markdown table as-is.
   - Reproduce the Reviewer's **Contract deviations** section directly.
   - **Never** convert tables to prose, bullet lists, or any other format.
   - After the tables, state the recommendation (must-fix vs. optional).

   **STOP — ALWAYS wait for the user's response before proceeding, regardless of severity. Never auto-proceed when findings exist.**

   Only if there are zero findings at all: proceed directly to Step 6b (auto-approve path).

e. After the user responds:
   - If there are **Critical or Major issues**: on user acknowledgement, enter the fix loop — there is no "Proceed anyway" choice for those severities. Go to sub-step (f).
   - If there are **no Critical or Major issues** (only Minor/Suggestion): proceed to Step 6b.
f. Fix loop for Critical/Major issues:
   1. Check the iteration count. If 5 iterations have already been completed, do not start a 6th. Present all remaining Critical/Major findings to the user and ask them to choose one of:
      Under `--auto`, the orchestrator does not present the `(a)/(b)/(c)` choice — it exits with state per `rules/do-workflow.md` → `### --auto (unattended orchestration)` (the green-commit cap retained gate; the task file remains the handoff artifact).
      - **(a) manual override** — authorize manual iteration beyond the cap; the orchestrator re-enters the fix loop only on explicit user instruction.
      - **(b) accept-without-commit** — accept the unresolved findings and proceed without committing this phase; record a Decision entry noting the unresolved findings and that the green-commit invariant is deliberately broken for this phase.
      - **(c) replan-and-abort** — abandon the current phase and return to planning.
      Do not continue to sub-step (f.2) until the user has chosen.
   2. Spawn the `crafter-implementer` agent. Provide it with: the list of Critical/Major issues from the review (severity, file, line, description), the approved phase contract, and accepted deviations for context. The Implementer reads files itself.
   3. Receive the fix summary. If the Implementer reports a blocker, stop and discuss with the user.
   4. Re-run **Step 5a (PHASE VERIFICATION)** on the newly changed files.
   5. Increment the iteration count, then re-run **Step 6 (REVIEW)** from the top (go back to sub-step (a)).

After review completes, record any notable decisions in the task file per `~/.claude/crafter/rules/task-lifecycle.md`.

## Step 6b — Phase Summary and Auto-Commit

After the review loop closes clean (no Critical or Major findings remain), the orchestrator chooses an approval path based on the active flags.

### Approval paths

#### `--auto` branch (precedes paths 1–3)

When `--auto` is set (`auto: true` in frontmatter): the orchestrator does **not** surface a Phase Summary to the user and does **not** wait for any approval signal. Instead:

1. Record any Critical/Major findings the auto-fix loop cleared before the review closed clean as `Decision (Auto-Fixed): <severity> — <description>` entries in the task file's `## Decisions` section.
2. Record any remaining Minor/Suggestion findings from the final review as tech-debt entries in the task file's `## Decisions` section (format: `Decision (Tech Debt — auto-recorded): <severity> — <description>`).
3. Record any manual-verification requirements as UAT buffer entries via the `crafter-buffer` skill (see `skills/crafter-buffer/SKILL.md`).
4. Commit automatically per `~/.claude/crafter/rules/post-change.md`.

For the canonical four-retained-gates and green-commit-invariant rules that govern this branch — including what constitutes a commit-blocking condition vs. a record-and-continue condition — see `rules/do-workflow.md` → `### --auto (unattended orchestration)`.

When `--auto` is **not** set, fall through to paths (1)–(3) below.

---

**Paths (1)–(3) apply only when `--auto` is not set.**

For non-`--auto` runs, the orchestrator produces and presents a structured **Phase Summary** to the user, then gates the commit on an approval signal. The summary is assembled from context already available to the orchestrator — no new task-file fields or agent invocations are needed:

- **What was implemented** — a brief statement of what the phase delivered (derived from the phase name/description in the task file).
- **Auto-fixed findings** — any Critical or Major findings that were raised during the fix loop and cleared before the review closed clean. List each by severity and short description.
- **Remaining Minor/Suggestion findings** — any open Minor or Suggestion items from the final review report.
- **Accepted Decisions** — any Decision entries recorded during this phase (from Step 5 drift-check handling or the review loop).

If there were no findings of any kind (review was clean on the first pass, fix loop never ran, no Decisions recorded), the summary may be omitted — the orchestrator proceeds directly via the auto-approve path below.

Choose the first path that applies:

#### (1) Auto-approve on clean summary (no user interaction required)

Conditions: zero remaining findings of any severity in the final review state (this covers both the case where the review was clean on the first pass AND the case where the fix loop ran and cleared all Critical/Major with no Minor/Suggestion remaining).

**Exception — manual verification:** If the phase plan or any of its steps explicitly states that verification of this phase requires manual testing (e.g., UI interaction, external integration, non-automatable scenarios), the orchestrator must **wait for explicit user confirmation** even on a fully clean summary. Matching is case-insensitive — "requires", "REQUIRES", "Required", etc., all trigger the wait. Do not introduce a new task-file schema for this flag — treat any plain-text statement that verification requires manual testing as sufficient to trigger the wait; mentions that no manual verification is needed do not trigger it. This exception overrides auto-approve entirely for that phase.

When auto-approve applies: present a one-line notice ("Phase clean — committing automatically.") and proceed directly to the commit per `~/.claude/crafter/rules/post-change.md`.

#### (2) Silence-as-approval — opt-in via `--fast` flag (see Skill options above)

Conditions: the crafter-do skill carries the `--fast` flag (declared in Skill options above) AND remaining Minor/Suggestion findings exist.

Present the Phase Summary and wait for the user's next turn; if that turn does not raise concerns about the summary, treat it as implicit approval and commit. Record each remaining Minor/Suggestion finding as a tech-debt entry in the task file's `## Decisions` section (format: `Decision (Tech Debt — auto-recorded): <severity> — <description>`), then proceed to the commit per `~/.claude/crafter/rules/post-change.md`.

Note: the manual-verification exception in path (1) also applies here — if manual verification is required, `--fast` does not bypass the explicit confirmation wait.

#### (3) Explicit approval — default

Conditions: remaining Minor/Suggestion findings exist AND the `--fast` flag is not set.

Present the Phase Summary and wait for an affirmative response from the user. **Silence does not count as approval.** Do not proceed to the commit until the user explicitly confirms (e.g., "approved", "looks good", "proceed"). If the user raises concerns, resolve them before committing.

### Commit

On approval (any path), run the commit per `~/.claude/crafter/rules/post-change.md`. The commit is automatic — no additional user prompt for the commit command itself.

After committing, continue to **Step 6a** (session break, Medium/Large scope) or **Steps 7–9** (last phase or Small scope).

## Step 6a — Session Break (Medium/Large scope only)

**Skip this step for Small scope** — proceed directly to Steps 7–9.

After a step's Execute → Step Drift Check cycle completes and the step is checked off:

1. If this was the **last step in the current phase**, proceed to Step 5a (Phase Verification) and Step 6 (Review).
2. If this was the **last step in the entire plan** and phase verification/review are complete, proceed directly to Steps 7–9.
3. Otherwise, suggest the user run `/clear` and then re-invoke `/crafter-do` to continue with the next step in a fresh context. If the user prefers to continue without clearing, go back to **Step 4 (EXECUTE)** for the next plan step.

The resume detection in Step 0 will pick up the active task file and continue from the next unchecked step or pending phase gate. This keeps each step's Execute → Drift Check cycle in a clean context window.

## Steps 7–9 — Post-Change

The per-phase commit for the final phase has already landed via Step 6b. Steps 7–9 cover any end-of-task follow-up work. If docs, skillbook, or STATE.md all require no updates, no follow-up commit is created.

Follow the post-change steps in `~/.claude/crafter/rules/post-change.md`. The checklist below is a quick-reference summary — `post-change.md` is the source of truth for details.

**MANDATORY CHECKLIST — do not skip any item:**

1. **Check docs** — review whether `{PROJECT_PATH}/{CRAFTER_DIR}/PROJECT.md` or `ARCHITECTURE.md` need updates (delegate ARCHITECTURE.md check to Implementer). If nothing needs updating, move on silently.
2. **Consolidated end-of-task commit** — if any of the following exist: PROJECT.md/ARCHITECTURE.md updates (item 1), a skillbook entry, or STATE.md changes (item 3), bundle them all into **one single consolidated commit** using conventional commits format. This commit is automatic per `~/.claude/crafter/rules/post-change.md`. Do not create separate commits for docs, skillbook, and STATE.md. If none of those updates are needed, no follow-up commit is created.
3. **Update STATE.md** — update `{PROJECT_PATH}/{CRAFTER_DIR}/STATE.md` (Recent Changes, Current Focus, Known Issues) and include this update in the consolidated commit (item 2). Show the user what changed.
4. **Complete the task file** — set Status to `completed`, fill in the `## Outcome` section, check off remaining plan steps. The task file is in `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/`.
5. **Suggest session wrap-up** — if there's more to do, suggest the user run `/clear` and start their next task with `/crafter-do` or `/crafter-debug` to keep context clean.

**Do not end the conversation until all 5 items above are addressed.**

## Step 9b — PR Composition (`--auto` only)

**Trigger:** Runs ONLY under `--auto` (`auto: true` in frontmatter), and ONLY after Steps 7–9 complete (which may or may not have produced a consolidated end-of-task commit). The latest commit on the work branch will be either the consolidated commit OR the final per-phase commit, depending on whether PROJECT.md / ARCHITECTURE.md / skillbook / STATE.md updates were needed. By the time this step runs, STATE.md is updated and the task file's `## Outcome` section is filled in. Non-`--auto` runs do not execute this step; the user composes the PR manually by invoking `gh pr create` themselves.

This step is the concrete implementation of the `Plan → Execute → Verify → Review → PR end-to-end` promise documented in `rules/do-workflow.md → ### --auto (unattended orchestration)`.

**Inputs:**

- `--run-dir`: `.crafter/run/<task-id>/` — the per-run scratch directory. The orchestrator already tracks `<task-id>` as the task-file basename without extension (e.g., `20260510-feat-gh-17-pr-composer`).
- `--task-file`: `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/<task-id>.md` — the task file path. Already known from the active task context.
- **Baseline body:** a minimal Summary + Test plan block, composed by the orchestrator (LLM-generated from the task file's `## Plan` → Approach paragraph and `## Outcome` section). The Test plan derives from the task file's acceptance criteria ONLY (extracted from the `## Request` section, which quotes the issue ACs, or from the issue body if available) — NOT from phase verification criteria, which would dump dozens of items. If `## Outcome` is empty when this step is reached, that indicates Steps 7–9 did not complete correctly — do NOT proceed to PR composition; exit with state via the Ad-hoc escape hatch.

**Action:**

1. **Compose the baseline body** (Summary + Test plan) from the task file. This is a brief, template-driven LLM composition step — not a new agent delegation. The orchestrator reads the task file's `## Plan → Approach` paragraph and `## Outcome` section (already in context from Steps 7–9) and emits a compact markdown block. Structure rules:
   - No leading newline; ends with exactly one trailing `\n` before concatenation with appended sections (which start with `## Manual QA Plan\n\n…`).
   - `## Summary` and `## Test plan` are H2 headings separated by exactly one blank line (`\n\n`); each heading is followed by one blank line before its body.
   - Test plan items are the issue's acceptance criteria ONLY (see Inputs above).

   ```
   ## Summary\n
   <1–3 sentences derived from ## Plan → Approach and ## Outcome>\n
   \n
   ## Test plan\n
   \n
   - <acceptance criterion 1 from issue ACs>\n
   - <acceptance criterion 2 from issue ACs>\n
   ...\n
   ```

   (The `\n` annotations above make the newline structure explicit; the actual output is plain markdown with those newlines.)

2. **Invoke the rendering subcommand** to produce the appended sections:

   ```sh
   crafter pr-body --run-dir .crafter/run/<task-id>/ --task-file {PROJECT_PATH}/{CRAFTER_DIR}/tasks/<task-id>.md
   ```

   The subcommand outputs the three appended sections (`## Manual QA Plan`, `## Known Gaps`, `## Decisions`) in that fixed order, omitting any section whose source is empty. If all three sources are empty, the output is empty (no appended sections).

3. **Concatenate** the baseline body and the subcommand output to form the full PR body.

4. **Derive the PR title:** use the first line of the most recent commit on the work branch (`git log -1 --format='%s'`); fall back to the task file's H1 heading if that is unavailable or empty.

5. **Open the PR** by passing the full body to `gh pr create`. To avoid shell-interpolation hazards (commit subjects routinely contain `"`, `` ` ``, `$`, `!`), hold the title in a variable and pass the variable — do NOT inline it in a double-quoted string:

   ```sh
   TITLE=$(git log -1 --format='%s')
   printf '%s' "<full-body>" | gh pr create --title "$TITLE" --body-file -
   ```

   Bash does not perform word-splitting or glob-expansion on the right-hand side of a variable assignment, so `TITLE` captures the full commit subject verbatim. Passing `"$TITLE"` to `gh` delivers it as a single argument regardless of embedded quotes or special characters. As an alternative, `gh` ≥ 2.40 supports `--title-file -` for stdin input, which is even safer when a stdin pipe is not already in use.

   Note: `gh pr create` will push the current branch to the remote as part of opening the PR. This is the **only** push in the `--auto` flow — the orchestrator does NOT run `git push` separately at any point. `rules/post-change.md` forbids a standalone `git push`; the push embedded in `gh pr create` is the one-time exception that opens the PR.

**Title derivation rule:** Use the first line of the most recent commit on the work branch (`git log -1 --format='%s'`); fall back to the task file's H1 heading if the git command fails or returns an empty string.

**Failure handling:** If `gh pr create` fails for any reason (network error, authentication failure, branch already has an open PR, branch not pushed, etc.):

1. Record the failure in the task file's `## Decisions` section:
   ```
   Decision (Auto-Recorded): PR creation failed — <error message or summary>
   ```
2. Do NOT run the cleanup hook (the run-dir is preserved for retry/debug).
3. Exit with state via the **Ad-hoc escape hatch** (per `rules/do-workflow.md → #### Ad-hoc escape hatch`). The task file remains the handoff artifact. The run terminates without violating the green-commit invariant (all per-phase and consolidated commits have already landed).

**Success handling:** On `gh pr create` success:

1. Print the PR URL to the user as a one-line notice:
   ```
   PR opened: <URL>
   ```
2. Run the cleanup hook: delete the run directory (`rm -rf .crafter/run/<task-id>/`). This is the PR-success cleanup trigger. If deletion fails, record a tech-debt note and continue — the cleanup failure is non-blocking.
3. Proceed to the session wrap-up (Step 7–9 item 5: suggest `/clear` for the next task).

**Reviewer note:** `## Manual QA Plan` items in the PR body are rendered as GitHub-flavored task list checkboxes (`- [ ] **Title** — verify text`). Reviewers can check them off directly in the PR UI as they complete each manual verification step.
