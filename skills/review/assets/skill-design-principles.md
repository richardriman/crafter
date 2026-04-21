# Skill Design Principles

These principles guide how this skill is built and maintained.
Reference this file when developing new features for the review skill.

## Script-driven logic

All deterministic decisions (project detection, file collection, language
mapping, UI prompts) live in `prepare-context.sh`, not in SKILL.md prose.
The script outputs structured sections (`== SECTION ==`) that tell the
orchestrator exactly what to do. This keeps behavior reproducible and
token-cheap.

## Minimal SKILL.md

SKILL.md only describes *what* to run and *how* to interpret the output.
No examples, no inline JSON, no duplicated logic. When adding new behavior,
add it to the script and document the new output section in one line.

## Orchestrator is a dispatcher

Claude reads the script output, matches the section header, and performs
the corresponding action. It should not invent new flows or add extra
questions beyond what the script specifies.
