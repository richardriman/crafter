# Architecture Overview

## System Intent

Crafter is a workflow orchestration toolkit for Claude Code projects, combining markdown-defined skills, role-specific agents, and a small Go CLI utility.

## Primary Components

- `skills/` — orchestration entrypoints and workflow control logic
- `agents/` — role-specialized execution units (planner, implementer, verifier, reviewer, analyzer)
- `rules/` — reusable policy/rule fragments consumed by skills
- `cli/` — deterministic utility binary (`crafter`) for operations that are brittle in prompt-only flows
- `templates/` — initialization templates for `.crafter/` context files
- `tests/` — installer and regression validation

## Interaction Model

1. A skill command starts orchestration.
2. The orchestrator delegates specialized tasks to agents in fresh contexts.
3. Agents read repository context and propose/perform scoped actions.
4. Human approval gates control plan acceptance, fixes, and final commit/merge decisions.
5. The CLI supports repeatable low-level utilities where deterministic behavior is required.

## Operational Boundaries

- Source-of-truth edits occur in repository files, not installed runtime copies.
- Documentation and context updates should accompany behavior changes.
- Repository owner remains the final merge authority.
