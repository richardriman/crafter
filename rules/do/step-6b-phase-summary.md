# Step 6b — Phase Summary and Auto-Commit

After the review loop closes clean (no Critical or Major findings remain), the orchestrator chooses an approval path based on the active flags.

### Approval paths

#### `--auto` branch (precedes paths 1–3)

When `--auto` is set (`auto: true` in frontmatter): the orchestrator does **not** surface a Phase Summary to the user and does **not** wait for any approval signal. Instead:

1. Record any Critical/Major findings the auto-fix loop cleared before the review closed clean as `Decision (Auto-Fixed): <severity> — <description>` entries in the task file's `## Decisions` section.
2. Record any remaining Minor/Suggestion findings from the final review as tech-debt entries in the task file's `## Decisions` section (format: `Decision (Tech Debt — auto-recorded): <severity> — <description>`).
3. Record any manual-verification requirements as UAT buffer entries via the `crafter-buffer` skill (see `skills/crafter-buffer/SKILL.md`).
4. Commit automatically per `{CRAFTER_HOME}/rules/post-change.md`.

For the canonical four-retained-gates and green-commit-invariant rules that govern this branch — including what constitutes a commit-blocking condition vs. a record-and-continue condition — see `rules/do-workflow.md` → `### --auto (unattended orchestration)`.

When `--auto` is **not** set, fall through to paths (1)–(3) below.

---

**Paths (1)–(3) apply only when `--auto` is not set.**

For non-`--auto` runs, the orchestrator produces and presents a structured **Phase Summary** to the user, then gates the commit on an approval signal. The summary is assembled from context already available to the orchestrator — no new task-file fields or agent invocations are needed:

- **What was implemented** — a brief statement of what the phase delivered (derived from the phase name/description in the task file).
- **Auto-fixed findings** — any Critical or Major findings that were raised during the fix loop and cleared before the review closed clean. List each by severity and short description.
- **Remaining Minor/Suggestion findings** — any open Minor or Suggestion items from the final review report.
- **Accepted Decisions** — any Decision entries recorded during this phase (from Step 5 drift-check handling or the review loop).

If there were no findings of any kind (review was clean on the first pass, fix loop never ran, no Decisions recorded), the summary may be omitted — the orchestrator proceeds directly via the auto-approve path below.

Choose the first path that applies:

#### (1) Auto-approve on clean summary (no user interaction required)

Conditions: zero remaining findings of any severity in the final review state (this covers both the case where the review was clean on the first pass AND the case where the fix loop ran and cleared all Critical/Major with no Minor/Suggestion remaining).

**Exception — manual verification:** If the phase plan or any of its steps explicitly states that verification of this phase requires manual testing (e.g., UI interaction, external integration, non-automatable scenarios), the orchestrator must **wait for explicit user confirmation** even on a fully clean summary. Matching is case-insensitive — "requires", "REQUIRES", "Required", etc., all trigger the wait. Do not introduce a new task-file schema for this flag — treat any plain-text statement that verification requires manual testing as sufficient to trigger the wait; mentions that no manual verification is needed do not trigger it. This exception overrides auto-approve entirely for that phase.

When auto-approve applies: present a one-line notice ("Phase clean — committing automatically.") and proceed directly to the commit per `{CRAFTER_HOME}/rules/post-change.md`.

#### (2) Silence-as-approval — opt-in via `--fast` flag (see Skill options above)

Conditions: the crafter-do skill carries the `--fast` flag (declared in Skill options above) AND remaining Minor/Suggestion findings exist.

Present the Phase Summary and wait for the user's next turn; if that turn does not raise concerns about the summary, treat it as implicit approval and commit. Record each remaining Minor/Suggestion finding as a tech-debt entry in the task file's `## Decisions` section (format: `Decision (Tech Debt — auto-recorded): <severity> — <description>`), then proceed to the commit per `{CRAFTER_HOME}/rules/post-change.md`.

Note: the manual-verification exception in path (1) also applies here — if manual verification is required, `--fast` does not bypass the explicit confirmation wait.

#### (3) Explicit approval — default

Conditions: remaining Minor/Suggestion findings exist AND the `--fast` flag is not set.

Present the Phase Summary and wait for an affirmative response from the user. **Silence does not count as approval.** Do not proceed to the commit until the user explicitly confirms (e.g., "approved", "looks good", "proceed"). If the user raises concerns, resolve them before committing.

### Commit

On approval (any path), run the commit per `{CRAFTER_HOME}/rules/post-change.md`. The commit is automatic — no additional user prompt for the commit command itself.

After committing, continue to **Step 6a** (session break, Medium/Large scope) or **Steps 7–9** (last phase or Small scope).
