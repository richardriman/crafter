# Step 3 — PLAN

The planning route depends on the scope classified at startup.

## Small scope — inline, no Planner spawn

Write the plan yourself, directly into the task file's `## Plan` section: a `**Plan status:** draft` line, 3–6 sentences covering the outcome, the scope boundary, what is out of scope, the seams the change happens at, how it will be verified, and the stop conditions — followed by a flat checklist of steps (`- [ ] Step 1: <outcome>`) closed by a `- [ ] Check` gate line. No phase headings. Then present it to the user and wait for approval.

## Medium / Large — delegate to `crafter-planner`

1. Spawn the `crafter-planner` agent.
2. Provide it with: the complete user request, the completeness/refinement notes, the task file path, the scope classification, high-level pointers to relevant modules or areas of code, and a mention of `{PROJECT_PATH}/{CRAFTER_DIR}/ARCHITECTURE.md` if it exists (the Planner reads it itself). Do not inject file contents.
3. The Planner writes the full plan directly to the task file and returns a structured summary. **Medium:** a flat step list with one contract for the whole task. **Large:** vertical phases with one contract per phase. Never a contract per step.
4. Present the Planner's summary to the user. It must include:
   - **Approach** — the overall strategy in 1–2 sentences
   - **Steps** — every step (grouped by phase under Large), with its outcome and relevant areas
   - **Assumptions** — explicit assumptions or competing interpretations
   - **Contract** — scope boundary, non-goals, and stop conditions
   - **Seams** — the interfaces the change happens at and is tested against
   - **Verification** — the verification evidence the contract requires
   - **Risks / unknowns** — any flags or open questions
   - A note that the full detailed plan is in the task file (mention the path)
5. **Wait for explicit user approval before proceeding.**

If the user requests changes, send the revised request back to the Planner (with the same task file path) and repeat until approved. **Cap: at most 3 planner revisions.** If the plan is still not approved after the third revision, stop re-spawning and ask the user how to proceed — the disagreement is about the request, not the plan.

## Both routes

Before presenting the plan, verify it closes every checklist with a `- [ ] Check` gate line (Small/Medium: one at the end of the flat checklist; Large: one at the end of each phase). Then present the plan and wait for approval.

Once the user approves, use the Edit tool directly to change `**Plan status:** draft` to `**Plan status:** approved` in the task file's `## Plan` section (an administrative update, like checking off completed steps), and in the same pass add any gate line found missing above.

Then continue to Step 4. Small and Medium execute the whole task in one Implementer spawn; Large executes one phase per spawn, and phase boundaries determine when the check pass and the commit run.
