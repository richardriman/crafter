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
| `crafter-step-runner` | Glue delegate for three `/crafter-do` steps — extension-skill discovery, Step 0 resume lookup, Step 1 scope assessment — returns a structured routing-relevant summary; never edits task files, branches, or commits |

## Model Configuration

When spawning agents via the Task tool, pass the `model` parameter according to this table:

| Agent | Model | Rationale |
|---|---|---|
| `crafter-planner` | `opus` | Deep reasoning for plan quality |
| `crafter-implementer` | `sonnet` | Good balance of code quality and speed |
| `crafter-verifier` | `sonnet` | Needs reliable instruction-following for output constraints |
| `crafter-reviewer` | `opus` | Thorough code analysis |
| `crafter-analyzer` | `sonnet` | Adaptive — use `opus` for large scope tasks that require deeper research |
| `crafter-step-runner` | `sonnet` | Lightweight instruction-following for structured lookup and assessment steps |

Always include the `model` parameter in every Task tool invocation. Do not rely on model inheritance from the orchestrator.

Agent files also include a fallback `model` for direct invocation (`/agents` without orchestrator). Orchestrator-provided `model` still takes precedence and remains the source of truth.

## Skillbook — Learned Guidelines

Before spawning any agent via the Task tool, check if the `crafter` CLI binary is available at `{CRAFTER_HOME}/bin/crafter` (or `.claude/crafter/bin/crafter` for local installs). If available:

1. Resolve `SKILLBOOK_FILE`:
   - Prefer `{PROJECT_PATH}/{CRAFTER_DIR}/skillbook.json` when `CRAFTER_DIR` is available.
   - Fallback for older contexts: if `CRAFTER_DIR` is not resolved, use `{PROJECT_PATH}/.crafter/skillbook.json` if that directory exists; otherwise use `{PROJECT_PATH}/.planning/skillbook.json`.
2. Run via Bash: `{CRAFTER_HOME}/bin/crafter skillbook get --agent <agent-short-name> --file <SKILLBOOK_FILE>`
3. Agent name mapping: strip the `crafter-` prefix (e.g., `crafter-implementer` -> `implementer`, `crafter-planner` -> `planner`).
4. If the command produces output (non-empty stdout), append it verbatim to the agent's task prompt. The output is already formatted as a "Learned Guidelines" markdown section.
5. If the command produces no output, the agent has no learned guidelines — proceed normally without mentioning it.
6. If the command fails (non-zero exit), log a warning but proceed with agent spawning — skillbook is optional.

If the CLI binary does not exist, skip skillbook injection silently.

## Skill Directives — Caveman and Ponytail

Before spawning any agent via the Task tool, re-read the caveman and ponytail markers immediately before spawning the agent — a fresh read, not the startup-cached state (see `rules/core.md` — **Skill Detection: Caveman and Ponytail**):

1. **Caveman (all agents, audience-based level):** If caveman is active, append the directive below, choosing the level by the agent's audience:
   - **caveman-full** for `crafter-implementer`, `crafter-planner`, `crafter-analyzer`, and `crafter-step-runner` — their output is agent-facing (the orchestrator consumes/digests it).
   - **caveman-lite** for `crafter-reviewer` and `crafter-verifier` — their reports are relayed verbatim to the user (see `rules/core.md` carve-out (a)), so they are human-facing and must stay in the lighter register.

   Append (replace `<LEVEL>` with `full` or `lite` per the above):

   ```
   ## Active skill directives

   **caveman-<LEVEL>** is active — apply caveman-<LEVEL> discipline to your reasoning and returned report. Drop filler, pleasantries, and hedging in whatever language you use (language-specific mechanics like dropping articles apply only where the language has them). Keep ALL technical substance verbatim: code, file paths, identifiers, numbers, and every required field, heading, and table of your mandated output format — compress only the free-text prose within them.

   **Never compress:** security warnings; confirmations of irreversible actions; multi-step sequences where order or completeness matters; and any deviation/discovery or classification text bound for a buffer entry (`[uat-worthy]`/`[gap-worthy]`, auto-routing lines) — that text is rendered into the PR body by `crafter pr-body` and must stay in neutral human voice.
   ```

2. **Ponytail (`crafter-implementer` and `crafter-planner` only):** If ponytail is active and the target agent is `crafter-implementer` or `crafter-planner`, add the ponytail line below. If the caveman directive above already emitted the `## Active skill directives` block, append this line to that same block (do not repeat the header). Otherwise emit the block yourself — the `## Active skill directives` header followed by this line:

   ```
   ## Active skill directives

   **ponytail** is active at level `<LEVEL>` — apply YAGNI / the-ladder / shortest-working-diff discipline. Definition: `rules/core.md` § Ponytail.
   ```

   Replace `<LEVEL>` with the level read from `$HOME/.claude/.ponytail-active`. Do not append this line for any other agent (reviewer, verifier, analyzer, step-runner).

3. **No-op:** When both markers are absent, append nothing and do not mention the skills.
