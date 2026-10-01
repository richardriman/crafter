# Standard Change Workflow Rules

### PLAN
- Always run a lightweight completeness check before planning. Synthesize it from what the user already said — do not interview them for information they have given or that you can find in the codebase yourself.
- If the request is not complete enough to plan, do targeted discussion and/or research before planning.
- A request is complete enough to plan when the goal, scope, non-goals, acceptance criteria, constraints, risks, and validation strategy are clear.
- Always propose a plan before taking implementation action.
- Write plans in plain, conversational language — not XML, not machine-readable syntax. Explain **why**, not just what.
- Write plans as execution contracts, not concrete implementation scripts.
- **Phases are for Large scope only.** Small and Medium produce a flat list of steps under one contract for the whole task. Large groups steps into vertical phases with one contract per phase.
- Each contract defines: **outcome, scope boundary, non-goals, seams, verification evidence, stop conditions**. There is no per-step contract.
- **Seams** are the public interfaces the change is agreed to happen at and is tested against — a function signature, endpoint, CLI flag, exported type, or file boundary. They are agreed up front so the Implementer knows what must not move and the Checker knows where to look.
- For non-trivial changes, mention alternatives that were considered.
- Surface assumptions and ambiguous interpretations explicitly.
- **Small scope is planned inline** by the orchestrator — 3–6 sentences written straight into the task file's `## Plan` section, no planning agent. Medium and Large delegate to `crafter-planner`.

### APPROVE
- Never proceed without explicit user approval of the plan.
- If the user has concerns or requests changes, revise the plan and wait again.
- "Looks good" or "go ahead" counts as approval. Silence does not.

### EXECUTE
- Implement exactly what was approved.
- Never change architecture without prior discussion.
- If something unexpected is discovered mid-execution that would materially change the plan, stop and inform the user before continuing.
- Avoid speculative additions ("while we're here" features, abstractions, configurability) unless explicitly approved.
- Delegate execution in units set by scope: **Small/Medium — the whole task in one Implementer spawn**; **Large — one whole phase per spawn**. There is no per-step delegation. Never implement beyond the contract that was handed over.
- The Implementer runs the tests, typechecks, and lint relevant to its change and reports the commands and their output as evidence. That evidence is an input to the check — it does not replace it.
- On resume, outcomes from an interrupted run may already exist. Inspect the current state first, finish what remains, and do not redo completed work.

### CHECK
- One `crafter-checker` pass per execution unit — per task under Small/Medium, per phase under Large. It covers drift against the contract and code review in a single fresh-context pass. There is no separate verification spawn and no separate review spawn.
- Drift findings are classified as: harmful drift, scope drift, beneficial local drift, or plan-obsoleting discovery. No drift at all is the OK case.
- Review issues carry a severity: Critical, Major, Minor, or Suggestion. The Checker reports **every** finding at every severity — filtering is the orchestrator's job, never the Checker's.
- **The Checker's report is always reproduced in full, verbatim, before any commit.** Copy the Drift findings, Diff summary, Issues found, and Contract deviations sections as markdown tables, as-is. Never convert them to prose or bullet lists.
- **Minor findings stop for a user decision:** fix all, pick numbers to fix, or defer all. Chosen Minors enter the fix loop; each deferred one is recorded as `Decision (Tech Debt — user-deferred): Minor — <description>`. Under `--auto`, Minor findings are recorded as `Decision (Tech Debt — auto-recorded)` without a stop.
- **Suggestion findings do not gate anything.** Record each one in the task file's `## Decisions` section as `Decision (Tech Debt — auto-recorded): Suggestion — <description>` and continue without waiting for the user. Auto-proceed suppresses the wait, never the display.
- **Critical and Major findings, and harmful drift, always STOP.** Present the report, wait for the user's response, then enter the fix loop — there is no "proceed anyway" for those. Minor findings in the same report are decided in the same stop.
- Scope drift requires user approval or replanning. Beneficial local drift may continue only when recorded as a `Decision (Orchestrator Accepted)` entry. A plan-obsoleting discovery returns to PLAN.
- Steps are checked off after the check pass, in one batch. A step stays unchecked only when a Critical or Major finding, or harmful / scope / plan-obsoleting drift, is attributed to it; Minor and Suggestion findings never block the tick.
- The user may ask for a deferred Minor finding to be fixed at any point — including after the commit and at the end of the run. Re-delegate it to the Implementer and run a delta Checker pass.

#### Fix loop

- Entry condition: Critical/Major findings, Minor findings the user chose to fix, or harmful drift. The iteration count is incremented at loop entry — the first pass is iteration 1.
- Each pass: re-delegate to `crafter-implementer` with the findings → spawn `crafter-checker` in **delta pass** mode with the files the fix changed plus the list of prior findings. The Checker reports the status of each prior finding and any new findings in the delta. Recall inside the delta stays full — no high-severity filtering.
- **Widening.** If the fix touched files outside the delta, the next pass runs a full Checker pass instead, then returns to delta passes. The iteration count and the cap are unaffected.
- **Cap: 5 iterations.** A 6th never starts automatically. If the cap is reached with Critical/Major or chosen Minor findings still present, stop and ask the user to choose:
  - **(a) manual override** — authorize iteration beyond the cap; re-enter the loop only on explicit user instruction.
  - **(b) accept-without-commit** — accept the unresolved findings and proceed without committing; record a Decision noting that the green-commit invariant is deliberately broken here.
  - **(c) replan-and-abort** — abandon the current work and return to planning.

  Under `--auto`, the cap-reached state does NOT present the (a)/(b)/(c) choice — the orchestrator exits with state, the task file remains the handoff artifact, unresolved findings are recorded as Decisions, and nothing is committed.

### Extension-skill supplemental-only invariant

**This invariant applies only when `--ext` is active.** Without the flag there is no extension-skill discovery and no pre-spawn extension check, so there is nothing to constrain. When `--ext` is active the invariant is binding for all `crafter-do` runs, including `--auto`.

Extension skills are **supplemental specialists** — they advise, annotate, or enrich workflow phases as domain specialists. They must never:

- **Replace a core agent** — extension skills cannot substitute for the Analyzer, Planner, Implementer, or Checker. Core-agent identities are fixed in `skills/crafter-do/SKILL.md`.
- **Bypass approval, check, or commit gates** — the APPROVE gate, the CHECK pass, and the green-commit invariant all apply regardless of which extension skills are active.
- **Violate the safety envelope** — all behaviors forbidden in `docs/skill-contract.md` → **Safety Envelope** apply to extension skills without exception.

Extension skill findings are advisory only. All workflow decisions remain exclusively with the core orchestrator and core agents.

### --auto (unattended orchestration)

`--auto` enables fully unattended orchestration (Symphony, CI bots, or any non-interactive context). After plan approval, the run executes Plan → Execute → Check → PR end-to-end without stopping for anything that does not threaten green commits. `--auto` and `--ext` are independent flags and may be combined.

#### Green-commit invariant

This is a binding rule for all `--auto` runs: **`--auto` MUST never produce a non-green commit.** If the fix loop cannot bring the work back to green within budget, `--auto` exits with state — the run terminates and the task file is left as the handoff artifact for a human or upstream orchestrator to pick up. It does NOT commit and continue. The four retained gates below are the only legal exit points from an `--auto` run; everything else must be handled automatically.

#### Retained gates

Four conditions cause `--auto` to stop. Each is an **exit + handoff via the task file as state, NOT an interactive pause** — the run terminates, leaving the task file with enough context for the orchestrator or a human to resume.

- **Initial clarification** — the request cannot be understood well enough to produce a plan. The task file records the blocking question(s) and the run stops before planning.
- **Plan approval** — the plan is ready and awaiting human approval. The run stops after planning; execution does not begin until a human approves.
- **Green-commit cap reached** — the fix loop has exhausted its iteration budget with Critical or Major findings still present after the final iteration. The unresolved findings are recorded as Decisions, nothing is committed, and the run terminates. If the budget is exhausted but all findings were cleared in the final iteration, that is a normal continue path — the gate fires only when findings remain.
- **Ad-hoc escape hatch** — the orchestrator is genuinely blocked by something outside the normal fix loop. See `#### Ad-hoc escape hatch` below.

#### Removed gates

The following conditions are gates in the default flow but are **not blocking under `--auto`**:

- **Manual-verification exception** — recorded into the UAT buffer rather than blocking execution.
- **Critical/Major findings the fix loop can clear within budget** — the loop fixes them and continues; what was auto-fixed is recorded in Decisions.
- **Critical/Major review findings routed to `gap`/`uat` by the Checker** — recorded as buffer entries, the run continues. They count as handled, not open, for the step tick, the Check gate tick, and the transition to Step 6b.
- **Minor findings** — recorded into Decisions as `Decision (Tech Debt — auto-recorded)` instead of stopping for a user decision. (Suggestion findings are auto-recorded in all runs.)
- **Drift outcomes that do not threaten green commits** — recorded into the Gaps or UAT buffer; execution continues.
- **All phase-summary approval gates** — under `--auto`, the Phase Summary is not surfaced to the user; the commit proceeds automatically once the work is green.

#### Ad-hoc escape hatch

The escape hatch is the catch-all exit for situations the other three retained gates do not cover. The other three are predictable points (clarification at the front, plan approval after planning, cap reached during the fix loop). The escape hatch handles genuinely unforeseen blockers that can surface anywhere else in the run.

**Trigger conditions** (illustrative, not exhaustive):

- Missing auth or secret the run cannot proceed without
- Hard contradiction in inputs (e.g., the task file's request and the approved plan diverge irreconcilably mid-execution)
- Infrastructure outage (CI service down, registry unreachable, etc.)
- Irrecoverable agent blocker (the Implementer or Checker hits a state it cannot recover from after exhausting in-step retries)

**Behavior on trigger:** same exit semantics as the other retained gates — exit with state, the task file remains the handoff artifact, the unresolved blocker is recorded as a Decision in the task file, no commit, run terminates without violating the green-commit invariant.

**Agent-side signal criteria:** each agent has a defined mechanism for emitting a blocker signal; see the respective agent prompt for the authoritative definition. In summary:

- The **Checker** (`agents/crafter-checker.md` → `## Behavior under --auto`) routes each Critical, Major, or drift finding into one of four buckets — `gap`, `uat`, `escape-hatch`, `auto-fixable` — with `auto-fixable` as the catch-all default. Only `escape-hatch` is a blocker signal; a plan-obsoleting discovery always routes there.
- The **Implementer** (`agents/crafter-implementer.md` → `## Behavior under --auto`) emits a blocker when it encounters a genuine impediment that cannot be classified as uat-worthy or gap-worthy.

**Rarity expectation:** in healthy `--auto` runs this exit should be uncommon. Most issues should be either auto-fixable within the fix-loop budget, recordable as Decisions or buffer entries, or caught by one of the other three gates. The escape hatch is the safety net, not a routine path.

### Run directory lifecycle

Each run that reaches execution gets a dedicated scratch directory: `{PROJECT_PATH}/.crafter/run/<task-id>/`, where `<task-id>` is the task-file basename without extension (e.g., `20260509-feat-gh-16-buffer-skill`). This is the same value the orchestrator already tracks as the active task identifier; no separate resolution step is needed.

**Creation — eager, at Step 4 (Execute) start.** The directory is created by the orchestrator at the beginning of Step 4, before the Implementer is spawned. Lazy creation (deferring to the first `crafter-buffer` call) is rejected because `crafter buffer` requires the directory to exist and does not create it (see `skills/crafter-buffer/SKILL.md` → "Creation behavior"). Eager creation at the start of execution — rather than during resume detection or planning — avoids creating empty directories for runs that are abandoned before execution. `mkdir -p` semantics are correct: if the directory already exists (resume scenario), the command is a no-op.

**Persistence.** The directory and its buffer files persist for the entire duration of the run. Sub-agents may append to buffer files at any point during execution.

**Per-run identity on resume.** Resuming a task reuses the same `<task-id>` and therefore the same directory. If the directory still exists from a prior session, its buffer files are retained and new entries are appended to them. This is intentional: buffer entries from an earlier session in the same task remain valid context for subsequent sessions.

**Cleanup.** Two triggers perform the same action — full deletion of the run directory and all its contents (`rm -rf {PROJECT_PATH}/.crafter/run/<task-id>/`):

1. **After successful `gh pr create`** (`--auto` runs only) — the primary cleanup trigger. The orchestrator runs the cleanup hook immediately after `gh pr create` succeeds (see `skills/crafter-do/SKILL.md` → Step 9b, "Success handling"). If `gh pr create` fails, cleanup is skipped and the run directory is preserved for retry/debug.
2. **On workspace teardown** — the safety-net trigger, active for all runs. Catches any run that exits via one of the four retained gates before reaching PR composition, as well as non-`--auto` runs where the user composes the PR manually.

**Git hygiene.** The run directory must never appear in commits. The Crafter repo's own `.gitignore` already includes `.crafter/run/`. Downstream projects MUST add `.crafter/run/` to their `.gitignore`.

**No per-run metadata artifact.** There is no `meta.json` or equivalent run-marker file inside the directory. Buffer entries carry `task_id` and `created_at` fields that provide sufficient per-entry traceability without a separate metadata file. GH#17 confirmed this: the PR composer reads only the two NDJSON buffer files and the task file's `## Decisions` section — no metadata artifact is needed. This remains the policy until a future change demonstrates a concrete need.

## Scope Detection

| Scope | Characteristics | Workflow |
|---|---|---|
| **Small** | 1–3 files, clear intent, isolated change | Inline completeness + scope → inline plan in the task file → approval → whole task in one Implementer spawn → one Checker pass → commit → post-change |
| **Medium** | Multiple files, clear intent, cross-cutting | Inline completeness + scope → `crafter-planner` (flat step list, one contract) → approval → whole task in one Implementer spawn → one Checker pass → commit → post-change |
| **Large** | Incomplete/vague request, architectural impact, many files, or unfamiliar territory | Inline completeness + scope → discuss/research until complete → `crafter-planner` (vertical phases, one contract per phase) → approval → per phase: one Implementer spawn → one Checker pass → commit → session break |

When scope is ambiguous, ask the user rather than guessing. However, if the user has already provided a clear, detailed request, do not ask them to repeat or clarify what they have already stated. Scope ambiguity means you cannot determine whether the change is Small/Medium/Large — it does not mean you need more information about the user's intent.
