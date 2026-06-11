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
     - Write and commit the bump locally:
       ```
       printf '%s\n' "<target-version>" > VERSION
       git add VERSION
       git commit -m "chore: bump version to <target-version>"
       ```
     - **Attempt a direct push first** (works for unprotected branches):
       ```
       git push origin HEAD
       ```
     - **If the direct push is rejected because the branch is protected** — any non-zero exit from `git push origin HEAD` may indicate branch protection; common phrases include `protected branch hook declined`, `GH006`, and `Changes must be made through a pull request`, but GitHub rulesets and other configurations can emit different wording. Treat ANY rejection of `git push origin HEAD` as a possible protection rejection and inspect the output before concluding it is something else — route the bump through a PR instead:
       - Resolve the default branch name (the result is the bare name, e.g. `main`; substitute it for `<default-branch>` everywhere below):
         ```
         git remote show origin | sed -n '/HEAD branch/s/.*: //p'
         ```
       - Move the bump commit to a dedicated branch and reset the local default branch back to its upstream (this assumes the local default branch tracks `origin/<default-branch>`, which is true for a normal clone):
         ```
         git branch chore/bump-<target-version>
         git reset --hard @{u}
         git checkout chore/bump-<target-version>
         git push -u origin chore/bump-<target-version>
         ```
       - Open the PR:
         ```
         gh pr create --base <default-branch> --head chore/bump-<target-version> \
           --title "chore: bump version to <target-version>" \
           --body "Bumps VERSION to <target-version> ahead of the v<target-version> release."
         ```
       - Attempt the squash merge immediately:
         ```
         gh pr merge chore/bump-<target-version> --squash --delete-branch
         ```
         If the merge is rejected because the PR is not yet mergeable — required status checks are still pending, or required approvals are missing — **do NOT treat this as a hard error and do NOT force-merge**. Instead, surface the blocker to the user: the release cannot continue until the bump PR satisfies the repository's branch-protection requirements (checks must pass and/or approval must be granted). As an alternative to polling, you may enable auto-merge so GitHub merges automatically once requirements are met:
         ```
         gh pr merge chore/bump-<target-version> --squash --auto --delete-branch
         ```
         Note: if auto-merge is not enabled on the repository, `gh pr merge --auto` will error. In that case, wait for the required checks/approvals to clear and re-run the plain squash merge: `gh pr merge chore/bump-<target-version> --squash --delete-branch`.
         Either way, the remaining release steps (tag creation, binary upload) must wait until the PR has actually merged. Do not proceed until the squash merge is confirmed.
       - Once the PR is merged, return to the default branch and reset to the canonical remote state (`pull --ff-only` is avoided here because a squash merge rewrites history — the local default branch's bump commit is not the squashed remote commit, so fast-forward would fail; resetting to the fetched upstream reliably lands the local branch on the merged bump commit):
         ```
         git checkout <default-branch>
         git fetch origin
         git reset --hard @{u}
         ```
       - Note: GitHub may auto-delete the topic branch on merge, so `--delete-branch` reporting that the branch is already gone is not an error — just continue. This only affects the topic branch and does not interfere with the subsequent fetch/reset of the default branch.
     - **Invariant:** the release tag (Step 5 — Create the Release and Upload CLI Binaries) must NOT be created until the bumped `VERSION` is on the default branch — meaning either the direct push succeeded, or the PR was fully merged and the local default branch reset to the upstream. This ensures `gh release create` tags the commit that actually contains the new VERSION.

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
   If `go` is not on PATH because it is managed by a version manager (e.g. mise or asdf), run the build through it instead — for example:
   ```
   cd cli
   mise exec -- make release
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
