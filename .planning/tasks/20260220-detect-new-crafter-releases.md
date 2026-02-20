# Task: Detect new Crafter releases and notify the user

## Metadata
- **Date:** 2026-02-20
- **Branch:** main
- **Status:** active
- **Scope:** Medium

## Request
Detect new Crafter releases and notify the user in Claude Code.

## Plan
- [ ] Create `VERSION` file at repo root with `0.1.0`
- [ ] Create `rules/update-check.md` — rule for 24h-cached GitHub Releases API check, silent on failure, non-blocking notification
- [ ] Modify `install.sh` to copy VERSION to install destination
- [ ] Modify all 4 commands (do.md, debug.md, status.md, map-project.md) to load update-check rule
- [ ] Add `/crafter:release` (internal, not distributed) to Ideas in STATE.md

## Decisions
- **Decision:** Use GitHub Releases (not just git tags). **Reason:** Conventional, supports release notes.
- **Decision:** Initial version 0.1.0. **Reason:** Pre-1.0, signals early but usable.
- **Decision:** All four commands check for updates. **Reason:** Broader coverage, trivial to add.
- **Decision:** Cache always global (~/.claude/crafter/.update-cache). **Reason:** Per-user concern, avoids .gitignore issues for local installs.
- **Decision:** `/crafter:release` as follow-up, internal (not distributed). **Reason:** Maintainer tool, not user-facing.

## Outcome
<!-- Filled on completion: what was actually done, commit SHA(s), any deviations from plan -->
