# Task: Remote install without cloning

## Metadata
- **Date:** 2026-02-22
- **Branch:** main
- **Status:** active
- **Scope:** Medium

## Request
Lepsi install, kterym budou moci uzivatele nainstalovat crafter, aniz by museli predtim naklonovat repozitar.

## Plan
- [x] Step 1: Rewrite `install.sh` — add remote auto-detection (if source files not found at SCRIPT_DIR, download tarball from GitHub), `--version` flag, temp dir cleanup via trap EXIT
- [x] Step 2: Update `README.md` — Quick Start leads with curl one-liner, clone as alternative for contributors
- [x] Step 3: Update `rules/update-check.md` — update notice shows curl command instead of local install.sh reference

## Decisions
<!-- Key decisions made during the workflow, in chronological order -->

## Outcome
<!-- Filled on completion: what was actually done, commit SHA(s), any deviations from plan -->
