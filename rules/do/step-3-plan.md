# Step 3 — PLAN

Delegate planning to the **`crafter-planner`** agent:

1. Spawn the `crafter-planner` agent.
2. Provide it with: the complete user request, the completeness/refinement notes, the task file path, high-level pointers to relevant modules or areas of code, and a mention of `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` if it exists (the Planner will read it itself). Do not inject file contents — the Planner uses its own Read/Grep/Glob tools to explore the codebase.
3. The Planner writes the full plan directly to the task file and returns a structured summary.
4. Present the Planner's summary to the user. The summary must include:
   - **Approach** — the overall strategy in 1–2 sentences
   - **Phases / steps** — every phase and step, with the outcome and relevant areas
   - **Assumptions** — explicit assumptions or competing interpretations the Planner identified
   - **Karpathy Contract** — scope boundaries, non-goals, drift checks, and stop conditions
   - **Verification criteria** — step drift checks and phase verification criteria
   - **Risks / unknowns** — any flags or open questions from the Planner
   - A note that the full detailed plan is in the task file (mention the path)
5. **Wait for explicit user approval before proceeding.**

If the user requests changes, send the revised request back to the Planner (with the same task file path) and repeat until approved.

Once the user approves, use the Edit tool directly to change `**Plan status:** draft` to `**Plan status:** approved` in the task file's `## Plan` section (this is an administrative update, like checking off completed steps).

If the approved plan contains **phases** (groups of steps under phase headings), execute one step at a time. Phase boundaries determine when phase verification and full review run.
