# Step 5a — PHASE VERIFICATION

When all steps in the current phase have passed drift checks, delegate phase verification to the **`crafter-verifier`** agent:

1. Spawn the `crafter-verifier` agent.
2. Provide it with: mode `phase verification`, the approved phase contract, phase verification criteria, accepted deviations, and the list of changed files. The Verifier reads and explores files itself.
3. Remind the Verifier in the task prompt: "Write your verification report as plain text in your response. Do not create any files."
4. Receive and present the verification report.

If phase verification fails, discuss the result with the user and decide whether to re-delegate to the Implementer, adjust the plan, or re-run a specific step drift check.
