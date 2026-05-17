# Step 1 — Completeness and scope

**If the effective request contains a clear, actionable request** (not just resume-intent words), do not ask the user "What do you want to do?" or similar — the user already told you. Instead, run a lightweight completeness check.

A request is complete enough to plan when these are clear: goal, scope, non-goals, acceptance criteria, constraints, risks, and validation strategy. For trivial requests, this can be a one-sentence assessment (e.g., "Completeness check passed because the requested one-line behavior and verification are explicit."). For non-trivial requests, identify missing pieces explicitly.

Based on the project context files, completeness check, and request, classify the scope:

- **Small** — touches 1–3 files, intent is clear, change is isolated
- **Medium** — touches multiple files, intent is clear, change is cross-cutting
- **Large** — incomplete/vague request, architectural impact, many files, or unfamiliar territory

**Extension skill check (supplemental only).** Before finalising the scope classification, check for compatible extension skills discovered at startup (see `{CRAFTER_HOME}/rules/do/extension-skills.md`). If any skill's `When-Applies` matches the request, record their names and capabilities. Pass this list as supplemental context when delegating to the Analyzer in Step 2 or when building plan context in Step 3, so those agents can consult the extension skills as domain specialists. Extension skills may contribute domain-specific completeness criteria; they cannot replace the orchestrator's scope classification or scope-gate decision. See `rules/do-workflow.md` → `### Extension-skill supplemental-only invariant`.

If the request is complete enough to plan, create the task file per `{CRAFTER_HOME}/rules/task-lifecycle.md` and continue to Step 3. Respect the main/master guard first — fresh task files should normally capture the approved topic branch, not `main/master`.
