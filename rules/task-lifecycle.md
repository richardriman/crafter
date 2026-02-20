# Task Lifecycle

## Task File Naming

- Location: `.planning/tasks/<filename>.md`
- Format: `YYYYMMDD-<topic>.md`
- Topic derivation from branch: take the full git branch name and sanitize it (non-alphanumeric characters become dashes, collapse consecutive dashes, lowercase). No prefix stripping — the full branch name is preserved for traceability.
- On main/master: derive the topic from the user's request (first few meaningful words, slug-ified using the same sanitization rules).
- Examples: branch `feat/add-health-check` → `20260220-feat-add-health-check.md`; branch `RR-do-cool-thing` → `20260220-rr-do-cool-thing.md`

## Resume Detection

Runs at workflow start, before scope detection.

1. Get the current git branch name.
2. Look for any file in `.planning/tasks/` whose Metadata section contains `**Status:** active`.
3. If on a feature branch: match files whose topic part corresponds to the sanitized branch name.
4. If on main/master: show all active task files and let the user choose.
5. If a match is found: read the task file, present its contents to the user, and ask whether to resume or start fresh.
   - If resuming, determine the appropriate workflow step from the task file:
     - Request filled but no Plan → go to scope detection / planning.
     - Plan filled → go to Execute.
     - Execution partially done → go to the next unchecked step.
   - If starting fresh: proceed normally (the old file stays as-is; a new one will be created after scope detection).
6. If no match is found: proceed normally. Task file creation happens after scope detection.

## Task File Creation

Runs after the first user-interaction gate (scope detection in `/crafter:do`, symptom collection in `/crafter:debug`).

1. Create the `.planning/tasks/` directory if it does not exist.
2. Create the task file from the `TASK.md` template with Metadata and Request filled in. Set Status to `active`.

## Task File Updates

Runs at each gate, silently — no user interaction needed.

- **After plan approval:** Write the approved plan to the Plan section. Use checkboxes (`- [ ]`) for each plan step if multi-step.
- **After each step is executed:** Check off the corresponding step (`- [x]`).
- **After fix approval (debug workflow):** Write the proposed fix to the Plan section.
- **After notable review findings:** Append to the Decisions section.

## Task File Completion

Runs during post-change, after commit.

- Fill in the Outcome section with the commit SHA and a brief summary.
- Set Status to `completed`.

## Edge Cases

- If the user abandons the task (says "stop", "cancel", etc.): set Status to `abandoned` with a note in the Outcome section.
- If multiple active files match: use the most recent by date prefix.
