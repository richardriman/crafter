# State

## Current Focus

Improving existing commands — refining the do/debug/map-project/status workflows.

## Recent Changes

| Date | Change | Commit |
|---|---|---|
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

- Small scope bypasses task file creation — post-change completion may fail
- Medium scope "one step at a time" lacks granularity guidance for the Planner
- Small scope flow through Steps 4–6 unclear — orchestrator may skip Verify+Review
- Review-fix iteration cap ambiguous — "would exceed the 3rd" vs "cap at 3"
- Resume detection doesn't distinguish mid-Execute break from post-E→V→R session break
- ARCHITECTURE.md handling inconsistent — `do.md` passes it to Planner but Planner doesn't use it; post-change reads it despite orchestrator ban
- do-workflow.md REVIEW section partially duplicates do.md Step 6 (divergence risk)
- map-project.md uses local-install fallback annotations on rule paths but do.md, debug.md, status.md do not
