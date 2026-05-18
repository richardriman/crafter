# Step 6 — REVIEW

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
