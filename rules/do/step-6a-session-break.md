# Step 6a — Session Break (Large scope only)

**Skip this step for Small and Medium scope** — they execute the whole task in one unit, so there is no phase boundary to break at; proceed directly to Steps 7–9.

This step is reached only from Step 6b, after the phase has been checked and committed. The phase behind you is always complete — the only question is what comes next:

1. If the committed phase was the **last phase in the plan**, proceed directly to Steps 7–9.
2. Otherwise, suggest the user run `/clear` and then re-invoke `/crafter-do` to start the next phase in a fresh context. If the user prefers to continue without clearing, go back to **Step 4 (EXECUTE)** for the next phase.

Phase boundaries are the only break points, and phases exist only under Large scope.

Resume detection at startup will pick up the active task file and continue from the next unchecked step or pending Check gate. This keeps each phase in a clean context window.
