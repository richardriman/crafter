# Skill Contract

> Status: Draft
> Date: 2026-05-16

A **Skill Contract** is a structured description that every Crafter-compatible skill should declare. It tells orchestrators, agents, and human reviewers exactly what the skill does, what it needs, what it produces, and where its boundaries lie — without dictating how the skill is implemented.

This document defines the eight contract fields, explains what belongs in each one, and shows a short illustrative example.

---

## The Eight Contract Fields

### 1. Capability

A concise, plain-language statement of what the skill does. One or two sentences at most. This is the human-readable headline — it should answer "what does this skill accomplish?" without reference to implementation details.

*Belongs here:* the core function of the skill, the primary user outcome.  
*Does not belong here:* trigger conditions, input requirements, or implementation choices.

---

### 2. When-Applies

The conditions under which this skill is the right tool to use. This may describe the type of task, the project context, the user intent, or combinations thereof. An orchestrator uses this field to decide whether to invoke the skill.

*Belongs here:* scope triggers, context signals, task patterns that match this skill.  
*Does not belong here:* how the skill works once selected; that is capability or outputs.

---

### 3. Required Inputs

The concrete inputs the skill must receive before it can execute. List each input by name and describe what it represents. Mark optional inputs clearly.

*Belongs here:* named parameters, environment values, files, flags, or context documents the skill reads.  
*Does not belong here:* internal variables computed by the skill itself.

---

### 4. Outputs

What the skill produces when it succeeds. List each output by name, type, and destination (e.g., a file written to a specific path, a commit, a message returned to the orchestrator).

*Belongs here:* artifacts, files, state changes, and return values the skill is responsible for producing.  
*Does not belong here:* side effects (those go in Allowed or Forbidden Side Effects).

---

### 5. Allowed Side Effects

Actions the skill is permitted to take beyond its primary outputs. Explicitly listing allowed side effects prevents ambiguity — if it is not listed here it is implicitly forbidden.

*Belongs here:* filesystem reads and writes outside the primary output path, spawning sub-agents, reading runtime configuration, appending to logs or buffers.  
*Does not belong here:* anything that would violate the safety envelope (see Forbidden Side Effects).

---

### 6. Forbidden Side Effects

Actions the skill must never take, regardless of the task or instructions. This is the **safety envelope** — the inviolable set of constraints that protect project integrity, user trust, and system stability.

Every Crafter-compatible skill inherits the following base forbidden side effects:

- **No silent commits** — the skill must not create a git commit without explicit user approval (or under an `--auto` run that has satisfied the green-commit invariant).
- **No out-of-scope file writes** — the skill must not write, overwrite, or delete files outside the paths declared in Outputs and Allowed Side Effects.
- **No credential exfiltration** — the skill must not read, log, or transmit secrets, API keys, or tokens beyond the context window boundary.
- **No destructive rewrites** — the skill must not rewrite or replace content that was not part of the approved change request, even if the existing content appears incorrect.
- **No scope expansion** — the skill must not implement features, refactors, or improvements that were not included in the approved plan step, even if they seem beneficial.
- **No bypassing approval gates** — the skill must not mark tasks complete, close review loops, or advance workflow state without satisfying the criteria required by those gates.
- **No home-directory runtime modification** — the skill must not modify installed runtime copies under `~/.claude` or equivalent user-level runtime directories.

The seven items above are the minimum baseline that every Crafter-compatible skill must respect; they are not an exhaustive list. A skill may declare additional forbidden side effects specific to its domain — these extend (never narrow) the base set. Field 6 in a skill's own contract is the per-skill declaration summary; the [Safety Envelope](#safety-envelope) section below is the full set of behavioral guardrails that applies to all skills regardless of domain.

---

### 7. Success Criteria

The observable conditions that confirm the skill completed its work correctly. These criteria must be checkable — either by automated verification or by a human reviewer following the description.

*Belongs here:* file existence checks, expected content patterns, exit conditions, test results, diff shapes.  
*Does not belong here:* aspirational quality statements that cannot be confirmed objectively.

---

### 8. Failure Criteria

The observable conditions that indicate the skill failed or must not proceed. Listing failure criteria explicitly allows orchestrators and agents to halt early and hand off cleanly rather than continuing into a bad state.

*Belongs here:* missing required inputs, conflicting state, unrecoverable errors, gate conditions the skill cannot satisfy.  
*Does not belong here:* transient errors that the skill is expected to retry and recover from internally.

---

## Safety Envelope

The safety envelope is the set of guarantees that extension skills must not weaken under any instruction. It supplements the base forbidden side effects in field 6 and applies to every Crafter-compatible skill without exception.

### Forbidden behaviors

An extension skill must never:

1. **Replace or shadow a core agent** — Skills must not impersonate, override, or silently substitute for the Analyzer, Planner, Implementer, Verifier, or Reviewer. Core-agent identities are fixed in `skills/crafter-do/SKILL.md`.

2. **Bypass plan-approval or commit-approval gates** — Skills must not advance workflow state past an approval gate without a recorded human (or authorized `--auto`) decision. The APPROVE gate and the green-commit invariant are defined in `rules/do-workflow.md`.

3. **Mutate task files or run-directory buffers directly** — Skills must not write to `.crafter/tasks/*.md` or `.crafter/run/<task-id>/*` except through the `crafter-buffer` skill or other sanctioned skills (sanctioned skills are those shipped under `skills/` in the Crafter core repo). The task file is the orchestrator's state artifact; unauthorized writes corrupt handoff context.

4. **Rewrite or replace an approved plan** — Once the APPROVE gate closes, the plan is sealed. Skills must not alter plan content, reorder steps, or amend phase contracts without re-entering the PLAN → APPROVE cycle.

5. **Disable, skip, or short-circuit drift checks, phase verification, or review** — The VERIFY and REVIEW gates in `rules/do-workflow.md` are mandatory. Skills must not suppress drift classifications, skip verification evidence, or omit the Reviewer's diff summary and issue tables.

6. **Perform destructive git operations** — Skills must not force-push, rewrite history (`git rebase -i`, `git reset --hard` on shared branches), delete remote branches, or perform any operation that would break the green-commit invariant defined in `rules/do-workflow.md`.

7. **Override or contradict Karpathy guardrails** — Skills must not introduce speculative abstractions, make surgical-change violations, or skip goal-driven verification. The guardrails (Think Before Coding, Simplicity First, Surgical Changes, Goal-Driven Execution) are defined in `rules/core.md`.

8. **Write persistent files in a language other than English** — All content written to `.crafter/*`, plans, skill contracts, and similar persistent artifacts must be in English. The language rules are defined in `rules/core.md`.

### Required passthroughs

An extension skill must always:

1. **Surface approval decisions to the orchestrator** — Approval gate outcomes (user confirmation, plan rejection, `--auto` retained-gate triggers) must be passed through unchanged. Skills must not swallow, reinterpret, or silently discard these signals.

2. **Write state through sanctioned skills** — All task-file and buffer state changes must go through the `crafter-buffer` skill or other sanctioned skills (those shipped under `skills/` in the Crafter core repo). Skills must not construct raw file writes that bypass the buffer skill's format, schema, or append semantics.

3. **Respect the run-directory lifecycle** — Skills must not create, rename, or delete run directories unilaterally. The run-directory lifecycle — created by the orchestrator at Step 4, cleaned up on PR success or workspace teardown — is defined in `rules/do-workflow.md`.

4. **Forward blocker signals to the orchestrator** — When a skill hits an irrecoverable state it cannot resolve internally, it must emit a blocker signal and halt rather than guessing, fabricating data, or continuing into a corrupted state.

5. **Maintain the green-commit invariant** — Skills must not commit, push, or create a PR unless the orchestrator has authorized it after all VERIFY and REVIEW gates passed. The invariant is binding for all `--auto` runs and defined in `rules/do-workflow.md`.

6. **Respect the active step contract's scope** — Skills must not implement future steps, adjacent improvements, or speculative features beyond the outcome declared in the current Karpathy Contract. The scope boundary and non-goals fields are authoritative.

7. **Apply language rules uniformly** — Skills must produce English output in all persistent artifacts and match the user's language in live conversational output, as defined in `rules/core.md`.

Skills that cannot satisfy all safety-envelope items must declare that incompatibility explicitly in their contract and escalate to the orchestrator rather than proceeding.

---

## Illustrative Example

The following example shows what a contract block looks like for a hypothetical `crafter-scaffold` skill that generates a new command stub.

```markdown
---
name: "crafter-scaffold"
description: "Generate a new Crafter command stub from a template"
---

## Skill Contract

### Capability
Generates a ready-to-edit markdown stub for a new Crafter command, placing it in the
correct directory and pre-populating standard frontmatter and section headings.

### When-Applies
Use when the user asks to add a new slash command to Crafter and no existing command
stub covers the intended function. Not applicable when editing an existing command.

### Required Inputs
- `command-name` (required) — the kebab-case name for the new command (e.g., `crafter-init`)
- `description` (required) — one-sentence description to embed in the frontmatter
- `target-dir` (optional, default: `commands/`) — destination directory for the stub

### Outputs
- `commands/<command-name>.md` — new command stub file

### Allowed Side Effects
- Reads `templates/command-stub.md` to source the stub template
- Reads existing files in `commands/` to check for name collisions

### Forbidden Side Effects
- Inherits all base forbidden side effects from the Crafter skill contract
- Must not modify any existing file in `commands/`
- Must not write outside `commands/` or `templates/`

### Success Criteria
- `commands/<command-name>.md` exists and contains valid YAML frontmatter
- Frontmatter `name` field matches the supplied `command-name`
- File passes a basic markdown lint check

### Failure Criteria
- `command-name` is empty or contains characters invalid for a filename
- A file with the same name already exists in `commands/`
- The template file `templates/command-stub.md` cannot be read
```

This example is intentionally concise. Real skills may need more entries in each field, but the structure stays the same — eight named sections, descriptive prose, no implementation details.
