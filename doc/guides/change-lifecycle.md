# Change Lifecycle

End-to-end flow for delivering a change in this repository.

## Stage overview

```
Backlog → change → in-progress → review → delivered
                        ↕
                     blocked
```

GitHub Issues labels drive stage transitions. There are no transition IDs — apply or remove labels directly.

## Stages

### 1. Backlog

Issue exists but is not yet selected for delivery.

- Issue has no delivery label yet, or sits in the general backlog.
- Validate against the Definition of Ready before selecting (see below).

### 2. `change` — Selected for delivery

Issue is prioritised and ready to start.

- Apply label: `change`
- Issue must have: problem statement, acceptance criteria, scope/non-goals, context links, verification plan.
- If any of these are missing, apply `blocked` and ask the owner before proceeding.

### 3. `in-progress` — Implementation started

Active work has begun.

- Apply label: `in-progress` (keep `change`)
- Create a branch: `feature|fix|refactor|chore|docs/GH-<id>-<slug>`
- Work in commits following [Conventional Commits](https://www.conventionalcommits.org/):
  `feat`, `fix`, `refactor`, `docs`, `chore`, `test`
- Keep scope surgical — no unrelated changes.

### 4. `review` — PR open, awaiting merge

Implementation is complete; PR is open for owner review.

- Apply label: `review` (remove `in-progress`)
- PR body must include:
  - `Closes #<issue-number>` (auto-closes issue on merge)
  - concise summary of intent and impact
  - verification notes (what was checked and how)
  - any follow-ups or unresolved risks
- Merge authority: repository human owner.

### 5. `blocked`

Progress is blocked pending external input or clarification.

- Apply label: `blocked`, note the blocker in a comment.
- Remove `blocked` and resume appropriate stage once resolved.

### 6. `delivered` — Merged and closed

PR squash-merged into `main`; issue auto-closed via `Closes #<id>` in PR body.

- Branch deleted after merge.
- No separate `delivered` label step needed — issue closure is the signal.

## Definition of Ready

A ticket is ready to start when it has:

- [ ] clear problem statement
- [ ] testable acceptance criteria
- [ ] known scope and non-goals
- [ ] context links (affected files, related issues)
- [ ] verification expectations

## Validation before merge

| Area | Command |
|---|---|
| CLI tests | `make test` (run inside `cli/`) |
| Installer tests | `tests/test_install.sh` (run from repo root) |
| Docs alignment | verify any behaviour change is reflected in the same PR |

## Priority order for selecting work

1. Blocked regressions / bugs with high user impact
2. Items already `in-progress`
3. Enhancements
4. Docs-only improvements

## References

- `AGENTS.md` — content policy and delivery conventions
- `.ai/agent/pm-instructions.md` — tracker configuration and label taxonomy
- `.ai/agent/pr-instructions.md` — PR platform integration
- `doc/guides/branch-protection.md` — `main` branch protection settings
