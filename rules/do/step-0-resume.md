# Step 0 — Resume Detection

Follow the resume detection procedure in `~/.claude/crafter/rules/task-lifecycle.md`.

**Important:** If the effective request contains resume-intent words (continue, resume, pokracuj, dál, further, next step, carry on, etc.), you must be thorough in searching for active tasks. Use Grep to search for active task metadata lines only (`^- \*\*Status:\*\* active$|^\*\*Status:\*\* active$`) in `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/` before concluding no active task exists.

If resuming an active task, first check the plan status in the task file:
- If the task file contains `**Work branch:** <branch>` and `<branch>` differs from the current branch, do not resume silently. Tell the user the expected branch and ask whether to switch branches, continue anyway, or start fresh.
- If the `## Plan` section still contains `_(pending)_` (no actual steps written yet) — go to Step 1 (scope detection).
- If `**Plan status:** draft` — go to Step 3 to present the plan summary and wait for user approval.
- If `**Plan status:** approved` — the task file's checkboxes are the source of truth. The first unchecked step (`- [ ]`) is the next step to execute — go to Step 4 to execute that plan step.
- Otherwise (Plan section contains unrecognized content) — present the task file to the user and ask how to proceed.

If not resuming, continue to Step 1.

**Branch sanity guard (mandatory):** When starting fresh on a non-main/master branch and no active task match was found, do not assume the current branch is correct just because it is not main/master. Apply the branch/request relevance check from `task-lifecycle.md`. If there is reasonable suspicion that the request does not belong to the current branch, ask the user how to proceed and wait for their instruction before scope detection.

**Main/master guard (mandatory):** When starting fresh on `main` or `master` and no active task match was found, do not plan or create a task file on that branch by default. Derive a suitable topic branch proposal from the request (choose an appropriate conventional prefix like `fix/`, `feature/`, `refactor/`, `docs/`, or `chore/`), present it to the user, and ask whether to create/switch to it. Only continue after the user explicitly accepts the topic branch or explicitly chooses to stay on `main/master` anyway.
