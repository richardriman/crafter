---
name: "crafter-release"
description: "Prepare and publish a GitHub Release with notes and CLI binary assets"
---

Prepare a GitHub Release for this project, including CLI binary assets.

**Requirements:** `gh` CLI must be installed and authenticated, and you must have push access to `origin`.

---

## Step 1 — Resolve Target Version (with auto-bump)

1. Ensure the working tree is clean before any automated version bump commit:
   ```
   git status --porcelain
   ```
   If non-empty, stop and ask the user to commit/stash first.

2. Sync remote tags so local state matches GitHub:
   ```
   git fetch origin --tags
   ```

3. Read `VERSION` as `<version-file>` (strip whitespace/newline).

4. Determine `<latest-tag>` from origin tags:
   ```
   git tag --list 'v*' --sort=-version:refname | head -n 1
   ```
   Set `<latest-version>` to `<latest-tag>` without the leading `v`. If there is no tag, treat `<latest-version>` as empty.

5. Resolve requested version from user input:
   - If the user provided a version argument (e.g. `0.8.1` or `v0.8.1`), normalize to `<requested-version>` without leading `v`.
   - If no version argument was provided, `<requested-version>` is empty for now.

6. Compare versions using semantic version ordering (`MAJOR.MINOR.PATCH`):
   - If `<latest-version>` is empty (first release):
     - Use `<requested-version>` when provided, otherwise use `<version-file>`.
   - If `<version-file>` is **greater than** `<latest-version>` and no conflicting `<requested-version>` is provided:
     - Use `<version-file>` as `<target-version>` (no bump).
   - If `<version-file>` is **equal to or lower than** `<latest-version>`:
     - Use `<requested-version>` if provided; otherwise ask the user for a new version.
     - Validate that `<target-version>` is strictly greater than `<latest-version>`. If not, ask again.
     - Update `VERSION`, commit, and push before continuing:
       ```
       printf '%s\n' "<target-version>" > VERSION
       git add VERSION
       git commit -m "chore: bump version to <target-version>"
       git push origin HEAD
       ```

7. After `<target-version>` is resolved, check whether release/tag `v<target-version>` already exists — verify both locally and on GitHub:
   ```
   git tag --list "v<target-version>"
   gh release view "v<target-version>" --json tagName 2>/dev/null
   ```
   If either check finds an existing tag/release, stop and report that version as already published.

---

## Step 2 — Gather Commits Since Last Release

1. Determine `<last-tag>` (most recent existing release tag before `v<target-version>`):
   ```
   git tag --list 'v*' --sort=-version:refname | head -n 1
   ```
   - If no tags exist, use the first commit as base:
     ```
     git rev-list --max-parents=0 HEAD
     ```

2. Collect commits since `<last-tag>`:
   ```
   git log <last-tag>..HEAD --oneline
   ```
   (Replace `<last-tag>` with the tag or commit hash from the previous step.)

---

## Step 3 — Generate Release Notes

From the commit list gathered in Step 2, generate structured release notes.

Group commits by conventional commit type:
- `feat` -> **Features**
- `fix` -> **Fixes**
- Everything else (`refactor`, `docs`, `chore`, `test`, `perf`, etc.) -> **Other Changes**

Write descriptions in human-readable form — do not copy raw commit messages verbatim. Summarize intent and impact in plain language. Omit empty sections.

Prioritize user-facing value:
- Omit purely internal maintainer/release-process changes when they are not useful to end users.
- Example to omit by default: internal release workflow/script-only changes.

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

## Step 4 — Review and Approve

Present the following:
- The proposed tag name: `v<target-version>`
- The full generated release notes

**Wait for explicit user approval before proceeding.**

If changes are requested, revise the release notes and present them again until approved.

---

## Step 5 — Create the Release and Upload CLI Binaries

1. Verify that `gh` is installed and authenticated:
   ```
   gh auth status
   ```
   If this fails, stop and ask the user to run `gh auth login`.

2. Create the GitHub Release using the approved notes. Write notes to a temporary file, then run:
   ```
   gh release create v<target-version> --title "v<target-version>" --notes-file /tmp/release-notes.md
   ```

3. Build release binaries:
   ```
   cd cli
   make release
   ```
   This must produce:
   - `cli/bin/crafter-darwin-arm64`
   - `cli/bin/crafter-darwin-amd64`
   - `cli/bin/crafter-linux-amd64`
   - `cli/bin/crafter-linux-arm64`
   If build fails or any expected file is missing, stop and report the error.

4. Upload binaries to the release:
   ```
   gh release upload v<target-version> \
     cli/bin/crafter-darwin-arm64 \
     cli/bin/crafter-darwin-amd64 \
     cli/bin/crafter-linux-amd64 \
     cli/bin/crafter-linux-arm64 \
     --clobber
   ```

5. Verify release assets after upload:
   ```
   gh release view v<target-version> --json url,assets
   ```
   Confirm all 4 expected binaries are present.

6. Report success and include:
   - release URL
   - uploaded asset names
