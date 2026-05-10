# Task: PR composer — Manual QA Plan / Known Gaps / Decisions sections (GH#17)

## Metadata
- **Date:** 2026-05-10
- **Work branch:** feat/gh-17-pr-composer
- **Status:** active
- **Scope:** Medium

## Request

Implement GitHub issue #17 — `[CHANGE] PR composer: include Manual QA Plan / Known Gaps / Decisions sections`.

The full contract is in the issue body at https://github.com/richardriman/crafter/issues/17. Summary:

**Context.** GH#15 introduced `--auto` mode and GH#16 introduced the per-run buffer infrastructure under `.crafter/run/<task-id>/`. The orchestrator's `--auto` flow ends with a PR step (`Plan → Execute → Verify → Review → PR end-to-end`), but no PR composer step is currently documented in the rules. This task introduces the PR composer step and wires it to the buffer files so that human reviewers see manual verifications, known gaps, and design decisions in the PR body.

**Issue body (verbatim):**

> When Crafter completes a run in `--auto` mode (#15), the PR description should include three additional sections derived from the run's buffers (#16) so that human reviewers see manual verifications, known gaps, and design decisions all in one place.
>
> Today the PR composer produces a standard summary; it does not read or include the buffer contents.

**Desired outcome.** Extend Crafter's PR composition step so that after the `gh pr create` body is built, it appends three optional sections — only when the corresponding source is non-empty:

```
## Manual QA Plan
<contents of .crafter/run/<task-id>/uat-buffer.jsonl, formatted as GitHub task list checkboxes>

## Known Gaps
<contents of .crafter/run/<task-id>/gaps-buffer.jsonl>

## Decisions
<contents of the task file's ## Decisions section>
```

Empty sources → corresponding section is omitted entirely (no empty headers).

**Note on issue text vs. reality:**

- The issue body refers to `uat-buffer.md` / `gaps-buffer.md` / `decisions.md`. GH#16 finalized the buffer files as `uat-buffer.jsonl` / `gaps-buffer.jsonl` (NDJSON). There is no `.crafter/run/<task-id>/decisions.md` — Decisions live in the **task file** under `## Decisions` (per `rules/task-lifecycle.md`). The PR composer must read from the actual locations: NDJSON buffers + task file `## Decisions` section.

**Acceptance criteria (from issue, adapted to actual file formats):**

- [ ] PR composer reads `.crafter/run/<task-id>/uat-buffer.jsonl` and includes content under `## Manual QA Plan` heading when non-empty
- [ ] PR composer reads `.crafter/run/<task-id>/gaps-buffer.jsonl` and includes content under `## Known Gaps` heading when non-empty
- [ ] PR composer reads the task file's `## Decisions` section and includes content under `## Decisions` heading when non-empty
- [ ] Empty sources produce no headers
- [ ] UAT entries are rendered as GitHub-flavored task list checkboxes (`- [ ] **Title** ...`) so reviewers can check them off in the PR UI
- [ ] Final PR body validates against existing PR template/lint rules
- [ ] Documentation explains that reviewers can check off `Manual QA Plan` items as they verify them

**Forward references that need resolving in this task:**

- `rules/do-workflow.md:142` — "Companion task GH#17 will add a second cleanup trigger after PR composition; until GH#17 lands, workspace teardown is the only active trigger." → this task wires the cleanup-after-PR trigger.
- `rules/do-workflow.md:146` — "this decision can be revisited in GH#17 or GH#18 if the PR composer finds it necessary" (re: per-run metadata artifact) → evaluate during planning whether a metadata artifact is needed; default expectation is no.

**Scope (from issue):**

- Update PR composer logic in workflow rules to read buffers + task-file Decisions and inject sections
- Define exact ordering and heading text
- Convert UAT entries to GitHub-flavored task list (`- [ ] **Title** ...`)
- Preserve existing PR body content (Summary, Validation, etc.)
- Document in PR template guidance
- Wire cleanup of `.crafter/run/<task-id>/` after PR composition

**Non-goals (from issue):**

- Editing existing open PRs (only applies on `gh pr create`)
- Rendering buffers as separate files attached to the PR
- Buffer translation/localization

**Context links:** Depends on #15 (`--auto`) and #16 (buffer skill). GH#18 (agent prompt updates) is out of scope.

## Plan

**Plan status:** approved

### Approach

This task delivers the GH#17 PR composer step: when `--auto` finishes a successful run with a green commit, the orchestrator must (1) build a PR body that includes the existing summary plus three optional appended sections (`## Manual QA Plan`, `## Known Gaps`, `## Decisions`) sourced from the per-run buffers and the task file's Decisions section, (2) call `gh pr create` with that body, and (3) clean up `.crafter/run/<task-id>/` after PR creation succeeds. The plan also resolves two GH#16 forward-reference hedges in `rules/do-workflow.md` (the cleanup-after-PR trigger on line 142 and the metadata-artifact "revisit in GH#17/#18" hedge on line 146).

The plan mirrors the GH#16 task file structure: architectural decisions are captured in `## Decisions` BEFORE any artifact is written, and each phase delivers a coherent, reviewable outcome. Three vertical phases:

- **Phase 1** records the four architectural decisions (rendering surface, Decisions extraction mechanism, baseline PR body shape, UAT checkbox rendering template) before any code or rule edit.
- **Phase 2** ships the rendering and assembly artifacts implied by Phase 1's surface decision (Go subcommand or shell-only) plus the section-rendering contracts. Tested in isolation against fixture buffers.
- **Phase 3** wires the rendering into the orchestrator workflow: adds the explicit PR composition step in `skills/crafter-do/SKILL.md`, documents PR-cleanup-trigger and resolves both forward references in `rules/do-workflow.md`, updates `.crafter/PROJECT.md`/`ARCHITECTURE.md`, and runs end-to-end smoke validation against the live task file.

### Why this approach

- **Decisions before construction (mirrors GH#16).** The "rendering-surface" question (orchestrator-only vs. Go subcommand vs. hybrid) is exactly analogous to GH#16's Decision 2. Locking it in `## Decisions` first keeps every later step traceable to one paragraph.
- **Skill remains the contract; CLI is plumbing if chosen.** Whatever Phase 1 picks, the orchestrator skill (`skills/crafter-do/SKILL.md`) stays the human-readable contract for when and how the PR composition step runs. A Go subcommand, if chosen, owns only deterministic NDJSON-to-Markdown rendering and Decisions extraction.
- **Render isolation before integration.** Phase 2 produces the rendering artifact and proves it with fixture inputs (sample buffers + sample task file) before Phase 3 wires it into `gh pr create`. This keeps the render contract reviewable independently of the workflow change.
- **One canonical paragraph for cleanup and metadata.** Both forward-reference hedges live in `rules/do-workflow.md`'s `### Run directory lifecycle` subsection; rewriting them in one Phase 3 step (rather than scattering edits) keeps the lifecycle contract self-consistent.
- **Smoke tests over heavy fixtures.** End-to-end PR-body validation uses the existing task file (or a tiny fixture run dir) — sufficient to prove the section assembly works against real artifacts, without inventing a new test framework.

### Phase 1 — Record architectural decisions before any artifact

Land four decisions in `## Decisions` before any code or rule edit. Each subsequent phase adapts to these decisions; if a decision changes, the affected step is re-scoped.

- [x] **Step 1 — Rendering-surface decision (Decision 1).** Decide where the rendering logic lives. Three enumerated options:
  - (a) **Orchestrator-only** — pure shell processing in the orchestrator skill (use `jq`/`grep`/`awk` to read NDJSON and the task file, build the appended sections inline, pass to `gh pr create --body`).
  - (b) **Go CLI subcommand** — new `crafter pr-body` subcommand reads `--run-dir` + `--task-file` and prints the assembled appended-sections Markdown (or the full body) to stdout; orchestrator pipes it into `gh pr create --body-file -`.
  - (c) **Hybrid** — Go CLI handles only the deterministic NDJSON-to-Markdown rendering of the two buffer files (one subcommand or two); orchestrator extracts the Decisions section from the task file via shell tooling and assembles the final body.
  Weighing factors must explicitly cover: (i) consistency with GH#16's hybrid decision (skill prose authoritative, Go for deterministic file ops), (ii) PROJECT.md policy "LLMs do JSON CRUD … poorly — a static binary handles these reliably", (iii) testability (Go gets table-driven `*_test.go` coverage; shell does not), (iv) shell-quoting fragility for entries with code fences and multi-line strings (the same risk that drove GH#16 toward Go), (v) minimum surface — `gh pr create` is invoked in shell anyway, so a small shell wrapper is unavoidable regardless of choice. Default recommendation if no further input: **Go CLI subcommand (option b)** mirroring GH#16's reasoning. Record chosen option as `Decision (User Accepted)` or `Decision (Orchestrator Accepted)` under `--auto`.

- [x] **Step 2 — Decisions extraction mechanism (Decision 2).** Decide how the `## Decisions` section is extracted from the task file. Two enumerated options:
  - (a) **Inside the chosen rendering surface** — if Decision 1 picks Go, the `crafter pr-body` subcommand reads the task file, scans for the `## Decisions` heading, and extracts everything until the next `## ` heading; if Decision 1 picks orchestrator-only, an `awk`/`sed` recipe in the skill does the same.
  - (b) **Pre-extraction in the orchestrator** — orchestrator runs a small extraction step regardless of Decision 1 and passes the extracted block as a `--decisions-file` flag to the renderer. Useful only if the renderer is Go and we want to avoid task-file path coupling. Evaluate which heading-extraction strategy survives edge cases: (i) `## Decisions` followed immediately by `## Outcome` (no content) → empty section, omit; (ii) `## Decisions` containing `### Decision (...)` H3 sub-headings → preserve verbatim; (iii) `## Decisions` followed by phase-verification H3 (`## Phase 4 — Verification evidence`) — the next `## ` heading boundary still works correctly. Default recommendation: option (a) co-located with chosen rendering surface. Record as `Decision`.

- [x] **Step 3 — Baseline PR body shape (Decision 3).** Crafter does not currently produce a PR body anywhere — `skills/crafter-do/SKILL.md` Step 6b ends with the per-phase commit and Steps 7-9 cover post-change housekeeping; no `gh pr create` is described. This task introduces the baseline. Decide explicitly:
  - The baseline-body sections and ordering (the canonical default is **Summary** + **Test plan** + the three appended sections; alternatives include adding **Validation** or **Changes** sections).
  - Whether the baseline body is generated by Crafter (LLM-composed Summary/Test-plan) or supplied pre-built by the caller and Crafter only appends the three sections.
  - The append order: `## Manual QA Plan` → `## Known Gaps` → `## Decisions` (matches issue body wording) — confirm and lock.
  Default recommendation: **Crafter generates a minimal Summary + Test plan baseline (LLM-composed in the orchestrator skill from the task file's Plan/Outcome), then appends the three issue-mandated sections in the order Manual QA Plan → Known Gaps → Decisions.** Record as `Decision`. This decision is the contract for Phase 3 Step 1 (the new PR composition step in `crafter-do/SKILL.md`).

- [x] **Step 4 — UAT checkbox rendering template (Decision 4).** Define the exact Markdown template for one UAT entry as a GitHub-flavored task list item, including how multi-line `verify` content is presented. Specify the template literally — every later step renders against this contract. Required choices:
  - **Title presentation:** `- [ ] **<title>**` (bold) or `- [ ] <title>` (plain).
  - **Verify content placement:** inline after an em-dash (`— <verify>`) for short verify text, vs. nested under the checkbox as a continuation paragraph (4-space indent) for multi-line verify text — and the rule for which to use (e.g., always nested? based on length? always inline with embedded `\n` becoming `<br>`?). The PR UI must still let reviewers check the box; nested content under the bullet preserves checkability in standard GFM.
  - **Whether to render `source` and `why_manual`** alongside (e.g., italic suffix, separate lines) or omit them entirely from the PR body. Issue acceptance criterion only mandates title + verify-as-checkbox; including source/why_manual is plan-defined polish.
  Default recommendation: `- [ ] **<title>** — <verify>` for single-line verify; for multi-line, `- [ ] **<title>**\n\n  <verify-with-2-space-indent>\n  \n  _Why manual:_ <why_manual>` with `source` omitted from the PR body (still present in the buffer). Record as `Decision`.

  **Define the Known Gaps rendering template as part of this step too** (the issue specifies "raw contents" but says nothing about formatting). Default: each gap entry rendered as a level-3 sub-heading or bullet — `- **<title>** — <detail>\n  - _Follow-up:_ <followup>` — with `source` omitted. Lock this as part of Decision 4 so Phase 2 Step 2 has an unambiguous contract.

**Phase 1 verification criteria:** `## Decisions` contains four new decision entries (rendering surface, Decisions extraction, baseline body shape, UAT+Gaps rendering templates) each with chosen option, alternatives considered, and rationale. Each decision references the issue/acceptance criterion or assumption it satisfies.

**Phase 1 drift criteria:** Drift if any decision is silently embedded in code or prose without a matching `## Decisions` entry; drift if the rendering-surface decision picks an option not enumerated above (e.g., a brand-new templating engine); drift if Decision 4 fails to specify both UAT and Gaps templates.

- [x] **Phase verification**
- [x] **Phase review**

### Phase 2 — Build the rendering artifact and prove it on fixtures

Land the rendering implementation chosen in Phase 1 Decision 1. The artifact must accept (run-dir, task-file) inputs and produce the three appended Markdown sections, with empty-source omission semantics. Validate against fixture inputs (sample buffer files + sample task file with sample `## Decisions` content) before any workflow wiring.

- [x] **Step 1 — Implement renderer for the two buffer files.** Read `<run-dir>/uat-buffer.jsonl` and `<run-dir>/gaps-buffer.jsonl`, skip the marker line (the first line whose `_marker` field is set per `skills/crafter-buffer/SKILL.md` "Creation behavior"), parse remaining lines as NDJSON, render each entry into the Markdown template chosen in Phase 1 Decision 4. Empty-after-marker, missing-file, and zero-byte-file cases must all produce no output (so the orchestrator can omit the corresponding `## Manual QA Plan` / `## Known Gaps` heading entirely). If Decision 1 chose Go, this lives in `cli/cmd/pr_body.go` (or two subcommands `pr-body uat`/`pr-body gap`) backed by `cli/internal/prbody/` (mirror `cli/internal/buffer/` layout — `types.go`, `format.go`, `format_test.go`). If Decision 1 chose shell, this lives as documented `jq` recipes in the relevant skill section. Output is **only the rendered Markdown content** (no `## Manual QA Plan` heading) — the heading is added by the assembler in Step 3 only when content is non-empty, so the empty-source rule remains exactly one place.

- [x] **Step 2 — Implement Decisions extraction.** Per Phase 1 Decision 2: locate the `## Decisions` heading in the task file, capture all content until the next `## ` heading (or EOF), trim leading/trailing blank lines, and return the body. If the section is missing, contains only blank lines, or contains only HTML comments / placeholders, return empty (so the assembler omits the heading). Edge cases the implementation MUST handle: (i) `## Decisions` followed immediately by another `## ` heading; (ii) `### Decision (...)` H3 sub-headings inside the section preserved verbatim; (iii) Decisions section appearing at end of file with no following `## ` heading; (iv) empty content between heading and next `## ` (just blank lines). If Decision 1 chose Go, this lives next to Step 1's renderer; if shell, the recipe is documented in the orchestrator skill.

- [x] **Step 3 — Implement section assembler.** Combine outputs from Steps 1 and 2 into the final appended-sections block, applying the empty-source omission rule (skip the heading entirely if the corresponding rendered content is empty), and the section ordering from Phase 1 Decision 3. The assembler outputs only the appended Markdown — the baseline body (Summary + Test plan) is constructed separately by the orchestrator (per Decision 3) and concatenated with the assembler output to form the full PR body. If Decision 1 chose Go, the assembler is the entry-point subcommand (e.g., `crafter pr-body --run-dir <path> --task-file <path>`); if shell, it is the orchestrator skill section that calls the renderers and concatenates.

- [x] **Step 4 — Fixture validation.** Build a minimal fixture: a temp run-dir with a marker-only `uat-buffer.jsonl`, a marker+two-entries `uat-buffer.jsonl`, a missing `gaps-buffer.jsonl`, and a tiny task-file with a non-trivial `## Decisions` section. Exercise all matrix cells: (a) all three sources empty → no appended sections; (b) only UAT non-empty → only `## Manual QA Plan` heading appears; (c) all three non-empty → all three headings in correct order; (d) UAT entry with multi-line `verify` containing fenced code blocks → renders correctly per Decision 4; (e) Decisions section with embedded `### Decision (...)` H3 sub-headings → preserved verbatim. Record evidence as plain text in this step's verification report. If Decision 1 chose Go, also add `*_test.go` table-driven cases mirroring the GH#16 `format_test.go` style covering the same matrix (marker-skip, empty-after-marker, multi-entry, code-fence escaping, oversized-input edge — the renderer should not impose a per-entry size cap since it reads, not writes, but should refuse to crash on any byte sequence the buffer can hold).

**Phase 2 verification criteria:** Renderer runs on the fixture, produces exactly the expected Markdown for each matrix cell. If Go-backed: `go build ./...` and `go test ./...` pass; the new subcommand's `--help` lists required flags. The marker line is skipped in both buffer files. Empty sources produce empty output. Multi-line and code-fence content survive rendering without quoting damage.

**Phase 2 drift criteria:** Drift if (a) any non-empty source's heading is omitted, or any empty source's heading appears, (b) the marker line is rendered as a buffer entry, (c) the section ordering deviates from Phase 1 Decision 3, (d) the renderer parses the buffer file in any way other than line-oriented NDJSON (e.g., loads the file as a JSON array), (e) Decisions extraction silently truncates content or includes content from after the next `## ` heading, (f) the renderer or assembler produces output that includes a heading concatenated to a non-empty body but missing a blank line before the next heading (PR body must remain valid Markdown).

- [x] **Phase verification**
- [x] **Phase review**

### Phase 3 — Wire PR composition into the workflow and resolve forward references

Land the orchestrator-side wiring: introduce the PR composition step in the `--auto` flow (and document it for non-`--auto` flows where relevant), wire the cleanup-after-PR trigger, resolve both GH#16 forward references in `rules/do-workflow.md`, and update `.crafter/` docs.

- [x] **Step 1 — Add the PR composition step to `skills/crafter-do/SKILL.md`.** Decide and document the placement: under `--auto`, the PR step runs after the FINAL phase's per-phase commit lands and before/instead of the current Steps 7-9 housekeeping (the current `--auto` Step 6b branch ends with a commit per `post-change.md` and then implicitly proceeds to Steps 7-9; the PR step interleaves between the final phase's commit and the consolidated end-of-task commit, OR runs after both — implementer must pick and justify in the commit message, and record as `Decision (Orchestrator Accepted)`). The new step must specify:
  - **Trigger:** runs only under `--auto` after the final phase commits clean (and after the consolidated end-of-task commit lands, if Decision in this step picks that ordering). Non-`--auto` runs continue to use the manual `gh pr create` path the user invokes themselves; mention this explicitly so reviewers know the step is `--auto`-only.
  - **Inputs:** `--run-dir` (already known to the orchestrator as `.crafter/run/<task-id>/`), the task file path, and the chosen baseline body (per Phase 1 Decision 3 — orchestrator-composed Summary + Test plan).
  - **Action:** invoke the rendering artifact from Phase 2 to produce the appended sections; concatenate with the baseline body; call `gh pr create --title "<title>" --body-file -` (heredoc or `printf` piped into stdin) per Crafter's existing `gh` usage idiom.
  - **Title derivation:** short, conventional-commit-style title derived from the task file metadata (e.g., the task topic + GH issue number). Document the derivation rule (one short sentence, e.g., "use the first line of the most recent commit on the work branch, or the task file's H1").
  - **Failure handling:** if `gh pr create` fails (network, auth, branch already has open PR, etc.), record the failure as a `Decision (Auto-Recorded): PR creation failed — ...` in the task file's `## Decisions` section, do NOT trigger the cleanup hook (Step 2), and exit with state per the four-retained-gates contract (this is an Ad-hoc escape hatch case per `rules/do-workflow.md`).
  - **Success handling:** on `gh pr create` success, return the PR URL to the user (one-line notice) and proceed to Step 2 (cleanup hook).

- [x] **Step 2 — Wire the cleanup-after-PR trigger and resolve `do-workflow.md` line 142 forward reference.** In `rules/do-workflow.md → ### Run directory lifecycle`, replace the "Companion task GH#17 will add a second cleanup trigger after PR composition; until GH#17 lands, workspace teardown is the only active trigger." sentence with definitive wording that lists both triggers as active: (a) **after successful `gh pr create`** under `--auto` (the new trigger this task adds), and (b) **on workspace teardown** (the existing trigger). Both perform the same `rm -rf .crafter/run/<task-id>/` action. Specify trigger ordering: PR-success cleanup is the primary trigger under `--auto`; teardown cleanup is the safety-net for runs that exit via one of the four retained gates before reaching PR composition. Implement the cleanup hook itself as part of the new PR composition step in `crafter-do/SKILL.md` (Phase 3 Step 1) — only after `gh pr create` returns success.

- [x] **Step 3 — Resolve the metadata-artifact hedge (`do-workflow.md` line 146).** Evaluate whether the PR composer needs a per-run metadata artifact (`meta.json` or equivalent) to function. Default expectation per the orchestrator's task prompt: **no.** The PR composer reads (i) the two buffer files, (ii) the task file's `## Decisions` section, (iii) optionally the most recent commit metadata for the title — none of which require a per-run metadata file. Record this as `Decision (Orchestrator Accepted): No per-run metadata artifact needed for GH#17 — confirmed.`. Then update `rules/do-workflow.md` line 146 to remove the "this decision can be revisited in GH#17 or GH#18 if the PR composer finds it necessary" hedge, replacing it with a definitive past-tense statement: "GH#17 confirmed no metadata artifact is needed; this remains the policy until a future change demonstrates a concrete need." If, against expectation, Phase 1 or Phase 2 surfaces a real need for a metadata artifact, STOP and surface it to the user as a scope expansion before proceeding (it would mean adding a fifth Phase 1 decision and is out of the current scope).

- [x] **Step 4 — End-to-end smoke validation.** From a clean working tree (or a controlled fixture), simulate the `--auto` end-of-run path: create a small fixture run-dir with non-trivial UAT and Gap buffer entries, point the renderer at the live task file (which already has a non-trivial `## Decisions` section), and produce the assembled PR body. Manually inspect the output for: (i) all three appended sections present and in correct order; (ii) UAT entries render as GFM checkboxes that GitHub will treat as actionable; (iii) the Decisions section preserves `### Decision (...)` H3 sub-headings and embedded markdown formatting verbatim; (iv) no marker lines appear; (v) the body validates as Markdown (no broken heading levels, no missing blank lines around fenced code blocks). Record evidence as plain text in this step's verification report. Do NOT actually call `gh pr create` against GitHub during this smoke (the validation is on the body string). The first real `gh pr create` invocation happens when this very task ships its own PR — that becomes the dogfood validation.

- [x] **Step 5 — Documentation updates.** Update `.crafter/PROJECT.md` "Key Decisions" table with a 2026-05-10 row for the GH#17 PR composer decision (rendering surface). Update `.crafter/ARCHITECTURE.md` to add (a) the new CLI subcommand under "Crafter CLI — Utility Binary → Current subcommands" if Phase 1 chose Go, (b) the new `cli/internal/prbody/` and `cli/cmd/pr_body.go` to the Structure tree (if applicable), (c) a short paragraph under "Key Patterns & Decisions" describing the PR composer step in the `--auto` workflow with a pointer to `skills/crafter-do/SKILL.md`. Add an entry to `.crafter/STATE.md` "Recent Changes" describing the GH#17 delivery (commit hash filled in by post-commit backfill per the GH#16 precedent — `5e8ceba`-style chore commit). Document in the new `crafter-do/SKILL.md` PR step that reviewers can check off `## Manual QA Plan` items in the PR UI as they verify them (issue acceptance criterion 7).

**Phase 3 verification criteria:** `skills/crafter-do/SKILL.md` contains a clearly placed and named PR composition step describing the trigger, inputs, action, title derivation, failure handling, and success handling. `rules/do-workflow.md → ### Run directory lifecycle` contains both cleanup triggers (PR-success and workspace-teardown) without any "until GH#17 lands" hedge. `rules/do-workflow.md` line 146 metadata-artifact hedge is replaced with definitive language. `.crafter/PROJECT.md`, `.crafter/ARCHITECTURE.md`, and `.crafter/STATE.md` all reflect the change. End-to-end smoke produces a valid PR body string against fixture+task-file inputs. All seven issue acceptance criteria each map to a specific file/section/decision.

**Phase 3 drift criteria:** Drift if (a) the PR step is added but not gated to `--auto` (i.e., would also fire in default or `--fast` runs without explicit user opt-in), (b) the cleanup hook fires before `gh pr create` returns success (cleanup must be conditional on PR success — failed PR runs leave the run-dir intact for retry/debug), (c) either forward-reference hedge is left intact in `rules/do-workflow.md`, (d) the rendering output diverges from the contract defined in Phase 1 Decision 4 (e.g., title bolding or verify-placement deviates silently), (e) the PR step's failure handler swallows the `gh pr create` error instead of routing it through the Ad-hoc escape hatch, (f) any new file is placed outside the surfaces listed in the Karpathy Contract below.

- [x] **Phase verification**
- [x] **Phase review**

### Karpathy Contract

**Scope boundaries:**
- This task delivers (a) the rendering artifact chosen in Phase 1 Decision 1 (Go subcommand `crafter pr-body` and supporting `cli/internal/prbody/` package, OR documented shell recipes in the orchestrator skill), (b) the new PR composition step in `skills/crafter-do/SKILL.md`, (c) the cleanup-after-PR hook implementation, (d) resolution of both forward-reference hedges in `rules/do-workflow.md`, and (e) the doc updates in `.crafter/PROJECT.md`, `.crafter/ARCHITECTURE.md`, and `.crafter/STATE.md`.
- Editable surface: `skills/crafter-do/SKILL.md` (PR step addition), `rules/do-workflow.md` (forward references + cleanup trigger), `cli/cmd/pr_body.go` + `cli/internal/prbody/` (new, conditional on Decision 1), `cli/cmd/pr_body_test.go` and `cli/internal/prbody/*_test.go` (new, conditional on Decision 1), `install.sh` (only if Decision 1 picks Go and a new subcommand needs no separate enumeration since `cli/Makefile` is `go build .` — confirm during Phase 2 Step 1; install.sh skill enumeration does NOT need updating since no new skill is introduced), `.crafter/PROJECT.md`, `.crafter/ARCHITECTURE.md`, `.crafter/STATE.md`, the task file itself.
- Off-limits: the installed copy under `~/.claude/crafter/` (per project rule); any agent file in `agents/`; `rules/core.md`, `rules/delegation.md`, `rules/post-change.md`, `rules/task-lifecycle.md` (no expected need to change them — if needed, surface as a stop condition); `skills/crafter-debug/`, `skills/crafter-status/`, `skills/crafter-map-project/`, `skills/crafter-buffer/` (the buffer skill's marker-skip contract is consumed but not modified); `templates/*.md`; `.gitignore` (already has `.crafter/run/`).

**Non-goals:**
- Editing the body of an already-open PR (only applies on `gh pr create`, per issue non-goals).
- Rendering buffers as separate file attachments to the PR (per issue non-goals).
- Buffer translation/localization (per issue non-goals).
- Switching agent prompts from blocking to buffer-append (GH#18).
- Defining when sub-agents append to buffers — that is GH#16's contract, already in place.
- Adding a per-run metadata artifact (`meta.json` etc.) — explicitly rejected in Phase 3 Step 3 default decision.
- Adding new skill files (no new skills are introduced — the rendering, if Go, is a CLI subcommand; the workflow change lives in the existing `crafter-do` skill).
- Generalizing the PR composer to non-`--auto` runs (default-mode users compose PRs manually today; this task does not change that flow).
- Writing a `templates/.gitignore` or installer-level git-ignore management.

**Drift checks:**
- Any architectural decision (rendering surface, extraction mechanism, baseline shape, rendering template) made silently in code rather than recorded in `## Decisions` first.
- Any deviation from the issue-mandated section ordering (`Manual QA Plan` → `Known Gaps` → `Decisions`).
- Any case where an empty source produces a heading, or a non-empty source omits its heading.
- Cleanup firing before `gh pr create` succeeds, OR cleanup not firing on PR success.
- Forward-reference hedges in `rules/do-workflow.md` left intact after Phase 3.
- New surface added outside the editable list above.

**Stop conditions:**
- Stop and ask if Phase 1 Decision 1 surfaces a fourth surface option (e.g., a Python helper, a templated `gh pr edit` flow) not enumerated above.
- Stop and ask if Phase 2 Step 4 fixture validation reveals a buffer-format edge case the GH#16 marker-skip contract doesn't cleanly handle (e.g., legacy buffer files without a marker line — should not exist post-GH#16, but flag if found).
- Stop and ask if Phase 3 Step 1 placement of the PR step in `crafter-do/SKILL.md` reveals an interaction with Step 6a's session-break suggestion or with the consolidated end-of-task commit ordering that the plan did not anticipate.
- Stop and ask if Phase 3 Step 3 finds a real need for a per-run metadata artifact (would expand scope significantly).
- Stop and ask if Decision 3's baseline-body shape requires the Planner or another agent to compose the Summary/Test-plan content (the default expects the orchestrator skill to template it from the task file alone; if more is needed, surface as scope expansion).

### Assumptions

- **`<task-id>` resolution is unchanged from GH#16.** `<task-id>` is the task-file basename without extension; the orchestrator already tracks it. The PR composer consumes it as `--run-dir .crafter/run/<task-id>/` and as the task-file path `.crafter/tasks/<task-id>.md`.
- **Buffer files always start with the marker line per GH#16.** `skills/crafter-buffer/SKILL.md` "Creation behavior" guarantees this; the renderer skips line 1 if `_marker` is set.
- **`## Decisions` heading-extraction is unambiguous.** Task files always have a `## Decisions` heading followed eventually by another `## ` heading (Outcome) or EOF. Empty Decisions sections are common (fresh task file) and must produce empty output. The current GH#16 task file is a strong example with `### Decision (...)` H3 sub-headings.
- **`gh pr create` is available and authenticated under `--auto`.** Per `rules/do-workflow.md → #### Ad-hoc escape hatch`, missing auth is a recognized escape-hatch trigger; the PR step's failure handling routes there if `gh` is missing or unauthenticated.
- **`--auto` is the only mode that triggers PR composition.** Default and `--fast` modes do not call `gh pr create` automatically (the user does it manually); only `--auto` adds the new step. This matches the `Plan → Execute → Verify → Review → PR end-to-end` wording in `rules/do-workflow.md → ### --auto`.
- **The baseline body is orchestrator-composed by default.** Per Phase 1 Decision 3 default, the Summary + Test plan baseline is generated by the orchestrator skill from task-file content (Plan + Outcome). If a future task wants the Planner agent to compose the body, that is a separate change.
- **Phase 1 Step 1 default is "Go CLI subcommand."** Mirrors GH#16 Decision 2 (hybrid → Go for deterministic file ops). User can override.
- **Phase 1 Step 4 default UAT template is `- [ ] **<title>** — <verify>` for short verify, nested-paragraph for multi-line.** This satisfies the issue acceptance criterion 5 (reviewers check items in PR UI). User can override the exact template formatting.
- **Auto Mode is active for this run.** All four Phase 1 decisions land as `Decision (Orchestrator Accepted)` with rationale, mirroring the GH#16 precedent for decisions made under `--auto`.

### Risks / unknowns

- **`gh pr create` integration point is currently undocumented.** The `--auto` flow ends with the green-commit per `post-change.md` and proceeds to Steps 7-9; there is no existing PR step to extend — this task introduces it. The placement (between final-phase commit and consolidated end-of-task commit, vs. after both) has implications for how the consolidated commit interacts with the open PR. Phase 3 Step 1 must pick and justify; flagging as the highest-risk scoping question in this plan.
- **Title derivation rule is under-specified in the issue.** The issue says nothing about PR title; the plan defaults to "first line of latest commit on work branch OR task file H1," but this is a planner-imposed convention. If the user wants a richer rule (e.g., LLM-composed from task scope), it adds a small step.
- **Decisions section can be very long.** The GH#16 task file's `## Decisions` section is ~100 lines with multiple `### Decision (...)` H3 sub-headings. Including it verbatim in PR body is correct per the issue, but reviewers may find it overwhelming. The plan does NOT add a length cap or summarization — that would be a UX decision beyond the issue's contract — but flagging so the user can choose to add a follow-up issue.
- **Multi-line UAT verify content with fenced code blocks could break PR-body Markdown rendering.** GitHub's PR body renderer is mostly GFM-compatible but not identical. Phase 2 Step 4 fixture validation must include at least one fenced-code-block UAT entry; if rendering drifts, Decision 4 may need revision.
- **Empty-source detection for the buffer files.** A buffer file with only the marker line (no entries) is structurally non-empty but semantically empty. The renderer must check post-marker line count, not file size or existence. Phase 2 Step 1 makes this explicit; flagging as a likely mistake-vector during implementation.
- **No existing PR-creation flow to extend.** Some Crafter contributors today run `gh pr create` manually after `--auto` completes. Introducing automatic PR creation under `--auto` is a behavior change for those users, even though `--auto` is documented as ending with PR. If anyone has scripted around the absence of automatic PR creation, this would surprise them. Low likelihood (the workflow is new), flagging for completeness.
- **Cleanup hook racing with the consolidated end-of-task commit.** If the PR composition step runs before the consolidated end-of-task commit, the run-dir is gone before STATE.md/skillbook updates happen — those updates do not depend on the run-dir, so this is fine. If it runs after, the run-dir persists slightly longer. Phase 3 Step 1's ordering choice locks this.
- **Forward-reference grep coverage.** Two known sites are flagged in the prompt (lines 142, 146 in `rules/do-workflow.md`). Phase 3 Step 2/3 must re-grep at start to confirm no other forward references to GH#17 exist in `skills/`, `rules/`, or `agents/` (Phase 2 may have added one as part of code comments). The implementer should not blindly trust the plan's enumeration.

## Decisions

### Decision 1 — Rendering surface (Go CLI subcommand) (Orchestrator Accepted) — 2026-05-10

**Chosen:** Option (b) — **Go CLI subcommand `crafter pr-body`**. A new `cli/cmd/pr_body.go` entry-point subcommand (backed by `cli/internal/prbody/`) reads `--run-dir` + `--task-file`, renders the three appended Markdown sections, and prints the result to stdout. The orchestrator pipes that output into `gh pr create --body-file -`.

**Alternatives considered:**

- **(a) Orchestrator-only shell processing** — `jq`/`grep`/`awk` recipes inline in the orchestrator skill. Rejected: (i) fragile shell quoting when UAT `verify` or Decisions prose contains fenced code blocks, multi-line strings, or special characters (the same failure mode that drove GH#16 toward Go); (ii) no unit-test coverage possible — a shell recipe is verified only by running it end-to-end; (iii) inconsistent with PROJECT.md Key Decisions (2026-03-24) policy: "LLMs do JSON CRUD, Jaccard similarity, and atomic writes poorly — a static binary with zero runtime deps handles these reliably." NDJSON parsing + Markdown rendering fit that category.

- **(c) Hybrid** — Go renders the two NDJSON buffer files; orchestrator extracts the Decisions section from the task file via `awk`/`sed` and assembles the final body. Rejected: splitting the rendering surface across two mechanisms complicates testing (each half is tested separately but the assembly contract is tested only end-to-end), and the Decisions extraction via `awk` is subject to the same quoting fragility risk for H3 sub-heading content. Keeping everything in one Go subcommand gives a single, testable entry point.

**Rationale:**

(i) **Consistency with GH#16's hybrid → Go progression.** GH#16 Decision 2 started from a hybrid framing (skill prose authoritative, Go for deterministic file ops) and resolved to Go-backed for every deterministic operation. The same logic applies here: the orchestrator skill remains the human-readable contract for *when* PR composition runs; Go owns the *how* of NDJSON-to-Markdown rendering and Decisions extraction.

(ii) **PROJECT.md policy.** The 2026-03-24 Key Decision explicitly names "JSON CRUD" and "atomic writes" as tasks the Go binary handles reliably. NDJSON parsing (reading `uat-buffer.jsonl` and `gaps-buffer.jsonl`) and Markdown assembly are deterministic file operations of the same character — they belong in the binary.

(iii) **Testability.** A Go subcommand backed by `cli/internal/prbody/` gets table-driven `*_test.go` coverage mirroring `cli/internal/buffer/format_test.go` from GH#16: matrix cells for empty sources, marker-only files, multi-entry files, code-fence content, and Decisions H3 sub-headings. Shell recipes cannot be unit-tested; they are verified only by executing the full `gh pr create` path.

(iv) **Shell-quoting fragility.** Buffer entries may contain fenced code blocks and multi-line `verify` strings (Phase 4 Step 2 of GH#16 confirmed this in real round-trips). Passing those values through shell variable expansion and heredoc construction is error-prone. A Go subcommand reads the NDJSON file bytes directly and emits Markdown without passing through a shell interpreter — quoting fragility is eliminated at the source.

(v) **Minimum surface.** `gh pr create` is invoked in shell regardless of this choice — a small shell wrapper remains unavoidable. Option (b) adds exactly one binary invocation (`crafter pr-body --run-dir ... --task-file ...`) before that wrapper; the orchestrator skill still controls the overall `gh pr create` call. The extra step is minimal and follows the exact pattern already established by `crafter buffer uat|gap`.

**Trade-offs accepted:**

- The Go subcommand expands the CLI binary surface by one subcommand and one internal package. This is the same trade-off made in GH#16 and is consistent with the existing project posture.
- `crafter pr-body` output is pure Markdown to stdout; the orchestrator is responsible for prepending the baseline body (Summary + Test plan) before passing the combined body to `gh pr create`. This separation of concerns is explicit and is locked in Phase 1 Decision 3.
- If `go build ./...` breaks for any reason (e.g., Go version incompatibility in CI), the PR composition step also breaks — the same risk as all other Go-backed skills.

### Decision 2 — Decisions extraction mechanism (Co-located with Go renderer) (Orchestrator Accepted) — 2026-05-10

**Chosen:** Option (a) — **Co-located with the Go renderer**. The `crafter pr-body` subcommand reads the task file directly, scans for the `## Decisions` heading, captures all content until the next `## ` heading (or EOF), and returns it as a string. The extracted content is assembled with the UAT and Gaps rendered output inside the same Go entry point. No separate extraction step, no additional CLI flag.

**Alternatives considered:**

- **(b) Pre-extraction in the orchestrator** — a shell `awk`/`sed` recipe extracts the Decisions block before invoking the renderer and passes it via a `--decisions-file` flag. Rejected: (i) introduces the same shell-quoting fragility that led to Go being chosen for NDJSON rendering in Decision 1 — the Decisions section contains fenced code blocks and multi-line content inside `### Decision (...)` H3 sub-headings (confirmed by the GH#16 task file's own `## Decisions` section, which is a real test case for edge case ii); (ii) adds an extra orchestrator-side step and an extra flag for a purely deterministic operation that Go can handle directly; (iii) splits the rendering surface — shell extracts one of the three sources while Go handles the other two — without any benefit, since Decision 1 already gave Go the responsibility for all deterministic file reading.

**Rationale:**

Decision 1 chose the Go subcommand precisely because deterministic file operations belong in the binary, not in shell. Heading-extraction from a Markdown file is a deterministic file operation of the same character as NDJSON parsing. Keeping it inside the same subcommand gives one self-contained entry point (`crafter pr-body --run-dir <path> --task-file <path>`) that the orchestrator invokes once, with one testable code path.

The four edge cases the implementation must handle:

(i) **`## Decisions` followed immediately by `## Outcome` (empty content).** The heading-boundary scanner advances to the next `## ` line without collecting any content. The returned string is empty (after trimming), so the assembler omits the `## Decisions` heading from the PR body entirely. No empty heading appears.

(ii) **`## Decisions` containing `### Decision (...)` H3 sub-headings.** The scanner captures lines verbatim until the next `## ` heading; H3 sub-headings are `### ` lines, which do not match the `## ` boundary condition. All H3 sub-heading lines — including the decision entries with their fenced code blocks and multi-line rationale — are preserved without modification. The GH#16 task file (`20260509-feat-gh-16-buffer-skill.md`) is the reference test case for this edge case.

(iii) **`## Decisions` followed by a phase-verification H3 such as `## Phase 4 — Verification evidence`.** The boundary is `## ` (two hashes + space), which matches the next H2-level heading regardless of its content. A `### Phase` line (three hashes) does not match and is captured as content; a `## Phase 4` line (two hashes) terminates extraction. The scanner correctly stops at the right boundary in both scenarios.

(iv) **Decisions section at end of file with no following `## ` heading.** The scanner reads until EOF; there is no boundary line, so all remaining lines after `## Decisions` are captured. This is the common state during active development when `## Outcome` has not yet been written. The trimmed non-empty content is returned and assembled into the PR body normally.

**Trade-offs accepted:**

- The Go subcommand now has a dependency on the task-file path in addition to the run-dir path. Both are already known to the orchestrator (it tracks `<task-id>` throughout the run), so the `--task-file` flag adds no new information-gathering step.
- Heading-extraction logic lives in the Go binary; if the task-file format ever changes (e.g., switching from `## Decisions` to `## Design Decisions`), the Go code must be updated alongside the rule. This is the same coupling cost as any other format assumption and is acceptable given the stability of the task-file format.
- Option (b)'s `--decisions-file` pattern would have allowed injecting arbitrary Decisions content independently of the task file (useful for testing). Under option (a), fixture testing of the extraction path uses a small temporary task-file fixture, which is equally straightforward.

### Decision 3 — Baseline PR body shape (Summary + Test plan + three appended sections, Crafter-composed) (Orchestrator Accepted) — 2026-05-10

**Chosen sub-choices:**

**(a) Baseline-body sections and ordering:** **Summary** + **Test plan** + three appended sections (`## Manual QA Plan` → `## Known Gaps` → `## Decisions`). No additional sections (`Validation`, `Changes`, etc.) are included in the baseline.

**(b) Who composes the baseline body:** **Crafter (orchestrator skill) generates the Summary and Test plan** from the task file's `## Plan` section content (Approach + phase/step descriptions) and `## Outcome` section (if present; otherwise elided). The caller does not supply a pre-built body. `crafter pr-body` outputs only the three appended sections; the orchestrator skill prepends the LLM-composed Summary + Test plan baseline before passing the combined body to `gh pr create`.

**(c) Append order of the three GH#17 sections:** `## Manual QA Plan` → `## Known Gaps` → `## Decisions`. This order matches the issue body verbatim and is confirmed and locked.

**Alternatives considered:**

- **(a-alt) Add `## Validation` or `## Changes` sections to the baseline.** Rejected: neither section is mandated by the issue, and neither is consistently applicable across all task types. Adding them now would require Crafter to know how to populate them from the task file, which is additional LLM-composition surface not warranted by the acceptance criteria. They can be added in a follow-up if a specific need is demonstrated.

- **(b-alt) Caller supplies a pre-built baseline body; Crafter only appends the three sections.** Rejected: `--auto` mode is fully unattended — there is no interactive caller to pre-compose the body. Requiring a pre-built body would break the `Plan → Execute → Verify → Review → PR end-to-end` contract by introducing a manual step (body drafting) that defeats unattended orchestration. The orchestrator skill already has access to the task file, which contains sufficient content (Approach, phases, Outcome) to compose a minimal summary.

- **(b-alt-2) Planner agent or a dedicated agent composes the Summary/Test-plan.** Rejected: this would require spawning another agent at PR composition time, expanding scope beyond GH#17. The orchestrator skill itself is the right place for a brief, template-driven LLM composition step that reads the task file's Plan and Outcome sections — no additional agent delegation is necessary. (This is the stop condition: if the Planner is needed, scope must be re-evaluated; the default expectation is that the orchestrator skill templates it from the task file alone.)

- **(c-alt) Different append order.** The issue body lists the sections in the order `Manual QA Plan` → `Known Gaps` → `Decisions`. There is no reason to deviate from the issue wording. The ordering is confirmed.

**Rationale:**

**(a) Minimal baseline is sufficient.** The issue acceptance criteria focus entirely on the three appended sections (UAT, Gaps, Decisions). A Summary + Test plan baseline is the conventional minimum for a human-readable PR body. Adding more baseline sections increases LLM composition complexity without a corresponding acceptance-criterion payoff.

**(b) Orchestrator-composed body preserves `--auto` unattended contract.** The `--auto` mode's core value is end-to-end execution without human intervention. Requiring the caller to supply a pre-built baseline breaks that contract. The task file's `## Plan → Approach` paragraph and phase/step list provide enough structured content for the orchestrator skill to emit a compact, informative Summary and a Test plan that lists the key acceptance criteria from the task file. This is exactly the same "read the task file and summarize" operation the orchestrator already performs in other contexts (phase summaries, review delegation).

**(b-note) Planner agent NOT required.** The orchestrator skill's LLM-composition of Summary + Test plan is a brief, template-driven operation from already-loaded task-file content — not a planning sub-task. This satisfies the stop condition: Decision 3's baseline body shape does NOT require the Planner or another agent to compose the Summary/Test-plan content.

**(c) Issue-mandated order is unambiguous and well-motivated.** `Manual QA Plan` first because it is the most action-oriented content for a reviewer who wants to start manual verification immediately. `Known Gaps` second because it contextualizes what the UAT plan does not cover. `Decisions` last because it is the longest section and the most reference-oriented — reviewers can scroll to it if they want rationale, rather than wading through it before reaching the actionable items.

**Trade-offs accepted:**

- The orchestrator skill gains a brief LLM-composition step (Summary + Test plan from task-file content) that did not exist before. This step is `--auto`-only and is template-driven from structured task-file fields; it is not open-ended generation.
- The PR body's Summary and Test plan sections are LLM-generated, not human-written. Under `--auto` this is the expected behavior; users who want a custom-worded PR body should compose it manually outside `--auto` mode (per the non-goal: non-`--auto` runs compose PRs manually today).
- `crafter pr-body` outputs only the three appended sections; it does not own the full PR body. The orchestrator skill is responsible for concatenation. This separation is consistent with Decision 1 (the CLI subcommand is plumbing; the skill is the contract) and with the GH#16 precedent for the buffer subcommand.
- The three appended sections are emitted by the Go subcommand in the fixed order `Manual QA Plan` → `Known Gaps` → `Decisions`. The orchestrator cannot reorder them without a new CLI flag; this is intentional — the order is a locked contract, not a runtime option.

### Decision 4 — UAT and Gaps rendering templates (Orchestrator Accepted) — 2026-05-10

**Chosen sub-choices:**

**(A) UAT title presentation:** `- [ ] **<title>**` — **bold title**. Bold makes the headline scannable when reviewers skim a long `## Manual QA Plan` section containing many entries. Plain is harder to parse at a glance. Issue acceptance criterion 5 specifies `- [ ] **Title** ...` explicitly; bold is the issue-mandated form, not just polish.

**(B) UAT verify content placement rule:** **threshold-based — single-line vs. multi-line determined by presence of `\n` in the stored `verify` value.**

- If `verify` contains no newline characters (single-sentence or short phrase): render inline after an em-dash — `- [ ] **<title>** — <verify>`. All content stays on one bullet line; the checkbox is unconditionally actionable.
- If `verify` contains one or more `\n` characters (multi-paragraph steps, embedded fenced code blocks, numbered lists, etc.): render `why_manual` on a continuation paragraph. Template:

  ```
  - [ ] **<title>**

    <verify — with \n rendered as literal newlines, each continuation line indented 2 spaces>

    _Why manual:_ <why_manual>
  ```

  The 2-space indent is the GFM list continuation indent that keeps all nested paragraphs inside the same bullet list item. GitHub renders this as one task-list entry with body content; the checkbox remains actionable because the GFM task list spec ties checkbox state to the `- [ ]` prefix of the bullet item, not to whether it has continuation content.

**(C) Source/why_manual handling in PR body:**

- **`source`** — **omitted from the PR body entirely.** `source` (e.g., `auth/callback.go:42`, `phase:3/step:1`) is useful for internal buffer navigation and debugging but adds noise for a PR reviewer whose job is to perform the verification step, not trace it back to a source location. The buffer NDJSON retains `source` for any tooling that needs it.
- **`why_manual`** — **included in multi-line UAT entries only** (as `_Why manual:_ <why_manual>` continuation line). It is omitted from single-line (inline) entries because `why_manual` is typically longer than `verify` in single-line cases and would make the inline form unwieldy. For multi-line entries the nested continuation paragraph already allocates vertical space, and explaining why the item cannot be automated is useful context for reviewers choosing which items to prioritize. `why_manual` is always omitted from Gap entries (Gaps have no `why_manual` field).
- **Gap `source`** — **omitted from the PR body**, same rationale as UAT `source`.

**Literal templates:**

**(A) UAT short-verify template** (no `\n` in `verify`):

```
- [ ] **<title>** — <verify>
```

**(B) UAT multi-line-verify template** (`\n` present in `verify`):

```
- [ ] **<title>**
  
  <verify, with each embedded \n replaced by a real newline followed by 2-space indent>
  
  _Why manual:_ <why_manual>
```

Note on the 2-space indent: the bullet prefix `- ` is 2 characters. GFM list continuation requires the continuation paragraph to be indented by at least the content column of the list item (2 spaces). All continuation lines — including blank separator lines between paragraphs — must use this indent. Blank separator lines between paragraphs MUST be exactly two spaces and nothing else (not a fully empty line), to keep the list-item context. The renderer must emit `  ` (2 spaces) on every blank separator line and `  ` (2-space prefix) before every non-blank continuation line.

**(C) Gap template** (always bullet with inline detail + nested follow-up):

```
- **<title>** — <detail, with \n rendered as newline + 2-space indent for continuation>

  _Follow-up:_ <followup>
```

If `followup` is empty or missing, the `_Follow-up:_` line is omitted entirely (same empty-source omission rule applied at field level).

**Rendered examples:**

**(1) Single-line UAT entry** — inline form:

```markdown
- [ ] **Confirm OAuth callback redirect** — Click Sign In; confirm browser lands on /dashboard not /login.
```

**(2) Multi-line UAT entry** — nested form (fenced code block in `verify`):

```markdown
- [ ] **Verify hover preview renders with empty dataset**
  
  Run the app with an empty dataset fixture:
  
  ` ` `bash
  DATA=fixtures/empty.json npm run dev
  ` ` `
  
  Hover over any row in the table and confirm the preview panel shows
  the placeholder message rather than a blank white box.
  
  _Why manual:_ Requires a browser and a running dev server; no headless test covers empty-dataset hover state.
```

**(3) Gap entry** — bullet with detail and follow-up:

```markdown
- **Database migrations lack a rollback path** — Migration 0042 adds an audit_log table and a trigger. The up migration is clean, but there is no corresponding down migration.

  _Follow-up:_ Add a down migration for 0042 that drops the trigger before dropping the table. Consider making the CI pipeline enforce that every up migration has a matching down migration.
```

**GFM checkbox actionability check:**

The multi-line UAT template uses the standard GFM list continuation pattern: `- [ ] **title**\n\n  body`. GitHub's task list implementation marks checkboxes as actionable based on the `- [ ]` prefix of the bullet item. Continuation paragraphs with 2-space indent are part of the same list item per the CommonMark spec (§5.3 lists), and GitHub's GFM rendering follows this rule. The checkbox is therefore clickable regardless of how much continuation content follows. This was verified mentally against the CommonMark spec and matches the GH#16 task file's own `### Decision (...)` H3 list patterns. Empirical confirmation is deferred to Phase 2 Step 4 fixture validation, which the plan's risk section already identifies as the checkpoint.

**Alternatives considered:**

- **(B-alt) Always inline with embedded `<br>` for newlines.** Rejected: fenced code blocks cannot be embedded in a single Markdown line even with `<br>` replacement — the backtick fence syntax requires a block context. A `verify` field containing a fenced code block (confirmed in GH#16 SKILL.md examples) would produce garbled output. The threshold-based rule handles this cleanly without any special-casing.
- **(B-alt2) Always nested (even for single-line verify).** Rejected: single-line `verify` entries are common (simple click-and-confirm steps); forcing them into a multi-paragraph form wastes vertical space in the PR body and obscures the signal-to-noise ratio for reviewers. Inline form for short verify is strictly more readable.
- **(C-alt) Include `source` in the PR body as italic prefix.** Rejected: `source` values like `auth/callback.go:42` or `phase:3/step:1` are internal navigation coordinates for buffer tooling and implementers, not actionable information for a PR reviewer. Including them adds visual clutter to the `## Manual QA Plan` and `## Known Gaps` sections without any reviewer benefit.
- **(C-alt2) Omit `why_manual` entirely.** Rejected: for multi-line UAT entries, `why_manual` directly answers the question a reviewer might ask — "why can't the CI suite verify this?" — which affects how they prioritize manual checking time. The value is brief and contextual; the vertical cost is one line per multi-line entry.

**Trade-offs accepted:**

- The threshold rule (inline vs. nested based on `\n` presence) means the renderer must inspect the `verify` string rather than applying a single unconditional template. This is a trivial string check in Go (`strings.Contains(entry.Verify, "\n")`) and adds no meaningful complexity.
- Multi-line UAT entries with fenced code blocks require the renderer to re-indent continuation lines (prefix each line with 2 spaces). The renderer must split on `\n` and re-join with `\n  ` — a standard Go `strings.Split`/`strings.Join` operation.
- Gap entries always use the bullet form (not a checkbox). This is correct per the issue: only UAT entries are task-list checkboxes (acceptance criterion 5); Gaps are informational, not items reviewers check off.
- `why_manual` is emitted only in multi-line entries (not in single-line). The Phase 2 renderer must implement this asymmetry explicitly. The Phase 2 Step 4 fixture matrix must cover both cases.

### Phase 1 review — auto-fixed and tech-debt findings (Auto Mode)

The Phase 1 review surfaced 2 Major findings and 8 Minor/Suggestion findings against Decision 4. The Major findings were resolved by fix-loop iteration 1; the re-review (iteration 1) closed both Majors with no new Critical/Major issues introduced. Recording the outcome below per `--auto` Step 6b.

- **Decision (Auto-Fixed): Major — Tautological prose at line 373 (Decision 4 multi-line UAT continuation indent).** Original wording described two visually-identical strings (`  ` vs. `  `) without distinguishing their roles. Rewrote line 373 to state concretely: blank separator lines between paragraphs MUST be exactly two spaces and nothing else (not a fully empty line); non-blank continuation lines carry a 2-space prefix before content. **Reason:** Phase 2 implementer needs an unambiguous target for the renderer.
- **Decision (Auto-Fixed): Major — Literal multi-line UAT template and prose disagreed on blank-separator-line shape.** The literal template (lines 365–371 area) and rendered example #2 (lines 396–408 area) showed fully empty separator lines while the prose said blanks must be 2-space-prefixed. Adopted **Resolution (A) — uniform 2-space-prefixed blanks throughout** the multi-line UAT template, applied to both the literal template and the rendered example. The Gap template's blank separator (line ~379) was deliberately left untouched as out-of-scope (not a Major finding); flagged as carry-forward Minor. **Reason:** internal consistency between the template contract and its illustrative example is required before Phase 2 starts implementing against the template.
- **Decision (Tech Debt — auto-recorded): Minor — Decision 4 lacks an explicit cross-reference to the buffer schema in `skills/crafter-buffer/SKILL.md`.** Phase 2 implementer can read the buffer SKILL.md directly; this is documentation polish.
- **Decision (Tech Debt — auto-recorded): Minor — Decision 4 calls bold UAT title "issue-mandated" which slightly overstates the source.** The issue example shows bold (`- [ ] **Title** ...`) but is illustrative; "we adopt bold to match the issue example" is the more accurate provenance. Local prose tweak.
- **Decision (Tech Debt — auto-recorded): Minor — Heading-format style (`(Orchestrator Accepted)` vs. GH#16's `(User Accepted)`) was flagged for confirmation.** The choice is intentional under `--auto` per the plan's Assumptions section (line 220 area). No change required; flag is informational.
- **Decision (Tech Debt — auto-recorded): Minor — Rendered example (2) inside Decision 4 uses spaced backticks (` `` ` ` ``  `) as a display workaround to avoid breaking the outer fenced block.** A one-line note inside Decision 4 would help future readers; the renderer itself emits real triple backticks (this was confirmed by the implementer's own note).
- **Decision (Tech Debt — auto-recorded): Minor — Decision 4 prose at line 373 (post-fix) still describes two visually-identical `  ` strings, distinguished by role rather than appearance.** Resolvable by switching to "two-space-only line" vs. "two-space indent before content" phrasing. Non-blocking; readability polish for a future pass.
- **Decision (Tech Debt — auto-recorded): Suggestion — Decision 3(b) does not specify what happens when `## Outcome` is empty (the common state during active development).** Default reading: Summary composes from `## Plan` alone in that case; Test plan derives from acceptance/verification criteria. Phase 3 Step 1 will need to lock this when wiring the orchestrator-side composition; recording here so it is not forgotten.
- **Decision (Tech Debt — auto-recorded): Suggestion — Asymmetry: Gap `followup` empty-omission rule is stated in Decision 4(C); UAT `why_manual` empty-omission rule for multi-line is not.** By analogy, an empty/missing `why_manual` should cause the `_Why manual:_` continuation line to be omitted. Phase 2 Step 1 should implement this consistently and Phase 2 Step 4 should cover it in the fixture matrix.
- **Decision (Tech Debt — auto-recorded): Suggestion — Add explicit cross-reference from Decision 3's section-ordering trade-off to Phase 2 drift criterion (c).** Visibility-only improvement.
- **Decision (Tech Debt — auto-recorded): Suggestion — Add a Phase 2 stop-condition for "GitHub renders the multi-line task list with checkbox not actionable."** The plan's Risks section already gestures at this (line ~227); making it a formal stop-condition would tighten the contract. Phase 2 Step 4 fixture validation is the empirical check; if it fails, Decision 4 must be revised before proceeding.
- **Decision (Tech Debt — auto-recorded): Suggestion — Gap template (line ~379) still has a fully empty blank separator line while the multi-line UAT template now uses 2-space-prefixed blanks (post-fix).** Carry-forward from iteration 1 review. Phase 2 fixture validation should confirm both render correctly under GFM; if the fully-empty line causes the Gap follow-up to escape the list-item context, the Gap template will need the same 2-space treatment.

### Phase 2 decisions

- **Decision (Orchestrator Accepted): Multi-line Gap `detail` lines are 2-space-indented via `indentBlock` for GFM list-item correctness; Gap blank separator before `_Follow-up:_` remains fully empty per locked template.** Decision 4's literal Gap template specifies the structure `- **<title>** — <detail>\n\n  _Follow-up:_ <followup>` but is silent on multi-line `detail` indentation. The implementer applied the same `indentBlock` helper used for multi-line UAT `verify` to multi-line Gap `detail` content for GFM correctness (continuation lines stay inside the bullet's list-item context). The blank separator between `detail` and `_Follow-up:_` is preserved as fully empty (`\n\n`), exactly per Decision 4's locked Gap template — only the *content* of multi-line detail gets the 2-space prefix. **Reason:** local, beneficial, does not alter the locked Gap separator, no impact on later steps. Recorded during GH#17 Phase 2 Step 1 drift check. (Refined in Phase 2 review fix-loop iteration 1: only `lines[1:]` are indented; the first line stays inline after `— ` per the deviation's literal description.)

### Phase 2 review — auto-fixed and tech-debt findings (Auto Mode)

The Phase 2 review surfaced 4 Major findings on the rendering implementation. Fix-loop iteration 1 cleared all four; the re-review (iteration 1) introduced no new Critical/Major findings. 15 Minor/Suggestion findings recorded below as tech debt.

- **Decision (Auto-Fixed): Major — `renderGapEntry` over-indented the first line of multi-line Gap `detail`.** The original implementation called `indentBlock(e.Detail)` and concatenated the result inline, producing `- **<title>** —   <first line>` (3 spaces between em-dash and content) and a stray prefix on the first line. Restructured to emit the first line inline after `— ` and indent only `lines[1:]` with `\n  ` (or `\n  ` for blank separators inside multi-paragraph detail). **Reason:** matches the literal Phase 2 Decision (Orchestrator Accepted) description ("continuation lines stay inside the bullet's list-item context") and produces correct GFM list-item rendering.
- **Decision (Auto-Fixed): Major — Multi-line Gap `detail` had zero test coverage.** Added `TestRenderGaps_MultiLineDetail` (single-newline → list continuation) and `TestRenderGaps_MultiParagraphDetail` (blank-line-separated paragraphs with 2-space-prefixed blank separators) with exact-output assertions. **Reason:** the accepted-deviation indent contract was unverified before; tests now lock it.
- **Decision (Auto-Fixed): Major — `bufio.Scanner` 64 KB token-size limit on buffer files (`readDataLines`).** Raised the scanner buffer cap to 1 MB via `scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)`; `bufio.ErrTooLong` is now wrapped with the file path and the limit so the operator can recover. **Reason:** corrupted or future-format buffer entries with >64 KB lines should fail with a clear error, not an opaque scanner error.
- **Decision (Auto-Fixed): Major — `bufio.Scanner` 64 KB token-size limit on task-file Decisions extraction (`ExtractDecisions`).** Same fix as above applied to `decisions.go`. **Reason:** task files have no per-line size cap; long fenced blocks or one-line tables would have triggered `bufio.ErrTooLong`. Now bounded at 1 MB with a clear error message.
- **Decision (Tech Debt — auto-recorded): Minor — `renderGapEntry` blank-line branch (`strings.TrimSpace(l) == ""`) is dead-equivalent to the else branch for canonical empty lines.** Behavior is fine; the conditional intent could be tightened or commented.
- **Decision (Tech Debt — auto-recorded): Minor — `renderGapEntry` no longer mirrors `renderUATEntry`'s shape (UAT uses `indentBlock`, Gap inlines its own loop post-fix).** Behavioral divergence is justified; structural divergence is incidental. An `indentContinuation(lines[1:])` helper could share the rule.
- **Decision (Tech Debt — auto-recorded): Minor — `maxScannerBytes` 1 MB cap duplicated across `format.go` and `decisions.go`.** Pull into a single package-level constant.
- **Decision (Tech Debt — auto-recorded): Minor — Test coverage gap: detail ending with `\n` (trailing newline) and multi-line detail combined with non-empty `followup` are not exact-asserted.**
- **Decision (Tech Debt — auto-recorded): Minor — `gofmt -l` flags `cli/internal/prbody/decisions.go`, `format.go`, `assemble_test.go` (var-block / doc-comment / struct-literal alignment).** Mechanical cleanup before next commit.
- **Decision (Tech Debt — auto-recorded): Minor — Malformed JSON data lines silently skipped in `readDataLines`.** A corrupt buffer line is invisible to the caller; consider counting and surfacing skips (warning to stderr or returned warning struct).
- **Decision (Tech Debt — auto-recorded): Minor — `isFenceDelimiter` matches indented and non-canonical fences too liberally.** Acceptable for current task files; flag as known simplification.
- **Decision (Tech Debt — auto-recorded): Suggestion — Add inline doc-comment at UAT multi-line builder explaining why `"  \n"` (trailing-space) keeps the list-item context per CommonMark §5.2.**
- **Decision (Tech Debt — auto-recorded): Suggestion — `cli/cmd/pr_body.go` `runPRBody` does not set `Args: cobra.NoArgs`.** Extra positional args silently succeed today.
- **Decision (Tech Debt — auto-recorded): Suggestion — `ExtractDecisions` boundary check is exact `line == "## Decisions"`.** Trailing whitespace or `## Decisions (Accepted)` would be silently ignored. Document the strictness in the doc-comment.
- **Decision (Tech Debt — auto-recorded): Suggestion — `Assemble`'s ordered if-block could be data-driven via `[]struct{ heading, body string }`** for easier future-section addition.
- **Decision (Tech Debt — auto-recorded): Suggestion — `ExtractDecisions` mixes I/O and the section state-machine.** Refactoring the scanner loop into a pure `extractDecisionsFromLines([]string) string` would improve unit-testability.
- **Decision (Tech Debt — auto-recorded): Suggestion — 1 MB scanner cap is a hard error rather than a graceful degrade (truncate-and-warn).** Acceptable for `--auto` (fail-fast); document somewhere.
- **Decision (Tech Debt — auto-recorded): Suggestion — Surface the offending line number in malformed-line warnings (when implemented per Minor #6/#10).**
- **Decision (Tech Debt — auto-recorded): Suggestion — Doc-comment on `UATEntry` should make explicit it is not a complete schema mirror** (only PR-relevant fields decoded; full buffer-round-trip needs the buffer package's struct).

### Phase 3 decisions

- **Decision (Orchestrator Accepted): Step 9b (PR Composition) is placed after Steps 7–9 (not between Steps 6b and 7–9) in `skills/crafter-do/SKILL.md`.** Rationale: the consolidated end-of-task commit, STATE.md update, and task-file `## Outcome` section must all be complete before `gh pr create` opens the PR and pushes the branch — this ensures the PR body reflects the final task state and the green-commit invariant is preserved (the consolidated commit is in history before the PR is opened). The run-dir is not needed by Steps 7–9, so deferring cleanup until after Step 9b is safe. **Reason:** ordering is non-obvious and worth recording so future maintainers do not "fix" the placement.
- **Decision (Orchestrator Accepted): No per-run metadata artifact needed for GH#17 — confirmed by Phase 2 implementation.** The PR composer (`cli/internal/prbody/`) reads only the two NDJSON buffer files (`uat-buffer.jsonl`, `gaps-buffer.jsonl`) and the task file's `## Decisions` section; it does not read or require any `meta.json` or equivalent run-marker file. The `rules/do-workflow.md` hedge ("can be revisited in GH#17 or GH#18 if the PR composer finds it necessary") has been removed and replaced with definitive past-tense policy. No metadata artifact will be introduced unless a future change demonstrates a concrete need.

### Phase 3 review — auto-fixed and tech-debt findings (Auto Mode)

The Phase 3 review surfaced 4 Major findings on Step 9b's wiring and the docs sweep. Fix-loop iteration 1 cleared all four; the re-review (iteration 1) introduced no new Critical/Major findings. 11 Minor/Suggestion findings recorded below as tech debt (some carry-forward from iteration 0, some surfaced by the iteration-1 fixes).

- **Decision (Auto-Fixed): Major — `gh pr create --title "<title>"` had a title-quoting hazard.** Commit subjects routinely contain `"`, `` ` ``, `$`, `!`. Step 9b now holds the title in a Bash variable (`TITLE=$(git log -1 --format='%s')`) and passes `"$TITLE"` to `gh pr create`. The Bash-no-word-split-on-RHS-of-assignment explanation is included; `gh ≥ 2.40`'s `--title-file -` is mentioned as alternative. **Reason:** prevents shell-injection / breakage on realistic commit subjects.
- **Decision (Auto-Fixed): Major — Step 9b's trigger precondition was wrong on the no-consolidated-commit path.** Steps 7–9 may produce no follow-up commit if no docs/skillbook/STATE.md updates are needed. Trigger now reads "may or may not have produced a consolidated end-of-task commit; the latest commit on the work branch will be either the consolidated commit OR the final per-phase commit, depending on whether updates were needed." **Reason:** correctly handles the legitimate path where Steps 7–9 are pure-housekeeping no-ops.
- **Decision (Auto-Fixed): Major — STATE.md TBD-backfill convention was undocumented.** New `## STATE.md commit-hash backfill` section added to `rules/post-change.md` codifying the convention with GH#16 (`5e8ceba`) precedent reference and a guard against introducing a different placeholder convention. **Reason:** prevents `TBD` from shipping in main if the backfill chore commit is forgotten; locks the convention before drift can creep in.
- **Decision (Auto-Fixed): Major — Step 9b's baseline-body specification was under-specified.** Tightened: explicit `\n` markers in the body skeleton; Test plan source pinned to issue ACs ONLY (extracted from task file's `## Request` section, NOT from phase verification criteria); Outcome-empty fallback path replaced with Ad-hoc escape hatch stop condition (Steps 7–9 must fill `## Outcome` before Step 9b runs). **Reason:** makes the orchestrator's body composition deterministic.
- **Decision (Tech Debt — auto-recorded): Minor — Step 9b's body-quoting example uses `printf '%s' "<full-body>"` which has the same shell-interpolation hazard as the title** (backticks, `$`, `!` in body content). Consider showing `--body-file <path>` or a heredoc with `<<'EOF'`.
- **Decision (Tech Debt — auto-recorded): Minor — `gh ≥ 2.40 --title-file -` mention is mutually exclusive with `--body-file -`** (single stdin stream). The mutual exclusion should be flagged inline.
- **Decision (Tech Debt — auto-recorded): Minor — Step 9b body skeleton shows literal `\n` tokens inside a fenced block.** A skimming reader might copy literal `\n` strings. Split into two blocks or move `\n` markers to inline prose.
- **Decision (Tech Debt — auto-recorded): Minor — Test plan AC source has two ordered fallbacks but no behavior specified when neither has bullet-list ACs.** Should empty Test plan also route through Ad-hoc escape hatch?
- **Decision (Tech Debt — auto-recorded): Minor — `rules/post-change.md` STATE.md backfill section's parenthetical for non-`--auto` path does not say *when* the backfill commit should happen.** Either narrow to `--auto` only or describe the non-`--auto` user-facing prompt.
- **Decision (Tech Debt — auto-recorded): Minor — STATE.md backfill section placement could read as if backfill must run before Session Wrap-Up.** A one-liner cross-reference to Step 9b would clarify ordering.
- **Decision (Tech Debt — auto-recorded): Suggestion — `gh pr create` push exception is documented in Step 9b but not in `rules/post-change.md` near the "Do not push to remote" rule.** A one-line back-reference would prevent confusion.
- **Decision (Tech Debt — auto-recorded): Suggestion — `Decision (Auto-Recorded): PR creation failed — <error>` template does not specify stderr capture rules** (verbatim, summarized, truncated). Specify a length cap or "first line of stderr" rule.
- **Decision (Tech Debt — auto-recorded): Suggestion — Step 9b's "If deletion fails, record a tech-debt note and continue" — destination unspecified.** Pick the task file's `## Decisions` section.
- **Decision (Tech Debt — auto-recorded): Suggestion — "Latest commit on work branch will be either consolidated OR final per-phase" is true only if no other commits land between Step 9 and Step 9b.** State the invariant explicitly.
- **Decision (Tech Debt — auto-recorded): Suggestion — `rules/post-change.md` backfill section references "GH#16 (`5e8ceba`)" without a one-clause gloss describing what the precedent commit did.**
- **Decision (Tech Debt — auto-recorded): Suggestion — Acceptance-criteria checkboxes at top of task file are still unchecked; an explicit per-AC mapping block would make the seven-to-N audit one-click.**

## Outcome
