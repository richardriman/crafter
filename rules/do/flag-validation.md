# Flag Validation

`crafter-do` accepts exactly two flags: **`--ext`** and **`--auto`** (frontmatter: `ext: true`, `auto: true`). Both default to off. They are **independent** — any combination is valid.

- **`--ext`** — enable extension-skill discovery and the pre-spawn extension checks. Without it, `rules/do/extension-skills.md` is never read and no discovery scan runs.
- **`--auto`** — fully unattended orchestration. See `rules/do-workflow.md` → `### --auto (unattended orchestration)`.

**`--fast` was removed.** If it is passed (or `fast: true` appears in frontmatter), produce this error and stop immediately — do not proceed to project resolution or any other workflow step:

> Error: `--fast` was removed; Minor findings now auto-proceed. Minor and Suggestion findings are recorded as `Decision (Tech Debt — auto-recorded)` entries and the commit continues without waiting, so silence-as-approval no longer has a purpose. Re-run without the flag.

`--project <path>` is not a skill flag — it is consumed by Project Resolution (see `skills/crafter-do/SKILL.md` → **Project Resolution**), not by this check.
