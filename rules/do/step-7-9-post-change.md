# Steps 7–9 — Post-Change

The per-phase commit for the final phase has already landed via Step 6b. Steps 7–9 cover any end-of-task follow-up work. If docs, skillbook, or STATE.md all require no updates, no follow-up commit is created.

Follow the post-change steps in `~/.claude/crafter/rules/post-change.md`. The checklist below is a quick-reference summary — `post-change.md` is the source of truth for details.

**MANDATORY CHECKLIST — do not skip any item:**

1. **Check docs** — review whether `{PROJECT_PATH}/{CRAFTER_DIR}/PROJECT.md` or `ARCHITECTURE.md` need updates (delegate ARCHITECTURE.md check to Implementer). If nothing needs updating, move on silently.
2. **Consolidated end-of-task commit** — if any of the following exist: PROJECT.md/ARCHITECTURE.md updates (item 1), a skillbook entry, or STATE.md changes (item 3), bundle them all into **one single consolidated commit** using conventional commits format. This commit is automatic per `~/.claude/crafter/rules/post-change.md`. Do not create separate commits for docs, skillbook, and STATE.md. If none of those updates are needed, no follow-up commit is created.
3. **Update STATE.md** — update `{PROJECT_PATH}/{CRAFTER_DIR}/STATE.md` (Recent Changes, Current Focus, Known Issues) and include this update in the consolidated commit (item 2). Show the user what changed.
4. **Complete the task file** — set Status to `completed`, fill in the `## Outcome` section, check off remaining plan steps. The task file is in `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/`.
5. **Suggest session wrap-up** — if there's more to do, suggest the user run `/clear` and start their next task with `/crafter-do` or `/crafter-debug` to keep context clean.

**Do not end the conversation until all 5 items above are addressed.**
