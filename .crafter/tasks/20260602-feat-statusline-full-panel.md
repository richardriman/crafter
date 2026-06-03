# Task: Extend crafter statusline into a full single-line status panel

## Metadata
- **Date:** 2026-06-02
- **Work branch:** feat/plan-progress-statusline
- **Status:** active
- **Scope:** Medium

## Request

Extend `crafter statusline` from a plan-position segment into a full single-line status panel that REPLACES the GSD status line (crafter panel = complete standalone panel; whoever installs it does not use GSD). Build on the current code of branch `feat/plan-progress-statusline` (the fallback cascade has already landed). Scope of THIS task = CLI rendering only (`cli/internal/statusline/*` + `cli/cmd/statusline.go` + tests); rewiring `install.sh`/`settings.json` (compose→replace) is a separate follow-up, NOT in this task.

**Context** — the statusline command receives a JSON payload from Claude Code on stdin. Verified against official docs that these fields are available for free (no transcript parsing):
- `model.display_name`
- `effort.level` — values low/medium/high/xhigh/max; field ABSENT when the model does not support effort
- `context_window.used_percentage` — number or null (null early in session / after `/compact`)
- `context_window.context_window_size` — 200000 default, 1000000 extended
- `cost.total_cost_usd`, `cost.total_lines_added`, `cost.total_lines_removed`
- `workspace.project_dir`
- branch is NOT in the payload → derive from git (verify how the current resolver does it after the cascade rewrite — `resolveActiveTask` was removed, replaced by `extractTaskMeta`/`classifyTasks`)

**Target layout** — one line, sections separated by `" │ "` (U+2502 with surrounding spaces), leading `"crafter"` label REMOVED:

    <plan> │ <model> │ <branch+diff> │ <ctx> │ <cost>

Sections:
- **plan**: keep the existing render `Phase 1/2 · 2/7 [██░░░░░░░░] 29%` but WITHOUT the leading `"crafter · "` prefix. Only when an active task exists.
- **model**: `Opus 4.8 1M (high)` = `model.display_name` + `" "` + capacity (`context_window_size` abbreviated: 1000000→`"1M"`, 200000→`"200k"`) + `" ("` + `effort.level` + `")"`. Verify the real `display_name` value.
- **branch+diff**: `⎇ <branch> +120/-30` — branch icon CONFIGURABLE with default `"⎇"` (U+2387, not Nerd font — not everyone has it); `+N` green (ANSI `\033[32m`), `-N` red (`\033[31m`) from `total_lines_added`/`total_lines_removed`.
- **ctx**: `[████░░░░░░] 42%` = progress bar (same style as the plan bar, `glyphFilled`/`glyphEmpty`) + `%` from `used_percentage`.
- **cost**: `$0.42` from `total_cost_usd`.

**Graceful degradation** (critical):
- `effort.level` absent → omit the `" (...)"` in the model section, keep the rest.
- `context_window.used_percentage` is null → omit the entire ctx section.
- no active task → plan section disappears, but the panel MUST render the rest (`model │ branch │ ctx │ cost`) — this is the "always returns something"; it must not collapse to an empty string just because there is no task. Reconcile with the existing fallback cascade (the rungs that prepend `"crafter ·"`).
- preserve the silent-fail posture (errors → degrade, never panic); always exit 0.

The approved section order is the layout above. Project name omitted for now (follow-up). `rate_limits`/PR/vim/`session_name` deferred.

Extend `statusline_test.go` with the new sections + degradation paths. Respect repo policy `CLAUDE.md` (no AI signatures in commits). Go builds via mise (`go` not on PATH — use `mise exec --`).

## Plan
**Plan status:** approved

### Approach

Today `crafter statusline` is a *segment producer*: `Render(workdir)` returns one short string chosen by a four-rung cascade, and the cascade strings all start with the literal `crafter · …`. The new spec inverts the design: the command becomes a *full status panel* assembled from up to five independent sections (`plan │ model │ vcs │ ctx │ cost`) joined by `" │ "`, with no leading `crafter` label, and it must **always render something** (model/vcs/ctx/cost don't depend on a Crafter task existing). The **vcs** section is a group: `<project> ⎇ <branch> +N/-N` — project name (dim grey), branch icon + branch, and the colored diff all share one ` │ `-delimited group.

The strategy is to treat the existing cascade output as exactly one section — the **plan section** — and build a thin assembler around it. The cascade keeps its *priority* logic but loses its `crafter · ` prefix and its rung-4 empty-string behavior at the panel level. The Claude Code stdin JSON payload (already read in `cli/cmd/statusline.go`) is the data source for the other four sections; the JSON struct must be extended to decode the new fields. Branch comes from git (already available via `readGitBranch`). The whole thing preserves always-exit-0 / silent-fail: every section degrades to "omit me" on missing/bad data, and the panel renders whatever sections survived.

This is a Medium, CLI-rendering-only change. No install.sh / settings.json rewiring (explicit follow-up). Three vertical phases, each leaving a coherent, tested state.

### Central architectural decision — reconciling the cascade with the always-on panel

The existing cascade produces four mutually-exclusive outcomes (active-current full segment / `✓ done` / `N active elsewhere` / empty), each prefixed `crafter · `, and rung 4 returns `""` so the caller prints nothing. The new panel removes the `crafter` label and must print model/branch/ctx/cost even when there is no task. These collide. Options weighed:

- **Option A — Demote the cascade to a "plan section" producer; a new panel assembler always runs.** Introduce a panel-assembly entry point that calls a *plan-section* function (the renamed/refactored cascade, with the `crafter · ` prefix dropped and rung-4 returning empty-section rather than empty-panel), then appends model/branch/ctx/cost sections, joins non-empty sections with `" │ "`, and returns the result. `Render(workdir)` becomes the panel assembler (or a new entry point that `runStatusline` calls). The cascade's *priority* semantics survive verbatim; only its prefix and its panel-level emptiness change.
- **Option B — Keep `Render` as-is, prepend/append the new sections in the cmd layer.** Rejected: scatters rendering across `cmd/statusline.go` and the package, breaks the silent-fail/test boundary, and leaves the `crafter · ` prefix fighting the spec.
- **Option C — Rewrite the cascade away entirely and fold task state into a flat panel.** Rejected: throws away tested rung precedence (rung1>rung2>rung3, tie-breaks, known-limitations) for no benefit; the spec explicitly says "reconcile with," not "delete."

**Decision: Option A.** The cascade becomes the *plan section* of the panel. Concretely:
- The four rungs keep their priority order and selection logic, but their output strings drop the leading `"crafter · "` (rung 1: `Phase 1/2 · 2/7 [██░░░░░░░░] 29%`; rung 2: `✓ done`; rung 3: `N active elsewhere`).
- Rung 4 ("nothing relevant") yields an **empty plan section**, NOT an empty panel — the panel then renders `model │ branch │ ctx │ cost`.
- The non-Crafter / no-git guards (`findCrafterDir`/`readGitBranch` returning empty) no longer abort the whole panel: they only suppress the plan section. The panel still renders the payload-derived sections. (Branch section still needs git; if git is absent the branch section degrades independently.)
- The panel joins only the non-empty sections with `" │ "`. If *every* section is empty (no task, payload empty, no git), the panel returns `""` — that is the one legitimate empty-panel case (and matches "graceful degradation to nothing renderable"), not a violation of "always render something," which the spec scopes to "don't collapse just because there's no task."

The Implementer chooses the concrete function names/signatures; the contract above fixes the behavior.

### Branch-icon configuration decision

Spec requires the branch glyph configurable, default `⎇` (U+2387), kept simple (Karpathy: smallest change). Options weighed:

- **Option A — Environment variable** (e.g. `CRAFTER_STATUSLINE_BRANCH_ICON`), read once at render time, default `⎇` when unset/empty. Smallest change: no flag plumbing through cobra, no config-file schema, statusline is invoked non-interactively by Claude Code so an env var set in the shell/Claude settings is the natural override surface. **Chosen.**
- **Option B — Cobra `--branch-icon` flag.** Rejected as larger: Claude Code's `statusLine` invocation is a fixed command line; flags are awkward to thread through settings.json and the spec defers install wiring. Still possible later without breaking the env var.
- **Option C — Config field in a crafter config file.** Rejected: no such config file exists today; introducing one is out of scope and over-engineered for a single glyph.

**Decision: Option A (env var).** Exact env var name is the Implementer's call within the `CRAFTER_STATUSLINE_*` namespace; default must be `⎇`. Empty/unset → default. The mechanism reads the value at panel-assembly time so tests can set it via `t.Setenv`.

### Assumptions / interpretations

- **A1.** `model.display_name` for Opus is literally `"Opus 4.8"` (or whatever Claude Code sends) — the renderer concatenates it verbatim; we do not hard-code or normalize it. The `"1M (high)"` suffix is composed from `context_window_size` + `effort.level`, not from `display_name`.
- **A2.** Capacity abbreviation uses **general k/M** (decided): `≥1_000_000 → "M"` (e.g. `1000000 → "1M"`), `≥1_000 → "k"` (e.g. `200000 → "200k"`, `128000 → "128k"`, `500000 → "500k"`), raw number below 1000. This covers sizes beyond the two pinned examples. (Per `## Decisions`.)
- **A3.** The model section renders even if capacity or effort is missing: `display_name` alone is the minimum. If `display_name` is empty → omit the whole model section.
- **A4.** The ctx bar reuses `glyphFilled`/`glyphEmpty`/`barSegments` and the same fill math as the plan bar, applied to `used_percentage` (0–100) instead of step ratio. `used_percentage` null → omit ctx section entirely (not "0%").
- **A5.** Diff numbers come from `cost.total_lines_added` / `cost.total_lines_removed`. They render as `+N` (green) `/-N` (red) appended to the branch token *only when there are changes*. When **both are zero**, the `+N/-N` suffix is **omitted entirely** (the branch token still renders). (Per `## Decisions` — resolves the former R2.)
- **A6.** Cost renders `$X.XX` (2 decimals) from `cost.total_cost_usd`, but the **cost section is omitted when `total_cost_usd` is zero or absent** (render only when cost > 0). This requires distinguishing zero from absent — the field is a pointer type (`*float64`) so a missing field and a `0` both omit the section, and any value > 0 renders. (Per `## Decisions` — resolves the former R3.)
- **A7.** The JSON struct currently lives in `cli/cmd/statusline.go` as `statuslineInput` and only decodes `workspace.current_dir`. The new fields are added there (or a struct is moved into the package). We keep `workspace.current_dir` for workdir resolution (don't regress) and **additionally decode `workspace.project_dir`**, whose basename feeds the project-name part of the VCS group (A10). (Per `## Decisions` — `project_dir` is now in scope; the earlier deferral is reversed.)
- **A8.** ANSI codes are emitted raw and rendered by Claude Code's statusLine: `\033[32m` green / `\033[31m` red for diff, `\033[2m` dim grey for the project name, reset `\033[0m`. Tests assert on the raw escape sequences.
- **A9.** Section separator is exactly `" │ "` (space, U+2502, space). Empty sections are dropped *before* joining so there are never doubled or leading/trailing separators.
- **A10.** Project name = basename of `workspace.project_dir`, styled dim grey (`\033[2m…\033[0m`, matching GSD). It sits at the **head of the VCS group**, immediately before the branch icon: `<project> ⎇ <branch> +N/-N`. Project, branch+icon, and diff are one ` │ `-delimited group. Within the group: empty/absent `project_dir` → omit only the project name (group still renders branch+diff); empty branch (no git/detached HEAD) → omit only the branch token (project name may render alone if present); both absent → the whole VCS group is empty and dropped from the panel. The single-space joins inside the group are also collapsed so a missing project name does not leave a leading space before `⎇`.

### Non-goals (whole task)

- No `install.sh` / `settings.json` changes (compose→replace wiring is a separate follow-up).
- No `rate_limits`, no PR/vim/`session_name` sections (all deferred). (Project name is **in scope** — it lives in the VCS group, see A10.)
- No new config file format; no cobra flag for the icon.
- No change to plan-parsing logic (`parsePlan`, gate detection, phase math) beyond dropping the `crafter · ` prefix at the render boundary.
- No transcript parsing — all data is from the stdin JSON payload + git.

### Relevant areas

- `cli/cmd/statusline.go` — stdin read + `statuslineInput` struct (extend with model/effort/context_window/cost; keep `workspace.current_dir`; **add `workspace.project_dir`**). The command passes payload data into the package renderer.
- `cli/internal/statusline/statusline.go` — `Render` entry + cascade orchestration + `segDone`/`segActiveElsewhereFmt` constants (drop `crafter · ` prefix; demote to plan section; add panel assembler).
- `cli/internal/statusline/parse.go` — `renderSegment`/`renderExecuting` (drop the `"crafter"` literal and `crafter · plan: …` strings from the plan-section output; the bar/glyph helpers are reused for ctx).
- `cli/internal/statusline/resolve.go` — `readGitBranch` (reused for the branch token in the VCS group; classification logic unchanged).
- `cli/internal/statusline/statusline_test.go` — extend with model / vcs-group (project+branch+diff) / ctx / cost sections, degradation paths, panel assembly, and updated cascade expectations (the existing rung tests assert the `crafter · ` prefix and must be updated to the new prefix-free strings).

---

### Phase 1 — Plan section loses its prefix; panel skeleton assembles around it

**Outcome:** The cascade no longer emits `crafter · `; its output is a *plan section*. A panel assembler joins the plan section (and, for now, only the plan section) into the final output, and the no-task case no longer aborts the whole panel — it just yields an empty plan section. Existing cascade behavior (priority, tie-breaks, guards) is preserved minus the prefix.

- [x] **Step 1.1 — Strip the `crafter · ` prefix from all plan-section output.** Update `renderSegment`/`renderExecuting` and the `segDone`/`segActiveElsewhereFmt` constants so the plan section reads `Phase 1/2 · 2/7 [██░░░░░░░░] 29%`, `✓ done`, `N active elsewhere`, `plan: awaiting approval`, `planning`. Update all existing tests that assert the old `crafter · …` strings.
  - *Outcome:* plan-section strings match the new spec; no `crafter` literal remains in section output.
  - *Scope boundary:* string/format changes only inside `parse.go` + the two constants; no math, no parsing changes.
  - *Non-goals:* don't touch cascade priority, don't add new sections yet.
  - *Simplicity constraint:* pure search-replace of the prefix; reuse existing bar/glyph code untouched.
  - *Drift criteria:* if you find yourself changing phase/step math, gate logic, or classification — stop, that's out of scope.
  - *Verification evidence:* `mise exec -- go test ./cli/internal/statusline/` green after updating the prefix assertions; grep shows no `"crafter ·"` literal in `parse.go`/`statusline.go` section output.
  - *Stop conditions:* tests pass with new prefix-free strings; nothing else changed.
- [x] **Step 1.2 — Introduce the panel assembler and demote the cascade to a plan-section producer.** Add a panel-assembly path that produces the plan section via the (now prefix-free) cascade, then joins non-empty sections with `" │ "`. In this step the only section is the plan section, so the panel output for an active/approved task equals the plan section string, and the no-task / no-Crafter / rung-4 case yields `""` (because no other sections exist yet). The non-Crafter and no-git guards must no longer short-circuit the entire panel — they only suppress the plan section.
  - *Outcome:* a single assembler function returns the joined panel; cascade returns a section, not a panel.
  - *Scope boundary:* `statusline.go` orchestration + signatures; `cmd/statusline.go` calls the new entry point. No payload fields yet.
  - *Non-goals:* don't add model/branch/ctx/cost; don't change `cmd` JSON struct yet.
  - *Simplicity constraint:* the join is "filter empty, `strings.Join(_, \" │ \")`"; no premature section abstraction beyond what two phases need.
  - *Drift criteria:* if the assembler grows section-specific logic inline (instead of delegating to per-section helpers added in Phase 2), stop and reconsider the seam.
  - *Verification evidence:* existing Render integration tests pass with adjusted expectations (active task → prefix-free plan string; empty dir → `""`); `mise exec -- go test ./cli/...`.
  - *Stop conditions:* panel == plan section for all current cases; guards no longer abort the panel.
- [x] **Phase 1 verification** — `mise exec -- go test ./cli/...` green; all updated cascade/Render tests reflect prefix-free strings; manual `echo '{}' | crafter statusline` (no task) returns empty (no other sections yet) and an active-task fixture returns the prefix-free plan string. Evidence: paste the failing-then-passing test names and one manual invocation.
- [x] **Phase 1 review** — diff confined to `parse.go`, `statusline.go`, `cmd/statusline.go`, and test updates; no parsing/math changes; cascade priority intact.

### Phase 2 — Payload-derived sections: model, vcs (project+branch+diff), ctx, cost

**Outcome:** The four payload-derived sections render per spec with full graceful degradation, joined after the plan section. The panel now "always renders something" whenever the payload carries usable data, regardless of task state. The **vcs** section is a group — `<project> ⎇ <branch> +N/-N` — with dim-grey project name, normal-styled branch+icon, and green/red diff. Branch icon configurable via env var (default `⎇`).

- [x] **Step 2.1 — Extend the stdin JSON struct to decode the new fields.** Add `model.display_name`, `effort.level`, `context_window.used_percentage` (pointer/`*float64` or `json.Number` so null is distinguishable from 0), `context_window.context_window_size`, `cost.total_cost_usd` (**pointer `*float64`** so zero/absent is distinguishable from a positive value, per A6), `cost.total_lines_added`, `cost.total_lines_removed`, and `workspace.project_dir` (per A7). Keep `workspace.current_dir`. Thread the decoded values into the panel assembler.
  - *Outcome:* the renderer receives all payload data it needs; null `used_percentage` and zero/absent `total_cost_usd` are each distinguishable from a real value.
  - *Scope boundary:* struct definition + plumbing from `cmd` into the package; no rendering yet.
  - *Non-goals:* don't render sections in this step.
  - *Simplicity constraint:* mirror the existing `statuslineInput` style; unknown fields ignored as today.
  - *Drift criteria:* if null-vs-zero (ctx) or zero-vs-absent (cost) can't be represented, fix the type now (don't paper over with sentinel 0).
  - *Verification evidence:* a decode unit test feeding a representative payload asserts each field (including absent `effort.level`, null `used_percentage`, absent vs zero `total_cost_usd`, and `project_dir`). `mise exec -- go test ./cli/...`.
  - *Stop conditions:* all fields decode; null/absent cases are observably distinct.
- [x] **Step 2.2 — Render the model and cost sections.** Model: `display_name + " " + abbrev(context_window_size) + " (" + effort.level + ")"`; omit `" (...)"` when `effort.level` absent; omit the whole section when `display_name` empty. Capacity abbreviation uses general k/M (`≥1_000_000 → "1M"`, `≥1_000 → "<n>k"`, raw below 1000) per A2. Cost: `$` + 2-decimal `total_cost_usd`, but **omit the cost section entirely when `total_cost_usd` is zero or absent** (render only when > 0) per A6.
  - *Outcome:* `Opus 4.8 1M (high)` and `$0.42` render and degrade (effort absent → `Opus 4.8 1M`; cost 0/absent → no cost section).
  - *Scope boundary:* two section helpers + their wiring into the assembler.
  - *Non-goals:* no vcs/ctx yet; no color codes here.
  - *Simplicity constraint:* capacity abbreviation is a tiny k/M helper; reuse for nothing else.
  - *Drift criteria:* if abbreviation logic grows past the simple k/M rule, stop.
  - *Verification evidence:* table tests for model (with/without effort; 1M / 200k / 128k / sub-1000; empty display_name) and cost (positive → `$X.XX`; zero → omitted; absent → omitted); `mise exec -- go test ./cli/internal/statusline/`.
  - *Stop conditions:* both sections match expected strings and degradation.
- [x] **Step 2.3 — Render the vcs group: project + branch + diff, with ANSI colors and configurable icon.** Build one grouped section `<project> ⎇ <branch> +N/-N`: (a) project name = basename of `workspace.project_dir`, dim grey (`\033[2m…\033[0m`); (b) branch icon (env var, default `⎇`) + branch from `readGitBranch`, normal style; (c) diff `+N` green (`\033[32m…\033[0m`) `/` `-N` red (`\033[31m…\033[0m`) from `total_lines_added`/`total_lines_removed`, **rendered only when the changes are non-zero** (both zero → omit the `+N/-N` suffix, keep the branch token) per A5. Intra-group degradation per A10: empty `project_dir` → omit only the project name (no leading space before `⎇`); empty branch → omit only the branch token; both empty → the whole vcs group is empty and dropped by the assembler.
  - *Outcome:* `crafter ⎇ feat/plan-progress +120/-30` with dim project + colored diff; custom icon via env var; each part degrades independently; no stray spaces when a part is omitted.
  - *Scope boundary:* one grouped-section helper + the env-var read + project-dir basename; reuse `readGitBranch`.
  - *Non-goals:* don't change `readGitBranch`/classification; don't add a cobra flag; don't split project into its own ` │ ` section (it is grouped, per A10).
  - *Simplicity constraint:* env var read once; raw ANSI constants; `filepath.Base` for the project name; no color library.
  - *Drift criteria:* if you need to shell out to `git diff` for line counts, stop — counts come from the payload. If you emit the project name as a separate top-level section, stop — it belongs inside the vcs group.
  - *Verification evidence:* tests assert the raw escape sequences (dim project, green/red diff), the default icon, a `t.Setenv` icon override, diff-omitted-when-zero, project-omitted-when-`project_dir`-empty (no leading space), branch-omitted-when-no-git, and whole-group-omitted-when-both-absent; `mise exec -- go test`.
  - *Stop conditions:* grouped vcs string renders correctly with all four intra-group degradation paths passing.
- [x] **Step 2.4 — Render the ctx section reusing the plan bar.** `[████░░░░░░] 42%` from `used_percentage` using `glyphFilled`/`glyphEmpty`/`barSegments` and the same fill math; omit the entire section when `used_percentage` is null.
  - *Outcome:* ctx bar renders at the given percentage; null → section absent.
  - *Scope boundary:* one section helper reusing existing bar code.
  - *Non-goals:* don't duplicate the bar logic — factor a shared bar builder if needed, but keep it minimal.
  - *Simplicity constraint:* if a shared bar helper is extracted, both plan and ctx call it; no behavior change to the plan bar.
  - *Drift criteria:* if extracting the shared bar changes plan-bar output, stop and preserve it.
  - *Verification evidence:* tests for several percentages (0/42/100), null omission, and that the plan bar output is unchanged; `mise exec -- go test`.
  - *Stop conditions:* ctx renders/degrades; plan bar regression-free.
- [x] **Phase 2 verification** — `mise exec -- go test ./cli/...` green. Manual evidence with example payloads piped to `crafter statusline`:
  - Full: `{"model":{"display_name":"Opus 4.8"},"effort":{"level":"high"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.42,"total_lines_added":120,"total_lines_removed":30},"workspace":{"current_dir":"<repo>","project_dir":"<repo>"}}` → `… │ Opus 4.8 1M (high) │ crafter ⎇ <branch> +120/-30 │ [████░░░░░░] 42% │ $0.42` (plan section present only if an active task is on the branch; project name = basename of `project_dir`, dim grey; diff green/red).
  - Degraded: same payload minus `effort`, with `used_percentage:null`, `total_cost_usd:0`, `total_lines_added:0`/`total_lines_removed:0` → model shows `Opus 4.8 1M`, ctx section absent, cost section absent, vcs group shows `crafter ⎇ <branch>` with no `+/-` suffix; separators correct (no doubled/leading/trailing `│`).
  - Project-absent: payload with empty `project_dir` → vcs group is `⎇ <branch> +N/-N` (no leading space before `⎇`).
  - No-task: payload with no matching task → plan section absent, rest renders.
  - Paste raw outputs (escape sequences visible) as evidence.
- [x] **Phase 2 review** — section helpers are isolated and independently degradable; the vcs group degrades part-by-part (project / branch / diff) with no stray spaces; ANSI codes correct (dim project, green/red diff); env-var icon default is `⎇`; cost omitted at zero/absent; capacity uses k/M; no doubled/leading/trailing separators; null-vs-zero and zero-vs-absent handled.

### Phase 3 — Panel cohesion, full degradation matrix, docs

**Outcome:** The assembled panel is verified end-to-end across the full degradation matrix (every section present, each section individually absent, all-absent → empty), the "always render something" invariant holds whenever any payload data exists, and the user-facing docs describing `crafter statusline` are updated to the panel behavior. Always-exit-0 / silent-fail confirmed.

- [ ] **Step 3.1 — Panel-assembly degradation matrix tests.** One test (or table) exercising: all five sections present (vcs group = project + branch + diff); plan absent; model absent (empty display_name); ctx absent (null `used_percentage`); cost absent (zero/absent `total_cost_usd`); and the vcs-group intra-degradation rows — project present + branch present + diff present; project absent (group = `⎇ branch +N/-N`); branch absent (group = `<project> +N/-N`? per A10 the branch token alone is omitted, so group = `<project>` then diff only if changes — assert the exact A10 outcome); diff zero (group = `<project> ⎇ branch`); whole vcs group absent (no project, no branch); and the all-absent case → `""`. Assert exact joined strings (separators, order, no stray spaces in the group) and the no-double-separator property.
  - *Outcome:* the assembler's join/filter behavior AND the vcs group's intra-degradation are pinned for every combination that matters.
  - *Scope boundary:* tests only (plus any tiny assembler/group fix the matrix reveals).
  - *Non-goals:* no new sections; no behavior change unless a real bug surfaces.
  - *Simplicity constraint:* drive through the panel entry point, not internal helpers, so the test reflects real output.
  - *Drift criteria:* if a "fix" the matrix prompts touches section semantics rather than the join/group assembly, re-scope it to Phase 2.
  - *Verification evidence:* `mise exec -- go test ./cli/internal/statusline/ -run Panel` green; matrix cases (including project present/absent and the vcs-group rows) enumerated in test names.
  - *Stop conditions:* every matrix row — including all project/branch/diff combinations inside the vcs group — asserts the exact expected panel string.
- [ ] **Step 3.2 — Confirm always-exit-0 / silent-fail at the command boundary.** Verify (test or scripted invocation) that malformed JSON, empty stdin, and a non-Crafter / non-git directory all exit 0 and never panic, with the panel degrading to whatever data exists (possibly `""`).
  - *Outcome:* the command never breaks the status bar; exit code is always 0.
  - *Scope boundary:* `cmd/statusline.go` posture + a guard test; no rendering changes.
  - *Non-goals:* don't add new error reporting.
  - *Simplicity constraint:* reuse the existing silent-fail pattern; don't restructure error handling.
  - *Drift criteria:* if you add any `os.Exit(1)` or error return that escapes, stop.
  - *Verification evidence:* `printf 'garbage' | crafter statusline; echo "exit=$?"` → `exit=0`; `echo '' | crafter statusline; echo "exit=$?"` → `exit=0`. Paste both.
  - *Stop conditions:* all three inputs exit 0 with no panic.
- [ ] **Step 3.3 — Update docs to describe the full panel.** Update the `crafter statusline` description in `.crafter/ARCHITECTURE.md` (CLI subcommands list) and `README.md` (and `PROJECT.md` if it documents the subcommand) from "single composable segment / four-rung cascade" to the full-panel behavior: the five sections `plan │ model │ vcs │ ctx │ cost`, the vcs group `<project> ⎇ <branch> +N/-N` (dim project name from `project_dir`, green/red diff, configurable icon), per-section degradation (effort/ctx/cost/project all optional), env-var icon, and always-renders-something. Keep install/settings wiring out (still a follow-up).
  - *Outcome:* docs match the shipped behavior; no stale "segment" framing.
  - *Scope boundary:* doc prose only; the subcommand bullet(s).
  - *Non-goals:* don't document install rewiring; don't invent config beyond the env var.
  - *Simplicity constraint:* edit the existing bullets in place; minimal prose.
  - *Drift criteria:* if doc edits imply behavior not implemented, stop and align.
  - *Verification evidence:* grep for the old "four-rung cascade / single composable segment" wording shows it updated; quote the new bullet.
  - *Stop conditions:* docs describe the panel accurately, install wiring still excluded.
- [ ] **Phase 3 verification** — `mise exec -- go test ./cli/...` fully green; degradation matrix + exit-0 evidence pasted; docs grep shows updated wording. Confirm the full-panel manual example and a degraded example one more time.
- [ ] **Phase 3 review** — no scope creep into install/settings; silent-fail/exit-0 intact; docs/code consistent; no AI signature in any commit (CLAUDE.md).

### Alternatives considered (summary)

- **Cascade reconciliation:** Option A (demote cascade to plan section) chosen over B (assemble in cmd layer — scatters logic) and C (delete the cascade — loses tested precedence). Reasoning above.
- **Branch-icon config:** env var (chosen) over cobra flag (awkward through settings.json, larger) and config file (no such file exists; over-engineered).
- **Bar reuse for ctx:** reuse existing `glyphFilled`/`glyphEmpty`/`barSegments` (optionally via a small shared helper) rather than a second bar implementation, to keep the plan and ctx bars visually identical and avoid drift.
- **Null context handling:** `*float64`/`json.Number` to distinguish null from 0, rather than a sentinel `-1`, so "no data" cleanly omits the section.
- **Project-name placement:** grouped *inside* the vcs section (`<project> ⎇ <branch> +N/-N`) rather than as its own top-level ` │ ` section, per the user decision — project and branch are shared VCS context and read as one unit. Rejected a standalone project section (would add a separator and dilute the grouping the user asked for).
- **Cost/diff zero handling:** omit the cost section and the `+/-` diff suffix at zero (decided), using a pointer type for cost to distinguish zero/absent from a positive value — chosen over always rendering `$0.00` / `+0/-0`, which the user judged as noise at session start.

### Risks / unknowns / flags

- **R1 (display_name value) — OPEN.** We assume `model.display_name` is the human string Crafter wants (`Opus 4.8`). If Claude Code sends something else (e.g. a slug), the model section text differs. Mitigation: render verbatim; flag for user confirmation against a real payload.
- **R2 (diff zero behavior, A5) — RESOLVED.** Decided: omit the `+N/-N` suffix entirely when both line counts are zero; the branch token still renders. (See `## Decisions`.)
- **R3 (cost zero/absent, A6) — RESOLVED.** Decided: omit the cost section when `total_cost_usd` is zero or absent; use a pointer type to distinguish zero/absent from a positive value. (See `## Decisions`.)
- **R4 (project_dir, A7/A10) — RESOLVED.** Decided: `workspace.project_dir` is in scope — its basename is the dim-grey project name at the head of the vcs group. No longer dead code. (See `## Decisions`.)
- **R5 (capacity abbreviation, A2) — RESOLVED.** Decided: general k/M (`≥1_000_000 → M`, `≥1_000 → k`), covering all sizes. (See `## Decisions`.)
- **R6 (ANSI in tests/terminals) — OPEN (low).** Tests assert raw escape sequences; this is deterministic. Confirm Claude Code's statusLine renders ANSI (the spec asserts green/red and dim grey, implying yes).
- **R7 (existing test churn) — informational.** Every current cascade/Render test that asserts `crafter · …` must be updated in Phase 1. This is expected, not a regression — but the diff will touch many test lines; reviewer should expect it.
- **R8 (project_dir vs current_dir, A7/A10) — OPEN (low).** The basename of `workspace.project_dir` may differ from the repo dir used for git/task resolution (`current_dir`). For the project *name* we trust `project_dir` as the user specified; if it is empty we omit the name (we do not fall back to `current_dir` unless the user asks). Flag if a fallback is wanted.

## Decisions

- **Decision (User Accepted):** Diff `+/-` is omitted entirely when both `total_lines_added` and `total_lines_removed` are zero (render only when there are changes). **Reason:** Resolves planner risk R2/A5 — keeps the panel clean; `+0/-0` is noise. The branch token itself still renders; only the `+N/-N` suffix is suppressed at zero.
- **Decision (User Accepted):** Cost section is omitted when `total_cost_usd` is zero or absent (render only when cost > 0). **Reason:** Resolves R3/A6 — cleaner panel at session start before the first API call. Implies distinguishing zero/absent (pointer type) per the plan's null-handling approach.
- **Decision (User Accepted):** Capacity abbreviation uses general k/M abbreviation (≥1_000_000 → `M`, ≥1_000 → `k`), so sizes beyond the pinned `1M`/`200k` (e.g. `128k`, `500k`) render sensibly. **Reason:** Resolves R5/A2.
- **Decision (User Accepted):** The project name IS in scope (user correction — explicitly NOT a follow-up). Sourced from `workspace.project_dir` (basename), styled dim grey (ANSI `\033[2m…\033[0m`, matching GSD). **Placement: it joins the VCS group** — the panel's branch section becomes `<project> ⎎ <branch> +N/-N` (project name immediately before the branch icon; project + branch + diff render as a single ` │ `-delimited group). Degradation: empty `project_dir` → omit only the project name, branch+diff still render; no git branch → branch token degrades independently. **Reason:** User explicitly wants the project name, grouped with branch as shared VCS context. Supersedes the earlier (incorrect) deferral of R4/A7 — `project_dir` is now consumed, not dead code.
- **Decision (Orchestrator Accepted):** Phase 2 Step 2.1 — `Render(workdir string)` is kept as a thin wrapper over the new `RenderPanel(Payload)` entry point instead of widening `Render`'s signature. **Reason:** Behavior-preserving (workdir / `current_dir` resolution unchanged; `RenderPanel` currently emits only the plan section); avoids churning 13 existing `Render(root)` test call sites while giving Steps 2.2–2.4 a clean payload-carrying seam. Local beneficial drift, no scope or later-step impact.
- **Decision (Tech Debt):** Phase 2 review Suggestion #2 — `costSection` renders a sub-cent positive cost (`0 < cost < 0.005`) as `$0.00` via `$%.2f`. Within contract (A6 says "render only when > 0"; `0.004 > 0`). Deferred by user; could later omit when the *formatted* value rounds to `$0.00`. Not fixed in this task.
- **Decision (Tech Debt):** Phase 2 review Suggestion #3 — `abbrevCapacity` integer-truncates non-clean sizes (`1_500_000 → "1M"`, `1_900 → "1k"`). Within contract (A2 chose the simple k/M rule; all real context-window sizes — 200k, 1M — are integer-clean). Deferred by user. Not fixed in this task.

## Outcome
