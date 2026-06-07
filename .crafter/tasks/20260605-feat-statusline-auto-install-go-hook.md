# Task: Statusline auto-install (drop opt-in) + node-free check-update hook + tty-safe renderer

## Metadata
- **Date:** 2026-06-05
- **Work branch:** feat/statusline-auto-install-go-hook
- **Status:** completed
- **Scope:** Large

## Request

Three related changes in the Crafter source. New framing: the Crafter statusline panel
is now a full-featured, standalone replacement for a generic external status line — so
it should install by default, not behind an opt-in flag. (See the wording constraint in
`## Decisions`.)

**1. Statusline installs automatically (drop `--with-statusline` as the gating flag).**
The default install (`install.sh`, both `--global` and `--local`) must wire the Crafter
statusline via the ALREADY-EXISTING smart-replace decision tree in `crafter install
statusline`:
- `statusLine` absent → set ours automatically.
- `statusLine` is ours (including an older version of ours) → update automatically.
- `statusLine` is foreign (GSD or any other tool) → on a real TTY, prompt whether to
  overwrite (keep `.bak` + echo the old command); when non-interactive (`curl | bash`,
  CI, no TTY) do NOT overwrite and print ready-to-paste guidance.

Remove `--with-statusline` as a control mechanism (statusline default ON). **Open
decision for discussion:** whether to remove the flag entirely or keep it as a
deprecated no-op, because existing `curl … | bash -s -- --with-statusline` invocations
would otherwise error on an unknown flag. Update `README.md` (the `--with-statusline`
section, ~lines 71–104) and `.crafter/ARCHITECTURE.md`. Invert the relevant
`tests/test_install.sh` cases: the current "default install (global+local) does NOT add
statusLine" expectations and Section G must reflect the new default-ON behavior.

**2. Check-update hook: node → Go.** Port `hooks/crafter-check-update.js` into a new Go
subcommand (`crafter check-update`): read VERSION (global `~/.claude/crafter/VERSION`
and project `.claude/crafter/VERSION`), read/write cache
`~/.claude/cache/crafter-update-check.json`, query GitHub `releases/latest` over HTTP
with a timeout (~5s), strictly-newer semver compare, 4h (14400s) freshness +
invalidation when the installed version changes, background refresh that does not block
session start, silent-fail (always exit 0, never break session start). `install.sh`
(`install_hook()`, ~lines 415–434): hook command becomes `"<crafter_bin>" check-update`
instead of `node "<hook_dest>"`; remove the copy of `hooks/crafter-check-update.js` and
the JS file itself. This removes the hook's node dependency entirely — also fixing the
case where bare `node` is not on PATH (the user's `node` is only a mise shim). The
existing `crafter update` subcommand lives in `cli/cmd/update.go` — share what makes
sense.

**3. Statusline tty-safe.** `cli/cmd/statusline.go:55` blindly does
`io.ReadAll(os.Stdin)` → when run manually in a terminal (stdin = tty) it blocks until
Ctrl-D ("hangs"). Detect a terminal (e.g. `term.IsTerminal(int(os.Stdin.Fd()))`) and
either skip reading stdin (render with an empty payload) or print a short hint; never
block. Under Claude Code (stdin = pipe carrying JSON) behavior is unchanged.

### Context / grounding
- Builds on `.crafter/tasks/20260601-plan-progress-statusline.md`,
  `20260602-feat-statusline-full-panel.md`, `20260603-feat-statusline-install-replace.md`
  (the last one introduced the decision tree + node→Go for `install.sh` settings editing;
  this task goes further: drops the opt-in, and also moves the check-update hook node→Go
  and makes the renderer tty-safe).
- Go via mise: `cd cli && mise exec -- go build/test/vet ./...`. Install tests:
  `bash tests/test_install.sh`, `bash -n install.sh`.
- Respect `CLAUDE.md`: NO AI/model signatures in commits, PRs, or release notes.
- Environment rule: work only with repo source (`install.sh`, `cli/`, `hooks/`,
  `templates/`, `rules/`), NEVER edit the installed copy `~/.claude/crafter/`.
- Vertical phases, each GREEN and independently committable.

## Plan

**Plan status:** approved

### Approach

Three largely independent changes, sliced into three vertical phases ordered so each
lands green and independently committable, and so any new Go behavior is proven by Go
unit tests before `install.sh` is rewired to depend on it:

1. **Phase 1 — tty-safe renderer.** Smallest, fully isolated change in
   `cli/cmd/statusline.go`. Make the command detect a terminal on stdin and skip the
   blocking `io.ReadAll(os.Stdin)` so a manual terminal invocation never hangs; under
   Claude Code (stdin is a pipe) behavior stays byte-identical.
2. **Phase 2 — node-free `crafter check-update` + install.sh rewire.** Port the JS
   SessionStart hook into a new Go subcommand, prove it with Go tests, then point
   `install.sh install_hook()` at `"<crafter_bin>" check-update`, stop copying the JS
   file, and delete the JS source. After this the hook has zero node dependency.
3. **Phase 3 — statusline installs by default.** Flip the statusline reconcile from
   opt-in (`--with-statusline`) to default-on in both install modes, reusing the
   existing decision tree unchanged. Invert the install tests, and update README +
   ARCHITECTURE prose using neutral, behavior-focused wording per the accepted Decision.

The decision tree, classifier, `.bak` recovery, and TTY-prompt logic in `install.sh`
`install_statusline()` already exist and are NOT re-implemented — Phase 3 only changes
WHEN the reconcile runs.

### Why this matters

Today the statusline is opt-in behind a flag, and the update-check hook depends on a
bare `node` on PATH (which fails when `node` is only a version-manager shim). Running
the renderer manually in a terminal hangs on stdin. This task makes the statusline a
default part of every install, removes the node dependency from the hook entirely, and
makes the renderer safe to run by hand — while preserving the existing silent-fail,
non-blocking, never-overwrite-foreign-without-asking posture.

### Assumptions / interpretations

- **A1.** The smart-replace decision tree (classifier, `--classify`, `--on-foreign`,
  `.bak`, foreign-keep guidance, TTY prompt in `install.sh`) is already correct and
  shipped (PR #42). Phase 3 reuses it verbatim and only removes the opt-in gating.
- **A2 (OPEN — needs user decision, see Risks).** Whether to (a) remove
  `--with-statusline` entirely or (b) keep it as a deprecated no-op. **Recommendation:
  (b) keep it as a deprecated no-op** that prints a one-line deprecation note and does
  nothing else, because `install.sh`'s arg parser currently treats any unknown flag as a
  hard error (`exit 1` with usage), so existing pinned one-liners
  (`curl … | bash -s -- --with-statusline`) and any CI that passes the flag would break
  on upgrade. A no-op preserves backward compatibility for one release cycle. The plan
  below is written for option (b); if the user picks (a), Step 3.1 drops the no-op branch
  and removes the flag from arg parsing + usage instead.
- **A3.** The Go `check-update` port keeps the cache contract byte-compatible with the JS
  hook: same cache file path (`~/.claude/cache/crafter-update-check.json`), same JSON
  field names (`update_available`, `installed`, `latest`, `checked`), same 14400s window,
  same notice text and update one-liner, same GitHub endpoint and `v`-strip + strict
  semver compare. This lets a mixed-version cache (written by the old JS, read by new Go,
  or vice versa) interoperate during upgrades.
- **A4.** "Non-blocking background refresh" is satisfied by spawning a detached
  self-invocation (e.g. `crafter check-update --refresh`) that does the GitHub fetch and
  cache rewrite, mirroring the JS `spawn(..., {detached:true}); child.unref()` pattern.
  The exact mechanism is the Implementer's choice as long as session start is never
  delayed by the network call. The synchronous foreground path only reads the cache and
  prints the notice.
- **A5.** Tty detection uses `os.Stdin.Stat()` + `(mode & os.ModeCharDevice) != 0`
  (stdlib only, no new dependency), preferred over `golang.org/x/term`, since the stack is
  cobra-only today and this keeps `go.mod` unchanged. The Implementer may choose
  `golang.org/x/term` only if it proves materially simpler — but adding a dependency must
  be justified.
- **A6.** Install tests are auto-discovered (`declare -F | grep '^test_'`), so inverting
  expectations means editing the existing `test_*` function bodies / renaming them, not a
  registration list. The real-binary harness (`_make_real_go_bin_dir` /
  `_build_real_crafter_bin`) is reused for any test that must observe a real settings.json
  mutation; the echo-only fake-go shim cannot mutate JSON.
- **A7.** When statusline becomes default-on and the crafter binary is absent (Go missing,
  build failed), the existing binary-absent posture is kept: print the warning and exit 0
  without writing statusLine — the install must not start failing just because statusline
  moved to default.

### Non-goals

- Do NOT re-implement or change the decision-tree behavior, classifier, `.bak` scheme, or
  foreign-keep guidance — those ship as-is.
- Do NOT change the rendered statusline panel content/format, the `crafter statusline`
  output under a real Claude Code pipe, or any `cli/internal/statusline` logic.
- Do NOT change `crafter update`'s behavior or flags. Sharing helpers with it is allowed
  but optional and must not alter its output.
- Do NOT change the hook's cache contract, notice text, freshness window, GitHub endpoint,
  or semver semantics (port behavior 1:1; A3).
- Do NOT bump the version, write release notes, or open a PR as part of these phases
  (those are post-task steps owned by the orchestrator).
- Do NOT touch the installed copy `~/.claude/crafter/` — repo source only.
- Do NOT add comparative/dismissive language about other tools in any prose, and do NOT
  justify default-on by reference to any specific external tool (accepted Decision).

### Relevant areas

- `cli/cmd/statusline.go` — the `io.ReadAll(os.Stdin)` line; tty detection goes here.
- `cli/cmd/statusline_test.go` — existing `withStdin` helper + never-breaks guard; add the
  tty-skip case here.
- `cli/cmd/` — add a new `check_update.go` (new subcommand registered on `rootCmd`); follow
  the grouped/standalone command precedent (`statusline.go`, `update.go`, `buffer.go`).
- `cli/cmd/update.go` — existing `crafter update`: reuse shared helpers (version-file read,
  GitHub `releases/latest` fetch, semver compare) where it reduces duplication. Note: there
  is currently NO shared version-read / GitHub-fetch / semver helper in Go — they live
  inline in update.go (installer fetch) and in the JS hook. The Implementer decides whether
  to extract a small shared internal package or keep it local to the new command.
- `cli/internal/` — if helpers are extracted, a small new internal package is the natural
  home (mirrors `claudesettings`, `statusline`, `prbody`). Optional.
- `hooks/crafter-check-update.js` — the port SOURCE; deleted at the end of Phase 2.
- `install.sh` — `install_hook()` (~415–434: drop JS copy, change `hook_cmd` to
  `"<crafter_bin>" check-update`); arg parsing for `--with-statusline` (~20, 557–560);
  `usage()` (~57, 63–64); `install_global()` (~501–503) and `install_local()` (~519–521)
  statusline call sites; `WITH_STATUSLINE` var (~20).
- `tests/test_install.sh` — Section E (hook tests ~1010–1100), Section G (statusline
  ~1300–1690), Section H (foreign paths ~1704–1949), and the syntax check (~1954).
- `README.md` — Statusline section (~71–105), update commands block (~58–69).
- `.crafter/ARCHITECTURE.md` — `### Dual Installation Model` (~117–121), the CLI subcommand
  list (~127–137), the `hooks/` line (~20–21).

---

### Phase 1 — Statusline renderer is tty-safe

**Phase outcome:** `crafter statusline` run manually in a terminal returns promptly
(renders a degraded panel from the cwd, or prints nothing/a short hint) instead of
blocking on stdin; under a pipe (Claude Code) the output is byte-identical to today.

- [x] **Step 1.1 — Detect a terminal on stdin and skip the blocking read.**
  - Outcome: in `runStatusline`, when stdin is a character device (tty) the command does
    NOT call `io.ReadAll(os.Stdin)`; it proceeds with an empty payload (so `RenderPanel`
    degrades to whatever the cwd yields) and exits 0 without hanging. When stdin is a pipe
    or file, the existing read + JSON decode + render path runs unchanged.
  - Scope boundary: only `cli/cmd/statusline.go` (the read/branch). No change to
    `cli/internal/statusline`, no change to the rendered panel format.
  - Non-goals: do not add interactive prompts, flags, or a new dependency (A5); do not
    alter the pipe path's bytes.
  - Simplicity constraint: a single tty check via `os.Stdin.Stat()` + `ModeCharDevice`
    (stdlib only). No new package, no `go.mod` change.
  - Drift criteria: drifting if the pipe-path output changes for any payload; if a new
    dependency is added without justification; if the command can still block on a tty; if
    behavior in `cli/internal/statusline` changes.
  - Verification evidence: `cd cli && mise exec -- go build ./... && mise exec -- go vet
    ./... && mise exec -- go test ./...` all green. Manual: running the built binary in a
    terminal with no piped input returns immediately (does not wait for Ctrl-D).
  - Stop conditions: stop once the tty branch is in place and tests pass; do not also
    refactor unrelated parts of the command.

- [x] **Step 1.2 — Add a Go test proving the tty path does not block and the pipe path is
      unchanged.**
  - Outcome: a test in `statusline_test.go` asserts that with a tty-like stdin (or the
    tty-detection branch taken) `runStatusline` returns nil without reading to EOF, and the
    existing pipe-fed cases still pass. The existing `TestRunStatusline_NeverBreaksStatusBar`
    and decode tests continue to pass unchanged.
  - Scope boundary: `cli/cmd/statusline_test.go` only.
  - Non-goals: no production changes here; no rewrite of existing test cases beyond what is
    needed to add the tty case.
  - Simplicity constraint: reuse the existing `withStdin` helper pattern; keep the new test
    focused on the no-block guarantee.
  - Drift criteria: drifting if existing tests are weakened or deleted, or if the test
    depends on real terminal allocation (must be deterministic in CI).
  - Verification evidence: `mise exec -- go test ./cmd/...` green, including the new case.
  - Stop conditions: stop once the no-block behavior is covered by a green test.

- [x] **Phase 1 verification:** `cd cli && mise exec -- go build ./... && mise exec -- go
  vet ./... && mise exec -- go test ./...` all green; manual terminal run of the built
  binary returns without hanging; piped JSON still renders the full panel identically.
- [x] **Phase 1 review:** Reviewer confirms the change is confined to the renderer command,
  no dependency was added without justification, the pipe path is byte-identical, and the
  tty path cannot block. Green-commit on clean review.

---

### Phase 2 — Node-free `crafter check-update` hook

**Phase outcome:** a new `crafter check-update` subcommand reproduces the JS hook's
behavior (silent-fail, exit 0, non-blocking background refresh, cache-compatible),
`install.sh` registers `"<crafter_bin>" check-update` as the SessionStart hook instead of
`node "<hook_dest>"`, the JS file is no longer copied and is deleted from the repo, and the
hook has zero node dependency. The install settings carry the new command and existing
SessionStart entries are preserved verbatim (idempotency unchanged).

- [x] **Step 2.1 — Implement `crafter check-update` (foreground notice + detached refresh)
      with Go unit tests.**
  - Outcome: new `cli/cmd/check_update.go` registers a `check-update` subcommand on
    `rootCmd`. Foreground behavior: resolve installed VERSION (global
    `~/.claude/crafter/VERSION`, else project `<cwd>/.claude/crafter/VERSION`); if neither
    exists, exit 0 silently; read the cache and, when `update_available===true` AND cached
    `installed` matches the installed version, print the update notice to stdout (text and
    update one-liner identical to the JS, A3). Background behavior: a detached
    self-invocation (A4, e.g. a hidden `--refresh` flag) respects the 14400s freshness
    window + invalidation when installed version changed, fetches GitHub
    `releases/latest` with a ~5s timeout, strips a leading `v`, does a strictly-newer
    semver compare, and rewrites the cache. Every path silently swallows errors and exits 0.
    Where it reduces duplication, reuse/extract shared helpers with `update.go` (version
    read, GitHub fetch, semver) — optional per A3/relevant-areas.
  - Scope boundary: new `cli/cmd/check_update.go` (+ optional small `cli/internal/...`
    helper package) and its tests; `root.go` registration happens via the file's `init()`
    (no edit to root.go needed). No change to `install.sh` yet — this step lands green
    purely as Go.
  - Non-goals: do not change `crafter update`; do not change the cache contract, notice
    text, freshness window, endpoint, or semver semantics; do not make the foreground path
    perform the network call; do not add a `node` dependency anywhere.
  - Simplicity constraint: prefer `net/http` (already used in update.go) for the fetch and
    stdlib for semver; spawn the detached child via `os/exec` mirroring the JS detach. Keep
    the command silent-fail with a single top-level "never return a non-zero/never panic"
    discipline.
  - Drift criteria: drifting if any foreground path can block on the network; if any path
    can exit non-zero or panic on missing/garbled cache, missing VERSION, or network
    failure; if the cache JSON shape or notice text diverges from the JS; if `crafter
    update` behavior changes.
  - Verification evidence: Go unit tests covering: no-VERSION → silent exit 0 + no notice;
    cache says update available + installed matches → notice printed; cache stale/missing
    → no notice on the foreground path; semver compare correct (equal/older/newer, including
    `v`-prefix strip and differing component counts); cache read/parse of garbled JSON →
    no panic, exit 0. `cd cli && mise exec -- go build/vet/test ./...` green.
  - Stop conditions: stop once foreground + refresh behavior is implemented and covered by
    green Go tests; do not also rewire install.sh in this step.

- [x] **Step 2.2 — Rewire `install.sh install_hook()` to the Go subcommand and drop the JS
      copy.**
  - Outcome: `install_hook()` registers `"<crafter_bin>" check-update` (no `node`), stops
    copying `hooks/crafter-check-update.js`, and no longer references `hook_dest`/the hooks
    dir for the JS file. The binary-absent posture is preserved (warn + skip registration,
    exit 0). The registration stays idempotent (same `crafter install hook` path). `bash -n
    install.sh` passes.
  - Scope boundary: `install.sh` `install_hook()` only (and the `hooks_dir`/`hook_dest`
    locals it owns). Do not touch statusline call sites (Phase 3) or arg parsing.
  - Non-goals: do not change `crafter install hook`'s Go logic; do not change settings
    schema; do not remove the `~/.claude/hooks` directory creation if other consumers rely
    on it (verify — if only the JS file used it, removing the copy is enough).
  - Simplicity constraint: minimal diff — swap the command string, remove the `cp` of the
    JS file. Keep the existing warn-and-skip-when-binary-missing guard.
  - Drift criteria: drifting if the hook command is anything other than the crafter
    binary + `check-update`; if `node` is still referenced anywhere in the hook path; if
    idempotency or the binary-absent posture regresses; if `bash -n` fails.
  - Verification evidence: `bash -n install.sh` clean; the E-series hook tests pass after
    Step 2.3 updates them; manual read confirms no `node`/`.js` reference remains in
    `install_hook()`.
  - Stop conditions: stop once `install_hook()` registers the Go command and the JS copy is
    gone.

- [x] **Step 2.3 — Update install tests for the node-free hook and delete the JS file.**
  - Outcome: delete `hooks/crafter-check-update.js`. Update Section E tests that asserted
    the JS file is copied / the settings command contains `crafter-check-update.js`
    (`test_hook_source_file_exists_in_repo`, `test_global_installs_hook_file`,
    `test_local_installs_hook_file`, `test_global_registers_hook_in_settings`,
    `test_local_registers_hook_in_settings`, `test_hook_registration_is_idempotent`) so
    they assert the new reality: no JS file is copied, and the registered SessionStart
    command is the crafter binary + `check-update`. Tests use the real-binary harness where
    a real settings mutation must be observed (A6). `bash tests/test_install.sh` green.
  - Scope boundary: `tests/test_install.sh` Section E + the repo-file-existence test, and
    deleting the JS source. Do not touch Section G/H (Phase 3).
  - Non-goals: do not weaken idempotency coverage; do not remove the idempotency test, just
    update its needle from `crafter-check-update.js` to the new command marker.
  - Simplicity constraint: edit the existing function bodies / rename where the name encodes
    a now-false expectation; keep the harness usage consistent with current patterns.
  - Drift criteria: drifting if any test still expects the JS file or a `node` command; if
    idempotency is no longer asserted; if the suite is left red.
  - Verification evidence: `bash tests/test_install.sh` all pass; grep shows no remaining
    `crafter-check-update.js` or `node ` reference in `install.sh` or the active tests.
  - Stop conditions: stop once the JS file is deleted and the hook tests are green against
    the Go command.

- [x] **Phase 2 verification:** `cd cli && mise exec -- go build/vet/test ./...` green;
  `bash -n install.sh` clean; `bash tests/test_install.sh` all pass; no `node`/`.js`
  reference remains in the hook path (`install.sh install_hook()` and Section E tests);
  `hooks/crafter-check-update.js` is deleted.
- [x] **Phase 2 review:** Reviewer confirms the Go port preserves the cache contract,
  notice text, freshness window, and silent-fail/non-blocking posture; the background
  refresh cannot delay session start; install.sh no longer depends on node; and tests
  meaningfully cover the new behavior. Green-commit on clean review.

---

### Phase 3 — Statusline installs by default

**Phase outcome:** both `install_global` and `install_local` run the existing statusline
reconcile on every install (no flag required), with the same decision-tree behavior
(absent→set, ours→update, foreign→ask on TTY / keep + guidance when non-interactive). The
opt-in flag is handled per the accepted flag decision (A2). Install tests reflect
default-on; README and ARCHITECTURE describe the new default in neutral, behavior-focused
wording.

- [x] **Step 3.1 — Run statusline reconcile by default; handle the `--with-statusline`
      flag per the accepted decision.**
  - Outcome: `install_global` and `install_local` call `install_statusline ...`
    unconditionally (remove the `if [[ -n "$WITH_STATUSLINE" ]]` gate around both call
    sites). Per A2 recommendation (b): keep `--with-statusline` as a deprecated no-op in
    arg parsing that prints a one-line deprecation note and otherwise does nothing; remove
    the now-unused `WITH_STATUSLINE` gating semantics. The binary-absent posture is
    preserved (warn + skip statusLine write, exit 0; A7). The non-interactive / TTY / keep
    / overwrite logic in `install_statusline()` is reused unchanged. `bash -n install.sh`
    passes.
  - Scope boundary: `install.sh` arg parsing (~557–560), `usage()` (~63–64),
    `install_global()`/`install_local()` statusline call sites, and the `WITH_STATUSLINE`
    var. Do not modify `install_statusline()`'s decision logic.
  - Non-goals: do not change the decision tree, the prompt wording, `.bak` behavior, or the
    Go `install statusline` command; do not make a missing binary fail the install.
  - Simplicity constraint: smallest diff that makes statusline default-on and keeps the
    one-liner backward-compatible. If the user chooses A2 option (a) instead, this step
    removes the flag from arg parsing + usage entirely (no no-op branch).
  - Drift criteria: drifting if statusline is still gated behind the flag for either mode;
    if a pinned `--with-statusline` one-liner now errors (under option b); if the foreign
    non-interactive path stops being keep-by-default; if a missing binary now fails the
    install; if `install_statusline()`'s decision logic is altered.
  - Verification evidence: `bash -n install.sh` clean; Section G/H tests green after Step
    3.2; manual: a default `--global` install on a clean settings.json now writes a
    crafter statusLine; on a foreign seed non-interactively it keeps + prints guidance.
  - Stop conditions: stop once statusline runs by default and the flag is handled per the
    decision; do not also touch hook or renderer code.

- [x] **Step 3.2 — Invert/extend install tests for default-on statusline.**
  - Outcome: update Section G so the two "default install does NOT add statusLine"
    expectations (`test_global_default_install_no_statusline`,
    `test_local_default_install_no_statusline`) become "default install DOES add the
    crafter statusLine" (using the real-binary harness, A6). The existing
    `--with-statusline` G2–G7 and H1–H5 cases are reconciled with the new default: either
    re-pointed to assert default-on behavior, or (under A2 option b) updated so passing the
    deprecated flag still succeeds (exit 0, deprecation note, statusline applied via the
    same decision tree). Foreign-seed default install must still keep + print guidance
    non-interactively and never hang. If A2 option (a) is chosen, any test that passes
    `--with-statusline` is updated to drop the flag and assert the unknown-flag error path
    is gone. `bash tests/test_install.sh` green.
  - Scope boundary: `tests/test_install.sh` Sections G and H. Do not touch Section E
    (Phase 2) or unrelated sections.
  - Non-goals: do not weaken the foreign-overwrite / keep / guidance / `.bak` coverage; do
    not remove the binary-absent test (re-point it to the default path).
  - Simplicity constraint: edit existing function bodies and rename where a name encodes a
    now-false expectation; keep real-binary harness usage consistent with current G/H
    tests.
  - Drift criteria: drifting if any test still asserts default install has no statusLine;
    if foreign-keep/overwrite/guidance/`.bak`/no-hang coverage is lost; if the deprecated
    flag (option b) is asserted to error; if the suite is left red.
  - Verification evidence: `bash tests/test_install.sh` all pass; the inverted G1 cases now
    assert a crafter statusLine is present on a default install.
  - Stop conditions: stop once Sections G/H reflect default-on and pass.

- [x] **Step 3.3 — Update README and ARCHITECTURE for default-on statusline (neutral
      wording).**
  - Outcome: `README.md` Statusline section no longer instructs users to pass
    `--with-statusline`; it states the statusline installs by default and describes the
    decision-tree behavior (absent→set; ours→update; foreign→asked before overwrite on a
    real terminal, kept + guidance when non-interactive). Document the flag's fate per A2
    (deprecated no-op, or removed). `.crafter/ARCHITECTURE.md` `### Dual Installation Model`
    and the CLI subcommand list / `hooks/` line are updated to describe default-on
    statusline and the node-free `check-update` hook. ALL prose is neutral and
    behavior-focused: no comparative/dismissive language about other tools, and default-on
    is not justified by reference to any specific external tool (accepted Decision; the
    full guardrail lives in `rules/core.md`).
  - Scope boundary: `README.md` (Statusline + update blocks) and `.crafter/ARCHITECTURE.md`
    (Dual Installation Model, subcommand list, hooks line). Optionally add a Key Decision
    row in `.crafter/PROJECT.md` if the orchestrator wants the decision recorded — flag,
    don't assume.
  - Non-goals: no release notes, no version bump, no marketing comparisons, no mention of
    any specific competing tool as a reason.
  - Simplicity constraint: update only the sentences that are now false; keep the
    decision-tree description accurate to the shipped behavior.
  - Drift criteria: drifting if docs still tell users to opt in; if any comparative/
    dismissive phrasing appears; if a specific external tool is cited to justify default-on;
    if the docs describe behavior that does not match the code.
  - Verification evidence: re-read the edited sections; confirm no `--with-statusline`
    opt-in instruction remains (beyond a deprecation note under option b) and no
    tool-comparison wording is present; ARCHITECTURE reflects node-free hook + default-on
    statusline.
  - Stop conditions: stop once the two docs match the shipped behavior in neutral wording.

- [x] **Phase 3 verification:** `bash -n install.sh` clean; `bash tests/test_install.sh`
  all pass; `cd cli && mise exec -- go build/vet/test ./...` green (regression guard);
  README + ARCHITECTURE describe default-on statusline + node-free hook in neutral wording;
  a default install writes the crafter statusLine and a foreign non-interactive seed keeps
  + prints guidance without hanging.
- [x] **Phase 3 review:** Reviewer confirms statusline is default-on for both modes, the
  decision tree is unchanged, the flag is handled per the accepted decision without
  breaking pinned one-liners (option b), tests cover default-on + foreign paths, and all
  prose is neutral and behavior-focused per the accepted Decision. Green-commit on clean
  review.

### Alternatives considered

- **Flag handling (A2): remove vs deprecated no-op.** Removing `--with-statusline` is
  cleaner long-term but breaks pinned one-liners and CI that pass the flag, because the
  arg parser hard-errors on unknown flags. A deprecated no-op preserves compatibility for
  one cycle at the cost of a tiny bit of dead code. Recommended: keep as deprecated no-op
  (b); surfaced to the user as an open decision.
- **Tty detection (A5): stdlib `Stat()`+`ModeCharDevice` vs `golang.org/x/term`.** The
  stdlib approach needs no new dependency and fits the cobra-only stack; `x/term` is more
  explicit but adds a module dependency for no functional gain. Chose stdlib.
- **Background refresh mechanism (A4): detached self-invocation vs goroutine vs cron.** A
  goroutine cannot outlive the short-lived hook process reliably, and the SessionStart hook
  must not block. A detached child mirrors the proven JS pattern and guarantees session
  start is never delayed. Chose detached self-invocation.
- **Shared helpers with `update.go` (A3): extract a package vs keep local.** Extracting a
  small internal package reduces duplication of version-read/fetch/semver but risks
  touching `update.go`'s tested behavior. Left to the Implementer as an optional
  simplification, with the hard constraint that `crafter update` behavior must not change.
- **Phase ordering.** Renderer first (smallest, isolated), then hook (Go proven before
  install rewire), then default-on statusline (depends on nothing new but is the largest
  test/doc surface). This keeps each phase green and independently committable per the
  green-commit invariant; the three changes have no hard inter-dependency.

### Risks / unknowns / flags

- **OPEN DECISION (A2) — must ask the user before Phase 3:** remove `--with-statusline`
  entirely, or keep it as a deprecated no-op? Recommendation: deprecated no-op (preserves
  pinned one-liners; the arg parser currently hard-errors on unknown flags). The plan is
  written for the no-op; switching to removal changes only Step 3.1/3.2/3.3 details.
- **Cache contract compatibility during upgrades (A3).** If field names, the cache path, or
  the freshness window drift from the JS hook, a cache written by an old JS hook and read
  by the new Go command (or vice versa during a partial upgrade) could misbehave. Mitigated
  by porting the contract 1:1; flag if any deviation is necessary.
- **Background-refresh determinism in tests.** The detached GitHub fetch is hard to test
  deterministically; Go tests should cover the foreground notice + semver + cache logic
  directly and avoid asserting on real network calls. Flag if a test needs network access.
- **Default-on statusline on existing installs.** Users upgrading who already have a
  foreign statusLine will hit the foreign rung on their next install; non-interactively the
  behavior is keep + guidance (no surprise overwrite), which is the intended posture — but
  worth calling out in the PR description (orchestrator-owned, not in these phases).
- **`~/.claude/hooks` directory.** Confirm during Step 2.2 whether anything other than the
  deleted JS file relied on that directory before removing its creation; if unsure, leave
  the `mkdir -p` in place (harmless) and only remove the JS `cp`.

This contract protects the desired end state — statusline installed by default with the
existing never-overwrite-foreign-without-asking decision tree intact, a node-free
non-blocking update-check hook ported 1:1, and a renderer that never hangs in a terminal —
while keeping each phase green and independently committable and surfacing the one open
flag decision for the user. It is the right approach because it sequences the work so new
Go behavior is proven by unit tests before `install.sh` is rewired to depend on it, reuses
the already-shipped decision tree rather than re-implementing it, and confines all
user-facing prose to the accepted neutral wording.

## Decisions

- **Decision (User Accepted):** All commit messages, PR titles/bodies, release notes,
  README, ARCHITECTURE, and task prose must describe the statusline behavior in neutral,
  behavior-focused terms only: it installs by default; when a foreign/third-party
  statusLine is already present, the user is asked before it is overwritten. Do not use
  comparative or dismissive language about other tools, and do not justify default-on
  install by reference to any specific external tool. **Reason:** neutral, behavior-focused
  wording fully conveys what changed and keeps the project's public voice professional.
- **Decision (User Accepted):** Resolve open decision A2 → **remove `--with-statusline`
  entirely** (option a), not the deprecated no-op (option b). Step 3.1 must drop the
  no-op branch and instead remove the flag from `install.sh` arg parsing + usage text.
  **Reason:** the user prefers a clean removal; pinned `curl … | bash -s -- --with-statusline`
  one-liners that error on the unknown flag are an acceptable trade-off.
- **Decision (Orchestrator Accepted):** Phase 1 Step 1.2 — the tty-skip branch cannot be
  directly exercised under `go test` (no allocatable tty fd; `os.Pipe()` yields
  `ModeCharDevice=0`), and adding a production-only injectable-stat seam is out of scope.
  Accepted an inverted proof: a never-EOF pipe test demonstrates the pipe path DOES block
  in `io.ReadAll` (the exact behavior the tty guard prevents), plus code inspection of the
  guard. **Reason:** strongest deterministic coverage achievable within step scope, with no
  production seam and no real-terminal CI dependency.
- **Decision (Tech Debt):** Phase 1 review Suggestion #2 — `TestRunStatusline_PipePath_BlocksOnNeverEOF`
  uses an 80 ms timing deadline; structured so a too-short window only risks a false failure
  (never a false pass), but could theoretically flake on a heavily loaded CI runner. Not
  hardened (margin/comment) this task. **Reason:** non-blocking, correctness is sound.
- **Decision (Tech Debt):** Phase 1 review Suggestion #3 — the tty guard condition in
  `statusline.go` (`err != nil || (fi.Mode()&os.ModeCharDevice) == 0`) could read marginally
  better as a named `isTTY` local; left inline (idiomatic Go, comment explains intent).
  **Reason:** cosmetic.
- **Decision (Orchestrator Accepted):** Phase 2 Step 2.3 — added `MkdirAll(filepath.Dir(path),
  0o755)` to `claudesettings.Save()` in `cli/internal/claudesettings/store.go` (outside the
  step's declared tests-only scope). Step 2.2 removed `mkdir -p "$HOME/.claude/hooks"` from
  `install_hook()`, which had also created `$HOME/.claude/` as a side effect that the
  local-install settings write relied on; without it, `Save` failed for local installs and the
  hook test could not go green. **Reason:** local, beneficial, purely additive (no-op when the
  dir exists), changes no behavior for existing callers (`crafter update`, statusline install,
  all `claudesettings` tests green), and makes the Go layer self-sufficient rather than
  depending on the installer to pre-create the directory — the better in-spirit fix vs.
  restoring the shell `mkdir`. Verifier classified it as beneficial local drift; Go + install
  suites both green.
- **Decision (Orchestrator Accepted):** Phase 2 review Minor #1 — FIXED. A stray untracked
  `go build` artifact `cli/cli` was present and not gitignored. Deleted it and added an
  anchored `/cli/cli` rule to `.gitignore` (does not affect `cli/` source or `cli/cmd/`).
  **Reason:** prevents accidental commit of a build artifact; trivial, no behavior impact.
- **Decision (Tech Debt):** Phase 2 review Suggestion #2 — `check-update`'s cache write uses
  non-atomic `os.WriteFile` (`cli/cmd/check_update.go`), unlike the temp-file+rename used by
  `claudesettings`/buffer/skillbook stores. **Reason:** byte-faithful to the JS hook
  (`fs.writeFileSync`), non-blocking (the foreground reader silently swallows garbled JSON),
  so correctness is sound; atomic write deferred as hardening.
- **Decision (Tech Debt):** Phase 2 review Suggestion #3 — fixtures in `cli/cmd/install_test.go`
  still use the literal `node ".../crafter-check-update.js"` hookCmd string. **Reason:** those
  tests exercise the generic registration machinery (command injected via `installHookCommand`,
  not asserted against real installer output), so they pass and are correct; the stale fixture
  string is cosmetic and the file is outside this phase's contracted change set.
- **Decision (Orchestrator Accepted):** Phase 2 review Suggestion #4 — ARCHITECTURE.md drift
  (still lists `hooks/crafter-check-update.js` and a "24h" window) is DEFERRED to Phase 3
  Step 3.3, which rewrites ARCHITECTURE.md for the node-free hook + default-on statusline.
  **Reason:** Step 3.3 already owns this prose; fixing it now would duplicate the edit.

## Outcome

**Completed.** Three vertical phases delivered on branch `feat/statusline-auto-install-go-hook`, each green and independently committed:

- **Phase 1 — tty-safe renderer** (commit `1dd873f`): `crafter statusline` detects a non-pipe stdin and skips the blocking `io.ReadAll(os.Stdin)`, so a manual terminal run returns promptly (degraded panel) instead of hanging; the Claude Code pipe path is byte-identical. (Committed in a prior session.)
- **Phase 2 — node-free `crafter check-update` hook** (commit `54049f4`): new `cli/cmd/check_update.go` ports the deleted `hooks/crafter-check-update.js` 1:1 — foreground notice + detached `--refresh` background fetch, byte-compatible cache contract (4h/14400s window), silent-fail/non-blocking, 17 Go tests. `install.sh install_hook()` registers `"<crafter_bin>" check-update`; the hook no longer depends on `node` on PATH. `claudesettings.Save()` gained `MkdirAll` (accepted beneficial local drift). Review: 4 Minor/Suggestion — #1 (stray `cli/cli` artifact) fixed + gitignored; #2 (non-atomic cache write) and #3 (stale node fixtures) recorded as tech debt; #4 (ARCHITECTURE drift) deferred to and resolved in Phase 3 Step 3.3.
- **Phase 3 — statusline installs by default** (commit `5f68370`): both `--global`/`--local` run the existing smart-replace decision tree on every install; `--with-statusline` removed entirely (now hard-errors). Install tests inverted to default-on (real-binary harness) + flag-rejection test (suite 63→64); README + `.crafter/ARCHITECTURE.md` updated in neutral wording. Review: clean (zero findings).

**End-of-task housekeeping:** PROJECT.md Key Decisions updated (default-on + node→Go hook); STATE.md Recent Changes + Current Focus updated; ARCHITECTURE.md required no further change (current from Step 3.3). All Go + install suites green; `bash -n install.sh` clean.

**Follow-up (orchestrator-owned, not part of these phases):** open a PR for the branch. PR-description notes: upgraders with an existing foreign statusLine hit the foreign rung (non-interactive → keep + guidance, no surprise overwrite); pinned `--with-statusline` one-liners will error after upgrade (accepted trade-off).
