# Step 6a — Session Break (Medium/Large scope only)

**Skip this step for Small scope** — proceed directly to Steps 7–9.

After a step's Execute → Step Drift Check cycle completes and the step is checked off:

1. If this was the **last step in the current phase**, proceed to Step 5a (Phase Verification) and Step 6 (Review).
2. If this was the **last step in the entire plan** and phase verification/review are complete, proceed directly to Steps 7–9.
3. Otherwise, suggest the user run `/clear` and then re-invoke `/crafter-do` to continue with the next step in a fresh context. If the user prefers to continue without clearing, go back to **Step 4 (EXECUTE)** for the next plan step.

The resume detection in Step 0 will pick up the active task file and continue from the next unchecked step or pending phase gate. This keeps each step's Execute → Drift Check cycle in a clean context window.
