# Task: Release skill aware of protected default branch

## Metadata
- **Date:** 2026-06-11
- **Work branch:** fix/release-skill-protected-main
- **Status:** completed
- **Scope:** Small

## Request
optimalizuj release skill, aby o tomhle vedel a nemusel porad dokola vynalezat kolo

Context: during the v0.13.0 release, the `crafter-release` skill's version-bump step ran `git push origin HEAD` directly against `main`, which is a protected branch. The push was rejected (`GH006 — Changes must be made through a pull request`) and the bump had to be manually re-routed through a PR (branch → `gh pr create` → squash-merge → sync). The skill should encode this so it does not reinvent the wheel each release. Also encode that `make release` may need to run through a version manager (e.g. `mise exec --`) when `go` is not on PATH.

## Plan

**Plan status:** approved

### Phase 1 — Protected-branch-aware version bump + mise-aware build
- [x] Step 1: Rewrite Step 1 sub-step 6 (bump path) of `.claude/skills/crafter-release/SKILL.md` so the bump commits locally, attempts a direct `git push origin HEAD`, and on a protected-branch rejection automatically routes the bump through a PR (move commit to `chore/bump-<version>`, reset local default branch, push branch, `gh pr create`, `gh pr merge --squash --delete-branch`, then `git checkout <default>` + `git pull --ff-only`). Add an explicit invariant: the release/tag must not be created until the bump is on the default branch. Add a mise-aware note to Step 5.3 build (`mise exec -- make release` when `go` is not on PATH).
- [x] Phase verification
- [x] Phase review

Karpathy Contract:
- **Outcome:** The release skill handles a version bump on a protected default branch without manual intervention.
- **Scope boundary:** Only `.claude/skills/crafter-release/SKILL.md`. Step 1 bump path + Step 5.3 build note.
- **Non-goals:** No change to release-notes generation, commit gathering, or binary upload logic. No behavioral change for unprotected repos (direct push stays the primary path).
- **Simplicity constraint:** Try-direct-then-PR-fallback; no upfront protection-detection API calls.
- **Drift criteria:** Unprotected flow stays effectively identical aside from the added fallback branch; tag-after-bump invariant added.
- **Verification evidence:** Re-read the edited Step 1 — direct push is still the first attempt, the PR fallback is correctly conditioned on a protected-branch rejection, and the tag-after-bump invariant is present. Step 5.3 has the mise note.
- **Stop conditions:** Any need to change release-notes/upload logic, or to add protection-detection API calls.

## Decisions
- **Decision (User Accepted):** Include the mise-aware build note in Step 5.3 in addition to the protected-main fix. **Reason:** Same "reinventing the wheel" class of problem hit during the v0.13.0 release.

## Outcome

Done in commit `534d0c3` (`fix(crafter-release): route version bump through PR on protected default branch`). One file changed: `.claude/skills/crafter-release/SKILL.md` (+~50/-1).

**Edit A — Step 1 sub-step 6 (bump path):** commit `VERSION` locally → attempt direct `git push origin HEAD` (primary, unchanged for unprotected repos) → on a protected-branch rejection, auto-route through a `chore/bump-<v>` topic branch + squash-merged PR, then `git checkout <default>` + `git fetch origin` + `git reset --hard @{u}`. Added: default-branch resolution command, ANY-non-zero-push-exit protection detection (rulesets emit varying wording), not-yet-mergeable handling (surface blocker to user, optional `--auto`, auto-merge-unavailable fallback), and a tag-after-bump invariant naming Step 5 explicitly.

**Edit B — Step 5 sub-step 3 (build):** `mise exec -- make release` note when `go` is not on PATH; `cd cli && make release` stays the default.

**Workflow:** crafter-do Small scope, single phase. Verifier: no drift. Reviewer round 1: 2 Major (gh pr merge under required checks; `pull --ff-only` after squash) + 4 Minor/Suggestion — all fixed in the fix loop. Reviewer round 2: prior findings resolved, 3 new optional (1 Minor + 2 Suggestion) — all fixed. No deviations from the approved plan.

No behavioral change for unprotected repos; no upfront `gh api` protection-detection (try-direct-then-fallback only).
