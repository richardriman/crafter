# Step 4 — EXECUTE

**Extension skill check — only when `--ext` is active.** Without the flag, skip this paragraph entirely. With it, check the extension skills discovered at startup (see `{CRAFTER_HOME}/rules/do/extension-skills.md`) whose `When-Applies` matches the execution unit, and include their names and capabilities in the context provided to the `crafter-implementer` so it can consult them as domain specialists. Extension skills cannot replace the Implementer as the writer or decision-maker. See `rules/do-workflow.md` → `### Extension-skill supplemental-only invariant`.

Create the run directory (`mkdir -p {PROJECT_PATH}/.crafter/run/<task-id>/`) before the first spawn, per `rules/do-workflow.md` → `### Run directory lifecycle`. The run directory always lives under `.crafter/run/` even when `CRAFTER_DIR` resolved to legacy `.planning`.

Delegate implementation to the **`crafter-implementer`** agent. The delegation unit depends on the scope classified at startup:

- **Small / Medium — the whole task in one spawn.**
- **Large — one whole phase per spawn.**

In both cases:

1. Spawn the `crafter-implementer` agent.
2. Provide it with the **full contract**: outcome, scope boundary, non-goals, seams, verification evidence, stop conditions, plus every step of the unit in order, relevant areas, and accepted deviations.
3. Ask for a **per-step report** (status, files changed, deviations per step) plus a single **Evidence** section listing the tests/typechecks it ran and their results, so the check pass can attribute findings to the step that produced them and check them against real evidence.
4. Include this line verbatim: "Some outcomes may already exist from an interrupted run — inspect the current state first, complete what remains, and do not redo work that is already done."

Do not inject file contents — the Implementer uses its own Read/Grep/Glob tools. Receive the implementation summary; if the agent reports a blocker, stop and discuss it with the user before continuing.

**Routing after execution.** Check off nothing yet — go straight to Step 5 (check), which classifies drift, reviews the code, and tells you which steps to tick.
