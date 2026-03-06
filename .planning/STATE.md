# State

## Current Focus

Improving existing commands — refining the do/debug/map-project/status workflows.

## Recent Changes

| Date | Change | Commit |
|---|---|---|
| 2026-03-06 | Smarter /crafter:do entry logic — monorepo support (`--project` flag + auto-discovery), Grep-based resume detection with resume-intent words, guardrails against ignoring clear user input, `{PROJECT_PATH}` across all rule files | 47745a3 |
| 2026-03-06 | Review STOP gate — unmissable formatting for review findings | eb10a08 |
| 2026-02-27 | Install script cleans target directory before copying — removes stale files from previous versions on upgrade | 5c1d41a |
| 2026-02-27 | Planner writes full plan to task file (draft/approved lifecycle), richer summary, stricter review gate, resume handles all plan states | 7722b9f |
| 2026-02-26 | Released v0.5.0 — review workflow always waits for user, detailed table output, update check and release fixes | [v0.5.0](https://github.com/richardriman/crafter/releases/tag/v0.5.0) |
| 2026-02-26 | Released v0.4.1 — agents prefer native tools over Bash, reducing permission prompts | [v0.4.1](https://github.com/richardriman/crafter/releases/tag/v0.4.1) |
| 2026-02-24 | Released v0.4.0 | [v0.4.0](https://github.com/richardriman/crafter/releases/tag/v0.4.0) |
| 2026-02-24 | Migrated meta-prompts to native agents, fixed all 11 known workflow issues, updated install + tests + docs | aec78c6 |
| 2026-02-23 | Moved update check to SessionStart hook (replaced inline rule with hooks/crafter-check-update.js) | 20acf60 |
| 2026-02-22 | Added test suite for install.sh (pure Bash, zero dependencies) | 8c302d3 |
| 2026-02-22 | Added internal `/crafter:release` command for GitHub Releases with AI-generated notes | 312b391 |
| 2026-02-22 | Added remote install via curl one-liner (auto-detection, --version flag, tarball download) | 68e54ad |
| 2026-02-21 | Added MIT licence | 976706f |

## Planned

- [ ] Optional project-level review rules — reviewer loads `.planning/review-rules.md` (if present) as additional context, allowing projects to define language-specific, framework-specific, or team-specific review criteria

## Ideas

- `/crafter:add-planned` — quick command for adding planned items to STATE.md

## Known Issues

None currently tracked.
