# Task: Move update check to SessionStart hook

## Metadata
- **Date:** 2026-02-23
- **Branch:** main
- **Status:** completed
- **Scope:** Medium

## Request
Replace the inline update-check mechanism (rules/update-check.md loaded by every command) with a Claude Code SessionStart hook, following the same pattern as GSD's `gsd-check-update.js`.

## Plan
- [x] Step 1: Create `hooks/crafter-check-update.js` — SessionStart hook script (sync notice from cache + background GitHub API check)
- [x] Step 2: Remove `update-check.md` references from all 4 command files (do.md, debug.md, status.md, map-project.md)
- [x] Step 3: Update `install.sh` — copy hook to `~/.claude/hooks/`, register in `~/.claude/settings.json` via `node -e`, remove `update-check.md` copy
- [x] Step 4: Update `tests/test_install.sh` — update expected files, add hook registration test
- [x] Step 5: Delete `rules/update-check.md`

Decisions:
- JSON merge in install.sh via `node -e` (Node.js always available with Claude Code)
- Hook always installed globally to `~/.claude/hooks/` regardless of install mode
- Cache location: `~/.claude/cache/crafter-update-check.json` (JSON, matching GSD pattern)

## Decisions
<!-- Key decisions made during the workflow, in chronological order -->

## Outcome
Commit `20acf60`. All 5 steps completed as planned. Review-fix loop triggered once on Step 3 (install.sh) — fixed nested `node -e` quoting via `process.env`, added `node` pre-check and `Array.isArray` guard. ARCHITECTURE.md tree ordering fixed as minor post-review cleanup.
