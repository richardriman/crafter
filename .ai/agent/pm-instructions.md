# PM Instructions — Crafter

Refer to the standard ADOS lifecycle here: `doc/guides/change-lifecycle.md`.

## 1) Tracker Configuration

- Tracker type: GitHub Issues
- Repository: `richardriman/crafter`
- Access method: `gh` CLI
- Backlog source of truth: GitHub Issues

## 2) Workflow States Mapping (GitHub Labels)

GitHub Issues uses labels for workflow progression (no transition IDs).

| ADOS stage | GitHub label |
|---|---|
| Selected for delivery | `change` |
| Implementation started | `in-progress` |
| Awaiting review | `review` |
| Blocked | `blocked` |
| Delivered | `delivered` |

## 3) Label Taxonomy

Core delivery labels:
- `change`
- `in-progress`
- `review`
- `blocked`
- `delivered`

Issue-type labels currently present in repository:
- `bug`
- `enhancement`
- `documentation`
- `question`
- `help wanted`
- `good first issue`

## 4) Backlog Source of Truth

All active delivery work is tracked in GitHub Issues for `richardriman/crafter`.

## 5) Conventions

- workItemRef format: `#<issue-number>` (example: `#123`)
- optional cross-reference: `GH-<id>`
- branch naming: `feature|fix|refactor|chore|docs/GH-<id>-<slug>`

## 6) Issue Validation Checklist

Before implementation starts, validate that each issue has:
- clear problem statement
- explicit acceptance criteria
- known scope boundaries/non-goals
- context links (affected files/docs/related issues)
- verification expectations

If critical details are missing, ask blocking questions and apply `blocked` until clarified.

## 7) Definition of Ready (DoR)

A ticket is ready when:
- acceptance criteria are testable
- required context is attached
- ambiguity is resolved with owner
- priority and intent are clear

## 8) Priority & Selection Rules

When selecting among candidates:
1. blocked regressions/bugs with high user impact
2. items already in `in-progress`
3. `enhancement`
4. docs-only improvements

## 9) Quality Gate References

No additional mandatory quality gates were specified beyond the repository's existing test and review workflow.

## 10) PR/MR Workflow Customizations

- Merge authority: repository human owner
- No in-repo CODEOWNERS file currently maintained
- Treat owner approval/review as the final merge gate
