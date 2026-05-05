# Task: Automatic commit after each successful phase

## Metadata
- **Date:** 2026-05-05
- **Work branch:** main
- **Status:** completed
- **Scope:** Medium

## Request
Two related changes that together produce "green commits" — commits land only when a phase is truly clean:

1. **Commit per phase (auto):** After a phase completes successfully, the orchestrator automatically creates a commit using conventional commits format, without waiting for explicit user instruction. No auto-push. Currently commits are gated behind an explicit user instruction ("only commit when the user explicitly says to").

2. **Critical/Major findings must be fixed before user sees the summary:** In the review loop, Critical and Major findings may no longer be skipped with "Proceed anyway" — they must be fixed. After the fix loop closes (no Critical/Major remain), the orchestrator presents a structured summary to the user: what was found (including what was already auto-fixed), and any remaining Minor/Suggestion findings for the user to decide on. Only after the user approves the phase as resolved does the commit happen.

Together: vertical-slice phases get a green commit — all critical/major issues resolved, user has reviewed the summary and signed off.

## Plan

**Plan status:** approved

### Approach

The expanded request couples two policy changes that together produce "green commits": (1) Critical/Major review findings become mandatory to fix — "Proceed anyway" is no longer offered for those severities; (2) once the review loop closes clean, the orchestrator presents a structured per-phase summary to the user, and only after the user signs off does an automatic conventional-commits commit land (no push, no further prompt for the commit itself). The work is prose-only across three files: `rules/post-change.md`, `rules/do-workflow.md`, and `skills/crafter-do/SKILL.md`. No new agents, no new verifier modes, no orchestration restructuring beyond aligning Step 6 / Step 6a / Steps 7–9 wording.

The change is best modeled as a vertical phase loop hook: at end-of-phase the orchestrator (a) drives the mandatory fix loop until Critical/Major are gone or the 5-iteration cap forces an explicit user decision, (b) presents a phase summary including auto-fixed items and any remaining Minor/Suggestion findings, (c) waits for the user's "approved" response (or, when the skill's `--fast` flag is set, treats silence as implicit approval; or auto-approves when the summary is fully clean and no manual verification is required), then (d) commits automatically. This is Medium scope because it changes review semantics, adds a new per-phase summary artifact to the orchestrator's output, introduces a metadata flag governing silence-as-approval, and rewires the relationship between Step 6 and the post-change commit. Two phases keep each landing reviewable on its own.

### Phase 1 — Tighten the review loop: Critical/Major are mandatory to fix

This phase changes review semantics only. After it ships, the orchestrator no longer offers "Proceed anyway" for Critical/Major findings, and the 5-iteration cap behavior is explicit. No commit timing changes yet.

- [x] **Step 1.1 — Update Step 6 review-handling in `skills/crafter-do/SKILL.md`**
  - Outcome: Sub-step (e) and (f) of Step 6 say that when Critical or Major issues are present, the orchestrator must run the fix loop — there is no "Proceed anyway" choice for those severities. Minor/Suggestion findings keep their informational behavior.
  - Relevant area: `/Users/ret/dev/ai/crafter/skills/crafter-do/SKILL.md` Step 6, sub-steps (d) through (f).
  - Drift criteria: Only Step 6 sub-steps (d)–(f) wording changes. Iteration counter, Verifier delegation, Reviewer table reproduction format, and Minor/Suggestion handling are untouched. No change to Step 5, 5a, 6a, 7–9.
  - Verification evidence: Read SKILL.md; sub-step (e) lists only "Fix and re-review" for Critical/Major; "Proceed anyway" appears only as a fallback at the 5-iteration cap (see Step 1.2).

- [x] **Step 1.2 — Define the 5-iteration cap behavior explicitly**
  - Outcome: SKILL.md states that if after 5 review iterations Critical or Major findings still remain, the orchestrator does NOT auto-proceed and does NOT commit. Instead it stops, presents the remaining findings to the user, and asks the user to choose: (a) manual override — keep iterating manually outside the cap, (b) accept-without-commit — accept the unresolved findings and proceed without committing (records a Decision and leaves the phase uncommitted, deliberately breaking the green-commit invariant), or (c) replan-and-abort — abandon the phase and re-plan. The cap is set to 5 automatic iterations.
  - Relevant area: `/Users/ret/dev/ai/crafter/skills/crafter-do/SKILL.md` Step 6 sub-step (f.1); `/Users/ret/dev/ai/crafter/rules/do-workflow.md` REVIEW section (the lines about the iteration cap and "Proceed anyway").
  - Drift criteria: The cap is exactly 5 (raised from the previous 3). The escape hatches are user-driven only — no automatic acceptance of unresolved Critical/Major. No new agent, no new verifier mode. Any pre-existing references to "3 iterations" / "3-iteration cap" must be updated to 5 wherever they appear inside the touched sections.
  - Verification evidence: Read both files; the cap behavior names the three user choices (manual override / accept-without-commit / replan-and-abort); "Proceed anyway" wording survives only as the cap-escape, not as a normal in-loop choice; no stale "3 iterations" wording remains in the touched sections.

- [x] **Step 1.3 — Align `rules/do-workflow.md` REVIEW prose with the new mandatory-fix rule**
  - Outcome: The REVIEW bullet that currently says "Critical or Major issues trigger the 'Fix and re-review' (recommended) or 'Proceed anyway' prompt" is rewritten so Critical/Major trigger only the fix loop in normal operation, with "Proceed anyway" reserved for the 5-iteration cap path.
  - Relevant area: `/Users/ret/dev/ai/crafter/rules/do-workflow.md` REVIEW section (specifically the bullet starting "After the user responds:" and the bullet about the iteration cap).
  - Drift criteria: Only those two bullets change. Reviewer output format (tables) and STOP wait-for-user behavior unchanged.
  - Verification evidence: Read the file; wording matches SKILL.md Step 6 exactly (no contradiction between the canonical rules and the skill prose).

#### Karpathy Contract — Phase 1

- **Outcome:** After this phase, the do-workflow's review semantics treat Critical/Major as must-fix in the normal path; the only way past unresolved Critical/Major is the explicit 5-iteration-cap user override.
- **Scope boundary:** Two files — `skills/crafter-do/SKILL.md` (Step 6) and `rules/do-workflow.md` (REVIEW section). Prose-only.
- **Non-goals:** No commit-timing changes. No phase-summary changes. No reviewer-agent changes. No changes to severity definitions. No new opt-out flags.
- **Simplicity constraint:** Edits should fit within Step 6 and the REVIEW section bullets. If structural reorganization of Step 6 is required, stop.
- **Drift criteria:** Harmful drift = silently auto-accepting Critical/Major, removing the iteration cap, or leaving the cap at 3. Scope drift = editing the Reviewer agent or other rules files. Beneficial local drift = small consistency edits between SKILL.md and `do-workflow.md` to keep wording aligned.
- **Verification evidence:** Read both files; "Proceed anyway" appears at most once and only inside the cap escape; Critical/Major in-loop handling has no "proceed anyway" alternative; iteration counter logic intact.
- **Stop conditions:** If aligning the cap behavior requires touching agents (`agents/crafter-reviewer.md`, etc.) or `rules/post-change.md`, stop and re-plan.

### Phase 2 — Auto-commit per phase, gated by user-approved summary

This phase adds the per-phase summary artifact and the automatic commit, and reconciles Steps 7–9 with the new flow. It depends on Phase 1 because the commit is only safe to automate once Critical/Major are guaranteed to be resolved.

- [x] **Step 2.1 — Rewrite the COMMIT rule in `rules/post-change.md`**
  - Outcome: The `## COMMIT` section says: when phase verification has passed, the review fix loop has closed (no Critical/Major remain), and the user has explicitly approved the phase summary, the orchestrator commits automatically using conventional commits format. No push to remote. The "one logical change = one commit" line stays. The "Update STATE.md" sub-section's "after a successful commit" wording is left coherent with the new flow (STATE.md update remains end-of-task per Phase 2 assumption 3).
  - Relevant area: `/Users/ret/dev/ai/crafter/rules/post-change.md` `## COMMIT` section (and the one-line bridge to `## Update STATE.md` if needed for coherence).
  - Drift criteria: Only the COMMIT subsection (and the immediately adjacent bridge line) changes. No new sections, no opt-out flag, no push behavior added, no changes to skillbook/task-file/wrap-up sections.
  - Verification evidence: Read the file; the rule names all three preconditions (phase verify + clean review + user summary approval), states "automatic, no push," and keeps conventional-commits format.

- [x] **Step 2.2 — Add the phase summary + commit hook to Step 6 / new Step 6b in `skills/crafter-do/SKILL.md`**
  - Outcome: After Step 6's review loop closes clean (no Critical/Major), the orchestrator presents a structured **Phase Summary** to the user covering: what was implemented in the phase, auto-fixed findings recorded during the fix loop, remaining Minor/Suggestion findings, and any accepted decisions. Approval semantics: (1) **Empty-summary auto-approve** — if there are zero remaining findings (nothing was found, or every Critical/Major was auto-fixed and no Minor/Suggestion remain), the orchestrator auto-approves and commits immediately, with one exception: if the phase or any of its steps is flagged as requiring manual verification (non-automatable testing — UI, external integration), the orchestrator still waits for explicit user confirmation. (2) **Explicit approval (default)** — when remaining Minor/Suggestion findings exist, the orchestrator waits for an affirmative response; silence does not count. (3) **Silence-as-approval (opt-in)** — only when the skill carries the `--fast` flag (introduced in Step 2.4) does silence count as implicit approval; in that case the orchestrator records remaining Minor/Suggestion findings as tech debt in the task file's Decisions section and commits. On approval (any path), the orchestrator runs the commit per `rules/post-change.md`. The simplest landing is a new Step 6b ("Phase summary and auto-commit") between current Step 6 and Step 6a; the Implementer chooses the exact location as long as the ordering Verify → Review → Summary → Approval (auto/explicit/silence-flag) → Commit → Step 6a is unambiguous.
  - Relevant area: `/Users/ret/dev/ai/crafter/skills/crafter-do/SKILL.md` end of Step 6, beginning of Step 6a, and the intro to Steps 7–9.
  - Drift criteria: Only one new sub-step (6b) is added; existing Steps 5, 5a, 6, 6a, 7–9 retain their identifiers. No new agent. The summary is produced by the orchestrator from already-available context (review report, decisions log, fix-loop history) — no new task-file fields and no new agent invocation. The three approval paths (auto on empty, explicit, silence-flag) must all be documented; manual-verification override must be explicit.
  - Verification evidence: Read SKILL.md end-to-end; the phase loop reads as Implement → Drift Check → Phase Verify → Review (with mandatory fix loop) → Phase Summary → (auto-approve if clean & no manual verification | silence-approve if `--fast` set | otherwise wait for explicit user approval) → Auto-commit (no push) → Step 6a (session break) or Steps 7–9 (if last phase).

- [x] **Step 2.3 — Update the MANDATORY CHECKLIST item 2 in SKILL.md and reconcile Steps 7–9 around a single consolidated end-of-task commit**
  - Outcome: Item 2 of the MANDATORY CHECKLIST no longer says "only commit when the user explicitly says to." It points to the per-phase auto-commit defined in Step 6b and `post-change.md`. For the final phase, the per-phase commit lands as normal. Steps 7–9 are reconciled so that any PROJECT.md/ARCHITECTURE.md updates, the skillbook entry, and the STATE.md update are bundled into **one consolidated end-of-task commit** (a single follow-up commit), instead of the prior "at most one follow-up commit" model that left the boundaries between docs / skillbook / STATE ambiguous. If none of those updates are needed, no follow-up commit is created. Item 3 (STATE.md update) is restated as part of the consolidated commit.
  - Relevant area: `/Users/ret/dev/ai/crafter/skills/crafter-do/SKILL.md` "MANDATORY CHECKLIST" items 2 and 3, plus the intro paragraph to Steps 7–9 and the wording in Steps 7, 8, 9 that previously implied separate commits.
  - Drift criteria: Items 1, 4, 5 unchanged. The "Do not end the conversation until all 5 items above are addressed" guarantee preserved. No double-commit at end of last phase. Steps 7–9 must produce at most ONE consolidated follow-up commit (docs + skillbook + STATE together), not multiple.
  - Verification evidence: Read SKILL.md; checklist item 2 is consistent with `post-change.md`'s new rule; final-phase walk-through yields exactly one phase commit plus at most one consolidated follow-up commit covering all of PROJECT.md/ARCHITECTURE.md updates, skillbook entry, and STATE.md.

- [x] **Step 2.4 — Add the `--fast` flag to the crafter-do skill metadata**
  - Outcome: `skills/crafter-do/SKILL.md` exposes a `--fast` metadata field (or equivalent front-matter / configuration line — Implementer's choice consistent with existing skill metadata conventions) that, when set, allows the orchestrator to treat user silence on the phase summary as implicit approval (per Step 2.2). The flag is documented in plain prose alongside the metadata as an **intentional speed-vs-explicitness trade-off**: with it on, phases ship faster and remaining Minor/Suggestion findings are auto-recorded as tech debt in the task Decisions; with it off (default), the orchestrator waits for an explicit affirmative response. The flag's behavior must not be hidden — its location in the skill and its effect must both be visible to a user reading SKILL.md.
  - Relevant area: `/Users/ret/dev/ai/crafter/skills/crafter-do/SKILL.md` metadata/front-matter area and the Step 6b prose (cross-reference).
  - Drift criteria: Exactly one new metadata field added. Default is unset/false (existing behavior unchanged for current users). No other skill metadata edited. The flag is documented in SKILL.md prose, not just declared. No equivalent flag added to other skills or rules files.
  - Verification evidence: Read SKILL.md; the flag appears in the skill's metadata block; a prose paragraph explains the trade-off and the tech-debt-logging behavior; Step 6b cross-references the flag by name; default behavior (flag absent or false) matches the explicit-approval path.

#### Karpathy Contract — Phase 2

- **Outcome:** Every successful phase ends with: a clean review (no Critical/Major), an approved phase summary (auto-approved when clean and no manual verification flagged; silence-approved when `--fast` is set; otherwise explicit user approval), and an automatic conventional-commits commit (no push). The final phase's commit is the same per-phase commit; Steps 7–9 add **one consolidated follow-up commit** covering docs (PROJECT.md/ARCHITECTURE.md), skillbook entry, and STATE.md — only if any such updates exist.
- **Scope boundary:** Two files — `rules/post-change.md` and `skills/crafter-do/SKILL.md`. Prose-only, plus one new metadata field on the crafter-do skill (`--fast`).
- **Non-goals:** No auto-push. No opt-out flag for the auto-commit itself (the silence flag is an approval-mode opt-in, not a commit opt-out). No new agent or "committer" role. No change to severity definitions, conventional-commits format, or task-file schema beyond logging tech debt to existing Decisions section. No STATE.md update at every phase (stays end-of-task, now bundled into the consolidated commit). No automatic commit when Critical/Major remain unresolved (Phase 1 guarantees they cannot at this point). No splitting of docs/skillbook/STATE into separate commits.
- **Simplicity constraint:** A single new sub-step (6b), one new metadata field, and rule wording changes. If a clean landing requires renumbering Steps 7–9 or introducing a new agent, stop.
- **Drift criteria:** Harmful drift = pushing to remote; committing without an approval signal of any kind; committing while Critical/Major outstanding (impossible after Phase 1 unless cap escape was taken); silence treated as approval when the flag is absent; auto-approval bypassing manual-verification flags; emitting more than one follow-up commit for docs/skillbook/STATE. Scope drift = editing files outside the two listed. Beneficial local drift = aligning the "after a successful commit" bridge wording in `post-change.md` for coherence — record as decision.
- **Verification evidence:** Read both files; commit preconditions are stated identically; no file references "only commit when the user explicitly says to"; the three approval paths and the manual-verification override are documented; the `--fast` flag is declared in metadata and explained in prose as an intentional trade-off; mental walk-through of a 2-phase task yields exactly 2 phase commits (plus at most 1 consolidated docs/skillbook/STATE follow-up); 1-phase Small task yields exactly 1 phase commit (plus at most 1 consolidated follow-up); empty-summary clean phase auto-commits without prompting (unless manual verification is flagged).
- **Stop conditions:** If grepping reveals "only commit when the user" or "Proceed anyway" wording in agents or other rules files we did not list, stop and ask the user before broadening scope. If the per-phase summary needs structured data the orchestrator does not already have, stop and re-plan (we do not want to introduce new task-file fields).

### Assumptions

1. **"Mandatory fix" semantics vs. the 5-iteration cap.** The cap is set to 5 automatic iterations (raised from the previous 3, per user direction). "Mandatory" applies to the normal in-loop choice — no "Proceed anyway" is offered for Critical/Major while iterations remain. If the cap is reached with Critical/Major still present, the orchestrator stops, presents the situation, and offers the user three explicit choices: (a) **manual override** — authorize manual iteration beyond the cap, (b) **accept-without-commit** — accept the unresolved findings and proceed without committing (the phase ends uncommitted, recorded as a Decision), or (c) **replan-and-abort** — abandon the phase. Option (b) explicitly breaks the green-commit invariant for that phase by user choice. All previous mentions of "3 iterations" / "3-iteration cap" in the touched sections are updated to 5.
2. **Approval semantics for the phase summary** has three paths, defined per user direction: (a) **Auto-approve on clean summary** — if there are no findings at all, or all Critical/Major were auto-fixed and no Minor/Suggestion remain, the orchestrator auto-approves and commits immediately, EXCEPT when the phase or step is flagged as requiring manual verification (non-automatable testing — UI, external integration), in which case explicit user confirmation is always required even on a clean review. (b) **Silence-as-approval (opt-in only)** — silence counts as implicit approval ONLY when the crafter-do skill carries the `--fast` metadata flag; with the flag set, remaining Minor/Suggestion findings are auto-recorded as tech debt in the task file's Decisions section and the orchestrator commits. (c) **Explicit approval (default)** — without the flag and with remaining Minor/Suggestion findings, silence does not count and the orchestrator waits for an affirmative response.
3. **Phase-summary content** is derived entirely from artifacts the orchestrator already has: the Reviewer's report (auto-fixed items appear because the fix loop ran on them and the next review pass cleared them), accepted Decisions in the task file, and the verifier's report. No new metadata is added to the task file beyond the `--fast` flag on the skill itself.
4. **End-of-task consolidated commit.** Per user direction, the final phase's per-phase commit lands as normal. Then any PROJECT.md/ARCHITECTURE.md updates, the skillbook entry, and the STATE.md update are bundled into a **single consolidated end-of-task commit** (one follow-up commit, not multiple). If none of those updates are needed, no follow-up commit is created. This replaces the previous "at most one follow-up commit" model with a single explicit consolidated commit.
5. **STATE.md update stays end-of-task** (Steps 7–9), bundled into the consolidated commit (assumption 4), not per-phase. This keeps per-phase commits free of STATE.md churn.
6. **The orchestrator (not an agent) executes the commit and produces the phase summary** — git operations are already orchestrator-level in `post-change.md`, and the summary is a presentation concern that fits the orchestrator role described in `skills/crafter-do/SKILL.md`.
7. **Conventional-commits subject** is derived from the task title and current phase name — no new task-file field is required.
8. **Single-phase Small tasks** still get exactly one auto-commit at end-of-phase, plus an optional consolidated follow-up if docs/skillbook/STATE.md updates exist.
9. **The `--fast` flag is in scope for this task.** It is documented as an intentional trade-off (speed over explicitness), not a hidden behavior — its location in the skill metadata and its effect must both be visible to a user reading SKILL.md.

### Alternatives considered

- **Skip the user-approval gate entirely and auto-commit immediately after the review loop closes clean.** Partially adopted (per user direction): a clean summary with no remaining findings auto-approves and commits without prompting, unless the phase or step is flagged as requiring manual verification. The gate still applies whenever Minor/Suggestion findings remain, unless `--fast` is set.
- **Treat the iteration cap as a hard fail (abort the phase).** Rejected — it removes user agency and conflicts with current rules that already let the user accept or override. Keeping three explicit user choices preserves flexibility while still preventing automatic acceptance of unresolved Critical/Major.
- **Always treat silence as approval.** Rejected — per user direction, silence counts only when the skill carries the `--fast` opt-in flag. This keeps the default explicit and surfaces the speed-vs-explicitness trade-off as a deliberate choice.
- **Emit separate commits for docs, skillbook, and STATE.md at end of task.** Rejected per user direction — the end-of-task follow-up is consolidated into a single commit covering all three artifacts.
- **Update STATE.md per phase commit.** Rejected — produces noisy commits with churn unrelated to the phase's logical change, and the codebase already centralizes STATE.md updates at end-of-task.
- **Commit only at end of task with a single roll-up commit.** Rejected — the request explicitly asks for one commit per phase, and per-phase commits give cleaner history and easier revert/bisect.
- **Introduce an opt-out flag (`--no-auto-commit`) for users who want the old behavior.** Rejected per existing project guidance — the user can always amend/reset locally.
- **Delegate the phase summary to a new agent.** Rejected — the summary is a presentation of data the orchestrator already holds; adding an agent hop has no benefit.

### Verification criteria (phase-level)

**Phase 1:**
- Grep across `rules/` and `skills/` for "Proceed anyway" — appears only inside the 5-iteration-cap escape, not in the normal Critical/Major path.
- Step 6 sub-steps in SKILL.md no longer offer "Proceed anyway" as a normal choice for Critical/Major.
- `rules/do-workflow.md` REVIEW section is consistent with SKILL.md Step 6 (no contradictions).
- The cap is set to 5 (no stale "3 iterations" / "3-iteration cap" wording remains in the touched sections).
- The cap behavior names the three user choices explicitly: manual override, accept-without-commit, replan-and-abort.

**Phase 2:**
- Grep across `rules/` and `skills/` for "only commit when the user" / "explicitly says" — zero matches that contradict the new auto-commit policy.
- `rules/post-change.md` `## COMMIT` section names the preconditions (phase verify + clean review + approval signal — auto/silence-flag/explicit), states "automatic, no push," keeps conventional-commits format.
- SKILL.md Step 6b (or equivalent insertion) makes the order Verify → Review → Summary → Approval (auto on clean / silence-approve if flagged / explicit otherwise) → Commit unambiguous, and explicitly documents the manual-verification override.
- The `--fast` flag is declared in the crafter-do skill metadata, defaults to off/absent, and is documented in prose as an intentional speed-vs-explicitness trade-off.
- Empty-summary auto-approve is documented and clearly excludes phases/steps flagged as requiring manual verification.
- MANDATORY CHECKLIST item 2 matches `post-change.md`.
- Steps 7–9 produce at most ONE consolidated follow-up commit covering docs (PROJECT.md/ARCHITECTURE.md), skillbook entry, and STATE.md — never multiple.
- Mental walk-through: 2-phase task → 2 phase commits (plus ≤1 consolidated docs/skillbook/STATE follow-up); 1-phase Small task → 1 phase commit (plus ≤1 consolidated follow-up); clean-summary phase with no manual verification → auto-commits without prompting; no double-commit at end of last phase.
- No file outside the listed three is modified.

### Risks / unknowns / flags

- ~~**Iteration cap edge case**~~ — Resolved by user direction: cap is raised to 5; on cap-hit the orchestrator offers three explicit choices (manual override / accept-without-commit / replan-and-abort).
- ~~**What counts as "user approval" of the summary**~~ — Resolved by user direction: three paths (auto-approve on clean summary unless manual verification is flagged; silence-as-approval only when `--fast` is set; otherwise explicit affirmative response required).
- ~~**Phase-summary verbosity**~~ — Resolved by user direction: clean phases auto-commit without a verbose summary gate (subject to manual-verification override).
- ~~**STATE.md and docs follow-up commit**~~ — Resolved by user direction: docs, skillbook, and STATE.md updates are bundled into a single consolidated end-of-task commit.
- **Existing in-flight tasks** with the old expectations may be surprised by the new behavior. Mitigation: rule change applies forward; user is requesting it explicitly.
- **Reviewer agent contract** — we are not editing `agents/crafter-reviewer.md`. If the Reviewer's output schema is needed to render the phase summary cleanly (or to detect "no remaining Minor/Suggestion" reliably for the auto-approve path), we may need a small consistency edit there. Flagging for exploration during Phase 2 implementation; if needed, surface as a stop condition.
- **Manual-verification flag location** — the empty-summary auto-approve path depends on a way to mark a phase/step as requiring manual verification. The current task-file template may not have an explicit field for this. The Implementer should look for an existing convention (e.g., a Karpathy Contract verification line that names manual testing) before introducing new schema; if a new field is required, that may broaden scope and should trigger a stop-and-confirm.

## Decisions

- **Decision (Orchestrator Accepted):** In Phase 1, the cap-escape was expressed as three named user choices (manual override / accept-without-commit / replan-and-abort) instead of preserving the literal "Proceed anyway" wording inside the cap path. **Reason:** Consistent with Assumption 1 of the plan and the Phase 1 "beneficial local drift" criterion; aligns SKILL.md and `do-workflow.md` wording verbatim.
- **Decision (Orchestrator Accepted):** Step 6b path (1) auto-approve required two follow-up fixes during Phase 2: first to extend coverage to the "fix loop ran successfully" case, then to remove an over-narrow "no accepted Decisions" condition that excluded the legitimate "zero findings + Decisions recorded" state. **Reason:** Both fixes were within Step 2.2's scope and closed real coverage gaps.
- **Decision (Orchestrator Accepted):** During Phase 2 verification, an additional fix was applied to `rules/post-change.md`: a new `## Consolidated End-of-Task Commit` section was added so the SKILL.md MANDATORY CHECKLIST item 2 cross-reference resolves, and `## Update STATE.md` was rewritten to align with the consolidated-commit timing. **Reason:** Phase 2 verification criterion 6 was failing because `post-change.md` did not document the consolidated commit that SKILL.md item 2 was pointing at.
- **Decision (User Accepted):** Future planned item recorded for `--fast` flag — auto-applying Minor/Suggestion findings with clear reviewer recommendations. Not implemented in this task; tracked as a follow-on idea.

## Outcome

Commit `7d24131`. Implemented in two phases:

- **Phase 1 (mandatory fix loop):** Critical/Major review findings now require the fix loop in normal flow — no "Proceed anyway" choice. Iteration cap raised from 3 to 5; on cap-hit the orchestrator offers three explicit user choices (manual override / accept-without-commit / replan-and-abort). Updates in `rules/do-workflow.md` and `skills/crafter-do/SKILL.md` Step 6.
- **Phase 2 (auto-commit + summary gate):** New Step 6b ("Phase Summary and Auto-Commit") in SKILL.md presents a structured phase summary then commits via one of three approval paths: auto-approve on clean summary, silence-approve when `--fast` is set, or explicit user approval (default). Manual-verification override forces explicit approval when the phase requires non-automatable testing. New `## Consolidated End-of-Task Commit` section in `rules/post-change.md`. MANDATORY CHECKLIST and Steps 7–9 reconciled around a single end-of-task follow-up commit covering docs + skillbook + STATE.md. New `--fast` metadata flag declared in SKILL.md (default off) and documented as a speed-vs-explicitness trade-off.

ARCHITECTURE.md updated to reflect the new commit gate semantics and Step 6b. Skillbook entry added for the implementer agent capturing the lesson on covering all end-of-flow states explicitly.
