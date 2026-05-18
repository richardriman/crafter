# Step 9b — PR Composition (`--auto` only)

**Trigger:** Runs ONLY under `--auto` (`auto: true` in frontmatter), and ONLY after Steps 7–9 complete (which may or may not have produced a consolidated end-of-task commit). The latest commit on the work branch will be either the consolidated commit OR the final per-phase commit, depending on whether PROJECT.md / ARCHITECTURE.md / skillbook / STATE.md updates were needed. By the time this step runs, STATE.md is updated and the task file's `## Outcome` section is filled in. Non-`--auto` runs do not execute this step; the user composes the PR manually by invoking `gh pr create` themselves.

This step is the concrete implementation of the `Plan → Execute → Verify → Review → PR end-to-end` promise documented in `rules/do-workflow.md → ### --auto (unattended orchestration)`.

**Inputs:**

- `--run-dir`: `.crafter/run/<task-id>/` — the per-run scratch directory. The orchestrator already tracks `<task-id>` as the task-file basename without extension (e.g., `20260510-feat-gh-17-pr-composer`).
- `--task-file`: `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/<task-id>.md` — the task file path. Already known from the active task context.
- **Baseline body:** a minimal Summary + Test plan block, composed by the orchestrator (LLM-generated from the task file's `## Plan` → Approach paragraph and `## Outcome` section). The Test plan derives from the task file's acceptance criteria ONLY (extracted from the `## Request` section, which quotes the issue ACs, or from the issue body if available) — NOT from phase verification criteria, which would dump dozens of items. If `## Outcome` is empty when this step is reached, that indicates Steps 7–9 did not complete correctly — do NOT proceed to PR composition; exit with state via the Ad-hoc escape hatch.

**Action:**

1. **Compose the baseline body** (Summary + Test plan) from the task file. This is a brief, template-driven LLM composition step — not a new agent delegation. The orchestrator reads the task file's `## Plan → Approach` paragraph and `## Outcome` section (already in context from Steps 7–9) and emits a compact markdown block. Structure rules:
   - No leading newline; ends with exactly one trailing `\n` before concatenation with appended sections (which start with `## Manual QA Plan\n\n…`).
   - `## Summary` and `## Test plan` are H2 headings separated by exactly one blank line (`\n\n`); each heading is followed by one blank line before its body.
   - Test plan items are the issue's acceptance criteria ONLY (see Inputs above).

   ```
   ## Summary\n
   <1–3 sentences derived from ## Plan → Approach and ## Outcome>\n
   \n
   ## Test plan\n
   \n
   - <acceptance criterion 1 from issue ACs>\n
   - <acceptance criterion 2 from issue ACs>\n
   ...\n
   ```

   (The `\n` annotations above make the newline structure explicit; the actual output is plain markdown with those newlines.)

2. **Invoke the rendering subcommand** to produce the appended sections:

   ```sh
   crafter pr-body --run-dir .crafter/run/<task-id>/ --task-file {PROJECT_PATH}/{CRAFTER_DIR}/tasks/<task-id>.md
   ```

   The subcommand outputs the three appended sections (`## Manual QA Plan`, `## Known Gaps`, `## Decisions`) in that fixed order, omitting any section whose source is empty. If all three sources are empty, the output is empty (no appended sections).

3. **Concatenate** the baseline body and the subcommand output to form the full PR body.

4. **Derive the PR title:** use the first line of the most recent commit on the work branch (`git log -1 --format='%s'`); fall back to the task file's H1 heading if that is unavailable or empty.

5. **Open the PR** by passing the full body to `gh pr create`. To avoid shell-interpolation hazards (commit subjects routinely contain `"`, `` ` ``, `$`, `!`), hold the title in a variable and pass the variable — do NOT inline it in a double-quoted string:

   ```sh
   TITLE=$(git log -1 --format='%s')
   printf '%s' "<full-body>" | gh pr create --title "$TITLE" --body-file -
   ```

   Bash does not perform word-splitting or glob-expansion on the right-hand side of a variable assignment, so `TITLE` captures the full commit subject verbatim. Passing `"$TITLE"` to `gh` delivers it as a single argument regardless of embedded quotes or special characters. As an alternative, `gh` ≥ 2.40 supports `--title-file -` for stdin input, which is even safer when a stdin pipe is not already in use.

   Note: `gh pr create` will push the current branch to the remote as part of opening the PR. This is the **only** push in the `--auto` flow — the orchestrator does NOT run `git push` separately at any point. `rules/post-change.md` forbids a standalone `git push`; the push embedded in `gh pr create` is the one-time exception that opens the PR.

**Title derivation rule:** Use the first line of the most recent commit on the work branch (`git log -1 --format='%s'`); fall back to the task file's H1 heading if the git command fails or returns an empty string.

**Failure handling:** If `gh pr create` fails for any reason (network error, authentication failure, branch already has an open PR, branch not pushed, etc.):

1. Record the failure in the task file's `## Decisions` section:
   ```
   Decision (Auto-Recorded): PR creation failed — <error message or summary>
   ```
2. Do NOT run the cleanup hook (the run-dir is preserved for retry/debug).
3. Exit with state via the **Ad-hoc escape hatch** (per `rules/do-workflow.md → #### Ad-hoc escape hatch`). The task file remains the handoff artifact. The run terminates without violating the green-commit invariant (all per-phase and consolidated commits have already landed).

**Success handling:** On `gh pr create` success:

1. Print the PR URL to the user as a one-line notice:
   ```
   PR opened: <URL>
   ```
2. Run the cleanup hook: delete the run directory (`rm -rf .crafter/run/<task-id>/`). This is the PR-success cleanup trigger. If deletion fails, record a tech-debt note and continue — the cleanup failure is non-blocking.
3. Proceed to the session wrap-up (Step 7–9 item 5: suggest `/clear` for the next task).

**Reviewer note:** `## Manual QA Plan` items in the PR body are rendered as GitHub-flavored task list checkboxes (`- [ ] **Title** — verify text`). Reviewers can check them off directly in the PR UI as they complete each manual verification step.
