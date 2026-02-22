# Task: Automate GitHub Releases

## Metadata
- **Date:** 2026-02-22
- **Branch:** main
- **Status:** active
- **Scope:** Medium

## Request
automatizace GitHub releases

## Plan
- [ ] Create `.claude/commands/release.md` — internal-only slash command that: reads VERSION, collects commits since last tag, generates structured release notes, presents for approval, creates git tag + GitHub Release via `gh release create`
- [ ] Update `.planning/STATE.md` — move `/crafter:release` from Ideas to Done
- [ ] Update `.planning/ARCHITECTURE.md` — add `.claude/commands/release.md` to directory tree (noted as internal-only)

## Decisions
<!-- Key decisions made during the workflow, in chronological order -->

## Outcome
<!-- Filled on completion: what was actually done, commit SHA(s), any deviations from plan -->
