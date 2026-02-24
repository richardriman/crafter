# Agent Delegation

Crafter commands run as **orchestrators**: the main context window manages the workflow and communicates with the user, while specialized agents do the actual work in fresh context windows.

Use the Task tool to spawn agents. Each agent is defined in the `agents/` directory (`crafter-planner`, `crafter-implementer`, etc.). Each agent:
- Runs as a native agent with its own tools (Read, Grep, Glob, Bash, etc.)
- Receives a task description and high-level pointers — it explores the codebase itself
- Returns a structured result to the orchestrator
- Has no memory of previous steps or other agents

The orchestrator is a **dispatcher** only: it manages workflow, communicates with the user, and delegates. It never reads code, never implements, and never reviews. It only holds: the current plan, the status of each step, and result summaries from agents.

## Agent Roles

| Agent | Role |
|---|---|
| `crafter-planner` | Researcher + architect — explores codebase deeply, produces implementation-ready plans with file:line references |
| `crafter-implementer` | Executor — mechanically executes the Planner's detailed plan |
| `crafter-analyzer` | Investigator — project mapping or research/investigation tasks |
| `crafter-verifier` | QA — checks verification criteria |
| `crafter-reviewer` | Code reviewer — reviews changes against plan and conventions |

## Model Configuration

When spawning agents via the Task tool, pass the `model` parameter according to this table:

| Agent | Model | Rationale |
|---|---|---|
| `crafter-planner` | `opus` | Deep reasoning for plan quality |
| `crafter-implementer` | `sonnet` | Good balance of code quality and speed |
| `crafter-verifier` | `haiku` | Lightweight read-and-check task |
| `crafter-reviewer` | `opus` | Thorough code analysis |
| `crafter-analyzer` | `sonnet` | Adaptive — use `opus` for large scope tasks that require deeper research |

Always include the `model` parameter in every Task tool invocation. Do not rely on model inheritance from the orchestrator.
