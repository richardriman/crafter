# Branch Protection Baseline — `main`

**GH-7** · Solo-maintainer mode

## Current settings (live)

Edit at: **GitHub → Settings → Branches → `main` protection rule**

### Enabled

| Setting | Value | Reason |
|---|---|---|
| Require linear history | ✅ | Keeps commit graph clean; enforces rebase/squash workflow |
| Require conversation resolution before merging | ✅ | Ensures review comments are addressed before merge |
| Allow force pushes | ❌ | Prevents history rewrite on `main` |
| Allow deletions | ❌ | Prevents accidental branch deletion |

### Intentionally disabled (solo-maintainer mode)

| Setting | Value | Reason |
|---|---|---|
| Require approving reviews | ❌ | No second reviewer; would block self-merge |
| Include administrators | ❌ | Would lock the sole maintainer out; not appropriate for single-person repo |
| Require signed commits | ❌ | Not enforced at repo level; can be revisited |
| Lock branch | ❌ | Not a read-only archive |

## Editing branch protection safely

1. Go to **GitHub → Settings → Branches**.
2. Click **Edit** on the `main` rule.
3. Change only the specific setting needed.
4. Save — changes take effect immediately; no deployment required.

⚠️ If you accidentally enable **Include administrators** while the only rule requires reviews, you will be locked out. Disable **Include administrators** first before adding review requirements.

## Multi-maintainer transition checklist

When the project grows beyond a single maintainer, revisit these settings in order:

1. **Enable required approving reviews** (set to 1 or more).
2. **Enable Include administrators** — this makes the rules apply to everyone equally.
3. **Consider adding CODEOWNERS** — auto-assigns reviewers by path.
4. **Consider requiring signed commits** — adds audit trail for all contributors.
5. Update this document and the issue tracker (open a `[CHANGE]` issue).

## References

- [GitHub docs: Managing a branch protection rule](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/managing-a-branch-protection-rule)
- `AGENTS.md` — merge authority policy for this repository
