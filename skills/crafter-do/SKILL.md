---
name: "crafter-do"
description: "Perform a change using Crafter workflow (adaptive: small/medium/large scope)"
fast: false
auto: false
---

Read and follow these rules:

<!-- Core rules -->
- `~/.claude/crafter/rules/core.md`
- `~/.claude/crafter/rules/do-workflow.md`
- `~/.claude/crafter/rules/delegation.md`
- `~/.claude/crafter/rules/task-lifecycle.md`

<!-- do/* capability modules -->
- `~/.claude/crafter/rules/do/flag-validation.md`
- `~/.claude/crafter/rules/do/project-resolution.md`
- `~/.claude/crafter/rules/do/extension-skills.md`
- `~/.claude/crafter/rules/do/step-0-resume.md`
- `~/.claude/crafter/rules/do/step-1-scope.md`
- `~/.claude/crafter/rules/do/step-2-discuss.md`
- `~/.claude/crafter/rules/do/step-3-plan.md`
- `~/.claude/crafter/rules/do/step-4-execute.md`
- `~/.claude/crafter/rules/do/step-5-drift.md`
- `~/.claude/crafter/rules/do/step-5a-phase-verification.md`
- `~/.claude/crafter/rules/do/step-6-review.md`
- `~/.claude/crafter/rules/do/step-6b-phase-summary.md`
- `~/.claude/crafter/rules/do/step-6a-session-break.md`
- `~/.claude/crafter/rules/do/step-7-9-post-change.md`
- `~/.claude/crafter/rules/do/step-9b-pr-composition.md`

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

Apply the resume detection procedure in `~/.claude/crafter/rules/do/step-0-resume.md`. This procedure establishes the resume state and applies the branch-sanity and main/master guards before Step 1 begins.

## Step 1 — Completeness and scope

Apply the completeness and scope procedure in `~/.claude/crafter/rules/do/step-1-scope.md`. This procedure runs the lightweight completeness check, classifies scope (Small/Medium/Large), applies the extension-skill supplemental-only check, and routes to Step 2 (if gaps remain) or creates the task file and continues to Step 3 (if the request is complete enough to plan).

## Step 2 — DISCUSS / RESEARCH (when incomplete or uncertain)

Apply the discuss/research procedure in `~/.claude/crafter/rules/do/step-2-discuss.md`. This procedure handles targeted clarifying questions and `crafter-analyzer` delegation for complex or codebase-dependent uncertainty, creates the task file, and continues to Step 3 once the request is complete enough to plan.

## Step 3 — PLAN

Apply the planning procedure in `~/.claude/crafter/rules/do/step-3-plan.md`. This procedure delegates planning to the `crafter-planner` agent, presents the structured summary (approach, phases/steps, assumptions, Karpathy Contract, verification criteria, risks) to the user, and waits for explicit user approval (revision loop back to `crafter-planner` if changes requested). Once the user approves, the orchestrator uses the Edit tool directly to change `**Plan status:** draft` to `**Plan status:** approved` in the task file's `## Plan` section. If the approved plan contains phases, execute one step at a time (Step 4).

## Step 4 — EXECUTE

Apply the execute procedure in `~/.claude/crafter/rules/do/step-4-execute.md`. This procedure performs the extension-skill supplemental-only check before delegating, delegates implementation to the `crafter-implementer` agent one step at a time, and routes to Step 5 (drift check) after each step and to Step 5a (phase verification) and Step 6 (phase review) after all steps in a phase pass.

## Step 5 — STEP DRIFT CHECK

Apply the step-drift-check procedure in `~/.claude/crafter/rules/do/step-5-drift.md`. This procedure delegates step drift verification to the `crafter-verifier` agent in mode `step drift check`, with the plain-text-report reminder ("Write your verification report as plain text in your response. Do not create any files."), and handles the Verifier's recommended action: **continue** — check off the step and continue; **record decision and continue** — the orchestrator appends a `Decision (Orchestrator Accepted)` entry to the task file and continues; **fix current step** — re-delegate the current step to the `crafter-implementer` before continuing; **ask user** — stop and ask the user, and if accepted the orchestrator appends a `Decision (User Accepted)` entry; **replan** — return to Step 3 with the new discovery.

## Step 5a — PHASE VERIFICATION

Apply the phase-verification procedure in `~/.claude/crafter/rules/do/step-5a-phase-verification.md`. This procedure delegates phase verification to the `crafter-verifier` agent in mode `phase verification`, with the plain-text-report reminder ("Write your verification report as plain text in your response. Do not create any files."). If phase verification fails, the orchestrator discusses the result with the user and decides whether to re-delegate to the Implementer, adjust the plan, or re-run a specific step drift check.

## Step 6 — REVIEW

Apply the review procedure in `~/.claude/crafter/rules/do/step-6-review.md`. This procedure performs the supplemental-only extension-skill check before delegating, delegates code review to the `crafter-reviewer` agent, reproduces the Reviewer's Diff summary, Issues found, Karpathy scorecard, and Contract deviations tables verbatim (never converting tables to prose), **STOPs and always waits for the user's response when any findings exist (never auto-proceeds)**, proceeding directly to Step 6b only when there are zero findings; if Critical or Major issues exist, runs the fix loop — checking the iteration count first and, if 5 iterations have already been completed, presenting all remaining Critical/Major findings and asking the user to choose **(a) manual override**, **(b) accept-without-commit**, or **(c) replan-and-abort** (under `--auto`, the orchestrator exits with state instead of presenting this choice), otherwise re-delegating fixes to the `crafter-implementer` agent, re-running Step 5a (PHASE VERIFICATION) on the newly changed files, incrementing the iteration count, and re-running Step 6 (REVIEW) from the top; then records any notable decisions in the task file.

## Step 6b — Phase Summary and Auto-Commit

Apply the phase-summary and auto-commit procedure in `~/.claude/crafter/rules/do/step-6b-phase-summary.md`. This procedure, after the review loop closes clean, chooses an approval path based on the active flags. Under `--auto`: the orchestrator appends `Decision (Auto-Fixed): <severity> — <description>` entries for any Critical/Major findings the fix loop cleared, appends `Decision (Tech Debt — auto-recorded): <severity> — <description>` entries for any remaining Minor/Suggestion findings, records any manual-verification requirements as UAT buffer entries via the `crafter-buffer` skill, and commits automatically under the green-commit invariant (see `rules/do-workflow.md` → `### --auto (unattended orchestration)`). When `--auto` is not set, the orchestrator chooses the first of three approval paths that applies: **(1) auto-approve on a fully clean summary** — if zero findings remain, presents a one-line notice and commits immediately, **except** when the phase plan explicitly requires manual testing, in which case the orchestrator always waits for explicit user confirmation regardless of finding count (the manual-verification exception overrides auto-approve and is not bypassed by `--fast`); **(2) `--fast` silence-as-approval** — if `--fast` is set and Minor/Suggestion findings remain, presents the Phase Summary and treats the user's next turn as implicit approval, recording each remaining finding as a `Decision (Tech Debt — auto-recorded)` entry, then commits (manual-verification exception still applies); **(3) explicit approval** — if neither condition above applies, presents the Phase Summary and waits for an affirmative response; **silence does not count as approval**; does not proceed to the commit until the user explicitly confirms. On approval (any path), runs the commit; after committing, continues to Step 6a (session break, Medium/Large scope) or Steps 7–9 (last phase or Small scope).

## Step 6a — Session Break (Medium/Large scope only)

Apply the session-break procedure in `~/.claude/crafter/rules/do/step-6a-session-break.md`. **Skip this step for Small scope** — proceed directly to Steps 7–9. This procedure applies three routing rules after a step's Execute → Step Drift Check cycle completes and the step is checked off: (1) if this was the last step in the current phase, proceed to Step 5a (Phase Verification) and Step 6 (Review); (2) if this was the last step in the entire plan and phase verification/review are complete, proceed directly to Steps 7–9; (3) otherwise, suggest the user run `/clear` and re-invoke `/crafter-do` to continue with the next step in a fresh context, or if the user prefers to continue without clearing, go back to Step 4 (EXECUTE) for the next plan step. The resume detection in Step 0 will pick up the active task file and continue from the next unchecked step or pending phase gate.

## Steps 7–9 — Post-Change

Apply the post-change procedure in `~/.claude/crafter/rules/do/step-7-9-post-change.md`. This procedure covers end-of-task follow-up work after the final per-phase commit has already landed via Step 6b. It follows the post-change steps defined in `post-change.md` (the source of truth for details).

**MANDATORY CHECKLIST — do not skip any item:**

1. **Check docs** — review whether `{PROJECT_PATH}/{CRAFTER_DIR}/PROJECT.md` or `ARCHITECTURE.md` need updates (delegate ARCHITECTURE.md check to Implementer).
2. **Consolidated end-of-task commit** — if any PROJECT.md/ARCHITECTURE.md updates, a skillbook entry, or STATE.md changes exist, bundle them into one single consolidated commit per `post-change.md`; if none of those updates are needed, no follow-up commit is created.
3. **Update STATE.md** — update `{PROJECT_PATH}/{CRAFTER_DIR}/STATE.md` (Recent Changes, Current Focus, Known Issues) and include this update in the consolidated commit.
4. **Complete the task file** — set Status to `completed`, fill in the `## Outcome` section, check off remaining plan steps (file is in `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/`).
5. **Suggest session wrap-up** — if there is more to do, suggest the user run `/clear` and start the next task with `/crafter-do` or `/crafter-debug`.

**Do not end the conversation until all 5 items above are addressed.**

## Step 9b — PR Composition (`--auto` only)

Apply the PR-composition procedure in `~/.claude/crafter/rules/do/step-9b-pr-composition.md`. This procedure runs ONLY under `--auto` (`auto: true` in frontmatter), and ONLY after Steps 7–9 complete. Non-`--auto` runs do not execute this step.

The procedure: composes a baseline Summary + Test plan body from the task file's `## Plan → Approach` and `## Outcome` sections (if `## Outcome` is empty, Steps 7–9 did not complete correctly — exit via the Ad-hoc escape hatch rather than proceeding); invokes `crafter pr-body --run-dir .crafter/run/<task-id>/ --task-file {PROJECT_PATH}/{CRAFTER_DIR}/tasks/<task-id>.md` to produce the appended sections; concatenates the two parts into the full PR body; derives the PR title from `git log -1 --format='%s'` (falling back to the task-file H1 if that command fails or returns empty); and opens the PR via `printf '%s' "<full-body>" | gh pr create --title "$TITLE" --body-file -`.

**The only push in the `--auto` flow is the one embedded in `gh pr create` — the orchestrator does NOT run `git push` separately at any point. `rules/post-change.md` forbids a standalone `git push`.**

On **failure** of `gh pr create`: record a `Decision (Auto-Recorded): PR creation failed — <error>` entry in the task file's `## Decisions` section; do NOT run the cleanup hook (preserve the run directory for retry/debug); exit via the Ad-hoc escape hatch (`rules/do-workflow.md → #### Ad-hoc escape hatch`).

On **success**: print the PR URL as a one-line notice (`PR opened: <URL>`); run the cleanup hook (`rm -rf .crafter/run/<task-id>/`); proceed to the session wrap-up (Step 7–9 item 5).
