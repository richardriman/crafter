# Task: Statusline fallback cascade — never silently empty when there is state to report

## Metadata
- **Date:** 2026-06-02
- **Work branch:** feat/plan-progress-statusline
- **Status:** completed
- **Scope:** Medium

## Request

Extend the existing `crafter statusline` Go subcommand (`cli/cmd/statusline.go` + `cli/internal/statusline/`) with a fallback cascade so the segment is never silently empty when there is useful state to report.

Current behavior: `Render(workdir)` resolves the active task whose `## Metadata` has BOTH `- **Status:** active` AND `- **Work branch:** <current git branch>`, parses its plan, and prints `crafter · Phase X/Y · a/b [bar] %`. On any miss it prints nothing.

Desired priority cascade:
1. **Active task for the CURRENT branch** → existing plan-position segment (unchanged).
2. **Else, completed task on the current branch** (`Status: completed`, `Work branch` = current branch) and no active task → show a "done on this branch" segment, e.g. `crafter · ✓ done`. Persists for as long as you are on that branch (no time-based expiry). If multiple completed tasks match, use the most-recent by filename.
3. **Else, active tasks on OTHER branches** → show a COUNT only, e.g. `crafter · 2 active elsewhere`. Count task files with `- **Status:** active` whose `Work branch` != current branch.
4. **Else** → empty output (current silent-fail posture preserved).

### Decisions already settled with the user (DISCUSS)

- **Branch strategy:** Work continues ON `feat/plan-progress-statusline` and lands in the already-open, unmerged **PR #42**. The statusline package exists only on this branch (not on `main`), so a fresh branch off `main` is not viable — there would be nothing to extend. This is a natural continuation of the not-yet-shipped statusline feature, so it folds into the same PR rather than a stacked/separate one.
- **Step 3 format = COUNT ONLY** (chosen over listing branch names) to keep the segment short, since the bar is shared with `gsd-statusline`.
- **"done on this branch" persists** as long as you are on the branch (chosen over last-completed-only).
- **Draft-plan tasks are included**, not dropped. An active task counts toward step 3 / is handled by step 1 regardless of whether its plan is still `draft` or already `approved`. (The existing parser already distinguishes draft vs approved and renders a planning/awaiting-approval edge state for step 1; that stays.)
- **Out of scope — the `- **Branch:**` alias.** The resolver matches only `- **Work branch:** `. One real task file (`20260421-skills-first-runtime-portability.md`) uses the non-standard `- **Branch:** main`. The user explicitly declined to address this. The resolver stays strict to the single `templates/TASK.md` field; that one task simply will not be counted by step 3. No alias, no edit to that task file. Documented as a known limitation, not a bug to fix here.
- **Contract preserved:** always exit 0; empty output on any problem/error (unchanged silent-fail posture). Segment strings stay English, consistent with the existing `crafter · Phase X/Y` segment.

### Scope

- Go code: `cli/internal/statusline/` (resolver + `Render` cascade) and any supporting parse logic.
- Table-driven tests in `cli/internal/statusline/statusline_test.go` — extend the existing suite (which already covers active-task rendering) for the new cascade branches.
- Docs that already describe the statusline: `README`, `ARCHITECTURE.md`, `PROJECT.md` — update for the new fallback behavior.

## Plan
**Plan status:** approved

### Why this change matters

The `crafter statusline` segment is the only ambient, always-on signal a developer gets about Crafter state while working in Claude Code. Today it speaks up only in one situation: there is an active task whose `## Metadata` lists both `- **Status:** active` and `- **Work branch:** <current branch>`. Every other situation — you just finished a task and are still on its branch, or you are on a branch with no task but have active work elsewhere — produces silence. Silence is indistinguishable from "Crafter isn't installed / isn't watching", which under-sells the tool and hides useful state. This task adds a priority cascade so the segment reports the *best available* truth instead of going dark, while never breaking the status bar (still always exit 0, still empty on any genuine problem).

This is an **extension of a working, unshipped feature** on its own branch — not a rewrite. Step 1 (the active-current-branch plan-position segment) must remain byte-for-byte identical in output. We are bolting two new fallback rungs and an empty terminal onto the front-door resolver.

### Complete request (restated)

Implement the four-rung priority cascade inside the statusline package so `Render(workdir)` returns:

1. **Active task for the current branch** → the existing plan-position segment, unchanged (`crafter · Phase X/Y · a/b [bar] %`, plus the `crafter · planning` / `crafter · plan: awaiting approval` edge states for draft/no-plan).
2. **Else, a completed task on the current branch** (status completed/done, work-branch = current branch, and no active task on this branch) → a short "done on this branch" segment. Persists as long as you stay on the branch; no time expiry. On multiple completed matches, most-recent-by-filename wins.
3. **Else, active tasks on other branches** → a **count-only** segment, e.g. `crafter · 2 active elsewhere`. Counts task files with `- **Status:** active` whose work-branch differs from the current branch.
4. **Else** → empty output.

**Acceptance criteria:**
- Existing test suite stays green with zero changes to existing cases; step-1 output unchanged.
- New table-driven cases cover rungs 2, 3, 4 and the edge cases enumerated below.
- A single directory scan classifies every task file once (no triple scan of the tasks dir).
- The contract holds: always exit 0 (command layer is untouched and already guarantees this), empty output on any error, English segment strings consistent with the `crafter · …` prefix and `·` separator.
- Docs describing the statusline are updated: `README.md` (Statusline section), `.crafter/ARCHITECTURE.md` (the two statusline mentions), `.crafter/PROJECT.md` (Key Decisions, if a line is warranted).

**Constraints (settled, not open for re-litigation):**
- Work stays on `feat/plan-progress-statusline` (folds into PR #42). Do NOT branch off main — the package does not exist on main.
- Resolver stays strict to `- **Work branch:** ` only. The `- **Branch:**` alias / the `20260421-skills-first-runtime-portability.md` task file are OUT OF SCOPE. No alias, no edit to that file. That task simply isn't counted by rung 3 — a documented known limitation, not a bug.
- Draft-plan active tasks are first-class: an active task counts toward rung 3 / is handled by rung 1 regardless of draft vs approved plan state.

### Assumptions / interpretations

- **A1 — "completed" status spelling.** The request and the resolver doc both speak of "completed"/"done". The existing test `TestResolveActiveTask_StatusNotActive` uses `"done"` as a representative non-active status, and recent commits use messages like "complete … task". I assume rung 2 should treat the task as completed when its `## Metadata` status is **either `completed` or `done`** (both are observed in the wild). The Implementer should confirm which value `templates/TASK.md` / `rules/task-lifecycle.md` actually writes on closure and match that as the primary, treating the other as a tolerated synonym. If only one is real, matching just that one is acceptable — but matching both is safer and costs nothing. **Flagged, not silently chosen** — see Risks.
- **A2 — rung-2 segment string.** I propose `crafter · ✓ done` (matches the request's own example and reuses the `✓` check glyph the user wrote). No phase/step detail — the plan is finished, so a position bar would be noise.
- **A3 — rung-3 segment string.** I propose `crafter · N active elsewhere`, e.g. `crafter · 2 active elsewhere`, with `N` the integer count. Singular/plural is NOT specially handled (`1 active elsewhere` is acceptable and consistent); introducing pluralization logic would be scope creep. The Implementer may keep it uniform.
- **A4 — count semantics for rung 3.** "Active tasks on other branches" counts **distinct task files** that are active and whose work-branch != current branch — not distinct branches. The request says "Count task files with Status active whose Work branch != current branch", so file-count is authoritative.
- **A5 — classification is per-file, single pass.** A file is read once; we extract its status and work-branch from the `## Metadata` section (same bounded scan as today), then the caller buckets it. No file is opened more than once per `Render`.
- **A6 — rung precedence is strict and short-circuiting.** Rung 1 wins over rung 2 even when both an active and a completed task exist on the current branch. Rung 2 wins over rung 3. We never blend rungs.
- **A7 — the `Render` → `parsePlan` → `renderSegment` pipeline for rung 1 is untouched.** Only rungs 2/3/4 are new code paths. Rung 2 and rung 3 do NOT call `parsePlan` (no plan position is shown).

### Non-goals

- No change to the cobra command layer (`cli/cmd/statusline.go`) — it already guarantees exit 0 and prints only non-empty output.
- No change to `parsePlan` / `renderSegment` plan-rendering logic, glyphs, bar math, or gate detection.
- No `- **Branch:**` alias, no second accepted work-branch field, no edit to the skills-first task file.
- No time-based expiry, no persistence store, no caching — rung 2 is recomputed live from the branch + tasks dir on every invocation.
- No branch-name listing in rung 3 (count only, settled decision).
- No new config flags, no new subcommands, no installer changes.
- No pluralization / i18n of segment strings.

### Relevant areas to inspect

- `cli/internal/statusline/resolve.go` — the heart of the change. `resolveActiveTask`, `isActiveTaskForBranch`, `findCrafterDir`, `readGitBranch`, `findGitDir`. The single-pass classification lives here.
- `cli/internal/statusline/statusline.go` — `Render` orchestrates the cascade; this is where rung ordering and the new segment strings are wired (or where calls to new helpers land).
- `cli/internal/statusline/parse.go` — read-only reference; rung 1 still routes through `parsePlan`/`renderSegment`. Confirm the new segment strings do NOT collide with existing ones.
- `cli/internal/statusline/statusline_test.go` — extend the existing table-driven suite and reuse `makeRepo`, `makeTaskFile`, `writeFile`, `Render` integration helpers. `makeTaskFile` already takes `(root, name, status, branch)` — it is ready for completed/other-branch fixtures as-is.
- `README.md` (Statusline section, lines ~71–94), `.crafter/ARCHITECTURE.md` (lines ~33 and ~129), `.crafter/PROJECT.md` (Key Decisions table). Docs update.

### Resolver refactor — design decision and alternatives

The core design question is how one directory scan yields all three signals. Decision and reasoning:

**Chosen: one pass that classifies each file once, returning a small result struct.** Introduce an internal scan that walks `<ctx>/.crafter/tasks/*.md` exactly once. For each `.md` file, extract `(status, workBranch)` from the `## Metadata` section using the *same bounded scan* that exists today. The caller then buckets each file: active+current-branch, completed+current-branch, or active+other-branch. From the buckets we derive (a) the rung-1 winner (lexicographically largest filename among active-current), (b) the rung-2 winner (lexicographically largest among completed-current), and (c) the count of active-other-branch files. `resolveActiveTask` is reframed as "scan + classify" returning this result; `Render` reads it and applies the cascade. This keeps a single `os.ReadDir` and a single open-per-file, preserves the silent-fail posture (any setup failure → empty/zero result → `Render` returns ""), and keeps step 1's selection logic identical (same filter, same tie-break).

**On `isActiveTaskForBranch`:** generalize it into a metadata extractor that returns the file's `(status, workBranch)` rather than a bool, and let the caller decide. The current bool predicate is a special case (`status=="active" && workBranch==branch`). This is the simplest surgical move: one function changes shape, the branch-equality and status checks move up one level into the classifier where all three rungs need them. Adding three sibling predicates (`isActiveForBranch`, `isCompletedForBranch`, `isActiveOtherBranch`) would re-open and re-scan the same metadata three times per file and duplicate the section-boundary parsing — rejected for that reason.

**Alternative A — keep `isActiveTaskForBranch` as-is, add two more predicates, three passes.** Rejected: triples the file I/O and duplicates the `## Metadata` boundary logic three ways; drift-prone and against the "one pass" requirement.

**Alternative B — extract a richer parsed-task struct and return a slice of all tasks from the scan, do all bucketing in `Render`.** Viable and clean, but pushes resolution policy into `statusline.go`. Acceptable if the Implementer prefers it, *provided* the single-scan / single-open invariant holds and step-1 selection stays identical. Listed as an allowed implementation variant, not the default.

**Alternative C — shell out to `git`/`grep`.** Rejected: the package deliberately reads `.git/HEAD` and files directly for zero-dependency, hermetic, testable behavior. No new external process.

The Implementer chooses concrete names and struct shape; the binding constraints are: one `ReadDir`, one open per file, identical rung-1 selection and tie-break, silent-fail intact.

---

### Phase 1 — Cascade resolver + Render rungs 2/3/4 (with tests)

**Outcome:** `Render(workdir)` implements the full four-rung cascade. Rung 1 output is byte-identical to today. Rungs 2 and 3 emit the new English segments; rung 4 emits empty. A single tasks-dir scan classifies every file once. The new behavior is covered by extensions to the existing table-driven suite, and the whole package builds and tests green.

**Scope boundary:** `cli/internal/statusline/resolve.go`, `cli/internal/statusline/statusline.go`, and `cli/internal/statusline/statusline_test.go`. No other files.

**Non-goals for this phase:** no docs edits (Phase 2), no command-layer edits, no parse.go logic changes.

**Simplicity constraint:** reuse the existing `## Metadata` bounded-scan logic and the existing tie-break (lexicographic filename). Add the minimum surface: one reshaped metadata extractor + one classifier/result type + cascade wiring + new segment string constants. No new packages, no new external deps, no abstraction beyond what three rungs require.

- [x] **Step 1.1 — Reshape metadata extraction to return `(status, workBranch)` per file.** Generalize the per-file `## Metadata` scan (today's `isActiveTaskForBranch`) so it yields the file's status and work-branch instead of a hard-coded active+branch bool, keeping the section-boundary semantics (start at `## Metadata`, stop at next `##` or EOF) and the 1 MB scanner buffer. On any open/read error it must yield an empty/zero result that classifies as "no match" — silent-fail preserved.
  - *Drift criteria:* drifting if a second file-open is introduced per file; if the metadata scan reads past the `## Metadata` section; if any error path surfaces an error to the caller instead of degrading to empty; if the bounded-scan buffer size changes.
  - *Verification evidence:* package compiles; existing `TestResolveActiveTask_*` cases still pass unchanged (they exercise the same active+branch outcome through the new shape).
  - *Stop condition:* stop once the extractor returns status+branch and the existing resolver semantics route through it; do not add completed/other-branch logic here (that's the classifier).

- [x] **Step 1.2 — Single-pass classifier producing the three signals.** Scan `<ctx>/.crafter/tasks/*.md` once; classify each file via the Step 1.1 extractor into: active-on-current-branch, completed-on-current-branch, active-on-other-branch. Derive the rung-1 winner and rung-2 winner by the existing lexicographic-largest-filename tie-break, and the rung-3 count as the number of active-other-branch files. Preserve every guard: no `.crafter/` dir, no git / detached HEAD, missing/empty tasks dir → all yield an empty/zero result (no panic, no error). Draft-vs-approved is irrelevant to classification (an active task is active regardless of plan state).
  - *Drift criteria:* drifting if `os.ReadDir` runs more than once per `Render`; if any file is opened more than once; if a completed-current task can win over an active-current task; if a non-`.md` entry or directory is classified; if the `- **Branch:**`-only task gets counted (it must not — strict `- **Work branch:** ` match).
  - *Verification evidence:* new table-driven tests assert the classifier/result for: active-current present, completed-current present (no active), N active-other, mixed (active+completed current → active selected), and zero-everything → empty result.
  - *Stop condition:* stop once the three signals are correctly produced from one scan; do not format strings here (that's Render).

- [x] **Step 1.3 — Wire the cascade in `Render` and add rung-2/rung-3 segment strings.** `Render` consumes the classifier result and applies strict precedence: rung 1 (active-current → existing `parsePlan`/`renderSegment` path, unchanged) → rung 2 (completed-current, no active → `crafter · ✓ done`) → rung 3 (active-other count > 0 → `crafter · N active elsewhere`) → rung 4 (empty `""`). Introduce the two new segment strings as named constants. Rung 2 and rung 3 must NOT call `parsePlan`.
  - *Drift criteria:* drifting if rung 1's output changes by even one byte; if rung 3 ever prints `crafter · 0 active elsewhere` (zero must fall through to empty); if precedence is reordered; if a non-English string or a separator other than `·` is used; if `parsePlan` is invoked for rungs 2/3.
  - *Verification evidence:* `Render` integration tests over temp dirs produce exactly `crafter · ✓ done`, `crafter · 2 active elsewhere`, `""`, and the unchanged rung-1 segment for the matching fixtures.
  - *Stop condition:* stop once all four rungs are wired and short-circuit correctly; no extra rungs, no config.

- [x] **Step 1.4 — Extend the table-driven test suite for all new rungs and edge cases.** Add cases reusing `makeRepo` / `makeTaskFile` / `writeFile` / `Render`. Cover the full edge matrix:
  - completed-current, no active → `crafter · ✓ done`;
  - multiple completed-current → most-recent-by-filename wins (assert which);
  - active-current AND completed-current both present → rung 1 wins (active segment, not `✓ done`);
  - one / multiple active-other-branch → `crafter · 1 active elsewhere` / `crafter · N active elsewhere`;
  - zero active-other and nothing else → empty string (NOT `0 active elsewhere`);
  - no `.crafter/` dir → empty; detached HEAD / no git → empty; tasks dir missing/empty → empty;
  - a task using only `- **Branch:** main` (non-standard) coexisting with the current branch → it is NOT counted by rung 3 (assert the count excludes it — this pins the documented known limitation as intended behavior).
  - *Drift criteria:* drifting if any existing test case is modified or deleted; if a new case asserts behavior outside the four-rung contract; if the non-standard-field exclusion case is omitted (it is the proof of the settled scope decision).
  - *Verification evidence:* `go test ./cli/...` (or `mise exec -- go test ./...` from `cli/`) passes; new cases visibly cover each rung and each enumerated edge.
  - *Stop condition:* stop once the matrix above is covered and green; do not add speculative cases for behaviors outside the contract (e.g. pluralization, branch listing).

- [x] **Phase verification** — `mise exec -- go build ./...` and `mise exec -- go test ./...` (run from `cli/`) both pass; the full edge matrix above is covered; rung-1 output confirmed byte-identical (existing cases unchanged); single-scan / single-open invariant holds; silent-fail / exit-0 posture intact.
- [x] **Phase review** — crafter-reviewer confirms surgical diff (only the three statusline files touched), no scope creep, segment strings English and `·`-consistent, the known-limitation test present, and no regression to step 1.

---

### Phase 2 — Documentation of the cascade

**Outcome:** The docs that describe the statusline accurately reflect the four-rung cascade, including the rung-2 and rung-3 segment formats and the documented `- **Branch:**` known limitation. No behavior change.

**Scope boundary:** `README.md`, `.crafter/ARCHITECTURE.md`, `.crafter/PROJECT.md`. Docs only.

**Non-goals for this phase:** no code changes; no new doc files; no installer/wiring doc changes (the `--with-statusline` mechanics are unchanged).

**Simplicity constraint:** extend the existing statusline prose in place; do not restructure the docs or add new top-level sections. Match the existing tone and the existing example-block style.

- [x] **Step 2.1 — Update `README.md` Statusline section.** Document the fallback cascade: the existing position/edge-state examples stay; add that when no active task matches the current branch, the segment falls back to `crafter · ✓ done` (completed task on this branch, persistent while on the branch) and then to `crafter · N active elsewhere` (count of active tasks on other branches), and finally to empty. Keep the "never breaks the status bar / silent when nothing to report" guarantee accurate. Note the known limitation: tasks using a non-standard work-branch field are not counted.
  - *Drift criteria:* drifting if the documented strings differ from the implemented constants; if the `--with-statusline` install mechanics are altered; if the silent-fail guarantee wording is weakened or removed.
  - *Verification evidence:* README Statusline section lists all four rungs with correct example strings.
  - *Stop condition:* stop once the cascade is described; no broader README edits.

- [x] **Step 2.2 — Update `.crafter/ARCHITECTURE.md` statusline mentions.** Refresh the two references (the `statusline.go` tree comment ~line 33 and the subcommand bullet ~line 129) so they describe the cascade rather than only "active task's plan position". Keep the example and the "silent when not a Crafter project" clause; extend with the fallback rungs concisely.
  - *Drift criteria:* drifting if the architecture description contradicts the README or the code; if unrelated architecture sections are edited.
  - *Verification evidence:* both statusline mentions describe the four-rung behavior.
  - *Stop condition:* stop once both mentions are accurate.

- [x] **Step 2.3 — Update `.crafter/PROJECT.md` if warranted.** The Key Decisions table already has the 2026-06-01 statusline row. Add/adjust at most one row noting the fallback-cascade extension and the settled scope (strict `- **Work branch:** ` only; `- **Branch:**` alias declined) so the decision is traceable. Keep it to a single concise row; if the Implementer judges the existing row sufficient, leave it and record that judgment.
  - *Drift criteria:* drifting if more than one row is added or unrelated decisions are touched; if the row re-opens a settled decision.
  - *Verification evidence:* PROJECT.md Key Decisions reflects the cascade scope (or a recorded judgment that the existing row suffices).
  - *Stop condition:* stop once the decision is traceable in one row.

- [x] **Phase verification** — README, ARCHITECTURE.md, and PROJECT.md describe the four-rung cascade and the known limitation; all documented strings match the implemented constants from Phase 1; no behavior or install-mechanics wording changed.
- [x] **Phase review** — crafter-reviewer confirms docs are accurate, consistent with code, scoped to the statusline sections, and free of model attribution.

### Risks / unknowns / flags

- **R1 — completed-status spelling (A1).** The exact closure value (`completed` vs `done`) written by `templates/TASK.md` / `rules/task-lifecycle.md` must be confirmed by the Implementer. Matching both is the safe default and is what this plan assumes; if the project guarantees exactly one, matching that one is fine. This is the only genuine ambiguity. Flagged here rather than guessed silently.
- **R2 — segment-string bikeshed.** `crafter · ✓ done` and `crafter · N active elsewhere` are proposals matching the request's own examples. If the user wants different wording, that is a trivial constant change — surface it before merge if there's any doubt.
- **R3 — known-limitation is intentional.** The non-standard `- **Branch:**` task being uncounted is *by decision*, asserted in Step 1.4 and documented in Phase 2. If a reader later files it as a bug, the test + docs are the record that it was deliberate.
- **R4 — Go toolchain access.** `go` is not on PATH in this shell; use `mise exec -- go …` for build/test (per project memory).

## Decisions
- **Decision (User Accepted):** Continue work on `feat/plan-progress-statusline` (PR #42) instead of a fresh branch off `main`. **Reason:** The statusline package does not exist on `main` yet; the feature is unshipped, so the extension folds naturally into the same PR.
- **Decision (User Accepted):** Drop the `- **Branch:**` alias / skills-first task-file fix from scope. **Reason:** User declined; keeping the resolver strict to the single documented `- **Work branch:**` field avoids a second accepted format. The non-standard task is left as-is and is simply not counted by step 3 (known limitation).
- **Decision (User Accepted):** Phase 2 (docs) review found no Critical/Major, two Suggestions. The user chose to fix #1: README rung-3 now documents that the count is not pluralized (`crafter · 1 active elsewhere` for a single task). #2 (ARCHITECTURE.md tree comment not enumerating rungs) needed no action — appropriate brevity, the full enumeration lives in the subcommand bullet. **Reason:** Complete reader-facing documentation of the singular form.
- **Decision (User Accepted):** Phase 1 review found no Critical/Major. The user chose to fix all three optional findings: (#1 Minor) removed the production-dead `resolveActiveTask`/`taskMatch` so resolution flows through a single path (`classifyTasks`/`Render`); (#2 Suggestion) the stale doc comment was removed with the function; (#3 Suggestion) repointed the `TestResolveActiveTask_*` cases to `TestClassifyTasks_*` (no coverage loss) and added `TestClassifyTasks_CompletedTieBreak` asserting the rung-2 most-recent-by-filename selection on `CompletedCurrent`. A re-review surfaced one gofmt regression (trailing blank line in `resolve.go`), fixed via `gofmt -w`. **Reason:** Clean single-resolution-path end-state and gofmt-clean tree before commit.

## Outcome

Both phases delivered on `feat/plan-progress-statusline` (PR #42).

**Phase 1 — cascade resolver + Render rungs 2/3/4 (commit `42fb291`):** `cli/internal/statusline/` now implements the four-rung priority cascade. `extractTaskMeta(path) taskMeta{status, workBranch}` generalizes the per-file `## Metadata` bounded scan; `classifyTasks(ctxDir, branch)` does a single `os.ReadDir` + one open per file, classifying each task into active-current / completed-current / active-other and deriving the rung-1 winner, rung-2 winner, and rung-3 count (lexicographic-largest-filename tie-break; `workBranch != ""` excludes the non-standard `- **Branch:**` task). `Render` applies strict short-circuit precedence: rung 1 (active-current → unchanged `parsePlan`/`renderSegment` segment) → rung 2 (`crafter · ✓ done`) → rung 3 (`crafter · %d active elsewhere`) → rung 4 (`""`). The dead `resolveActiveTask`/`taskMatch` were removed (review #1/#2); their tests were repointed to `TestClassifyTasks_*` with no coverage loss, and `TestClassifyTasks_CompletedTieBreak` was added (review #3). Rung-1 output is byte-identical; always exit 0, empty on any error. Test suite extended with 11 cascade cases (incl. the known-limitation exclusion), all green; gofmt-clean.

**Phase 2 — documentation (commit `24727ed`):** README Statusline section rewritten to document all four rungs (incl. the no-pluralization `crafter · 1 active elsewhere` note), the silent-fail/exit-0 guarantee, and the known limitation; `.crafter/ARCHITECTURE.md` two statusline mentions refreshed to the four-rung cascade; one new `.crafter/PROJECT.md` Key Decisions row (2026-06-02). All documented strings verified against the code constants; `·` is U+00B7; no model attribution.

**Deviations:** `Render` consumes `classifyTasks` directly rather than going through `resolveActiveTask` (allowed plan Alternative B; avoids a double scan) — `resolveActiveTask` was subsequently deleted as dead code during review. Status matching settled on exactly `completed` (canonical per `templates/TASK.md` + `rules/task-lifecycle.md`), not the `done` synonym the plan's A1 mused about. The `- **Branch:**` alias / skills-first task file remain out of scope (user decision) — that one task is simply not counted by rung 3, pinned as intended behavior by `TestRender_KnownLimitation_NonStandardBranchField` and documented as a known limitation.

**Verification:** `mise exec -- go build ./...`, `mise exec -- go vet ./...`, and `mise exec -- go test ./...` (from `cli/`) all green; phase verifications 6/6 + 6/6 PASS; both phase reviews 0 Critical/0 Major.
