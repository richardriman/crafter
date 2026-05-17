# State

## Current Focus

`crafter-do` core-capability decomposition Slice 3 complete on branch `refactor/crafter-do-slice-3-steps-1-2` (Steps 1–2 Scope+Discuss extracted; both phases delivered, verified, reviewed). Branch ready for PR. Slices 1–2 already on `main`.

## Recent Changes

| Date | Change | Commit |
|---|---|---|
| 2026-05-17 | `crafter-do` core-capability decomposition Slice 3 — `## Step 1 — Completeness and scope` + `## Step 2 — DISCUSS / RESEARCH …` extracted verbatim into `rules/do/step-1-scope.md` + `rules/do/step-2-discuss.md` (move-and-link, H2→H1; binding-preserving stubs matching the Step 0 convention), loader/installer/test wired ×2, `{CRAFTER_HOME}` policy applied (step-1 ×2, step-2 ×1), ledger extended. Pre-extraction binding gate GO (5/5); verifier 7/7 + 6/6 PASS; reviewer 1 Suggestion (fixed); test 45/0. | 9668a3b + 69f4220 |
| 2026-05-17 | `crafter-do` core-capability decomposition Slice 2 — `## Step 0 — Resume Detection` extracted verbatim into `rules/do/step-0-resume.md` (move-and-link, H2→H1; binding-preserving pointer stub retains heading + states resume state/branch-sanity/main-master guards established before Step 1), loader/installer/test wired, `{CRAFTER_HOME}` policy applied to the module's `task-lifecycle.md` ref, design-note ledger extended. Pre-extraction binding gate GO; verifier 7/7 + 6/6 PASS; reviewer no findings; test 45/0. | 53e6143 + 1319d28 |
| 2026-05-17 | `crafter-do` core-capability decomposition Slice 1 — Phase 1 design note (`docs/core-capabilities.md`: taxonomy + `{CRAFTER_HOME}` runtime-path policy), Phase 2 preamble extraction (Flag Validation / Project Resolution / Extension Skills → `rules/do/*`, byte-identical, installer deploys `rules/do/`), Phase 2 review #2 resolved (loader grouping comments), Phase 3 runtime-path hygiene (`extension-skills.md` global path → `{CRAFTER_HOME}/skills/`, sibling task delineated). Verifier 7/7 PASS, reviewer no findings. Follow-up slices (Step 0–9b modules) recorded in task Outcome. | 8133f12 + b2d74e8 + 7142673 |
| 2026-05-16 | Composable skill contracts — added `docs/skill-contract.md` with the eight-field Skill Contract and Safety Envelope, cross-linked plugin/architecture docs, wired `crafter-do` to discover compatible extension skills at Steps 1/4/6 as supplemental-only specialists, and codified the invariant in `rules/do-workflow.md` | ac1f40b + 5fe353f + e31096b |
| 2026-05-10 | Agent `--auto` semantics (GH#18) — Reviewer three-bucket classification (auto-fixable/uat/gap), Verifier routing metadata per recommendation, Implementer buffer-worthy discovery tagging; escape hatch placeholder in `do-workflow.md` replaced with agent-side signal documentation; completes `--auto` pipeline | 498ecec + c94f9c3 + a5d5b9f |
| 2026-05-10 | PR composer (GH#17) — `crafter pr-body` Go subcommand reads per-run NDJSON buffers and task file, renders `## Manual QA Plan`, `## Known Gaps`, `## Decisions` sections; Step 9b in `skills/crafter-do/SKILL.md` wires it into the `--auto` end-of-task flow; GH#16 forward references resolved | 679a06e + ebd48af + a74ab5e + 9b86594 |
| 2026-05-10 | `crafter-buffer` skill (GH#16) — NDJSON `.crafter/run/<task-id>/` buffers, Go subcommand `crafter buffer uat\|gap`, run-directory lifecycle in `rules/do-workflow.md`, `.crafter/run/` in `.gitignore`, GH#15 forward references resolved | a786116 |
| 2026-05-09 | `--auto` flag for crafter-do (GH#15) — unattended orchestration mode with binding green-commit invariant, four retained gates (initial clarification, plan approval, green-commit cap reached, ad-hoc escape hatch), and parser-level mutual exclusion with `--fast`. Phase 1 documented the contract in `rules/do-workflow.md`; Phase 2 wired the flag into `skills/crafter-do/SKILL.md` (frontmatter, Skill options, Flag Validation block, Step 6b restructure). Forward references to GH#16/#17/#18 for buffer skill, PR composer, and agent prompt updates. | 828815e + 308f828 |
| 2026-05-05 | Green commits + per-phase auto-commit — Critical/Major review findings mandatory to fix (5-iteration cap with three user choices on cap-hit), new Step 6b Phase Summary and Auto-Commit with three approval paths (auto / `--fast` silence / explicit), manual-verification override, consolidated end-of-task commit for docs+skillbook+STATE.md, new `--fast` metadata flag (default off) | 7d24131 |
| 2026-03-25 | Skillbook system — Go CLI binary (`crafter skillbook`) with get/add/init subcommands, prompt integration (delegation.md + post-change.md), install script binary download, documentation | 1600677 |
| 2026-03-06 | Workflow hardening — `git -C {PROJECT_PATH}` branch detection, English-only task files, mandatory post-change checklist, scope expansion rule | 5d46339 |
| 2026-03-06 | Smarter /crafter:do entry logic — multi-project workspace support (`--project` flag + auto-discovery), Grep-based resume detection with resume-intent words, guardrails against ignoring clear user input, `{PROJECT_PATH}` across all rule files | 47745a3 |
| 2026-03-06 | Review STOP gate — unmissable formatting for review findings | eb10a08 |
| 2026-02-27 | Install script cleans target directory before copying — removes stale files from previous versions on upgrade | 5c1d41a |
| 2026-02-27 | Planner writes full plan to task file (draft/approved lifecycle), richer summary, stricter review gate, resume handles all plan states | 7722b9f |
| 2026-02-26 | Released v0.5.0 — review workflow always waits for user, detailed table output, update check and release fixes | [v0.5.0](https://github.com/richardriman/crafter/releases/tag/v0.5.0) |
| 2026-02-26 | Verifier temp-file hardening — upgraded verifier instructions to prevent Bash `cat`/temporary-file verification reports; released v0.4.2 | cd5d03a |
| 2026-02-26 | Released v0.4.1 — agents prefer native tools over Bash, reducing permission prompts | [v0.4.1](https://github.com/richardriman/crafter/releases/tag/v0.4.1) |
| 2026-02-24 | Released v0.4.0 | [v0.4.0](https://github.com/richardriman/crafter/releases/tag/v0.4.0) |
| 2026-02-24 | Migrated meta-prompts to native agents, fixed all 11 known workflow issues, updated install + tests + docs | aec78c6 |
| 2026-02-23 | Moved update check to SessionStart hook (replaced inline rule with hooks/crafter-check-update.js) | 20acf60 |
| 2026-02-23 | Context window optimization — compact verifier output, session-break guidance, targeted checkbox edits, and planner staging for long plans | 85b642c + 7839161 |
| 2026-02-22 | Added test suite for install.sh (pure Bash, zero dependencies) | 8c302d3 |
| 2026-02-22 | Added internal `/crafter:release` command for GitHub Releases with AI-generated notes | 312b391 |
| 2026-02-22 | Added remote install via curl one-liner (auto-detection, --version flag, tarball download) | 68e54ad |
| 2026-02-21 | Added MIT licence | 976706f |
| 2026-02-20 | Automatic update detection — VERSION file, GitHub Releases API check, install-time VERSION copy, and follow-up `/crafter:release` idea | 850e59a + 34ebfb1 |

## Planned

- [x] ~~Refactor `crafter-do` into composable core capability modules — Slice 1 complete (`.crafter/tasks/20260517-refactor-crafter-do-core-capabilities.md`): design note + preamble extraction + runtime-path policy~~. **Follow-up slices** (Step 0–9b capability modules) listed in that task's `## Outcome` for the next planning session
- [ ] Optional project-level review rules — reviewer loads `.crafter/review-rules.md` (if present) as additional context, allowing projects to define language-specific, framework-specific, or team-specific review criteria
- [ ] Skills-first runtime portability — active parked task in `.crafter/tasks/20260421-skills-first-runtime-portability.md`; skills-first baseline and wrapper removal landed, but adapter/build/test rollout across VS Code, Copilot CLI, and OpenCode is not complete on `main`
- [ ] Model profiles — matice agent × profil (quality/balanced/budget) → model tier. Prompt-only, orchestrátor čte config a předává `--model` agentům. Inspirace: Nightshift `model-profiles.ts`
- [x] ~~Skillbook — self-learning agents~~ (implemented in 1600677 as Go CLI binary)
- [ ] Holdout validation — nezávislý agent ověří implementaci proti kritériím, která implementer neviděl. Informační bariéra čistě v prompt designu. Inspirace: Nightshift holdout pattern
- [ ] `--fast` auto-apply Minor recommendations — when reviewer's output contains a clear recommendation for a Minor/Suggestion finding, the orchestrator (under `--fast`) auto-applies the recommendation, documents it as accepted, and continues. Requires defining a stable "clear recommendation" detection pattern from the reviewer output.

## Ideas

- `/crafter:add-planned` — quick command for adding planned items to STATE.md
- Wonder/Reflect pattern (inspirace OctopusGarden) — dvou-fázová diagnostika při zaseknutí: Wonder (divergentní brainstorming neobvyklých příčin) → Reflect (chirurgický konzervativní fix). Temperature control přes `claude -p` není dostupný, ale dá se nahradit prompt engineeringem. Mohlo by obohatit `crafter:debug`.
- Holdout validation — verifier testuje proti kritériím, která implementer neviděl. Satisfaction scoring (0–100) místo binary pass/fail.

## Known Issues

None currently tracked.
