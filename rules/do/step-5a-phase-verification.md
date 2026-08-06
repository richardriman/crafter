# Step 5a — PHASE CHECK

The phase check is the single verification pass for a phase: it covers per-step drift and the **phase verification criteria** in one Verifier spawn over one diff. How much per-step work it does depends on what already ran — steps that passed their own Step 5 drift check are not re-classified.

It runs for every scope. Under Small/Medium it is the only drift check there is. Under Large it follows the per-step drift checks and covers the phase criteria plus any drift those checks could not see across steps. A phase with a single step runs only the phase check — never a separate step drift check over the same diff.

Delegate to the **`crafter-verifier`** agent:

1. Spawn the `crafter-verifier` agent.
2. Provide it with: mode `phase check`, the approved phase contract **including every step contract in it**, the phase verification criteria, accepted deviations, the Implementer's per-step report, the list of changed files, and — critically — **which steps already passed a Step 5 drift check**. On the Large path, name them; on the Small/Medium path, state that none were checked individually. Omitting this makes the Verifier fall back to a full per-step re-classification and the de-redundancy is lost. The Verifier reads and explores files itself.
3. Remind the Verifier in the task prompt: "Write your verification report as plain text in your response. Do not create any files."
4. Receive and present the report — copy the per-step drift table as-is.

Then act on it:

1. **Batch-tick.** Check off every step whose recommended action is `continue`, in one pass over the task file. Leave the rest unchecked. A step reported as `already checked` carries `continue` and counts as clean — under Large it was ticked after its own Step 5 drift check, so there is nothing left to tick. The orchestrator remains the only writer of the task file.
2. **Handle each remaining step** by its recommended action, using the same rules as Step 5: `record decision and continue` → append a `Decision (Orchestrator Accepted)` entry and tick the step; `ask user` → stop and ask; `replan` → return to Step 3; `fix current step` → run the fix-and-re-verify cycle:
   - Re-delegate that step to the Implementer with the step contract and the drift the Verifier reported.
   - Spawn the Verifier in mode `targeted re-check`, scoped to that step's contract and the fix diff — not another full phase check.
   - On pass, tick the step. If the same drift is still reported, repeat: **at most 2 re-delegations per step** (the same cap as Step 5), then stop and ask the user (accept / revise scope / replan), or exit via the Ad-hoc escape hatch under `--auto`.
3. **Failed phase criteria:** discuss the result with the user and decide whether to re-delegate to the Implementer, adjust the plan, or accept.
4. Under `--auto`, route each drift item by the Verifier's `Auto-routing` line **per item** — the routing vocabulary (`gap` / `uat` / `no-buffer` / `escape-hatch`) is unchanged.

Proceed to Step 6 only when every step of the phase is checked off and every phase criterion passes.

**Inside the review fix loop:** every pass spawns the Verifier in mode `targeted re-check` instead of this full phase check, naming only the files the fix changed and the criteria/steps they could affect. This full check re-runs only when a fix reached outside that delta.
