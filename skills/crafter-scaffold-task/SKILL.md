---
name: "crafter-scaffold-task"
description: "Scaffold a new Crafter task file from templates/TASK.md (orchestrator/agent-only)"
user-invocable: false
---

# Crafter Task Scaffold Skill

This is a **composable, agent-only** skill. It is invoked by the Crafter
orchestrator (or another sanctioned agent) during the task-lifecycle setup —
not by the user directly. Its single job is to materialize a new task file
from the canonical template so every task starts with the same structure
(Metadata, verbatim Request, Plan placeholder, Decisions, Outcome) instead of
being hand-assembled each time.

It performs a pure scaffold: it does **not** plan, implement, commit, or fill
the `## Plan` section — that remains the Planner agent's responsibility per
`rules/task-lifecycle.md`.

## Procedure

1. Resolve the context directory: prefer `.crafter/`, fall back to legacy
   `.planning/` only if `.crafter/` does not exist (mirror the resolution in
   `skills/crafter-status/SKILL.md`). The task file lives in
   `<context-dir>/tasks/`.
2. Read the canonical template `templates/TASK.md`. Do not invent structure —
   the template's section order and headings are authoritative.
3. Derive the filename: `<YYYYMMDD>-<slug>.md`, where `<YYYYMMDD>` is the
   supplied date with dashes removed and `<slug>` is the kebab-case title
   (lowercase, non-alphanumerics collapsed to single dashes, trimmed).
4. Collision check: if `<context-dir>/tasks/<filename>` already exists, **stop**
   and report the collision. Never overwrite an existing task file.
5. Fill the template:
   - `# Task:` ← the brief title
   - `**Date:**` ← supplied date (`YYYY-MM-DD`)
   - `**Work branch:**` ← supplied branch (must not be `main`/`master` for
     fresh work; if omitted, leave the template comment in place rather than
     guessing)
   - `**Status:**` ← `active`
   - `**Scope:**` ← supplied scope (`Small` | `Medium` | `Large`)
   - `## Request` ← the original user request, quoted verbatim, unedited
   - Leave `## Plan` as `_(pending)_` with its guidance comment intact
   - Leave `## Decisions` and `## Outcome` as their template comments
6. Write the file to `<context-dir>/tasks/<filename>` and return the path to
   the caller. Do not commit.

## Skill Contract

### Capability

Generates a ready-to-plan Crafter task file from `templates/TASK.md`, with
Metadata and the verbatim Request pre-populated, in the resolved context
directory's `tasks/` folder.

### When-Applies

Use during task-lifecycle setup when the orchestrator needs to create a new
task file and no task file for this work exists yet. Not applicable when
resuming or editing an existing task, and never invoked directly by the user
(`user-invocable: false`).

### Required Inputs

- `title` (required) — brief task title used for the `# Task:` heading and to
  derive the slug
- `request` (required) — the original user request, quoted verbatim
- `date` (required) — task date in `YYYY-MM-DD`
- `scope` (required) — `Small` | `Medium` | `Large`
- `work-branch` (optional) — topic branch where the task is resumed/executed;
  if absent, the template's branch comment is left untouched

### Outputs

- `<context-dir>/tasks/<YYYYMMDD>-<slug>.md` — new task file populated from
  `templates/TASK.md`

### Allowed Side Effects

- Reads `templates/TASK.md` to source the structure
- Reads existing files in `<context-dir>/tasks/` to perform the collision check
- Suggests `git mv .planning .crafter` if only the legacy directory exists

### Forbidden Side Effects

- Inherits all base forbidden side effects from `docs/skill-contract.md`
- Must not write the `## Plan` content — planning is the Planner agent's job
- Must not overwrite or modify an existing task file
- Must not write outside `<context-dir>/tasks/`
- Must not create a git commit
- Must not edit installed runtime copies under `~/.claude` or equivalents

### Success Criteria

- `<context-dir>/tasks/<YYYYMMDD>-<slug>.md` exists
- Its section headings match `templates/TASK.md` in order
- Metadata fields are filled from the inputs; `Status` is `active`
- `## Request` contains the supplied request verbatim
- `## Plan` still reads `_(pending)_` with its guidance comment

### Failure Criteria

- `title`, `request`, `date`, or `scope` is missing or empty
- A task file with the derived filename already exists
- `templates/TASK.md` cannot be read
- The resolved context directory has no `tasks/` directory and one cannot be
  created within the declared output path
