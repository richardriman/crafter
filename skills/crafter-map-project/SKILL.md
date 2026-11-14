---
name: "crafter-map-project"
description: "Analyze codebase and generate/update .crafter/ context files"
---

Read and follow these rules:
- `{CRAFTER_HOME}/rules/core.md`
- `{CRAFTER_HOME}/rules/delegation.md`
- `{CRAFTER_HOME}/rules/workspace.md`

You are the **orchestrator**. Your job is to manage the mapping workflow and communicate with the user. You delegate codebase analysis to the Analyzer agent and present its results for approval before writing any files.

---

## Step 0 — Workspace Check

Run the detection in `{CRAFTER_HOME}/rules/workspace.md` § Detection from the current working directory, then:

- **Workspace detected, cwd is the workspace root** → workspace mode: in Step 1, also ask the Analyzer to draft creating or updating the workspace description; Step 2 adds case **D**.
- **Workspace detected, cwd is below the root (e.g. a nested repo)** → workspace mode: map this repo as usual and tell the user in one sentence that the workspace description lives in `<root>` and is updated by running `/crafter-map-project` there.
- **No workspace, but the cwd is a workspace candidate** → ask one question: draft a workspace description here? Yes → treat as the workspace root case above. No → continue exactly as without it.
- **Otherwise** → continue exactly as without it.

**In workspace mode, `STATE.md` is neither drafted nor written** — skip it in Steps 1–2 and skip Step 3; an existing `STATE.md` is left alone. Exception: if the run entered workspace mode only through the candidate question and case **D** ends without writing the description, fall back to normal mapping: re-spawn the Analyzer (mode A, normal mapping, no workspace mode) to draft `STATE.md` (use only that draft), then present and write it as in Step 2 case A/B and Step 3. Never write `STATE.md` without that Analyzer draft.

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
3. In workspace mode, say so in the prompt: no `STATE.md` draft; at the workspace root, also a workspace-description draft in the shape of `{CRAFTER_HOME}/rules/workspace.md` § Workspace description (pass that path).
4. Receive the analysis report and proposed `.crafter/` file contents.

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

### D — Workspace description

When Step 0 chose the workspace root case:

1. Present the Analyzer's draft of the `## Workspace` section (shape: `{CRAFTER_HOME}/rules/workspace.md` § Workspace description; it must contain `### Repositories`, which together with the heading is the detection marker) and ask the questions it flagged (unclear relationships, external repos, conventions).
2. Propose the target file: the file that already holds the description (an instruction file or `.crafter/WORKSPACE.md`) → else the existing `CLAUDE.md` → else the existing `AGENTS.md` → else a new `CLAUDE.md`. Only when no instruction file already holds a workspace-marker `## Workspace` section may the user choose `.crafter/WORKSPACE.md` instead; otherwise the target is that instruction file, because its section wins detection.
3. If the target is `CLAUDE.md` or `AGENTS.md` and it has a `## Workspace` heading that is not a workspace marker (no `### Repositories` inside it, per `{CRAFTER_HOME}/rules/workspace.md` § Detection), it has an unrelated meaning: propose renaming that heading rather than overwriting it.
4. **Wait for approval of both the content and the target file before writing.**
5. After approval: for `.crafter/WORKSPACE.md`, write the description over the whole file (the whole file is the description). For `CLAUDE.md` / `AGENTS.md`, replace exactly the workspace-marker `## Workspace` section as bounded in `{CRAFTER_HOME}/rules/workspace.md` § Section boundary and nothing beyond it, or append it when absent (after renaming an unrelated heading, per item 3); leave the rest of the file unchanged.

## Notes

- All generated content in `.crafter/` files must be in English.
- Never overwrite existing content without showing the user what will change.
- `.planning/` is legacy fallback only; `.crafter/` is the primary path.

---

## Step 3 — Update STATE.md

Skip this step in workspace mode (Step 0), except in the Step 0 fallback to normal mapping. Otherwise, after writing context files, update the active context directory's `STATE.md` (`.crafter/STATE.md` or legacy fallback `.planning/STATE.md`):
- Add an entry to **Recent Changes**
- Update **Current Focus** if it has shifted

Show the user what was updated.
