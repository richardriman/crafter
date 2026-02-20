# Standard Change Workflow Rules

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
- Never change architecture without prior discussion.
- If something unexpected is discovered mid-execution that would materially change the plan, stop and inform the user before continuing.

### VERIFY
- Check each verification criterion defined in the plan.
- Run tests if applicable.
- Report clearly what passed and what (if anything) did not.

### REVIEW
- Show diffs after execution.
- Highlight any deviations from the approved plan, even minor ones.
- After receiving the review report, categorize issues by severity.
- If Critical or Major issues exist, present them to the user with two options: fix automatically or proceed anyway.
- If the user chooses to fix: delegate only the Critical/Major issues to the Implementer, then re-Verify and re-Review.
- Cap the review-fix loop at 3 iterations. After the 3rd review, present remaining issues and recommend the user decide manually.
- Minor issues and Suggestions are informational — they do not trigger the fix loop.
- Always wait for the user's assessment before moving on.

## Scope Detection

| Scope | Characteristics | Workflow |
|---|---|---|
| **Small** | 1–3 files, clear intent, isolated change | Direct plan → execute → verify → review (with fix loop) → commit |
| **Medium** | Multiple files, clear intent, cross-cutting | Plan with numbered steps → execute and verify and review (with fix loop) per step |
| **Large** | Vague request, architectural impact, many files, unfamiliar territory | Research/discuss first → plan with steps → execute per step → verify per step → review (with fix loop) per step |

When scope is ambiguous, ask the user rather than guessing.
