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
- Wait for the user's assessment before moving on.

## Scope Detection

| Scope | Characteristics | Workflow |
|---|---|---|
| **Small** | 1–3 files, clear intent, isolated change | Direct plan → execute → verify → review → commit |
| **Medium** | Multiple files, clear intent, cross-cutting | Plan with numbered steps → execute and verify and review per step |
| **Large** | Vague request, architectural impact, many files, unfamiliar territory | Research/discuss first → plan with steps → execute per step → verify per step → review per step |

When scope is ambiguous, ask the user rather than guessing.
