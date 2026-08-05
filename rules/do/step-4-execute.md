# Step 4 — EXECUTE

**Extension skill check (supplemental only).** Before delegating, check for compatible extension skills discovered at startup (see `{CRAFTER_HOME}/rules/do/extension-skills.md`) whose `When-Applies` matches the work being delegated — the whole phase for Small/Medium, the current step for Large. If any match, include their names and capabilities in the context provided to the `crafter-implementer` agent so it can consult them as domain specialists during implementation. Extension skills cannot replace the `crafter-implementer` as the writer or decision-maker. See `rules/do-workflow.md` → `### Extension-skill supplemental-only invariant`.

Delegate implementation to the **`crafter-implementer`** agent. The delegation unit depends on the scope classified at startup.

**Small / Medium — the whole phase in one spawn:**

1. Spawn the `crafter-implementer` agent.
2. Provide it with the **full phase contract**: every step of the phase in order, each with its outcome, scope boundary, non-goals, drift criteria, verification evidence and stop conditions — plus phase context, relevant areas, accepted deviations, and matching extension skills.
3. Ask for a **per-step report** (status, files changed, deviations per step) so the phase check can classify drift step by step.
4. Include this line verbatim: "Some outcomes may already exist from an interrupted run — inspect the current state first, complete what remains, and do not redo work that is already done."

**Large — one step per spawn:**

1. Spawn the `crafter-implementer` agent.
2. Provide it with: the current step contract, phase context, relevant areas, non-goals, drift criteria, verification evidence, accepted deviations, and stop conditions.

In both cases: do not inject file contents — the Implementer uses its own Read/Grep/Glob tools to explore the codebase. Receive the implementation summary; if the agent reports a blocker, stop and discuss it with the user before continuing.

**Routing after execution.** Small/Medium: check off nothing yet — go straight to Step 5a (phase check), which classifies per-step drift and tells you which steps to tick. Large: run Step 5 (drift check) after each step and tick that step, then Step 5a after the last step of the phase; a single-step phase skips Step 5. Then Step 6 (phase review).
