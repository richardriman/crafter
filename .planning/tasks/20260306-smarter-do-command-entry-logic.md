# Task: Smarter /crafter:do entry logic — resume detection, monorepo support, input handling

## Metadata
- **Date:** 2026-03-06
- **Branch:** main
- **Status:** completed
- **Scope:** Large

## Request
The initial decision logic in /crafter:do has several issues:
1. When the user provides clear instructions with the request, the orchestrator sometimes ignores them and asks what to do anyway.
2. Resume detection fails to find active tasks — claims no active task exists even when there's an in-progress task with unchecked steps.
3. In a monorepo/workspace setup (e.g. /workspace/rust, /workspace/elixir — each with their own .planning/), there's no reliable way to tell Crafter which subproject to work on. Saying "v rust" in natural language is unreliable — the orchestrator doesn't understand it refers to a subdirectory.

User works primarily on main branch. Workspace structure: subdirectories like rust/, elixir/ each with their own .planning/.

## Plan
**Plan status:** approved

This task improves three aspects of the `/crafter:do` entry logic: monorepo/workspace support, more reliable resume detection, and not ignoring clear user input. The changes touch three files: `commands/do.md`, `rules/task-lifecycle.md`, and `rules/do-workflow.md`. All changes are additive — single-project setups see no behavioral difference.

### Stage 1 — Monorepo support and project resolution

The core idea: add a "Step -1" to `commands/do.md` that resolves the project root before anything else. This resolution determines the base path for all `.planning/` references throughout the workflow. The task lifecycle rules also need updating so `.planning/tasks/` paths are relative to the resolved project root.

- [x] **Step 1: Add project resolution logic to `commands/do.md`** (file: `commands/do.md`, lines 14–20)

  Insert a new section **before** Step 0 (between the "Read the project context files" block and the `$ARGUMENTS` line). This section:

  1. **Parse `--project <path>` from `$ARGUMENTS`.** If `$ARGUMENTS` contains `--project <path>` (e.g., `--project rust fix the parser`), extract the path as `PROJECT_PATH` and strip `--project <path>` from the remaining arguments. The `--project` flag can appear anywhere in the arguments but conventionally comes first. The path is a relative directory path (e.g., `rust`, `rust/`, `packages/frontend`).
  2. **If no `--project` specified**, check if `.planning/` exists at CWD.
     - If yes: `PROJECT_PATH` = `.` (current directory). Done.
     - If no: scan one level deep for directories containing `.planning/` (i.e., check `*/.planning/`).
       - If exactly one found: use it as `PROJECT_PATH`, inform the user (e.g., "Found project in `rust/`, using it. Tip: next time you can use `--project rust` to skip this detection.").
       - If multiple found: list them and ask the user which one to use. In the prompt, mention the `--project` shortcut so users discover it naturally (e.g., "Found projects: `rust/`, `elixir/`. Which one would you like to work on? (Tip: you can skip this next time with `/crafter:do --project rust ...`)"). Wait for response.
       - If none found: proceed with `.` as `PROJECT_PATH` (the normal single-project path — `.planning/` may be created later by the workflow).
  3. **Apply `PROJECT_PATH`** to all subsequent `.planning/` references. Rewrite the "Read the project context files" instructions (currently lines 14–18) so they reference `{PROJECT_PATH}/.planning/STATE.md` and `{PROJECT_PATH}/.planning/PROJECT.md` instead of hardcoded `.planning/`. Add a note that the orchestrator must use `PROJECT_PATH` as the base for all `.planning/` paths throughout the entire workflow (task files, ARCHITECTURE.md references passed to agents, etc.).

  Also update the `$ARGUMENTS` line (line 20) to reflect that arguments may have been modified by `--project` extraction (the remaining arguments after stripping `--project <path>`).

- [x] **Step 2: Update task lifecycle rules for project-relative paths** (file: `rules/task-lifecycle.md`, lines 1–8 and 12–26)

  Update the Task Lifecycle rules so all `.planning/tasks/` references are described as relative to the project root (which the orchestrator determines). Specifically:

  - In "Task File Naming" (line 5): change `.planning/tasks/<filename>.md` to `{PROJECT_PATH}/.planning/tasks/<filename>.md` and add a note that `PROJECT_PATH` is determined by the orchestrator's project resolution step.
  - In "Resume Detection" (line 16): change `.planning/tasks/` to `{PROJECT_PATH}/.planning/tasks/`.
  - In "Task File Creation" (line 31): same treatment.

  This is a documentation/instruction change only — the orchestrator interprets these rules and applies the project path.

### Stage 2 — Stronger resume detection

The current resume detection instructions are too terse — the orchestrator sometimes skips them or fails to actually read task files. This stage makes the instructions more explicit and adds resume-intent word detection.

- [x] **Step 3: Strengthen resume detection in `rules/task-lifecycle.md`** (file: `rules/task-lifecycle.md`, lines 12–26)

  Rewrite the "Resume Detection" section to be more explicit and procedural:

  1. Keep step 1 (get branch name).
  2. Rewrite step 2 to be emphatic and efficient: "**Use Grep to search efficiently.** Run a Grep for `**Status:** active` across all files in `{PROJECT_PATH}/.planning/tasks/`. This returns only files with active tasks — do not read every file individually. Then Read only the matched files to determine the task details (request, plan status, checkboxes). Do not skip this step or assume no tasks exist without searching."
  3. Keep steps 3-4 (branch matching vs. main branch behavior).
  4. Add a new step between current 4 and 5: "If the user's request (`$ARGUMENTS`) contains resume-intent words — including but not limited to: 'continue', 'resume', 'pokracuj', 'dál', 'further', 'next step', 'carry on' — treat resume detection as **high priority**. If no active tasks are found on the first scan, try reading the directory listing again and check all task files more carefully before concluding there are none. Only after confirming no active tasks exist should you fall through to scope detection."
  5. Keep steps 5-6 (match found / no match behavior).

- [x] **Step 4: Add resume-intent awareness to `commands/do.md` Step 0** (file: `commands/do.md`, lines 24–34)

  Add a note to Step 0 in `commands/do.md` that reinforces the resume-intent logic:

  After "Follow the resume detection procedure in `~/.claude/crafter/rules/task-lifecycle.md`.", add:

  "**Important:** If `$ARGUMENTS` contains resume-intent words (continue, resume, pokracuj, dál, next step, carry on, etc.), you must be thorough in searching for active tasks. Actually read all files in `{PROJECT_PATH}/.planning/tasks/` before concluding no active task exists."

### Stage 3 — Respect clear user input

- [x] **Step 5: Prevent orchestrator from ignoring clear input** (file: `commands/do.md`, lines 36–47, and `rules/do-workflow.md`, lines 52–59)

  Two changes:

  **In `commands/do.md`, Step 1** (lines 36–47): Add a preamble before scope detection:

  "**If `$ARGUMENTS` contains a clear, actionable request** (not just resume-intent words), proceed directly to scope detection. Do NOT ask the user 'What do you want to do?' or similar — the user already told you via `$ARGUMENTS`. Only ask clarifying questions if `$ARGUMENTS` is empty or genuinely vague/ambiguous."

  **In `commands/do.md`, Step 2** (lines 48–56): Add a qualifier to the clarifying questions block:

  "Only ask these questions if the request is genuinely vague or ambiguous — meaning you cannot determine what files or areas of code are involved. If the user has provided specific details in their request (e.g., 'add retry logic to the HTTP client', 'fix the parser for nested expressions'), those details are sufficient to proceed to planning. Do not re-ask what the user has already stated."

  **In `rules/do-workflow.md`** (line 59, after "When scope is ambiguous, ask the user rather than guessing."): Add: "However, if the user has already provided a clear, detailed request, do not ask them to repeat or clarify what they have already stated. Scope ambiguity means you cannot determine whether the change is Small/Medium/Large — it does not mean you need more information about the user's intent."

### Alternatives considered

- **Prefix parsing (`rust: fix X`, `in rust/: continue`)** — rejected as a "hidden feature" that users would never discover. The explicit `--project` flag is standard CLI convention and is self-documenting. The interactive fallback naturally teaches users about `--project` via the tip in the prompt.
- **Separate `/crafter:do-in` command for monorepo** — rejected because it fragments the entry point. A single command with an optional `--project` flag is more ergonomic and discoverable.
- **Automatic CWD change to subproject** — rejected because changing CWD has side effects on git operations and relative paths throughout the session. Using a `PROJECT_PATH` prefix is safer.
- **Storing project path in a config file** — rejected as overengineering for an instruction-based system. The orchestrator just needs to resolve it at the start of each session.

### Verification criteria

1. In a workspace with `rust/.planning/` and `elixir/.planning/`, running `/crafter:do --project rust fix X` should look for tasks in `rust/.planning/tasks/`.
2. In a workspace with one subproject `rust/.planning/`, running `/crafter:do fix X` (no `--project`) should auto-detect and use `rust/`, mentioning the `--project` shortcut.
3. In a workspace with multiple subprojects and no `--project`, the orchestrator should list them, ask which one, and mention the `--project` shortcut in the prompt.
4. In a single-project setup (`.planning/` at root), behavior is unchanged.
5. Resume detection with "continue" or "pokracuj" reliably finds active tasks by actually reading files.
6. When `$ARGUMENTS` contains "fix the parser in rules/core.md", the orchestrator proceeds to scope detection without asking "What do you want to do?".
7. No TypeScript/syntax errors (not applicable — all files are Markdown).

### Unknowns / flags

- The `{PROJECT_PATH}` notation is used as a placeholder in the instruction text. Since these are natural-language instructions interpreted by an LLM (not code), this should be clear enough. But if the orchestrator struggles with it, it may need to be rephrased more explicitly.

## Decisions

## Outcome
Commit `47745a3`. All 5 steps implemented: project resolution with `--project` flag and auto-discovery, Grep-based resume detection with resume-intent words, guardrails against ignoring clear user input, `{PROJECT_PATH}` applied across all rule files.
