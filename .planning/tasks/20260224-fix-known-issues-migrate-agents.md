# Task: Fix known issues in do.md workflow + migrate meta-prompts to native agents

## Metadata
- **Date:** 2026-02-24
- **Branch:** main
- **Status:** active
- **Scope:** Large

## Request
Fix the known issues in do.md workflow (all 11 items from STATE.md Known Issues) and migrate meta-prompts to native `.claude/agents/` format as part of the same task.

## Plan

### Stage 1 — Create Native Agents (steps 1–5)

- [x] Step 1: Create `crafter-planner.md` in `agents/`. YAML frontmatter (name, description, tools: Read, Grep, Glob, Bash). Role: deep code explorer + plan producer. Removes $CONTEXT, adds "use Read/Grep/Glob to explore codebase". Explicitly uses ARCHITECTURE.md if provided (#7). Adds Medium scope granularity guidance (#2). Produces implementation-ready plans with file:line references.
- [x] Step 2: Create `crafter-implementer.md` in `agents/`. Tools: Read, Write, Edit, Bash, Grep, Glob. Role: mechanical execution of detailed plans. Removes $CONTEXT, receives plan in task prompt.
- [x] Step 3: Create `crafter-verifier.md` in `agents/`. Tools: Read, Grep, Glob, Bash. Removes $CONTEXT.
- [x] Step 4: Create `crafter-reviewer.md` in `agents/`. Tools: Read, Grep, Glob, Bash. Adds diff summary requirement (#8). Uses ARCHITECTURE.md if provided.
- [x] Step 5: Create `crafter-analyzer.md` in `agents/`. Tools: Read, Grep, Glob, Bash. Broadened: two modes — (A) Project Mapping, (B) Research/Investigation (#6).

### Stage 2 — Fix Workflow Files (steps 6–10)

- [x] Step 6: Fix `do.md` — task file for all scopes (#1), explicit Small V+R flow (#3), resume detection clarity (#5), ARCHITECTURE.md passing to Planner (#7), agent references, delegation model change (orchestrator = dispatcher).
- [x] Step 7: Fix `do-workflow.md` — iteration cap "3 iterations, no 4th" (#4), deduplicate REVIEW section (#9), align "show diffs" (#8).
- [x] Step 8: Fix `do.md` Step 6 iteration cap language (#4).
- [x] Step 9: Fix `map-project.md` — remove fallback annotations (#11), agent references.
- [x] Step 10: Rewrite `delegation.md` — agent-based model, remove `claude --print` (#10), orchestrator = dispatcher / Planner = researcher / Implementer = executor.

### Stage 3 — Infrastructure and Cleanup (steps 11–14)

- [x] Step 11: Fix `post-change.md` — ARCHITECTURE.md delegation clarity (#7), agent references.
- [x] Step 12: Update `install.sh` — copy from `agents/` to `$base/agents/` instead of meta-prompts.
- [x] Step 13: Update `tests/test_install.sh` — expected file list, run tests.
- [x] Step 14: Update ARCHITECTURE.md, update `debug.md` agent references, delete `meta-prompts/`.

## Decisions
<!-- Key decisions made during the workflow, in chronological order -->
- **Decision:** Combine known-issues fix with meta-prompts → agents migration. **Reason:** Several known issues (inconsistent delegation, missing tool restrictions, claude --print guidance) are resolved automatically by the migration.
- **Decision:** Switch delegation model from push (orchestrator pre-reads code, injects $CONTEXT) to pull (agents explore codebase themselves via Read/Grep/Glob). **Reason:** Orchestrator was doing research work that belongs to Planner/Implementer. Planner (Opus) does deep code exploration, produces implementation-ready plans; Implementer (Sonnet) executes mechanically.

## Outcome
<!-- Filled on completion: what was actually done, commit SHA(s), any deviations from plan -->
