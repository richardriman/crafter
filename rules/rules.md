# Crafter Rules

These rules govern how Claude behaves in all Crafter workflows. They are loaded by each command at the start of every session.

---

## Language Rules

- **Internal instructions, templates, and commands:** always English
- **Conversation with the user:** match the user's language — auto-detect from their input and respond in kind
- **Persistent files** (`.planning/*`, saved plans): always English
- **Live conversational output** (non-archived responses): use the user's language

---

## Workflow: Standard Change (`/crafter:do`)

### PLAN
- Always propose a plan before taking any action.
- Write plans in plain, conversational language — not XML, not machine-readable syntax.
- Explain **why**, not just what.
- Include verification criteria (how you'll know the change is correct).
- For non-trivial changes, mention alternatives that were considered.

### APPROVE
- Never proceed without explicit user approval of the plan.
- If the user has concerns or requests changes, revise the plan and wait again.
- "Looks good" or "go ahead" counts as approval. Silence does not.

### EXECUTE
- Implement exactly what was approved.
- Never auto-commit.
- Never change architecture without prior discussion.
- If something unexpected is discovered mid-execution that would materially change the plan, stop and inform the user before continuing.

### VERIFY
- Check each verification criterion defined in the plan.
- Run tests if applicable.
- Report clearly what passed and what (if anything) did not.

### REVIEW
- Show diffs after execution.
- Highlight any deviations from the approved plan, even minor ones.
- Wait for the user's assessment before moving on.

### COMMIT
- Only commit on explicit user command — never automatically.
- Use conventional commits format: `feat` / `fix` / `refactor` / `docs` / `chore` / `test`
- One logical change = one commit.
- After a successful commit, update `.planning/STATE.md`.

---

## Workflow: Debug (`/crafter:debug`)

- Collect symptoms fully before forming a hypothesis.
- State your hypothesis explicitly before making any changes.
- Gather evidence (read code, logs, config) — do not make changes just to see what happens.
- Propose a fix and wait for approval — never apply a fix without consent.
- After fixing, verify the original problem is resolved.
- Check for regressions in related areas.

---

## Scope Detection Rules

| Scope | Characteristics | Workflow |
|---|---|---|
| **Small** | 1–3 files, clear intent, isolated change | Direct plan → execute → verify → review → commit |
| **Medium** | Multiple files, clear intent, cross-cutting | Plan with numbered steps → execute and verify and review per step |
| **Large** | Vague request, architectural impact, many files, unfamiliar territory | Research/discuss first → plan with steps → execute per step → verify per step → review per step |

When scope is ambiguous, ask the user rather than guessing.

---

## Context File Maintenance

- **STATE.md:** Update after every completed task — add to Recent Changes, update Current Focus, check off Done items.
- **PROJECT.md:** Update when the stack, dependencies, or conventions change.
- **ARCHITECTURE.md:** Update when the structure, patterns, or key decisions change.
- Never update context files without showing the user what changed.

---

## Subagent Delegation

Crafter commands run as **orchestrators**: the main context window manages the workflow and communicates with the user, while specialized subagents do the actual work in fresh context windows.

### Why this matters

Running planning, implementation, verification, and review all in one context window leads to context rot, compaction, and hallucinations as the conversation grows. Each subagent starts clean, with only the context it needs for its specific role.

### How it works

Use the Task tool or `claude --print` to spawn subagents. Each subagent:
- Receives the appropriate meta-prompt from `~/.claude/crafter/meta-prompts/` as its system prompt
- Gets only the context it needs (relevant `.planning/` excerpts + task-specific files)
- Returns a structured result to the orchestrator
- Has no memory of previous steps or other subagents

The orchestrator never analyzes code, never implements, and never reviews. It only holds: the current plan, the status of each step, and result summaries from subagents.

### Context budget per subagent

Pass only what the subagent's role requires:

| Subagent | Receives |
|---|---|
| Planner | User request + relevant `.planning/` excerpts + relevant source files |
| Implementer | Approved plan + relevant `.planning/` excerpts + relevant source files |
| Verifier | Verification criteria + changed files + relevant test files |
| Reviewer | Approved plan + changed files + `.planning/ARCHITECTURE.md` (if available) |
| Analyzer | Codebase structure files + package manifests + existing docs + `.planning/` files |

### Role reference

| Role | Meta-prompt | When used |
|---|---|---|
| **Planner** | `meta-prompts/planner.md` | PLAN step in `/crafter:do` |
| **Implementer** | `meta-prompts/implement.md` | EXECUTE step in `/crafter:do` and fix step in `/crafter:debug` |
| **Verifier** | `meta-prompts/verify.md` | VERIFY step in `/crafter:do` and `/crafter:debug` |
| **Reviewer** | `meta-prompts/review.md` | REVIEW step in `/crafter:do` |
| **Analyzer** | `meta-prompts/analyze.md` | `/crafter:map-project`, research phase in Large scope tasks, hypothesis analysis in `/crafter:debug` |

---

## BMAD Party Mode Integration

If the user's request involves any of the following, consider suggesting a BMAD party mode session first:

- Architectural decisions with significant tradeoffs
- Choosing between multiple valid approaches
- Cross-domain impact analysis
- Brainstorming or ideation sessions

**Crafter works completely independently of BMAD.** This is a recommendation, not a requirement. After a BMAD session, the user can feed the conclusions into `.planning/` files and proceed with `/crafter:do`.

---

## General Principles

- **Never commit without explicit user approval.**
- **When uncertain, ask — don't guess or assume.**
- Plans are written for humans: conversational, clear, and reasoned.
- Show your reasoning — explain why, not just what.
- Verification criteria are defined during planning, not after the fact.
- One logical change = one commit.
- Respect the existing code style and conventions of the project.
