# Agent Delegation

Crafter commands run as **orchestrators**: the main context window manages the workflow and communicates with the user, while specialized agents do the actual work in fresh context windows.

Use the Task tool to spawn agents. Each agent is defined in the `agents/` directory (`crafter-planner`, `crafter-implementer`, etc.). Each agent:
- Runs as a native agent with its own tools (Read, Grep, Glob, Bash, etc.)
- Receives a task description and high-level pointers — it explores the codebase itself
- Returns a structured result to the orchestrator
- Has no memory of this run's other steps or agents, beyond its own project-scoped `.claude/agent-memory/<agent>/MEMORY.md`

**Git hygiene.** Agent memory is per-developer and must never appear in commits. The Crafter repo's own `.gitignore` already includes `.claude/agent-memory/`. Downstream projects MUST add `.claude/agent-memory/` to their `.gitignore`.

The orchestrator is a **dispatcher** only: it manages workflow, communicates with the user, and delegates. It never reads code, never implements, and never reviews. It only holds: the current plan, the status of each step, and result summaries from agents.

## Agent Roles

| Agent | Role |
|---|---|
| `crafter-planner` | Researcher + architect — explores enough context to produce execution contracts with outcomes, boundaries, seams, and verification evidence (Medium/Large scope only) |
| `crafter-implementer` | Executor — implements the approved contract it was given (the whole task for Small/Medium, one whole phase for Large), runs the relevant tests/typechecks, and reports evidence and deviations |
| `crafter-analyzer` | Investigator — project mapping or research/investigation tasks |
| `crafter-checker` | QA + code reviewer in one — checks drift against the contract and reviews the code in a single fresh-context pass (modes: full pass, delta pass) |

## Model Configuration

When spawning agents via the Task tool, pass the `model` parameter according to this table:

| Agent | Model | Effort | Rationale |
|---|---|---|---|
| `crafter-planner` | `fable` | medium | Strongest tier; medium effort keeps plans dense without padding |
| `crafter-implementer` | `fable` | low | Strongest tier; low effort curbs verbosity in reports and diffs |
| `crafter-checker` | `fable` | medium | Strongest tier; medium effort for drift and review without invented findings |
| `crafter-analyzer` | `fable` | low | Strongest tier; low effort keeps research output terse |

`fable` replaced `opus`, which proved verbose and prone to invent findings; effort is set low/medium to curb verbosity and padding.

Always include the `model` parameter in every Task tool invocation. Do not rely on model inheritance from the orchestrator.

The `Effort` column is **documentary only** — the Task tool has no `effort` parameter. Each agent's effort is honored automatically from the `effort:` field in its own frontmatter (`agents/crafter-*.md`); the column records the intended tier so the two stay in sync. Do not attempt to pass effort at spawn time.

Agent files also include a fallback `model` for direct invocation (`/agents` without orchestrator). Orchestrator-provided `model` still takes precedence and remains the source of truth.

## Skill Directives — Caveman and Ponytail

Before spawning any agent via the Task tool, re-read the caveman and ponytail markers per `rules/core.md` — **Skill Detection: Caveman and Ponytail** (canonical), then:

1. **Caveman (all agents, audience-based level, English only):** If caveman is active **and** the detected user conversation language (per § Report Language / `rules/core.md` — **Language Rules**) is English, append the directive below, choosing the level by the agent's audience. For any other language append nothing for caveman (ponytail is unaffected): caveman's mechanics are English-specific, and for other languages the hard output-format limits in each agent's own prompt (e.g. `agents/crafter-checker.md` § Output format) do the work.
   - **caveman-full** for `crafter-implementer`, `crafter-planner`, and `crafter-analyzer` — their output is agent-facing (the orchestrator consumes/digests it).
   - **caveman-lite** for `crafter-checker` — its report is relayed verbatim to the user (see `rules/core.md` carve-out (a)), so it is human-facing and must stay in the lighter register.

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

   Replace `<LEVEL>` with the level read from `$HOME/.claude/.ponytail-active`. Do not append this line for any other agent (checker, analyzer).

3. **No-op:** When both markers are absent, append nothing and do not mention the skills.

## Report Language (always on)

Independent of the caveman and ponytail markers, and emitted on **every** spawn: the orchestrator names the user's conversation language (as detected per `rules/core.md` — **Language Rules**) in the spawned agent's prompt, and the agent writes the free-text prose of its returned report — finding descriptions, recommendations, summaries — in that language. Only that final report prose switches language: the agent's reasoning, intermediate notes, tool queries, any scratch or plan text, and MEMORY.md entries stay English regardless of the user's language — English intermediate work is cheaper and terser, and translation happens once, at the report.

Always English regardless of the user's language: code, identifiers, file paths, required headings, table column names, status-line formats, every persistent file (`.crafter/*`, task files, plans — see `rules/task-lifecycle.md`), and — the same carve-out the caveman directive carries in item 1 above — any deviation/discovery or classification text bound for a buffer entry (`[uat-worthy]`/`[gap-worthy]`, auto-routing lines): `crafter pr-body` renders it into the PR body and the buffers under `.crafter/run/` are persistent files. Crafter targets no specific language; it follows whichever one the user writes in.

**No-filler rule (always on, language-independent).** The spawn prompt always carries the following, regardless of the caveman/ponytail markers and of the user's language — it replaces caveman for non-English users and is cheaper than a style directive:

- The returned report contains no greeting, no preamble or intro sentence, no closing summary or "overall" verdict, no pleasantries, and no hedging filler.
- Only the sections the agent's output format mandates, in order.
- Free-text cells and paragraphs say what is needed in the fewest sentences.
