# Standard Change Workflow Rules

### PLAN
- Always run a lightweight completeness/refinement check before planning.
- If the request is not complete enough to plan, do targeted discussion and/or research before planning.
- A request is complete enough to plan when the goal, scope, non-goals, acceptance criteria, constraints, risks, and validation strategy are clear.
- Always propose a plan before taking implementation action.
- Write plans in plain, conversational language — not XML, not machine-readable syntax.
- Explain **why**, not just what.
- Write plans as vertical execution contracts, not concrete implementation scripts.
- Include step drift criteria and phase verification criteria (how you'll know the change is correct and still inside scope).
- For non-trivial changes, mention alternatives that were considered.
- Surface assumptions and ambiguous interpretations explicitly.
- Include a Karpathy Contract for each phase and step: outcome, scope boundary, non-goals, simplicity constraint, drift criteria, verification evidence, and stop conditions.

### APPROVE
- Never proceed without explicit user approval of the plan.
- If the user has concerns or requests changes, revise the plan and wait again.
- "Looks good" or "go ahead" counts as approval. Silence does not.

### EXECUTE
- Implement exactly what was approved.
- Never change architecture without prior discussion.
- If something unexpected is discovered mid-execution that would materially change the plan, stop and inform the user before continuing.
- Avoid speculative additions ("while we're here" features, abstractions, configurability) unless explicitly approved.
- Execute one step at a time. Do not implement future steps early.

### VERIFY
- After each step, run a lightweight step drift check against that step's Karpathy Contract.
- Step drift checks classify drift as: no drift, harmful drift, scope drift, beneficial local drift, or plan-obsoleting discovery.
- Harmful drift blocks the next step until the current step is fixed.
- Scope drift requires user approval or replanning.
- Beneficial local drift may continue only when recorded as an accepted decision.
- After all steps in a phase pass drift checks, run phase verification against the phase verification criteria.
- Run tests if applicable.
- Report clearly what passed, what failed, and what workflow action is required.
- Verify goals, not just activity: each criterion must map to observable evidence.

### REVIEW
- Run full review after phase verification passes, not after every step.
- Run full review after an individual step only when the step is high-risk: security/auth, data migration, public API, architecture, concurrency, destructive behavior, or a verifier concern.
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

  ### Karpathy scorecard
  | Principle | Status | Evidence |
  |-----------|--------|----------|
  | Think Before Coding | PASS/FLAG | ... |
  | Simplicity First | PASS/FLAG | ... |
  | Surgical Changes | PASS/FLAG | ... |
  | Goal-Driven Execution | PASS/FLAG | ... |
  ```

  After the tables, state the recommendation (must-fix vs. optional).

- **STOP — ALWAYS wait for the user's response before proceeding, regardless of severity. Never auto-proceed when findings exist.** Only if there are literally zero findings: proceed automatically to the next workflow step.

- **After the user responds:** Critical or Major issues trigger the mandatory fix loop — there is no "Proceed anyway" choice for those severities. Only Minor/Suggestion findings → proceed.
- The `crafter-reviewer` agent produces a diff summary and issue report as part of its review output.
- Issues are categorized by severity (Critical, Major, Minor, Suggestion).
- Only Critical and Major issues trigger the fix loop.
- The review-fix loop runs a maximum of 5 iterations — a 6th iteration never starts automatically. If the cap is reached with Critical/Major findings still present, the orchestrator stops and asks the user to choose:
  - **(a) manual override** — authorize manual iteration beyond the cap; the orchestrator re-enters the fix loop only on explicit user instruction.
  - **(b) accept-without-commit** — accept the unresolved findings and proceed without committing this phase; record a Decision entry noting the unresolved findings and that the green-commit invariant is deliberately broken for this phase.
  - **(c) replan-and-abort** — abandon the current phase and return to planning.
- Minor issues and Suggestions are informational only.

## Scope Detection

| Scope | Characteristics | Workflow |
|---|---|---|
| **Small** | 1–3 files, clear intent, isolated change | Completeness check → contract plan → execute step(s) → drift check per step → phase verification → phase review (with fix loop) → commit |
| **Medium** | Multiple files, clear intent, cross-cutting | Completeness check → contract plan with vertical phase(s) → execute one step at a time → drift check per step → phase verification and review per phase |
| **Large** | Incomplete/vague request, architectural impact, many files, or unfamiliar territory | Completeness check → research/discuss until complete → contract plan with vertical phases → execute one step at a time → drift check per step → phase verification and review per phase |

When scope is ambiguous, ask the user rather than guessing. However, if the user has already provided a clear, detailed request, do not ask them to repeat or clarify what they have already stated. Scope ambiguity means you cannot determine whether the change is Small/Medium/Large — it does not mean you need more information about the user's intent.
