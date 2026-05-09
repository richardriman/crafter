# Task: UAT and Gaps buffers + crafter-buffer skill (GH#16)

## Metadata
- **Date:** 2026-05-09
- **Work branch:** feat/gh-16-buffer-skill
- **Status:** active
- **Scope:** Medium

## Request

Implement GitHub issue #16 — `[CHANGE] UAT and Gaps buffers + crafter-buffer skill`.

The full contract is in the issue body at https://github.com/richardriman/crafter/issues/16. Summary of what this task delivers:

**Context.** GH#15 introduced `--auto` mode (already shipped — see commit 308f828 + 828815e). Under `--auto`, sub-agents must not block on findings the auto-fix loop cannot resolve; they must instead record those findings to a persistent location that surfaces to humans after the run. This task adds that buffer mechanism.

**Issue body excerpt (file format superseded — see Decision 1):**

**Buffers** under `.crafter/run/<task-id>/`:
- `uat-buffer.md` — manual QA items (UI flows, external integrations, business decisions humans must validate)
- `gaps-buffer.md` — known gaps / tech debt / out-of-scope follow-ups intentionally deferred by the agent
- `decisions.md` — already exists; keep separate

**Skill `crafter-buffer`** with two append operations:
- `crafter-buffer uat --title "..." --source "..." --verify "..." --why-manual "..."`
- `crafter-buffer gap --title "..." --source "..." --detail "..." --followup "..."`

Both append a structured Markdown block to the corresponding buffer file (creating it if missing).

**Lifecycle.** `.crafter/run/<task-id>/` is created at run start, persists during the run, and is cleaned up after PR composition (or on workspace teardown). The directory must be in `.gitignore` so buffers never leak into commits.

**Resolved question** _(see Decision 1)_**:** unified schema (`kind: uat | gap`) vs. distinct schemas — the issue defers to implementation; the task documents the chosen decision in `## Decisions`.

**Acceptance criteria** (from issue):
- [ ] `skills/crafter-buffer/SKILL.md` exists and documents the two append operations
- [ ] Buffer file format documented with at least 2 example entries each (UAT, Gap)
- [ ] `crafter-buffer uat ...` creates the file if missing and appends a well-formed entry
- [ ] `crafter-buffer gap ...` does the same for the gaps buffer
- [ ] `.crafter/run/<task-id>/` directory creation + lifecycle is documented in workflow rules
- [ ] `.gitignore` (or template equivalent) excludes `.crafter/run/`
- [x] Schema decision (unified vs distinct) is documented _(see Decision 1)_

**Out of scope:**
- PR composer extension that reads buffers (separate issue, GH#17)
- Agent prompt updates that switch from blocking to buffer-append (GH#18)
- Persistence across runs / centralized cross-task store / real-time UI

**Forward references in GH#15 contract:**
- `rules/do-workflow.md` already mentions UAT buffer and Gaps buffer as forward references to GH#16
- `skills/crafter-do/SKILL.md` Step 6b also references the buffer skill as `*(buffer skill defined in companion task GH#16)*`

After this task, those forward references resolve to real artifacts.

## Plan

**Plan status:** approved

### Approach

This task delivers the runtime artifacts the GH#15 contract forward-references: a per-run `.crafter/run/<task-id>/` workspace, two append-only Markdown buffers inside it (`uat-buffer.md`, `gaps-buffer.md`), a new `crafter-buffer` skill that documents the two append operations, lifecycle wording in `rules/do-workflow.md`, and a `.gitignore` rule that prevents the run directory from leaking into commits. The architectural question is "Markdown-only skill that calls a shell command, a new Go subcommand, or a hybrid?" — this task answers that question explicitly in `## Decisions` before the skill is written, because the answer drives every other step.

The plan is split into four small vertical phases. Phase 1 records the schema decision (unified vs. distinct entry shape) and the implementation-surface decision (skill-only / Go subcommand / hybrid) before any artifact is built — those two decisions constrain every later phase. Phase 2 lands the `crafter-buffer` skill and (if the decision in Phase 1 calls for it) the supporting Go subcommand, with example entries embedded in the skill prose. Phase 3 lands the workflow integration: lifecycle wording for `.crafter/run/<task-id>/` in `rules/do-workflow.md`, the `.gitignore` rule, and resolution of the GH#15 forward references so they point at real artifacts. Phase 4 runs the verification round-trips required by the issue (append-then-read, code-fence-safe parsing, concurrency note) and end-of-task housekeeping. Each phase is reviewable in isolation.

### Why this approach

- **Decisions before construction.** The two open architectural questions (schema shape, Markdown-only vs. Go subcommand) are recorded in `## Decisions` first. Encoding the decision before writing the skill avoids re-doing prose if the schema flips, and gives the reviewer one paragraph to trace every later choice back to.
- **Skill is the authoritative interface, not the implementation.** Even if a Go subcommand backs the appends, the skill remains the documented contract — agents read SKILL.md, not Go source. This matches the existing pattern: `skills/crafter-do/SKILL.md` is the contract, `cli/cmd/skillbook_*.go` is the deterministic backing.
- **Lifecycle wording lives where the workflow lives.** `.crafter/run/<task-id>/` creation/cleanup is a workflow concern (run start, after PR composition), not a skill concern. Putting it in `rules/do-workflow.md` keeps the workflow self-contained and makes the GH#15 forward references resolve to one canonical paragraph instead of being scattered.
- **`.gitignore` rule is single-source.** The repo currently has only a tiny project-local `.gitignore` (`tmp/`); there is no `templates/.gitignore`. The plan adds the `.crafter/run/` ignore at the project root and documents in `rules/do-workflow.md` that downstream projects should add the same line. A full gitignore-template extension is out of scope.
- **Forward-reference resolution is explicit and minimal.** Eight existing call sites in `rules/do-workflow.md` and `skills/crafter-do/SKILL.md` use phrases like "(forward reference to GH#16)" or "(buffer skill defined in companion task GH#16)". Phase 3 replaces those annotations with concrete pointers to the new skill / file paths in a tightly scoped edit, without rewriting the surrounding contract.

**Plan amendment (fix-loop iteration #1, 2026-05-09):** The buffer file format was changed from Markdown blocks (`.md`) to NDJSON (`.jsonl`) during review fix-loop iteration #1. This is a deliberate deviation from the issue body wording; the deviation is captured and approved in Decision 1 (schema) in `## Decisions`.

### Phase 1 — Record schema and implementation-surface decisions

Land the two architectural decisions in the task file's `## Decisions` section before writing any artifact. After this phase, every subsequent step has an unambiguous contract to follow.

- [x] **Step 1 — Schema decision (unified vs. distinct).** Decide whether UAT and Gap entries share one schema with a `kind: uat | gap` discriminator, or stay as two distinct schemas with their own field sets (UAT: `title`, `source`, `verify`, `why-manual`; Gap: `title`, `source`, `detail`, `followup`). Record the chosen option, the alternatives considered, and the rationale as a `Decision (User Accepted)` entry in `## Decisions`. The deciding factors to weigh: (a) issue spec lists distinct field names per kind, suggesting distinct schemas are the natural choice; (b) two distinct files (`uat-buffer.md`, `gaps-buffer.md`) already separate the kinds at the file level, so a shared `kind:` field inside entries is redundant; (c) future PR composer (GH#17) will read the two files separately. Default recommendation if no further input: **distinct schemas** — but explicitly call this out for user confirmation.
- [x] **Step 2 — Implementation-surface decision (skill-only / Go subcommand / hybrid).** Decide whether `crafter-buffer uat …` and `crafter-buffer gap …` are (a) shell snippets the agent assembles inline (skill prose only, agent runs `mkdir -p`/`cat >>` itself), (b) a new Go subcommand `crafter buffer uat|gap …` mirroring the `crafter skillbook add` pattern with atomic writes, or (c) a hybrid where the skill documents the contract but delegates to the Go subcommand for the deterministic append. Weighing factors: (i) PROJECT.md states "LLMs do JSON CRUD, Jaccard similarity, and atomic writes poorly — a static binary with zero runtime deps handles these reliably" — append-with-atomic-write fits that policy; (ii) entries may contain code fences and multi-line strings, which agents quote inconsistently in shell; (iii) concurrency (sub-agents calling buffer in the same run) benefits from a single atomic-write code path; (iv) however, a Go subcommand expands distribution surface (cross-compilation, version pinning) and may be overkill if the append is line-oriented and idempotent. Record the chosen option and rationale as a `Decision (User Accepted)` entry. The plan's later phases adapt to the chosen surface (Phase 2 either ships only a skill, or ships a skill plus `cli/cmd/buffer*.go`).
- [x] **Step 3 — Concurrency policy.** Record a one-paragraph `Decision` on concurrency: whether concurrent calls (multiple sub-agents in the same run appending to the same buffer) are explicitly serialized, advisory-locked, or simply documented as "callers run sequentially under the orchestrator." This decision must be recorded before Phase 2 because it changes whether the skill needs to document a lock idiom or just a sequential expectation.

**Drift criteria:** Drift if any architectural decision is made implicitly (e.g., implementer picks a schema by writing it without a recorded `## Decisions` entry); drift if the schema decision is recorded as anything other than one of the two enumerated options; drift if any decision silently substitutes a different buffer-file format (Markdown, JSON array, YAML, etc.) for the agreed NDJSON / `.jsonl` line-append format recorded in Decision 1.

**Verification criteria:** `## Decisions` contains three new entries — schema, implementation surface, concurrency — each with a clear chosen option and rationale. The user explicitly approves them (or the orchestrator records `Decision (Orchestrator Accepted)` with rationale under `--auto`/Auto Mode, per the project's recent precedent).

### Phase 2 — Ship `crafter-buffer` skill (and Go subcommand if Phase 1 chose that)

Create `skills/crafter-buffer/SKILL.md` per the conventions used by `skills/crafter-do/SKILL.md` and `skills/crafter-status/SKILL.md` — YAML frontmatter with `name` + `description`, prose-first body. Document the two append operations with at least two example entries each (UAT, Gap) showing the NDJSON entry (single-line JSON object terminated by `\n`). If Phase 1 Step 2 chose the Go-backed surface, also add `cli/cmd/buffer.go` + `cli/cmd/buffer_uat.go` + `cli/cmd/buffer_gap.go` (or equivalent layout, mirroring `skillbook` cobra structure) with atomic-write helpers. The skill always remains the authoritative interface; the Go subcommand is purely deterministic plumbing if chosen.

- [ ] **Step 1 — Create `skills/crafter-buffer/SKILL.md` with frontmatter and prose contract.** Frontmatter: `name: "crafter-buffer"`, `description: "Append a UAT or Gap entry to the current run's buffer (.crafter/run/<task-id>/)"`. Body sections, in this order: (a) what the skill does and when to call it; (b) the two operations with their flag signatures (`crafter buffer uat --title "..." --source "..." --verify "..." --why-manual "..."` and `crafter buffer gap --title "..." --source "..." --detail "..." --followup "..."`) — flag list comes from the issue and Phase 1 Step 1 schema decision; (c) the NDJSON entry shape for one entry per kind, including (i) a minimal valid entry, (ii) an entry whose `verify`/`detail` field contains a fenced code block expressed as a JSON-escaped string (`\n` newlines, backticks unescaped), and (iii) an entry with a multi-line value embedded via JSON string escaping. Show how callers should pass the value via the `crafter buffer` flag; (d) at least two example UAT entries and two example Gap entries showing realistic content including at least one entry with multi-line content and one with a fenced code block; (e) creation behavior — file is created if missing, headed by a fixed first-line marker so future tooling can detect the file kind; (f) concurrency note matching Phase 1 Step 3's decision.
- [ ] **Step 2 — (Conditional on Phase 1 Step 2 picking Go-backed or hybrid) Add `crafter buffer` subcommand to the CLI.** Mirror the existing `crafter skillbook` cobra structure: `cli/cmd/buffer.go` (parent command), `cli/cmd/buffer_uat.go` (uat subcommand with required flags from Phase 1 Step 1), `cli/cmd/buffer_gap.go` (gap subcommand). Implementation lives in a new `cli/internal/buffer/` package mirroring `cli/internal/skillbook/` (types, store with atomic append, format). The atomic-append must handle: file-not-exists (create with a header), entry serialization that preserves code fences and multi-line content (the format chosen in Step 1), and concurrent-append safety per Phase 1 Step 3. If Phase 1 chose skill-only (no Go), this step is skipped and explicitly marked `~~struck through~~` with a note pointing at the recorded decision.
- [ ] **Step 3 — Wire installer to deploy the new skill (and binary if applicable).** Confirm `install.sh` already deploys all `skills/crafter-*/SKILL.md` (it does, per ARCHITECTURE.md "Distribution"). If Phase 1 chose Go-backed, confirm the new subcommand is reachable through the existing build pipeline (`cli/Makefile` cross-compile targets pick up new cobra commands automatically). No new installer-level changes are expected; if any are required, this step is the place to surface them and re-confirm scope with the user.

**Drift criteria:** Drift if (a) the example entries do not include at least one with code fences and one multi-line, (b) drift if the entry shape introduces a delimiter or framing concept inside the JSON object beyond the keys defined in Decision 1 — every entry MUST be a self-contained JSON object on a single line, with no out-of-band delimiters, (c) the Go subcommand introduces a runtime dependency beyond what the existing CLI already uses, (d) any new file is placed outside `skills/crafter-buffer/` or `cli/cmd/`+`cli/internal/buffer/`.

**Verification criteria:** Phase 2 verification: `skills/crafter-buffer/SKILL.md` exists with frontmatter; documents both `crafter buffer uat` and `crafter buffer gap` operations; embeds at least 2 NDJSON example entries per kind (one minimal, one with code-fence content via JSON string escaping); the Go subcommand `crafter buffer uat|gap --run-dir <path>` exists, accepts the schema-1 fields as flags, and appends a single NDJSON line to `<run-dir>/<kind>-buffer.jsonl` (creating the file if missing). `crafter buffer uat|gap --help` runs without error.

### Phase 3 — Wire run-directory lifecycle and resolve GH#15 forward references

Land the workflow-level integration: document `.crafter/run/<task-id>/` lifecycle in `rules/do-workflow.md`, add `.crafter/run/` to `.gitignore`, and rewrite the eight forward-reference annotations in `rules/do-workflow.md` and `skills/crafter-do/SKILL.md` so they point at the real `crafter-buffer` skill and real file paths.

- [ ] **Step 1 — Document `.crafter/run/<task-id>/` lifecycle in `rules/do-workflow.md`.** Add a short subsection (location: near the top of the workflow rules or as a sibling of `### --auto`, whichever the implementer judges more discoverable — flag the chosen location for review). The subsection must state: (a) **when the run directory is created** — the decision between eager creation (at orchestrator run-start, before any buffer call) and lazy creation (on first `crafter-buffer` call) is made **inside this Phase 3 step**, not pre-committed by Phase 1; the implementer must evaluate both options against the existing crafter-do run-start sequence and pick one, justifying the choice in the commit message; the chosen option must be recorded as a `Decision (User Accepted)` (or `Decision (Orchestrator Accepted)` under `--auto`) entry in `## Decisions`, mirroring the Phase 1 pattern; (b) it persists for the duration of the run; (c) it is cleaned up after PR composition or workspace teardown — and "cleanup" means delete the directory, not just empty it; (d) it must be in `.gitignore` so the run files never leak; (e) it is per-run, not per-task across runs (resuming a task reuses the same `<task-id>` and therefore the same directory if it still exists); this follows the assumption recorded in `### Assumptions` (`<task-id>` is the task-file basename without extension), and Phase 3 Step 1 does not need to re-decide it. If Phase 3 implementation discovers the assumption is incorrect or insufficient, the implementer must surface it as a Decision entry before proceeding; (f) any per-run metadata artifact (if any — e.g., a `meta.json` run marker) and its content are also decided here in Phase 3, not pre-committed in Phase 1; if a per-run metadata artifact is added, its filename, schema, and lifecycle (creation timing, cleanup) must also be recorded as a `Decision` entry in `## Decisions`. The wording should explicitly link to the future PR-composer task (GH#17) for the cleanup trigger so the dependency is traceable.
- [ ] **Step 2 — Add `.crafter/run/` to the project's `.gitignore` and document the requirement for downstream projects.** The Crafter repo's own `.gitignore` currently contains only `tmp/`; append `.crafter/run/` to it. Then add a one-line note in the lifecycle subsection from Step 1: "Downstream projects MUST add `.crafter/run/` to their `.gitignore`." Do NOT introduce a new `templates/.gitignore` file in this task — the issue scopes this as "`.gitignore` (or template equivalent) excludes `.crafter/run/`", and adding a new template file is broader than required. If the user wants a gitignore template later, that is a separate task.
- [ ] **Step 3 — Resolve GH#15 forward references in `rules/do-workflow.md`.** Replace each "(forward reference to GH#16)" / "(GH#16)" annotation with a concrete pointer to `skills/crafter-buffer/SKILL.md` (or the relevant section of it). Sites to update (8 total based on the current grep): `rules/do-workflow.md` lines 36, 38, 87 (the section-level note may stay if it now describes resolved artifacts in past tense, or be removed entirely — implementer's call, justified in commit message), 110, 113, 114, 132. The wording must remain self-consistent with the surrounding `--auto` contract; do not change the semantics of any retained gate, removed gate, or invariant.
- [ ] **Step 4 — Resolve GH#15 forward reference in `skills/crafter-do/SKILL.md`.** Replace `*(buffer skill defined in companion task GH#16)*` at line 257 with a concrete pointer to `skills/crafter-buffer/SKILL.md`. The surrounding Step 6b prose for the `--auto` branch must remain byte-equivalent except for that annotation.

**Drift criteria:** Drift if (a) any rewrite changes the meaning of any retained or removed gate, (b) `.gitignore` rule is added without the documentation note (or vice versa), (c) a new `templates/.gitignore` file is created (out of scope for this task), (d) lifecycle wording introduces a cleanup trigger other than "PR composition or workspace teardown" without flagging it, (e) the `<task-id>` resolution rule is left ambiguous (e.g., "task slug" vs. "task filename" vs. "task ID" not pinned to one).

**Verification criteria:** `rules/do-workflow.md` contains the lifecycle subsection covering creation, persistence, cleanup, gitignore, and per-run identity; the project root `.gitignore` contains `.crafter/run/`; no remaining "GH#16" or "forward reference" annotations exist in `rules/do-workflow.md` or `skills/crafter-do/SKILL.md`; cross-checking SKILL.md/do-workflow.md against the new `skills/crafter-buffer/SKILL.md` shows no orphan pointers in either direction.

### Phase 4 — Verification round-trips and end-of-task housekeeping

Run the three verification round-trips the issue requires (append-then-read for each kind, code-fence-safe parsing, concurrency note), close the task file, and update STATE.md / skillbook / ARCHITECTURE.md per the standard end-of-task pattern.

- [ ] **Step 1 — Round-trip verification (UAT and Gap).** From a clean working tree, simulate or actually run a `crafter-buffer uat …` invocation followed by reading `.crafter/run/<task-id>/uat-buffer.jsonl` to confirm the entry parses back into its constituent fields without loss. Do the same for `crafter-buffer gap …`. Run each twice in succession on the same file to confirm the append (not overwrite) behavior and the inter-entry delimiter survives. Record the verification evidence as plain text in the task file's `## Outcome` section (or as a separate code block in the verification step's drift report).
- [ ] **Step 2 — Append-then-read round trip with awkward content.** Append three entries via `crafter buffer uat` (or `gap`) whose payload exercises JSON string escaping: (i) a `verify`/`detail` field containing a fenced code block (` ```bash ... ``` `) expressed as a JSON-escaped multi-line string; (ii) an entry containing literal triple backticks inline; (iii) an entry with embedded blank lines. Confirm the buffer file remains valid NDJSON: line count equals entry count (`wc -l <buffer>` matches the number of appends), each line parses as JSON (verification tool of the implementer's choice — `jq -c . <buffer>` if available, else `python3 -c "import json,sys; [json.loads(l) for l in sys.stdin]" <buffer>` — must succeed with no errors), and the parsed values round-trip back to the original strings.
- [ ] **Step 3 — Concurrency note verification.** Confirm the concurrency policy chosen in Phase 1 Step 3 is documented in both the skill (SKILL.md) and the workflow rules where relevant. If the policy is "sequential under the orchestrator," confirm `rules/do-workflow.md` agent-spawning sections do not implicitly enable parallel sub-agents touching the buffer. If the policy involves locking, confirm the lock idiom is documented and tested with two simulated concurrent appends.
- [ ] **Step 4 — End-of-task housekeeping.** Per `rules/post-change.md`: update `.crafter/STATE.md` Recent Changes with the new buffer skill and run-directory lifecycle; update `.crafter/ARCHITECTURE.md` to add `skills/crafter-buffer/` (and, if applicable, `cli/internal/buffer/` + the new cobra commands) to the Structure tree and to mention the skill under "Skills"; capture one or two skillbook observations (one Implementer-level, one orchestrator-level if appropriate); fill in `## Outcome` in the task file; commit the consolidated end-of-task changes.

**Drift criteria:** Drift if (a) verification is declared "passing" without recorded evidence in the task file or commit message, (b) ARCHITECTURE.md update omits the new skill or the new CLI surface (if Go-backed) — that is the single canonical structure document, (c) STATE.md update is skipped, (d) the commit bundles unrelated changes.

**Verification criteria:** All three verification round-trips have recorded evidence; `STATE.md`, `ARCHITECTURE.md`, task file `## Outcome`, and skillbook are all updated in a single consolidated end-of-task commit per `rules/post-change.md`; the eight issue acceptance criteria each map to a specific file/section.

### Karpathy Contract

**Scope boundaries:**
- This task delivers `skills/crafter-buffer/SKILL.md`, the run-directory lifecycle wording in `rules/do-workflow.md`, the `.gitignore` rule, the resolution of all GH#15 forward references, and (conditionally on Phase 1) `cli/cmd/buffer*.go` + `cli/internal/buffer/`.
- Editable surface: `skills/crafter-buffer/` (new), `rules/do-workflow.md`, `skills/crafter-do/SKILL.md` (annotation site only), `.gitignore` (root), `cli/` (conditionally), `.crafter/STATE.md`, `.crafter/ARCHITECTURE.md`, `.crafter/skillbook.json`, the task file itself.
- Off-limits: the installed copy under `~/.claude/crafter/` (per project rule); any agent file in `agents/`; `rules/core.md`, `rules/delegation.md`, `rules/post-change.md`, `rules/task-lifecycle.md` (no expected need to change them); `skills/crafter-debug/`, `skills/crafter-status/`, `skills/crafter-map-project/`; `templates/*.md`; `install.sh` (no change expected; if any required, surface as a stop condition).

**Non-goals:**
- PR composer extension that reads the buffers (GH#17).
- Agent prompt updates that switch sub-agents from "block on finding" to "append to buffer and continue" (GH#18).
- Persistence of buffers across runs, central cross-task buffer store, real-time UI surfacing, or a search/query CLI over buffers.
- A `templates/.gitignore` file or any installer-level gitignore management for downstream projects.
- Generalizing the skill beyond the two operations the issue defines (e.g., adding a `decisions` operation that competes with the task-file `## Decisions` section).
- Touching `decisions.md` — the issue mentions "`decisions.md` already exists; keep separate" but no such file exists in the codebase. The plan treats this as the issue meaning "the task file's `## Decisions` section, kept separate from the new buffers." Flagged in Risks.

**Drift checks:**
- Any architectural decision (schema shape, surface choice, concurrency model) made silently in implementation rather than in `## Decisions` first.
- Any expansion of the entry format beyond the issue-spec field set.
- Any shell-quoting fragility in the chosen append idiom that would break on entries with code fences.
- Any forward reference left unresolved after Phase 3.
- Any cleanup behavior beyond "delete the directory after PR composition or workspace teardown" (no archival, no zip, no transfer to STATE.md).

**Stop conditions:**
- Stop and ask if Phase 1 Step 1 cannot reach a clear schema decision because the issue's distinct field names appear to overlap (they shouldn't, but flag if they do).
- Stop and ask if Phase 1 Step 2 surfaces a third option not enumerated above (e.g., a tiny shell helper installed alongside the skill).
- Stop and ask if Phase 3 Step 1 reveals a conflict between "create at run start" and existing run-start steps in `crafter-do` (e.g., resume detection happens before any run-directory could exist).
- Stop and ask if `decisions.md` actually exists somewhere the planner missed (re-grep before implementing — see Risks).
- Stop and ask if the Go subcommand option is chosen and the implementer finds an unexpected interaction with `install.sh` binary distribution.

### Assumptions

- **The issue's "`decisions.md` already exists; keep separate" line refers to the task file's `## Decisions` section, not a literal `decisions.md` file** — no such file exists in the repo. Treating this as a documentation-only constraint to "keep the new buffers separate from how Decisions are recorded today." Flagged in Risks.
- **`<task-id>` resolution rule.** The plan assumes `<task-id>` is the task-file basename without extension (e.g., `20260509-feat-gh-16-buffer-skill`), matching the existing task-file naming convention in `.crafter/tasks/`. The implementer should confirm this in Phase 3 Step 1 and lock it in the documentation.
- **Phase 1 Step 1's default recommendation is "distinct schemas."** This is consistent with the issue's listing of distinct field names per kind and with the file-level separation. The user can override this default during Phase 1 Step 1.
- **Phase 1 Step 2's default recommendation is "hybrid (skill prose authoritative, Go subcommand for the deterministic append).** This matches the existing `skillbook` precedent (skill calls `crafter skillbook add`) and the project policy ("LLMs do JSON CRUD and atomic writes poorly"). The user can override this during Phase 1 Step 2 — and if "skill-only (shell-based)" is chosen, Phase 2 Step 2 is struck and Phase 4 Step 2 (code-fence safety) becomes the binding test.
- **Phase 1 Step 3's default recommendation is "sequential under the orchestrator (no locking)."** Sub-agents under the existing crafter-do flow are spawned one at a time, not in parallel; documenting this expectation is sufficient. If the user wants advisory locking, it adds work to Phase 2 Step 2.
- **Auto Mode is active.** Per the orchestrator's note, Auto Mode is currently active for this run. The plan accommodates this by recording all decisions in `## Decisions` (which is the right thing to do under any mode) and by relying on the `Decision (Orchestrator Accepted)` precedent (already established in GH#15) when Phase 1 decisions need to be made without an interactive pause.
- **No Reviewer/Implementer agent prompt changes are needed for this task.** The agents already read the rule files and skill files at run time; once `skills/crafter-buffer/SKILL.md` exists and `rules/do-workflow.md` is updated, the agents transparently pick it up. Prompt-level rewrites are GH#18.

### Risks / unknowns

- **`decisions.md` mention in the issue body.** The issue body says "`decisions.md` — already exists; keep separate" but the codebase has no such file. Two possible meanings: (1) the issue author meant "the task file's `## Decisions` section is the existing decisions surface — keep buffers separate from it"; (2) the issue author intended a `.crafter/run/<task-id>/decisions.md` file that this task should also create and document. The plan adopts meaning (1). If meaning (2) is correct, Phase 1 needs an extra step to scope a third buffer file and this task grows in surface. **Recommended: confirm with user before Phase 1 starts.**
- **Cleanup trigger ambiguity.** The issue says cleanup happens "after PR composition or workspace teardown." PR composition is GH#17, which is not delivered in this task. So in practice, in the immediate post-GH#16 world, cleanup happens only on "workspace teardown" (which is itself underspecified). The plan documents both triggers but does NOT implement a cleanup mechanism — the buffers will accumulate in `.crafter/run/` until GH#17 lands or until a human deletes the directory. This is intentional but worth flagging.
- **Run-start trigger ambiguity.** "Created at run start" is unambiguous in unattended runs (`--auto`) but less clear in interactive runs where the user may abort before the first buffer call. The plan recommends "lazy creation on first buffer call" as a simpler alternative that avoids creating empty directories for runs that never need them, and asks the user to confirm in Phase 3 Step 1. If the user prefers eager creation at run start, that adds a small step to `crafter-do`'s Step 0 / Step 1.
- **Concurrency policy unverified.** The plan defaults to "sequential under the orchestrator" because that matches the existing single-active-sub-agent pattern. If the future Symphony / CI integration spawns parallel sub-agents touching the same buffer, this policy will need to be revisited (likely as part of GH#18). Flagged so it does not get lost.
- **Go subcommand vs. skill-only is a real branching point.** If Phase 1 Step 2 picks Go-backed/hybrid, Phase 2 Step 2 ships Go code and the verification round-trip (Phase 4 Step 1) can run a real `crafter buffer uat` command. If Phase 1 Step 2 picks skill-only, Phase 2 Step 2 is skipped and Phase 4 Step 2 becomes more important (we're trusting agents to quote multi-line content correctly in shell). The plan supports either branch but the verification cost is asymmetric.
- **`tasks/` placement vs. `run/` placement.** Task files live in `.crafter/tasks/`; per-run buffers will live in `.crafter/run/<task-id>/`. These are deliberately separate because tasks survive runs (as resume state and historical record) while runs do not. Flagged for the implementer to keep distinct.
- **Forward-reference annotation count.** The current grep shows 8 sites referencing GH#16 (lines 36, 38, 87, 110, 113, 114, 132 in `rules/do-workflow.md` + line 257 in `skills/crafter-do/SKILL.md`). A new edit to either file before Phase 3 starts could add or remove sites; the implementer should re-grep at the start of Phase 3 to confirm the count, not blindly trust this plan's enumeration.

## Decisions

### Decision (User Accepted): Buffer entry schema — 2026-05-09

**Chosen:** Distinct schemas per kind (UAT, Gap), with shared base fields. Buffer file format is **NDJSON** (one JSON object per line; file extensions `.jsonl`). Each entry is a self-contained JSON object — no multi-line entry blocks.

**Per-entry shared base fields** (every entry has these):
- `id` — stable per-entry identifier (e.g., short hash or ULID; final scheme decided in Phase 2). Generated by the `crafter buffer` subcommand at append time and NOT a flag the caller passes in.
- `kind` — `"uat"` or `"gap"` (redundant with file but useful when entries are aggregated)
- `created_at` — ISO 8601 timestamp of when the entry was appended
- `created_by` — identity of the calling agent (e.g., `"crafter-implementer"`, `"crafter-reviewer"`)
- `task_id` — the originating task (matches `<task-id>` in the run-directory path)
- `title` — short human-readable headline
- `source` — pointer to the originating site (file:line, phase id, review finding id, etc.)

**UAT-specific fields:** `verify` (what to manually verify), `why_manual` (why the verification cannot be automated). (Use `snake_case` keys consistently in JSON.)

**Gap-specific fields:** `detail` (description of the gap / tech debt), `followup` (recommended action to close the gap).

**Per-run metadata is NOT part of this schema.** Anything aggregate (run-start timestamp, originating skill/flag, source branch, etc.) is a Phase 3 concern (run-directory lifecycle) and may be carried by separate run-level artifacts decided there.

**Alternatives considered:**
- Markdown blocks per entry (per issue body wording) — rejected: code-fence/quoting/multi-line escaping fragility; harder for downstream PR composer to parse; conflict with single-write atomic append.
- Unified schema with `kind: uat | gap` discriminator only — rejected: kind-specific fields differ.
- JSON array re-written on every append — rejected: read-modify-write race risk and bounded buffer scaling problems.

**Rationale:** NDJSON gives single-write atomic appends on POSIX (`O_APPEND` + bounded entry size), naturally handles code fences and multi-line content via JSON string escaping, and is trivially parseable by the future PR composer. The shared per-entry base provides triage ergonomics; per-run aggregate metadata stays out of this schema and is owned by Phase 3.

**Encoding contract:** UTF-8, LF (`\n`) line terminator, ISO 8601 UTC for `created_at` (suffix `Z`), JSON object key ordering not significant.

**Issue contract note:** The issue body (`https://github.com/richardriman/crafter/issues/16`) describes "Markdown block" entries with `.md` file extensions. This decision **deliberately deviates** to NDJSON / `.jsonl` for the reasons above. Approved by issue author (richardriman) on 2026-05-09 during fix-loop iteration #1.

### Decision (User Accepted): Implementation surface — 2026-05-09

**Chosen:** Hybrid — `skills/crafter-buffer/SKILL.md` documents the contract and prose; the deterministic append work is delegated to a new Go subcommand `crafter buffer uat|gap …` (mirroring `cli/cmd/skillbook*` structure). **Write strategy is NDJSON line-append:** one `O_APPEND | O_WRONLY | O_CREAT` open, encode the entry as a single-line JSON object, single `write(2)` of `<bytes>\n`. No temp file, no rename, no read-modify-write of existing buffer content.

**Target directory resolution:** the subcommand accepts a required `--run-dir <path>` flag (passed in by the orchestrator/calling skill) — analogous to `crafter skillbook --file`. This decision intentionally does NOT require the subcommand to auto-resolve `<task-id>` from the working tree; that responsibility stays with the caller, which already knows the active task. Phase 2 implements this contract.

The filename within the run-dir is fixed by kind: `crafter buffer uat ...` writes to `<run-dir>/uat-buffer.jsonl`, `crafter buffer gap ...` writes to `<run-dir>/gaps-buffer.jsonl`. The filename is not a flag.

**Alternatives considered:**
- Skill-only (rejected — fragile shell quoting around code fences and multi-line strings).
- Pure Go subcommand without a skill (rejected — agents read SKILL.md; the skill is the documented contract).
- Read-modify-write append on Markdown (rejected — concurrency/atomicity risk and scaling).

**Rationale:** Matches PROJECT.md Key Decisions (2026-03-24) policy of pushing deterministic file ops into the Go binary. NDJSON line-append is the simplest mechanism that satisfies single-write atomicity on POSIX without needing temp+rename or a lock manager.

### Decision (User Accepted): Concurrency policy — 2026-05-09

**Chosen:** Sequential writes under the orchestrator + single-syscall `O_APPEND` line-append. No lock manager. No temp+rename for the buffer files (NDJSON line-append makes it unnecessary).

**Atomicity assumption:** POSIX `write(2)` to a file opened with `O_APPEND` is atomic relative to other `O_APPEND` writers up to a per-platform size limit (`PIPE_BUF` is the conservative reference: at least 512 bytes on macOS, 4096 on Linux; for regular files, typical implementations atomic to at least the page size, but the formal POSIX guarantee is the conservative bound). Phase 2 enforces an entry-size cap below this conservative bound to keep the guarantee real. Cross-filesystem / NFS scenarios are explicitly **not supported** in this PoC; if Crafter ever needs to run on those, future work must add `fsync` + advisory locking.

**Resume semantics:** on resume of an existing task, `O_APPEND` to an existing buffer file is the natural primitive — no special handling needed for the buffer files themselves. Run-directory existence handling is a Phase 3 concern.

**Alternatives considered:**
- `flock` advisory locking (rejected — adds a lock-manager surface for a problem that does not exist today).
- Temp file + atomic `rename(2)` (rejected for buffers — would require read-modify-write of the whole file every append, defeating the "append-only" model).
- Fully unsynchronized writes (rejected — torn writes for entries near the size limit).

**Rationale:** Single `O_APPEND write(2)` of a length-bounded NDJSON line gives crash-safe append-only semantics with the simplest possible mechanism. Sequential expectation under the orchestrator is documented but not enforced by code; future Symphony/CI parallelism (GH#18 territory) can revisit if needed.

## Outcome

_(pending)_
