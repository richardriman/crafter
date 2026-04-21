# Task Lifecycle

## Task File Naming

- Location: `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/<filename>.md` — where `PROJECT_PATH` and `CRAFTER_DIR` are determined by the orchestrator's project/context resolution step (see `skills/crafter-do/SKILL.md`).
- Format: `YYYYMMDD-<topic>.md`
- Topic derivation from branch: get the git branch name from the project directory (e.g., `git -C {PROJECT_PATH} branch --show-current`) and sanitize it (non-alphanumeric characters become dashes, collapse consecutive dashes, lowercase). No prefix stripping — the full branch name is preserved for traceability.
- On main/master: derive the topic from the user's request (first few meaningful words, slug-ified using the same sanitization rules).
- Examples: branch `feat/add-health-check` → `20260220-feat-add-health-check.md`; branch `RR-do-cool-thing` → `20260220-rr-do-cool-thing.md`
- **Language:** All task file content — topic slug, request description, plan, decisions, outcome — must always be written in English, regardless of the user's conversation language. This reinforces the core rule: "Persistent files (.crafter/*, saved plans): always English."

## Resume Detection

Runs at workflow start, before scope detection.

1. Get the current git branch name from the project directory: `git -C {PROJECT_PATH} branch --show-current`.
2. **Use Grep to search efficiently.** Run a Grep for `**Status:** active` across all files in `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/`. This returns only files with active tasks — do not read every file individually. Then Read only the matched files to determine the task details (request, plan status, checkboxes). Do not skip this step or assume no tasks exist without searching.
3. If the user's request (`$ARGUMENTS`) contains resume-intent words — including but not limited to: "continue", "resume", "pokracuj", "dál", "further", "next step", "carry on" — treat resume detection as **high priority**. If no active tasks are found on the first scan, try reading the directory listing again and check all task files more carefully before concluding there are none. Only after confirming no active tasks exist should you fall through to scope detection.
4. If on a feature branch: match files whose topic part corresponds to the sanitized branch name.
5. If on main/master: show all active task files and let the user choose.
6. If a match is found: read the task file, present its contents to the user, and ask whether to resume or start fresh.
   - If resuming, determine the appropriate workflow step from the task file:
     - Request filled but Plan section still contains `_(pending)_` → go to scope detection / planning.
     - Plan filled with `**Plan status:** draft` → go to plan approval (present plan summary to user and wait for approval).
     - Plan filled with `**Plan status:** approved` → go to Execute (the first unchecked step is next).
     - Otherwise (unrecognized Plan content) → present to user and ask how to proceed.
   - If starting fresh: proceed normally (the old file stays as-is; a new one will be created after scope detection).
7. If no match is found and you are on a feature branch (not main/master): run a branch/request relevance sanity check before proceeding. Compare the effective request (`$ARGUMENTS`) with the branch topic at a high level. If there is a reasonable suspicion that the request is unrelated to the current branch (for example, stale branch context, clearly different task intent, or low topical overlap), do not proceed silently. Ask the user how to continue and wait for a decision.
   - Recommended prompt: "You are on branch `<branch>`, but this request may be unrelated. Should I continue on this branch, or switch/start from another branch first?"
8. If no match is found and either (a) you are on main/master, or (b) the user confirms the current feature branch is correct: proceed normally. Task file creation happens after scope detection.

## Task File Creation

Runs after the first user-interaction gate (scope detection in `/crafter:do`, symptom collection in `/crafter:debug`).

1. Create the `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/` directory if it does not exist.
2. Create the task file from the `TASK.md` template with Metadata and Request filled in. Set Status to `active`.

## Task File Updates

Runs at each gate, silently — no user interaction needed.

- **After planning:** The Planner agent writes the full plan directly to the Plan section (with checkboxes for each step) and sets `**Plan status:** draft`. After the user approves the plan, the orchestrator changes the status to `**Plan status:** approved` (administrative edit via Edit tool). These are the only two valid states for the plan status field.
- **After each step's full cycle (Execute → Verify → Review):** Check off the corresponding step — use a targeted Edit on just the checkbox line (change `- [ ]` to `- [x]`) rather than rewriting the full task file. This avoids pulling the entire file into context each time.
- **After fix approval (debug workflow):** Write the proposed fix to the Plan section.
- **After notable review findings:** Append to the Decisions section.
- **After scope expansion:** If the scope expands or the request is refined during discussion (e.g., additional steps are added to the plan), update the Request section to reflect the final agreed-upon scope before proceeding to execution. The Request should serve as an accurate record of what was actually done, not just the initial input.

## Task File Completion

Runs during post-change, after commit.

- Fill in the Outcome section with the commit SHA and a brief summary.
- Check off any remaining plan steps (`- [ ]` → `- [x]`).
- Set Status to `completed`.

## Edge Cases

- If the user abandons the task (says "stop", "cancel", etc.): set Status to `abandoned` with a note in the Outcome section.
- If multiple active files match: use the most recent by date prefix.
