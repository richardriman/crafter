# Post-Change Steps

## Check Documentation

Review whether the changes affect any `{PROJECT_PATH}/.planning/` context files beyond STATE.md:

- **PROJECT.md** — update if the stack, dependencies, or conventions changed.
- **ARCHITECTURE.md** — delegate this check to the **Implementer** agent. The Implementer reads ARCHITECTURE.md, compares it with what changed, and proposes updates if needed. The orchestrator does not read ARCHITECTURE.md directly.

If updates are needed, show the proposed changes to the user and wait for approval before applying.

If nothing needs updating, move on silently.

## COMMIT

**Only commit when the user explicitly says to.**

Use conventional commits format: `feat` / `fix` / `refactor` / `docs` / `chore` / `test`

One logical change = one commit.

## Update STATE.md

After a successful commit, update `{PROJECT_PATH}/.planning/STATE.md`:
- Add an entry to **Recent Changes**
- Update **Current Focus** if it has shifted
- Remove or update any relevant **Known Issues** entries

Show the user what was updated.

## Complete Task File

If a task file exists for the current workflow (in `{PROJECT_PATH}/.planning/tasks/`), complete it per `~/.claude/crafter/rules/task-lifecycle.md`.

## Update Skillbook

After completing the task file, reflect on the task and extract observations for the project's skillbook. Only do this if the `crafter` CLI binary is available at `~/.claude/crafter/bin/crafter`.

1. Review what happened during the task: Did the implementer struggle with something project-specific? Did the reviewer flag a recurring pattern? Did the planner miss something about the project structure?
2. Formulate 0-3 observations (only if genuinely useful — do not force observations for trivial tasks). Each observation needs:
   - **agent**: which agent this applies to (implementer, reviewer, planner, verifier, analyzer)
   - **rule**: the learned guideline, written as an instruction
   - **rationale**: what happened that led to this observation
3. For each observation, run via Bash:
   ```
   ~/.claude/crafter/bin/crafter skillbook add \
     --agent "<agent>" \
     --rule "<rule text>" \
     --rationale "<rationale text>" \
     --task "<task-filename>" \
     --file {PROJECT_PATH}/.planning/skillbook.json
   ```
4. The CLI handles deduplication and confidence promotion automatically. If a similar skill already exists, it will be merged and promoted.
5. Briefly tell the user what was learned (e.g., "Added 2 observations to the project skillbook: ...").
6. If the CLI binary is not available or the command fails, skip silently — skillbook is optional.

Focus on project-specific patterns, not general programming knowledge:
- Good: "This project uses X pattern for Y", "Tests need Z setup", "The review found A was a recurring issue"
- Bad: "Always use descriptive variable names" (too generic), "Fixed a typo" (not a pattern)

## Session Wrap-Up

After completing the task, suggest that the user can start a fresh session for the next piece of work:

> If there's more to do, you might want to run `/clear` and then start your next task with `/crafter:do` or `/crafter:debug` to keep the context clean.
