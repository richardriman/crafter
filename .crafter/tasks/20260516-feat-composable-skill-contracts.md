# Task: Composable skill contracts for crafter-do

## Metadata
- **Date:** 2026-05-16
- **Work branch:** feat/composable-skill-contracts
- **Status:** completed
- **Scope:** Medium

## Request
Add a concrete composable skill extension model to `crafter-do`: the workflow should be able to discover and use additional useful skills with compatible contracts, especially as supplemental specialists, without weakening core Crafter guarantees.

Acceptance criteria:
- Define what a Crafter skill contract declares: capability, when it applies, required inputs, outputs, allowed/forbidden side effects, and success/failure criteria.
- Update `crafter-do` workflow instructions so it can evaluate compatible extension skills at relevant workflow phases.
- Keep v1 conservative: extension skills are supplemental only and cannot bypass or replace core planner, implementer, verifier, reviewer, approval, or commit gates.
- Document safety constraints that prevent extension skills from weakening Crafter guarantees.
- Keep changes scoped to source files in this repository, not installed runtime copies.

## Plan

**Plan status:** approved

### 1. Complete request

Give Crafter a concrete, supplemental-only extension model so `crafter-do` can recognize and consult third-party / project-local "extension skills" at the right workflow phases — without ever letting them replace the core Planner / Implementer / Verifier / Reviewer / approval / commit pipeline. The deliverable is **documentation + prompt edits in source files only**:

- A written **Skill Contract** spec describing what any Crafter-compatible skill (core or extension) must declare so it can be evaluated for compatibility: capability, applicability ("when it applies"), required inputs, outputs, allowed/forbidden side effects, success/failure criteria.
- A written **safety envelope** that lists the core guarantees extension skills cannot weaken, expressed as forbidden behaviors and required passthroughs to core agents/gates.
- Prompt edits to `skills/crafter-do/SKILL.md` (and supporting rule fragments as needed) so the orchestrator can, at a defined set of phases, **discover** available extension skills, **evaluate** their declared contracts for relevance and compatibility, and **delegate consultative work** to them strictly as supplemental specialists.

Why this matters: Crafter already has a plugin system design (`docs/plugin-system.md`) that says plugins are additive, but there is no concrete contract that defines what makes a skill "compatible" or how `crafter-do` should actually consider one. Without that, "composable skills" remains a slogan. With it, projects can plug in stack-specific specialists (e.g., a Rails review companion, a frontend a11y checker) and `crafter-do` can use them safely.

**Acceptance criteria** (mirrored from the request):
- Skill contract fields documented: capability, when-applies, inputs, outputs, allowed/forbidden side effects, success/failure criteria.
- `crafter-do` updated to evaluate compatible extension skills at relevant workflow phases.
- v1 is conservative — extension skills are supplemental only; they cannot bypass or replace Planner, Implementer, Verifier, Reviewer, approval, or commit gates.
- Safety constraints are documented in source.
- Changes affect only `skills/`, `rules/`, `docs/`, `.crafter/` source files in this repo — no edits to installed runtime copies under `~/.claude/...`.

**Validation strategy:** All changes are markdown. Validation is editorial — the changed files must be internally consistent, cross-references must resolve, and the existing core workflow (Steps 0–9b in `crafter-do`) must keep its current gates with no semantic regression. No code tests required unless a non-doc file is touched.

### 2. Assumptions / interpretations

- **Scope of "extension skill"** = any skill (in `skills/` or under a plugin's `skills/`) that is *not* one of the core crafter-* orchestrator skills (`crafter-do`, `crafter-debug`, `crafter-map-project`, `crafter-status`, `crafter-buffer`) and is *not* a core agent (`crafter-planner/implementer/verifier/reviewer/analyzer`). Existing local skills like `review/SKILL.md` are conceptual examples of "supplemental specialists" — they illustrate the shape, but we are not refactoring them as part of this task.
- **"Compatible" = declares a Skill Contract block.** Discovery is by presence of the contract section in a skill's `SKILL.md`. Skills without a contract are simply not considered as extensions by `crafter-do`. No registry, no manifest changes to `plugin.md`.
- **Discovery mechanism is prompt-only** — `crafter-do` instructs itself to scan known skill locations (project `.claude/skills/`, global `~/.claude/skills/`, and plugin-provided `skills/` per `docs/plugin-system.md`) and read each `SKILL.md` frontmatter + contract block. No new CLI subcommand, no Go code, no auto-loader. This matches Crafter's "convention over configuration" principle.
- **Relevant phases for extension evaluation (v1):** Step 1 (completeness/scope) — extensions may offer analysis input; Step 4 (Execute) — extensions may be consulted by the Implementer as specialists (e.g., "is there a `frontend-a11y` extension whose contract matches this step?"); Step 6 (Review) — extensions may produce *additional* findings appended to the core Reviewer report. They never replace the core agent's verdict.
- **Failure mode of an extension is non-fatal.** If an extension errors, is missing, or returns nothing useful, the core workflow continues unaffected.
- **The contract spec lives in `docs/skill-contract.md`** as the canonical document, with a short pointer added to `docs/plugin-system.md` and a short pointer added to `skills/crafter-do/SKILL.md`. (Alternative considered: embed in plugin-system.md — rejected because the contract concerns *all* skills, not just plugin-supplied ones.)
- **No frontmatter schema change is required** in v1 — the contract is a structured markdown section inside `SKILL.md`. (Alternative considered below.)

### 3. Non-goals

- No plugin manager, marketplace, install command, or auto-update for extensions.
- No automatic runtime scanning implemented in Go/CLI — discovery is prompt-driven only.
- No override / replacement semantics: extension skills cannot substitute for core agents or core gates, period.
- No refactor of existing skills (`review/`, etc.) to declare contracts — we only define the shape; backfilling is out of scope.
- No changes to installed runtime copies (`~/.claude/...`, `~/.copilot/...`).
- No new task-file fields, no new buffer types, no new CLI subcommands.
- No changes to `crafter-debug`, `crafter-status`, `crafter-map-project`, or agent definitions in this task.

### 4. Relevant areas

- `skills/crafter-do/SKILL.md` — the main integration site. Add an "Extension Skills" section + targeted hooks inside Steps 1, 4, 6.
- `rules/do-workflow.md` — possibly add a short subsection codifying the supplemental-only invariant alongside other invariants (green-commit, retained gates).
- `rules/core.md` — possibly extend the Karpathy-inspired guardrails reference with a note that extensions must respect them.
- `docs/plugin-system.md` — cross-reference the new contract doc; the plugin-system already says "additive only", so this fits naturally.
- `docs/skill-contract.md` — **new file**, the canonical spec.
- `skills/review/SKILL.md` — read-only reference for "shape of a specialist skill"; do not modify.
- `.crafter/ARCHITECTURE.md` — read-only; ensure planning respects skills-first conventions documented there.

### 5. Vertical phases and steps

The work is naturally two vertical slices. Each phase ends in a coherent, reviewable, committable state.

#### Phase A — Define the Skill Contract and safety envelope (doc-only)

Goal: after Phase A the repo contains a complete written specification of what makes a skill Crafter-compatible and what extension skills must never do. `crafter-do` is **not yet** wired to use it — that is Phase B. Phase A is independently valuable as a design artifact.

- [x] **A1. Author `docs/skill-contract.md`** — define the contract sections every Crafter-compatible skill should declare (capability, when-applies, required inputs, outputs, allowed side effects, forbidden side effects, success criteria, failure criteria) and give a short illustrative example block. Keep it descriptive, not prescriptive about implementation.
  - *Karpathy contract:* outcome = one new markdown doc explaining the contract; boundary = only this file is created; non-goals = no edits to other files, no schema enforcement; simplicity = single doc, single example; drift criteria = adding fields beyond the 8 listed, embedding it in another doc instead, or adding tooling; verification evidence = file exists, contains all 8 contract fields, contains one example, links nowhere broken; stop conditions = if scope creeps into changing frontmatter schema, stop and reconsider.

- [x] **A2. Add the safety envelope section to `docs/skill-contract.md`** — enumerate the guarantees extension skills cannot weaken (cannot replace core agents, cannot bypass approval/commit gates, cannot mutate task files or buffers directly, cannot rewrite plans, cannot disable drift/review/verification, cannot perform destructive git operations, must respect language and Karpathy guardrails from `rules/core.md`). Frame as a "Forbidden behaviors" + "Required passthroughs" pair.
  - *Karpathy contract:* outcome = safety envelope is explicit and exhaustive enough to ground later prompt rules; boundary = same single doc; non-goals = no detection/enforcement logic; simplicity = two flat lists; drift = adding speculative future-tense rules ("when we add X..."), turning it into a policy treatise; verification = each forbidden behavior maps to a concrete Crafter gate or invariant; stop = if a forbidden item has no corresponding existing core mechanism, surface the gap rather than invent one.

- [x] **A3. Cross-link the new doc** — add a short pointer paragraph in `docs/plugin-system.md` ("Skill compatibility — see `docs/skill-contract.md`") and add a one-line link in `.crafter/ARCHITECTURE.md` under conventions or key patterns (so future readers find it).
  - *Karpathy contract:* outcome = two surgical pointer edits, no content moved; boundary = only these two files touched in this step; non-goals = no rewriting of plugin-system.md, no new ARCHITECTURE.md section beyond a one-liner; simplicity = ≤ 3 added lines per file; drift = duplicating contract content into either file; verification = both files render and point to the new doc; stop = if either file needs structural edits to host the pointer cleanly, reduce to one pointer or note it as a follow-up.

**Phase A verification:** `docs/skill-contract.md` exists with all 8 contract fields + safety envelope; cross-links from `docs/plugin-system.md` and `.crafter/ARCHITECTURE.md` resolve; no other files modified; no behavioral change in `crafter-do` yet. Reviewer should confirm the doc is self-contained and that the safety envelope items each map to an existing Crafter gate/invariant.

#### Phase B — Wire `crafter-do` to discover and evaluate extension skills

Goal: after Phase B, `crafter-do` knows how to *consider* compatible extension skills at the three relevant workflow phases (completeness/scope, execute, review) as supplemental specialists, with the safety envelope explicitly cited. Core gates and agents are unchanged.

- [x] **B1. Add an "Extension Skills" section to `skills/crafter-do/SKILL.md`** — placed before the step-by-step procedure. Define: what counts as an extension skill, how to discover them (scan project + global + plugin `skills/` locations and read each `SKILL.md` for a Skill Contract block), v1 supplemental-only invariant, pointer to `docs/skill-contract.md` for the contract shape and safety envelope.
  - *Karpathy contract:* outcome = one new section in crafter-do; boundary = no edits to existing steps in this sub-step; non-goals = no new CLI calls, no new agents, no new buffer types; simplicity = one section, ≤ ~30 lines; drift = duplicating contract content from `docs/skill-contract.md`, inventing discovery mechanics that need code; verification = the section is internally consistent and references the doc rather than restating it; stop = if discovery looks like it needs a CLI helper, stop and raise it under risks.

- [x] **B2. Add targeted evaluation hooks in Steps 1, 4, and 6 of `crafter-do`** — short surgical inserts that say "before delegating to <core agent>, check for compatible extension skills whose contract applies; if any apply, pass their names + capabilities as supplemental context to the core agent so it can consult them as specialists. Extension skills cannot replace the core agent or its verdict." Each insert must explicitly cite the supplemental-only rule.
  - *Karpathy contract:* outcome = three small additive inserts; boundary = only Step 1, Step 4, Step 6 are touched, no semantic change to existing instructions; non-goals = no new approval gates, no fan-out to extensions at every step (only the three identified phases); simplicity = each insert is a short paragraph or bullet, no new sub-steps; drift = adding extension hooks to Verifier/Planner/commit gates, adding "extension can override" language anywhere; verification = each insert names the relevant core agent, names the supplemental-only rule, and points to `docs/skill-contract.md`; stop = if a hook would require restructuring an existing step (e.g., splitting Step 6 into 6/6c), stop and reconsider — that signals scope drift.

- [x] **B3. Codify the supplemental-only invariant in `rules/do-workflow.md`** — add a short subsection (likely near the existing "Removed gates" / "Retained gates" area or under the workflow rules) stating: extension skills are supplemental specialists; they cannot replace core agents; they cannot bypass approval, drift check, review, or commit; they must respect the safety envelope in `docs/skill-contract.md`. This makes the invariant binding across all `crafter-do` runs, parallel to the green-commit invariant.
  - *Karpathy contract:* outcome = one short authoritative subsection in the canonical rules file; boundary = only this section is added; non-goals = no restructure of existing invariants; simplicity = ≤ ~15 lines, mirrors phrasing/structure of existing invariants; drift = repeating the contract spec verbatim, adding enforcement mechanics; verification = the rule references `docs/skill-contract.md` and is consistent with the crafter-do prompt edits from B1/B2; stop = if it starts duplicating safety envelope content, trim and cross-reference instead.

- [x] **B4. Final consistency sweep** — read the four touched files (`docs/skill-contract.md`, `docs/plugin-system.md`, `skills/crafter-do/SKILL.md`, `rules/do-workflow.md`) end-to-end and confirm terminology, cross-references, and constraints are aligned. No behavioral edits — only correction of inconsistencies introduced during B1–B3.
  - *Karpathy contract:* outcome = internally consistent doc set; boundary = wording fixes only; non-goals = adding new content, restructuring; simplicity = touch only what is actually inconsistent; drift = drive-by improvements unrelated to consistency; verification = a fresh read of crafter-do + do-workflow + skill-contract tells one coherent story; stop = if a real semantic gap is found (not just wording), surface it as a discovery rather than fixing silently.

**Phase B verification:** Re-read of `crafter-do` shows extension-skill hooks present at exactly Steps 1, 4, 6, each citing supplemental-only and the safety envelope. `rules/do-workflow.md` contains the supplemental-only invariant. Existing core gates (plan approval, drift check, phase verification, review fix-loop cap, --auto retained gates, commit) read identically in semantics to before. No `~/.claude/...` or other installed paths are touched. Light editorial review (Reviewer agent) confirms internal consistency and Karpathy scorecard PASS on Simplicity / Surgical Changes / Goal-Driven.

### 6. Karpathy Contract — overall

- **Outcome:** A documented, supplemental-only skill-extension model that `crafter-do` knows how to consult at three defined workflow phases, with an explicit safety envelope that prevents weakening of core guarantees.
- **Scope boundary:** source files only (`docs/`, `skills/crafter-do/`, `rules/`, `.crafter/ARCHITECTURE.md` pointer). No code, no CLI, no agent changes, no installed-runtime edits.
- **Non-goals:** plugin manager, override semantics, automated discovery via Go, schema enforcement, frontmatter changes, backfill of existing skills.
- **Simplicity constraint:** one new doc, one new section in `crafter-do`, three small inserts inside `crafter-do`, one short subsection in `do-workflow.md`, two cross-link edits. Nothing else.
- **Drift criteria:** any of these signal drift — touching agent files, adding CLI subcommands, modifying buffers/task-file schema, introducing override/replace semantics, editing files under `~/.claude` or `~/.copilot`, adding extension hooks at gates other than Steps 1/4/6, redefining core gates.
- **Verification evidence:** the four-file set reads as one consistent story; existing `crafter-do` core gates are preserved verbatim; reviewer Karpathy scorecard PASS on Simplicity / Surgical / Goal-Driven.
- **Stop conditions:** stop and re-plan if discovery looks impossible without code; if any safety envelope item lacks a corresponding existing Crafter gate; if hooks cannot fit into Steps 1/4/6 without restructuring; if a reviewer flags ambiguity about supplemental-only.

### 7. Alternatives considered

- **Frontmatter schema extension** (add structured YAML fields `contract:` with sub-fields) — rejected for v1 because it forces a parser/validator (out of scope) and because LLM agents are equally good at reading a labeled markdown section. Can be revisited if/when a CLI validator is built.
- **Putting the contract spec inside `docs/plugin-system.md`** — rejected because the contract applies to any skill (including core ones conceptually), not just plugin-provided skills. Keeping it in its own doc gives it a stable home and avoids overloading plugin-system.md.
- **Adding extension hooks at every workflow step** — rejected. v1 is conservative; three hooks (Steps 1/4/6) cover the natural specialist insertion points (analysis, implementation context, supplemental review) without touching gating steps.
- **Auto-loading extension skills via a new CLI subcommand (`crafter skills list-contracts`)** — rejected as overengineering for v1. Discovery is prompt-driven; if usage proves the pattern is hot, a CLI helper can be added later (parallels how `crafter pr-body` was added only when LLM rendering proved unreliable).
- **Letting extensions optionally replace core agents under an opt-in flag** — explicitly rejected; this directly contradicts the supplemental-only acceptance criterion.

### 8. Risks / unknowns / flags

- **R1. Discovery realism.** Prompt-only discovery assumes the orchestrator will actually scan project/global/plugin `skills/` locations. If installed Claude Code contexts don't expose those directories to the model, the discovery will silently no-op. **Flag for the user:** confirm prompt-only discovery is acceptable for v1, knowing extensions may be invisible if the model isn't pointed at the right directories.
- **R2. Wording ambiguity on "supplemental specialist" vs. "consulted by core agent".** The Implementer (Step 4) is the natural consumer of an extension's *advice*, but the prompt phrasing must make it unambiguous that the extension does not run code or write files — the Implementer does. Mitigation: explicit phrasing in B2 + B3, cross-reference to the safety envelope.
- **R3. Backfill expectation.** Once the contract is documented, readers may expect existing skills (e.g., `review/`) to declare contracts. This task does not backfill. **Flag:** confirm the plan defers backfilling to a separate task.
- **R4. ARCHITECTURE.md pointer placement.** `.crafter/ARCHITECTURE.md` is already long; adding even a one-liner needs to land in the right section without disturbing structure. If the right section isn't obvious, the implementer should skip the ARCHITECTURE pointer (recording a Decision) and keep the doc cross-link only in `docs/plugin-system.md`.
- **R5. Possible overlap with future plugin-system v2.** If the plugin-system doc evolves to formally register plugin-provided skills, the contract spec may need to align. v1 keeps both docs independent and cross-linked to minimize coupling.

---

**Summary:** This contract protects a small, additive, doc-and-prompt-only change that introduces a concrete Skill Contract + supplemental-only safety envelope and wires `crafter-do` to consult compatible extension skills at three specific phases. It is the right approach because it makes the "composable skills" concept concrete without introducing code, registries, or override semantics, and because every change traces directly to an explicit acceptance criterion.

## Decisions
- **Decision:** Use `feat/` as the branch prefix for this task. **Reason:** User prefers the shorter prefix over `feature/`.
- **Decision (Orchestrator Accepted):** Accepted local A3 formatting drift: `docs/plugin-system.md` received 4 added lines instead of the planned ≤3 because the extra blank line is required for idiomatic Markdown heading structure. **Reason:** The change is local, preserves the short pointer-only intent, duplicates no contract content, and does not affect later steps.

## Outcome
- Implemented the composable skill contract model in two committed phases:
  - `ac1f40b` (`docs: add skill contract specification`) added `docs/skill-contract.md`, documented the eight contract fields and Safety Envelope, and cross-linked `docs/plugin-system.md` plus `.crafter/ARCHITECTURE.md`.
  - `5fe353f` (`docs: wire extension skills into crafter-do`) added the `crafter-do` Extension Skills section, Step 1/4/6 supplemental-only hooks, and the binding supplemental-only invariant in `rules/do-workflow.md`.
- Verified and reviewed both phases. Phase B review closed clean with no findings.
- Accepted deviation: `docs/plugin-system.md` used a 4-line Markdown section instead of the planned ≤3-line pointer so the cross-link fits the document's heading-per-topic style.
