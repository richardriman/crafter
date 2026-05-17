# Task: Refactor crafter-do into composable core capability modules

## Metadata
- **Date:** 2026-05-17
- **Work branch:** feat/composable-skill-contracts
- **Status:** active
- **Scope:** Large

## Request
Create a new plan, continuing on the existing `feat/composable-skill-contracts` branch, for the next step after composable skill contracts: refactor `crafter-do` toward composable **core capability modules** rather than external extension-skill packaging.

Context from discussion:
- Extension skills live at global or project level and are installed separately; Crafter ships no default extension skills. Extension packaging/distribution is out of scope for this follow-up.
- The important follow-up is to examine how to refactor `crafter-do` itself: decompose the current monolithic workflow into internal, Crafter-distributed core capability modules while preserving behavior.
- The previous composable skill contract work may have under-accounted for the fact that Crafter now supports more than Claude Code; it also supports Copilot CLI. The plan must account for multi-runtime source/install behavior instead of assuming Claude-only paths.

Expected planning direction:
- Identify a safe decomposition strategy for `skills/crafter-do/SKILL.md` and related `rules/*` files.
- Preserve existing `crafter-do` user-facing behavior and green-commit/review/approval gates.
- Keep external extension-skill installation/distribution out of scope.
- Treat Claude Code and Copilot CLI runtime support as a first-class constraint.
- Prefer an incremental refactor plan with verification gates over a large rewrite.

## Plan

**Plan status:** draft

### 1. Complete request

Refactor `skills/crafter-do/SKILL.md` (currently ~450 lines, monolithic orchestrator) toward a composable set of **internal core capability modules** — Crafter-distributed, sanctioned, loaded by the orchestrator the same way `rules/do-workflow.md`, `rules/post-change.md`, etc. are loaded today. The goal is to reduce the monolith's cognitive load, isolate per-capability semantics so they can evolve independently, and make it cheaper to reason about each gate without scrolling 450 lines of intertwined prose. **No user-visible behavior change.** Every existing gate (flag validation, project resolution, extension-skill discovery, resume detection, Steps 0–9b, the green-commit invariant, the review-fix cap, `--auto` retained gates, the supplemental-only invariant) must read identically after the refactor.

This task delivers a **first safe slice** of the decomposition, not the full decomposition. The deliverable is:

1. A short design note (in `docs/`) that fixes the capability-module taxonomy and the runtime-path policy used by all new modules. This document is what later slices will follow.
2. A first concrete extraction of 2–3 self-contained "preamble" capabilities (the non-step sections at the top of `crafter-do`) into `rules/` fragments, with `skills/crafter-do/SKILL.md` reduced to a thin loader for those fragments at exactly the same insertion points.
3. A runtime-path hygiene pass scoped strictly to the newly created/edited files, so that multi-runtime support (Claude Code today, Copilot CLI / OpenCode tomorrow) is not made worse — and ideally made marginally easier for follow-up slices.

Why this matters: `crafter-do/SKILL.md` is the entry point that every Crafter workflow loads. A 450-line monolith with three different abstraction levels (preamble, workflow gates, post-change) is hostile to safe edits — every previous behavioral change (GH#15 `--auto`, GH#16 buffer, GH#17 PR composer, GH#18 sub-agent semantics, the composable-contracts work just merged on this branch) had to be threaded through the same single file. Extracting the natural capability seams that already exist as section headings into named modules turns the orchestrator into a manifest plus glue, which is what it conceptually already is.

Why "first safe slice" instead of a full decomposition: a full extraction of all sections is mechanically straightforward but creates a large diff that is hard to review and easy to regress. The composable-contracts task on this branch chose the same conservative path (Phase A doc-only, then Phase B small inserts) and that paid off. We do the same here: design + first slice = one PR; remaining slices = follow-up tasks gated on this slice landing cleanly.

**Acceptance criteria for this plan:**

- Names a target taxonomy of capability modules (every existing section of `crafter-do/SKILL.md` maps to exactly one named future module — even if most are deferred to follow-ups).
- Identifies the first slice explicitly and bounds it (2–3 modules at the natural preamble seam).
- Preserves user-visible behavior: no gate semantics change, no new gates, no removed gates, no contract changes.
- Treats Claude Code and Copilot CLI as first-class: source files use a runtime-neutral reference convention where it does not break existing behavior, and the installer remains the single place that knows about a specific runtime's installed-path layout.
- Defines verification gates (per step and per phase) consistent with the existing drift-check / phase-verification / review-fix-cap / green-commit invariants.
- Surfaces assumptions, alternatives considered, and risks/unknowns explicitly.

**Validation strategy:** All changes in this task are markdown (sources only — no installed-runtime edits, no Go code, no CLI subcommands, no agent edits). Validation is editorial: the post-refactor `crafter-do` reads the new modules at exactly the same insertion points, and a re-reading of the four main flows (default, `--fast`, `--auto`, resume) produces identical semantics. No automated tests are required by the refactor itself; existing `tests/test_install.sh` should still pass because the installer's deployed file set is unchanged (new module files are added to the same `rules/` deployment).

### 2. Assumptions / interpretations

- **"Core capability module" = an internal, Crafter-distributed markdown fragment** with a short, focused responsibility, loaded by a core skill the same way `rules/*.md` are loaded today. It is **not** an extension skill (those are supplemental, third-party, project-scoped). The two concepts live side by side: core capabilities are part of Crafter, extension skills are not. The composable-contracts safety envelope continues to apply to extension skills only.
- **Target location for new modules: `rules/`** — likely under a subdirectory such as `rules/capabilities/` or `rules/do/` — so the existing installer deployment to `crafter/rules/` picks them up with no installer change other than copying the new files. The Planner does not fix the exact path; the Implementer chooses inside the contract.
- **Loading convention stays prompt-driven** — `crafter-do/SKILL.md` continues to use a "Read and follow these rules" list at the top and inline references inside steps. No new mechanism, no manifest, no Go code, no CLI subcommand.
- **No content is rewritten.** Extraction is move-and-link: the prose moves to its module file, and the original section in `crafter-do/SKILL.md` is replaced with a one- to three-line pointer (sometimes a one-paragraph orientation paragraph plus the link, if the section in place provided structural context). Drive-by clarifications, reorderings, or "while-we're-here" edits are not allowed in this task.
- **Runtime path policy (new):** the new capability modules should reference the installed rules directory by a runtime-neutral phrase (e.g., "the installed Crafter rules directory" or a documented placeholder like `{CRAFTER_HOME}/rules/...`) rather than hard-coding `~/.claude/crafter/rules/...`. Existing hard-coded references in files that are **not touched by this slice** are left alone — broader normalization is a separate task (the active `20260421-skills-first-runtime-portability.md` task already owns it). This task only commits to the new files following the policy and to not regressing the situation.
- **First-slice candidates** (the orchestrator does not pre-commit to all three — Implementer picks 2 or 3 inside the contract): (a) Flag Validation (lines ~62–68), (b) Project Resolution (lines ~70–104), (c) Extension Skills (lines ~106–137). These three are top-of-file, self-contained, free of cross-references to other steps, and have a stable surface. Resume Detection (Step 0) is deliberately **not** in the first slice because it cross-references `rules/task-lifecycle.md` and the branch sanity / main-master guards — a richer surface that benefits from being done after the pattern is proven.
- **Installer behavior assumption:** `install.sh` currently copies the fixed list of `rules/*.md` files explicitly (lines 311–317). If new module files are introduced under a subdirectory, the installer needs a single change (a directory copy or expanded explicit list). This is in scope for the slice that introduces the first new module file.
- **Multi-runtime install posture:** today `install.sh` writes to `~/.claude/` only (Claude Code surface). Copilot CLI is referenced as an aspiration in tasks but not implemented in `install.sh`. This task does not add Copilot install support; it only ensures the new module files do not bake in Claude-only assumptions in their **content**. Adding Copilot/OpenCode install targets remains the responsibility of `20260421-skills-first-runtime-portability.md`.

**Competing interpretations the user may want to resolve:**

- *Interpretation A (chosen):* extract preamble sections first (flag validation, project resolution, extension skills); leave workflow steps (0–9b) intact for follow-up slices.
- *Interpretation B:* extract one workflow step (e.g., Step 6b Phase Summary, the densest single block) instead, to prove the pattern on a higher-risk surface.
- *Interpretation C:* skip extraction entirely and only deliver the design note + runtime-path policy in this task, executing the first slice in a separate follow-up.

The plan below assumes **A**. If the user prefers **B** or **C**, the Phase 2 contract changes and we re-plan that phase.

### 3. Non-goals

- **No external extension-skill packaging or distribution.** That remains explicitly out of scope per the request. The composable-contracts work just landed; this task does not touch extension-skill semantics.
- **No new gates, no removed gates, no semantic changes** to plan approval, drift check, phase verification, review fix-loop cap, green-commit invariant, `--auto` retained gates, manual-verification exception, `--fast` semantics, supplemental-only invariant.
- **No agent edits.** `agents/crafter-*.md` are out of scope.
- **No CLI subcommand additions** and no changes to the Go binary.
- **No buffer or task-file schema changes.**
- **No edits to installed runtime copies** under `~/.claude/...`, `~/.copilot/...`, etc.
- **No backfill of existing skills** (`crafter-debug`, `crafter-map-project`, `crafter-status`, `crafter-buffer`) into the new module convention. They can be refactored later if the pattern proves out.
- **No installer-level multi-runtime support added** (Copilot CLI / OpenCode install targets stay deferred to `20260421-skills-first-runtime-portability.md`).
- **No reordering of existing `crafter-do` step numbers**, no renaming of existing rules files, no changes to `docs/skill-contract.md` or `docs/plugin-system.md` content.
- **No drive-by improvements** to extracted prose. Move-and-link only.

### 4. Relevant areas

- `skills/crafter-do/SKILL.md` — primary integration site. Sections to map onto future capability modules; first-slice extractions land here as pointer references.
- `rules/do-workflow.md` — adjacent to several extraction candidates (drift check, review, `--auto` retained gates, run-directory lifecycle, scope detection); review for cross-references but **do not edit in the first slice** unless the Implementer finds a strictly mechanical link to update.
- `rules/post-change.md`, `rules/task-lifecycle.md`, `rules/delegation.md`, `rules/core.md` — existing fragments; act as reference precedents for tone, length, frontmatter style, and "Read and follow these rules" loading convention.
- `docs/skill-contract.md`, `docs/plugin-system.md` — read-only context. The new design note for *core* capabilities lives separately and explicitly distinguishes itself from the *extension* skill contract.
- `.crafter/ARCHITECTURE.md` — possibly a one-line cross-reference under conventions; do not restructure.
- `install.sh` (lines 277–339) — the `install_to()` function explicitly enumerates rule files. The first slice introduces new files under `rules/` and may need a single mechanical change here (either an added line per file or a directory copy). Verify against `tests/test_install.sh`.
- `tests/test_install.sh` — sanity check that adding files under `rules/` doesn't regress installer tests.
- `.crafter/tasks/20260421-skills-first-runtime-portability.md` — sibling active task. The runtime-path policy in this plan is consistent with what that task is meant to deliver, but does not preempt it.

### 5. Vertical phases and steps

The work is three phases. Phase 1 produces design artifacts that make the rest cheap. Phase 2 is the first safe extraction slice. Phase 3 is the runtime-path hygiene pass scoped to the touched files. Each phase ends in an independently committable, reviewable state.

#### Phase 1 — Capability taxonomy and runtime-path policy (design only)

After Phase 1 the repo contains a short written design document that (a) lists every section of `crafter-do/SKILL.md` and the named capability module it will eventually map to, (b) fixes the loading convention and naming scheme for the new module files, (c) fixes the runtime-neutral reference convention for installed paths, and (d) marks which modules are in this task's slice vs. deferred to follow-ups. No prompt edits, no extractions yet — Phase 1 is independently valuable as a contract for Phase 2 and any follow-up task.

- [ ] **1.1 — Author the capability taxonomy design note.** Produce a single new doc (location TBD by Implementer — likely `docs/core-capabilities.md` or similar) listing each existing section of `crafter-do/SKILL.md` with a one-line description and the name of its future capability module. Mark every entry as "Slice 1 (this task)", "Deferred — follow-up", or "Stays inline in crafter-do" (the top-of-file frontmatter + loader list is expected to stay inline). The doc explicitly states: this is for core, Crafter-distributed capabilities only — extension-skill compatibility is governed by `docs/skill-contract.md` and is unrelated.
  - **Karpathy contract:** outcome = one new design doc that maps the monolith onto named modules; scope = only this doc created; non-goals = no edits to other files, no implementation, no extension-skill discussion; simplicity = single file, single table; drift = expanding into an implementation script, redefining gates, blurring extension vs. core boundary; verification = every existing top-level section heading of `crafter-do/SKILL.md` appears in the table with a target module name and a slice tag; stop conditions = if any section resists naming because it spans gates, stop and surface in risks before forcing a name.

- [ ] **1.2 — Add the runtime-path policy section to the design note.** Append a short policy section to the same doc fixing the convention for referencing installed Crafter paths from inside the new module files. Define the placeholder or phrasing (the Implementer picks the concrete form within the contract), define what stays runtime-specific (the installer), and explicitly note that **existing** hard-coded `~/.claude/...` references in **untouched** files are not normalized by this task. Cross-link to the sibling task `20260421-skills-first-runtime-portability.md` as the broader owner of multi-runtime work.
  - **Karpathy contract:** outcome = a runtime-path policy section appended to the same design doc; scope = same single doc; non-goals = no edits to existing files, no installer changes yet, no Copilot/OpenCode install logic; simplicity = ≤ ~20 lines; drift = expanding into a multi-runtime adapter design, redefining install layout, contradicting `20260421-skills-first-runtime-portability.md`; verification = the policy names exactly one canonical reference form for new module files, names the installer as the sole runtime-specific surface, and cross-links the sibling task; stop conditions = if specifying the policy concretely requires a new installer feature, stop — that is out of scope for this task.

- [ ] **Phase 1 verification.** The design doc exists, lists every `crafter-do` section with a target module name and a slice tag, contains the runtime-path policy section, and resolves cross-links. No other files modified. A reviewer reading just this doc should be able to predict what Phase 2 will touch.
- [ ] **Phase 1 review.** Standard `crafter-reviewer` pass; Karpathy scorecard PASS on Simplicity / Surgical Changes / Goal-Driven.

#### Phase 2 — First extraction slice (preamble capabilities)

After Phase 2 the preamble of `crafter-do/SKILL.md` (the sections above Step 0, i.e., Flag Validation, Project Resolution, Extension Skills) is reduced to short pointer references to new capability modules under `rules/`. The original prose lives unchanged inside the new module files. Behavior identical. Step 0 onward untouched.

- [ ] **2.1 — Extract Flag Validation.** Move the "Flag Validation (before anything else)" section verbatim into a new capability module file under `rules/` (exact name and subdirectory chosen by Implementer per the taxonomy from 1.1). Replace the section in `crafter-do/SKILL.md` with a short pointer ("Apply the flag-validation procedure in `<path>`."). Verify install.sh deploys the new file — if a new explicit entry or a directory-copy is needed in `install_to()`, add it in this step.
  - **Karpathy contract:** outcome = one moved section + one new module file + (if needed) one mechanical installer edit; scope = exactly this section, this module file, and the installer; non-goals = touching other sections, rewriting prose, adding new gates, normalizing all `~/.claude/...` references repo-wide; simplicity = pure move; drift = reordering content, rewording, "while-we're-here" cleanup, splitting the section across multiple files; verification = `crafter-do/SKILL.md` retains a pointer at the same line range; the new module file contains the full original prose; the mutual-exclusion error message is byte-identical; `tests/test_install.sh` passes; stop = if extraction forces a semantic change to the mutual-exclusion rule, stop.

- [ ] **2.2 — Extract Project Resolution.** Same pattern: move "Project Resolution (before anything else)" into its capability module, leave a pointer in place. The `PROJECT_PATH` / `CRAFTER_DIR` resolution semantics, the `.crafter` → `.planning` legacy fallback, the migration offer, and the `--project` flag handling all stay byte-identical.
  - **Karpathy contract:** outcome = one moved section + one new module file; scope = exactly this section and module; non-goals = changing resolution order, changing the migration prompt, normalizing all path references repo-wide, touching the legacy `.planning` migration; simplicity = pure move; drift = altering the user-facing strings ("Found project in ...", "Tip: ..."), reordering bullets, splitting; verification = `crafter-do/SKILL.md` retains a pointer; the user-facing strings appear verbatim in the new file; cross-references in the same file (e.g., the line that says "Use `{PROJECT_PATH}/{CRAFTER_DIR}` as the base ...") still resolve after the move; stop = if the section cannot be cleanly excised because another step quotes the resolution rule, stop and re-plan.

- [ ] **2.3 — Extract Extension Skills.** Move the "Extension Skills" section into its capability module. This section is the bridge between core and supplemental-only; the module's introductory paragraph must keep the explicit distinction (core capability modules ≠ extension skills) so future readers don't confuse the two. The discovery priority table (project / parent / global) and the supplemental-only invariant stay verbatim.
  - **Karpathy contract:** outcome = one moved section + one new module file; scope = exactly this section and module; non-goals = touching the supplemental-only invariant in `rules/do-workflow.md`, rewriting `docs/skill-contract.md`, changing discovery semantics; simplicity = pure move plus a short opening paragraph distinguishing core vs. extension; drift = redefining what an extension skill is, adding override/replace language, adding new discovery locations; verification = the discovery table reads identically; cross-references in Steps 1, 4, 6 of `crafter-do/SKILL.md` ("see `## Extension Skills`") are updated to point to the new module; stop = if changing the cross-reference style would cascade to other unrelated sections, stop and discuss.

- [ ] **2.4 — Cross-reference sweep.** Read the post-extraction `crafter-do/SKILL.md` end-to-end and verify every remaining reference to the three extracted sections now points to the correct new module file. Verify the rules-loader list at the top of `crafter-do/SKILL.md` includes the new module paths (or the installer's deployment covers them via directory copy — choose one consistent pattern). No other content edits.
  - **Karpathy contract:** outcome = consistent cross-references after extractions; scope = wording fixes only; non-goals = adding new content, restructuring, normalizing unrelated paths; simplicity = touch only what is inconsistent; drift = improvements unrelated to consistency, reordering steps, renaming files; verification = a fresh read of `crafter-do/SKILL.md` produces no dangling internal references; stop = if a real semantic gap surfaces (not just wording), record and replan rather than fix silently.

- [ ] **Phase 2 verification.** Re-read `crafter-do/SKILL.md` confirms: the three preamble sections are now pointer references; no other sections were touched; all cross-references resolve; the rules-loader list (or installer deployment) covers the new module files. `tests/test_install.sh` passes. A semantic diff of behavior against the pre-refactor file produces zero behavioral deltas across the four flows (default, `--fast`, `--auto`, resume).
- [ ] **Phase 2 review.** Standard `crafter-reviewer` pass; Karpathy scorecard PASS on Simplicity / Surgical Changes / Goal-Driven. Reviewer must confirm move-and-link discipline (no rewording).

#### Phase 3 — Runtime-path hygiene scoped to touched files

After Phase 3, the new capability modules introduced in Phase 2 follow the runtime-path policy from Phase 1.2. Files not touched by Phase 2 are not normalized. This phase is intentionally narrow — it's the proof-of-policy step, not a repo-wide sweep.

- [ ] **3.1 — Apply the runtime-path policy to the new module files.** In the capability modules created in Phase 2, replace any `~/.claude/crafter/...` references (introduced because the original prose used them) with the runtime-neutral reference convention defined in Phase 1.2. Do **not** edit existing files outside the Phase 2 set. If the resulting modules need a one-paragraph "how to resolve the installed rules directory" footnote, add it once in the policy-establishing module rather than repeating it.
  - **Karpathy contract:** outcome = the three new module files conform to the runtime-path policy; scope = only the new module files from Phase 2 + at most a one-paragraph footnote; non-goals = repo-wide normalization, editing `rules/post-change.md` or `rules/delegation.md` or other existing files, adding installer logic for Copilot, modifying `crafter-do/SKILL.md` pointer lines (their paths already point to the new module locations); simplicity = a small set of string substitutions plus optionally one short footnote; drift = expanding to normalize other files, adding install-time path detection logic, introducing new placeholders beyond what 1.2 defined; verification = the new module files contain zero hard-coded `~/.claude/...` references; the footnote (if added) appears exactly once; pointer lines in `crafter-do/SKILL.md` still resolve under the existing Claude install layout; stop = if applying the policy requires installer changes, stop — that escalates scope.

- [ ] **3.2 — Confirm sibling task delineation.** Add a single short note to the design doc from Phase 1 (and optionally a one-line cross-reference inside the active `.crafter/tasks/20260421-skills-first-runtime-portability.md` Outcome section if appropriate) acknowledging that this task established the policy and applied it to N module files, and the sibling task remains the owner of the broader normalization. This is a small administrative edit to keep the two tasks coherent.
  - **Karpathy contract:** outcome = clear delineation between this task's scope and the sibling task's scope; scope = at most two files touched (design doc + optionally the sibling task file); non-goals = restructuring either task, restating the sibling task's plan, claiming responsibility for repo-wide normalization; simplicity = ≤ ~5 lines added total; drift = expanding into a meta-task about runtime portability; verification = the cross-reference resolves and both tasks read consistently; stop = if the sibling task file edit would interfere with its active plan, skip the sibling edit and keep only the design-doc note.

- [ ] **Phase 3 verification.** New module files contain zero hard-coded `~/.claude/...` references; design doc reflects what was actually shipped; sibling task delineation is clear. Touched files outside Phase 2's set: at most the design doc (always) and optionally the sibling task file (Outcome section only).
- [ ] **Phase 3 review.** Standard `crafter-reviewer` pass; Karpathy scorecard PASS on Surgical Changes (no scope creep into untouched files) and Goal-Driven (policy from 1.2 applied exactly once).

#### Post-task housekeeping

- [ ] **STATE.md and skillbook update** per `rules/post-change.md` — consolidated end-of-task commit.
- [ ] **Task file completion** per `rules/task-lifecycle.md`.
- [ ] **Follow-up tasks identified** — list the remaining capability modules (Step 0, Steps 1–9b) from the Phase 1 taxonomy that were deferred, as candidates for the next slice. Do not create those task files in this task — just record the list in `## Outcome` so the next planning session can pick them up.

### 6. Karpathy Contract — overall

- **Outcome:** A short design note plus a first concrete slice that extracts the preamble of `crafter-do/SKILL.md` into 2–3 named capability modules under `rules/`, with byte-identical behavior and a runtime-path policy that the rest of the decomposition can follow.
- **Scope boundary:** source files only — `docs/` (one new design doc), `rules/` (2–3 new module files, possibly under a new subdirectory), `skills/crafter-do/SKILL.md` (pointer replacements for the extracted sections only), `install.sh` (at most one mechanical edit to deploy the new files), `.crafter/ARCHITECTURE.md` (optional one-liner). No agent edits, no CLI changes, no installed-runtime edits, no buffer/task-file schema changes.
- **Non-goals:** extension-skill packaging/distribution, override/replace semantics, full decomposition of all `crafter-do` sections, multi-runtime installer support, edits to `crafter-debug`/`crafter-map-project`/`crafter-status`/`crafter-buffer`, agent edits, repo-wide path normalization, prose rewriting during extraction.
- **Simplicity constraint:** one design doc, 2–3 new module files, ≤ 3 pointer replacements in `crafter-do/SKILL.md`, at most one mechanical installer edit, at most one architecture-doc cross-link. Move-and-link discipline only. No reorderings, no renames of existing files, no semantic changes.
- **Drift criteria:** any of these signal drift — adding new gates or removing existing ones; rewording prose during extraction (any non-mechanical edit); editing files under `~/.claude/...` or other installed runtime paths; adding CLI subcommands or Go code; extending extension-skill semantics; backfilling other `crafter-*` skills; expanding the runtime-path policy into a multi-runtime adapter design; touching `agents/*.md`; touching the buffer or task-file schema; normalizing `~/.claude/...` references outside the Phase 2 file set.
- **Verification evidence:** every existing gate of `crafter-do` (default, `--fast`, `--auto`, resume) reads identically post-refactor; the design doc enumerates every section of the pre-refactor file with a target module name; the three extracted sections appear verbatim in their module files; `tests/test_install.sh` passes; reviewer Karpathy scorecard PASS on Simplicity / Surgical Changes / Goal-Driven.
- **Stop conditions:** stop and re-plan if (a) any extraction forces a semantic change to a gate, (b) extraction reveals a hidden cross-reference that requires editing a file outside the Phase 2 set in a non-mechanical way, (c) the runtime-path policy cannot be applied without installer changes, (d) a reviewer flags ambiguity between "core capability module" and "extension skill", or (e) the user prefers Interpretation B or C from Section 2.

### 7. Alternatives considered

- **Full decomposition in one task** (extract every section of `crafter-do/SKILL.md` into capability modules in one slice) — rejected. The diff would be huge, the regression surface large, and review would be hostile. The composable-contracts task on this branch chose conservative slicing for the same reason and it worked. Better to prove the pattern on a small, low-risk surface and let follow-up tasks repeat it.
- **Extract a workflow step instead of the preamble** (e.g., Step 6b Phase Summary, the densest single gate) — viable but riskier as a *first* slice. Step 6b depends on flag state, manual-verification exception phrasing, and the review fix-loop close condition — three semantic surfaces in one extraction. The preamble (flag validation, project resolution, extension skills) is mechanically simpler and almost dependency-free. Recorded as Interpretation B if the user prefers.
- **Design note only, no extraction** — viable (Interpretation C) and the safest possible task. Rejected as the *recommended* path because it leaves the monolith intact; the value of the design doc is much higher when paired with one concrete proof-of-pattern extraction.
- **Introduce a new "core skill contract" alongside the existing extension skill contract** — rejected for now. Core capability modules are internal and Crafter-distributed; their contract is the design note from Phase 1, not a public spec. Conflating the two would re-open semantic boundaries the composable-contracts task just clarified.
- **Add a frontmatter schema or YAML manifest for capability modules** — rejected for v1. The existing `rules/*.md` files have no manifest and the orchestrator loads them by name. Same convention works here. A manifest can be added later if discovery becomes lossy.
- **Repo-wide normalization of `~/.claude/...` references in this task** — rejected. That is the sibling task `20260421-skills-first-runtime-portability.md`'s responsibility; doing it here would inflate scope and conflict with that task's plan.
- **Adding Copilot CLI install support to `install.sh` as part of this task** — rejected. The user named multi-runtime as a constraint, not a feature to add here. The constraint is satisfied by ensuring the new module files don't bake in Claude-only assumptions; actually shipping a Copilot install target is the sibling task's deliverable.

### 8. Risks / unknowns / flags

- **R1. Cross-references hidden inside the extracted sections.** Project Resolution sets `PROJECT_PATH` and `CRAFTER_DIR` used throughout the file. Extraction must not break those bindings — the orchestrator must still treat the variables as established by the time Step 0 starts. **Mitigation:** the pointer left in `crafter-do/SKILL.md` must explicitly say "Apply this procedure and set `PROJECT_PATH` and `CRAFTER_DIR` before continuing", so the binding contract survives the move. **Flag for user:** if a reviewer judges that referencing-by-pointer weakens this contract, we may need to keep the procedure inline and only extract Flag Validation + Extension Skills in this slice.
- **R2. Installer file enumeration vs. directory copy.** `install_to()` lists rule files explicitly. Introducing new files under a new `rules/` subdirectory may push us toward a `cp -R` of `rules/`. That is a small but real change in installer style. **Flag for user:** approve switching to directory copy if needed, or require the Implementer to keep explicit per-file lines.
- **R3. Multi-runtime aspiration vs. current reality.** The repo says Crafter "supports Copilot CLI" in task descriptions, but `install.sh` only writes to `~/.claude/`. The runtime-path policy in Phase 1.2 is therefore forward-looking — it does not get exercised by any current runtime other than Claude. **Flag for user:** confirm a forward-looking policy is acceptable now, or defer Phase 1.2 + Phase 3 entirely until the sibling task lands Copilot install support.
- **R4. Test surface.** `tests/test_install.sh` validates the installer's deployed file set. Adding new module files requires updating either the test's expected file list or its globbing logic. **Mitigation:** the Implementer must run the test after the Phase 2 installer edit and adjust the test in the same step if needed. **Flag for user:** if the test is fragile, this may surface as a small additional installer-side change.
- **R5. Reviewer drift on "move-and-link discipline".** "Pure move" is easy to violate by reflex (a one-word clarification here, a list-ordering tweak there). **Mitigation:** Reviewer must explicitly diff each extracted block against the original prose and flag any non-mechanical change.
- **R6. Interpretation choice.** The plan assumes Interpretation A (preamble extraction). If the user prefers B (a workflow-step first) or C (design-only), Phase 2 changes shape. **Flag for user:** confirm Interpretation A before approval, or signal B/C and we revise.
- **R7. Naming the new subdirectory.** "Capability module" is a deliberately neutral term. The actual subdirectory name (e.g., `rules/capabilities/`, `rules/do/`, `rules/modules/`) is left to the Implementer inside Phase 2.1's contract; if the user has a strong preference, name it now to remove that micro-decision from execution.

---

This contract protects the existing user-visible behavior of `crafter-do` (every gate, every flag, every retained-gate semantic, every approval path) while creating the structural foundation — a short design doc plus one proven extraction slice — that makes the rest of the decomposition cheap to land in follow-up tasks. The approach is the right one because it mirrors the conservative slicing that worked on the just-merged composable-contracts task on the same branch, treats Claude Code and Copilot CLI as first-class without overreaching into the sibling runtime-portability task, and keeps every change inside source files only with no installed-runtime edits and no behavior change.

## Decisions
- **Decision:** Continue this planning task on branch `feat/composable-skill-contracts`. **Reason:** The work is a direct follow-up to the composable skill contracts branch and the user explicitly requested continuation on the existing branch.
- **Decision:** Keep extension-skill packaging/distribution out of scope. **Reason:** Extension skills are expected to live at global or project level and be installed separately; Crafter does not ship default extensions today.

## Outcome
<!-- Filled on completion: what was actually done, commit SHA(s), any deviations from plan -->
