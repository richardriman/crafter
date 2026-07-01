# Task: Caveman/ponytail skill detection support in Crafter agents

## Metadata
- **Date:** 2026-07-01
- **Work branch:** feat/caveman-ponytail-detect
- **Status:** completed
- **Scope:** Medium

## Request
> Chci přidat podporu skillů caveman a ponytail na bázi detekce, zda jsou přítomny. U cavemana pak ještě rozlišovat, kdy se komunikuje s člověkem, tam použít režim lite, pro reasoning a čistě agent-facing věci pak caveman full.

## Plan
**Plan status:** approved

### 1. Complete request

Teach Crafter's orchestrator and agents to detect whether the external Claude Code skills **caveman** (terse communication compression) and **ponytail** (YAGNI-first code discipline) are active in the current session, and to adapt behavior accordingly:

- **Detection** is deterministic and performed orchestrator-side (so subagents — which never receive the skills' SessionStart hook injection — can still be told). The canonical signal is the presence of the marker files `$HOME/.claude/.caveman-active` and `$HOME/.claude/.ponytail-active`; each file's content is the configured level (`lite` / `full` / `ultra`). Presence = the skill is active for this session.
- **Caveman** adaptation is audience-based: **lite** for human-facing prose (the orchestrator talking to the user) and **full** for reasoning and pure agent-facing output (inter-agent summaries, structured reports returned to the orchestrator). Caveman must never touch commits, PR titles/bodies, or release notes (those stay neutral English per `CLAUDE.md`), must never mangle content the orchestrator relays verbatim (Reviewer tables, scorecards), must honor caveman's own Auto-Clarity carve-outs (security warnings, irreversible-action confirmations, multi-step sequences write normally), and must not weaken the Jargon Confinement guardrail.
- **Ponytail** adaptation applies **only** to the two code-writing/planning agents — `crafter-implementer` and `crafter-planner` — as YAGNI / the-ladder / shortest-working-diff discipline. It does not apply to reviewer, verifier, analyzer, or step-runner.

This matters because Crafter users who run these skills expect the whole session — including delegated agents — to honor the discipline they turned on; today the subagents silently ignore it because they run in fresh contexts without the SessionStart injection.

**Scope:** Medium, purely additive prompt-engineering across Markdown skill/rule/agent files. No Go/CLI changes. No installer changes.

**Acceptance criteria:**
1. When neither marker is present, behavior is byte-for-byte unchanged from today (no regression, no mention of the skills).
2. When `.caveman-active` is present, the orchestrator's user-facing prose is compressed (caveman-lite) with all carve-outs respected, and every delegated agent is signaled to produce caveman-full reasoning/reports.
3. When `.ponytail-active` is present, only `crafter-implementer` and `crafter-planner` receive the ponytail discipline directive; other agents are unaffected.
4. Commits, PR bodies, release notes, and verbatim-relayed tables remain untouched by caveman regardless of marker state.

**Validation strategy:** This is Markdown prompt engineering, so verification is documentation review for internal consistency (correct marker paths, complete carve-out enumeration, correct ponytail scoping, DRY definitions), plus an optional manual smoke test running `/crafter-do` with markers present vs. absent. There is no automated test target for prompt files (`tests/test_install.sh` covers only the installer).

### 2. Assumptions / interpretations

- **A1 — Marker files are the canonical signal.** `$HOME/.claude/.caveman-active` / `.ponytail-active` exist only while the skill is toggled on and contain the level as text. I treat presence as the activation trigger. (Confirmed by inspecting the live install: both files exist and contain `full`.)
- **A2 — Caveman level is chosen by audience, not by the marker's value (settled decision).** The orchestrator applies lite to human-facing prose and full to agent-facing output regardless of whether the marker says `lite`, `full`, or `ultra`. The marker's specific level is presence-detection only for caveman; it is NOT a ceiling. Audience always wins.
- **A3 — Ponytail level is passed through.** For ponytail, the marker's level (`lite`/`full`/`ultra`) is forwarded to the implementer/planner directive so the user's configured intensity is honored, since the settled decision gives no audience split for ponytail.
- **A4 — Reuse the existing injection idiom.** The orchestrator already appends skillbook "Learned Guidelines" to agent task prompts (`rules/delegation.md`) and already signals `--auto` in task prompts with matching `## Behavior under --auto` sections in agent files. The adaptation directive should follow this exact established pattern rather than inventing a new mechanism.
- **A5 — Shared definitions + detection live in `rules/core.md` / `rules/delegation.md` (settled decision).** Both orchestrators (`crafter-do`, `crafter-debug`) load `core.md` and `delegation.md`. The caveman/ponytail semantics, detection procedure, and human-facing caveman policy live in these shared rules (phrased generically for "the orchestrator"), not inlined only in `skills/crafter-do/SKILL.md`. This is orchestrator-side, DRY, and gives `crafter-debug` parity for free.
- **A6 — `HOME`/`.claude` resolution.** The global `$HOME/.claude/` markers are canonical. A project-local `.claude/.caveman-active` override is out of scope unless it already exists in practice (it does not, per inspection).

### 3. Non-goals

- No changes to the Go CLI, `install.sh`, statusline, or any binary.
- No attempt to *install*, *enable*, *disable*, or *configure* caveman or ponytail — Crafter only reacts to their presence.
- No caveman/ponytail application to commits, PR titles/bodies, or release notes — those remain neutral English (`CLAUDE.md`), full stop.
- No new levels, no Crafter-specific caveman/ponytail sub-modes, no user-facing flags to force the modes on/off (detection only).
- No reformatting of Reviewer/Verifier structured tables — verbatim relay stays verbatim.
- No change to which agents exist or their model tiers.

### 4. Relevant areas

- `skills/crafter-do/SKILL.md` — orchestrator entry; detection ordering and the human-facing/user-communication surface. Note the many agent-spawn points (Steps 1–6, step-runner) and the verbatim-relay instruction at Step 6a.
- `rules/core.md` — Language Rules, Jargon Confinement, Karpathy guardrails; the natural home for the shared caveman/ponytail semantics and (per A5) the detection + human-facing caveman policy.
- `rules/delegation.md` — the single pre-spawn injection point (skillbook pattern at lines ~41–54); the natural home for one "before spawning any agent, append the adaptation directive" rule.
- `agents/crafter-implementer.md`, `agents/crafter-planner.md` — get both a caveman and a ponytail behavior section (mirror the existing `## Behavior under --auto` sections).
- `agents/crafter-analyzer.md`, `agents/crafter-reviewer.md`, `agents/crafter-verifier.md`, `agents/crafter-step-runner.md` — get a caveman behavior section only.
- `CLAUDE.md` — the neutral-voice / no-signature policy caveman must not violate (reference only; not edited).
- `skills/crafter-debug/SKILL.md` — second orchestrator, in scope for human-facing caveman-lite parity. It already reads `core.md` at startup (before its Step 1 user dialog), so shared detection runs automatically; its user-communication surface is Steps 1–4 (symptom collection, hypothesis presentation, fix proposal). Wire it via the generic `core.md` policy plus a short pointer only if an explicit detection-ordering cue is warranted.

### 5. Vertical phases and steps

---

#### Phase 1 — Orchestrator detects the skills and adapts its own voice

Delivers a coherent, independently reviewable slice: with the markers present, the orchestrator detects them and speaks to the user in caveman-lite (with all carve-outs); with them absent, nothing changes. Agent-facing propagation comes in Phase 2. The shared semantic definitions written here are consumed by Phase 2.

- [x] **Step 1.1 — Detection procedure + shared skill semantics (in shared rules).** Add an orchestrator-side detection procedure to the shared rules (`rules/core.md`, per settled decision A5) that checks for `$HOME/.claude/.caveman-active` and `$HOME/.claude/.ponytail-active`, records each skill's active state and level, and establishes these as workflow context available at every delegation point. Add a canonical definition block describing what caveman-lite vs caveman-full mean, that caveman selection is audience-based and presence-only (A2), and that ponytail's level is passed through (A3). Phrase the block generically for "the orchestrator" so both `crafter-do` and `crafter-debug` inherit it. No behavior change yet beyond the orchestrator now *knowing* the state.
  - **Outcome:** Detection runs deterministically at orchestrator startup (both orchestrators, since both load `core.md`); a single shared definition of the two disciplines exists and is referenceable by later steps.
  - **Scope boundary:** Marker reads + shared definitions in `core.md`/`delegation.md` only. No changes to how the orchestrator talks or how agents are spawned yet.
  - **Non-goals:** No agent-prompt injection; no user-prose compression; no CLI.
  - **Simplicity constraint:** Two file-existence checks and one content read per marker; reuse existing rule-file structure; no helper scripts, no new CLI subcommand.
  - **Drift criteria:** FLAG if detection touches anything other than the two named marker paths; if it tries to enable/disable the skills; if it invents levels beyond lite/full/ultra; if definitions are duplicated instead of centralized; if the block is phrased for `crafter-do` only in a way that excludes `crafter-debug`.
  - **Verification evidence:** The detection text names the exact two paths and the lite/full/ultra level semantics; the shared definition block is present, self-consistent, and generic across both orchestrators; absent-marker path explicitly yields unchanged behavior.
  - **Stop conditions:** Stop if a spawn or startup path is found that does not load `core.md` (would break inherited detection) — report before broadening scope.

- [x] **Step 1.2 — Human-facing caveman-lite with carve-outs.** Specify that when caveman is active, the orchestrator compresses its own user-facing prose to caveman-lite, and enumerate the carve-outs explicitly: never compress or alter (a) content relayed verbatim (Reviewer diff/issues/scorecard tables, Verifier reports the orchestrator reproduces), (b) plan-approval and other human-in-the-loop gate prompts to the point of losing clarity, (c) security warnings / irreversible-action confirmations / multi-step sequences (caveman Auto-Clarity), (d) commit messages, PR titles/bodies, release notes (stay neutral English per `CLAUDE.md`), (e) persistent-file English (`.crafter/*`, plans, task files per `core.md` Language Rules). Reaffirm Jargon Confinement is unchanged.
  - **Outcome:** With `.caveman-active` present, the orchestrator's conversational output is terser (lite) while every carve-out is provably preserved; with it absent, prose is unchanged.
  - **Scope boundary:** Orchestrator's own user-facing/human-communication prose only.
  - **Non-goals:** No agent output changes; no ponytail here; no touching verbatim relay content or persisted artifacts.
  - **Simplicity constraint:** A single policy paragraph plus an explicit carve-out list; no per-message conditionals scattered across every step.
  - **Drift criteria:** FLAG if any carve-out (a)–(e) is omitted or weakened; if caveman is applied to commits/PRs/release notes; if verbatim relay is reformatted; if the change alters behavior when the marker is absent.
  - **Verification evidence:** All five carve-out categories present and cross-referenced to `CLAUDE.md` and `core.md` Language Rules; the Step 6a verbatim-relay instruction remains authoritative over caveman.
  - **Stop conditions:** Stop if honoring caveman-lite would require restructuring the human-in-the-loop gate wording in a way that could reduce approval clarity — flag for user decision.

- [x] **Step 1.3 — crafter-debug human-facing caveman-lite parity.** Ensure `crafter-debug` honors the same human-facing caveman-lite policy at its user-communication surface (symptom collection, hypothesis presentation, fix proposal, verification reporting — Steps 1–4 and the final report). Because the human-facing policy and detection live in shared `core.md` (Step 1.1/1.2) phrased generically for "the orchestrator," and `crafter-debug/SKILL.md` already reads `core.md` at startup before its Step 1 dialog, parity should fall out of the shared rule with detection ordering already satisfied. This step's job is to confirm that inheritance actually holds for `crafter-debug`, and to add a short pointer in `skills/crafter-debug/SKILL.md` only if an explicit detection-ordering or caveman-surface cue is needed for clarity (mirroring whatever ordering cue `crafter-do` gets). Keep it lean — do not duplicate the policy text into `crafter-debug`.
  - **Outcome:** With `.caveman-active` present, `crafter-debug`'s user-facing dialog and reports are caveman-lite with all carve-outs preserved, identical in policy to `crafter-do`; with it absent, unchanged.
  - **Scope boundary:** `skills/crafter-debug/SKILL.md` only, and only a pointer/ordering cue if needed — the substantive policy stays in shared `core.md`.
  - **Non-goals:** No re-statement of the caveman policy in `crafter-debug`; no changes to `crafter-debug`'s debugging workflow logic; no ponytail wiring here (agent-facing propagation is shared via Phase 2).
  - **Simplicity constraint:** Prefer zero new prose in `crafter-debug` if inheritance is clean; at most one short pointer line. No duplicated carve-out list.
  - **Drift criteria:** FLAG if the caveman policy is copied into `crafter-debug` instead of referenced; if `crafter-debug`'s workflow logic is altered; if the change affects absent-marker behavior.
  - **Verification evidence:** Reading `crafter-debug/SKILL.md` plus shared `core.md`, a reviewer can trace that its user-facing surface is governed by the same caveman-lite policy and that detection runs (via `core.md` load) before its first user-facing output.
  - **Stop conditions:** Stop if `crafter-debug` turns out to bypass `core.md` for any user-facing output path, or if its mandated report structure conflicts with caveman-lite — flag rather than force.

**Phase 1 verification:** Reading the edited rules/skills, a reviewer can confirm: detection targets exactly the two marker paths and lives in shared `core.md`; absent-marker behavior is unchanged; caveman-lite applies to both orchestrators' (`crafter-do` and `crafter-debug`) human-facing prose with all five carve-outs enumerated; the policy is defined once and inherited (not duplicated into `crafter-debug`); no commit/PR/verbatim content is affected.

---

#### Phase 2 — Agents honor the disciplines when delegated

Extends detection into the delegated agents: caveman-full for every agent's reasoning and returned reports, ponytail for the two code/plan agents only. Builds directly on the Phase 1 detection state and shared definitions.

- [x] **Step 2.1 — Single pre-spawn propagation rule.** In `rules/delegation.md` (alongside the existing skillbook injection rule), add one rule: before spawning any agent, if caveman is active, append a compact caveman-full directive to that agent's task prompt (all agents); if ponytail is active, additionally append a ponytail directive **only** when the target agent is `crafter-implementer` or `crafter-planner`, carrying the passed-through level (A3). This is the sole propagation mechanism — no per-step edits in `SKILL.md`.
  - **Outcome:** Every agent spawn deterministically carries the correct discipline directive(s) based on detection, with ponytail correctly narrowed to two agents.
  - **Scope boundary:** One additive rule in `delegation.md`. Reuses the skillbook idiom.
  - **Non-goals:** No editing of individual Step 1–6 spawn instructions; no ponytail for reviewer/verifier/analyzer/step-runner; no caveman on commits.
  - **Simplicity constraint:** One rule, one conditional per skill, an explicit two-agent allowlist for ponytail; no new injection channel.
  - **Drift criteria:** FLAG if ponytail leaks to any agent outside the two-agent allowlist; if the rule fires when markers are absent; if it duplicates the caveman semantics instead of referencing the Phase 1 definition.
  - **Verification evidence:** The rule names the exact allowlist (`crafter-implementer`, `crafter-planner`) for ponytail and "all agents" for caveman; references the shared definition; is a no-op when inactive.
  - **Stop conditions:** Stop if a spawn point is discovered that bypasses `delegation.md` (would need its own handling) — report before broadening scope.

- [x] **Step 2.2 — Caveman behavior section in all six agent files.** Add a concise, self-contained `## Behavior under caveman` section to each of the six agent files, mirroring the existing `## Behavior under --auto` structure: when the task prompt signals caveman-full, compress reasoning and the returned report per caveman discipline (drop articles/filler/hedging, keep all technical substance), but keep tables, code, file paths, and identifiers verbatim, and never compress across the Auto-Clarity carve-outs. Agent sections are self-contained because agents do not load `core.md`.
  - **Outcome:** Any agent, when signaled, returns caveman-full output without losing technical substance or structured tables.
  - **Scope boundary:** One new section per agent file; no change to existing role/constraints/output-format sections beyond the addition.
  - **Non-goals:** No ponytail content here; no change to the required output-format tables/fields; no behavior when the signal is absent.
  - **Simplicity constraint:** One short section per file, uniform wording; no per-agent bespoke caveman rules beyond what each agent's output already contains.
  - **Drift criteria:** FLAG if a caveman section instructs an agent to alter its mandated output-format structure (e.g., collapse the Reviewer scorecard), or to compress security/irreversible/multi-step content.
  - **Verification evidence:** All six agent files contain the section; each preserves that file's existing required output structure; wording is consistent across files.
  - **Stop conditions:** Stop if honoring caveman would conflict with an agent's mandated output format (e.g., Verifier's fixed summary line) — flag rather than override the format.

- [x] **Step 2.3 — Ponytail behavior section in implementer and planner only.** Add a `## Behavior under ponytail` section to `crafter-implementer.md` and `crafter-planner.md` only: when the task prompt signals ponytail (at the passed level), apply the-ladder / YAGNI / shortest-working-diff discipline to code and plans, while never simplifying away input validation, error handling, security, accessibility, or explicitly requested behavior. Confirm no ponytail section is added to the other four agents.
  - **Outcome:** The two code/plan agents honor ponytail discipline when signaled; the other four remain unaffected.
  - **Scope boundary:** Exactly two agent files edited for ponytail.
  - **Non-goals:** No ponytail in reviewer/verifier/analyzer/step-runner; no caveman content; no weakening of the "when NOT to be lazy" safety carve-outs.
  - **Simplicity constraint:** One short section in each of exactly two files; reference ponytail's own discipline rather than re-deriving it.
  - **Drift criteria:** FLAG if a ponytail section appears in any of the other four agent files; if it omits the safety carve-outs (validation/error-handling/security/accessibility/explicit requests).
  - **Verification evidence:** Only `crafter-implementer.md` and `crafter-planner.md` contain the ponytail section; the safety carve-outs are present; the other four files have no ponytail text.
  - **Stop conditions:** Stop if ponytail discipline would appear to conflict with the Implementer's existing "smallest correct change" constraint in a contradictory way — reconcile in wording, or flag.

**Phase 2 verification:** A reviewer can confirm: the propagation rule fires only when active and scopes ponytail to exactly two agents; all six agent files carry a consistent caveman section that preserves mandated output formats; exactly two agent files carry the ponytail section with safety carve-outs intact; nothing changes when both markers are absent.

### 6. Karpathy Contract (summary)

- **Outcome (whole task):** Crafter detects active caveman/ponytail via two marker files and adapts — caveman-lite for the orchestrator's human-facing prose, caveman-full for all agent reasoning/reports, ponytail discipline for the implementer and planner only — with commits/PRs/release notes/verbatim relays and safety carve-outs fully protected.
- **Scope boundary:** Additive Markdown edits to `rules/core.md`, `rules/delegation.md`, `skills/crafter-do/SKILL.md` and `skills/crafter-debug/SKILL.md` (as needed for ordering/parity), and the six `agents/*.md` files. No CLI, installer, or non-Markdown changes.
- **Non-goals:** Installing/toggling the skills; touching commits/PRs/release notes; reformatting structured relay; expanding ponytail beyond two agents; changing absent-marker behavior.
- **Simplicity constraint:** Reuse the existing skillbook/`--auto` injection idiom; one detection procedure, one propagation rule, centralized definitions, uniform per-agent sections. No new mechanisms, scripts, or CLI subcommands.
- **Drift criteria:** Any behavior change when markers are absent; ponytail leaking beyond the two-agent allowlist; caveman touching commits/PRs/release notes or verbatim tables; dropped carve-outs; duplicated (non-DRY) definitions.
- **Verification evidence:** Documentation review confirming correct paths, complete carve-out lists, correct scoping, DRY definitions; optional manual smoke of `/crafter-do` with markers present vs absent.
- **Stop conditions:** Discovery of an agent spawn path that bypasses `delegation.md`, or a user-facing output path (in either orchestrator) that bypasses `core.md`; a mandated agent output-format that caveman cannot honor without restructuring.

### 7. Alternatives considered

- **Inject the full behavioral spec into every agent task prompt (no agent-file edits).** Leanest in file count, but the orchestrator would have to assemble variable spec text at every spawn, and agents invoked directly via `/agents` would never adapt. Rejected in favor of the established crafter idiom (signal in task prompt + `## Behavior under X` section in agent files), which is consistent with `--auto` and more discoverable/durable.
- **Detect via plugin-install registry (`~/.claude/plugins/installed_plugins.json`).** Rejected: install ≠ active. The `.X-active` markers mean the skill is actually toggled on for the session, which is the correct semantic and is a single cheap file check.
- **Put detection + human-facing caveman only in `skills/crafter-do/SKILL.md`.** Rejected (settled decision A5): it duplicates definitions and denies `crafter-debug` parity. Shared `core.md`/`delegation.md` placement is DRY, orchestrator-side, and gives `crafter-debug` human-facing and agent-facing parity for free.
- **Rely on the skills' own SessionStart hook injection.** Rejected: subagents run in fresh contexts and never receive that injection — the whole reason detection is orchestrator-driven.

### 8. Risks / unknowns / flags

- **No automated test coverage:** Prompt-file changes have no CI gate; correctness rests on documentation review and optional manual smoke. Absent-marker no-op behavior is the key regression guard to verify by inspection.
- **Marker path portability (A6):** Assumes `$HOME/.claude/`. If a user runs a local `.claude/` install with its own markers, the plan does not detect the local variant (out of scope) — flag if local-install support is later desired.

## Decisions
- **Decision:** Detection mechanism = filesystem check performed by the orchestrator, result passed to agents. **Reason:** Deterministic; works for subagents that do not receive SessionStart hook injection.
- **Decision:** Ponytail discipline applies only to code-writing/planning agents (crafter-implementer, crafter-planner). **Reason:** YAGNI/shortest-diff discipline is meaningful where code is authored/designed; not for review/verify/analyze roles.
- **Decision:** Caveman mode selection = lite when communicating with a human, full for reasoning and pure agent-facing output.
- **Decision (Flag A2 resolved):** Caveman level chosen by audience only — marker is presence-detection, its lite/full/ultra value is NOT a ceiling. **Reason:** Simpler; audience always wins.
- **Decision (crafter-debug parity):** Include `crafter-debug` human-facing caveman-lite wiring in this task (not deferred). Both orchestrators (`crafter-do`, `crafter-debug`) get human-facing caveman surfaces; agent-facing propagation is shared via core.md/delegation.md. **Reason:** User wants full parity now. Expands scope to also edit `skills/crafter-debug/SKILL.md`.
- **Decision (Flag A5 resolved):** Shared definitions + detection live in `rules/core.md` / `rules/delegation.md` (orchestrator-side, DRY), not inlined only in SKILL.md. **Reason:** DRY + crafter-debug parity for free.

## Outcome

Implemented in two phases on branch `feat/caveman-ponytail-detect`.

**Phase 1 (commit `53f02d8`)** — Added a shared `## Skill Detection: Caveman and Ponytail` section to `rules/core.md`: detection via marker files `$HOME/.claude/.caveman-active` + `.ponytail-active` (presence = active, content = level); caveman is audience-based (lite for human-facing prose, full for agent-facing output; marker level presence-only, `ultra` collapses to audience-based selection; compression is language-aware); five carve-outs (verbatim relay, HITL gates, Auto-Clarity, commits/PRs/release notes, persisted English); Jargon Confinement reaffirmed. Ponytail scoped to `crafter-implementer` + `crafter-planner`, level passed through. `crafter-debug` inherits the human-facing policy for free (reads core.md first, self-declares as orchestrator) — zero duplication (Step 1.3 needed no file change). Verifier 5/5 PASS; reviewer 0 Critical/Major, 3 optional (2 fixed: ultra ambiguity + language-aware compression; 1 no-change-needed: Step 6a verbatim authority confirmed in crafter-do SKILL.md).

**Phase 2 (commit `45d6512`)** — `rules/delegation.md` got one pre-spawn propagation rule (mirrors the skillbook idiom): caveman-full directive appended to every spawned agent, ponytail directive (passed-through level) only to crafter-implementer/crafter-planner; no-op when markers absent. All six `agents/*.md` gained a self-contained `## Behavior under caveman` section (fixed-format agents — reviewer/verifier/step-runner — explicitly preserve mandated output structure); crafter-implementer + crafter-planner additionally gained `## Behavior under ponytail` with safety carve-outs intact. Verifier 4/4 PASS; reviewer 0 Critical/Major, 2 optional (both fixed: analyzer Mode-B parenthetical made mode-agnostic; "drop articles" made language-aware across all six files).

**Invariant held throughout:** when neither marker is present, behavior is byte-for-byte unchanged. Purely additive Markdown; no Go/CLI/installer changes.

**Files changed:** `rules/core.md`, `rules/delegation.md`, `agents/crafter-implementer.md`, `agents/crafter-planner.md`, `agents/crafter-analyzer.md`, `agents/crafter-reviewer.md`, `agents/crafter-verifier.md`, `agents/crafter-step-runner.md`.

**Deferred (out of scope):** local `.claude/` install marker detection (A6 — global `$HOME/.claude/` only); no automated test coverage for prompt files (documentation review + optional manual smoke is the guard).
