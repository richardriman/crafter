# Task: Plan-progress statusline (crafter statusline subcommand)

## Metadata
- **Date:** 2026-06-01
- **Work branch:** feat/plan-progress-statusline
- **Status:** completed
- **Scope:** Medium

## Request
> Po vzoru GSD, progress bar to Claude bottom baru, kde se nacházíme v rámci plánu. Je to řešeno prostřednictvím hooku asi.

(Translated) Following GSD's example, add a progress bar to Claude Code's bottom bar (statusline) showing where we are within the current plan. The user assumed it is solved via a hook.

Research established that the GSD reference uses Claude Code's native `statusLine` settings key (a command whose stdout becomes the bar), not a `hooks` entry. The feature for Crafter derives plan position from the active task file in `.crafter/tasks/` (the one whose `**Status:** active` and `**Work branch:**` match the current git branch), counting step checkboxes and identifying the current phase.

## Plan
**Plan status:** approved

### 1. Complete request

Add a new Go CLI subcommand `crafter statusline` that renders the current Crafter plan position as a single composable segment for Claude Code's native `statusLine` bar, GSD-style. The subcommand reads the Claude Code stdin JSON payload, resolves the active task file (the `.crafter/tasks/*.md` whose metadata has `**Status:** active` and `**Work branch:** <current git branch>`), parses the plan, and prints exactly one line such as:

```
crafter · Phase 2/3 · 7/12 [███████░░] 58%
```

The four `Decision (User Accepted)` entries in `## Decisions` are settled and govern this plan:
1. Implement as a Go subcommand (not a JS hook).
2. The subcommand prints only its own composable segment; `install.sh` wires it via an opt-in `--with-statusline` flag and never overwrites an existing `statusLine`.
3. Display granularity is phase + steps + bar + percent.
4. Edge states: render nothing when not a Crafter project or no active task; render a planning/awaiting-approval state when an active task exists but the plan is `_(pending)_` or `**Plan status:** draft`.

**Why this matters:** Crafter already tracks fine-grained plan position in task-file checkboxes, but that state is invisible during a session. Surfacing it in the always-on statusline gives the user continuous orientation ("where are we in the plan") at zero interaction cost, matching a pattern users already value from GSD — without coupling to GSD or forcing it as the sole statusline.

**Acceptance criteria:**
- `crafter statusline` reads stdin JSON, resolves the active task, and prints the segment format above on stdout; exit code is always 0 and output is empty on any problem (missing project, no active task, parse failure, no stdin).
- Empty/edge states behave exactly per Decision 4.
- Default `install.sh` (global and local) does NOT touch `statusLine`. `--with-statusline` sets `statusLine` only when absent; on collision it prints a ready-to-paste composite wrapper and does not modify the existing value.
- Go unit tests (table-driven) cover the parser and renderer including all edge states; `tests/test_install.sh` covers the opt-in wiring and the no-overwrite/collision behavior.
- Docs (README, and a note flagged for ARCHITECTURE/PROJECT) describe the new subcommand and opt-in flag.

**Validation strategy:** `go test ./...` in `cli/` for the subcommand; `tests/test_install.sh` for install wiring; manual smoke check piping a sample JSON payload into the built binary.

### 2. Assumptions / interpretations

- **A1 — Active-task detection mirrors resume logic.** The subcommand finds the active task by scanning `<project>/.crafter/tasks/*.md` for metadata lines `**Status:** active` and `**Work branch:** <branch>` matching the current git branch, mirroring `rules/task-lifecycle.md` resume detection. If multiple match, pick the most recent by date prefix (matches the documented tie-break). The legacy `.planning/` directory is treated as a fallback only if `.crafter/` is absent (consistent with the skillbook path policy in ARCHITECTURE.md); I recommend supporting `.crafter` first and treating `.planning` fallback as optional — flagged in Risks.
- **A2 — Git branch resolution.** The subcommand needs the current branch to match `Work branch`. Recommended: read `.git/HEAD` directly (resolve `ref: refs/heads/<branch>`) walking up from `current_dir`, rather than shelling out to `git`. Rationale: no external process per render (statuslines render frequently), no dependency on `git` being on PATH, and graceful empty-output on detached HEAD or no repo. The Implementer may choose `git rev-parse --abbrev-ref HEAD` instead if it proves simpler, but must justify and keep the silent-fail posture. (This is a design choice the Implementer owns within these constraints.)
- **A3 — Phase/step counting model (Decision 3 mechanics).** Among `### Phase`/`#### Phase` headings, total phases = count of phase headings. Current phase = the phase heading governing the first unchecked step. Step counts (done/total) count only genuine work-step checkboxes and EXCLUDE the per-phase gate lines (`Phase verification`, `Phase review`) and any end-of-task post-change checkboxes. Rationale: gates are workflow ceremony, not user-facing plan steps; including them would distort "how much of the plan is done." The Implementer must detect gate/post-change lines by their conventional wording (e.g. lines matching `Phase N verification`/`Phase N review`/`STATE.md`/`Task file completion`) and exclude them. This exclusion rule is a stated design choice, not a guess — see Risks for the heading-level (`###` vs `####`) ambiguity.
- **A4 — Percent basis.** Percent = round(done_steps / total_steps × 100) over the work-step count (gates excluded), and the 10-segment bar uses `floor(pct/10)` filled glyphs, matching GSD's `renderProgressBar`. Phase shown as `Phase <current>/<total>`.
- **A5 — Plan-state detection.** `## Plan` containing `_(pending)_` → no plan yet → planning state. `**Plan status:** draft` → awaiting-approval state. `**Plan status:** approved` → executing state with counts. These mirror `rules/do/step-0-resume.md`.
- **A6 — Color/glyph posture.** Use the block glyphs `█`/`░` for the bar (already used across the repo and GSD). Keep ANSI color minimal or omit it for the first cut; if color is used it must degrade gracefully (the segment must still be readable as plain text). The Implementer owns the exact styling within "tasteful, degrades gracefully."
- **A7 — Segment label.** The leading label is the literal `crafter` followed by ` · ` separators (middle-dot + spaces), matching the target string.

### 3. Non-goals

- NOT building a full statusline (model, dir, context-window segments) — Crafter prints only its own segment.
- NOT composing/merging multiple statuslines automatically — on collision the installer only PRINTS a ready-to-paste wrapper; it does not author or install a tee wrapper script.
- NOT adding any new config file or config schema beyond the `--with-statusline` flag.
- NOT modifying GSD or reading GSD state; the GSD file is reference-only.
- NOT touching the installed copy under `~/.claude/crafter/` — source files only.
- NOT changing default install behavior; statusline wiring is strictly opt-in.
- NOT live-watching files or caching; each invocation is a one-shot read.

### 4. Relevant areas

- `cli/cmd/root.go`, `cli/main.go` — cobra root and version wiring (registration pattern).
- `cli/cmd/pr_body.go`, `cli/cmd/buffer_uat.go` — subcommand idioms (`SilenceUsage`, `RunE`, print-only-when-nonempty, registering on root).
- `cli/internal/prbody/` — internal-package + table-driven test conventions (`decisions.go`/`decisions_test.go` for markdown-section parsing patterns, `assemble.go` for the print-empty-on-empty posture, live-fixture smoke tests).
- `cli/Makefile`, `cli/go.mod` — build targets and module path (`github.com/richardriman/crafter/cli`).
- `templates/TASK.md` — canonical task-file structure (Metadata, `## Plan`, `_(pending)_`, `**Plan status:**`).
- `.crafter/tasks/20260518-refactor-crafter-do-slice-5-steps-5-5a.md` and similar — real phased task files (note: real files use `#### Phase N` headings and bold gate lines `- [x] **Phase 1 verification.**`).
- `rules/task-lifecycle.md`, `rules/do/step-0-resume.md` — authoritative active-task detection and plan-state semantics.
- `install.sh` — `install_hook()` (the Node-based settings.json edit pattern to mirror), `install_to()` / `install_global()` / `install_local()`, argument parsing, binary path resolution (`$base/crafter/bin/crafter`, `~/.local/bin/crafter` symlink).
- `tests/test_install.sh` — install test harness (`_run_installer`, settings.json assertions, idempotency).
- `~/.claude/hooks/gsd-statusline.js` and `~/.claude/settings.json` — reference only for the `statusLine` key shape and stdin contract (do NOT couple).
- `README.md` (CLI Binary / Skills sections), `docs/` — documentation surfaces.

### 5. Vertical phases and steps

The work splits into three independently verifiable vertical phases. Phase 1 delivers a working, tested subcommand (the core value). Phase 2 makes it installable opt-in. Phase 3 documents it. Each phase leaves the system coherent.

#### Phase 1 — `crafter statusline` subcommand renders plan position (Go + unit tests)

- [x] Step 1.1: Add the `statusline` cobra subcommand wired into the root command, reading stdin JSON and printing only on success — a thin command layer that delegates parsing/rendering to a new internal package, following the `pr_body.go` idiom (always exit 0, print empty line / nothing on any error).
- [x] Step 1.2: Implement active-task resolution in the internal package — given a working directory (from `workspace.current_dir`, falling back to cwd) and the current git branch, locate the matching active task file per A1/A2, with silent empty result when no project / no match.
- [x] Step 1.3: Implement plan parsing + rendering in the internal package — detect plan state (none/draft/approved per A5), count phases and work-steps excluding gates (A3/A4), and render the segment string for executing, planning/awaiting-approval, and empty states (A6/A7).
- [x] Step 1.4: Add table-driven unit tests covering: each plan state, gate-exclusion counting, multi-phase current-phase detection, percent/bar math, branch-mismatch (no render), missing/empty stdin, and a non-Crafter directory; include at least one live-fixture smoke test against a real phased task file in `.crafter/tasks/` (skipped if absent), mirroring the prbody live-test pattern.
- [x] Phase 1 verification
- [x] Phase 1 review

#### Phase 2 — Opt-in install wiring registers the segment without clobbering (install.sh + tests)

- [x] Step 2.1: Add a `--with-statusline` flag to `install.sh` argument parsing (default off; documented in `usage()`), threaded through to global and local install paths.
- [x] Step 2.2: Add a `install_statusline()` routine (invoked only when the flag is set) that mirrors `install_hook()`'s Node-based settings.json edit: resolve the correct binary invocation for the install mode (global vs local; global uses the `~/.local/bin/crafter` / `$base/crafter/bin/crafter` location, local uses the project `.claude/crafter/bin/crafter`), and set the top-level `statusLine` key to `{ "type": "command", "command": "<crafter> statusline" }` ONLY when no `statusLine` exists. On collision, leave the existing value untouched and print a ready-to-paste composite wrapper command (tee stdin to both the existing command and `crafter statusline`, join outputs) to stdout.
- [x] Step 2.3: Add `tests/test_install.sh` cases: default install (global + local) does NOT add `statusLine`; `--with-statusline` on a clean settings.json sets it to invoke `crafter statusline`; `--with-statusline` when a `statusLine` already exists does NOT overwrite it and prints the composite-wrapper guidance; opt-in wiring is idempotent.
- [x] Phase 2 verification
- [x] Phase 2 review

#### Phase 3 — Documentation describes the feature and opt-in (docs)

- [x] Step 3.1: Document `crafter statusline` and the `--with-statusline` opt-in in `README.md` (CLI Binary section and/or a short statusline subsection), including the rendered example and the composability/no-overwrite note; flag (do not edit here) that `.crafter/ARCHITECTURE.md` "Crafter CLI — Utility Binary" subcommand list and PROJECT.md may need a follow-up note — those ARCHITECTURE edits are delegated to the Implementer as a separate explicit step if the orchestrator approves.
- [x] Phase 3 verification
- [x] Phase 3 review

### 6. Karpathy Contract

**Phase 1 — subcommand + tests**
- Outcome: `crafter statusline` parses the active task and emits the correct segment for every state, fully unit-tested.
- Scope boundary: only `cli/` (new `cli/cmd/statusline.go`, new `cli/internal/<pkg>/`, tests). No install or docs changes.
- Non-goals: no install wiring, no GSD coupling, no config, no other-segment rendering.
- Simplicity constraint: reuse existing markdown-scanning idioms (bufio scanner, fence-awareness pattern from `decisions.go`); no new dependencies; one internal package.
- Drift criteria: introducing a config file, shelling to git without justification, rendering more than the Crafter segment, non-zero exit on error, or counting gate lines as steps.
- Verification evidence: `go test ./...` green in `cli/`; manual `echo '<json>' | crafter statusline` shows the example for an approved fixture and empty for a non-Crafter dir.
- Stop conditions: stop and flag if the real task-file checkbox/heading conventions cannot be parsed unambiguously into phase/step counts (heading-level or gate-detection ambiguity).

**Step 1.1** — Outcome: subcommand registered, stdin read, always exit 0. Scope: `cli/cmd/statusline.go` + root registration. Non-goals: parsing logic (delegated to internal pkg). Simplicity: mirror `pr_body.go`. Drift: business logic in the cmd layer. Evidence: `crafter statusline --help` works; piping empty stdin prints nothing, exit 0. Stop: if cobra root cannot register without touching unrelated commands.

**Step 1.2** — Outcome: active task file resolved or empty result. Scope: resolution function + git-branch read. Non-goals: plan parsing. Simplicity: prefer `.git/HEAD` read; single directory scan. Drift: adding a watcher/cache, recursive unbounded walks. Evidence: unit tests for match, no-match, branch-mismatch, non-project. Stop: if `.planning` legacy support proves to materially expand scope — flag and default to `.crafter` only.

**Step 1.3** — Outcome: correct segment string per state. Scope: parse + render functions. Non-goals: I/O, file discovery. Simplicity: one render function with explicit state switch; glyph constants reused. Drift: ANSI color that breaks plain-text readability; adding fields beyond phase/steps/bar/percent. Evidence: unit tests assert exact strings for executing/draft/pending/empty. Stop: if gate-exclusion can't be done deterministically from wording.

**Step 1.4** — Outcome: table-driven tests cover all states + edges. Scope: `*_test.go` in the internal pkg. Non-goals: changing production code beyond test-driven fixes. Simplicity: mirror prbody test style; live test skips if fixture absent. Drift: non-hermetic tests that fail when run elsewhere. Evidence: tests pass locally and are hermetic except the explicitly-skippable live smoke test. Stop: none expected.

**Phase 2 — install wiring**
- Outcome: opt-in `--with-statusline` wires the segment without ever clobbering an existing statusLine; default install untouched.
- Scope boundary: `install.sh` and `tests/test_install.sh` only.
- Non-goals: changing default install, authoring a tee wrapper script, touching the Go code.
- Simplicity constraint: mirror `install_hook()`'s Node-based JSON edit; reuse existing binary-path resolution; no new helper files.
- Drift criteria: overwriting an existing `statusLine`, making statusline wiring run by default, editing `~/.claude/crafter/`, or adding a separate wrapper script artifact.
- Verification evidence: `tests/test_install.sh` green; default-install assertions confirm no `statusLine`; opt-in assertions confirm set/no-overwrite/idempotent.
- Stop conditions: stop and flag if Claude Code's `statusLine` schema differs from the reference shape observed in `~/.claude/settings.json`.

**Step 2.1** — Outcome: flag parsed and threaded. Scope: arg-parse + usage text. Non-goals: the settings edit. Simplicity: follow `--global/--local` parsing style. Drift: changing default mode. Evidence: `--help` lists the flag; unknown-flag guard still works. Stop: none.

**Step 2.2** — Outcome: settings.json edited correctly (set-if-absent, print-on-collision), per install mode. Scope: new install function + invocation guarded by the flag. Non-goals: default-path behavior. Simplicity: reuse the `install_hook()` Node-edit template. Drift: overwrite-on-collision; node-missing path must skip gracefully like `install_hook()`. Evidence: tested in 2.3. Stop: if global-vs-local binary path can't be resolved consistently with existing logic.

**Step 2.3** — Outcome: install tests prove all four behaviors. Scope: `tests/test_install.sh` additions. Non-goals: production install changes beyond test-driven fixes. Simplicity: reuse `_run_installer` harness and settings.json assertion style. Drift: tests that touch the real `~/.claude`. Evidence: suite green; new tests visible in output. Stop: none.

**Phase 3 — docs**
- Outcome: README documents the subcommand + opt-in; ARCHITECTURE/PROJECT follow-up flagged.
- Scope boundary: `README.md` (and only if orchestrator approves, a delegated ARCHITECTURE/PROJECT note step).
- Non-goals: rewriting docs structure; editing ARCHITECTURE within this step without approval.
- Simplicity constraint: a concise subsection consistent with existing CLI Binary / Skills docs tone.
- Drift criteria: scope creep into unrelated doc sections.
- Verification evidence: README renders the example and the opt-in/no-overwrite note; reviewer confirms accuracy against the implemented behavior.
- Stop conditions: none expected.

### 7. Alternatives considered

- **JS hook (GSD-style) instead of Go subcommand** — Ruled out by Decision 1: deterministic parsing belongs in the Go binary (consistent with `buffer`/`pr-body`/`skillbook`), testable in Go, no Node dependency at render time.
- **Installer auto-composes a tee wrapper on collision** — Ruled out by Decision 2: the platform `statusLine` key is singular and composition is environment-specific; silently authoring/installing a wrapper risks clobbering or surprising behavior. Printing a ready-to-paste wrapper keeps the user in control.
- **Shell out to `git` for the branch** — Viable but adds a process per render and a PATH dependency; reading `.git/HEAD` is cheaper and self-contained (A2). Left as an Implementer choice within the silent-fail constraint.
- **Including phase gate checkboxes in the step total** — Ruled out (A3): gates are workflow ceremony and would distort the user-facing "plan progress" reading; excluded by wording detection.
- **Reading plan state from STATE.md (like GSD)** — Ruled out: Crafter's authoritative plan position lives in the active task file's checkboxes (per `rules/do/step-0-resume.md`), which is finer-grained and matches Decision 3.

### 8. Risks / unknowns / flags

- **R1 — Heading level for phases (`###` vs `####`).** `templates/TASK.md` shows `### Phase`, but real task files (e.g. the slice-5 file) use `#### Phase N`. The parser must accept both. Flagged so the Implementer treats "any heading line containing `Phase N`" robustly rather than assuming one level.
- **R2 — Gate/post-change checkbox detection.** Excluding gates (A3) relies on conventional wording (`Phase N verification`, `Phase N review`, `STATE.md`, `Task file completion`, `Follow-up note`). If a future plan phrases gates differently, counts could drift. Recommend matching on the established gate wording and treating only clearly-labeled work steps (e.g. `N.M` numbered steps or non-gate checkboxes) as steps; confirm the heuristic against several real task files during Step 1.4.
- **R3 — `statusLine` schema stability.** The reference shape is taken from the local `~/.claude/settings.json`. If the Claude Code `statusLine` schema differs across versions, the installer wiring may need adjustment. Low risk (matches documented GSD usage) but flagged.
- **R4 — `.planning/` legacy fallback scope.** Whether the subcommand must support legacy `.planning/tasks/` is unclear. Recommend `.crafter/` only for the first cut (matches the project's own layout) and flag if legacy support is desired — adding it later is cheap.
- **R5 — Local-install branch/cwd mismatch.** For local installs, the `crafter statusline` binary path is project-relative; if a user opens Claude Code in a different cwd the segment will correctly render empty (no active project found). This is acceptable per Decision 4 but worth noting for the docs.
- **R6 — ARCHITECTURE/PROJECT edits.** Per the task brief these are delegated to the Implementer as a flagged follow-up, not performed by the Planner. Phase 3 only flags them; the orchestrator should confirm whether to include the ARCHITECTURE subcommand-list update in this task or defer it.

This contract protects a single, surgical outcome: a tested, silent-fail-safe `crafter statusline` segment plus strictly opt-in, never-clobbering install wiring and accurate docs — keeping the change deterministic, in the Go binary where Crafter's other parsers live, and fully reversible (default install behavior is unchanged).

## Decisions
- **Decision (User Accepted):** Implement as a Go subcommand `crafter statusline` (not a standalone JS hook). **Reason:** Matches the established Key Decision that deterministic rendering/parsing belongs in the Go binary (`crafter buffer`, `crafter pr-body`, `crafter skillbook`); testable in Go, no Node dependency at render time.
- **Decision (User Accepted):** `crafter statusline` prints only its own composable segment (no assumption it is the sole statusline). `install.sh` registers it via an opt-in `--with-statusline` flag and MUST NOT overwrite an existing `statusLine`. **Reason:** The platform `statusLine` key is singular (one command, one line); composability lets the Crafter segment be appended to an existing statusline (e.g. GSD) without clobbering it. On collision the installer prints a ready-to-paste composite wrapper command rather than overwriting.
- **Decision (User Accepted):** Display granularity is phase + steps + bar, e.g. `crafter · Phase 2/3 · 7/12 [███████░░] 58%`. **Reason:** Most informative; uses Crafter's fine-grained step-checkbox data.
- **Decision (User Accepted):** Empty/edge states — render nothing when not a Crafter project or no active task; when an active task exists but the plan is `_(pending)_`/`draft`, show a planning/awaiting-approval state rather than checkbox counts. **Reason:** Avoids noise outside Crafter while still signalling pre-execution states.
- **Decision (User Accepted):** The `ARCHITECTURE.md` CLI-subcommand-list update (adding `crafter statusline`) is IN SCOPE for this task (Phase 3 / post-change), not a separate follow-up. **Reason:** Keeps the documented CLI surface consistent within the same change that introduces the subcommand.
- **Decision (Orchestrator Accepted):** Step 2.3 fixed a genuine bug in `install_statusline()` (introduced in Step 2.2): the `node -e` body used a top-level `return`, which Node v26 rejects as `SyntaxError: Illegal return statement` (under `-e`, scripts run in module context), so every `--with-statusline` install failed with exit 1. Fix: wrapped the `node -e` body in an IIFE `(function(){ … })()`. **Reason:** Minimal, test-driven correction enabling the Step 2.3 tests to pass; preserves the set-if-absent / print-on-collision behavior; flagged by the Implementer and confirmed by the Verifier; no effect on later steps.
- **Decision (Orchestrator Accepted):** Step 1.2 introduced a dead no-op (`_ = fmt.Sprintf(...)`) and an otherwise-unused `fmt` import in `cli/internal/statusline/resolve.go` (artifact of the silent-fail-no-stderr posture). Deferred its removal to Step 1.3 rather than spawning a separate fix. **Reason:** Step 1.3 already modifies the same package (`Render` + the resolve error path); the cleanup is a one-line local change with no behavioral impact and no effect on later steps.

## Outcome

All three phases delivered, verified, and reviewed clean. Suite grew 47 → 58 install tests; Go `cli/internal/statusline` tests green.

**Phase 1 — `crafter statusline` subcommand (commit `c904718`):** new Go subcommand (`cli/cmd/statusline.go` + `cli/internal/statusline/`) reads Claude Code's stdin JSON, resolves the active task file (mirroring resume detection: `**Status:** active` + `**Work branch:**` matching the current branch via `.git/HEAD`), parses the plan, and prints one composable segment (`crafter · Phase 2/3 · 7/12 [█████░░░░░] 58%`). Always exit 0; empty output on any problem; planning (`_(pending)_`) and awaiting-approval (`draft`) edge states; gate/post-change checkbox lines excluded from step counts. Table-driven unit tests + skippable live-fixture smoke test.

**Phase 2 — opt-in install wiring (commit `722f1c2`):** `--with-statusline` flag (default off) + `install_statusline()` mirroring `install_hook()`'s Node settings.json edit. Sets top-level `statusLine` only when absent; on collision never overwrites and prints ready-to-paste composite-wrapper guidance (a `bash -c '...'` that tees stdin once to both the existing command and `crafter statusline`). Both global and local modes. 11 install tests (section G).

**Phase 3 — documentation (commit `1f73d5e`):** README `### Statusline` subsection (segment format, edge states, silent-fail, opt-in flag, never-overwrite policy); ARCHITECTURE.md CLI subcommand enumeration + source-tree entries; PROJECT.md Key Decisions row.

**Notable fix-loop findings (all resolved):** collision guidance originally used a never-set `$STDIN` (Major) → correct stdin-tee; guidance was not valid pasteable JSON → built via `JSON.stringify`; single quotes in an existing command broke the `bash -c '...'` wrapper (Major) → POSIX `'\''` escaping; the single-quote regression test used `bash -n -c` which only checks outer syntax (false-confidence) → now executes the command to detect syntax errors; the rendered bar example (`[███████░░░]` for 58%) was inconsistent with `filled = pct/10` → corrected to `[█████░░░░░]` in both README and ARCHITECTURE.md.

**Deviation from plan:** Step 2.2's `node -e` body uses an IIFE wrapper because Node v26 rejects a top-level `return` under `-e` (recorded as a Decision). No scope changes; all four user-accepted decisions honored (Go subcommand, opt-in non-clobbering install, phase+steps+bar+percent granularity, edge states; ARCHITECTURE update in scope).
