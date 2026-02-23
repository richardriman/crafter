# Task: Context Window Optimization

## Metadata

- **Date:** 2026-02-23
- **Branch:** main
- **Scope:** Medium
- **Status:** active

## Request

Optimize context window usage in the orchestrator for multi-step tasks. The orchestrator accumulates full subagent outputs across every Execute-Verify-Review cycle, eventually exhausting the window. Four targeted changes agreed with user after discussion/analysis phase.

## Plan

- [ ] Step 1: Compact Verifier output format — replace table with summary line + FAILs only (`meta-prompts/verify.md`)
- [ ] Step 2: Add Step 6a compaction instruction for orchestrator — replace verbose Implementer/Verifier outputs with one-liners after step completion (`commands/do.md`)
- [ ] Step 3: Use targeted Edit instead of Write for task file checkboxes (`rules/task-lifecycle.md`)
- [ ] Step 4: Add staging constraint to Planner — split plans over 5 steps into self-contained stages (`meta-prompts/planner.md`)

## Decisions

- Reviewer output is never compacted — review is core to Crafter philosophy
- Auto-compaction is disabled by user preference — all savings must be structural

## Outcome

