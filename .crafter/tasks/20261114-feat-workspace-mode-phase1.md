# Task: Workspace mode Phase 1 — workspace detection, workspace description, conventions precedence

## Metadata
- **Date:** 2026-11-14
- **Work branch:** feat/workspace-mode-phase1
- **Status:** completed
- **Scope:** Medium

## Request
Implement Phase 1 of `docs/workspace-mode.md`: workspace detection and the workspace description in CLAUDE.md/AGENTS.md via `/crafter-map-project`, plus the conventions-precedence rule (workspace/project conventions win over Crafter's default branch naming). Scope strictly Phase 1 per the design doc's Decisions (Workspace description, Conventions precedence) and the Tasks bullet on detection from a nested repo only as far as detection itself; multi-repo task execution, worktrees, multi-repo check/review, linked PRs and migration are later phases. Single-repo projects must behave exactly as today. Keep everything generic (no real workspace/company names).

## Plan
**Plan status:** approved

### 1. Complete request
Deliver Phase 1 of `docs/workspace-mode.md`: (a) one precise workspace-detection rule, (b) `/crafter-map-project` drafts the workspace description into the workspace's CLAUDE.md / AGENTS.md after user approval, (c) the conventions-precedence rule for branch naming and commit format. Why: Phase 2 (multi-repo tasks) needs a described workspace and a settled branch-naming rule to build on. Acceptance: a single-repo project with no workspace signal behaves exactly as today; all text is generic (no real company, repo, person, tracker).

**Detection rule (the definition this plan fixes):** walk from the current working directory up through its ancestors — up to the user's home directory (inclusive) when the cwd is inside it, otherwise up to the filesystem root; the first directory whose `CLAUDE.md` or `AGENTS.md` contains a line exactly `## Workspace`, or which contains `.crafter/WORKSPACE.md`, is the **workspace root** and a workspace is detected. Nothing else activates it — not nested git repos, not nested `.crafter/` dirs, not `--project`. Instruction-file section wins over `WORKSPACE.md` (which is read as a supplement only). Detection runs after Project Resolution and does **not** change `PROJECT_PATH`/`CRAFTER_DIR` in Phase 1 — a workspace root with its own `.crafter/` still resolves to itself, the one-level scan and `--project` work exactly as today.

**Workspace candidate (map-project only):** no workspace detected and the cwd has at least two non-hidden direct subdirectories containing `.git` (dir or file). It only triggers an offer to draft a description; it never activates workspace mode.

**Phase 1 behavior when a workspace is detected in `/crafter-do`:** read the description for orientation, use its conventions for the branch proposal, and say one sentence that the workspace was detected and this run still changes only `PROJECT_PATH` (multi-repo tasks come later). No question, no gate, no task relocation.

**STATE.md call:** in Phase 1, `/crafter-map-project` never creates or updates `STATE.md` when a workspace is detected (root or nested repo); an existing one is left alone. `/crafter-do` Steps 7–9 and startup STATE.md handling stay unchanged until Phase 2 (task files in the workspace root give the replacement orientation); `/crafter-status` stays Phase 5. Reason: suppressing STATE.md in `/crafter-do` before task-file orientation exists would remove Current Focus with nothing replacing it; map-project skipping it is harmless.

### 2. Assumptions / interpretations
- Marker heading is the exact line `## Workspace`; subsections inside it use `###`, so the section ends at the next `## ` heading or EOF. Fixed subsection shape: `### Repositories` (table: Name | Base branch | Location — relative path = cloned, URL = external), `### Relationships`, `### Conventions` (branch naming incl. optional tracker-ID format, PR title/labels, commit format), `### Run and test`.
- Conventions precedence also applies to single repos whose own CLAUDE.md / AGENTS.md defines branch/commit conventions (settled decision: "workspace/project conventions win"). With no such conventions, behavior is identical to today.
- Order for the branch name: workspace description convention → else the strictest convention among touched repos' instruction files (Phase 1: just `PROJECT_PATH`) → else Crafter's `<prefix>/<slug>` fallback. A tracker ID is included when the request carries one; if the convention wants one and none is known, the proposal goes to the user in the existing branch question — never a new gate, never required.
- Description target file at write time: the file already holding the section → else existing `CLAUDE.md` → else existing `AGENTS.md` → else create `CLAUDE.md`; the user confirms the target with the content and may choose `.crafter/WORKSPACE.md` instead.

### 3. Non-goals
Multi-repo tasks, task relocation to the workspace root, Repos table in plans, worktrees, multi-repo check/review, linked PRs, migration, `/crafter-status` changes, STATE.md changes in `/crafter-do`, `/crafter-debug` workspace behavior, `/crafter-review`, Go CLI/statusline, main/master guard honoring a non-`main` base branch, mapping every nested repo from the root, `.claude/CLAUDE.md` as a detection location.

### 4. Relevant areas
`docs/workspace-mode.md`; `rules/` (new fragment), `rules/task-lifecycle.md`, `rules/do/step-0-resume.md`, `rules/post-change.md`, `rules/do/step-7-9-post-change.md`; `skills/crafter-do/SKILL.md` (rules list, Project Resolution → context read, Startup, master-plan table); `rules/do/project-resolution.md` (must stay behaviorally unchanged); `skills/crafter-map-project/SKILL.md`; `agents/crafter-analyzer.md` (mode A); `install.sh` + `tests/test_install.sh` `_EXPECTED_FILES_REL`; `.crafter/ARCHITECTURE.md` structure tree; `README.md`.

### 5. Steps
- [x] Step 1: The detection rule and description shape exist in exactly one place — a new per-concern rules fragment (crafter-* naming not required for rules) — installed by `install.sh`, listed in `tests/test_install.sh`, and shown in `.crafter/ARCHITECTURE.md`'s tree.
- [x] Step 2: `/crafter-do` loads that fragment, runs detection after Project Resolution, and when detected reads the description for orientation and emits the one-sentence notice; Project Resolution text and outcomes unchanged (SKILL.md and `rules/do/project-resolution.md` stay in agreement).
- [x] Step 3: Conventions precedence lands in `rules/task-lifecycle.md` (branch naming) and the main/master guard in `rules/do/step-0-resume.md` (plus any mirrored wording in `skills/crafter-do/SKILL.md`); `rules/post-change.md`'s conventional-commits line becomes the fallback behind the repo's own commit format (PR title follows the commit subject as today).
- [x] Step 4: `/crafter-map-project` loads the fragment; detected workspace root → Analyzer proposes create/update of the description; workspace candidate → one question offering to draft it; nested repo under a detected workspace → today's per-repo mapping plus a one-sentence pointer to the root; in all workspace cases STATE.md is neither drafted nor written. Draft + target file are approved before writing; writing replaces only the `## Workspace` section. Analyzer mode A documents the description sub-case (evidence-based drafting, unclear relationships/external repos flagged as questions, no STATE.md draft in a workspace).
- [x] Step 5: README gains a short user-facing note: workspace description via `/crafter-map-project`, branch conventions precedence, `--project` as secondary override.
- [x] Check

### 6. Contract
- **Outcome:** a described workspace is detected by one rule; map-project drafts/updates the description with approval; branch names and commit format follow workspace/project conventions with Crafter's defaults as fallback.
- **Scope boundary:** the files in § 4 only; no Go code; no new agent or skill.
- **Non-goals:** § 3.
- **Seams:** see § 7.
- **Verification evidence:** `bash tests/test_install.sh` passes with the new fragment in the expected list; Implementer reports a scenario walkthrough citing rule text for: (1) single repo, no markers → unchanged; (2) cwd = workspace root with `## Workspace` in CLAUDE.md, with and without root `.crafter/`; (3) cwd = nested repo under a described root; (4) only `.crafter/WORKSPACE.md`; (5) `--project` inside a workspace; (6) undescribed root with nested repos in map-project; (7) branch proposal with workspace convention, repo-only convention, none; grep shows `## Workspace` marker spelled identically everywhere and no real names.
- **Stop conditions:** the precedence rule cannot be expressed without changing single-repo behavior beyond "project conventions win when defined"; detection would need to change `PROJECT_PATH`; any step needs Go/statusline changes; the marker heading collides with an existing Crafter heading convention.

### 7. Seams
- The detection rule + `## Workspace` section shape in the new rules fragment (consumed by `/crafter-do`, `/crafter-map-project`, later phases).
- `/crafter-do` startup: Project Resolution outputs (`PROJECT_PATH`, `CRAFTER_DIR`) unchanged; new detected-workspace root available for orientation and conventions only.
- `/crafter-map-project` user-visible flow: approval gate before any write; written artifact = `## Workspace` section in CLAUDE.md / AGENTS.md (or `.crafter/WORKSPACE.md` if chosen).
- Branch-proposal and commit-format rules in `rules/task-lifecycle.md`, `rules/do/step-0-resume.md`, `rules/post-change.md`.
- Installed file list: `install.sh` ↔ `tests/test_install.sh` `_EXPECTED_FILES_REL`.

### 8. Alternatives considered
- Detect by nested `.git` scan — rejected by the design (finds repos, not relationships); kept only as a map-project offer trigger.
- Put the rule in `rules/core.md` — loaded by every skill incl. debug/status; a per-concern fragment matches the architecture convention and later phases will grow it.
- `templates/WORKSPACE.md` — the shape is a few lines and must live next to the detection marker; a template adds an install entry without adding a source of truth.
- Fold detection into Project Resolution (re-root to workspace) — that is Phase 2 behavior and would change today's resolution.
- Suppress STATE.md in `/crafter-do` now — removes orientation before its replacement exists.

### 9. Risks / unknowns / flags
- `## Workspace` may already exist in some user's CLAUDE.md with another meaning → false detection; effect is only orientation + one sentence + conventions read. Confirm the marker name.
- Workspace-candidate threshold "≥1 nested git repo" also fires on repos with a submodule (one extra skippable question in map-project). Alternative: ≥2.
- Conventions precedence changes branch proposals for single repos that define their own conventions (settled, but visible).
- Walk-up to filesystem root could match an unrelated ancestor (e.g. a home-level CLAUDE.md with the marker).
- Main/master guard still keys on `main`/`master`, not a workspace-declared base branch (deferred; flag for Phase 2).

**Summary:** the contract protects today's single-repo behavior and Project Resolution while giving later phases a single, unambiguous workspace signal and a user-approved description to build on.

## Decisions
- **Decision (User Accepted):** Marker heading stays `## Workspace`. **Reason:** short and readable; a heading collision is handled by renaming during mapping.
- **Decision (User Accepted):** A workspace *candidate* (for `/crafter-map-project` only) needs at least **two** non-hidden subdirectories containing `.git`, not one. **Reason:** one nested repo also matches a repo with a single submodule.
- **Decision (User Accepted):** The upward search for a workspace root stops at the user's home directory (inclusive), never the filesystem root. **Reason:** avoids matching an unrelated ancestor directory.
- **Decision (Tech Debt — auto-recorded):** Suggestion — `rules/task-lifecycle.md:21`: "Include an issue-tracker ID when the request carries one" is unconditional while the same paragraph says branch proposals are unchanged with no defined convention; unclear whether the ID is added without a convention.
- **Decision (Tech Debt — auto-recorded):** Suggestion — `skills/crafter-map-project/SKILL.md:91` / `rules/workspace.md:11`: marker detection and section boundaries do not exclude fenced code blocks; an instruction file quoting the shape could trigger false detection or truncate the replaced section.
- **Decision (Tech Debt — auto-recorded):** Suggestion — plan text (Approach/detection definition) still says search to the filesystem root and candidate "at least one", contradicting the accepted Decisions.
- **Decision (User Accepted):** Plan amended after the check pass (D1 plan-obsoleting discovery): `rules/do/step-7-9-post-change.md` joins Relevant areas so its consolidated-commit wording follows § COMMIT; the detection search stops at home only when the cwd is inside home, otherwise it continues to the filesystem root (fixes R1: devcontainer paths like `/workspaces/<name>` sit outside home); plan text aligned with the accepted Decisions (R7). D2 (cwd outside home → check only cwd) is superseded by this rule.
- **Decision (Tech Debt — auto-recorded):** Suggestion (R5@1) — precedence order differs: branch naming lets the workspace win over the repo, commit/PR format lets the repo win over the workspace; neither `rules/workspace.md` § Phase 1 behavior nor README states the difference.
- **Decision (Tech Debt — auto-recorded):** Suggestion (R6@1) — detection starts from the cwd, not `PROJECT_PATH`; with `--project` pointing outside the cwd's tree the wrong (or no) workspace conventions apply.
- **Decision (User Accepted):** Scope drift N1–N3 (pass 1) accepted as a plan amendment: PR title and labels follow repo instructions → workspace `### Conventions` → commit subject; `rules/do/step-9b-pr-composition.md` and `skills/crafter-do/SKILL.md` Step 9b join Relevant areas.
- **Decision (Tech Debt — user-deferred):** Minor (R4@1) — `/crafter-debug` loads `task-lifecycle.md` and commits via `post-change.md` but does not load `rules/workspace.md` or run detection, so in a workspace its branch/commit naming ignores workspace conventions (`/crafter-debug` is out of scope and may be removed).
- **Decision (User Accepted):** Detection marker tightened (fix for R2@1): a workspace root needs the exact line `## Workspace` **and** a `### Repositories` subsection inside that section, so an unrelated `## Workspace` heading in a single-repo instruction file does not activate workspace mode.
- **Decision (Tech Debt — auto-recorded):** Suggestion (R5@2) — README promises a `## Workspace` draft when run in a workspace root, but an undescribed root is offered one only via the candidate rule (≥2 direct subdirectories with `.git`); roots with repos under e.g. `services/*` get no offer.
- **Decision (Tech Debt — auto-recorded):** Suggestion (R6@2) — detection runs from the cwd, not `PROJECT_PATH` (duplicate of R6@1); a `## Workspace` line inside a fenced code block also counts.
- **Decision (Orchestrator Accepted):** Beneficial local drift (D1@3) — the section boundary is now defined once in `rules/workspace.md` § Section boundary: the `## Workspace` section ends just before the next level-1 or level-2 heading or EOF, and lines inside fenced code blocks never count. This supersedes the plan's "next `## ` heading or EOF". **Reason:** required by the R1@2/R3@2 fix; scope unchanged.
- **Decision (Tech Debt — auto-recorded):** Suggestion (R4@3/R5@3) — the Step 9b PR-title snippet runs both branches in sequence; executed literally without a convention it would title the PR with the placeholder (`rules/do/step-9b-pr-composition.md:48`, `skills/crafter-do/SKILL.md:349`).
- **Decision (Tech Debt — auto-recorded):** Suggestion (R6@3) — no rule for which file wins when both `CLAUDE.md` and `AGENTS.md` carry the marker.
- **Decision (Tech Debt — auto-recorded):** Suggestion (R7@3) — `rules/workspace.md` § Phase 1 behavior says `### Conventions` drives only branch naming, but commit format and PR title/labels also use it, with the opposite repo/workspace priority.
- **Decision (Tech Debt — auto-recorded):** Suggestion (R3@4) — `--label <name>` has no quoting guidance; label names with spaces or colons split into several arguments (`rules/do/step-9b-pr-composition.md:54`, `skills/crafter-do/SKILL.md:355`).
- **Decision (Tech Debt — auto-recorded):** Suggestion (R4@4) — map-project fallback sentence uses "mode A" (Analyzer Project Mapping) next to "Step 2 case A/B"; the letter A means two different things.
- **Decision (Tech Debt — auto-recorded):** Suggestion (R5@4) — the nested-repo bullet says "map this repo as usual" while the workspace-mode paragraph forbids `STATE.md` and case A still lists `STATE.md` without a workspace exception; the rule must be assembled from three places.
- **Decision (Tech Debt — auto-recorded):** Suggestion (R6@4) — duplicate of R6@3: no precedence when both `CLAUDE.md` and `AGENTS.md` carry the marker; case D step 5 may write the file that did not win detection.
- **Decision (Tech Debt — auto-recorded):** Suggestion (R4@5) — the rationale under the Step 9b title block still only mentions capturing the commit subject verbatim; the convention-derived heredoc branch is not explained (`rules/do/step-9b-pr-composition.md:67`).
- **Decision (User Accepted):** Manual override of the 5-iteration fix-loop cap — iteration 6 authorized to fix R1@5/R2@5 (unquoted convention labels) and R3@5 (§ Phase 1 behavior omits commit/PR conventions).
- **Decision (Tech Debt — auto-recorded):** Suggestion (R3@6) — outside home the upward search climbs to `/`; with `--project` outside the cwd a foreign ancestor instruction file with the marker could impose conventions in an unattended `--auto` run.
- **Decision (Tech Debt — auto-recorded):** Suggestion (R4@6) — PR precedence (repo → workspace) is ambiguous per field (title vs labels) and on whether a general PR-content rule in a repo's instructions counts as a "defined convention".
- **Decision (User Accepted):** Second manual override of the fix-loop cap — iteration 7 authorized to fix R1@6 (split title/PR blocks lose `$TITLE`) and R2@6 (Step 9b git/gh commands run in the cwd instead of `PROJECT_PATH`).
- **Decision (Orchestrator Accepted):** Beneficial local drift (D1@7) — the PR body is now written via a quoted heredoc to `.crafter/run/<task-id>/pr-body` and passed with `--body-file`, instead of `printf '%s' "<full-body>" |`. **Reason:** removes the body interpolation risk; stays within the amended Step 9b scope; the run dir cleanup removes the file.
- **Decision (Tech Debt — auto-recorded):** Suggestion (R5@7/R6@7) — a convention that defines only labels (no title rule) has no explicit block in Step 9b step 5.
- **Decision (User Accepted):** Step 9b reverted to its `main` state instead of a further fix iteration; PR title/label conventions move to Phase 4 (linked PRs), to be written together with a test that actually runs the PR-creation blocks. Open findings R1@7–R4@7 are dropped with the reverted code. Pre-existing on `main` and recorded for Phase 4: Step 9b runs `git log`/`gh` in the cwd instead of `PROJECT_PATH` (wrong repo under `--auto --project`), and the PR body is interpolated via `printf '%s' "<full-body>"`.
- **Decision (Tech Debt — user-deferred):** Minor (R1@8) — `/crafter-debug` commits via `rules/post-change.md` § COMMIT, which now references the workspace `### Conventions`, but debug does not load `rules/workspace.md` or run detection (duplicate of R4@1). Deferred on the user's instruction to commit Phase 1 after the Step 9b revert.
- **Decision (Tech Debt — user-deferred):** Minor (R2@8) — `rules/post-change.md` § COMMIT says "the repo's own `CLAUDE.md` / `AGENTS.md`" without naming `{PROJECT_PATH}`; with cwd = workspace root and `PROJECT_PATH` = nested repo only the root's instruction file is auto-loaded. Deferred on the same instruction.
- **Decision (Tech Debt — auto-recorded):** Suggestions (R3@8–R7@8) — repo-vs-workspace precedence when `PROJECT_PATH` is the workspace root; no precedence between `CLAUDE.md` and `AGENTS.md` markers; "the `## Workspace` section in the instruction file wins" ambiguous for a non-marker heading; start directory not stated as absolute before the ancestor walk; fenced-block rule names only opening fences.

## Outcome
Phase 1 implemented: new `rules/workspace.md` (detection marker `## Workspace` + `### Repositories` in CLAUDE.md/AGENTS.md or `.crafter/WORKSPACE.md`; start directory = resolved `PROJECT_PATH` for `/crafter-do`, cwd for `/crafter-map-project`; upward search to home when inside it, else to `/`; section boundary defined once; description shape; Phase 1 behavior), installed and listed in the install test. `/crafter-do` runs detection after Project Resolution without changing `PROJECT_PATH`; conventions precedence for branch naming (`rules/task-lifecycle.md`, main/master guard) and commit format (`rules/post-change.md`, step-7-9). `/crafter-map-project` drafts the `## Workspace` section with approval, offers it for candidates (≥2 nested repos), skips `STATE.md` in workspace mode with a fallback. README section. PR title/label conventions were reverted (Step 9b back to `main`) and moved to Phase 4. Fix loop: 7 iterations (2 manual overrides) plus the revert; deferred items listed in Decisions. Install test not re-run after the final edits (sandbox blocks `mktemp`; last confirmed 64/0 by the first check pass).
