# Task: Review workflow — always wait for user + show detailed findings

## Metadata
- **Date:** 2026-02-26
- **Branch:** main
- **Status:** active
- **Scope:** Medium

## Request
Two changes to the review step in the Crafter workflow:

1. **Always wait for user acknowledgment** after review — currently the orchestrator skips waiting when there are only Minor/Suggestion findings. The user wants to be consulted after every review, regardless of severity.

2. **Show all findings in detail** — the reviewer should list every individual occurrence rather than providing a summary. The output should be clear and easy to scan.

## Plan

- [ ] Step 1: Restructure Step 6 (REVIEW) in `commands/do.md` — auto-proceed only on zero findings, wait for user on any finding, then branch by severity
- [ ] Step 2: Strengthen REVIEW rules in `rules/do-workflow.md` — always wait when findings exist, auto-proceed only on clean review
- [ ] Step 3: Update reviewer output format in `agents/crafter-reviewer.md` — table (#, What, Where, Severity), recommendations below, every finding listed individually

## Decisions
<!-- Key decisions made during the workflow -->

## Outcome
<!-- To be filled after completion -->
