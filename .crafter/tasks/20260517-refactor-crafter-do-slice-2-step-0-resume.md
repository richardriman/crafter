# Task: Refactor crafter-do Slice 2 — extract Step 0 Resume Detection into a core capability module

## Metadata
- **Date:** 2026-05-17
- **Work branch:** refactor/crafter-do-slice-2-step-0
- **Status:** active
- **Scope:** Medium

## Request
Slice 2 of the crafter-do core-capability decomposition. Extract the `## Step 0 — Resume Detection` section of `skills/crafter-do/SKILL.md` into a new internal core capability module `rules/do/step-0-resume.md`, following exactly the move-and-link pattern, taxonomy, naming scheme, loading convention, and `{CRAFTER_HOME}` runtime-path policy established in `docs/core-capabilities.md` and proven in Slice 1 (task `.crafter/tasks/20260517-refactor-crafter-do-core-capabilities.md`).

Constraints:
- Behavior must be byte-identical across all four flows (default, `--fast`, `--auto`, resume).
- Validate that the `task-lifecycle.md` reference, the branch-sanity guard, and the main/master guard cross-references survive the pointer before extracting, per the deferred-slice note in `docs/core-capabilities.md` (Step 0 cross-references `rules/task-lifecycle.md`).
- Pure move-and-link only: no prose rewriting, no reordering, no other sections touched.
- Out of scope: any other Step extraction, extension-skill semantics, agent edits, CLI/Go changes, repo-wide path normalization.

## Plan

**Plan status:** approved

### 1. Complete request

Extract the `## Step 0 — Resume Detection` section of `skills/crafter-do/SKILL.md` into a new internal core capability module `rules/do/step-0-resume.md`, applying the exact move-and-link pattern, taxonomy, naming scheme, loading convention, and `{CRAFTER_HOME}` runtime-path policy established in `docs/core-capabilities.md` and proven in Slice 1 (`.crafter/tasks/20260517-refactor-crafter-do-core-capabilities.md`). This is Slice 2 of the deferred follow-up sequence; Step 0 was deliberately deferred from Slice 1 because it cross-references `rules/task-lifecycle.md` and carries two mandatory guards (branch-sanity, main/master) whose bindings must be validated to survive the pointer before extracting.

The Step 0 prose moves verbatim into the new module. The `## Step 0 — Resume Detection` section in `SKILL.md` is reduced to a thin pointer stub that (a) keeps the exact `## Step 0 — Resume Detection` heading so every existing "Step 0" reference elsewhere in the file still resolves, and (b) preserves the binding contracts the prose carried (resume state and branch/main-master guards are established before Step 1 begins). The loader list gains one entry under the existing `<!-- do/* capability modules -->` group.

Why this matters: `crafter-do/SKILL.md` is the entry point every Crafter workflow loads. Step 0 is the highest-value next isolation per the design note's recommended deferred-slice order — it gates whether the workflow resumes an active task or starts fresh, and it owns the branch-safety guards. Pulling it into a named module shrinks the monolith and lets resume/guard semantics evolve in one focused file, exactly as Slice 1 did for the preamble.

**Acceptance criteria:**

- A new module `rules/do/step-0-resume.md` exists containing the full Step 0 prose verbatim (single `# Title` H1, no carried-over `## Step 0 — Resume Detection` H2 — same convention as the three Slice 1 modules), with the one `~/.claude/crafter/rules/task-lifecycle.md` reference normalized to `{CRAFTER_HOME}/rules/task-lifecycle.md` per the runtime-path policy.
- `skills/crafter-do/SKILL.md` keeps the `## Step 0 — Resume Detection` heading followed by a short pointer that preserves the binding contracts (resume routing + both guards must be applied before Step 1).
- The loader list in `SKILL.md` includes `~/.claude/crafter/rules/do/step-0-resume.md` under the existing `<!-- do/* capability modules -->` group (loader-list/pointer paths intentionally stay `~/.claude/...`, same carve-out as Slice 1).
- `install.sh` deploys the new module and `tests/test_install.sh` expects it; `tests/test_install.sh` passes.
- Behavior byte-identical across all four flows (default, `--fast`, `--auto`, resume). Every inbound reference to "Step 0" and every forward reference from Step 0 (to Steps 1/3/4 and to `task-lifecycle.md`) still resolves.
- Two independently committable phases, each with a per-phase verification gate and a `crafter-reviewer` gate; Karpathy scorecard PASS on Simplicity / Surgical Changes / Goal-Driven.

**Validation strategy:** Editorial + installer test. Re-reading the post-extraction `SKILL.md` for the four flows produces identical semantics; a literal diff of the moved prose against the original block shows only the mechanical `{CRAFTER_HOME}` substitution and the H2→H1 normalization; `tests/test_install.sh` passes after the installer + test edits.

### 2. Assumptions / interpretations

- **Module path is fixed by the design note:** `rules/do/step-0-resume.md` (taxonomy table row for Step 0). The `rules/do/` subdirectory was confirmed in Slice 1; this slice follows it without re-deciding.
- **Module structure mirrors Slice 1 modules:** single `# Title` H1 (e.g. `# Step 0 — Resume Detection` or `# Resume Detection`), no carried-over H2, prose otherwise verbatim. The exact H1 wording is the Implementer's choice within the move-and-link contract; it must not change meaning.
- **`{CRAFTER_HOME}` applies to one line.** Step 0's only concrete runtime path is `~/.claude/crafter/rules/task-lifecycle.md` (appears once as a "Follow the resume detection procedure in ..." reference; a second mention is the bare `task-lifecycle.md` in the branch-sanity guard, which is not a runtime path and stays as-is). That one full path becomes `{CRAFTER_HOME}/rules/task-lifecycle.md` in the module. SKILL.md loader-list and pointer lines stay `~/.claude/...` (installer-resolved) — identical carve-out to Slice 1.
- **Installer change IS required (deviation from Slice 1's "likely none" framing).** `install.sh` enumerates each `rules/do/*.md` file explicitly (lines 321–323), and `tests/test_install.sh` enumerates each expected file (lines 295–297). A new module therefore needs exactly one mechanical `cp` line in `install.sh` and one expected-path entry in `tests/test_install.sh`. The `mkdir -p "$rules_dest/do"` already exists from Slice 1 — no directory-creation change needed.
- **Lower-drift cross-reference strategy (chosen):** keep the `## Step 0 — Resume Detection` heading in SKILL.md as a thin pointer stub rather than deleting the heading and rewriting every referrer. This preserves all existing "Step 0" references (Step 6a's "The resume detection in Step 0 will pick up...", Step 1's resume/continue flow) with zero edits to other sections — strictly lower drift than renaming/retargeting referrers. Justified below in §7.

**Competing interpretations:**

- *Interpretation A (chosen):* thin pointer stub under the retained `## Step 0 — Resume Detection` heading; referrers untouched.
- *Interpretation B:* delete the Step 0 heading, move all content out, and update every "Step 0" referrer to point at the module. Rejected — higher drift, touches sections outside the extraction target, risks weakening the "Step 0 runs before Step 1" ordering contract.

If the user prefers B, Phase 1 changes shape and the cross-reference sweep expands.

### 3. Non-goals

- No other Step extraction (Steps 1–9b stay inline; this is Slice 2 only).
- No prose rewriting, reordering, or "while-we're-here" edits. Move-and-link only.
- No edits to `rules/task-lifecycle.md` (the resume procedure it owns is unchanged; only the *reference path* to it is normalized inside the new module).
- No changes to extension-skill semantics, `agents/*.md`, CLI/Go code, buffer or task-file schema.
- No edits to any SKILL.md section other than: the Step 0 stub, the loader list, and (cross-reference sweep) any genuinely dangling internal reference — none expected under Interpretation A.
- No repo-wide `~/.claude/...` normalization; only the one `task-lifecycle.md` path inside the new module (sibling task `20260421-skills-first-runtime-portability.md` owns the broader sweep).
- No edits to installed runtime copies under `~/.claude/...`.
- No reordering or renaming of existing Step numbers or rules files.

### 4. Relevant areas

- `skills/crafter-do/SKILL.md` — extraction target is lines ~93–110 (`## Step 0 — Resume Detection` through the main/master guard). Loader list lines 16–19 (`<!-- do/* capability modules -->` group). Inbound "Step 0" references: Step 1 (lines ~112, 126 — resume/continue flow) and Step 6a (line ~310, "The resume detection in Step 0 will pick up..."). Forward references from Step 0: Steps 1/3/4 and `task-lifecycle.md`.
- `rules/do/step-0-resume.md` — new file (does not exist yet).
- `rules/do/flag-validation.md`, `rules/do/project-resolution.md`, `rules/do/extension-skills.md` — precedent for module structure (single H1, tone, `{CRAFTER_HOME}` usage).
- `docs/core-capabilities.md` — governing design note: taxonomy row for Step 0, naming scheme, loading convention, runtime-path policy, deferred-slice ordering note (Step 0 cross-ref validation requirement).
- `install.sh` lines 320–323 — `rules/do/` deployment block; one `cp` line to add.
- `tests/test_install.sh` lines 295–297 — expected `rules/do/` files; one entry to add.
- `.crafter/tasks/20260517-refactor-crafter-do-core-capabilities.md` — Slice 1 precedent: Phase 2 (extraction) / Phase 3 (runtime hygiene) structure, Karpathy contracts, `## Decisions` conventions (loader grouping comments, redundant-H2 removal, R1 binding-preservation pattern).

### 5. Vertical phases and steps

Two phases, mirroring Slice 1's conservative shape. Phase 1 = the extraction (module + pointer + loader entry + installer/test wiring + cross-reference sweep) — independently committable and behavior-identical. Phase 2 = runtime-path hygiene scoped to the new module + bookkeeping. Each phase ends in a reviewable state with a `crafter-reviewer` gate.

#### Phase 1 — Extract Step 0 into a capability module (move-and-link)

After Phase 1, `## Step 0 — Resume Detection` in `SKILL.md` is a thin pointer stub under its original heading; the full Step 0 prose lives verbatim in `rules/do/step-0-resume.md`; the loader list and installer/test deploy the new module; all cross-references resolve. Behavior identical across all four flows.

- [x] **1.1 — Validate Step 0 bindings survive a pointer (pre-extraction check).** — GO; all 5 bindings PRESERVED, no verbatim quotes elsewhere, Interpretation A confirmed. Before moving anything, confirm in writing (in the step's working notes / drift evidence) that: (a) the `task-lifecycle.md` resume reference, (b) the resume-intent thoroughness rule, (c) the plan-status routing branches (pending → Step 1, draft → Step 3, approved → Step 4), (d) the branch-sanity guard, and (e) the main/master guard can all be preserved by a pointer that keeps the `## Step 0 — Resume Detection` heading and an explicit "establish resume state and apply both guards before Step 1" sentence. If any binding cannot survive a pointer (e.g., another section quotes Step 0 prose verbatim rather than referencing it by name), stop and surface as a risk before extracting.
  - **Karpathy contract:** outcome = a written go/no-go on pointer-preservability of all five Step 0 bindings; scope = analysis only, zero file edits; non-goals = moving prose, editing referrers; simplicity = a short evidence note, not a redesign; drift = beginning the extraction before the check is recorded, or expanding into a referrer rewrite; verification = each of the five bindings has an explicit preserved/at-risk verdict citing the inbound/forward reference it depends on; stop = any binding judged not pointer-preservable → halt, record risk, do not extract.

- [x] **1.2 — Create `rules/do/step-0-resume.md` with verbatim Step 0 prose.** Move the full `## Step 0 — Resume Detection` body (from the `task-lifecycle.md` reference through the main/master guard) into the new module. Single `# Title` H1, no carried-over `## Step 0 — Resume Detection` H2 (same convention as the three Slice 1 modules and Slice 1 Decision "redundant carried-over H2 removed"). Prose otherwise byte-identical — the `{CRAFTER_HOME}` substitution is deferred to Phase 2, so this step copies the `~/.claude/crafter/rules/task-lifecycle.md` reference exactly as-is.
  - **Karpathy contract:** outcome = one new module file containing the full original Step 0 prose; scope = exactly this file; non-goals = rewording, reordering, splitting, normalizing paths (deferred to Phase 2), touching SKILL.md yet; simplicity = pure move + H2→H1; drift = any non-mechanical change to the resume routing, the branch-sanity guard, or the main/master guard wording; verification = a literal diff of the module body against the original SKILL.md block shows only the H2→H1 line change and no other differences; stop = if the section cannot be cleanly excised because another section quotes its prose verbatim, stop and re-plan.

- [x] **1.3 — Replace the SKILL.md Step 0 section with a binding-preserving pointer stub.** Keep the `## Step 0 — Resume Detection` heading. Replace the body with a short pointer (Slice 1 R1 pattern): point to `~/.claude/crafter/rules/do/step-0-resume.md` and explicitly state that the procedure establishes resume state and applies the branch-sanity and main/master guards before Step 1 continues — so the "Step 0 runs before Step 1" ordering contract and the guard bindings survive the move. Pointer path stays `~/.claude/...` (installer-resolved carve-out, same as Slice 1).
  - **Karpathy contract:** outcome = the Step 0 section is a thin stub under its original heading, preserving all bindings; scope = only the Step 0 section of SKILL.md; non-goals = touching any other section, changing the heading text, rewording referrers; simplicity = a 2–4 line pointer mirroring the Slice 1 preamble pointers; drift = dropping the "before Step 1 / guards apply" binding language, altering the heading, editing adjacent sections; verification = the heading is unchanged; the stub explicitly names resume routing + both guards as established before Step 1; no other SKILL.md section diffs; stop = if a pointer cannot carry the ordering/guard binding without rewording a referrer, stop and surface (Interpretation B territory).

- [x] **1.4 — Wire the loader list, installer, and install test.** Add `~/.claude/crafter/rules/do/step-0-resume.md` to the SKILL.md loader list under the existing `<!-- do/* capability modules -->` group (after the three Slice 1 entries, consistent with Slice 1's resolved grouping convention). Add one `cp` line for the new module in `install.sh`'s `rules/do/` block (the `mkdir -p "$rules_dest/do"` already exists). Add one expected-path entry (`crafter/rules/do/step-0-resume.md`) to `tests/test_install.sh`. Run `tests/test_install.sh`.
  - **Karpathy contract:** outcome = the new module is loaded by SKILL.md and deployed+tested by the installer; scope = loader list (one line), `install.sh` (one cp line), `tests/test_install.sh` (one entry); non-goals = switching the installer to a directory copy, reordering existing entries, renaming files; simplicity = exactly three one-line additions, each mirroring the adjacent Slice 1 entry; drift = restructuring the loader list or installer enumeration, normalizing other entries; verification = loader entry sits in the `do/*` group; `tests/test_install.sh` passes; stop = if the test framework cannot accommodate one added file without broader changes, stop and flag (risk R4).

- [x] **1.5 — Cross-reference sweep.** Read the post-extraction `SKILL.md` end-to-end. Confirm every inbound "Step 0" reference still resolves to the retained heading (Step 6a's "The resume detection in Step 0 will pick up...", Step 1's resume/continue flow) and every forward reference from the now-stubbed Step 0 / new module (to Steps 1/3/4 and `task-lifecycle.md`) still resolves. No content edits beyond what 1.3 produced — under Interpretation A no referrer should need changing.
  - **Karpathy contract:** outcome = confirmed zero dangling internal references after extraction; scope = read-only verification + only consistency fixes if a genuine dangle exists; non-goals = improvements, restructuring, renaming, retargeting referrers that already resolve; simplicity = touch nothing if nothing dangles; drift = "improving" referrers that already work, expanding the stub; verification = a fresh read produces no broken "Step 0" or forward references; stop = if a real semantic gap surfaces (not cosmetic), record and replan rather than silently patch.

- [x] **Phase 1 verification.** — crafter-verifier 7/7 PASS; test 45/0. `rules/do/step-0-resume.md` contains the full Step 0 prose verbatim (only H2→H1 differs); SKILL.md's `## Step 0 — Resume Detection` heading is retained with a binding-preserving pointer stub; loader list includes the new module in the `do/*` group; `install.sh` + `tests/test_install.sh` updated and the test passes; all inbound/forward Step 0 references resolve; no SKILL.md section other than the Step 0 stub and loader list changed; semantic behavior identical across default / `--fast` / `--auto` / resume.
- [x] **Phase 1 review.** — crafter-reviewer no findings; Karpathy scorecard all PASS. `crafter-reviewer` pass; Karpathy scorecard PASS on Simplicity / Surgical Changes / Goal-Driven. Reviewer must literally diff the moved prose against the original block and confirm move-and-link discipline (no rewording, only H2→H1).

#### Phase 2 — Runtime-path hygiene + bookkeeping

After Phase 2 the new module follows the `{CRAFTER_HOME}` runtime-path policy and the design-note ledger reflects what shipped. Narrow, proof-of-policy scope — no repo-wide sweep.

- [ ] **2.1 — Apply the `{CRAFTER_HOME}` policy to the new module.** In `rules/do/step-0-resume.md`, replace the single concrete runtime path `~/.claude/crafter/rules/task-lifecycle.md` with `{CRAFTER_HOME}/rules/task-lifecycle.md` per the policy in `docs/core-capabilities.md`. The bare `task-lifecycle.md` mention in the branch-sanity guard is not a runtime path and stays unchanged. No other files touched; SKILL.md pointer/loader lines keep `~/.claude/...` (installer-resolved carve-out).
  - **Karpathy contract:** outcome = the new module has zero hard-coded `~/.claude/...` references; scope = only `rules/do/step-0-resume.md`; non-goals = repo-wide normalization, editing SKILL.md, editing `rules/task-lifecycle.md`, adding installer path logic, adding a footnote (Slice 1 needed none for an analogous single substitution); simplicity = exactly one string substitution; drift = normalizing other files, introducing new placeholders, rewording surrounding prose; verification = the module contains zero `~/.claude/...` references and exactly one `{CRAFTER_HOME}/rules/task-lifecycle.md`; the bare `task-lifecycle.md` guard mention is unchanged; stop = if applying the policy would require installer changes, stop (scope escalation).

- [ ] **2.2 — Update the design-note ledger and confirm sibling delineation.** In `docs/core-capabilities.md`, update the "Applied in this task" / runtime-path ledger to record that Slice 2 created `rules/do/step-0-resume.md` and applied the `{CRAFTER_HOME}` policy to its one `task-lifecycle.md` reference, restating that repo-wide normalization remains owned by `20260421-skills-first-runtime-portability.md`. Keep it to a few lines; do not restructure the doc.
  - **Karpathy contract:** outcome = the design note accurately reflects Slice 2's shipped scope; scope = a few lines in `docs/core-capabilities.md` (and optionally one cross-reference line if the doc's structure clearly invites it); non-goals = restructuring the taxonomy, restating the sibling task's plan, claiming repo-wide normalization, re-tagging other taxonomy rows; simplicity = ≤ ~5 added/edited lines; drift = expanding into a meta runtime-portability narrative; verification = the ledger names Slice 2, the new module, the one substitution, and the sibling-task owner; stop = if the edit would contradict the existing Slice 1 ledger entry, stop and reconcile explicitly.

- [ ] **Phase 2 verification.** `rules/do/step-0-resume.md` has zero hard-coded `~/.claude/...` references and exactly one `{CRAFTER_HOME}/rules/task-lifecycle.md`; the branch-sanity guard's bare `task-lifecycle.md` is unchanged; `docs/core-capabilities.md` ledger reflects Slice 2; no files outside this phase's set touched; behavior still byte-identical.
- [ ] **Phase 2 review.** `crafter-reviewer` pass; Karpathy scorecard PASS on Surgical Changes (no scope creep into untouched files) and Goal-Driven (policy applied exactly once).

#### Post-task housekeeping

- [ ] **STATE.md and skillbook update** per `rules/post-change.md` — consolidated end-of-task commit.
- [ ] **Task file completion** per `rules/task-lifecycle.md` (Status → completed, `## Outcome` filled, steps checked).
- [ ] **Follow-up note** — record in `## Outcome` that Slice 3 (Steps 1–2 per the design-note order) remains the next candidate.

### 6. Karpathy Contract — overall

- **Outcome:** Step 0 Resume Detection extracted into `rules/do/step-0-resume.md` with byte-identical behavior, a binding-preserving pointer stub under the retained heading, loader/installer/test wiring, and the `{CRAFTER_HOME}` policy applied to the module's single runtime path.
- **Scope boundary:** source files only — `rules/do/step-0-resume.md` (new), `skills/crafter-do/SKILL.md` (Step 0 stub + one loader line), `install.sh` (one cp line), `tests/test_install.sh` (one entry), `docs/core-capabilities.md` (ledger lines). No agent edits, no CLI/Go, no installed-runtime edits, no schema changes, no edits to `rules/task-lifecycle.md`.
- **Non-goals:** other Step extractions, prose rewriting, reordering, referrer rewrites, repo-wide path normalization, extension-skill changes, installer style change to directory copy.
- **Simplicity constraint:** one new module, one stub replacement, one loader line, one cp line, one test entry, one `{CRAFTER_HOME}` substitution, ≤ ~5 ledger lines. Move-and-link discipline only.
- **Drift criteria:** any reword/reorder/split of Step 0 prose; dropping the "before Step 1 / guards apply" binding from the stub; altering the `## Step 0 — Resume Detection` heading; editing any other SKILL.md section or any referrer that already resolves; normalizing `~/.claude/...` outside the new module; switching the installer to a directory copy; editing `rules/task-lifecycle.md`; touching agents/CLI/schema.
- **Verification evidence:** literal diff of moved prose vs. original shows only H2→H1; all four flows read identically; all inbound/forward Step 0 references resolve; `tests/test_install.sh` passes; module has zero `~/.claude/...` refs; reviewer Karpathy scorecard PASS on Simplicity / Surgical Changes / Goal-Driven.
- **Stop conditions:** stop and re-plan if (a) any Step 0 binding cannot be pointer-preserved (1.1 no-go), (b) extraction forces editing a section outside the Step 0 stub non-mechanically, (c) the install test cannot absorb one added file without broader change, (d) the `{CRAFTER_HOME}` policy needs installer changes, or (e) the user prefers Interpretation B.

### 7. Alternatives considered

- **Delete the Step 0 heading and retarget all referrers (Interpretation B)** — rejected. Higher drift: touches Step 1 and Step 6a (sections outside the extraction target), and risks weakening the implicit "Step 0 establishes resume state before Step 1" ordering contract. Keeping the heading as a thin stub preserves every existing "Step 0" reference with zero referrer edits — strictly lower drift, exactly the conservative posture Slice 1 used.
- **Defer the `{CRAFTER_HOME}` substitution out of this task** — rejected. Slice 1 established that the policy applies to new module files in the same task that creates them; deferring would leave the module non-compliant and contradict the design note.
- **Switch `install.sh` to a `cp -R rules/do/` directory copy** — rejected for this slice. The installer's explicit per-file enumeration is the established pattern (Slice 1 kept it); one added `cp` line is lower-drift than changing installer style, and `tests/test_install.sh` enumerates files explicitly too, so a directory copy would not reduce churn.
- **Skip the pre-extraction binding validation (1.1)** — rejected. The design note explicitly flags Step 0's `task-lifecycle.md` cross-reference and guards as the thing to validate before extracting; skipping it would ignore the one risk that made Step 0 a deferred slice.

### 8. Risks / unknowns / flags

- **R1. Binding preservation (Slice 1 R1 analog).** Step 0 carries an implicit "runs before Step 1, establishes resume state and applies both guards" contract that prose-by-location currently enforces. Mitigation: the 1.1 pre-check gates the extraction, and the 1.3 stub explicitly restates the binding. Flag for user: if a reviewer judges a pointer weakens the guard ordering, the procedure may need to stay partly inline (escalate to Interpretation B or a hybrid).
- **R2. Installer change is required (corrects Slice 1's "likely none" framing).** Confirmed: `install.sh` and `tests/test_install.sh` both enumerate `rules/do/*` files explicitly, so one cp line + one test entry are mandatory. Low risk, fully mechanical, but it is a real edit — not zero as the task brief hypothesized.
- **R3. Inbound "Step 0" references.** Step 6a ("The resume detection in Step 0 will pick up...") and Step 1's resume/continue flow depend on the heading existing. Mitigation: Interpretation A retains the heading; the 1.5 sweep verifies. If the user picks B, these become required edits.
- **R4. Install-test fragility.** `tests/test_install.sh` uses an explicit expected-file list; adding one entry is straightforward, but if the test has stricter ordering/exact-count assertions the entry may need careful placement. Mitigation: run the test in 1.4 and adjust placement only (no logic change); escalate if logic changes are needed.
- **R5. Move-and-link reflex drift.** "Pure move" is easy to violate with a one-word tidy in the guard prose. Mitigation: reviewer must literally diff moved prose vs. original and confirm only H2→H1 changed.

---

This contract protects the byte-identical behavior of `crafter-do`'s resume detection and its branch-safety guards while extracting Step 0 into a named module exactly as Slice 1 extracted the preamble. The approach is right because it reuses the proven, lowest-drift move-and-link pattern, gates the extraction on an explicit binding-preservation check (the precise risk that made Step 0 a deferred slice), keeps the heading as a stub so no referrer changes, and corrects the one Slice-1 assumption (installer change) that does not hold here.

## Decisions

## Outcome
<!-- Filled on completion: what was actually done, commit SHA(s), any deviations from plan -->
