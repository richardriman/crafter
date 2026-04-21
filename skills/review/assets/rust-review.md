## Rust Clippy (Static Analysis)

### Rule
The `prepare-context.sh` script automatically runs `cargo clippy` when Rust files
are detected. Its output is included in the `== RUST-CLIPPY ==` section.

### What clippy checks
Clippy applies hundreds of lints organized by category:
- **correctness** — code that is outright wrong or very useless (on by default)
- **suspicious** — code that is most likely wrong or useless (on by default)
- **style** — code that should be written in a more idiomatic way
- **complexity** — code that does something simple but in a complex way
- **perf** — code that can be written to run faster
- **pedantic** — lints which are rather strict or have occasional false positives

The project's `Cargo.toml` or `clippy.toml` may configure which lints are active.

### How to use in review
The agent should:
1. Summarize clippy findings (not just dump raw output)
2. Correlate them with its own findings (avoid duplicates)
3. Flag any clippy errors that are especially critical (correctness, suspicious)
4. Note if clippy was not available — the review continues without it
