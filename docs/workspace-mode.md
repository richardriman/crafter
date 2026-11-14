# Workspace Mode (Multi-Repo)

> Status: Draft
> Date: 2026-11-14

**Workspace mode** lets Crafter treat a workspace — a root repo plus several nested repos that only work together — as the unit of work. One task, one plan, one check and one review span every repo the change touches.

Single-repo projects behave exactly as today. Workspace mode activates only when a workspace is detected.

---

## Problem & Motivation

- **One repo per run.** `--project <path>` picks one nested repo for `/crafter-do` and `/crafter-review`; every agent spawn gets `repository: {PROJECT_PATH}` (`rules/delegation.md` § Target Repository). Few users know the flag, and they should not have to.
- **Detection stops at the root.** Auto-detection scans one level deep only when the cwd has no `.crafter/`. A workspace root with its own `.crafter/` never reaches its nested repos.
- **Changes are cross-repo.** In a typical workspace (a root repo holding the devcontainer, run/test tooling and docs, plus nested repos), a change usually spans 2 repos (two backends, or backend + frontend), sometimes 4. Today that means one of:
  - separate, unlinked runs;
  - paired per-repo tasks sharing a branch name, with the contract written into one task's Outcome;
  - one hand-maintained multi-repo task file in the workspace root.
- **Team visibility is uneven.** Nested repos' `.crafter/` are versioned and shared via PRs; the workspace root `.crafter/` may be unversioned, so tasks there are invisible to the team.
- **Worktrees break the environment.** The devcontainer mounts the whole workspace folder, and run tooling (Procfile, Makefile, volumes) uses fixed nested-repo paths. Ad-hoc git worktrees end up scattered inside the workspace and are invisible or broken inside the container — a worktree's `.git` file carries an absolute host gitdir path.

---

## Goals / Non-Goals

**Goals**

- A cross-repo change is one task with one plan, one cross-repo contract, one check and one review.
- Users never need `--project` in a workspace.
- Task files are shared team artifacts.
- Work happens in worktrees that the devcontainer can run and test.

**Non-goals**

- Changing single-repo behavior.
- Depending on any specific issue tracker.
- Merging PRs — Crafter never merges.
- `/crafter-debug` in workspace mode.

---

## Terminology

| Term | Meaning |
|------|---------|
| **Workspace** | A set of repos that only work together, described in one place. |
| **Workspace root** | The top-level repo holding the devcontainer, run/test tooling, docs and the workspace description. |
| **Nested repo** | A repo cloned at a fixed relative path inside the workspace root. |
| **External repo** | A repo listed in the workspace description by URL, not cloned locally (wiki, devops, gitops). |
| **Workspace description** | The section of the workspace's CLAUDE.md / AGENTS.md describing repos, relationships and conventions. |
| **Touched repo** | A repo the current task changes. |
| **Cross-repo contract** | An interface one repo provides and another consumes, with its fixed shape. |
| **Linked PR set** | One PR per touched repo (plus the workspace root), each linking the others with the merge order. |

---

## Decisions

### Workspace description

- **Lives in the workspace's CLAUDE.md / AGENTS.md.** Not all work runs through Crafter; instruction files reach every agent.
- **`.crafter/WORKSPACE.md` is only a fallback/supplement**; instruction files take precedence. `@` imports are not relied on.
- **Content:** a repo table (name, main/base branch, location — relative path = cloned locally, URL = external); relationships (who provides which API to whom, generated copies, shared translations); branch/PR/label conventions; how to run and test the system. — Relationships are what scanning cannot discover.
- **`/crafter-map-project` drafts the section in a workspace; the user approves before it is written.** — The description is team-owned text.
- **External repos are context only by default.** If a task must change one (e.g. add an env var in gitops), Crafter clones it into a scratch location and it follows the normal branch/PR path. — No silent edits to repos nobody has locally.

### Conventions precedence

- **Workspace/project conventions win over Crafter's defaults.** Crafter's `<prefix>/<slug>` (fix/feature/refactor/docs/chore) is a fallback, not a rule. — The team's conventions predate Crafter.
- **One branch name, decided by the workspace, identical in all touched repos.** If the workspace sets none, use the strictest convention among the touched repos (e.g. `<type>/<TRACKER-ID>-<slug>` when any repo requires an ID). — One name links the change everywhere.
- **PR title, labels and commit format follow each repo's own instructions.** — Repos keep their own review rules.
- **Issue-tracker ID is used in branch names when the task has one.** Tracker integration (e.g. via MCP) is optional; if none is available, nothing is enforced. ID format is configurable in the workspace description. — Crafter is open source and must not depend on a tracker; preview environments may key off the ID.

### Tasks

- **A task lives in the workspace root `.crafter/tasks/`, even when it touches one repo.** — One place to find all work.
- **Starting `/crafter-do` inside a nested repo detects the parent workspace and creates the task there.** The starting repo is the default candidate; the Planner adds others. — Users start where they are.
- **Task file naming:** `<date>-<TRACKER-ID>-<slug>.md` with an ID, else `<date>-<slug>.md`. — Avoids collisions in a shared directory.
- **The plan/contract gains a Repos table** (repo, base branch, what changes), **cross-repo contracts**, and **merge order**. — Replaces the "contract in one task's Outcome" pattern.
- **Nested repos keep their `.crafter/` (PROJECT, ARCHITECTURE) as per-repo context only**, passed to agents for touched repos; no new tasks are created there.
- **`--project` stays as a manual override** to force a single repo; documented as secondary.

### Execution

- **Same branch name in every touched repo.**
- **Work happens in git worktrees, never by switching branches in the main checkouts.** — Branch switching collided in practice.
- **Worktrees are not created inside the workspace folder.** Exact layout relative to the devcontainer is decided in Phase 0.
- **Commit only when the check of the whole change passes.** — The green-commit rule applies to the task, not per repo.

### Check, review, delivery

- **One Checker checks the cross-repo contract against all repos' diffs at once.** — Drift across a boundary is only visible with both sides.
- **One Reviewer sees all repos' ranges at once.** — Catches boundary bugs, e.g. a backend renames a field the frontend still decodes.
- **`/crafter-review` accepts several PRs** (`/crafter-review <PR1> <PR2>`), alongside today's single PR, `base..head` range, or uncommitted working tree.
- **One PR per repo, each with a "Related PRs" section listing the others and the merge order.** — Reviewers see the whole change.
- **All PRs are opened together after the whole check is green; Crafter never merges.**

---

## Team Collaboration

- **The task file is a shared team artifact.** It is committed to the workspace root repo on the task branch; the workspace root gets its own PR in the linked PR set (task file + any glue changes).
- **No `STATE.md` in a workspace.** Orientation at run start comes from active task files and the workspace description. `/crafter-status` lists active tasks (title, ID, branch, repos, plan state). Steps 7–9 and `/crafter-map-project` do not write `STATE.md` there; an existing one is left alone. — A tracker-driven workspace gets no value from it, one shared Current Focus is meaningless in a team, and resume and statusline already run on task files.
- **Migration of nested-repo tasks:** completed tasks stay as history; only the operator's own active tasks (on their branch) are migrated, and only with consent; teammates' tasks are never touched.
- **Mixed adoption:** resume detection looks in both the workspace root and nested repos; new tasks are created only in the workspace. When Crafter notices old-style use (running inside a nested repo under a workspace, or a new task about to be created there), it suggests switching in one sentence and never forces.
- **Worktree instances are local per person**; tracker IDs make branch names unique across people.

---

## Phased Plan

| Phase | Scope |
|-------|-------|
| **0** | Spike/prototype: per-task worktree layout outside the workspace folder that still lets the devcontainer run and test the whole system. The result becomes a rule in Crafter and in the workspace's instructions. |
| **1** | Workspace detection and description in CLAUDE.md via `/crafter-map-project`; conventions-precedence rule. |
| **2** | Multi-repo task in `/crafter-do`: plan with Repos table, cross-repo contracts and merge order; same branch everywhere; worktrees per Phase 0; whole-task commit; task file in the workspace root repo. |
| **3** | Multi-repo Checker and Reviewer; `/crafter-review` with several PRs. |
| **4** | Linked PRs with merge order. |
| **5** | Migration of nested-repo tasks with consent; old-style detection hint; `/crafter-status` from task files. |

Phases 0 and 1 can run in parallel; Phase 2 needs both.

---

## Open Questions

### Where do per-task worktrees live, and how does the devcontainer see them? (Phase 0)

**Candidate under evaluation:** a per-task workspace instance outside the main folder — a worktree of the workspace root repo, plus worktrees of touched repos at their canonical nested paths on the task branch, and untouched repos at their base so the whole system runs. It gets its own isolated container stack (a second full checkout with its own compose project name already works in practice).

**To validate:**

- The whole stack starts.
- Git works inside the container — worktree gitdir paths are absolute host paths; either mount the main repos' git dirs or use relative worktree paths.
- Time and disk cost per instance.
- Sharing dependency caches and the database.

---

## Rejected Alternatives

| Alternative | Why rejected |
|-------------|--------------|
| Repo-only worktrees outside the workspace | The devcontainer cannot see them; run tooling uses fixed nested-repo paths. |
| A single hidden worktree dir inside the workspace | Visible to the container, but run tooling still points at the main checkouts; worktrees inside the workspace folder were rejected. |
| Paired per-repo tasks | The cross-repo contract gets lost in one task's Outcome. |
| Auto-scanning nested `.git` instead of a description | Finds repos but not their relationships. |
| Per-repo reviewers | Miss cross-repo boundary bugs. |
| `.crafter/WORKSPACE.md` as primary source | Not all work runs through Crafter; `@` imports are not relied on. |
