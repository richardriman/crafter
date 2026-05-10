# Standard Change Workflow Rules

### PLAN
- Always run a lightweight completeness/refinement check before planning.
- If the request is not complete enough to plan, do targeted discussion and/or research before planning.
- A request is complete enough to plan when the goal, scope, non-goals, acceptance criteria, constraints, risks, and validation strategy are clear.
- Always propose a plan before taking implementation action.
- Write plans in plain, conversational language — not XML, not machine-readable syntax.
- Explain **why**, not just what.
- Write plans as vertical execution contracts, not concrete implementation scripts.
- Include step drift criteria and phase verification criteria (how you'll know the change is correct and still inside scope).
- For non-trivial changes, mention alternatives that were considered.
- Surface assumptions and ambiguous interpretations explicitly.
- Include a Karpathy Contract for each phase and step: outcome, scope boundary, non-goals, simplicity constraint, drift criteria, verification evidence, and stop conditions.

### APPROVE
- Never proceed without explicit user approval of the plan.
- If the user has concerns or requests changes, revise the plan and wait again.
- "Looks good" or "go ahead" counts as approval. Silence does not.

### EXECUTE
- Implement exactly what was approved.
- Never change architecture without prior discussion.
- If something unexpected is discovered mid-execution that would materially change the plan, stop and inform the user before continuing.
- Avoid speculative additions ("while we're here" features, abstractions, configurability) unless explicitly approved.
- Execute one step at a time. Do not implement future steps early.

### VERIFY
- After each step, run a lightweight step drift check against that step's Karpathy Contract.
- Step drift checks classify drift as: no drift, harmful drift, scope drift, beneficial local drift, or plan-obsoleting discovery.
- Harmful drift blocks the next step until the current step is fixed.
- Scope drift requires user approval or replanning.
- Beneficial local drift may continue only when recorded as an accepted decision.
- After all steps in a phase pass drift checks, run phase verification against the phase verification criteria.
- **Under `--auto`:** drift handling has three branches; default and `--fast` drift handling are unchanged.
  - **Drift that does NOT threaten green commits** — record as a Decision (Orchestrator Accepted) or Gaps buffer entry (see `skills/crafter-buffer/SKILL.md`) and continue without pausing.
  - **Drift that DOES threaten green commits** — treat as a fix-loop trigger (re-delegate to the Implementer); if the verifier further classifies the drift as plan-obsoleting, route to the Ad-hoc escape hatch exit (see the `### --auto (unattended orchestration)` section).
  - **Verifier "ask user" recommendation** — if non-blocking, downgrade to "record and continue" (Decision or Gaps buffer entry); if blocking, route to the Ad-hoc escape hatch exit.
- Run tests if applicable.
- Report clearly what passed, what failed, and what workflow action is required.
- Verify goals, not just activity: each criterion must map to observable evidence.

### REVIEW
- Run full review after phase verification passes, not after every step.
- Run full review after an individual step only when the step is high-risk: security/auth, data migration, public API, architecture, concurrency, destructive behavior, or a verifier concern.
- **Output format is mandatory** — reproduce the Reviewer's **Diff summary** and **Issues found** tables directly. Copy the markdown tables as-is. **Never** convert tables to prose, bullet lists, or any other format. Expected structure:

  ```
  ### Diff summary
  | File | Changes |
  |------|---------|
  | ...  | ...     |

  ### Issues found
  | # | Severity | File | Line | Description |
  |---|----------|------|------|-------------|
  | ...                                        |

  ### Karpathy scorecard
  | Principle | Status | Evidence |
  |-----------|--------|----------|
  | Think Before Coding | PASS/FLAG | ... |
  | Simplicity First | PASS/FLAG | ... |
  | Surgical Changes | PASS/FLAG | ... |
  | Goal-Driven Execution | PASS/FLAG | ... |
  ```

  After the tables, state the recommendation (must-fix vs. optional).

- **STOP — ALWAYS wait for the user's response before proceeding, regardless of severity. Never auto-proceed when findings exist.** Only if there are literally zero findings: proceed automatically to the next workflow step.

- **After the user responds:** Critical or Major issues trigger the mandatory fix loop — there is no "Proceed anyway" choice for those severities. Only Minor/Suggestion findings → proceed.
- The `crafter-reviewer` agent produces a diff summary and issue report as part of its review output.
- Issues are categorized by severity (Critical, Major, Minor, Suggestion).
- Only Critical and Major issues trigger the fix loop.
- The review-fix loop runs a maximum of 5 iterations — a 6th iteration never starts automatically. If the cap is reached with Critical/Major findings still present, the orchestrator stops and asks the user to choose:
  - **(a) manual override** — authorize manual iteration beyond the cap; the orchestrator re-enters the fix loop only on explicit user instruction.
  - **(b) accept-without-commit** — accept the unresolved findings and proceed without committing this phase; record a Decision entry noting the unresolved findings and that the green-commit invariant is deliberately broken for this phase.
  - **(c) replan-and-abort** — abandon the current phase and return to planning.

  Under `--auto`, the cap-reached state does NOT present the (a)/(b)/(c) choice. Instead, the orchestrator exits with state — the task file remains the handoff artifact, with the unresolved Critical/Major findings recorded as Decisions and the phase left uncommitted. The run terminates without violating the green-commit invariant. See the `### --auto (unattended orchestration)` section below for the full contract.

- Minor issues and Suggestions are informational only.

### --auto (unattended orchestration)

`--auto` enables fully unattended orchestration (Symphony, CI bots, or any non-interactive context). After plan approval, the run executes Plan → Execute → Verify → Review → PR end-to-end without stopping for anything that does not threaten green commits.

**`--auto` is mutually exclusive with `--fast`.** Passing both flags produces a clear parser-level error and the workflow does not start. `--auto` strictly supersedes and replaces `--fast`; it is not an extension of it. The rejection happens at the orchestrator entry point, before any project or task-file work begins.

#### Green-commit invariant

This is a binding rule for all `--auto` runs: **`--auto` MUST never produce a non-green commit.** If the auto-fix loop cannot bring the phase back to green within budget, `--auto` exits with state — the run terminates and the task file is left as the handoff artifact for a human or upstream orchestrator to pick up. It does NOT commit and continue. The four retained gates below are the only legal exit points from an `--auto` run; everything else must be handled automatically.

#### Retained gates

Four conditions cause `--auto` to stop. Each is an **exit + handoff via the task file as state, NOT an interactive pause** — the run terminates, leaving the task file with enough context for the orchestrator or a human to resume.

- **Initial clarification** — the Analyzer cannot understand the ticket well enough to produce a plan. The task file records the blocking question(s) and the run stops before planning.
- **Plan approval** — PLAN.md is ready and awaiting human approval. The run stops after the Analyzer produces the plan; execution does not begin until a human approves.
- **Green-commit cap reached** — the Critical/Major fix loop has exhausted its iteration budget (configured in the REVIEW section above) with Critical or Major findings still present after the final iteration. The unresolved findings are recorded as Decisions in the task file, the phase is left uncommitted, and the run terminates. Note: if the budget is exhausted but all findings were cleared in the final iteration, that is a normal continue path — the cap-reached gate fires only when Critical/Major findings remain after the last allowed iteration.
- **Ad-hoc escape hatch** — the orchestrator is genuinely blocked by something outside the normal fix-loop: missing auth/secret, hard contradiction in inputs, infrastructure outage, or an irrecoverable agent blocker. Full details are in `#### Ad-hoc escape hatch` below.

#### Removed gates

The following conditions are gates in the default or `--fast` flows but are **not blocking under `--auto`**:

- **Manual-verification exception** — recorded into the UAT buffer rather than blocking execution.
- **Critical/Major review findings the auto-fix loop can clear within budget** — the loop fixes them and continues; what was auto-fixed is recorded in Decisions.
- **Minor/Suggestion review findings** — recorded into Decisions as tech debt, same as today's `--fast` behavior.
- **Karpathy scorecard FLAGs** — recorded into Decisions or Gaps buffer depending on nature; execution continues.
- **Step-drift outcomes that do not threaten green commits** — recorded into Gaps buffer or UAT buffer; execution continues.
- **All phase-summary approval gates between phases** — under `--auto`, Phase Summary is not surfaced to the user; the commit proceeds automatically once the phase is green.

#### Ad-hoc escape hatch

The escape hatch is the catch-all exit for situations the other three retained gates do not cover. The other three gates are predictable points (initial clarification at the front, plan approval after planning, cap reached during the review fix-loop). The escape hatch handles genuinely unforeseen blockers that can surface anywhere else in the run.

**Trigger conditions** (illustrative, not exhaustive):

- Missing auth or secret the run cannot proceed without
- Hard contradiction in inputs (e.g., the task file's request and the approved plan diverge irreconcilably mid-execution)
- Infrastructure outage (CI service down, registry unreachable, etc.)
- Irrecoverable agent blocker (Implementer or Verifier hits a state it cannot recover from after exhausting in-step retries)

**Behavior on trigger:** Same exit semantics as the other retained gates — exit with state, the task file remains the handoff artifact, the unresolved blocker is recorded as a Decision (or a structured blocker entry) in the task file's Decisions section, no commit, run terminates without violating the green-commit invariant.

**Agent-side signal criteria:** Each agent has a defined mechanism for emitting a blocker signal; see the respective agent prompt for the authoritative definition. In summary:

- The **Reviewer** (`agents/crafter-reviewer.md` → `## Behavior under --auto`) does NOT emit blocker signals — its three-bucket decision tree (gap → uat → auto-fixable) is exhaustive, with `auto-fixable` as the catch-all default, so every Critical or Major finding is always classified deterministically.
- The **Verifier** (`agents/crafter-verifier.md` → `## Behavior under --auto`) emits a blocker via a "replan" recommendation or an "ask user" recommendation with `escape-hatch` routing, covering missing auth/secret, hard contradiction, infrastructure outage, or irrecoverable state.
- The **Implementer** (`agents/crafter-implementer.md` → `## Behavior under --auto`) emits a blocker when it encounters a genuine impediment that cannot be classified as uat-worthy or gap-worthy.

**Rarity expectation:** In healthy `--auto` runs this exit should be uncommon. Most issues should be either auto-fixable within the fix-loop budget, recordable as Decisions or Gaps buffer, or caught by one of the other three gates. The escape hatch is the safety net, not a routine path.

### Run directory lifecycle (`.crafter/run/<task-id>/`)

Each run that reaches execution gets a dedicated scratch directory: `.crafter/run/<task-id>/`, where `<task-id>` is the task-file basename without extension (e.g., `20260509-feat-gh-16-buffer-skill`). This is the same value the orchestrator already tracks as the active task identifier; no separate resolution step is needed.

**Creation — eager, at Step 4 (Execute) start.** The directory is created by the orchestrator at the beginning of Step 4, before the first agent is spawned for the first plan step. Lazy creation (deferring to the first `crafter-buffer` call) is rejected because `crafter buffer` requires the directory to exist and does not create it (see `skills/crafter-buffer/SKILL.md` → "Creation behavior"). Eager creation at the start of execution (after plan approval, before the first agent is spawned) — rather than during resume detection or planning — avoids creating empty directories for runs that are abandoned before execution. `mkdir -p` semantics are correct: if the directory already exists (resume scenario), the command is a no-op.

**Persistence.** The directory and its buffer files persist for the entire duration of the run. Sub-agents may append to buffer files at any point during execution.

**Per-run identity on resume.** Resuming a task reuses the same `<task-id>` and therefore the same directory. If the directory still exists from a prior session, its buffer files are retained and new entries are appended to them. This is intentional: buffer entries from an earlier session in the same task remain valid context for subsequent sessions.

**Cleanup.** Two triggers perform the same action — full deletion of the run directory and all its contents (`rm -rf .crafter/run/<task-id>/`):

1. **After successful `gh pr create`** (`--auto` runs only) — the primary cleanup trigger. The orchestrator runs the cleanup hook immediately after `gh pr create` succeeds (see `skills/crafter-do/SKILL.md` → Step 9b, "Success handling"). If `gh pr create` fails, cleanup is skipped and the run directory is preserved for retry/debug.
2. **On workspace teardown** — the safety-net trigger, active for all runs. Catches any run that exits via one of the four retained gates (plan rejected, scope-change abort, green-commit violation, escape-hatch blocker) before reaching PR composition, as well as non-`--auto` runs where the user composes the PR manually.

**Git hygiene.** The run directory must never appear in commits. The Crafter repo's own `.gitignore` already includes `.crafter/run/`. Downstream projects MUST add `.crafter/run/` to their `.gitignore`.

**No per-run metadata artifact.** There is no `meta.json` or equivalent run-marker file inside the directory. Buffer entries carry `task_id` and `created_at` fields that provide sufficient per-entry traceability without a separate metadata file. GH#17 confirmed this: the PR composer reads only the two NDJSON buffer files and the task file's `## Decisions` section — no metadata artifact is needed. This remains the policy until a future change demonstrates a concrete need.

## Scope Detection

| Scope | Characteristics | Workflow |
|---|---|---|
| **Small** | 1–3 files, clear intent, isolated change | Completeness check → contract plan → execute step(s) → drift check per step → phase verification → phase review (with fix loop) → commit |
| **Medium** | Multiple files, clear intent, cross-cutting | Completeness check → contract plan with vertical phase(s) → execute one step at a time → drift check per step → phase verification and review per phase |
| **Large** | Incomplete/vague request, architectural impact, many files, or unfamiliar territory | Completeness check → research/discuss until complete → contract plan with vertical phases → execute one step at a time → drift check per step → phase verification and review per phase |

When scope is ambiguous, ask the user rather than guessing. However, if the user has already provided a clear, detailed request, do not ask them to repeat or clarify what they have already stated. Scope ambiguity means you cannot determine whether the change is Small/Medium/Large — it does not mean you need more information about the user's intent.
