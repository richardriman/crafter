# Task Lifecycle

## Task File Naming

- Location: `{PROJECT_PATH}/.planning/tasks/<filename>.md` — where `PROJECT_PATH` is determined by the orchestrator's project resolution step (see `commands/do.md`).
- Format: `YYYYMMDD-<topic>.md`
- Topic derivation from branch: take the full git branch name and sanitize it (non-alphanumeric characters become dashes, collapse consecutive dashes, lowercase). No prefix stripping — the full branch name is preserved for traceability.
- On main/master: derive the topic from the user's request (first few meaningful words, slug-ified using the same sanitization rules).
- Examples: branch `feat/add-health-check` → `20260220-feat-add-health-check.md`; branch `RR-do-cool-thing` → `20260220-rr-do-cool-thing.md`

## Resume Detection

Runs at workflow start, before scope detection.

1. Get the current git branch name.
2. **Use Grep to search efficiently.** Run a Grep for `**Status:** active` across all files in `{PROJECT_PATH}/.planning/tasks/`. This returns only files with active tasks — do not read every file individually. Then Read only the matched files to determine the task details (request, plan status, checkboxes). Do not skip this step or assume no tasks exist without searching.
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
7. If no match is found: proceed normally. Task file creation happens after scope detection.

## Task File Creation

Runs after the first user-interaction gate (scope detection in `/crafter:do`, symptom collection in `/crafter:debug`).

1. Create the `{PROJECT_PATH}/.planning/tasks/` directory if it does not exist.
2. Create the task file from the `TASK.md` template with Metadata and Request filled in. Set Status to `active`.

## Task File Updates

Runs at each gate, silently — no user interaction needed.

- **After planning:** The Planner agent writes the full plan directly to the Plan section (with checkboxes for each step) and sets `**Plan status:** draft`. After the user approves the plan, the orchestrator changes the status to `**Plan status:** approved` (administrative edit via Edit tool). These are the only two valid states for the plan status field.
- **After each step's full cycle (Execute → Verify → Review):** Check off the corresponding step — use a targeted Edit on just the checkbox line (change `- [ ]` to `- [x]`) rather than rewriting the full task file. This avoids pulling the entire file into context each time.
- **After fix approval (debug workflow):** Write the proposed fix to the Plan section.
- **After notable review findings:** Append to the Decisions section.

## Task File Completion

Runs during post-change, after commit.

- Fill in the Outcome section with the commit SHA and a brief summary.
- Set Status to `completed`.

## Edge Cases

- If the user abandons the task (says "stop", "cancel", etc.): set Status to `abandoned` with a note in the Outcome section.
- If multiple active files match: use the most recent by date prefix.
