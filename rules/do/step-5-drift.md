# Step 5 — STEP DRIFT CHECK (Large scope only)

**Large scope only.** For Small and Medium scope the Implementer executes the whole phase in one spawn, and per-step drift is classified retroactively by the phase check — see `{CRAFTER_HOME}/rules/do/step-5a-phase-verification.md`. A phase with a single step never runs this step at either scope; it goes straight to the phase check, so the same diff is never verified twice.

Delegate verification to the **`crafter-verifier`** agent:

1. Spawn the `crafter-verifier` agent.
2. Provide it with: mode `step drift check`, the current step contract, phase context, non-goals, implementer summary, accepted deviations, changed files, and permission to inspect relevant `git diff` output. The Verifier reads and explores files itself.
3. Remind the Verifier in the task prompt: "Write your verification report as plain text in your response. Do not create any files."
4. Receive the verification report.
5. Present the report to the user clearly.

Handle the Verifier's recommended action:

- **continue:** check off the completed step and continue.
- **record decision and continue:** if the drift is local, beneficial, and does not affect scope or later steps, append a `Decision (Orchestrator Accepted)` entry to the task file and continue.
- **fix current step:** re-delegate the current step to the Implementer, then spawn the Verifier in mode `targeted re-check`, scoped to that step's contract and the fix diff — not another full step drift check. On pass, tick the step and continue. **Cap: at most 2 re-delegations for the same drift.** If the third check still reports the same drift, stop re-delegating and ask the user how to proceed (accept, revise scope, or replan); under `--auto`, exit via the Ad-hoc escape hatch (`rules/do-workflow.md` → `#### Ad-hoc escape hatch`) instead of asking.
- **ask user:** stop and ask the user whether to accept the drift, revise scope, or replan. If accepted, append a `Decision (User Accepted)` entry.
- **replan:** return to Step 3 with the new discovery.
