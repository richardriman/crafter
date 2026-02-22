# State

## Current Focus

Improving existing commands — refining the do/debug/map-project/status workflows.

## Recent Changes

| Date | Change | Commit |
|---|---|---|
| 2026-02-22 | Added test suite for install.sh (pure Bash, zero dependencies) | TBD |
| 2026-02-22 | Added internal `/crafter:release` command for GitHub Releases with AI-generated notes | 312b391 |
| 2026-02-22 | Added remote install via curl one-liner (auto-detection, --version flag, tarball download) | 68e54ad |
| 2026-02-21 | Added MIT licence | 976706f |
| 2026-02-21 | Added automatic update check with version detection (VERSION file, 24h-cached GitHub Releases API, all commands) | 850e59a |
| 2026-02-20 | Added model configuration for subagent delegation (opus/sonnet/haiku per role) | f3de9d0 |
| 2026-02-20 | Added session workflow guidance (P2a) and task state persistence with decision records (P2b) | fdc881c |
| 2026-02-20 | Added review-fix loop to do-workflow (Step 6 with severity-based fix delegation, 3-iteration cap) | 4ba6c1d |
| 2026-02-20 | Split rules.md into per-concern fragments, optimized context window loading | 143ec46 |
| 2026-02-20 | Added Ideas section to STATE.md and template | ab13162 |
| 2026-02-20 | Enforced STATE.md and documentation update rules in all commands | ed48a88 |
| 2026-02-19 | Extracted install_to() to eliminate duplication in install.sh | 8e9fad5 |
| 2026-02-19 | Refactored install.sh to support --global and --local modes | c7bc499 |
| 2026-02-19 | Added meta-prompts and converted commands to orchestrator pattern | bdc4440 |
| 2026-02-19 | Swapped VERIFY and REVIEW order (verify before review) | a12c110 |
| 2026-02-19 | Created initial Crafter repository structure | c84a06d |

## Done

- [x] Initial repository structure (commands, rules, templates, README)
- [x] Orchestrator/subagent architecture with five meta-prompts
- [x] Installer script with global and local modes
- [x] Refactored install.sh to use shared install_to() function
- [x] Design philosophy documentation
- [x] BMAD integration documentation
- [x] Context window optimization (rule fragment split, selective .planning/ loading)
- [x] Update STATE.md after completing and approving each step
- [x] Update documentation before committing
- [x] **P1** Add review-fix loop to do-workflow (review findings → Implementer fix → Verify → Review → repeat or proceed)
- [x] **P2** Session-level workflow guidance (post-completion `/clear` + `/crafter:do` reminder, soft nudge when working outside crafter)
- [x] **P2** Task state persistence with decision records — `.planning/tasks/YYYYMMDD-<topic>.md` tracked in git; active task state (resume) during work, permanent decision record after completion
- [x] **P3** Model configuration for subagent delegation — assign model tiers (opus/sonnet/haiku) per role in `rules/delegation.md`
- [x] **P3** Detect new Crafter releases and notify the user — VERSION file, `rules/update-check.md`, 24h-cached GitHub Releases API check in all commands
- [x] Remote install without cloning — curl one-liner with auto-detection, --version flag, tarball download from GitHub
- [x] Internal `/crafter:release` command for preparing GitHub Releases with AI-generated release notes (not distributed)
- [x] Test suite for install.sh — pure Bash, no external dependencies

## Planned

- [ ] Optional project-level review rules — reviewer loads `.planning/review-rules.md` (if present) as additional context, allowing projects to define language-specific, framework-specific, or team-specific review criteria

## Ideas

- `/crafter:add-planned` — quick command for adding planned items to STATE.md

## Known Issues

## Notes

- Commands reference `~/.claude/crafter/...` (global) with fallback to `.claude/crafter/...` (local)
