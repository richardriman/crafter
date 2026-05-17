# Flag Validation

`--auto` and `--fast` are mutually exclusive. If both flags are active (`auto: true` AND `fast: true` in frontmatter, or equivalent invocation context indicating both are set), produce a clear error and stop immediately — do not proceed to project resolution, resume detection, or any other workflow step:

> Error: `--auto` and `--fast` are mutually exclusive — pass at most one. `--auto` strictly supersedes `--fast` per `rules/do-workflow.md` → `### --auto`.

See `rules/do-workflow.md` → `### --auto (unattended orchestration)` for the canonical mutual-exclusion rule.
