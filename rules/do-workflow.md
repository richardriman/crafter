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
- **Output format is mandatory** — reproduce the Reviewer's **Diff summary** and **Issues found** tables directly. Copy the markdown tables as-is. **Never** convert tables to prose, bullet lists, or any other format. Expected structure:

  ```
  ### Diff summary
  | File | Changes |
  |------|---------|
  | ...  | ...     |

  ### Issues found
  | # | Severity | File | Line | Description |
  |---|----------|------|------|-------------|
  | ...                                        |
  ```

  After the tables, state the recommendation (must-fix vs. optional).

- **STOP — ALWAYS wait for the user's response before proceeding, regardless of severity. Never auto-proceed when findings exist.** Only if there are literally zero findings: proceed automatically to the next workflow step.

- **After the user responds:** Critical or Major issues trigger the "Fix and re-review" (recommended) or "Proceed anyway" prompt. Only Minor/Suggestion findings → proceed.
- The `crafter-reviewer` agent produces a diff summary and issue report as part of its review output.
- Issues are categorized by severity (Critical, Major, Minor, Suggestion).
- Only Critical and Major issues trigger the fix loop.
- The review-fix loop runs a maximum of 3 iterations — a 4th iteration never starts.
- Minor issues and Suggestions are informational only.

## Scope Detection

| Scope | Characteristics | Workflow |
|---|---|---|
| **Small** | 1–3 files, clear intent, isolated change | Direct plan → execute → verify → review (with fix loop) → commit |
| **Medium** | Multiple files, clear intent, cross-cutting | Plan with numbered steps → execute and verify and review (with fix loop) per step |
| **Large** | Vague request, architectural impact, many files, unfamiliar territory | Research/discuss first → plan with steps → execute per step → verify per step → review (with fix loop) per step |

When scope is ambiguous, ask the user rather than guessing. However, if the user has already provided a clear, detailed request, do not ask them to repeat or clarify what they have already stated. Scope ambiguity means you cannot determine whether the change is Small/Medium/Large — it does not mean you need more information about the user's intent.
