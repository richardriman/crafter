---
name: "crafter:do"
description: "Perform a change using Crafter workflow (adaptive: small/medium/large scope)"
---

Read and follow all rules from `~/.claude/crafter/rules.md`.

Then read the project context files (if they exist):
- `.planning/PROJECT.md`
- `.planning/ARCHITECTURE.md`
- `.planning/STATE.md`

The user's request is: $ARGUMENTS

---

## Step 1 — Auto-detect scope

Analyze the request and classify it:

- **Small** — touches 1–3 files, intent is clear, change is isolated
- **Medium** — touches multiple files, intent is clear, change is cross-cutting
- **Large** — vague or ambiguous request, architectural impact, many files, or unfamiliar territory

## Step 2 — DISCUSS / RESEARCH (Large scope only)

If scope is **Large**, pause here. Ask the user clarifying questions before planning:
- What is the desired outcome?
- Are there constraints or preferences?
- Are there approaches to explore?

Do not proceed until you have enough information to write a solid plan.

## Step 3 — PLAN

Write a plan in plain, conversational language. Include:

- **What** you will do and **why**
- **Files** that will be affected
- **Alternatives considered** (for Medium and Large scope)
- **Verification criteria** — how you and the user will know the change is correct

**Wait for explicit user approval before proceeding.**

If the user has concerns or requests changes to the plan, revise it and wait again.

## Step 4 — EXECUTE

Implement exactly what was approved in the plan.

- Never auto-commit.
- Never change architecture without a discussion.
- If you discover something unexpected mid-execution that would change the plan, stop and inform the user.

For **Medium** and **Large** scope: execute one step at a time and pause for review after each step.

## Step 5 — REVIEW

Show a diff of all changes made. Highlight any deviation from the approved plan, even minor ones.

Wait for the user's assessment.

## Step 6 — VERIFY

Check each verification criterion defined in the plan:

- Run relevant tests if applicable.
- Report clearly what passed and what (if anything) did not.

## Step 7 — COMMIT

**Only commit when the user explicitly says to.**

Use conventional commits format:
- `feat:` new feature
- `fix:` bug fix
- `refactor:` code restructuring without behavior change
- `docs:` documentation only
- `chore:` tooling, config, dependencies
- `test:` adding or updating tests

One logical change = one commit.

## Step 8 — Update STATE.md

After a successful commit, update `.planning/STATE.md`:
- Add an entry to **Recent Changes**
- Update **Current Focus** if it has shifted
- Check off any items in **Done**

Show the user what was updated.
