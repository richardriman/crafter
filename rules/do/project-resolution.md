# Project Resolution

Determine the project root path (`PROJECT_PATH`) so all Crafter context references point to the right place.

1. **Check for `--project <path>` in `$ARGUMENTS`.** If the arguments contain `--project <path>` (e.g., `--project rust fix the parser`), extract the path as `PROJECT_PATH` and strip `--project <path>` from the remaining arguments. The `--project` flag can appear anywhere in the arguments but conventionally comes first. The path is a relative directory path (e.g., `rust`, `rust/`, `packages/frontend`). After extracting the path, verify the directory exists. If it does not exist, tell the user: "Directory `<path>` not found — please check the path and try again." and stop (do not continue the workflow). Use the remaining arguments (after stripping) as the effective `$ARGUMENTS` for all subsequent steps.

2. **If no `--project` was specified**, check whether `.crafter/` exists at the current working directory.
   - If yes: set `PROJECT_PATH` to `.` (current directory). Done.
   - If no: scan one level deep for non-hidden directories (skip names starting with `.`) containing `.crafter/` (i.e., check `*/.crafter/`, excluding `.*/.crafter/`).
      - **Exactly one found:** use it as `PROJECT_PATH`. Inform the user, e.g., "Found project in `rust/`, using it. Tip: you can use `--project rust` to skip this detection next time."
      - **Multiple found:** list them and ask the user which one to use. Mention the `--project` shortcut so users discover it naturally, e.g., "Found projects: `rust/`, `elixir/`. Which one would you like to work on? (Tip: skip this next time with `/crafter-do --project rust ...`)" — then **wait for the user's response** before continuing.
      - **None found:** repeat this exact scan using legacy `.planning/` paths as fallback (`.planning/`, `*/.planning/`). If still none found, set `PROJECT_PATH` to `.` (the normal single-project path — `.crafter/` may be created later by the workflow).

3. **Resolve context directory name (`CRAFTER_DIR`) inside `PROJECT_PATH`.**
   - If `{PROJECT_PATH}/.crafter/` exists: set `CRAFTER_DIR` to `.crafter`.
   - Else if `{PROJECT_PATH}/.planning/` exists: set `CRAFTER_DIR` to `.planning` (legacy fallback), then proactively offer migration:
     - Recommended command: `git -C {PROJECT_PATH} mv .planning .crafter`
     - Ask the user whether to run it now.
     - If user approves and the command succeeds: set `CRAFTER_DIR` to `.crafter`.
     - If user declines or the command fails: continue with `.planning`.
   - Else: set `CRAFTER_DIR` to `.crafter` (new project default).

**Important:** Use `{PROJECT_PATH}/{CRAFTER_DIR}` as the base for all context paths throughout the entire workflow — task files, context files, architecture references passed to agents, everything.

After `--project` extraction, the remaining text is the **effective request** — this is what all subsequent steps refer to when they mention "the user's request" or `$ARGUMENTS`.
