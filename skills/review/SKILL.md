---
name: review
description: >
  Performs thorough code review using a dedicated review agent. Use this skill
  whenever the user wants to check code quality, requests a review, asks for
  feedback on an implementation, or wants to improve their code. The skill
  spawns a custom code-review agent that NEVER modifies files, compiles, or
  runs tests.
argument-hint: "[project or files]"
user-invocable: true
---

# Code Review Skill

This skill uses a custom `code-review` agent (via the `Agent` tool) to perform
reviews. The orchestrator's job is to prepare context and spawn the agent — the
heavy review work happens in a separate, clean context.

## When to use

- The user wrote or modified code and wants feedback
- "Review this", "check my code", "what do you think about this implementation"
- Before a merge request / pull request
- During refactoring — verifying the approach is sound

## What the review agent does

- Analyzes readability, variable and function naming
- Reviews architecture and code structure
- Looks for potential security issues (SQL injection, XSS, etc.)
- Flags performance problems (O(n²), unnecessary DB queries, etc.)
- Comments on adherence to principles (DRY, SOLID, separation of concerns)
- Evaluates code complexity against language-specific rules
- Suggests concrete improvements with code examples

## What the review agent NEVER does

- Does not run code or terminal commands
- Does not compile source files
- Does not run tests (that is a separate agent's job)
- Does not write or modify files

---

## Instructions for Claude (orchestrator)

When performing a review, follow these steps:

### 1. Parse skill arguments

The skill may receive arguments via `/review <args>`. Parse them as follows:

- **Project name** (e.g. `/review rust`, `/review frontend-v2`) — use as `--project=<name>`,
  skip to step 3.
- **File paths** (e.g. `/review src/main.rs src/lib.rs`) — pass as positional args to the
  preparation script in step 3.
- **No arguments** (`/review`) — proceed to step 2 to determine the project interactively.

### 2. Gather context and determine project

Ask the user (if not already clear from context):

- What to review (specific files, a branch, staged changes, etc.)
- Focus area if any (security, performance, readability, architecture)

If the project is still unknown, run:

```bash
.claude/skills/review/scripts/prepare-context.sh --list-projects
```

### 3. Run the preparation script

```bash
# Single project / after choosing project:
.claude/skills/review/scripts/prepare-context.sh --project=<name> [file1 file2 ...]
```

### Handling script output

The script output always starts with a `== SECTION ==` header. Handle each:

- **`== ASK_USER ==`** — The next lines contain JSON for `AskUserQuestion`.
  Pass it directly to `AskUserQuestion`. Then:
  - If followed by `== ON_YES ==`, run the command on that line **using the Bash tool**
    when user selects "Yes". On "No", stop.
  - If followed by `== OTHER_PROJECTS ==`, and user selects "Other", display the
    listed projects as text and ask user to specify. Then re-run the script with
    `--project=<selected project>`.
  - Otherwise, use the selected label as `--project=<label>` and re-run the script.
- **`== AGENT_INSTALLED ==`** — Tell user to restart Claude Code and run /review again.
- **`== PROJECT ==`** — Normal review output follows (LANGUAGES, RULES, FILES).
  Proceed to spawning the review agent.

### 4. Spawn the code-review agent

Use the `Agent` tool with `subagent_type: "code-reviewer"` to spawn the review agent.
Compose the prompt from the script output and the user's focus area:

```
Agent(
  subagent_type: "code-reviewer",
  prompt: """
    Review the following code changes.

    Focus: <user's focus area, or "general review" if not specified>

    <paste RULES section from script output — these are the project's
     language-specific review criteria that MUST be checked>

    <paste FILES section — the agent will read and review these files>

    Output format for each finding:
    - **File:line** — exact location
    - **Severity** — 🔴 critical / 🟡 moderate / 🟢 minor
    - **Description** — what the problem is and why it matters
    - **Suggestion** — concrete fix or improvement

    End with a brief Summary (2-3 sentences overall impression).
  """
)
```

### 5. Present the results

Display the agent's review output to the user. If the user wants to go deeper
on specific findings, spawn the agent again with a focused follow-up prompt.

---

## Adding language-specific rules

To add rules for a new language, create a file in the `assets/` folder following the
naming convention:

```
{language}-{topic}.md
```

Examples: `assets/elm-complexity.md`, `assets/ruby-style.md`, `assets/elixir-patterns.md`

The `prepare-context.sh` script will automatically pick up matching files based
on detected languages. Supported language mappings:

| Extensions                              | Language   |
| --------------------------------------- | ---------- |
| `.elm`                                  | elm        |
| `.rb`, `.rake`                          | ruby       |
| `.ex`, `.exs`, `.eex`, `.heex`, `.leex` | elixir     |
| `.js`, `.jsx`                           | javascript |
| `.ts`, `.tsx`                           | typescript |
| `.rs`                                   | rust       |
| `.py`                                   | python     |
| `.sh`, `.bash`                          | bash       |
| `.css`, `.scss`                         | css        |

---

## Notes

- If the user wants to run tests → delegate to the `test-runner` skill
- If the user wants to fix the code → the orchestrator can do that directly
- The review agent is **read-only** and **stateless** — each call is independent
- The agent source of truth is `assets/code-review-agent.md` — the copy in
  `.claude/agents/` is overwritten by `--install-agent`. Do not edit it directly.
