# Task: orchestrator drops caveman/ponytail skill directives when spawning subagents

## Metadata
- **Date:** 2026-07-03
- **Work branch:** fix/skill-directive-propagation
- **Status:** completed
- **Scope:** Medium

## Request

Source ticket: `BUGFIX-skill-directive-propagation.md` (repo root).

Crafter is designed to propagate active `caveman` / `ponytail` skill state into every spawned subagent via a task-prompt `## Active skill directives` block (`rules/delegation.md` § "Skill Directives — Caveman and Ponytail"). In practice the orchestrator does **not** do it: subagents run without the directive and produce normal verbose prose.

**Root cause (proven):** adherence failure, not missing config. The delegation rule is a soft "remember to re-read the markers and append the directive before every spawn". Nothing enforces it at the spawn site, so in a long `/crafter-do` run the orchestrator loads the rule into context but skips the step when spawning.

**Diagnostic to confirm first:**
1. Verify caveman/ponytail SessionStart hooks reliably write `~/.claude/.caveman-active` / `.ponytail-active` at session start (rule out a missing-marker second bug). If markers unreliable → separate finding, do not silently fix here.
2. Trace which `rules/do/step-*.md` steps actually spawn Task subagents, and whether any restate the directive requirement at the spawn site or rely solely on the distant `rules/delegation.md` section.

**Goal:** make skill-directive injection reliable — whenever a marker is present at spawn time, the matching `## Active skill directives` block MUST be appended per existing audience rules (caveman-full agent-facing / caveman-lite reviewer+verifier; ponytail only implementer+planner). No dependence on the LLM remembering a distant rule. Prefer the smallest change that removes the adherence gap.

**Candidate directions (planner picks):** bake the directive requirement into the task-prompt template/checklist at each `rules/do/step-*.md` spawn site; or a hard self-check gate immediately before every Task spawn; a file-driven hook only if prompt-level cannot be made reliable (heavier, must ship via installer).

**Scope:** repo `/Users/ret/dev/ai/crafter`. Likely files: `rules/delegation.md`, `rules/do/step-*.md`, possibly `rules/core.md`. Agent defs already carry "Behavior under ponytail" — extend only if needed for symmetry.

**Non-goals:** do NOT touch caveman/ponytail plugins; no git-committed repo hook; do NOT change the audience-based level policy; do NOT rely on the plugins' `SubagentStart` hook as the fix.

**Acceptance:**
- Dry run / trace shows: markers present ⇒ every spawned subagent's task prompt contains the correct `## Active skill directives` block with the right level per audience; markers absent ⇒ nothing appended, no mention of the skills (byte-for-byte unchanged).
- Reliability no longer depends on the orchestrator recalling a distant rule.

## Plan

**Plan status:** approved

### Trace correction (read before implementing)

The ticket's candidate direction (a) points at `rules/do/step-*.md` as the spawn sites. **That is wrong** — those modules are read *internally by the delegated agents* (each spawn instruction says "The agent internally reads `{CRAFTER_HOME}/rules/do/step-N.md`"). The orchestrator never reads them at spawn time. The actual runtime spawn sites — the 11 `Spawn the **\`crafter-<agent>\`** agent` instructions the orchestrator itself executes — all live in **`skills/crafter-do/SKILL.md`**:

| Line | Agent | Audience → directive level |
|------|-------|----------------------------|
| 103 | `crafter-step-runner` (extension-skills) | caveman-full; no ponytail |
| 171 | `crafter-step-runner` (step-0-resume) | caveman-full; no ponytail |
| 181 | `crafter-step-runner` (step-1-scope) | caveman-full; no ponytail |
| 189 | `crafter-analyzer` (step 2) | caveman-full; no ponytail |
| 195 | `crafter-planner` (step 3) | caveman-full; **ponytail** |
| 201 | `crafter-planner` (re-spawn on plan changes) | caveman-full; **ponytail** |
| 209 | `crafter-implementer` (step 4) | caveman-full; **ponytail** |
| 219 | `crafter-verifier` (step 5 drift) | caveman-**lite**; no ponytail |
| 231 | `crafter-verifier` (step 5a phase verif) | caveman-**lite**; no ponytail |
| 239 | `crafter-reviewer` (step 6) | caveman-**lite**; no ponytail |
| 265 | `crafter-implementer` (step 6 fix loop) | caveman-full; **ponytail** |
| 327 | `crafter-implementer` (steps 7–9 ARCHITECTURE.md check) | caveman-full; **ponytail** |

(12 rows — line 201 is a re-spawn narrated in the orchestrator-residue block; treat it as a spawn site too.)

This is the root cause: the enforcement gap is at the SKILL.md spawn sites, and the fix must remove that gap **once, where all spawns route through** — not by editing the agent-internal step modules.

### Approach

The existing `rules/delegation.md` §"Skill Directives — Caveman and Ponytail" already holds the correct, complete rule (audience policy, exact directive block text, no-op behavior). Do **not** rewrite that rule — it is correct; it is just distant and unenforced. The structural parallel is the Skillbook rule (`rules/delegation.md` §"Skillbook — Learned Guidelines"), which suffers the identical "distant soft reminder" shape.

The smallest change that removes the adherence gap: add a **hard pre-spawn gate at the spawn site** that the orchestrator physically reads at each `Spawn the ...` instruction, pointing back to the delegation rule as the source of truth. Prefer a single shared gate injected once near the spawn-site cluster / master-plan map, plus a short per-spawn marker, over duplicating the full directive block at all 12 sites (duplication is drift-prone and violates the ladder). The Implementer chooses the exact mechanism within these boundaries; the requirement is that **the orchestrator cannot reach a `Task` spawn without a co-located instruction to apply the delegation §"Skill Directives" rule**.

Ponytail note: this is prompt-only Markdown in `skills/crafter-do/SKILL.md` (and only if strictly needed, a one-line cross-reference tweak-spot in `rules/delegation.md`). No new files, no CLI, no installer hook — candidate (c) is explicitly ruled out because prompt-level enforcement is reliable here (same mechanism the Skillbook rule already relies on). Skipped: a file-driven `SubagentStart` installer hook — add only if a dry-run trace proves prompt-level enforcement still drops directives, which is not expected.

---

### Phase 1 — Spawn-site enforcement gate in the orchestrator

**Phase outcome:** every `Spawn the ...` instruction in `skills/crafter-do/SKILL.md` is co-located with an enforceable directive-injection gate, so the orchestrator applies `rules/delegation.md` §"Skill Directives" at each spawn without recalling a distant rule. Markers absent ⇒ byte-for-byte unchanged behavior.

**Phase scope boundary:** `skills/crafter-do/SKILL.md` only (plus, if needed, a one-line cross-reference sharpening in `rules/delegation.md`). No changes to the audience policy, the directive block text, `rules/core.md`, agent files, CLI, or installer.

**Phase non-goals:** rewriting the delegation §"Skill Directives" rule; changing audience-based levels; editing `rules/do/step-*.md`; touching plugins or plugin hooks.

**Phase simplicity constraint:** shortest working diff. One shared gate + light per-spawn markers, not 12 copies of the directive block.

**Phase verification (must pass before phase closes):** a manual dry-run trace over all 12 spawn sites (see Verification section) shows each site now routes through the gate, with correct audience level per the table above.

- [x] **Step 1.1 — Add a shared pre-spawn directive gate to SKILL.md.**
  - Outcome: SKILL.md contains a single, unmissable pre-delegation gate (co-located with the spawn-site cluster — e.g. adjacent to where the master-plan map or the first spawn instruction lives) that instructs the orchestrator, before *every* `Task` spawn, to apply `rules/delegation.md` §"Skill Directives — Caveman and Ponytail": re-read the two markers fresh, and append the `## Active skill directives` block per the audience policy (or append nothing when both markers are absent). The gate references the delegation rule as source of truth — it does **not** restate the block text or the audience table.
  - Scope boundary: adding the gate text; no change to any individual spawn instruction's payload description yet.
  - Non-goals: duplicating the directive block; changing the delegation rule; per-agent level lists inline.
  - Simplicity constraint: one gate block, prose only, no new headings that fragment the workflow map. Model it on the existing "Skillbook — Learned Guidelines" enforcement shape.
  - Drift criteria (STOP if): the gate copies the full directive block text; the gate restates or alters the audience policy; the gate lands in `rules/do/step-*.md` instead of SKILL.md; the change touches marker-writing or plugins.
  - Verification evidence: the gate is present in SKILL.md and cross-references `rules/delegation.md` §"Skill Directives"; a Grep for the gate anchor resolves; reading from any spawn instruction back-references the gate.
  - Stop condition: gate added and cross-reference resolves — stop; do not also rewrite the delegation rule.

- [x] **Step 1.2 — Attach a per-spawn directive marker at each of the 12 spawn sites.**
  - Outcome: each `Spawn the ...` instruction (lines per the trace table) carries a short, co-located marker naming the audience level(s) that apply to that specific agent (caveman-full vs caveman-lite; ponytail yes/no), pointing at the gate/delegation rule for the block text. This makes the correct level unmissable *at the point of spawn*, closing the "orchestrator forgot which level" failure mode without duplicating block text.
  - Scope boundary: the 12 spawn instructions in SKILL.md only. The marker names the level and defers block text to the rule.
  - Non-goals: inlining the full `## Active skill directives` block at each site; adding markers to non-spawn prose; changing any spawn's functional payload (step id, passed context).
  - Simplicity constraint: a single short line/clause per spawn site, consistent wording. Reviewer/verifier sites get caveman-lite + no-ponytail; implementer/planner sites get caveman-full + ponytail; analyzer/step-runner sites get caveman-full + no-ponytail — matching the trace table and the existing audience policy exactly.
  - Drift criteria (STOP if): any site is assigned a level that contradicts the trace table / delegation audience policy; the re-spawn at line 201 or the ARCHITECTURE.md-check spawn at line 327 is missed; a marker restates the full block text; a marker changes the audience policy rather than referencing it.
  - Verification evidence: all 12 sites carry a marker; each marker's level matches the trace table; Grep count of markers == count of spawn instructions.
  - Stop condition: all 12 sites marked with correct levels — stop; do not extend markers into agent step modules.

- [x] **Step 1.3 — (Conditional) sharpen the delegation rule's cross-reference.** — SKIPPED (not needed). `rules/delegation.md` §"Skill Directives" already reads as mandatory ("Before spawning any agent... re-read the markers immediately before spawning"); enforcement now lives at the SKILL.md spawn-site gate, and delegation.md is not read at spawn time either, so a back-cross-reference adds no enforcement. YAGNI.
  - Outcome: only if Step 1.1/1.2 revealed that `rules/delegation.md` §"Skill Directives" is ambiguous about being a spawn-site gate (e.g. it reads as advisory), add at most a one-line note pointing back to the SKILL.md gate so the two artifacts reinforce each other. If the rule already reads as mandatory, skip this step and record "not needed" — YAGNI.
  - Scope boundary: at most one line in `rules/delegation.md`; the block text and audience table stay verbatim.
  - Non-goals: rewriting the rule; changing levels or block text; adding a second copy of the policy.
  - Simplicity constraint: one line max, or skip entirely.
  - Drift criteria (STOP if): the edit grows beyond one cross-reference line; the block text or audience policy changes.
  - Verification evidence: either a one-line cross-reference exists, or a task-file note records the step was skipped as unnecessary.
  - Stop condition: cross-reference added or explicitly skipped — stop.

---

### Verification (acceptance)

Prompt-only change ⇒ verification is a **manual dry-run trace**, not an automated test (no code path executes markdown). Required evidence for phase close:

1. **Markers-present trace (positive):** Walk all 12 spawn sites in `skills/crafter-do/SKILL.md`. For each, confirm the orchestrator, reading top-to-bottom, hits the shared gate (Step 1.1) and the per-spawn marker (Step 1.2) *before* the `Task` invocation, and that the marker's level matches the trace table:
   - caveman-full + ponytail: lines 195, 201, 209, 265, 327 (planner + implementer sites)
   - caveman-full + no ponytail: lines 103, 171, 181, 189 (step-runner + analyzer sites)
   - caveman-lite + no ponytail: lines 219, 231, 239 (verifier + reviewer sites)
2. **Markers-absent trace (negative / no-op):** Confirm the gate and markers instruct "append nothing, no mention of skills" when both markers are absent — behavior byte-for-byte unchanged, satisfying `rules/core.md` §"Skill Detection" no-op guarantee.
3. **Diagnostic #1 (verify-only, no code change):** Confirm the `.caveman-active`/`.ponytail-active` markers are written by the external caveman/ponytail plugins, NOT by crafter (Grep of `cli/`, `install.sh`, `rules/` finds no crafter-owned marker writer). This session's live `SubagentStart` context (`PONYTAIL MODE ACTIVE — level: full`) already proves the plugin hooks fire. Record as a finding: marker reliability is out of crafter's scope; no second bug in-repo. If a future run shows missing markers, that is a plugin-side finding, filed separately.
4. **No-scope-leak check:** `git diff` touches only `skills/crafter-do/SKILL.md` (and at most one line of `rules/delegation.md`). No changes to plugins, CLI, installer, `rules/core.md`, agent files, or `rules/do/step-*.md`.

---

### Assumptions / interpretations

- **A1 (load-bearing):** The runtime spawn sites are in `skills/crafter-do/SKILL.md`, and `rules/do/step-*.md` are agent-internal (confirmed by the "agent internally reads" phrasing at every spawn). The fix targets SKILL.md, overriding the ticket's candidate-(a) file pointer.
- **A2:** A prompt-level gate is sufficient (same enforcement class as the already-working Skillbook rule). Installer/hook direction (c) stays ruled out unless a dry-run proves prompt-level insufficient.
- **A3:** The re-spawn at line 201 and the ARCHITECTURE.md-check implementer spawn at line 327 are genuine spawn sites and must be covered (12 total, not 10).
- **A4:** The two agent files already carrying "Behavior under ponytail" (`crafter-implementer.md:74`, `crafter-planner.md:67`) need no change — the directive is injected orchestrator-side per the audience policy; agent-side symmetry already exists.

### Non-goals (hard, restated)

- Do not touch caveman/ponytail plugins or rely on their `SubagentStart` hook as the fix.
- No git-committed repo hook in any target project; no installer/CLI hook.
- Do not change the audience-based level policy or the directive block text.
- Do not edit `rules/do/step-*.md` (agent-internal, not spawn sites).
- Do not add markers to `rules/core.md` §"Skill Detection" — reference only.

### Alternatives considered (Medium scope)

- **Duplicate the full `## Active skill directives` block at each of the 12 sites.** Rejected: drift-prone (12 copies of block text + audience logic to keep in sync), violates the ladder. A shared gate + level marker gives the same enforcement with a fraction of the diff.
- **Edit `rules/do/step-*.md` per ticket candidate (a).** Rejected: those files are read by delegated agents, not the orchestrator at spawn time — editing them would not reach the spawn decision point. Would leave the bug unfixed while appearing addressed.
- **File-driven `SubagentStart` installer hook (candidate c).** Rejected as heavier and out of scope: requires CLI/installer work and a target-project hook the non-goals forbid; prompt-level enforcement is reliable here. Reconsider only if the dry-run trace proves prompt enforcement still drops directives.
- **Rewrite the delegation §"Skill Directives" rule to be "stronger".** Rejected: the rule content is already correct; strengthening distant prose does not fix a spawn-site adherence gap. Enforcement must live where the spawn happens.

### Risks / unknowns / flags

- **R1 (low):** Prompt-level enforcement is still LLM adherence, not a hard runtime guard — a co-located gate is materially more reliable than a distant rule (the proven failure mode) but cannot be *proven* 100% without runtime instrumentation crafter doesn't have. Mitigation: co-locate at the spawn point + per-site level marker. If a real run still shows drops, escalate to candidate (c). This is the residual risk the contract accepts.
- **R2 (verify-only):** If markers turn out unreliable in some environment, that is a plugin-side second bug outside this task's scope (diagnostic #1) — file separately, do not fix here.
- **R3:** Line numbers in the trace table are snapshot-current; the Implementer must match spawn instructions by content (`Spawn the **\`crafter-...\`** agent`), not by line number, since edits shift lines.

## Decisions

- **Decision:** Fix targets `skills/crafter-do/SKILL.md`, NOT `rules/do/step-*.md` as the ticket's candidate (a) suggested. **Reason:** the step modules are read *internally by delegated agents* ("The agent internally reads..."), never by the orchestrator at spawn time. All spawn instructions the orchestrator executes live in SKILL.md — that is the actual enforcement gap.
- **Decision:** Step 1.3 (delegation.md cross-reference) SKIPPED as unnecessary. **Reason:** `rules/delegation.md` §"Skill Directives" already reads as mandatory; enforcement now lives at the SKILL.md gate, and delegation.md is not read at spawn time either, so a back-cross-reference adds no enforcement. YAGNI.
- **Decision (Reviewer-Fixed):** Major — Step 5 "fix current step" and Step 5a "re-delegate to the Implementer" are implementer re-delegations phrased as prose (not "Spawn the ..."), originally missing level markers (nearest marker was the wrong lite/no-ponytail verifier level). Added co-located caveman-full; ponytail markers at both; broadened the gate scope clause to cover prose re-delegations. 14 markers total.
- **Decision (Reviewer-Fixed):** Suggestion — normalized the two prose-re-delegation markers from `caveman-full + ponytail` to `caveman-full; ponytail` for wording consistency with the other 12.
- **Decision:** Diagnostic #1 confirmed — `.caveman-active`/`.ponytail-active` markers are written by the external caveman/ponytail plugins, not crafter (no marker-writer in `cli/`, `install.sh`, `rules/`, `hooks/`). Marker reliability is out of crafter's scope; no second in-repo bug.

## Outcome

**Done.** Closed the adherence gap where the orchestrator skipped injecting the caveman/ponytail `## Active skill directives` block into spawned subagents' task prompts.

**Change (prompt-only Markdown, `skills/crafter-do/SKILL.md` only):**
- Added a `## Pre-Spawn Gate — Skill Directives` section above the first spawn site (`## Extension Skills`), instructing the orchestrator before EVERY `Task` spawn — including prose re-delegations, not only "Spawn the ..." instructions — to apply `rules/delegation.md` §"Skill Directives — Caveman and Ponytail" (fresh marker re-read; append the block per audience policy, or nothing when both markers absent). Gate references the rule as source of truth — no directive-block or audience-table duplication.
- Added a co-located level marker at all **14** Task-spawn / re-delegation paths: 7× `caveman-full; ponytail` (planner ×2, implementer ×5 incl. the two prose re-delegations at Step 5 / Step 5a), 4× `caveman-full; no-ponytail` (step-runner ×3, analyzer), 3× `caveman-lite; no-ponytail` (verifier ×2, reviewer).

**Invariant preserved:** markers absent ⇒ nothing appended, no mention of the skills (byte-for-byte unchanged).

**Verification:** phase verification 6/6 PASS (after fixing a gate-ordering defect where the gate initially sat after the first spawn site); review closed clean (2 Major fixed in iteration 1, 1 Suggestion fixed). Acceptance dry-run trace confirmed each of the 14 sites routes through the gate with the correct audience level.

**Residual risk (accepted):** prompt-level enforcement is still LLM adherence, not a hard runtime guard — co-location at the spawn point is materially more reliable than the prior distant rule, but not provably 100%. Escalate to a file-driven installer hook (candidate c) only if a future real run still drops directives.

**Commit:** `304711f` (SKILL.md phase change) + consolidated end-of-task commit (task file + STATE.md).
