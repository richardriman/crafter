# Project

> Crafter is a lightweight, human-in-the-loop AI development methodology for Claude Code — a set of conventions, context files, and skills that keep the developer in control at every step.

## Stack

- **Language:** Markdown (skills, rules, agents, templates, documentation)
- **Language:** Go (CLI utility binary — `cli/` directory)
- **Scripting:** Bash (install.sh only)
- **Runtime platform:** Claude Code CLI (custom skills)
- **Framework:** cobra (Go CLI only — core project remains pure conventions-and-prompts)

## Conventions

- **Naming:** kebab-case for filenames, UPPER-CASE for planning template files
- **Structure:** top-level directories by function (skills, agents, rules, templates, docs, cli)
- **Commits:** conventional commits format (feat/fix/refactor/docs/chore/test)
- **Language:** all persistent files in English; conversational output matches user's language
- **Source vs install:** When working on Crafter itself, always modify source files in the repository (`skills/`, `agents/`, `rules/`, `templates/`, `cli/`). Never modify installed copies in `~/.claude/crafter/`.

## Key Decisions

| Date | Decision | Reason |
|---|---|---|
| 2026-02-19 | Orchestrator/agent architecture | Prevents context rot in long conversations — each agent starts with fresh context |
| 2026-02-19 | Two installation modes (global/local) | Global is convenient for solo developers; local is committable and team-shareable |
| 2026-02-19 | Adaptive scope detection in /crafter:do | One command handles everything from one-line fixes to cross-cutting refactors |
| 2026-03-24 | Go CLI binary for deterministic utilities | LLMs do JSON CRUD, Jaccard similarity, and atomic writes poorly — a static binary with zero runtime deps handles these reliably |
| 2026-05-10 | PR composer rendered as Go CLI subcommand (`crafter pr-body`) | Mirrors GH#16 buffer pattern — deterministic NDJSON→markdown rendering belongs in the binary, not in LLM prose |
| 2026-06-01 | Plan-progress statusline via `crafter statusline`; installer wires it by default on every install; `--with-statusline` opt-in flag removed (hard-errors if passed) | Claude Code's native status bar integration; silent-fail posture and set-if-absent wiring avoids breaking existing setups; default-on removes a manual step that was easy to overlook |
| 2026-06-05 | SessionStart update-check hook ported from `hooks/crafter-check-update.js` to Go subcommand `crafter check-update`; `install.sh` registers the Go binary directly | Eliminates Node.js as a runtime dependency of the hook; 4h cache window; silent-fail and non-blocking; byte-compatible cache contract |
| 2026-06-02 | Statusline fallback cascade (rungs 2–4): completed-branch → `crafter · ✓ done`; active-elsewhere count → `crafter · N active elsewhere`; else empty; resolver strict to `- **Work branch:**` only (`- **Branch:**` alias explicitly declined) | Avoids silent segment when there is useful state; count-only format keeps bar short; strict single field keeps resolver simple and the non-standard case is a documented known limitation |
| 2026-06-03 | Statusline expanded from a single plan segment to a full multi-section panel `plan │ model │ vcs │ ctx │ cost`, each section degrading independently | One status bar surfaces plan position, model/effort, repo+diff, context use, and cost together; independent degradation keeps the silent-fail posture and never collapses to empty merely because no task is active |
| 2026-07-01 | Caveman/ponytail skill detection via marker files (`$HOME/.claude/.caveman-active` / `.ponytail-active`); propagated to agents at spawn via `rules/delegation.md`; caveman is audience-based (lite=human-facing, full=agent-facing regardless of marker level); ponytail scoped to `crafter-implementer` and `crafter-planner` only | Subagents don't receive SessionStart hook injection; orchestrator-side detection + single pre-spawn propagation rule ensures all agents honor the user's active disciplines without duplicating policy |
