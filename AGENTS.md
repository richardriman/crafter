# AGENTS.md

Global repository policy for all coding agents.

## Highest-priority content policy

Do **not** sign outputs with model identity in commits, pull requests, or releases.

Forbidden in commit messages, PR text, and release text:

- any `Co-authored-by` trailer
- any `Signed-off-by` trailer
- any model/vendor attribution ("Claude", "Copilot", "GPT", "Gemini", etc.)
- any "AI-generated" signature line or footer

## Enforcement rules

1. Before finalizing commit text, PR text, or release notes, strip model signatures/attributions.
2. Use concise, project-focused wording only.
3. If an automation path conflicts with this policy, prefer this policy for generated text content.

## Project-specific working agreement (Crafter)

### Repository map

- `skills/` — canonical Crafter workflow definitions (source of truth)
- `agents/` — role-specific native agent definitions
- `rules/` — modular workflow rule files
- `cli/` — Go utility binary (`crafter`)
- `templates/` — `.crafter/` initialization templates
- `tests/` — installer and regression tests
- `docs/` — supplementary project documentation

### Source-of-truth editing rules

- Modify repository source files only.
- Do **not** modify installed runtime copies in `~/.claude/...`.
- Keep changes surgical and scoped to the requested task.
- Do not refactor unrelated code while completing a change.

### Delivery conventions

- Backlog source of truth: GitHub Issues (`richardriman/crafter`).
- Workflow labels: `change`, `in-progress`, `review`, `blocked`, `delivered`.
- Work item references: `#<issue-number>` and optional `GH-<id>`.
- Branch naming: `feature|fix|refactor|chore|docs/GH-<id>-<slug>`.
- PR merge authority: repository human owner (no in-repo CODEOWNERS file currently maintained).

### Validation hints

- CLI tests: run `make test` in `cli/`.
- Installer tests: run `tests/test_install.sh` from repository root.
- Keep documentation updates aligned with behavior changes in the same PR when applicable.
