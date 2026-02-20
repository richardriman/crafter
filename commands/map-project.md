---
name: "crafter:map-project"
description: "Analyze codebase and generate/update .planning/ context files"
---

Read and follow all rules from `~/.claude/crafter/rules.md` (or `.claude/crafter/rules.md` if installed locally).

You are the **orchestrator**. Your job is to manage the mapping workflow and communicate with the user. You delegate codebase analysis to the Analyzer subagent and present its results for approval before writing any files.

---

## Step 1 — Scan and Delegate Analysis

Collect the relevant files to analyze:
- Directory layout
- Package manifests: `package.json`, `Cargo.toml`, `go.mod`, `requirements.txt`, `pyproject.toml`, `Gemfile`, `pom.xml`, etc.
- Existing README or documentation
- Configuration files (`.env.example`, `docker-compose.yml`, CI configs, etc.)
- Source code structure (entry points, main modules, test layout)
- Existing `.planning/` files (if any)

Delegate analysis to the **Analyzer** subagent:

1. Spawn a subagent using `~/.claude/crafter/meta-prompts/analyze.md` as its system prompt (or `.claude/crafter/meta-prompts/analyze.md` if installed locally).
2. Provide it with: all the files collected above.
3. Receive the analysis report and proposed `.planning/` file contents.

---

## Step 2 — Determine the Situation

Use the Analyzer's report to determine which case applies:

### A — Large existing CLAUDE.md (>50 lines)

If a `CLAUDE.md` exists with more than 50 lines:

1. Present the Analyzer's proposed decomposition to the user:
   - What content belongs in `.planning/PROJECT.md`?
   - What belongs in `.planning/ARCHITECTURE.md`?
   - What belongs in `.planning/STATE.md`?
   - What (if anything) genuinely doesn't fit `.planning/` and should stay in CLAUDE.md?
2. **Wait for user approval before writing any files.**
3. After approval:
   - Write the `.planning/` files with the redistributed content.
   - Reduce `CLAUDE.md` to the Crafter snippet (plus any approved non-planning content).

### B — Clean project (no .planning/ files)

If `.planning/` does not exist or its files are empty templates:

1. Present the Analyzer's draft content for `PROJECT.md`, `ARCHITECTURE.md`, and `STATE.md`.
2. Ask the user clarifying questions for anything the Analyzer flagged as unclear:
   - Project purpose and target users
   - Conventions or preferences not visible in the code
   - Current priorities and planned work
3. **Wait for review and approval before writing files.**
4. After approval, write the `.planning/` files.

### C — Existing .planning/ files

If `.planning/` files already exist and have content:

1. Present the Analyzer's proposed updates — what would change in each file and why.
2. **Wait for user approval before modifying any file.**
3. After approval, apply the updates.

---

## Notes

- All generated content in `.planning/` files must be in English.
- Never overwrite existing content without showing the user what will change.
- When in doubt about something, ask — don't guess.

---

## Step 3 — Set Up CLAUDE.md

After writing the `.planning/` files, ensure the Crafter snippet is present in `CLAUDE.md`.

Read the snippet content from `~/.claude/crafter/templates/claude-md.snippet` (or `.claude/crafter/templates/claude-md.snippet` if installed locally).

Apply the following logic:

- **No `CLAUDE.md`** → create it containing only the snippet above.
- **`CLAUDE.md` exists, no `<!-- crafter:start -->` marker** → append the snippet to the end of the file.
- **`CLAUDE.md` exists, `<!-- crafter:start -->` marker present** → replace everything between `<!-- crafter:start -->` and `<!-- crafter:end -->` (inclusive) with the snippet above.

If `CLAUDE.md` already has content beyond the snippet, preserve it.

## Step 4 — Update STATE.md

After writing the `.planning/` files and setting up `CLAUDE.md`, update `.planning/STATE.md`:
- Add an entry to **Recent Changes**
- Update **Current Focus** if it has shifted
- Check off any items in **Done**

Show the user what was updated.
