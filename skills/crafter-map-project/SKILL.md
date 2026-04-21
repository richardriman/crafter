---
name: "crafter-map-project"
description: "Analyze codebase and generate/update .crafter/ context files"
---

Read and follow these rules:
- `~/.claude/crafter/rules/core.md`
- `~/.claude/crafter/rules/delegation.md`

You are the **orchestrator**. Your job is to manage the mapping workflow and communicate with the user. You delegate codebase analysis to the Analyzer agent and present its results for approval before writing any files.

---

## Step 1 — Scan and Delegate Analysis

Collect the relevant files to analyze:
- Directory layout
- Package manifests: `package.json`, `Cargo.toml`, `go.mod`, `requirements.txt`, `pyproject.toml`, `Gemfile`, `pom.xml`, etc.
- Existing README or documentation
- Configuration files (`.env.example`, `docker-compose.yml`, CI configs, etc.)
- Source code structure (entry points, main modules, test layout)
- Existing `.crafter/` files (if any)
- Existing legacy `.planning/` files (if any)

Delegate analysis to the **Analyzer** agent:

1. Spawn the `crafter-analyzer` agent.
2. Provide it with high-level pointers to what to analyze (the list from Step 1). Do not inject file contents — the Analyzer explores the codebase itself.
3. Receive the analysis report and proposed `.crafter/` file contents.

---

## Step 2 — Determine the Situation

Use the Analyzer's report to determine which case applies:

### A — Clean project (no .crafter/ files)

If `.crafter/` does not exist or its files are empty templates:

1. Present the Analyzer's draft content for `PROJECT.md`, `ARCHITECTURE.md`, and `STATE.md`.
2. Ask the user clarifying questions for anything the Analyzer flagged as unclear:
   - Project purpose and target users
   - Conventions or preferences not visible in the code
   - Current priorities and planned work
3. **Wait for review and approval before writing files.**
4. After approval, write the `.crafter/` files.

### B — Existing .crafter/ files

If `.crafter/` files already exist and have content:

1. Present the Analyzer's proposed updates — what would change in each file and why.
2. **Wait for user approval before modifying any file.**
3. After approval, apply the updates.

---

### C — Legacy `.planning/` project

If `.planning/` exists but `.crafter/` does not:
1. Tell the user this is a legacy Crafter layout and recommend migration.
2. Propose migration command: `git mv .planning .crafter`.
3. Wait for explicit approval before running migration.
4. If migration succeeds, continue with `.crafter/`.
5. If the user declines migration, continue using `.planning/` as fallback for this run only.

## Notes

- All generated content in `.crafter/` files must be in English.
- Never overwrite existing content without showing the user what will change.
- `.planning/` is legacy fallback only; `.crafter/` is the primary path.

---

## Step 3 — Update STATE.md

After writing context files, update the active context directory's `STATE.md` (`.crafter/STATE.md` or legacy fallback `.planning/STATE.md`):
- Add an entry to **Recent Changes**
- Update **Current Focus** if it has shifted

Show the user what was updated.
