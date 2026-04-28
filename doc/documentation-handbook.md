# Documentation Handbook

This handbook defines how documentation is authored and maintained for this repository.

## 1) Purpose

Documentation in this project must:
- help contributors understand intent quickly
- make changes traceable from issue → spec → implementation
- stay accurate as the codebase evolves

## 2) Documentation Structure

Use these locations:

- `doc/00-index.md` — entry point and document map
- `doc/overview/` — high-level project context
  - `north-star.md`
  - `architecture-overview.md`
- `doc/spec/features/` — feature-level specifications
- `doc/spec/nonfunctional.md` — cross-cutting non-functional requirements
- `doc/decisions/` — decision records and index
- `doc/guides/` — workflow and operational guides
- `doc/templates/` — reusable templates

## 3) Authoring Standards

All docs should:
- be concise and scannable
- use explicit headings and bullet points
- separate facts, assumptions, and open questions
- avoid hidden tribal knowledge
- prefer examples over abstract wording

When a section is unknown, use `TODO:` with owner/context instead of guessing.

## 4) Naming Conventions

- use lowercase kebab-case filenames
- keep names descriptive and stable
- feature spec filenames should match feature intent
  - example: `skillbook-learning-system.md`

## 5) Required Content by Document Type

### Overview docs
Must explain:
- why the project exists
- who it serves
- system boundaries and key components

### Feature specs
Must include:
- problem statement
- scope and non-goals
- acceptance criteria (testable)
- risks/assumptions
- verification approach

### Decision docs
Must include:
- context
- decision
- consequences
- date and status

## 6) Change Management Rules

For any non-trivial behavior change:
1. update or add the relevant spec/overview/decision doc
2. keep docs aligned with merged behavior
3. link work item references (`GH-<id>` or `#<id>`) when applicable

## 7) Quality Checklist (Docs Review)

Before considering docs complete:
- [ ] structure follows this handbook
- [ ] claims are verifiable from repo context
- [ ] acceptance criteria are specific and testable
- [ ] no stale or contradictory statements
- [ ] open questions are marked with `TODO:`

## 8) Ownership

Repository owner is the final approver for documentation quality and merge readiness.
