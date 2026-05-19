# Task: Crafter-jargon context bleed — terminology cleanup + prompt-design guardrail

## Metadata
- **Date:** 2026-05-19
- **Work branch:** chore/crafter-jargon-context-bleed
- **Status:** completed
- **Scope:** Medium

## Request

Implement the "Crafter-jargon context bleed" backlog item recorded in `.crafter/STATE.md` `## Planned` (also captured in PR #34). Two parts:

(a) **Confusing term cleanup.** Replace the confusing metaphor "seam" with a clear canonical English term. **Decision: the canonical term is `split point`** (user-chosen at planning time; candidates were split point / cut point / module boundary). Add a one-line glossary definition to `docs/core-capabilities.md`. **Sweep of existing task-file prose is explicitly out of scope** (user chose "Skip sweep" — task files are historical artifacts; the glossary + guardrail govern future occurrences).

(b) **Prompt-design guardrail (the deeper fix).** Add an explicit instruction (likely in `rules/core.md` Language/Communication rules and/or agent prompts) that agents MUST describe the user's own project domain in domain-neutral / the user's own terms, and reserve crafter-internal vocabulary (`gate`, `drift`, `seam`/`split point`, `surface`, `binding`, `escape hatch`) strictly for crafter workflow mechanics — never project it onto the user's codebase. `gate`/`gated` stays a legitimate term *for crafter's own gates*; the guardrail is about NOT exporting crafter jargon into unrelated domain explanations.

## Plan

**Plan status:** approved

### Complete request

Surgical, documentation/instruction-layer changes to Crafter itself:

- **(a) Terminology cleanup.** "seam" is a confusing metaphor. The canonical English replacement is **`split point`** (locked decision). Add a single one-line glossary definition of `split point` to `docs/core-capabilities.md`. No code or task-file prose sweep (locked out of scope).
- **(b) Prompt-design guardrail.** Add a durable, clearly-named guardrail subsection in `rules/core.md` (single source of truth for the content) plus a one-line POINTER to it in each prose-generating agent prompt. Agents/orchestrator must describe the user's own project domain in domain-neutral / the user's own terms, and reserve crafter-internal vocabulary (`gate`, `drift`, `seam`/`split point`, `surface`, `binding`, `escape hatch`) strictly for crafter workflow mechanics — never projecting it onto the user's codebase. The carve-out is generalized: ALL listed terms stay fully legitimate when describing crafter's OWN workflow mechanics; the prohibition is solely about exporting them onto the user's unrelated domain.

**Why this matters:** The crafter workflow is permanently loaded in the context window. Its metaphor-laden vocabulary lexically primes the model and bleeds into prose about the user's unrelated domain (real failure: a Gantt panel described as "gatovat panel … na všech švech … bez flag gate"). This makes developer-facing prose jargon-heavy and harder to understand. A core.md guardrail provides the single-source content; thin pointers in the prose-generating agents reinforce adherence at the point of generation.

**Acceptance criteria:**
1. `docs/core-capabilities.md` contains exactly one concise glossary definition of `split point`.
2. `rules/core.md` contains an explicit, durable, clearly-named guardrail subsection with (i) the positive domain-neutral / user's-own-terms instruction, (ii) the explicit reserved-jargon list, (iii) a generalized carve-out stating ALL listed terms remain legitimate for crafter's own mechanics (explicitly naming `drift` alongside `gate`/`gated` as the worked example), and (iv) ONE concrete contrasting bad/good example using the STATE.md Gantt case.
3. Each of `agents/crafter-planner.md`, `agents/crafter-implementer.md`, `agents/crafter-reviewer.md` contains exactly ONE added line that POINTS to the core.md guardrail — no duplication of the jargon list or guardrail content into the agents.
4. No other rule sections, agent prompts, or task files are rewritten beyond what (1)–(3) require.
5. No new files; no installer/loader wiring changes (all touched files are already deployed).

**Validation strategy:** By inspection only — this is prose, no Go code and no behavioral logic, so there are NO automated tests for the content. Verification = grep + reading the added text. The existing `tests/test_install.sh` should still pass unchanged (no install surface touched) and is the only automated check, used solely as a no-regression guard.

### Assumptions / interpretations

- **Single-vs-multi placement — RESOLVED (user-directed): MULTI via pointer-only reinforcement, content single-sourced in core.md.** The guardrail CONTENT lives in exactly one place — a clearly-named subsection of `rules/core.md` (deployed at `install.sh` line 313, loaded by all workflows). The prose-generating agents get a ONE-LINE pointer to that subsection — never a copy of the jargon list or guardrail text. This keeps a single source of truth (no content drift) while reinforcing adherence at the point of user-facing prose generation. This is now a settled decision, not an open flag.
- **Which agents are "prose-generating" — verified.** The three agents that produce user-facing domain prose are `crafter-planner.md` (writes plans/explanations), `crafter-implementer.md` (writes change descriptions/summaries), and `crafter-reviewer.md` (writes review prose). `crafter-verifier.md` and `crafter-analyzer.md` produce structured/diagnostic output rather than narrative user-domain prose, so they are excluded to keep the change minimal; the Implementer may include one of them ONLY if it clearly emits user-facing domain narrative. All three target agents are deployed by `install.sh` (lines 344, 345, 347) and share `## Constraints` and `## Output format` sections — natural one-line homes. No new install wiring is required because all touched files are existing deployed files.
- **Glossary placement.** `docs/core-capabilities.md` has no existing glossary and its sections are taxonomy/loading/runtime-path specific. A small, clearly-scoped "Glossary" (or "Terminology") subsection appended near the end is cleaner than wedging the definition into an unrelated section. The Implementer may instead attach the single line to an existing section if a natural home exists, as long as it is one concise definition and easy to find. Keep it minimal either way.
- **`docs/` is NOT deployed by the installer** (verified: `install.sh` has no `docs/` copy lines). So the glossary in `core-capabilities.md` is a maintainer-facing reference, not a runtime-loaded instruction. The behavioral fix that actually stops the bleed is the `rules/core.md` guardrail (which IS deployed and loaded). The glossary is documentation hygiene that anchors the canonical term; it is correct and intended that it is not shipped to user installs.
- **Seam footprint verified:** `grep -rni "seam"` across `rules/ agents/ docs/ skills/ templates/` returns zero non-"seamless" hits, and `split point` does not yet exist anywhere. STATE.md's claim is accurate; no operative-file sweep is needed or in scope.
- The guardrail must name the offending vocabulary explicitly (the STATE.md list) so the instruction is concrete, while clarifying via a GENERALIZED carve-out that ALL listed terms remain valid when describing crafter's own workflow mechanics — not a `gate`-only carve-out. `drift` is named alongside `gate`/`gated` as the worked example of a heavily-used internal mechanic term that must keep working (step drift check, scope drift, beneficial local drift).

### Non-goals

- NOT sweeping or rewriting existing task-file prose (locked out of scope).
- NOT rewriting unrelated sections of `rules/core.md`, `docs/core-capabilities.md`, the agent prompts, or any workflow rule.
- NOT duplicating the guardrail content or the jargon list into the agent prompts — agents get a ONE-LINE pointer only; core.md remains the single source of truth.
- NOT touching Go code, `install.sh`, the loader, or `tests/`.
- NOT adding any new file, glossary document, or rule fragment.
- NOT re-litigating the canonical term (`split point` is locked) or the single-vs-multi placement (resolved: MULTI via pointer-only reinforcement).
- NOT renaming or redefining any reserved term (`gate`/`gated`/`drift`/`seam`/`split point`/`surface`/`binding`/`escape hatch`) for crafter's own usage — only constraining their export onto the user's domain.

### Relevant areas

- `rules/core.md` — has "Language Rules", "General Principles", "Karpathy-Inspired Guardrails". The new clearly-named guardrail subsection belongs here, within/adjacent to "Language Rules".
- `docs/core-capabilities.md` — design note; append/attach the one-line `split point` glossary definition. End of file (after `### Applied in this task`) or a new short Glossary subsection.
- `agents/crafter-planner.md`, `agents/crafter-implementer.md`, `agents/crafter-reviewer.md` — prose-generating agents; each gets ONE pointer line in a natural communication/output section (each has `## Constraints` and `## Output format`). No content duplication.
- `.crafter/STATE.md` `## Planned` — authoritative problem statement and the canonical jargon list to reference; the bullet is checked off via normal post-change steps (not part of plan steps).
- `install.sh` (read-only confirmation): `rules/core.md` deployed at line 313; the three agents deployed at lines 344/345/347; `docs/` not deployed — no wiring required, confirms no install changes.

### Phase 1 — Canonical term defined + bleed guardrail installed

Single coherent phase: all changes are tiny, mutually reinforcing prose edits that together leave the repo in a complete, reviewable state (term anchored + durable single-source guardrail live + agents pointing to it).

- [x] **Step 1 — Define `split point` in the design note.** Add exactly one concise glossary definition of `split point` to `docs/core-capabilities.md`, in a small clearly-labelled Glossary/Terminology subsection or attached to a natural existing section. The definition states plainly what a split point is (the point where a module/file is divided during decomposition/refactoring — the concept formerly called "seam") in domain-neutral English that translates well. One line of definition; no broader doc rewrite.
- [x] **Step 2 — Install the jargon-bleed guardrail subsection in `rules/core.md`.** Add a clearly-named subsection (not a buried bullet), within or adjacent to "Language Rules", containing: (i) the positive instruction that agents and the orchestrator MUST describe the user's own project/domain in domain-neutral language or the user's own terms; (ii) the explicit reserved-jargon list `gate`, `drift`, `seam`/`split point`, `surface`, `binding`, `escape hatch`; (iii) a GENERALIZED carve-out stating ALL listed terms remain fully legitimate when describing crafter's OWN workflow mechanics, with `drift` named alongside `gate`/`gated` as the worked example (step drift check, scope drift, beneficial local drift), the prohibition being solely about exporting them onto the user's unrelated domain; (iv) ONE concrete contrasting example — bad: the STATE.md Gantt case ("gatovat panel … na všech švech … bez flag gate"), good: the same described in plain domain terms. Smallest addition that is unambiguous and durable; do not restructure existing core.md sections.
- [x] **Step 3 — Add a one-line pointer in each prose-generating agent prompt.** In each of `agents/crafter-planner.md`, `agents/crafter-implementer.md`, `agents/crafter-reviewer.md`, add exactly ONE line in a natural communication/output section (e.g. `## Constraints` or `## Output format`) instructing the agent to follow the `rules/core.md` jargon-confinement guardrail and not export crafter vocabulary onto the user's domain. The line POINTS to core.md — it must NOT restate the jargon list or guardrail content (single source of truth stays in core.md). One line per agent; no other edits to these files.

#### Phase 1 verification (by inspection — no automated content tests exist)

- [x] `grep -ni "split point" docs/core-capabilities.md` returns exactly one definition occurrence, and reading it confirms a single concise, translatable definition.
- [x] `grep -ni "jargon\|user's own\|domain-neutral\|drift" rules/core.md` shows the new clearly-named guardrail subsection; reading it confirms it contains (i) the positive instruction, (ii) the full reserved-jargon list, (iii) the GENERALIZED carve-out that explicitly names `drift` alongside `gate`/`gated` and states ALL listed terms stay legitimate for crafter's own mechanics, and (iv) the one bad/good Gantt example.
- [x] `grep -ni "core.md\|jargon\|crafter vocabulary" agents/crafter-planner.md agents/crafter-implementer.md agents/crafter-reviewer.md` shows exactly one pointer line per file; reading confirms each POINTS to the core.md guardrail and does NOT duplicate the jargon list / guardrail content.
- [x] `git diff --stat` shows only `docs/core-capabilities.md`, `rules/core.md`, and the three named agent files changed (plus the task file / STATE.md via normal post-change steps) — no other files, no new files, no `install.sh`/`tests/`/Go changes.
- [x] `bash tests/test_install.sh` still passes (no install surface touched) — no-regression guard only.
- [x] Re-grep confirms still zero "seam" (non-"seamless") occurrences introduced into operative files.

#### Phase 1 review gate

- [x] Reviewer confirms: (1) glossary is one minimal line and the term matches the locked `split point` decision; (2) the core.md guardrail is a clearly-named subsection that is durable, explicit, lists the full jargon, has the GENERALIZED carve-out naming `drift` alongside `gate`/`gated`, and includes the one bad/good Gantt example; (3) each of the three named agents has exactly one pointer line with NO content/jargon-list duplication (single source of truth preserved); (4) no scope creep into unrelated sections, no new files, no install/Go/test changes.

### Karpathy Contract

**Phase 1**
- **Outcome:** The canonical term `split point` is defined once in the design note; a durable, clearly-named single-source guardrail in `rules/core.md` stops crafter jargon from bleeding into user-domain prose; and the three prose-generating agents point to it.
- **Scope boundary:** Only `docs/core-capabilities.md` (one definition), `rules/core.md` (one guardrail subsection), and `agents/crafter-planner.md` / `crafter-implementer.md` / `crafter-reviewer.md` (one pointer line each). Normal post-change updates to the task file / STATE.md bullet are allowed via standard workflow steps.
- **Non-goals:** No task-file sweep, no rule restructuring, no content/jargon-list duplication into agents, no Go/install/test changes, no new files, no edits to verifier/analyzer agents.
- **Simplicity constraint:** Smallest possible prose additions; one definition, one named guardrail subsection, one pointer line per agent; no speculative wording.
- **Drift criteria:** Drift if any file outside the five targets is edited for content; if guardrail CONTENT or the jargon list is duplicated into any agent (pointer-only is required); if a pointer becomes multi-line or restates the list; if `split point` is defined more than once or the term deviates from the locked decision; if existing sections are reworded; if the carve-out is `gate`-only rather than generalized across all listed terms; if any new file or install wiring appears. **Note:** adding the one-line pointer to the three prose-generating agents is NOW IN SCOPE and is NOT drift.
- **Verification evidence:** Grep outputs + read-through of all additions; `git diff --stat` scope (five files); `tests/test_install.sh` green as a no-regression guard. Explicitly inspection-based — no automated tests exist for this prose.
- **Stop conditions:** Stop and escalate if defining the term cleanly requires touching unrelated taxonomy content, if the guardrail wording would require rewording existing core.md rules to remain coherent, or if any target agent has no natural single-line communication/output home without restructuring it.

**Step 1**
- **Outcome:** One concise, translatable `split point` definition exists in `docs/core-capabilities.md`.
- **Scope boundary:** `docs/core-capabilities.md` only; one definition (plus at most a short subsection heading).
- **Non-goals:** No rewrite of taxonomy/loading/runtime-path sections; no multi-line essay; no examples sprawl.
- **Simplicity constraint:** Single definition line; minimal heading if any.
- **Drift criteria:** More than one definition, edits to unrelated sections, or term deviating from `split point`.
- **Verification evidence:** `grep -ni "split point" docs/core-capabilities.md` = one definition; read confirms clarity.
- **Stop conditions:** Stop if no clean placement exists without altering unrelated content — escalate.

**Step 2**
- **Outcome:** A durable, explicit, clearly-named jargon-confinement guardrail subsection is live in `rules/core.md` as the single source of truth.
- **Scope boundary:** `rules/core.md` only; a named subsection within/adjacent to "Language Rules".
- **Non-goals:** No restructuring of existing sections; no agent edits; no redefinition of any reserved term for crafter's own usage.
- **Simplicity constraint:** Smallest unambiguous subsection containing the positive instruction, the full jargon list, the generalized carve-out, and exactly one bad/good example.
- **Drift criteria:** Edits to other core.md sections; omission of any of the four required parts; a `gate`-only (non-generalized) carve-out; failure to name `drift` alongside `gate`/`gated` in the carve-out; more than one example.
- **Verification evidence:** Grep shows the named subsection; read confirms (i) positive instruction, (ii) full reserved list, (iii) generalized carve-out naming `drift` + `gate`/`gated`, (iv) one Gantt bad/good example.
- **Stop conditions:** Stop if the guardrail cannot be added without reworking existing rules — escalate.

**Step 3**
- **Outcome:** Each of the three prose-generating agents has exactly one line pointing to the core.md guardrail.
- **Scope boundary:** `agents/crafter-planner.md`, `agents/crafter-implementer.md`, `agents/crafter-reviewer.md` only; one line each in a natural communication/output section.
- **Non-goals:** No duplication of the jargon list / guardrail content; no other edits to these agents; no edits to verifier/analyzer.
- **Simplicity constraint:** One line per agent that POINTS to `rules/core.md` and forbids exporting crafter vocabulary onto the user's domain.
- **Drift criteria:** Any pointer that restates the jargon list or guardrail content; multi-line additions; edits beyond the one line; touching non-target agents; any install-wiring change (none needed — files already deployed).
- **Verification evidence:** Grep shows exactly one pointer line per file; read confirms it references core.md and contains no duplicated content.
- **Stop conditions:** Stop if an agent has no natural single-line home without restructuring it — escalate.

### Alternatives considered

- **Single-source-only in core.md, no agent pointers.** Considered, but the user directed MULTI: a core.md guardrail PLUS thin agent pointers. The chosen design keeps content single-sourced (no drift) while reinforcing adherence exactly where user-facing prose is generated — the leanest way to satisfy the directive.
- **Duplicating the guardrail content/jargon list into each agent prompt.** Rejected: `rules/core.md` is the deployed single source of truth; copying content invites divergence and violates Simplicity First. Pointer-only reinforcement gets the proximity benefit without the drift cost.
- **A standalone glossary doc / new rule fragment.** Rejected: adds a file that would need loader/installer wiring and contradicts the "smallest set of additions" guidance; one line in the existing design note plus the core.md guardrail is sufficient.
- **A `gate`-only carve-out.** Rejected per user refinement: every listed reserved term (not just `gate`) is legitimate for crafter's own mechanics, so the carve-out must be generalized and explicitly name `drift` as the worked example of a heavily-used internal term.
- **Sweeping task-file prose to replace "seam".** Rejected — explicitly locked out of scope; task files are historical artifacts.
- **Putting the glossary line in `rules/core.md` so it ships.** Rejected: the design note is the right home for decomposition vocabulary; the behavioral fix (guardrail) already ships via core.md, and `docs/` deliberately stays repo-internal.

### Risks / unknowns / flags

- **Single-vs-multi placement — RESOLVED, no longer an open flag.** User-directed decision: MULTI via pointer-only reinforcement, content single-sourced in `rules/core.md`. The three prose-generating agents (planner/implementer/reviewer) get a one-line pointer; no content duplication. No open question remains here.
- **Low risk overall:** prose-only, five existing deployed/repo-internal files, no install or Go surface (all touched agents verified deployed at `install.sh` lines 344/345/347; `rules/core.md` at 313; `docs/` repo-internal). Main residual risk is wording that is too vague to change model behavior — mitigated by requiring the explicit jargon list, the generalized carve-out naming `drift`, and the concrete bad/good example in the verification evidence.

**Summary:** This contract protects a minimal, surgical outcome — anchor the canonical term `split point` once, install a durable clearly-named jargon-confinement guardrail (with a generalized carve-out and one worked bad/good example) as the single source of truth in `rules/core.md`, and add one pointer line to each of the three prose-generating agents — the leanest change that both clarifies terminology and structurally stops crafter jargon from bleeding into user-domain prose without introducing content drift.

## Decisions
- **Decision (User Accepted):** Canonical English replacement for "seam" is `split point`. **Reason:** Idiomatic in refactoring, translates well, primary-English requirement satisfied.
- **Decision (User Accepted):** Task-file prose sweep is out of scope. **Reason:** Task files are historical artifacts; rewriting old prose adds noise without value — glossary + guardrail address future occurrences.
- **Decision (Orchestrator Accepted):** Step 1 also touched `.crafter/skillbook.json` (automated `appliedCount`/`updatedAt` increments from `crafter skillbook get`). **Reason:** Inherent, zero-content workflow side effect of skillbook delegation per `rules/delegation.md`; not implementer content drift, no impact on later steps.

## Outcome

**Commit:** `99662b1`

**What was done (Steps 1–3 + Suggestion fixes):**

- **Step 1** — Added a `## Glossary` subsection to `docs/core-capabilities.md` with one concise definition of `split point` (the point where a module or file is divided during decomposition or refactoring — the canonical replacement for the informal term "seam").
- **Step 2** — Added a `## Jargon Confinement` guardrail subsection to `rules/core.md` containing: (i) the positive instruction that agents must describe the user's domain in domain-neutral / the user's own terms; (ii) the explicit reserved-jargon list (`gate`, `drift`, `seam`/`split point`, `surface`, `binding`, `escape hatch`); (iii) a generalized carve-out explicitly naming `drift` alongside `gate`/`gated` — all listed terms remain fully legitimate when describing crafter's own workflow mechanics; (iv) one contrasting bad/good example using the Gantt-panel case from STATE.md, with an English gloss for the Czech bad example.
- **Step 3** — Added exactly one pointer line to each of `agents/crafter-planner.md`, `agents/crafter-implementer.md`, and `agents/crafter-reviewer.md` directing the agent to follow the `rules/core.md` Jargon Confinement guardrail. No guardrail content or jargon list was duplicated into the agents; single source of truth preserved.
- **Suggestion fix 1** — Reviewer suggested sharpening the bad-example gloss; Czech phrase was reconstructed from context and an English translation added.
- **Suggestion fix 2** — Reviewer suggested rewording the pointer line in the reviewer agent for parallelism with the other two; wording aligned.

**Accepted deviations:**
- `split point` term was locked at planning time (user-chosen); no re-evaluation during implementation.
- Task-file prose sweep was explicitly locked out of scope; historical task-file occurrences of "seam" left unchanged.
- Czech reconstructed Gantt example accepted as a good-faith reconstruction (original source was STATE.md planning prose); English gloss added to satisfy verifier re-check.
- `skillbook.json` `appliedCount`/`updatedAt` auto-increments from `crafter skillbook get` at Step 1 were an inherent workflow side effect, not implementer content drift (documented in Decisions).

**Follow-up:** The "Karpathy Contract" term softening / communication improvement was recorded as a new `- [ ]` backlog item in `STATE.md ## Planned`.
