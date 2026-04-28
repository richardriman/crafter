# PR Instructions — Crafter

Refer to `doc/guides/pr-platform-integration.md` for platform integration guidance.

## Platform Configuration

- Platform: GitHub
- Host: `github.com`
- Repository: `richardriman/crafter`
- Access method: GitHub CLI (`gh`)
- Self-hosted instance: No

## Operational Policy

- PR workflow is managed through `gh` CLI.
- Merge decision is made by the repository human owner.
- There is no in-repo CODEOWNERS file currently maintained.

## Operations Reference

| Operation | Preferred command |
|---|---|
| Show repo status | `git status` |
| List open PRs | `gh pr list` |
| View PR details | `gh pr view <number>` |
| Create PR | `gh pr create --title "<title>" --body "<body>"` |
| Check PR checks | `gh pr checks <number>` |
| Update PR metadata | `gh pr edit <number> ...` |
| Merge PR | `gh pr merge <number> --squash` (or repo-preferred mode) |

## PR Content Expectations

Every PR should include:
- work item reference (`#<issue-number>` and optional `GH-<id>`)
- concise summary focused on intent and impact
- verification notes (what was tested and how)
- explicit callout of any follow-ups or unresolved risks

## Review & Merge Expectations

- Keep PR scope surgical and easy to review.
- Resolve all requested review changes before merge.
- Do not merge without human-owner approval.
