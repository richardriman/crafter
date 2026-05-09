# Task: --auto flag for crafter-do (GH#15) — flag plumbing + green-commit-protected unattended semantics

## Metadata
- **Date:** 2026-05-09
- **Work branch:** feat/GH-15-auto-flag
- **Status:** completed
- **Scope:** Medium

## Request

Implement GitHub issue #15 — `[CHANGE] --auto flag for crafter-do (YOLO mode for unattended orchestration)`.

The full refined contract is in the issue body at https://github.com/richardriman/crafter/issues/15 (revised 2026-05-09 after design discussion). Summary of what this task delivers:

**A new `--auto` flag for `crafter-do`** for unattended orchestration (Symphony, CI bots, etc.). After plan approval, the run executes Plan → Implement → Verify → Review → PR end-to-end without stopping for anything that does not threaten green commits.

**Binding invariant: green commits.** `--auto` MUST never produce a non-green commit. If the auto-fix loop cannot bring the phase back to green within budget, `--auto` exits with state and hands off to the orchestrator — it does NOT commit and continue.

**Four retained gates** (each = exit + handoff, not interactive pause):
1. Initial clarification (Analyzer cannot understand the ticket)
2. Plan approval (PLAN.md ready, awaiting human approval)
3. Green-commit cap reached (Critical/Major fix loop exhausted budget)
4. Ad-hoc escape hatch (genuinely blocked: missing auth/secret, hard contradiction)

**Removed gates** (compared to `--fast`): manual-verification exception → UAT buffer; Critical/Major findings the loop can clear → fix and continue; Minor/Suggestion → Decisions tech debt; Karpathy FLAGs → Decisions/Gaps; non-blocking step drift → Gaps/UAT; all phase summary approval gates.

**`--auto` is mutually exclusive with `--fast`** at parser level. Passing both produces a clear error. `--auto` strictly supersedes `--fast`.

**Scope of this task** (per issue):
- Add `--auto` flag handling to `skills/crafter-do/SKILL.md`
- Define `--auto` semantics in `rules/do-workflow.md` (four retained gates, green-commit invariant, mutual exclusion with `--fast`)
- Document the four retained gates and green-commit invariant
- Update `--fast` documentation to mention the mutual exclusion

**Out of scope for this task** (covered by companion issues #16/#17/#18):
- Implementation of UAT/Gaps buffers and `crafter-buffer` skill (#16)
- PR composer extension reading buffers (#17)
- Agent prompt updates for buffer-instead-of-block under `--auto` (#18)

This task is **pure flag plumbing + workflow documentation** — it lays the contract that #16/#17/#18 will then consume. References to UAT buffer / Gaps buffer / `crafter-buffer` skill in workflow rules should be present as forward references (so the contract is complete) but their implementation is owned by #16.

## Plan

**Plan status:** approved

### Approach

This is a documentation-only change to two source files (`rules/do-workflow.md` and `skills/crafter-do/SKILL.md`) that lays the contract for a new `--auto` flag on `crafter-do`. The contract makes "green commits" a binding invariant for unattended runs and reduces the human gates from "every phase summary" to exactly four well-defined exit points (clarification, plan approval, fix-loop cap reached, ad-hoc escape hatch). Everything else that the standard/`--fast` flows treat as a pause becomes either a non-blocking buffer entry (forward reference to GH#16) or an automatic record-and-continue Decision.

The plan is executed in two vertical phases. Phase 1 lands the **rules-level contract** (the source of truth that Phase 2 will reference). Phase 2 lands the **skill-level surface** (the flag declaration, prose, and Step 6b restructure that operationalizes the rules). Splitting the work this way keeps each phase reviewable in isolation: Phase 1 is reviewable as a workflow-policy document; Phase 2 is reviewable as a flag plumbing change that consumes a now-stable policy. It also lets the reviewer trace every retained gate to a single rules paragraph and a single SKILL.md paragraph, which is exactly the self-consistency check the verification requires.

### Why this approach

- **Rules first, skill second** — `rules/do-workflow.md` is the canonical contract referenced from `skills/crafter-do/SKILL.md`. Writing the contract first means the skill changes can simply cite the new sections instead of duplicating them, which keeps both files in lockstep and avoids drift.
- **Keep changes additive** — default behavior (no flag), `--fast` semantics, and `--project` resolution must remain byte-equivalent in user-facing behavior. The contract treats `--auto` as a new branch on top of the existing flow, never as a replacement.
- **Forward references are explicit** — every mention of a UAT buffer, Gaps buffer, or `crafter-buffer` skill is annotated as a forward reference to GH#16 so the next reader (and the implementer of GH#16) understands the dependency without ambiguity.

### Phase 1 — Document `--auto` semantics in `rules/do-workflow.md`

Land the workflow-level contract that defines what `--auto` means before any code path references it. After this phase, `rules/do-workflow.md` is internally consistent and complete on its own; `skills/crafter-do/SKILL.md` will then point at it in Phase 2.

- [x] Step 1: Add a top-level `### --auto (unattended orchestration)` section to `rules/do-workflow.md` that states the green-commit invariant, the four retained gates (Initial clarification, Plan approval, Green-commit cap reached, Ad-hoc escape hatch), the removed gates (manual-verification exception, Critical/Major findings the loop can clear, Minor/Suggestion findings, Karpathy FLAGs, non-blocking step drift, all phase-summary approval gates), and the `--auto`/`--fast` mutual exclusion. Each retained gate must be described as "exit + handoff via the task file as state, not interactive pause." Forward references to UAT/Gaps buffers and the `crafter-buffer` skill are annotated `(buffer skill defined in companion task GH#16)`.
- [x] Step 2: Update the existing **REVIEW** section's fix-loop cap text so the three user-choice options `(a) manual override`, `(b) accept-without-commit`, `(c) replan-and-abort` carry an explicit clause: under `--auto`, the cap-reached state does not present a choice — the orchestrator exits with state (the task file remains the handoff artifact, with the unresolved Critical/Major findings recorded as Decisions and the phase left uncommitted) and the run terminates without violating green commits.
- [x] Step 3: Update the existing **VERIFY** section to document `--auto`-specific drift handling: drift that does not threaten green commits is recorded as a Decision (Orchestrator Accepted) or Gap (forward reference to GH#16) and the run continues; drift that threatens green commits is treated as a fix-loop trigger (re-delegate to Implementer or, if the verifier classifies it as plan-obsoleting, exit via the ad-hoc escape hatch). The verifier's `ask user` recommendation is downgraded to "record and continue" when it is non-blocking and routed to the escape-hatch exit when it is blocking. Default and `--fast` behavior are unchanged.
- [x] Step 4: Add a short **Ad-hoc escape hatch** subsection inside the new `--auto` section listing example trigger conditions (missing auth/secret, hard contradiction in inputs, infrastructure outage, irrecoverable agent blocker) and clarifying that this task only documents the orchestrator behavior on receiving such a signal — agent-side recognition (Implementer/Verifier emitting a structured blocker) is owned by GH#18.
- [x] Phase verification
- [x] Phase review

#### Phase 1 Karpathy Contract

- **Outcome:** `rules/do-workflow.md` documents the complete `--auto` contract: green-commit invariant, four retained gates, removed gates with forward references, `--auto`/`--fast` mutual exclusion, and the modifications to REVIEW (fix-loop cap exit) and VERIFY (drift downgrade) under `--auto`. The file is self-consistent and complete on its own.
- **Scope boundary:** Only `rules/do-workflow.md` is modified in this phase. No SKILL.md, no agent files, no templates, no rules outside `do-workflow.md`.
- **Non-goals:** Do not implement UAT/Gaps buffers; do not modify Reviewer/Verifier/Implementer agent prompts; do not change default or `--fast` user-visible behavior; do not change `rules/post-change.md`, `rules/task-lifecycle.md`, `rules/core.md`, or `rules/delegation.md` unless an existing cross-reference becomes literally false (none are expected to).
- **Simplicity constraint:** Add new sections and brief clauses rather than restructuring existing prose. Do not rewrite the REVIEW or VERIFY sections — append narrowly scoped `--auto` clauses where they belong.
- **Drift criteria:** Drift if (a) any text changes the meaning of default/`--fast` behavior, (b) any text introduces buffer mechanics rather than referencing them as forward dependencies, (c) any retained gate is described as an interactive pause instead of "exit with state," (d) the green-commit invariant is qualified with exceptions other than the documented escape hatch.
- **Verification evidence:** Reviewer can find one paragraph or list item satisfying each of the 8 acceptance criteria from issue #15 (flag exists at parser level, mutual exclusion, four gates, green-commit invariant, removed gates enumerated, fix-loop cap exit, drift downgrade, escape-hatch trigger language).
- **Stop conditions:** Stop and ask the user before proceeding if (a) the existing REVIEW or VERIFY prose cannot accommodate the `--auto` clauses without rewriting (suggesting a deeper refactor than planned), (b) any cross-reference from `rules/post-change.md` or `rules/task-lifecycle.md` would become false under the new `--auto` text, or (c) the green-commit invariant collides with an existing rule statement.

### Phase 2 — Add `--auto` to `skills/crafter-do/SKILL.md`

Operationalize the Phase 1 contract at the skill surface: declare the flag, document it in Skill options, restructure Step 6b so the `--auto` branch precedes the existing three approval paths, and add the mutual-exclusion check at the top of the orchestrator prose. After this phase, the skill is consistent with `rules/do-workflow.md` and a reviewer can trace every retained gate to one paragraph in each file.

- [x] Step 1: Add `auto: false` to the YAML frontmatter (mirroring `fast: false`), and add a new `### --auto (default: off)` subsection to the existing `## Skill options` block. The new subsection explains the unattended-orchestration intent, lists the four retained gates and the green-commit invariant by referencing `rules/do-workflow.md`, states the mutual exclusion with `--fast`, and updates the existing `### --fast` subsection to mention the mutual exclusion symmetrically. Cross-reference Step 6b for the approval-path branch that consumes the flag.
- [x] Step 2: Add a `--auto`/`--fast` mutual-exclusion check at the very top of the orchestrator prose (before "Project Resolution"), so the rejection happens before any project work begins. The check produces a clear error message and stops the workflow. The location is intentionally before project resolution because the rejection must be unambiguous and independent of project state.
- [x] Step 3: Restructure `## Step 6b — Phase Summary and Auto-Commit` to branch on `--auto` first, then fall through to the existing three approval paths when `--auto` is not set. Under `--auto`: there is no Phase Summary surfaced to the user, no waiting for approval; the orchestrator records remaining Minor/Suggestion findings as Decisions (tech debt) and any manual-verification requirements as UAT buffer entries `(buffer skill defined in companion task GH#16)`, then commits automatically per `rules/post-change.md`. The existing paths (1) auto-approve, (2) `--fast` silence-approve, (3) explicit approval remain unchanged in wording for the non-`--auto` branch.
- [x] Step 4: Add a brief `--auto`-specific clause to the existing fix-loop cap handling reference in Step 6 (or wherever Step 6 currently mentions the cap-reached choice), noting that under `--auto` the orchestrator does not present the (a)/(b)/(c) choice and instead exits with state per `rules/do-workflow.md` REVIEW section. This is a single sentence cross-reference; the canonical text lives in the rules file.
- [x] Phase verification
- [x] Phase review

#### Phase 2 Karpathy Contract

- **Outcome:** `skills/crafter-do/SKILL.md` declares `auto: false` in frontmatter, documents `--auto` in Skill options, enforces `--auto`/`--fast` mutual exclusion at the top of the orchestrator prose, and restructures Step 6b so the `--auto` branch precedes the three existing approval paths. Every retained gate referenced in SKILL.md has matching prose in `rules/do-workflow.md` and vice versa.
- **Scope boundary:** Only `skills/crafter-do/SKILL.md` is modified in this phase. No agent files, no other skills, no rules.
- **Non-goals:** Do not change the wording of the existing three approval paths; do not change `--fast` semantics beyond adding the mutual-exclusion mention; do not change project resolution or resume detection; do not introduce a new task-file schema for `--auto` state.
- **Simplicity constraint:** Reuse the existing `--fast` documentation pattern (frontmatter boolean, Skill options subsection, Step 6b path). Treat `--auto` as a new top-level branch in Step 6b, not a rewrite of the existing structure.
- **Drift criteria:** Drift if (a) the SKILL.md prose contradicts `rules/do-workflow.md` on any retained gate, removed gate, or invariant, (b) the mutual-exclusion check appears anywhere other than before project resolution, (c) the existing three approval paths' wording or conditions change, (d) the `--auto` branch in Step 6b mentions any user pause that is not one of the four retained gates.
- **Verification evidence:** Reviewer can (i) find `auto: false` in frontmatter, (ii) find a `### --auto` subsection in Skill options that names all four retained gates and the green-commit invariant, (iii) find the mutual-exclusion check before project resolution, (iv) find the `--auto` branch at the top of Step 6b that records-and-commits without surfacing a Phase Summary, (v) trace every gate mentioned in SKILL.md to a corresponding paragraph in `rules/do-workflow.md`.
- **Stop conditions:** Stop and ask the user before proceeding if (a) the Step 6b restructure cannot keep the existing three paths byte-equivalent for non-`--auto` runs without significant rewriting, (b) the mutual-exclusion check placement before project resolution conflicts with an existing pre-resolution check, or (c) the frontmatter `auto: false` declaration interacts with the installer or any tool in an unforeseen way (the implementer should confirm by reading `install.sh` if uncertain).

### Alternatives considered

- **Single phase covering both files.** Rejected because reviewing rules and skill together produces a larger diff with more places to drift, and because Phase 2 explicitly cites Phase 1 — splitting lets the rules contract stabilize first.
- **Putting the `--auto` branch at the bottom of Step 6b after the existing three paths.** Rejected because the existing three paths' conditions (clean review, `--fast`, default) all assume a Phase Summary is surfaced; placing `--auto` last would require defensive "unless `--auto` was set" qualifiers in each existing path. Branching on `--auto` first is structurally cleaner.
- **Documenting the green-commit invariant only in SKILL.md.** Rejected because the invariant constrains REVIEW and VERIFY behavior, which live in `rules/do-workflow.md`. Putting it only in the skill would create an asymmetry where the rules say "ask user" and the skill says "exit with state."
- **Introducing a new task-file field for `--auto` exit state.** Rejected per the issue's verification plan footer ("handoff artifact in expected location") and the existing pattern: the task file with its Plan checkboxes and Decisions section already serves as the handoff artifact. No new schema is needed.
- **Defining the ad-hoc escape hatch trigger conditions exhaustively.** Rejected as over-prescriptive — the issue intentionally lists representative triggers ("missing auth/secret, hard contradiction, infrastructure outage") rather than an exhaustive list. The plan documents the exit semantics; agent-side recognition is GH#18's responsibility.

### Risks / unknowns / flags

- **Forward references to GH#16 readability.** The `(buffer skill defined in companion task GH#16)` annotation is the chosen idiom. If a reviewer finds it intrusive, an alternative is a single footnote at the top of the new `--auto` section explaining that buffer mechanics are owned by GH#16 and all `(buffer)` mentions in the section refer to it. The implementer may choose the footnote form if the inline annotations clutter the prose.
- **Step 6 cap-reached cross-reference.** Step 6 in SKILL.md currently embeds the (a)/(b)/(c) choice in full. Phase 2 Step 4 adds a brief `--auto` cross-reference there. If the cap-reached text proves to need its own restructure for clarity under `--auto`, that would expand Phase 2 — flag to the user before doing it.
- **`auto: false` frontmatter and installer.** The installer copies SKILL.md verbatim, so adding a frontmatter key should be transparent. The Phase 2 stop condition asks the implementer to confirm by reading `install.sh` if anything looks unusual.
- **Drift handling subtlety under `--auto`.** Phase 1 Step 3 distinguishes "drift that threatens green commits" from "drift that does not." This classification is currently the verifier's judgment call. The plan documents the orchestrator behavior given the verifier's classification; agent-side classification refinements are GH#18's responsibility. If the documented behavior is ambiguous on a real verifier output, the escape hatch is the safe default.
- **`crafter-debug` skill is not in scope.** This task does not touch `skills/crafter-debug/SKILL.md`. If the user later wants `--auto` for debug runs, that is a separate task — flag if the user expected it in this scope.

### Verification (phase-level)

- **Phase 1 verification:** `rules/do-workflow.md` contains a top-level `--auto` section that names all four retained gates and the green-commit invariant; the REVIEW section's fix-loop cap text contains an `--auto` exit clause; the VERIFY section contains an `--auto` drift-handling clause; the removed-gates list is present with forward references to GH#16; the `--auto`/`--fast` mutual exclusion is stated. Default and `--fast` user-visible behavior described in the file are unchanged.
- **Phase 2 verification:** `skills/crafter-do/SKILL.md` has `auto: false` in frontmatter, a `### --auto` subsection in Skill options that mirrors the `--fast` pattern, a mutual-exclusion check before project resolution, and a Step 6b that branches on `--auto` first then falls through to the three existing paths. Cross-checking SKILL.md against `rules/do-workflow.md`: every retained gate, the green-commit invariant, and the mutual exclusion appear in both files with no contradiction.
- **End-of-task verification (full task):** Walk each of the 8 issue acceptance criteria and identify the file:section that satisfies it. Walk each of the 5 E2E scenarios from the issue verification plan and trace it through the documented workflow to confirm the contract produces the expected runtime behavior.

## Decisions

- **Decision (Tech Debt — auto-recorded) — Phase 2 Review, 2026-05-09:** Minor #3 — markdown rendering ambiguity for `--auto` cross-reference clause inserted at same indentation as `(a)/(b)/(c)` lead-in in Step 6f.1. Verifier classified as marginal-but-PASS; readability nit only. Could be moved to a parenthetical at end of lead-in line in a future polish.
- **Decision (User Accepted) — Phase 2 polish follow-up, 2026-05-09:** Applied Suggestions #4 and #5 from the Phase 2 review (commit `f74b146`). #4: orientation line at SKILL.md line 16 generalized to mention both `--fast` and `--auto`. #5: four approval-branch labels in Step 6b (`--auto` branch + paths 1/2/3) promoted from bold pseudo-headings to `####` markdown headings for consistent TOC visibility. Body text byte-equivalent. **Reason:** User explicitly requested fixes for findings 1, 2, 4, 5; #1 and #2 were already applied in 308f828; #3 (markdown indentation in Step 6f.1) deliberately skipped per verifier's marginal-but-PASS classification.
- **Decision (Auto Mode — Phase 2 commit, 2026-05-09):** User enabled Claude Code Auto Mode mid-task. Phase 2 review found 0 Critical/Major findings and 5 Minor/Suggestion findings; orchestrator applied a polish iteration for Minor #1 + #2 (cross-consistency tighten-ups), recorded the remaining Minor/Suggestion findings as tech-debt Decisions above, and committed Phase 2 without explicit approval prompt per Auto Mode minimize-interruptions guidance. Re-verification confirmed PASS post-polish.
- **Decision (User Accepted) — Phase 1 Review, 2026-05-09:** User opted to apply Minor/Suggestion fixes #1–#4 and #7 from the Phase 1 review via a polish iteration. Fixes #5 and #6 are deferred. **Reason:** #5 (adding GH#18 to the section-level footnote) and #6 (tightening the Retained-gates one-liner that duplicates the Ad-hoc escape hatch trigger list) are pure stylistic suggestions that do not threaten the contract. The user judged them not worth the prose churn; they remain as known tech debt in the task record but require no follow-up work.
- **Decision (Orchestrator Accepted) — Phase 1 Review fix iteration, 2026-05-09:** The implementer applied Fix #3 ("Gaps buffer" terminology alignment) at one additional site beyond the reviewer's enumerated lines — line 132 (rarity note inside `#### Ad-hoc escape hatch`, "Decisions or Gaps" → "Decisions or Gaps buffer"). **Reason:** This is strictly within the stated Fix #3 scope rule ("apply consistently throughout the `### --auto` section") and improves uniformity rather than expanding scope.
- **Decision (Orchestrator Accepted) — Phase 1 Step 1, 2026-05-09:** Two local beneficial deviations from the step contract recorded together. (a) `.crafter/skillbook.json` was modified alongside `rules/do-workflow.md` — `appliedCount: 0 → 1` and `updatedAt` timestamp. **Reason:** This is an automated metadata side-effect of orchestrator-level `crafter skillbook get` Bash calls (used to fetch agent guidelines per `delegation.md`), not implementer-driven content. It will recur on every step. Including it in the commit is preferable to repeatedly reverting benign tracking metadata; this does not threaten the green-commit invariant. (b) The new `### --auto` section uses `####` subheadings (Green-commit invariant, Retained gates, Removed gates) — implementer chose them as navigational aid for the ~30-line section despite the contract's "flow narrative + lists" guidance. **Reason:** Local readability win, scope-bounded to the new section, no impact on later steps; verifier independently assessed this as acceptable.
- **Decision (User Accepted) — 2026-05-09:** Issue #15 body refined before planning. Added green-commit invariant as binding constraint; promoted "green-commit cap reached" and "ad-hoc escape hatch" from implicit behavior to explicit retained gates (2 → 4); made `--auto` mutually exclusive with `--fast` at parser level. **Reason:** Original issue language said Critical/Major findings would be buffered "not blocking", which would silently break the green-commit invariant — hard contradiction. User chose green-commit as the higher-precedence rule. Mutual exclusion with `--fast` chosen to eliminate ambiguity in flag combination semantics.

## Outcome

**Delivered:** GH#15 `--auto` flag contract for crafter-do unattended orchestration — pure flag plumbing + workflow documentation, no code execution paths. Lays the contract that companion tasks GH#16 (UAT/Gaps buffer skill), GH#17 (PR composer), and GH#18 (agent prompts) will consume.

**Phase 1 commit:** `828815e` — `docs(rules): document --auto flag contract in do-workflow.md (GH#15 Phase 1)` — added `### --auto (unattended orchestration)` section to `rules/do-workflow.md` with green-commit invariant, four retained gates, removed gates list, and `--auto`/`--fast` mutual exclusion; added `--auto` clauses to existing REVIEW (fix-loop cap exit) and VERIFY (drift downgrade) sections; added `#### Ad-hoc escape hatch` subsection.

**Phase 2 commit:** `308f828` — `feat(skill): wire --auto flag into crafter-do skill (GH#15 Phase 2)` — added `auto: false` to YAML frontmatter, `### --auto (default: off)` Skill options subsection, `## Flag Validation (before anything else)` block before Project Resolution, restructured `## Step 6b — Phase Summary and Auto-Commit` so the `--auto` branch precedes the existing three approval paths (paths byte-equivalent for non-`--auto` runs), and a single-sentence `--auto` cross-reference in Step 6f.1 above the `(a)/(b)/(c)` cap-reached choices. `### --fast` subsection updated with symmetric mutual-exclusion mention.

**Verification:** Phase 1 and Phase 2 both passed phase verification. Phase 1 review applied fixes #1–#4 + #7 in a polish iteration; #5 + #6 deferred as recorded tech debt. Phase 2 review applied polish for Minor #1 + #2 (cross-consistency tighten-ups for SKILL.md ↔ `rules/do-workflow.md`); Minor #3 + Suggestions #4–#5 recorded as tech-debt Decisions.

**End-of-task commit:** consolidates ARCHITECTURE.md `### Human-in-the-Loop Gates` paragraph for the new `--auto` mode, STATE.md Recent Changes entry, and 2 new skillbook observations (one reviewer, one implementer).

**Deviations from plan:** None beyond the four accepted deviations recorded in `## Decisions` (bullet labels for retained gates in `### --auto`, structural repositioning of `### Phase Summary content` block in Step 6b, inline placement of `--auto` cross-reference in Step 6f.1, and skillbook/task-file metadata side-effects from orchestrator-level Bash calls).
