# Extension Skills

An **extension skill** is any Crafter-compatible skill — beyond the core agents shipped with Crafter — that declares a Skill Contract block in its own `SKILL.md`. Extension skills act as **supplemental specialists**: they advise, annotate, or enrich workflow phases, but they never replace a core agent or bypass an approval gate.

**v1 invariant — supplemental only.** Extension skills in v1 are advisory. They may contribute observations, checklists, or domain-specific findings, but all workflow decisions (plan approval, commit approval, phase gates) remain exclusively with the core orchestrator and core agents.

### Discovery

When the workflow starts, scan these `skills/` directories in priority order (most specific to least):

| Priority | Location | Scope |
|---|---|---|
| 1 (highest) | `{PROJECT_PATH}/.claude/crafter/skills/` | Current project |
| 2 | First `../.claude/crafter/skills/` found walking up parent directories | Parent project |
| 3 (lowest) | `~/.claude/crafter/skills/` | Global (user-wide) |

For each `skills/<skill-name>/SKILL.md` found, read it and check whether it contains a `## Skill Contract` section. If it does, the skill is Crafter-compatible and eligible for consideration. If the same skill name appears at multiple levels, the most specific one wins.

### Safety envelope

Every extension skill must satisfy the full safety envelope defined in `docs/skill-contract.md` → **Safety Envelope**. If a skill's contract cannot satisfy all envelope items, it must declare that incompatibility explicitly and the orchestrator must not invoke it.

### Where extension skills apply

Compatible extension skills may be considered at three workflow phases:

- **Step 1 (Completeness and scope)** — a skill whose `When-Applies` matches the request may contribute domain-specific completeness criteria.
- **Step 4 (Execute)** — a skill may be consulted as a domain specialist during implementation delegation.
- **Step 6 (Review)** — a skill may provide additional review criteria beyond the core Reviewer's checklist.

In all cases the core agent for that phase runs first and is authoritative. Extension skill findings are supplemental context.
