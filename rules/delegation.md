# Subagent Delegation

Crafter commands run as **orchestrators**: the main context window manages the workflow and communicates with the user, while specialized subagents do the actual work in fresh context windows.

Use the Task tool or `claude --print` to spawn subagents. Each subagent:
- Receives the appropriate meta-prompt from `meta-prompts/` as its system prompt
- Gets only the context it needs (relevant `.planning/` excerpts + task-specific files)
- Returns a structured result to the orchestrator
- Has no memory of previous steps or other subagents

The orchestrator never analyzes code, never implements, and never reviews. It only holds: the current plan, the status of each step, and result summaries from subagents.

## Model Configuration

When spawning subagents via the Task tool, pass the `model` parameter according to this table:

| Subagent | Model | Rationale |
|---|---|---|
| Planner | `opus` | Deep reasoning for plan quality |
| Implementer | `sonnet` | Good balance of code quality and speed |
| Verifier | `haiku` | Lightweight read-and-check task |
| Reviewer | `opus` | Thorough code analysis |
| Analyzer | `sonnet` | Adaptive — use `opus` for Large scope tasks that require deeper research |

Always include the `model` parameter in every Task tool invocation. Do not rely on model inheritance from the orchestrator.
