# Workspace

A **workspace** is a set of repos that only work together: a **workspace root** (the top-level repo with the run/test tooling and docs) plus **nested repos** cloned at fixed relative paths inside it.

## Detection

Run after Project Resolution. The **start directory** is the resolved `PROJECT_PATH` in `/crafter-do` and the current working directory in `/crafter-map-project`. Check the start directory, then each ancestor directory. If the start directory is inside the user's home directory, stop after the home directory (inclusive) — never search above it; otherwise search up to the filesystem root.

The first directory that matches either of these is the **workspace root**, and a workspace is detected:

- its `CLAUDE.md` or `AGENTS.md` contains a line that is exactly `## Workspace` **and**, inside that section (see § Section boundary), a line that is exactly `### Repositories` — a `## Workspace` heading without it is not a marker; or
- it contains `.crafter/WORKSPACE.md`.

Nothing else activates a workspace — not nested git repos, not nested `.crafter/` directories, not `--project`. When both sources exist, the `## Workspace` section in the instruction file wins; `.crafter/WORKSPACE.md` is read only as a supplement.

Detection never changes `PROJECT_PATH` or `CRAFTER_DIR`. Project Resolution, its one-level scan and `--project` work exactly as without a workspace.

**Workspace candidate** (`/crafter-map-project` only): no workspace is detected and the cwd has at least **two** non-hidden direct subdirectories that contain `.git` (a directory or a file). A candidate only triggers an offer to draft the description; it never activates a workspace.

## Workspace description

The `## Workspace` section in the workspace root's `CLAUDE.md` or `AGENTS.md` (or the whole of `.crafter/WORKSPACE.md`). Subsections use `###`. The `### Repositories` subsection is required in the instruction-file form — it is part of the detection marker. Shape:

```markdown
## Workspace

### Repositories

| Name | Base branch | Location |
|------|-------------|----------|
| root | main | `.` |
| backend | main | `services/backend` |
| docs-wiki | main | https://example.com/org/docs-wiki.git |

A relative path means the repo is cloned locally; a URL means an external repo (context only).

### Relationships

Who provides which API to whom, generated copies, shared translations — what scanning cannot discover.

### Conventions

- Branch naming: e.g. `<type>/<TRACKER-ID>-<slug>`; tracker-ID format, if any (e.g. `ABC-123`).
- PR title and labels.
- Commit format.

### Run and test

How to start and test the whole system.
```

### Section boundary

The `## Workspace` section starts at its heading line and ends just before the next level-1 or level-2 heading (a line starting with `# ` or `## `) or at the end of the file. Lines inside fenced code blocks (opened by `` ``` `` or `~~~`) never count — not as the `## Workspace` line, not as `### Repositories`, and not as a boundary.

## Phase 1 behavior

When a workspace is detected, `/crafter-do` reads the description for orientation, uses its `### Conventions` for the branch proposal and the commit format, and tells the user in one sentence that the workspace was detected and this run still changes only `PROJECT_PATH` — multi-repo tasks come later. No question, no gate, no change to where the task file lives. The branch proposal puts the workspace before the repo (`rules/task-lifecycle.md` § Conventions precedence), while the commit format puts the repo's own instructions first (`rules/post-change.md` § COMMIT). PR title and label conventions may be documented in the description but are not applied yet.
