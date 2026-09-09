# Step 5 — CHECK

One `crafter-checker` spawn per execution unit — the whole task under Small/Medium scope, the current phase under Large. It covers drift against the contract and code review in a single fresh-context pass. There is no separate verification spawn and no separate review spawn.

**Extension skill check — only when `--ext` is active.** Without the flag, skip this paragraph entirely. With it, check for compatible extension skills discovered at startup (see `{CRAFTER_HOME}/rules/do/extension-skills.md`) whose `When-Applies` matches the work being checked, and include their names and capabilities in the context passed to the Checker as supplemental review context. Extension skill findings are advisory only and cannot replace the Checker's report. See `rules/do-workflow.md` → `### Extension-skill supplemental-only invariant`.

## Full pass

1. Spawn the `crafter-checker` agent in mode `full pass`.
2. Provide it with: the approved contract (including its seams and verification evidence), accepted deviations, the Implementer's per-step report **including its Evidence section**, the list of changed files, and a mention of `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` if available. The Checker reads and explores files itself — do not inject file contents.
3. Remind the Checker in the task prompt: "Write your report as plain text in your response. Do not create any files."
4. Initialize the fix-loop iteration count at 0.

**Relay the report verbatim.** Reproduce the Checker's **Drift findings**, **Diff summary**, **Issues found**, and **Contract deviations** sections as-is — copy the markdown tables, never convert them to prose or bullet lists. After the tables, state the recommendation. This relay happens on **every** pass, before any commit, regardless of severity.

## Acting on the report

**Buffered-finding carve-out (`--auto`).** Under `--auto`, a Critical or Major finding routed to `gap` or `uat` and recorded as a buffer entry counts as **handled, not open**, for every progress condition in this step — the step tick (sub-step 1), the gate tick (sub-step 7), and the transition to Step 6b (sub-step 9).

1. **Batch-tick.** Check off the unit's steps in one pass over the task file. A step stays **unchecked** only when a Critical or Major finding, or harmful / scope / plan-obsoleting drift, is attributed to it — subject to the **buffered-finding carve-out** above, so a step whose only Critical/Major findings were buffered is ticked. Minor and Suggestion findings never block the tick — the step is checked off and the finding is recorded as tech debt per sub-step 2. The orchestrator remains the only writer of the task file.
2. **Minor and Suggestion issues, and beneficial local drift — auto-proceed.** Record each as a `Decision (Tech Debt — auto-recorded): <severity> — <description>` entry (or `Decision (Orchestrator Accepted)` for beneficial local drift) in the task file's `## Decisions` section, then complete sub-steps 7–9 below and continue **without waiting for the user**. The verbatim relay above already gave the user full visibility; auto-proceed suppresses the wait, not the display.
3. **Critical or Major issues, or harmful drift — STOP.** Present the report and wait for the user's response, then enter the fix loop. There is no "proceed anyway" for those severities. **Under `--auto`:** do not wait — route every Critical and Major finding by the Checker's classification table: `auto-fixable` enters the fix loop directly; `gap` and `uat` become buffer entries and the run continues; `escape-hatch` exits with state via the **Ad-hoc escape hatch** (`rules/do-workflow.md` → `#### Ad-hoc escape hatch`).

   **Reachability send-back — the orchestrator never downgrades a severity.** It has no code context; only the Checker assigns or lowers a severity. This applies to **any new Critical or Major finding at fix-loop entry** — from the first pass or from a delta pass re-entering via sub-step 5 below. A finding that names neither a trigger nor one of the Checker's three reachability states is **sent back to the `crafter-checker`**, spawned in mode `delta pass` with `reachability check on finding #N`. Findings that already name a trigger, or already say "cannot determine — severity kept", go straight to the fix loop. The send-back is a Checker-only pass: no Implementer spawn, no code change, so it does **not** increment the fix-loop iteration count and leaves the 5-iteration cap untouched. At most one send-back per finding, tracked by the finding number in the relayed report — no separate bookkeeping file; whatever the Checker returns the second time is final. It happens **after** the verbatim relay of the report that triggered it, and the report it returns is relayed verbatim too. Identical interactive and under `--auto`.
4. **Scope drift** — stop and ask the user whether to accept the drift, revise scope, or replan. If accepted, append a `Decision (User Accepted)` entry. **Under `--auto`:** do not ask — route the drift by the Checker's classification table (`gap` / `uat` / `auto-fixable` / `escape-hatch`), recording `gap`/`uat` as buffer entries and continuing; `auto-fixable` enters the fix loop and `escape-hatch` exits with state.
5. **Plan-obsoleting discovery** — return to Step 3 with the new discovery. **Under `--auto`:** the Checker always routes it to `escape-hatch` — exit with state via the **Ad-hoc escape hatch** (`rules/do-workflow.md` → `#### Ad-hoc escape hatch`) instead of returning to Step 3.
6. **`--auto` routing.** Every branch above routes through the Checker's classification table — vocabulary `gap` / `uat` / `auto-fixable` / `escape-hatch`; the per-branch handling is defined in sub-steps 3–5. Minor/Suggestion findings need no routing — they are recorded as tech debt per sub-step 2.
7. **Tick the Check gate.** Once the pass closes clean — every step of the unit ticked and no Critical or Major finding open — change the unit's `- [ ] Check` gate line to `- [x] Check` in the task file. The **buffered-finding carve-out** above applies.
8. **Record decisions.** Record any notable decisions in the task file's `## Decisions` section per `{CRAFTER_HOME}/rules/task-lifecycle.md`.
9. **Continue to Step 6b** — only after sub-steps 7 and 8, and only when no Critical or Major findings and no unresolved drift remain. The **buffered-finding carve-out** above applies.

Sub-steps 7–9 are mandatory whenever leaving Step 5 toward Step 6b, regardless of which branch led there.

## Fix loop (Critical/Major and harmful drift)

1. **Increment the iteration count at loop entry** — the first pass is iteration 1. If the incremented value would exceed 5, do NOT start that pass. Present all remaining Critical/Major findings and ask the user to choose one of:
   - **(a) manual override** — authorize iteration beyond the cap; re-enter the loop only on explicit user instruction.
   - **(b) accept-without-commit** — accept the unresolved findings and proceed without committing; record a Decision noting that the green-commit invariant is deliberately broken here.
   - **(c) replan-and-abort** — abandon the current work and return to planning.

   Under `--auto`, do NOT present the (a)/(b)/(c) choice — exit with state per `rules/do-workflow.md` → `### --auto (unattended orchestration)`. Do not continue to sub-step 2 until the user has chosen (non-`--auto`).
2. Spawn the `crafter-implementer` agent with: the list of Critical/Major findings and harmful drift (severity, file, line, description), the approved contract, and accepted deviations. The Implementer reads files itself and re-runs the relevant checks.
3. Receive the fix summary. If the Implementer reports a blocker, stop and discuss with the user.
4. **Delta pass.** Spawn the `crafter-checker` in mode `delta pass`, passing the approved contract, only the files the fix changed, and the list of prior findings. The Checker reports the status of each prior finding and any new findings in the delta; recall inside the delta stays full and the verbatim relay is unchanged.
5. Relay the delta report — the verbatim relay covers the delta report's **Prior findings status**, **Not re-checked**, and **Widening required** sections as well as its new-finding tables — then: if no Critical/Major findings and no harmful drift remain, the loop is closed → run sub-steps 7–9 of **Acting on the report** → Step 6b. Otherwise go back to sub-step 1 with the findings that remain.
6. **Widening.** If the Checker reports that the fix touched files outside the delta, the next pass runs a **full pass** instead of a delta pass, then returns to delta passes. The iteration count and the 5-cap are unaffected.
