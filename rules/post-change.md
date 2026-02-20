# Post-Change Steps

## Check Documentation

Review whether the changes affect any `.planning/` context files beyond STATE.md:

- **PROJECT.md** — update if the stack, dependencies, or conventions changed.
- **ARCHITECTURE.md** — update if the structure, patterns, or key decisions changed.

If updates are needed, show the proposed changes to the user and wait for approval before applying.

If nothing needs updating, move on silently.

## COMMIT

**Only commit when the user explicitly says to.**

Use conventional commits format: `feat` / `fix` / `refactor` / `docs` / `chore` / `test`

One logical change = one commit.

## Update STATE.md

After a successful commit, update `.planning/STATE.md`:
- Add an entry to **Recent Changes**
- Update **Current Focus** if it has shifted
- Check off any items in **Done**
- Remove or update any relevant **Known Issues** entries

Show the user what was updated.
