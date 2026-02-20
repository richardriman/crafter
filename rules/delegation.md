# Subagent Delegation

Crafter commands run as **orchestrators**: the main context window manages the workflow and communicates with the user, while specialized subagents do the actual work in fresh context windows.

Use the Task tool or `claude --print` to spawn subagents. Each subagent:
- Receives the appropriate meta-prompt from `meta-prompts/` as its system prompt
- Gets only the context it needs (relevant `.planning/` excerpts + task-specific files)
- Returns a structured result to the orchestrator
- Has no memory of previous steps or other subagents

The orchestrator never analyzes code, never implements, and never reviews. It only holds: the current plan, the status of each step, and result summaries from subagents.
