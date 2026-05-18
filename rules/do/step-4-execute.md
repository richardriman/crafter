# Step 4 — EXECUTE

**Extension skill check (supplemental only).** Before delegating, check for compatible extension skills discovered at startup (see `{CRAFTER_HOME}/rules/do/extension-skills.md`) whose `When-Applies` matches the current step. If any match, include their names and capabilities in the context provided to the `crafter-implementer` agent so it can consult them as domain specialists during implementation. Extension skills cannot replace the `crafter-implementer` as the writer or decision-maker for any step. See `rules/do-workflow.md` → `### Extension-skill supplemental-only invariant`.

Delegate implementation to the **`crafter-implementer`** agent:

1. Spawn the `crafter-implementer` agent.
2. Provide it with: the current step contract, phase context, relevant areas, non-goals, drift criteria, verification evidence, accepted deviations, and stop conditions. Do not inject file contents — the Implementer uses its own Read/Grep/Glob tools to explore the codebase.
3. Receive the implementation summary from the agent.
4. If the agent reports a blocker, stop and discuss it with the user before continuing.

**All scopes** execute one step at a time. For Small scope this is usually one phase with one or a few steps. After each step, run Step 5 (drift check). After all steps in a phase pass drift checks, run Step 5a (phase verification) and Step 6 (phase review).
