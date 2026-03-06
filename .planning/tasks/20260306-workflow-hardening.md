# Task: Fix multi-project branch detection, enforce English task files, strengthen post-change checklist

## Metadata
- **Date:** 2026-03-06
- **Branch:** main
- **Status:** completed
- **Scope:** Small

## Request
Three workflow hardening fixes:
1. **Branch detection in multi-project workspaces** — branch name is read from the root repo instead of the subproject's repo when using `--project`. Fix by using `git -C {PROJECT_PATH}` in task-lifecycle.md.
2. **Enforce English in task files** — task file content (topic slug, request, plan, decisions) sometimes gets written in the user's language. Add an explicit English-only rule to task-lifecycle.md to reinforce the existing core.md rule.
3. **Post-change steps skipped** — after completing larger tasks, the orchestrator forgets to complete task files, commit changes, or update STATE.md. Expand the terse "Steps 7–9" in do.md into an inline mandatory checklist.
4. **Request section goes stale** — when scope expands during discussion, the Request section should be updated to reflect the final agreed-upon scope. Add a rule to task-lifecycle.md.

## Plan
**Plan status:** approved

All changes are in a single file: `/Users/ret/dev/ai/crafter/rules/task-lifecycle.md`.

**What and why:** In a multi-project workspace, subdirectories can be independent git repos. Running `git branch --show-current` without specifying a directory defaults to the root repo's branch, which is wrong when `PROJECT_PATH` points to a subproject. The fix adds explicit `git -C {PROJECT_PATH}` instructions at the two places where branch name is obtained.

**Files affected:** `rules/task-lifecycle.md` (lines 5–8 and 15).

- [x] **Step 1 — Update "Topic derivation from branch" (line 7).** Change the phrase "take the full git branch name" to specify that the branch must be obtained from within `PROJECT_PATH`. New wording: `Topic derivation from branch: get the git branch name from the project directory (e.g., \`git -C {PROJECT_PATH} branch --show-current\`) and sanitize it (non-alphanumeric characters become dashes, collapse consecutive dashes, lowercase). No prefix stripping — the full branch name is preserved for traceability.`

- [x] **Step 2 — Update "Resume Detection" step 1 (line 15).** Change `Get the current git branch name.` to: `Get the current git branch name from the project directory: \`git -C {PROJECT_PATH} branch --show-current\`.`

- [x] **Step 3 — Add English-only rule for task files (lines 5–8 area).** Add a note to the "Task File Naming" section in `rules/task-lifecycle.md` stating that all task file content — topic slug, request description, plan, decisions — must always be written in English, regardless of the user's language. This reinforces the existing `core.md` rule ("Persistent files (.planning/*): always English") directly where task files are created.

- [x] **Step 4 — Expand post-change enforcement in `commands/do.md` (lines 170–172).** Replace the current terse "Steps 7–9" section (which just says "Follow the post-change steps in `post-change.md`") with an explicit inline checklist that the orchestrator cannot skip. The new section should keep the reference to `post-change.md` for full details but add a mandatory inline checklist summarizing the key actions that are currently being forgotten. Replace lines 170–172 with content structured like this:

  ```
  ## Steps 7–9 — Post-Change

  Follow the post-change steps in `~/.claude/crafter/rules/post-change.md`.

  **MANDATORY CHECKLIST — do not skip any item:**

  1. **Check docs** — review whether `{PROJECT_PATH}/.planning/PROJECT.md` or `ARCHITECTURE.md` need updates (delegate ARCHITECTURE.md check to Implementer).
  2. **Commit** — ask the user whether to commit. Do not silently skip this. Use conventional commits format.
  3. **Update STATE.md** — after a successful commit, update `{PROJECT_PATH}/.planning/STATE.md` (Recent Changes, Current Focus, Known Issues). Show the user what changed.
  4. **Complete the task file** — set Status to `completed`, fill in the `## Outcome` section, check off remaining plan steps. The task file is in `{PROJECT_PATH}/.planning/tasks/`.
  5. **Suggest session wrap-up** — tell the user they can `/clear` and start fresh.

  **Do not end the conversation until all 5 items above are addressed.**
  ```

  The key insight: the orchestrator reads `do.md` as its primary instruction file. A single-line reference to another file is easy to skip in a long context window. An inline checklist with a bold "MANDATORY" header and an explicit "do not end the conversation" gate makes it much harder to forget.

- [x] **Step 5 — Add Request update rule to `rules/task-lifecycle.md`.** In the "Task File Updates" section, add a new bullet: "**After scope expansion:** If the scope expands or the request is refined during discussion (e.g., additional steps are added to the plan), update the Request section to reflect the final agreed-upon scope before proceeding to execution. The Request should serve as an accurate record of what was actually done, not just the initial input."

**Alternatives considered:** Adding a general note at the top of the file instead of inline changes — rejected because each location should be self-contained so agents don't miss context when reading a specific section. For Step 4 specifically, considered only strengthening the wording in `post-change.md` — rejected because the problem is that the orchestrator never gets to `post-change.md` in the first place; the fix must be where the orchestrator actually reads (in `do.md`).

**Files affected:** `rules/task-lifecycle.md` (lines 5–8 and 15), `commands/do.md` (lines 170–172).

**Verification criteria:**
- Lines 7 and 15 of `rules/task-lifecycle.md` both reference `{PROJECT_PATH}` for branch detection.
- Task File Naming section contains an explicit English-only rule for task file content.
- `commands/do.md` "Steps 7–9" section contains an inline checklist with at least 5 items covering: docs check, commit, STATE.md update, task file completion, and session wrap-up.
- The `commands/do.md` section includes a bold "do not end the conversation" or equivalent enforcement gate.
- The reference to `post-change.md` is preserved (not removed).
- Task File Updates section in `rules/task-lifecycle.md` contains a rule about updating the Request section when scope expands.
- No other files need changes.
- Both files remain valid Markdown with no formatting issues.

**Unknowns / flags:** None — the scope is clear and limited to instruction text.

## Decisions

## Outcome
Commit `5d46339`. All 5 steps implemented: `git -C {PROJECT_PATH}` branch detection, English-only task file rule, mandatory post-change checklist aligned with post-change.md, scope expansion rule for Request section, plus completion step added to task-lifecycle.md.
