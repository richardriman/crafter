# Plugin System — Design Document

> Status: Draft
> Date: 2026-03-01

## Motivation

Crafter is a general-purpose AI development methodology. Some extensions don't belong in core because they are project-specific, team-specific, or depend on external infrastructure (e.g., MCP servers). A plugin system allows extending Crafter without forking it.

### Use Cases

1. **Project-specific review rules** — The core Reviewer agent is intentionally generic. Projects with specific stacks (Rails, React, Go, etc.) need stricter or more targeted review criteria. A plugin can bundle review rules for a given stack (e.g., N+1 query detection, framework-specific security patterns).

2. **External task management integration** — Connecting Crafter's task lifecycle (`.planning/tasks/`) with an external system via MCP server. This requires additional rules for task synchronization and possibly new commands. It is not a general open-source feature, so it does not belong in core.

## Plugin Structure

A plugin is a directory with a known layout — a subset of Crafter's own structure:

```
my-plugin/
├── plugin.md       # manifest (required)
├── rules/          # additional rule fragments
├── commands/       # additional slash commands
├── agents/         # additional agent definitions
├── templates/      # additional templates
└── hooks/          # additional hooks
```

All subdirectories are optional. Only `plugin.md` is required.

## Plugin Manifest — `plugin.md`

The manifest describes what the plugin provides and what it requires.

```markdown
# Plugin Name

> Short description of what this plugin does.

## Provides

- rules/task-lifecycle-ext.md — extends task lifecycle with MCP sync
- commands/sync-tasks.md — manual task synchronization command

## Requires

- MCP server: `task-manager` — must be configured in Claude Code settings
```

The manifest serves three purposes:
1. **Documentation** — human-readable description for developers.
2. **Discoverability** — `/crafter-status` can list active plugins with their descriptions.
3. **Prerequisite validation** — the orchestrator can warn if declared requirements are not met.

## Lookup Chain

Plugins are discovered by scanning up to three directory levels, from most specific to least:

| Priority | Location | Scope |
|---|---|---|
| 1 (highest) | `.claude/crafter/plugins/` | Current project |
| 2 | First `../.claude/crafter/plugins/` found walking up parent directories | Parent project (e.g., integration repo) |
| 3 (lowest) | `~/.claude/crafter/plugins/` | Global (user-wide) |

### Parent lookup

Starting from the current project root, walk up the directory tree. Stop at the **first** parent directory that contains `.claude/crafter/plugins/`. This covers the common pattern of an integration repo with nested sub-repos:

```
integration-repo/
├── .claude/crafter/plugins/
│   └── task-sync/              ← shared by all sub-repos
├── sub-repo-1/
│   └── .claude/crafter/plugins/
│       └── rails-review/       ← specific to sub-repo-1
└── sub-repo-2/                 ← inherits task-sync from parent
```

In this setup:
- `sub-repo-1` sees: `rails-review` (project) + `task-sync` (parent)
- `sub-repo-2` sees: `task-sync` (parent)

### Conflict resolution

If the same plugin name exists at multiple levels, the most specific one wins (project > parent > global). Plugins at the same level are independent — no ordering guarantees between them.

## Integration with Core Workflows

Plugins extend existing workflows through a single mechanism: **rule merging**.

### Rule extensions

A plugin provides additional rule files that core commands load alongside built-in rules. The naming convention determines how they are loaded:

- `rules/<core-rule>-ext.md` — loaded **after** the corresponding core rule. Example: `rules/task-lifecycle-ext.md` is loaded when the orchestrator loads `task-lifecycle.md`.
- `rules/<custom-name>.md` — standalone rules, loaded when explicitly referenced by a plugin command.

Core commands include a generic directive at the end of their rule loading:

> After loading core rules, scan active plugins for matching `*-ext.md` rule files and append them.

This keeps the extension mechanism simple — plugins add context and constraints, they do not replace core behavior.

### Commands and agents

Plugin commands and agents work the same as core ones — they are markdown files that Claude Code discovers by directory convention. No special registration is needed. Plugin commands should use a namespace prefix to avoid collisions (e.g., `/crafter-task-sync` rather than `/task-sync`).

### Hooks

Plugin hooks follow Claude Code's native hook mechanism. They are registered in the project's or user's `settings.json` — the plugin manifest documents which hooks need to be configured, but Crafter does not auto-register them.

## Installation

### Manual (v1)

Copy the plugin directory into the appropriate `plugins/` location:

```bash
# Project-local
cp -r my-plugin .claude/crafter/plugins/

# Global
cp -r my-plugin ~/.claude/crafter/plugins/
```

### Future considerations

- `crafter plugin add <git-url>` — clone a plugin repo into `plugins/`
- `crafter plugin list` — show active plugins from all levels with their source
- `crafter plugin remove <name>` — remove a plugin

These are out of scope for the initial implementation.

## Example: Task Sync Plugin

```
task-sync/
├── plugin.md
└── rules/
    └── task-lifecycle-ext.md
```

**plugin.md:**
```markdown
# Task Sync

> Synchronizes Crafter task lifecycle events with an external task management
> system via MCP server.

## Provides

- rules/task-lifecycle-ext.md — on task create/update/complete, sync with MCP

## Requires

- MCP server: `task-manager`
```

**rules/task-lifecycle-ext.md:**
```markdown
## Task Synchronization (Plugin: task-sync)

After creating, updating, or completing a task file in `.planning/tasks/`,
synchronize the change with the external task management system:

1. On task creation — create a corresponding external task via MCP `task-manager`
2. On status change — update the external task status
3. On task completion — mark the external task as done
4. Include a link to the external task in the task file header
```

## Example: Rails Review Plugin

```
rails-review/
├── plugin.md
└── rules/
    └── review-rules.md
```

**plugin.md:**
```markdown
# Rails Review

> Adds Rails-specific review criteria to the Reviewer agent.

## Provides

- rules/review-rules.md — Rails review checklist
```

**rules/review-rules.md:**
```markdown
## Rails Review Criteria (Plugin: rails-review)

In addition to general review criteria, check for:

- N+1 queries — flag ActiveRecord calls inside loops
- Mass assignment — ensure strong parameters are used
- Unscoped queries — flag `Model.all` or `Model.find` without scoping
- Missing database indexes — flag queries on columns without indexes
- Callback side effects — flag `after_save` / `after_commit` with external calls
```

## Design Principles

1. **Convention over configuration** — plugins work by placing files in the right directories, not by editing config files.
2. **Additive only** — plugins extend behavior, they do not replace or override core rules.
3. **No runtime dependencies** — plugins are markdown files, same as core Crafter. No package manager, no build step.
4. **Transparent** — active plugins and their effects are visible via `/crafter-status`.
