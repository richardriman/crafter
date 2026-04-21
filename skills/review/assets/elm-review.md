## Elm Review (Static Analysis)

### Rule
The `prepare-context.sh` script automatically runs `elm-review` when Elm files
are detected. Its output is included in the `== ELM-REVIEW ==` section.

### What elm-review checks
The project's `review/src/ReviewConfig.elm` defines the active rules,
which typically include:
- No debug code (NoDebug.Log, NoDebug.TodoOrToString)
- No exposing/importing everything
- No unused dependencies and variables
- No premature let computations
- No recursive update patterns
- Pipeline style enforcement
- Code simplification opportunities

### How to use in review
The agent should:
1. Summarize elm-review findings (not just dump raw output)
2. Correlate them with its own findings (avoid duplicates)
3. Flag any elm-review errors that are especially critical
