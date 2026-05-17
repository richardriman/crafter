# Core Capabilities — Taxonomy and Decomposition Design Note

> Status: Draft
> Date: 2026-05-17
> Task: `.crafter/tasks/20260517-refactor-crafter-do-core-capabilities.md`

## Scope of this document

This document governs **core, Crafter-distributed capabilities only** — internal markdown fragments that Crafter ships and loads as part of the `crafter-do` workflow. It is not concerned with extension skills.

**Extension skills are a separate concept.** An extension skill is a supplemental, third-party skill declared with a Skill Contract block; it lives outside the Crafter core repo and is discovered at project, parent, or global scope. Extension-skill compatibility, safety envelopes, and discovery rules are governed exclusively by `docs/skill-contract.md`. Nothing in this document affects extension-skill semantics.

A **core capability module** is an internal, Crafter-distributed markdown fragment with a short, focused responsibility. It is loaded by `skills/crafter-do/SKILL.md` (or another core skill) the same way `rules/do-workflow.md`, `rules/post-change.md`, etc. are loaded today — by a "Read and follow these rules" reference. Core capability modules are part of Crafter itself. They are not extension skills and are not governed by `docs/skill-contract.md`.

---

## Capability taxonomy

The table below maps every existing top-level section and block of `skills/crafter-do/SKILL.md` to a named future capability module and a slice tag.

**Slice tags:**

- **Slice 1 (this task)** — extracted in the current task (`20260517-refactor-crafter-do-core-capabilities.md`).
- **Deferred — follow-up** — deferred to a future slice task; the section stays inline in `crafter-do/SKILL.md` for now.
- **Stays inline in crafter-do** — intentionally not extracted; structurally belongs in the orchestrator entry point.

| Section / block | One-line description | Future capability module | Slice tag |
|---|---|---|---|
| Frontmatter (`---` block) + "Read and follow these rules" loader list | YAML skill metadata and the list of core rule files loaded at startup | — (orchestrator manifest; no extraction) | Stays inline in crafter-do |
| `## Skill options` — `### --fast` and `### --auto` subsections | Declares the two behavioral flags, their trade-offs, and the mutual-exclusion rule for human readers | — (user-facing flag documentation; no extraction) | Stays inline in crafter-do |
| Orchestrator identity block (`You are the orchestrator…`) | Single paragraph establishing orchestrator role and delegation contract | — (structural anchor; no extraction) | Stays inline in crafter-do |
| `## Flag Validation (before anything else)` | Enforces mutual exclusion of `--auto` and `--fast` before any other work starts | `rules/do/flag-validation.md` | Slice 1 (this task) |
| `## Project Resolution (before anything else)` | Resolves `PROJECT_PATH` and `CRAFTER_DIR` from `--project` flag or auto-scan; handles `.planning` legacy fallback and migration offer | `rules/do/project-resolution.md` | Slice 1 (this task) |
| Project-context read block (prose block after `---`, before `## Extension Skills`) | Instructs the orchestrator to read `STATE.md` and selected sections of `PROJECT.md` for orientation | — (structural glue; tightly coupled to project resolution output) | Stays inline in crafter-do |
| `## Extension Skills` | Defines extension-skill discovery (priority table), safety envelope reference, and supplemental-only invariant for the three workflow phases where extension skills apply | `rules/do/extension-skills.md` | Slice 1 (this task) |
| `## Step 0 — Resume Detection` | Detects active tasks to resume, checks branch sanity, enforces main/master guard | `rules/do/step-0-resume.md` | Deferred — follow-up |
| `## Step 1 — Completeness and scope` | Completeness check, scope classification (Small/Medium/Large), extension-skill check for Analyzer delegation | `rules/do/step-1-scope.md` | Deferred — follow-up |
| `## Step 2 — DISCUSS / RESEARCH` | Handles incomplete requests via targeted clarification or Analyzer delegation | `rules/do/step-2-discuss.md` | Deferred — follow-up |
| `## Step 3 — PLAN` | Delegates planning to `crafter-planner`, presents summary, waits for approval, marks plan approved | `rules/do/step-3-plan.md` | Deferred — follow-up |
| `## Step 4 — EXECUTE` | Delegates implementation to `crafter-implementer`, one step at a time; extension-skill check before delegation | `rules/do/step-4-execute.md` | Deferred — follow-up |
| `## Step 5 — STEP DRIFT CHECK` | Delegates step drift verification to `crafter-verifier`; handles recommended actions | `rules/do/step-5-drift.md` | Deferred — follow-up |
| `## Step 5a — PHASE VERIFICATION` | Delegates phase-level verification to `crafter-verifier` after all steps in a phase pass drift checks | `rules/do/step-5a-phase-verification.md` | Deferred — follow-up |
| `## Step 6 — REVIEW` | Delegates code review to `crafter-reviewer`; manages fix loop (5-iteration cap), Critical/Major handling, extension-skill supplemental context | `rules/do/step-6-review.md` | Deferred — follow-up |
| `## Step 6b — Phase Summary and Auto-Commit` | Post-review approval paths: `--auto` branch, auto-approve on clean summary, `--fast` silence-as-approval, explicit default; commits via `post-change.md` | `rules/do/step-6b-phase-summary.md` | Deferred — follow-up |
| `## Step 6a — Session Break` | Suggests `/clear` and re-invoke between phases for Medium/Large scope; skipped for Small | `rules/do/step-6a-session-break.md` | Deferred — follow-up |
| `## Steps 7–9 — Post-Change` | End-of-task housekeeping: doc checks, consolidated commit, STATE.md update, task file completion, session wrap-up; delegates to `post-change.md` | `rules/do/step-7-9-post-change.md` | Deferred — follow-up |
| `## Step 9b — PR Composition` | `--auto`-only PR composition: baseline body, `crafter pr-body` subcommand, `gh pr create`, cleanup hook | `rules/do/step-9b-pr-composition.md` | Deferred — follow-up |

### Notes on "Stays inline" decisions

- **Frontmatter + loader list:** The YAML frontmatter and the "Read and follow these rules" list are the structural entry point of the skill. They cannot be extracted because they are the mechanism by which all other modules are loaded. Any new capability module files introduced by Slice 1 will be added to this loader list.
- **`## Skill options` (`--fast` / `--auto`):** This section is user-facing documentation embedded in the skill file for discoverability. It describes flags that are formally defined in `rules/do-workflow.md`. Moving it would break the user's reference point without adding structural benefit.
- **Orchestrator identity block:** A single structural paragraph that names the orchestrator role. It has no extractable rule content and is too small to justify a module.
- **Project-context read block:** The three-line instruction to read `STATE.md` and `PROJECT.md` is tightly coupled to the variables set by Project Resolution (`PROJECT_PATH`, `CRAFTER_DIR`). Extracting it separately would create a fragment too small and too context-dependent to stand alone.

---

## Loading convention

Core capability modules follow the same loading convention as existing rule files:

- New module files are placed under a new subdirectory of `rules/` — the recommended default is `rules/do/`, but the final subdirectory name is confirmed by the Phase 2.1 Implementer inside that step's contract (see risk R7 in the task plan).
- `skills/crafter-do/SKILL.md` loads them by adding entries to the "Read and follow these rules" list at the top of the file.
- No new mechanism, no YAML manifest, no Go code, no CLI subcommand. The prompt-driven loading that works for `rules/core.md`, `rules/do-workflow.md`, etc. works identically for the new modules.
- The installer (`install.sh`) copies new files under the chosen subdirectory as part of the same `rules/` deployment. Slice 1 includes the mechanical installer edit needed to deploy the new subdirectory.

### Naming scheme

Module files use `rules/do/<step-or-concept>.md` (or the equivalent path under whichever subdirectory the Phase 2.1 Implementer confirms) where `<step-or-concept>` matches the taxonomy table above. The `do/` subdirectory is the recommended default because these modules are all specific to the `crafter-do` orchestrator; the Implementer chooses the final name within the Phase 2.1 contract. The module paths shown in the taxonomy table above are illustrative recommended paths consistent with this default — they are not locked until Phase 2.1. If a future module is shared across multiple orchestrators, it can live directly in `rules/` without a subdirectory — that decision belongs to the task that introduces it.

---

## Deferred slices (follow-up tasks)

The sections tagged "Deferred — follow-up" in the taxonomy table — Step 0 through Step 9b — are candidates for extraction in subsequent slice tasks. The recommended order for follow-up slices is:

1. **Step 0 (Resume Detection)** — high-value isolation; cross-references `rules/task-lifecycle.md`, so validate that those bindings survive the pointer before extracting.
2. **Steps 1–2 (Scope + Discuss)** — low cross-reference density; natural pair.
3. **Steps 3–4 (Plan + Execute)** — medium density; test the pattern on agent-delegation prose.
4. **Steps 5, 5a (Drift + Phase Verification)** — straightforward procedure; few bindings.
5. **Steps 6, 6b, 6a (Review + Phase Summary + Session Break)** — highest density; extract together since they share flag-state and approval-path logic.
6. **Steps 7–9, 9b (Post-Change + PR Composition)** — already heavily delegated to `rules/post-change.md`; extraction is thin pointer work.

This ordering is advisory, not binding. Follow-up tasks own their own scope and sequencing.

---

## Runtime-path policy

### Canonical reference form for new module files

All new core capability module files introduced by this task (and by subsequent slice tasks) MUST refer to the installed Crafter rules directory using the placeholder:

```
{CRAFTER_HOME}/rules/...
```

This is the single, unambiguous canonical form. Do **not** hard-code any concrete runtime path (e.g., `~/.claude/crafter/rules/...`) inside module content.

### Installer as the sole runtime-specific surface

The installer (`install.sh`) is the only place in the codebase that knows the concrete installed-path layout for a given runtime. Today it deploys to `~/.claude/crafter/` (Claude Code surface). Copilot CLI and OpenCode are forward-looking aspirations referenced in task descriptions but not yet implemented in `install.sh`. The `{CRAFTER_HOME}` placeholder used in module content is intentionally runtime-neutral: each runtime's install target becomes the concrete expansion of `{CRAFTER_HOME}` at the point of deployment. The concrete expansion semantics of `{CRAFTER_HOME}` — how and where it resolves at install time, and whether that implies any substitution mechanism — are deliberately not defined here and are deferred to the sibling task `.crafter/tasks/20260421-skills-first-runtime-portability.md`.

No module file, skill file, or rules fragment (other than `install.sh`) should resolve or reference a concrete runtime install path.

### Existing hard-coded references are out of scope here

Files **not touched by this task** that contain hard-coded `~/.claude/...` references are deliberately **not** normalized here. Broader, repo-wide normalization of those references is the responsibility of the sibling task:

`.crafter/tasks/20260421-skills-first-runtime-portability.md`

This task only commits to (a) the new module files following the `{CRAFTER_HOME}/rules/...` convention and (b) not making the existing situation worse.

### Applied in this task

The runtime-path policy was established here and applied to **1 of the 3** Phase 2 module files: `rules/do/extension-skills.md` contained one hard-coded `~/.claude/crafter/skills/` reference in its discovery table, which was replaced with `{CRAFTER_HOME}/skills/`. The other two modules (`rules/do/flag-validation.md` and `rules/do/project-resolution.md`) contained no runtime install paths and needed no change. Repo-wide normalization of existing hard-coded references remains the responsibility of `.crafter/tasks/20260421-skills-first-runtime-portability.md`.

**Slice 2** (task `.crafter/tasks/20260517-refactor-crafter-do-slice-2-step-0-resume.md`) created `rules/do/step-0-resume.md` and applied the `{CRAFTER_HOME}` policy to its one `task-lifecycle.md` runtime reference (`~/.claude/crafter/rules/task-lifecycle.md` → `{CRAFTER_HOME}/rules/task-lifecycle.md`). The bare `task-lifecycle.md` mention in the branch-sanity guard is not a runtime path and was left unchanged. Repo-wide normalization remains owned by `.crafter/tasks/20260421-skills-first-runtime-portability.md`.

**Slice 3** (task `.crafter/tasks/20260517-refactor-crafter-do-slice-3-steps-1-2.md`) created `rules/do/step-1-scope.md` and `rules/do/step-2-discuss.md` and applied the `{CRAFTER_HOME}` policy to their runtime references: `step-1-scope.md` had two (`~/.claude/crafter/rules/do/extension-skills.md` → `{CRAFTER_HOME}/rules/do/extension-skills.md` and `~/.claude/crafter/rules/task-lifecycle.md` → `{CRAFTER_HOME}/rules/task-lifecycle.md`); `step-2-discuss.md` had one (`~/.claude/crafter/rules/task-lifecycle.md` → `{CRAFTER_HOME}/rules/task-lifecycle.md`). The bare `rules/do-workflow.md` and `crafter-analyzer` mentions are not runtime install paths and were left unchanged. Repo-wide normalization remains owned by `.crafter/tasks/20260421-skills-first-runtime-portability.md`.
