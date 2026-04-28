# Feature Spec: Skillbook Learning System

## Problem Statement

Agents need a lightweight, project-specific memory mechanism that improves future task execution without introducing noisy or duplicate guidance.

## Scope

This feature defines behavior for:
- storing project-scoped observations
- deduplicating similar observations
- promoting confidence based on repeated evidence
- retrieving top relevant learnings for agent prompts

## Non-Goals

- global cross-project memory
- opaque autonomous policy changes without review
- replacing explicit repository documentation

## Functional Requirements

1. Provide deterministic commands to initialize, read, and append skillbook data.
2. Deduplicate similar observations using Jaccard overlap heuristics.
3. Track confidence levels and allow promotion with repeated corroboration.
4. Return a prioritized subset of learnings for prompt injection.
5. Perform file writes atomically to avoid corruption.

## Data & Storage Constraints

- Skillbook file is project-local and versionable in project context directory.
- Stored content must exclude secrets or credentials.
- Schema evolution must remain backward-compatible or provide explicit migration path.

## Acceptance Criteria

- `crafter skillbook init` creates a valid empty skillbook structure.
- `crafter skillbook add` appends or merges an observation according to deduplication rules.
- Repeated similar observations increase confidence tier as defined by implementation.
- `crafter skillbook get` returns top-ranked items ordered by confidence and relevance.
- All write operations are atomic and leave no partial file state on failure.

## Verification Approach

- unit tests for Jaccard scoring, formatting, and storage behavior
- command-level tests for init/add/get flows
- manual spot checks that retrieved learnings improve downstream agent prompts

## Risks & Assumptions

- dedup threshold tuning may under-merge or over-merge observations
- low-quality observations can pollute recommendations if not reviewed
- assumes project context path is writable during command execution
