---
name: "crafter:map-project"
description: "Analyze codebase and generate/update .planning/ context files"
---

Read and follow all rules from `~/.claude/crafter/rules.md`.

---

## Step 1 — Scan the Codebase

Explore the project directory to understand its structure:

- Directory layout
- Package manifests: `package.json`, `Cargo.toml`, `go.mod`, `requirements.txt`, `pyproject.toml`, `Gemfile`, `pom.xml`, etc.
- Existing README or documentation
- Configuration files (`.env.example`, `docker-compose.yml`, CI configs, etc.)
- Source code structure (entry points, main modules, test layout)

---

## Step 2 — Determine the Situation

### A — Large existing CLAUDE.md (>50 lines)

If a `CLAUDE.md` exists with more than 50 lines:

1. Read it carefully, section by section.
2. Propose a decomposition:
   - What content belongs in `.planning/PROJECT.md`?
   - What belongs in `.planning/ARCHITECTURE.md`?
   - What belongs in `.planning/STATE.md`?
   - What (if anything) genuinely doesn't fit `.planning/` and should stay in CLAUDE.md?
3. Show the proposed split clearly to the user.
4. **Wait for user approval before writing any files.**
5. After approval:
   - Write the `.planning/` files with the redistributed content.
   - Reduce `CLAUDE.md` to the Crafter snippet (plus any approved non-planning content).

### B — Clean project (no .planning/ files)

If `.planning/` does not exist or its files are empty templates:

1. Generate draft content for `PROJECT.md`, `ARCHITECTURE.md`, and `STATE.md` from the codebase analysis.
2. Ask the user clarifying questions for anything that cannot be inferred from the code:
   - Project purpose and target users
   - Conventions or preferences not visible in the code
   - Current priorities and planned work
3. Show the generated content to the user.
4. **Wait for review and approval before writing files.**
5. After approval, write the `.planning/` files.

### C — Existing .planning/ files

If `.planning/` files already exist and have content:

1. Re-scan the codebase for changes since the files were last updated.
2. Propose specific updates to each file (show what would change and why).
3. **Wait for user approval before modifying any file.**
4. After approval, apply the updates.

---

## Notes

- All generated content in `.planning/` files must be in English.
- Never overwrite existing content without showing the user what will change.
- When in doubt about something, ask — don't guess.
