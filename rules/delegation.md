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
| `crafter-planner` | Researcher + architect — explores enough context to produce vertical execution contracts with outcomes, boundaries, drift criteria, and verification evidence |
| `crafter-implementer` | Executor — implements the current approved step inside its contract and reports deviations/discoveries |
| `crafter-analyzer` | Investigator — project mapping or research/investigation tasks |
| `crafter-verifier` | QA — checks verification criteria |
| `crafter-reviewer` | Code reviewer — reviews changes against plan and conventions |

## Model Configuration

When spawning agents via the Task tool, pass the `model` parameter according to this table:

| Agent | Model | Rationale |
|---|---|---|
| `crafter-planner` | `opus` | Deep reasoning for plan quality |
| `crafter-implementer` | `sonnet` | Good balance of code quality and speed |
| `crafter-verifier` | `sonnet` | Needs reliable instruction-following for output constraints |
| `crafter-reviewer` | `opus` | Thorough code analysis |
| `crafter-analyzer` | `sonnet` | Adaptive — use `opus` for large scope tasks that require deeper research |

Always include the `model` parameter in every Task tool invocation. Do not rely on model inheritance from the orchestrator.

Agent files also include a fallback `model` for direct invocation (`/agents` without orchestrator). Orchestrator-provided `model` still takes precedence and remains the source of truth.

## Skillbook — Learned Guidelines

Before spawning any agent via the Task tool, check if the `crafter` CLI binary is available at `~/.claude/crafter/bin/crafter` (or `.claude/crafter/bin/crafter` for local installs). If available:

1. Resolve `SKILLBOOK_FILE`:
   - Prefer `{PROJECT_PATH}/{CRAFTER_DIR}/skillbook.json` when `CRAFTER_DIR` is available.
   - Fallback for older contexts: if `CRAFTER_DIR` is not resolved, use `{PROJECT_PATH}/.crafter/skillbook.json` if that directory exists; otherwise use `{PROJECT_PATH}/.planning/skillbook.json`.
2. Run via Bash: `~/.claude/crafter/bin/crafter skillbook get --agent <agent-short-name> --file <SKILLBOOK_FILE>`
3. Agent name mapping: strip the `crafter-` prefix (e.g., `crafter-implementer` -> `implementer`, `crafter-planner` -> `planner`).
4. If the command produces output (non-empty stdout), append it verbatim to the agent's task prompt. The output is already formatted as a "Learned Guidelines" markdown section.
5. If the command produces no output, the agent has no learned guidelines — proceed normally without mentioning it.
6. If the command fails (non-zero exit), log a warning but proceed with agent spawning — skillbook is optional.

If the CLI binary does not exist, skip skillbook injection silently.
