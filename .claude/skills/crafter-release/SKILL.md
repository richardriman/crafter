---
name: "crafter-release"
description: "Prepare and publish a GitHub Release with AI-generated release notes"
---

Prepare a GitHub Release for this project.

**Requirements:** `gh` CLI must be installed and authenticated. This skill does NOT bump the VERSION file and does NOT commit anything.

---

## Step 1 — Gather Context

1. Sync remote tags so local state matches GitHub:
   ```
   git fetch --tags
   ```
2. Read the `VERSION` file to get the current version (referred to as `<VERSION>` below).
3. Check whether the release `v<VERSION>` already exists — verify **both** locally and on GitHub:
   ```
   git tag --list "v<VERSION>"
   gh release view "v<VERSION>" --json tagName 2>/dev/null
   ```
   If either check finds an existing tag/release, warn the user and stop. The release has already been published, or the VERSION file has not been updated.
4. Run the following to find the most recent release tag:
   ```
   git tag --list 'v*' --sort=-version:refname
   ```
   - If one or more tags exist, take the first result as `<last-tag>`.
   - If no tags exist, find the first commit with:
     ```
     git rev-list --max-parents=0 HEAD
     ```
     and use that commit hash as the base.
5. Run the following to collect all commits since the last release:
   ```
   git log <last-tag>..HEAD --oneline
   ```
   (Replace `<last-tag>` with the tag or commit hash from the previous step.)

---

## Step 2 — Generate Release Notes

From the commit list gathered in Step 1, generate structured release notes.

Group commits by conventional commit type:
- `feat` -> **Features**
- `fix` -> **Fixes**
- Everything else (`refactor`, `docs`, `chore`, `test`, `perf`, etc.) -> **Other Changes**

Write descriptions in human-readable form — do not copy raw commit messages verbatim. Summarize the intent and impact of each change in plain language. Omit empty sections.

Use this structure:

```
## <Title summarizing the release>

<One-line description of what this release delivers overall>

### Features
- <Human-readable description of each feat commit>

### Fixes
- <Human-readable description of each fix commit>

### Other Changes
- <Human-readable description of each remaining commit>
```

---

## Step 3 — Review and Approve

Present the following:
- The proposed tag name: `v<VERSION>`
- The full generated release notes

**Wait for explicit user approval before proceeding.**

If changes are requested, revise the release notes and present them again until approved.

---

## Step 4 — Create the Release

1. Verify that `gh` is installed and authenticated:
   ```
   gh auth status
   ```
   If this fails, stop and ask the user to run `gh auth login`.

2. Create the GitHub Release using the approved notes. Write the approved release notes to a temporary file (e.g. `/tmp/release-notes.md`), then run:
   ```
   gh release create v<VERSION> --title "v<VERSION>" --notes-file /tmp/release-notes.md
   ```

3. Report success and include the URL of the newly created release.
