# Task: Fix Verifier agent still using Bash cat + temp files despite constraints

## Metadata
- **Date:** 2026-02-26
- **Branch:** main
- **Status:** completed
- **Scope:** Small

## Request
The crafter-verifier agent almost always uses `cat > /tmp/verification_summary.txt << 'EOF'` to write its verification report, despite constraints added in v0.4.1 (commit c58ad0f) that say "prefer native tools" and "do not create temporary files." The instructions aren't effective enough — need a stronger approach to prevent this behavior.

## Plan

- [x] Step 1: Update `rules/delegation.md` — change verifier model from `haiku` to `sonnet`
- [x] Step 2: Restructure `agents/crafter-verifier.md` — add Critical Rules section after Role, strengthen constraint language (NEVER + specific examples), repeat constraint in Output format section
- [x] Step 3: Add reinforcement to orchestrator spawn instructions in `commands/do.md` and `commands/debug.md`

## Decisions
<!-- Key decisions made during the workflow -->

## Outcome
<!-- To be filled after completion -->
