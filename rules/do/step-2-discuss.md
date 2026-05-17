# Step 2 — DISCUSS / RESEARCH (when incomplete or uncertain)

If the completeness check finds gaps, pause and ask targeted clarifying questions. Ask only for missing information — do not re-ask what the user has already stated.

For complex or codebase-dependent uncertainty, delegate to the **`crafter-analyzer`** agent with the user's request, the missing completeness fields, and high-level pointers to relevant areas of the codebase. Do not inject file contents — the Analyzer uses its own Read/Grep/Glob tools to explore the codebase. Present the Analyzer's findings to the user to inform the discussion.

Do not proceed to planning until the request is complete enough to plan. Once it is complete, create the task file per `~/.claude/crafter/rules/task-lifecycle.md` and continue to Step 3.
