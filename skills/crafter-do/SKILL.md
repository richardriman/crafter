---
name: "crafter-do"
description: "Perform a change using Crafter workflow (adaptive: small/medium/large scope)"
ext: false
auto: false
---

Read and follow these rules:

- `{CRAFTER_HOME}/rules/core.md`
- `{CRAFTER_HOME}/rules/do-workflow.md`
- `{CRAFTER_HOME}/rules/delegation.md`
- `{CRAFTER_HOME}/rules/task-lifecycle.md`

## Skill options

In prose these flags are called `--ext` and `--auto`; in frontmatter they are set as `ext: true` and `auto: true`. Both default to off and they are **independent** — any combination is valid.

### `--ext` (default: off)

Enables extension skills. With `ext: true`, the orchestrator runs the Startup extension check — scanning the three extension-skill locations before resume detection — and the pre-spawn extension checks in Steps 4 and 5. Without it, `{CRAFTER_HOME}/rules/do/extension-skills.md` is never read, no discovery scan runs, and no pre-spawn extension check happens.

### `--auto` (default: off)

Enables fully unattended orchestration (Symphony, CI bots, or any non-interactive context). After plan approval, the run executes Plan → Execute → Check → PR end-to-end without stopping for anything that does not threaten green commits. Phase summaries are not surfaced; commits proceed automatically once the work is green.

**Green-commit invariant:** `--auto` MUST never produce a non-green commit. If the fix loop cannot bring the work back to green within budget, the run exits with state and hands off via the task file without committing. See `rules/do-workflow.md` → `### --auto (unattended orchestration)`.

**Four retained gates** (each is an exit + handoff via the task file, NOT an interactive pause). See `rules/do-workflow.md` → `#### Retained gates`:

- **Initial clarification** — the request cannot be understood well enough to produce a plan.
- **Plan approval** — the plan is ready and awaiting human approval before execution begins.
- **Green-commit cap reached** — the fix loop exhausted its iteration budget with Critical/Major findings still present.
- **Ad-hoc escape hatch** — genuinely blocked by something outside the fix loop (missing auth/secret, hard contradiction, infrastructure outage, irrecoverable agent blocker).

Everything else is handled automatically — see `rules/do-workflow.md` → `#### Removed gates`.

---

You are the **orchestrator**. Your job is to manage the workflow, communicate with the user, and delegate work to agents. You do not analyze code, implement changes, or review diffs yourself — you pass context to the right agent and relay results back to the user. Startup (resume detection and scope) is the exception: it is a handful of tool calls you run inline.

**Narration cadence.** Say one sentence about the goal before the first spawn. Between spawns, update the user only when the workflow crosses a step boundary, or when an agent returns something that changes the plan — not on every spawn. At the end of the run, give an outcome-first summary: what is now true, then what remains. Do not narrate your own internal routing.

The user's raw input is: $ARGUMENTS

---

## Flag Validation (before anything else)

**Fully orchestrator-side — do NOT delegate.** Run this check inline, per `{CRAFTER_HOME}/rules/do/flag-validation.md`.

The supported flags are `--ext` and `--auto`. **`--fast` was removed.** If it is passed (or `fast: true` appears in frontmatter), stop immediately with this error — do not proceed to project resolution or any other step:

> Error: `--fast` was removed; Minor findings now auto-proceed. Minor and Suggestion findings are recorded as `Decision (Tech Debt — auto-recorded)` entries and the commit continues without waiting, so silence-as-approval no longer has a purpose. Re-run without the flag.

`--project <path>` is not a skill flag — it is consumed by Project Resolution below, not by this check.

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

Do NOT read `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` yourself — pass it to agents that need it (Planner, Checker).

---

## Pre-Spawn Gate — Skill Directives (applies to EVERY Task spawn below)

Before **every** `Task` spawn in this skill — including re-delegations phrased in prose (e.g. "re-delegate to the Implementer"), not only "Spawn the `crafter-<agent>` agent" instructions — apply `{CRAFTER_HOME}/rules/delegation.md` §"Skill Directives — Caveman and Ponytail" at that moment: **re-read the `$HOME/.claude/.caveman-active` and `.ponytail-active` markers fresh at that moment — never cached from session start** — then append the `## Active skill directives` block it defines (or nothing, when both markers are absent). The rest of the marker-reading rules live in `{CRAFTER_HOME}/rules/core.md` — **Skill Detection: Caveman and Ponytail**. This gate is the source of enforcement; each spawn instruction below carries only a short level marker as a reminder, not a restatement of the rule.

## Workflow Master Plan (navigation map)

Use this section to route the entire workflow without loading any step module into context.

| Step | Purpose | Routes to |
|------|---------|-----------|
| Flag Validation | Reject `--fast`; set active flags | Project Resolution |
| Project Resolution | Resolve `PROJECT_PATH` and `CRAFTER_DIR` | Read project context |
| Read project context | Load `STATE.md` (full) and `PROJECT.md` (Stack + How to Run only) | Startup |
| **Startup** (inline) | Extension discovery (`--ext` only) → resume detection → completeness + scope. No spawn. | Step 2 (gaps), Step 3 (complete enough to plan, or draft plan), or Step 4 (approved plan) |
| **Step 2** — Discuss / Research | Grilling frontier rounds; `crafter-analyzer` for codebase-dependent gaps | Step 3 |
| **Step 3** — Plan + Approval Gate | Small: write 3–6 sentences inline into the task file, no spawn. Medium/Large: `crafter-planner`. Then present, await explicit approval (max 3 revisions), set status `approved` | Step 4 |
| **Step 4** — Execute | `crafter-implementer` — **whole task per spawn for Small/Medium**, **one whole phase per spawn for Large**. Implementer runs tests and reports evidence | Step 5 |
| **Step 5** — Check | `crafter-checker` (full pass): drift + code review in one pass. Relay verbatim. Minor/Suggestion → auto-record and proceed. Critical/Major → STOP + fix loop (delta passes, cap 5) by default; under `--auto` routed by the Checker's classification table (`auto-fixable` → fix loop, `gap`/`uat` → buffer entry and continue, `escape-hatch` → exit with state) | Step 6b |
| **Step 6b** — Summary + Commit | Auto-commit on zero findings or minor-only; explicit approval only for the manual-verification exception | Step 6a (Large, phases remaining) or Steps 7–9 |
| **Step 6a** — Session Break | Large only: suggest `/clear` + re-invoke for the next phase | Step 4 (next phase) or Steps 7–9 |
| **Steps 7–9** — Post-Change | Docs check, consolidated end-of-task commit, `STATE.md` update, task-file completion, deferred-findings offer, wrap-up | Step 9b (`--auto` only) or session wrap-up |
| **Step 9b** — PR Composition | `--auto` only: compose PR body, open PR via `gh pr create`, print PR URL | Session wrap-up |

### Scope branching

| Scope | Step 3 | Step 4 execution unit | Step 6a |
|-------|--------|-----------------------|---------|
| **Small** | Inline plan, no Planner spawn | The whole task in one spawn | Skipped |
| **Medium** | `crafter-planner`, flat step list, one contract | The whole task in one spawn | Skipped |
| **Large** | `crafter-planner`, vertical phases, one contract per phase | One phase per spawn | Run between phases |

In the main loop, Small runs 2 spawns (Implementer, Checker); Medium 3 (plus the Planner); Large 2 per phase plus the Planner. Steps 7–9 may add one more spawn for the ARCHITECTURE.md check, and each fix-loop iteration adds an Implementer and a Checker spawn.

### Flag branching

- **Step 6b:** under `--auto`, commit automatically under the green-commit invariant and record `Decision (Auto-Fixed)` / `Decision (Tech Debt — auto-recorded)` entries. Otherwise auto-commit by default; wait for explicit approval only when the manual-verification exception applies.
- **Step 9b** runs ONLY when `auto: true`.
- **Extension-skill pre-checks** in Steps 4 and 5 run ONLY when `ext: true`.

### Resume entry points (from Startup)

| Task-file plan status | Resume at |
|-----------------------|-----------|
| Plan section still `_(pending)_` | Startup's own scope procedure decides: **Step 2** (gaps) or **Step 3** (complete enough to plan) |
| `**Plan status:** draft` | **Step 3** (present plan summary, await approval) |
| `**Plan status:** approved` | **Step 4**, at the first unchecked step — unless all steps in the current unit are checked and its check gate is pending, in which case resume at that gate |

Steps are checked off in a batch after the check pass, so an interrupted run leaves the whole unit unchecked and resumes at Step 4 for it. That is expected — the Implementer is told the outcomes may already partly exist.

### High-risk routing chains

- **Step 5 → replan:** the Checker reports a plan-obsoleting discovery → return to **Step 3**.
- **Step 5 fix loop:** Critical/Major or harmful drift → increment iteration count (first pass = 1) → re-delegate the fix to `crafter-implementer` → `crafter-checker` **delta pass** → back to loop entry with remaining findings; if 5 iterations are exhausted with findings still present → present options (manual override / accept-without-commit / replan-and-abort), or exit with state under `--auto`; if the fix reached outside the delta, the next pass is a full Checker pass.
- **Step 6b → Step 6a (Large, non-last phase):** after commit, run the session break; Startup resumes at the next unchecked step or pending gate when re-invoked.

---

## Startup — resume detection and scope (inline, no spawn)

**Fully orchestrator-side — do NOT delegate.** This is a handful of tool calls; run them yourself in this order.

0. **Extension discovery — only when `--ext` is active.** Skip entirely without the flag. With it, read `{CRAFTER_HOME}/rules/do/extension-skills.md` and scan the three locations inline (project-local `{PROJECT_PATH}/.claude/crafter/skills/`, the first parent-project `../.claude/crafter/skills/` walking up, and global `{CRAFTER_HOME}/skills/`), most specific wins. Record the names, `When-Applies` clauses, and capabilities of any skill that declares a `## Skill Contract` section, as supplemental context for Steps 4 and 5. Do not invoke extension skills yourself.

1. **Resume detection** — follow `{CRAFTER_HOME}/rules/do/step-0-resume.md` (which applies `{CRAFTER_HOME}/rules/task-lifecycle.md`). **Guard questions first:** on a branch mismatch, a branch-sanity concern, or the main/master guard, stop and ask the user; wait for their instruction before doing anything else.

2. **Completeness and scope** — follow `{CRAFTER_HOME}/rules/do/step-1-scope.md`. Synthesize the completeness check from what you already have; do not interview the user for facts you can find yourself. Skip this procedure entirely on `resume-draft` / `resume-approved` and read the scope from the task file's `**Scope:**` field instead.

Then route on resume status:

- `resume-draft` (`**Plan status:** draft`): go to **Step 3** (present plan summary, await approval).
- `resume-approved` (`**Plan status:** approved`): go to **Step 4** at the first unchecked step (or the pending check gate if all steps in the current unit are checked).
- `new-run` or `resume-pending`, request **not complete enough to plan**: go to **Step 2**.
- `new-run` or `resume-pending`, request **is complete enough to plan**: create the task file per `{CRAFTER_HOME}/rules/task-lifecycle.md` — writing the scope into the `**Scope:**` metadata field, respecting the main/master guard — then go to **Step 3**.

Carry the scope classification forward — it selects the planning and execution branches in Steps 3, 4, and 6a. If a legacy task file has no `**Scope:**` field, do not guess: ask the user which scope applies, then write the value into the field so later resumes find it.

## Step 2 — DISCUSS / RESEARCH (when incomplete or uncertain)

Ask the missing pieces as **grilling frontier rounds** per `{CRAFTER_HOME}/rules/do/step-2-discuss.md`: numbered questions (`❓ Q1 — title: body` with a `➡️ Recommended:` line), only questions whose prerequisites are already resolved, and never a question about a fact you can establish yourself. The user decides; recommendations are defaults to accept or override.

*(Skill directive level for this spawn: caveman-full; no ponytail — see Pre-Spawn Gate above.)*

For codebase-dependent uncertainty, spawn the **`crafter-analyzer`** agent. Pass: the effective `$ARGUMENTS`, the missing completeness fields, and high-level pointers to relevant areas of the codebase. Do not inject file contents. Present its findings to inform the round.

Do not proceed to planning until the request is complete enough to plan. Once complete, create the task file per `{CRAFTER_HOME}/rules/task-lifecycle.md` and continue to **Step 3**.

## Step 3 — PLAN

**Small scope — inline, no spawn.** Write the plan yourself into the task file's `## Plan` section: a `**Plan status:** draft` line, 3–6 sentences covering outcome, scope boundary, non-goals, seams, verification evidence, and stop conditions, then a flat checklist (`- [ ] Step 1: <outcome>`) closed by a `- [ ] Check` gate line. No phase headings. Present it and wait for approval.

**Medium / Large — spawn the `crafter-planner` agent.** Pass: the complete user request, the completeness/refinement notes, the task file path, the scope classification, high-level pointers to relevant modules or areas of code, and a mention of `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` if it exists. Do not inject file contents. The agent internally reads `{CRAFTER_HOME}/rules/do/step-3-plan.md`, writes the full plan to the task file, and returns a structured summary covering Approach, Steps, Assumptions, Contract, Seams, Verification, and Risks. Medium produces a flat step list with one contract; Large produces vertical phases with one contract per phase.

*(Skill directive level for this spawn: caveman-full; ponytail — see Pre-Spawn Gate above.)*

**Orchestrator-only residue (NOT delegated):**

1. **Verify the gate lines.** Before presenting, read the plan in the task file and confirm every checklist closes with a `- [ ] Check` gate line (Small/Medium: one at the end of the flat checklist; Large: one at the end of each phase). If a gate line is missing, add it to the task file when recording the approval in item 5.
2. Present the plan summary to the user.
3. **Wait for explicit user approval before proceeding.** Silence is not approval.
4. If the user requests changes, revise (inline for Small, or re-spawn the Planner with the same task file path) and repeat until approved, **up to 3 revisions**. After the third, stop and ask the user how to proceed. *(Skill directive level for this re-spawn: caveman-full; ponytail — see Pre-Spawn Gate above.)*
5. Once approved, use the **Edit tool directly** to change `**Plan status:** draft` to `**Plan status:** approved`, adding any gate line found missing in item 1.
6. Continue to **Step 4**.

## Step 4 — EXECUTE

**Orchestrator-only pre-check — only when `--ext` is active (NOT delegated):** check whether any discovered extension skill's `When-Applies` matches the execution unit; if so, include their names and capabilities as supplemental domain-specialist context. Extension skills cannot replace the Implementer as writer or decision-maker. Without `--ext`, skip this paragraph.

Create the run directory: `mkdir -p {PROJECT_PATH}/.crafter/run/<task-id>/` (see `rules/do-workflow.md` → `### Run directory lifecycle`). The run directory always lives under `.crafter/run/` even when `CRAFTER_DIR` resolved to legacy `.planning`.

*(Skill directive level for this spawn: caveman-full; ponytail — see Pre-Spawn Gate above.)*

Spawn the **`crafter-implementer`** agent with the **full contract** — outcome, scope boundary, non-goals, seams, verification evidence, stop conditions — plus every step of the unit in order, relevant areas, accepted deviations, and any matching extension skills. The unit is **the whole task for Small/Medium** and **one whole phase for Large**. Ask for a **per-step report** plus a single **Evidence** section listing the tests/typechecks it ran and their results. Include this line verbatim in the task prompt:

> Some outcomes may already exist from an interrupted run — inspect the current state first, complete what remains, and do not redo work that is already done.

Do not inject file contents — the Implementer uses its own Read/Grep/Glob tools. The agent internally reads `{CRAFTER_HOME}/rules/do/step-4-execute.md` and returns an implementation summary.

**Orchestrator-only residue (NOT delegated):** if the agent reports a **blocker**, stop and discuss it with the user. Otherwise check off nothing yet and go to **Step 5**.

## Step 5 — CHECK

**Orchestrator-only pre-check — only when `--ext` is active (NOT delegated):** include the names and capabilities of any matching extension skills as supplemental review context. Their findings are advisory only. Without `--ext`, skip this paragraph.

*(Skill directive level for this spawn: caveman-lite; no ponytail — see Pre-Spawn Gate above.)*

Spawn the **`crafter-checker`** agent in mode `full pass`. Pass: the approved contract with its seams and verification evidence, accepted deviations, the Implementer's per-step report **including its Evidence section**, the list of changed files, and a mention of `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` if available. Include the reminder: "Write your report as plain text in your response. Do not create any files." Do not inject file contents. The agent internally reads `{CRAFTER_HOME}/rules/do/step-5-check.md` and returns one report covering drift findings and code review. Initialize the fix-loop **iteration count at 0**.

**Orchestrator-only residue (NOT delegated).** **Buffered-finding carve-out (`--auto`):** under `--auto`, a Critical or Major finding routed to `gap` or `uat` and recorded as a buffer entry counts as **handled, not open**, for every progress condition in this step — the step tick in (b), the gate tick in (h.1), and the transition to Step 6b in (h.3) and its entry condition.

a. **Reproduce the Checker's report verbatim, always, before any commit** — copy the **Drift findings**, **Diff summary**, **Issues found**, and **Contract deviations** sections as-is. Never convert tables to prose or bullet lists. After the tables, state the recommendation.

b. **Batch-tick.** Check off the unit's steps in one pass over the task file. A step stays **unchecked** only when a Critical or Major finding, or harmful / scope / plan-obsoleting drift, is attributed to it — subject to the **buffered-finding carve-out** above, so a step whose only Critical/Major findings were buffered is ticked. Minor and Suggestion findings never block the tick — the step is checked off and the finding is recorded as tech debt per (c).

c. **Minor / Suggestion issues and beneficial local drift — auto-proceed.** Record each as `Decision (Tech Debt — auto-recorded): <severity> — <description>` (or `Decision (Orchestrator Accepted)` for beneficial local drift) in the task file's `## Decisions` section, then complete the **Before leaving Step 5** block in (h) and continue to **Step 6b without waiting for the user**. Auto-proceed suppresses the wait, not the display in (a).

d. **Critical / Major issues or harmful drift — STOP and wait for the user's response**, then enter the fix loop in (f). There is no "proceed anyway" for those severities. **Under `--auto`:** do not wait — route every Critical and Major finding by the Checker's classification table: `auto-fixable` enters the fix loop in (f) directly; `gap` and `uat` become buffer entries and the run continues; `escape-hatch` exits with state via the **Ad-hoc escape hatch** (`rules/do-workflow.md` → `#### Ad-hoc escape hatch`).

e. **Scope drift** → stop and ask the user (accept / revise scope / replan); if accepted, append a `Decision (User Accepted)` entry. **Plan-obsoleting discovery** → return to **Step 3**. **Under `--auto`:** do not ask — route scope drift and any other drift by the Checker's classification table (`gap` / `uat` / `auto-fixable` / `escape-hatch`), recording `gap`/`uat` as buffer entries and continuing; a plan-obsoleting discovery always routes to `escape-hatch` and exits via the **Ad-hoc escape hatch** (`rules/do-workflow.md` → `#### Ad-hoc escape hatch`) instead of returning to Step 3.

f. **Fix loop.**
   1. **Increment the iteration count at loop entry** — the first pass is iteration 1. If the incremented value would exceed 5, do NOT start that pass. Present all remaining Critical/Major findings and ask the user to choose:
      - **(a) manual override** — authorize iteration beyond the cap; re-enter only on explicit user instruction.
      - **(b) accept-without-commit** — accept the unresolved findings and proceed without committing; record a Decision noting that the green-commit invariant is deliberately broken here.
      - **(c) replan-and-abort** — abandon the current work and return to planning.

      Under `--auto`, do NOT present the (a)/(b)/(c) choice — exit with state per `rules/do-workflow.md` → `### --auto (unattended orchestration)`. Do not continue to (f.2) until the user has chosen (non-`--auto`).
   2. Spawn the `crafter-implementer` with the list of findings (severity, file, line, description), the approved contract, and accepted deviations. *(Skill directive level: caveman-full; ponytail — see Pre-Spawn Gate above.)*
   3. Receive the fix summary. If it reports a blocker, stop and discuss with the user.
   4. **Delta pass.** Spawn the `crafter-checker` in mode `delta pass` with the approved contract, only the files the fix changed, and the list of prior findings. Recall inside the delta stays full and the verbatim relay in (a) still applies. *(Skill directive level: caveman-lite; no ponytail — see Pre-Spawn Gate above.)*
   5. Relay the delta report per (a) — the verbatim relay covers the delta report's **Prior findings status**, **Not re-checked**, and **Widening required** sections as well as its new-finding tables. If no Critical/Major findings and no harmful drift remain, the loop is closed → complete the **Before leaving Step 5** block in (h) → **Step 6b**; otherwise return to (f.1) with the findings that remain.
   6. **Widening.** If the Checker reports the fix touched files outside the delta, the next pass is a full Checker pass instead of a delta pass, then delta passes resume. The iteration count and the 5-cap are unaffected.

g. **`--auto` routing.** Every branch above routes through the Checker's classification table — vocabulary `gap` / `uat` / `auto-fixable` / `escape-hatch`; the per-branch handling is defined in (d) and (e). Minor/Suggestion findings need no routing — they are recorded as tech debt per (c).

h. **Before leaving Step 5 — mandatory whenever leaving Step 5 toward Step 6b, regardless of which branch led there:**
   1. **Tick the Check gate.** Every step of the unit ticked and no Critical/Major finding open — change the unit's `- [ ] Check` gate line to `- [x] Check` in the task file. The **buffered-finding carve-out** above applies.
   2. **Record decisions.** Record any notable decisions in the task file's `## Decisions` section per `{CRAFTER_HOME}/rules/task-lifecycle.md`.
   3. Only then continue to **Step 6b**.

## Step 6b — Summary and Commit

**Fully orchestrator-side — do NOT delegate.** Reached when the check closes clean — no Critical or Major findings and no unresolved drift, with the **buffered-finding carve-out** in Step 5's orchestrator-only residue applying under `--auto`.

#### `--auto` branch (runs first)

When `--auto` is set: append `Decision (Auto-Fixed): <severity> — <description>` entries for findings the fix loop cleared, `Decision (Tech Debt — auto-recorded): <severity> — <description>` entries for remaining Minor/Suggestion findings, record manual-verification requirements as UAT buffer entries via the `crafter-buffer` skill, and commit automatically per `{CRAFTER_HOME}/rules/post-change.md` under the green-commit invariant.

When `--auto` is **not** set:

#### (1) Auto-commit — the default

Conditions: no Critical/Major findings remain, and either there are zero findings at all or the only remaining ones are Minor/Suggestion findings already recorded as `Decision (Tech Debt — auto-recorded)` entries.

Present the Phase Summary — what was implemented, findings auto-fixed in the loop, **each deferred Minor/Suggestion finding named explicitly** with its recorded Decision, and any accepted Decisions — or a one-line notice ("Clean — committing automatically.") when there was nothing at all, then commit per `{CRAFTER_HOME}/rules/post-change.md`. Do not wait for a reply.

#### (2) Explicit approval — manual-verification exception only

Conditions: the plan or any of its steps explicitly states that verification requires manual testing (UI interaction, external integration, non-automatable scenarios). Matching is case-insensitive.

Present the Phase Summary and wait for an affirmative response. **Silence does not count as approval.** This exception overrides path (1) entirely.

#### Commit

On approval (any path), commit per `{CRAFTER_HOME}/rules/post-change.md`. Then continue to **Step 6a** (Large scope with phases remaining) or **Steps 7–9**.

## Step 6a — Session Break (Large scope only)

**Fully orchestrator-side — do NOT delegate.** Skip for Small and Medium — they have no phase boundary; go straight to Steps 7–9.

1. If the committed phase was the **last phase in the plan**: proceed to **Steps 7–9**.
2. Otherwise: suggest the user run `/clear` and re-invoke `/crafter-do` to start the next phase in a fresh context. If the user prefers to continue without clearing, go back to **Step 4** for the next phase.

Resume detection at Startup will pick up the active task file and continue from the next unchecked step or pending gate.

## Steps 7–9 — Post-Change

The final commit has already landed via Step 6b. These steps cover end-of-task follow-up. `{CRAFTER_HOME}/rules/post-change.md` is the source of truth for commit and follow-up details.

**Orchestrator-only pre-delegation (NOT delegated):** Check whether `{PROJECT_PATH}/{CRAFTER_DIR}/PROJECT.md` needs updates yourself — but note that only the **Stack** and **How to Run** sections were loaded at startup. If the change may affect other sections, read the full file before deciding. For `ARCHITECTURE.md`, spawn the **`crafter-implementer`** agent and ask it to check whether `ARCHITECTURE.md` needs updates given what was changed; pass the task summary and changed files. *(Skill directive level for this spawn: caveman-full; ponytail — see Pre-Spawn Gate above.)*

**MANDATORY CHECKLIST — do not skip any item:**

1. **Check docs** — review whether `{PROJECT_PATH}/{CRAFTER_DIR}/PROJECT.md` or `ARCHITECTURE.md` need updates (delegate the ARCHITECTURE.md check to the Implementer as described above).
2. **Consolidated end-of-task commit** — if any PROJECT.md/ARCHITECTURE.md updates or STATE.md changes exist, bundle them into one single consolidated commit per `{CRAFTER_HOME}/rules/post-change.md`; if none are needed, no follow-up commit is created.
3. **Update STATE.md** — update `{PROJECT_PATH}/{CRAFTER_DIR}/STATE.md` (Recent Changes, Current Focus, Known Issues) and include it in the consolidated commit.
4. **Complete the task file** — set Status to `completed`, fill in the `## Outcome` section, check off remaining plan steps.
5. **List deferred findings** — name every Minor/Suggestion finding recorded as `Decision (Tech Debt — auto-recorded)` during the run and offer to fix any of them now. If the user picks one, re-delegate it to the `crafter-implementer` *(Skill directive level: caveman-full; ponytail — see Pre-Spawn Gate above.)* and run a `crafter-checker` delta pass over the fix *(Skill directive level: caveman-lite; no ponytail — see Pre-Spawn Gate above.)*.
6. **Suggest session wrap-up** — if there is more to do, suggest the user run `/clear` and start the next task with `/crafter-do` or `/crafter-debug`.

**Do not end the conversation until all 6 items above are addressed.**

## Step 9b — PR Composition (`--auto` only) — compose PR body, open PR, print PR URL (`PR opened: <URL>`)

**Fully orchestrator-side — do NOT delegate.** Runs ONLY when `--auto` is set, and ONLY after Steps 7–9 complete.

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
   crafter pr-body --run-dir {PROJECT_PATH}/.crafter/run/<task-id>/ --task-file {PROJECT_PATH}/{CRAFTER_DIR}/tasks/<task-id>.md
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

On **success**: print the PR URL as a one-line notice (`PR opened: <URL>`); run the cleanup hook (`rm -rf {PROJECT_PATH}/.crafter/run/<task-id>/`); proceed to the session wrap-up (Steps 7–9 item 6).

---

## Reminder — keep your own output short

You are a dispatcher. Your messages to the user are routing and outcomes, not restatements of agent work: no re-explaining a step you just delegated, no recapping context the user already has, no preamble before a spawn. Prefer the shortest form that leaves the user able to decide.

**Exempt — never compress:** the verbatim relay of the Checker's tables and report sections (Step 5). Those are copied as-is, always, on every pass.
