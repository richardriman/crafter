# Architecture

## Structure

```
<!-- Paste an ASCII directory tree here and annotate each entry -->
project/
├── src/
│   ├── <!-- module/ -->    # <!-- what lives here -->
│   └── <!-- module/ -->    # <!-- what lives here -->
├── tests/                  # <!-- test layout -->
└── <!-- other dirs -->
```

## Navigation — Where to Find What

| What | Where |
|---|---|
| <!-- HTTP route handlers --> | <!-- src/routes/ --> |
| <!-- Business logic --> | <!-- src/services/ --> |
| <!-- Database models --> | <!-- src/models/ --> |
| <!-- Configuration --> | <!-- src/config/ or config/ --> |
| <!-- Tests --> | <!-- tests/ or __tests__/ --> |

## Key Patterns & Decisions

<!-- Document significant technical choices with rationale. -->

### <!-- Pattern or Decision Name -->
<!-- What it is, why it was chosen, and any important caveats. -->

### <!-- Pattern or Decision Name -->
<!-- What it is, why it was chosen, and any important caveats. -->

## Conventions

- <!-- What belongs in which layer (e.g. no business logic in route handlers) -->
- <!-- Naming conventions specific to this codebase -->
- <!-- What NOT to do (common pitfalls, anti-patterns to avoid) -->
- <!-- Error handling approach -->
