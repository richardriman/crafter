# Task: Fix install.sh stale-binary problem

## Metadata
- **Date:** 2026-05-19
- **Work branch:** fix/install-sh-stale-binary
- **Status:** completed
- **Scope:** Small

## Request

Fix install.sh stale-binary problem (revealed during v0.10.0 release):

**PROBLEM**
- `install.sh` `install_to` (around line 353) blindly copies `$SCRIPT_DIR/cli/bin/crafter` into the install destination if present, regardless of its age/version.
- `install.sh` `_download_cli_binary` (lines 157–161) then sees a pre-built binary at `$dest_bin` and returns early, so the matching release asset is never downloaded.
- Net effect: a local clone with a stale `cli/bin/crafter` (e.g. left over from a previous dev build) silently overrides every subsequent `./install.sh` run with an outdated binary. This bit us today — VERSION was 0.10.0 but the installed crafter binary was 0.7.1.

**SCOPE (recommendation A + B from prior discussion)**
1. Stop the blind copy in `install_to`: remove (or make conditional/version-checked) the `cp` of `$SCRIPT_DIR/cli/bin/crafter` to `$crafter_dest/bin/crafter`. The single source of truth for the installed binary should be `_download_cli_binary` (which prefers release assets and falls back to a source build).
2. Make `make release` in `cli/Makefile` remove any stale unsuffixed `cli/bin/crafter` before building, so a `make release` build can't leave a misleading artifact behind.

**CONSTRAINTS**
- Preserve existing `install.sh` behavior for the `curl | bash` remote path (already unaffected because tarball never contains `cli/bin/`).
- Preserve source-build fallback when `go` is available and no release asset exists.
- Keep VERSION 0.10.0 — this is a follow-up fix, not a re-release.
- No model attribution in commits/PR (CLAUDE.md rule).

**OUT OF SCOPE**
- Changing release tagging, GitHub Actions, or hook scripts.
- The `--force-binary` flag idea (option C) — defer.

**VERIFICATION**
- After change, running `./install.sh --global` from a clone that has a stale `cli/bin/crafter` must NOT install that stale binary; it must download the release asset matching VERSION (or fall back to source build).
- `make release` must not leave a stray `cli/bin/crafter` behind.

## Plan
**Plan status:** approved

### Goal
Make `_download_cli_binary` the single source of truth for the installed CLI binary, so a stale `cli/bin/crafter` in a developer's local clone can no longer silently override the release asset (or the source-build fallback). Also make `make release` in `cli/Makefile` self-cleaning so it cannot leave a misleading unsuffixed `cli/bin/crafter` artifact behind.

### Assumptions / interpretations
- The blind-copy block in `install_to()` (currently around lines 352–356 of `install.sh`) is the only place where `cli/bin/crafter` gets propagated into the install destination. `_download_cli_binary` always runs afterwards in both `--global` and `--local` paths, so removing the blind copy does not create an "unhandled" gap — the download/build path takes over.
- The early-return guard in `_download_cli_binary` (lines 157–161: `if [[ -x "$dest_bin" ]]; then return 0; fi`) is kept **as-is**. Rationale: once the blind copy is gone, nothing in `install_to` pre-populates `$dest_bin`. The guard still serves a legitimate purpose: idempotency on re-invocation (rare but harmless) and avoiding double-work if some future caller pre-populates the binary deliberately. Removing it would be a behavior change outside this fix's scope. The comment above it ("install_to may have already copied a local pre-built binary…") becomes stale and should be updated to reflect the new reality.
- `_download_cli_binary` already creates `$crafter_dest/bin/` via `mkdir -p "$(dirname "$dest_bin")"` (line 181). The `mkdir -p "$crafter_dest/bin"` at line 350 of `install_to` is therefore redundant after removing the copy block, but is harmless and not in scope to remove.
- For `cli/Makefile`: the simplest expression is `rm -f bin/crafter` as the first step of the `release` recipe (after the existing `mkdir -p bin`). The `clean` target is unaffected; the `build` target is unaffected.
- The existing test suite `tests/test_install.sh` contains two tests (`test_global_copies_local_cli_binary`, `test_local_copies_local_cli_binary`) whose **assertions remain valid** (they only assert that `$dest/bin/crafter` exists after install). However, the **mechanism** they're exercising changes: under the new behavior, the binary at the destination is produced by `_download_cli_binary`'s source-build fallback (since `REMOTE_MODE=0` in those tests and `go` is on PATH in CI) rather than by a copy. The assertions still hold — but the test names and inline comments become misleading. The Implementer should update the inline comments and may rename the tests for clarity; the assertions themselves should continue to pass without modification.

### Non-goals
- Do NOT introduce a `--force-binary` flag or any new CLI flag.
- Do NOT modify `_download_cli_binary`'s release-asset lookup, source-build fallback, or platform detection.
- Do NOT change `VERSION` (stays at 0.10.0).
- Do NOT modify GitHub Actions, release tagging, hooks, or the `curl | bash` remote install path.
- Do NOT touch the Makefile `build` or `clean` targets.
- Do NOT add a new test scenario specifically for "stale binary present" — the manual verification scenario suffices (deferred to a future hardening task if desired).

### Relevant areas
- `/Users/ret/dev/ai/crafter/install.sh` — `install_to()` blind-copy block (around line 353); `_download_cli_binary()` early-return guard and its comment (lines 157–161).
- `/Users/ret/dev/ai/crafter/cli/Makefile` — `release` target (lines 12–17).
- `/Users/ret/dev/ai/crafter/tests/test_install.sh` — comment updates only on the two tests at lines ~449 and ~538 (and possibly the matching `test_global_links_cli_to_home_local_bin` / `test_local_does_not_link_cli_to_home_local_bin` which also create fake `cli/bin/crafter` as a setup convenience). The Implementer should confirm assertions still pass after the install.sh change and adjust inline comments to no longer describe a "copy" mechanism.

### Phase 1 — Fix stale-binary install path (single phase, two steps)

**Phase outcome:** `./install.sh --global` (and `--local`) from a local clone with any stale `cli/bin/crafter` present installs the binary obtained from `_download_cli_binary` (release asset, then source-build fallback) — never the stale local artifact. `make -C cli release` cannot leave a stray unsuffixed `cli/bin/crafter` behind.

**Phase scope boundary:** Only the install.sh blind-copy removal, the stale-comment refresh on the kept early-return, the Makefile `release` cleanup line, and minor test-comment hygiene. No new flags, no new helper functions, no refactors.

**Phase non-goals:** No version-checking logic, no integrity checks, no new install modes, no behavior change for `_download_cli_binary` itself.

**Phase simplicity constraint:** Deletions and one-line additions only. No new control flow. No new variables. No new tests.

**Phase drift criteria (any one = drift):**
- A new flag, function, or env var is introduced in install.sh.
- `_download_cli_binary`'s body (other than the comment above the early-return) is modified.
- The Makefile `build` or `clean` targets are modified, or `release` adds anything beyond removing the stale unsuffixed binary.
- VERSION is touched.
- New test cases are added, or existing test assertions are weakened.

**Phase verification evidence:**
- `grep -n 'cli/bin/crafter' /Users/ret/dev/ai/crafter/install.sh` returns no hits inside `install_to()` (only allowed mention is the stale comment removal — i.e. ideally zero matches in install.sh after the fix).
- Manual reproduction: place an older real `crafter` binary (e.g. v0.7.1) at `cli/bin/crafter`, run `./install.sh --global`, then `~/.claude/crafter/bin/crafter --version` prints `0.10.0` (or, if release assets are unreachable and `go` is absent, the install warns and skips — but never reports 0.7.1).
- `rm -rf cli/bin && make -C cli release && ls cli/bin/` lists only the four `crafter-<os>-<arch>` artifacts and no bare `crafter`. Repeating with a pre-existing `cli/bin/crafter` (e.g. `touch cli/bin/crafter`) still ends with no bare `crafter` after `make release`.
- `bash tests/test_install.sh` (or whatever harness runner the repo uses) passes end-to-end with all existing tests green.

**Phase stop conditions:** Stop and re-plan if (a) removing the blind copy breaks any test whose assertion is about destination-file existence (not about copy mechanism) — that would mean `_download_cli_binary`'s fallback isn't firing in the test harness and needs investigation; (b) the Makefile change interacts with parallel-make or another rule in a way not visible from the current Makefile; (c) any in-tree code other than `install.sh` reads `cli/bin/crafter` directly (Implementer must grep to confirm none exists).

#### Steps

- [x] **Step 1 — Remove blind copy in `install.sh` and refresh stale comment**
  - **Outcome:** `install_to()` no longer copies `$SCRIPT_DIR/cli/bin/crafter` into the install destination under any condition. The comment above the early-return in `_download_cli_binary` is updated to no longer claim `install_to` may have pre-copied a binary.
  - **Scope boundary:** Delete the `if [[ -f "$SCRIPT_DIR/cli/bin/crafter" ]]; then … fi` block (and the comment above it). Keep `mkdir -p "$crafter_dest/bin"` so the destination directory still exists before `_download_cli_binary` runs (defensive; `_download_cli_binary` will also `mkdir -p`). Update the comment block above `_download_cli_binary`'s early-return to reflect that the guard now exists only for idempotency on re-invocation, not for pre-populated binaries.
  - **Non-goals:** Do not remove or alter the `if [[ -x "$dest_bin" ]]; then return 0; fi` early-return itself. Do not change any other line of `_download_cli_binary`. Do not touch `_link_cli_into_path`.
  - **Simplicity constraint:** Pure deletion plus a comment-text edit. No new conditionals.
  - **Drift criteria:** Any change to logic in `_download_cli_binary`; any additional flag/env-var; any change outside the `install_to` blind-copy block and the early-return comment.
  - **Verification evidence:** `grep -n 'cli/bin/crafter' /Users/ret/dev/ai/crafter/install.sh` returns no matches; `bash -n /Users/ret/dev/ai/crafter/install.sh` passes syntax check; reading lines 277–360 confirms only the deletion happened.
  - **Stop conditions:** Stop if the comment refresh requires more than one or two sentences, or if any other callsite in install.sh references `SCRIPT_DIR/cli/bin/crafter`.

- [x] **Step 2 — Make `cli/Makefile` `release` target clean the stale unsuffixed binary**
  - **Outcome:** Running `make -C cli release` removes any pre-existing `cli/bin/crafter` before producing the four suffixed release artifacts, so a `make release` build can never leave a misleading bare `crafter` binary behind in `cli/bin/`.
  - **Scope boundary:** Add a single `rm -f bin/crafter` line to the `release` recipe (after `mkdir -p bin`, before the first `go build`). The `build`, `test`, and `clean` targets are unchanged.
  - **Non-goals:** Do not change `RELEASE_LDFLAGS`, the suffixed target list, or the platform matrix. Do not introduce a phony dependency between `release` and `clean`.
  - **Simplicity constraint:** One new line in one recipe.
  - **Drift criteria:** Any change outside the `release` recipe; any change to other targets; any new variable.
  - **Verification evidence:** `touch cli/bin/crafter && make -C cli release && test ! -f cli/bin/crafter` succeeds; the four `cli/bin/crafter-<os>-<arch>` files all exist after the run.
  - **Stop conditions:** Stop if `make release` already (somehow) deletes the bare binary on this machine (would mean the bug repro requires a different precondition we haven't identified) or if `bin/crafter` is referenced by some other Make target downstream.

### Test hygiene (within Phase 1, no separate step)
After Steps 1–2, the Implementer should run `bash /Users/ret/dev/ai/crafter/tests/test_install.sh` (using the project's normal test invocation) and:
- Confirm all tests still pass.
- If `test_global_copies_local_cli_binary` / `test_local_copies_local_cli_binary` pass on their assertions but their **comments** describe a now-defunct "copy" mechanism, update only the inline comments (and optionally rename for clarity) — do NOT change their setup or assertions. If they fail, that's a drift signal: stop and report rather than weakening assertions.

### Alternatives considered
Small scope, but two alternatives were weighed and rejected:
- **Version-check the blind copy** (only copy if `cli/bin/crafter --version` matches `VERSION`): rejected because it adds runtime branching, depends on `--version` working on the stale binary, and still leaves two competing sources of truth. Removing the copy entirely is simpler and correct.
- **Remove the early-return in `_download_cli_binary` as well**: rejected because, once the blind copy is gone, nothing pre-populates `$dest_bin`, so the guard is a harmless idempotency net. Removing it would be a scope creep with no current benefit.

### Risks / unknowns / flags
- **Test harness behavior:** In CI/local test runs, `_download_cli_binary` will hit the source-build path (since `REMOTE_MODE=0`). This adds a real `go build` to two tests that previously only `cp`'d a fake script. Build time is small but non-zero; if the test environment lacks a working Go toolchain the relevant tests will fail. Flag for the Implementer: verify Go is available in the test environment before declaring drift.
- **Comment-only test updates:** If the Implementer feels the existing tests no longer match their names after Step 1, that's a naming concern — flag it in `## Decisions` rather than expanding scope here.
- **No new automated test for stale-binary scenario:** Coverage relies on manual verification. A future hardening task may add a dedicated regression test, but that's out of scope per the task constraints.

### Summary
This contract protects a small, surgical fix: kill the blind copy that lets a developer's stale local build poison `./install.sh`, and make `make release` self-clean so the same artifact can't accumulate. The single source of truth for the installed CLI binary becomes `_download_cli_binary`, which already prefers the matching release asset and falls back to a source build — exactly what users expect when they run `./install.sh --global` from a clone.

## Decisions

- **Decision (Orchestrator Accepted):** Phase verification verifier flagged a `{CRAFTER_HOME}` substitution regression and two deleted tests. Investigation showed the verifier compared against commit `f898dc9` on the unmerged branch `feat/crafter-do-progressive-loading`, not against `main`. Actual `git diff main` shows the three expected edits only (install.sh ±10/12, cli/Makefile +1, tests/test_install.sh ±35 lines all comment-only, zero deleted test functions). No regression. Proceeding to Step 6 review.

## Outcome

**Status:** completed on branch `fix/install-sh-stale-binary` (commit `b3efebd`).

**What was done:**

- `install.sh::install_to()` — removed the unconditional `cp "$SCRIPT_DIR/cli/bin/crafter" "$crafter_dest/bin/crafter"` block (and its preceding `# Local clone install: copy pre-built binary if it exists` comment). The `mkdir -p "$crafter_dest/bin"` line is kept so the destination directory exists before `_download_cli_binary` runs.
- `install.sh::_download_cli_binary()` — refreshed the now-stale comment above the early-return guard. The early-return itself (`if [[ -x "$dest_bin" ]]; then return 0; fi`) is unchanged and remains as an idempotency net.
- `cli/Makefile` — added `rm -f bin/crafter` to the `release` recipe (after `mkdir -p bin`, before the first `go build`). `build`, `test`, and `clean` targets unchanged.
- `tests/test_install.sh` — inline-comment-only updates on `test_global_copies_local_cli_binary` and `test_local_copies_local_cli_binary` to describe the new mechanism (binary now produced by `_download_cli_binary`'s source-build fallback path in these tests). Zero assertions modified; test function count unchanged (45 == 45 vs `main`).

**Net effect:** `_download_cli_binary` is now the single source of truth for the installed CLI binary (GitHub release asset → source-build fallback). A stale developer-local `cli/bin/crafter` can no longer silently override it. `make release` cannot leave a misleading unsuffixed `cli/bin/crafter` behind.

**Verification evidence captured during phase verification:**

- `grep -n 'cli/bin/crafter' install.sh` → zero matches.
- Manual stale-binary repro: fake `crafter version 0.7.1-fake` placed at `cli/bin/crafter`, `HOME=$(mktemp -d) bash install.sh --global` → installed binary reported `crafter version dev` (source-build), never `0.7.1-fake`.
- `touch cli/bin/crafter && make -C cli release && test ! -f cli/bin/crafter` → printed `CLEANED`; all four `cli/bin/crafter-<os>-<arch>` present.
- `bash tests/test_install.sh` → 45 passed, 0 failed.

**Deviations from plan:** None on the implementation side. One process note recorded under `## Decisions`: the phase-verification verifier initially flagged a `{CRAFTER_HOME}` regression that turned out to be a false positive (verifier had compared against the unmerged `feat/crafter-do-progressive-loading` branch instead of `main`).
