# State

## Current Focus

Improving existing commands — refining the do/debug/map-project/status workflows.

## Recent Changes

| Date | Change | Commit |
|---|---|---|
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

## Planned

- [x] Update STATE.md after completing and approving each step
- [x] Update documentation before committing
## Ideas

- `/crafter:add-planned` — quick command for adding planned items to STATE.md

## Known Issues

- No test suite for install.sh

## Notes

- Commands reference `~/.claude/crafter/...` (global) with fallback to `.claude/crafter/...` (local)
