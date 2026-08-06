# Post-Change Steps

## Check Documentation

Review whether the changes affect any `{PROJECT_PATH}/{CRAFTER_DIR}/` context files beyond STATE.md:

- **PROJECT.md** — update if the stack, dependencies, or conventions changed.
- **ARCHITECTURE.md** — delegate this check to the **Implementer** agent. The Implementer reads ARCHITECTURE.md, compares it with what changed, and proposes updates if needed. The orchestrator does not read ARCHITECTURE.md directly.

If updates are needed, show the proposed changes to the user and wait for approval before applying.

If nothing needs updating, move on silently.

## COMMIT

The orchestrator commits **automatically** — no user prompt for the commit itself — when all three preconditions are met:

1. **Phase verification passed** — the Verifier has signed off on the phase.
2. **Clean review** — the review fix loop has closed with no Critical or Major findings remaining.
3. **Approval signal received** — the orchestrator has received an approval signal via one of the three paths defined in the phase summary approval gate in `{CRAFTER_HOME}/skills/crafter-do/SKILL.md` (auto-approve on clean summary | silence-approve when the `--fast` flag is set | explicit user approval as default).

Use conventional commits format: `feat` / `fix` / `refactor` / `docs` / `chore` / `test`

One logical change = one commit.

**Do not push to remote.** Pushing is explicitly forbidden as part of the automatic commit flow.

## Update STATE.md

At end-of-task (after the final phase's per-phase commit), update `{PROJECT_PATH}/{CRAFTER_DIR}/STATE.md` and include this change in the consolidated end-of-task commit (see section below):
- Add an entry to **Recent Changes**
- Update **Current Focus** if it has shifted
- Remove or update any relevant **Known Issues** entries

Stage the STATE.md edit but do not commit it on its own — it lands as part of the consolidated end-of-task commit (see section below).

Show the user what was updated.

## Complete Task File

If a task file exists for the current workflow (in `{PROJECT_PATH}/{CRAFTER_DIR}/tasks/`), complete it per `{CRAFTER_HOME}/rules/task-lifecycle.md`.

## Consolidated End-of-Task Commit

After the final phase's per-phase commit has landed, the orchestrator may produce one **consolidated follow-up commit** covering end-of-task housekeeping:

- PROJECT.md / ARCHITECTURE.md updates
- STATE.md update

If any of these updates exist, bundle them into a **single** commit using conventional commits format. Do not create separate commits for docs and STATE.md. If none of these updates are needed, no follow-up commit is created.

This commit shares the same constraints as the per-phase commit: automatic, no push to remote, conventional commits format.

## STATE.md commit-hash backfill

When the orchestrator updates `STATE.md` "Recent Changes" with a row referencing the current task before the per-task commit lands, use a `TBD` placeholder for the commit hash. After the PR is opened (or the per-task commit lands locally if no PR is being opened), a small `chore(state): backfill GH#NN commit hash in STATE.md` commit replaces `TBD` with the real SHA. This pattern was first used in GH#16 (`5e8ceba`) and is the standard way to handle the chicken-and-egg between "STATE.md updated" and "the commit that updates STATE.md."

Do NOT introduce a different placeholder convention — keep it consistent with the GH#16 precedent.

## Session Wrap-Up

After completing the task, suggest that the user can start a fresh session for the next piece of work:

> If there's more to do, you might want to run `/clear` and then start your next task with `/crafter-do` or `/crafter-debug` to keep the context clean.
