# Step 5 — STEP DRIFT CHECK

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
