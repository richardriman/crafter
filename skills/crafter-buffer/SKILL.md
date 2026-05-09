---
name: "crafter-buffer"
description: "Append a UAT or Gap entry to the current run's buffer (.crafter/run/<task-id>/)"
---

## What this skill does

This skill appends a single NDJSON entry to a per-run buffer file under `.crafter/run/<task-id>/`. Sub-agents (implementer, verifier, reviewer) use it to record findings that should not block the current run — UAT items (manual QA that humans must validate) or Gap items (deferred tech-debt / out-of-scope follow-ups). It is most relevant under `--auto` mode (see GH#15 / `skills/crafter-do/SKILL.md`) where blocking on unresolvable findings would defeat unattended orchestration, but it is equally useful in standard mode whenever a finding should be persistent rather than embedded in chat. The run-directory lifecycle (creation, persistence, cleanup) is defined in Phase 3 of task `20260509-feat-gh-16-buffer-skill`.

## Operations

Both operations require the caller to supply `--run-dir`, which is the `.crafter/run/<task-id>/` path the orchestrator already knows. The filename within that directory is fixed by kind — the filename is not a flag.

**Note on naming:** CLI flags use kebab-case (`--why-manual`, `--run-dir`) following shell ergonomics convention. JSON keys in the buffer file use snake_case (`why_manual`, `created_at`) following JSON portability convention. The Go subcommand translates between the two; callers and downstream tools must not conflate them.

### `crafter buffer uat` — append a UAT (manual-QA) entry

```
crafter buffer uat \
  --run-dir <path>      \
  --title "..."         \
  --source "..."        \
  --created-by "..."    \
  --task-id "..."       \
  --verify "..."        \
  --why-manual "..."
```

Required flags: `--run-dir`, `--title`, `--source`, `--created-by`, `--task-id`, `--verify`, `--why-manual`.

Appends one NDJSON line to `<run-dir>/uat-buffer.jsonl`. Creates the file (with its marker line) if it does not exist.

### `crafter buffer gap` — append a Gap (tech-debt / deferred follow-up) entry

```
crafter buffer gap \
  --run-dir <path>      \
  --title "..."         \
  --source "..."        \
  --created-by "..."    \
  --task-id "..."       \
  --detail "..."        \
  --followup "..."
```

Required flags: `--run-dir`, `--title`, `--source`, `--created-by`, `--task-id`, `--detail`, `--followup`.

Appends one NDJSON line to `<run-dir>/gaps-buffer.jsonl`. Creates the file (with its marker line) if it does not exist.

## Entry shape

The subcommand generates `id`, `kind`, and `created_at` at append time. The caller provides `--created-by` (the calling agent or skill name, e.g. `crafter-implementer`, `crafter-reviewer`, `crafter-do`) and `--task-id` (the task-file basename without extension, e.g. `20260509-feat-gh-16-buffer-skill`), as well as `--title`, `--source`, and the kind-specific value flags.

### UAT entry (annotated)

```json
{
  "id":         "<subcommand-generated, 12-char hex>",
  "kind":       "uat",
  "created_at": "2026-05-09T14:30:00Z",
  "created_by": "<calling agent name, e.g. crafter-implementer>",
  "task_id":    "<task-file basename without extension>",
  "title":      "<short headline>",
  "source":     "<file:line | phase:step | review-finding-id>",
  "verify":     "<what to manually verify>",
  "why_manual": "<why automation is not possible>"
}
```

### Gap entry (annotated)

```json
{
  "id":         "<subcommand-generated, 12-char hex>",
  "kind":       "gap",
  "created_at": "2026-05-09T14:30:00Z",
  "created_by": "<calling agent name>",
  "task_id":    "<task-file basename without extension>",
  "title":      "<short headline>",
  "source":     "<file:line | phase:step | review-finding-id>",
  "detail":     "<description of the gap or tech debt>",
  "followup":   "<recommended action to close the gap>"
}
```

In the buffer files, each entry is a single line (no pretty-printing). The annotated forms above are for documentation only.

### Escaping reference

#### (i) Minimal valid entry

CLI invocation:

```sh
crafter buffer uat \
  --run-dir .crafter/run/20260509-feat-gh-16-buffer-skill \
  --title "Confirm OAuth callback redirect" \
  --source "auth/callback.go:42" \
  --created-by "crafter-implementer" \
  --task-id "20260509-feat-gh-16-buffer-skill" \
  --verify "Click Sign In; confirm browser lands on /dashboard not /login." \
  --why-manual "Requires a live OAuth provider and browser session."
```

Resulting NDJSON line:

```
{"id":"a1b2c3d4e5f6","kind":"uat","created_at":"2026-05-09T14:30:00Z","created_by":"crafter-implementer","task_id":"20260509-feat-gh-16-buffer-skill","title":"Confirm OAuth callback redirect","source":"auth/callback.go:42","verify":"Click Sign In; confirm browser lands on /dashboard not /login.","why_manual":"Requires a live OAuth provider and browser session."}
```

#### (ii) Value containing a fenced code block

Backticks pass through unescaped in JSON strings. Newlines become `\n`. Double-quotes inside the value become `\"`.

CLI invocation (heredoc recommended for multi-line values):

```sh
crafter buffer uat \
  --run-dir .crafter/run/20260509-feat-gh-16-buffer-skill \
  --title "Verify hover preview renders with empty dataset" \
  --source "ui/preview.tsx:118" \
  --created-by "crafter-implementer" \
  --task-id "20260509-feat-gh-16-buffer-skill" \
  --verify "$(cat <<'EOF'
Run the app with an empty dataset fixture:

\`\`\`bash
DATA=fixtures/empty.json npm run dev
\`\`\`

Hover over any row in the table and confirm the preview panel shows
the placeholder message rather than a blank white box.
EOF
)" \
  --why-manual "Requires a browser and a running dev server; no headless test covers empty-dataset hover state."
```

Resulting NDJSON line:

```
{"id":"b2c3d4e5f6a7","kind":"uat","created_at":"2026-05-09T14:31:00Z","created_by":"crafter-implementer","task_id":"20260509-feat-gh-16-buffer-skill","title":"Verify hover preview renders with empty dataset","source":"ui/preview.tsx:118","verify":"Run the app with an empty dataset fixture:\n\n```bash\nDATA=fixtures/empty.json npm run dev\n```\n\nHover over any row in the table and confirm the preview panel shows\nthe placeholder message rather than a blank white box.","why_manual":"Requires a browser and a running dev server; no headless test covers empty-dataset hover state."}
```

#### (iii) Multi-line value with embedded blank lines

CLI invocation:

```sh
crafter buffer gap \
  --run-dir .crafter/run/20260509-feat-gh-16-buffer-skill \
  --title "Rate-limit middleware not applied to internal endpoints" \
  --source "middleware/ratelimit.go:67" \
  --created-by "crafter-reviewer" \
  --task-id "20260509-feat-gh-16-buffer-skill" \
  --detail "$(cat <<'EOF'
The rate-limit middleware wraps public API routes but is skipped for
/internal/* endpoints.

This was intentional during prototyping (internal callers are trusted)
but should be revisited before production: a compromised internal
service could abuse the endpoints without throttling.
EOF
)" \
  --followup "Apply rate-limit middleware to /internal/* or add an explicit allow-list with documented justification."
```

Resulting NDJSON line:

```
{"id":"c3d4e5f6a7b8","kind":"gap","created_at":"2026-05-09T14:32:00Z","created_by":"crafter-reviewer","task_id":"20260509-feat-gh-16-buffer-skill","title":"Rate-limit middleware not applied to internal endpoints","source":"middleware/ratelimit.go:67","detail":"The rate-limit middleware wraps public API routes but is skipped for\n/internal/* endpoints.\n\nThis was intentional during prototyping (internal callers are trusted)\nbut should be revisited before production: a compromised internal\nservice could abuse the endpoints without throttling.","followup":"Apply rate-limit middleware to /internal/* or add an explicit allow-list with documented justification."}
```

## Examples

Realistic entries as they appear in the buffer files (single lines; wrapped here for readability inside the code block).

### UAT examples

**UAT example 1 — fenced code block in `verify`:**

```
{"id":"d4e5f6a7b8c9","kind":"uat","created_at":"2026-05-09T15:00:00Z","created_by":"crafter-verifier","task_id":"20260509-feat-gh-16-buffer-skill","title":"Confirm export CSV survives special characters in column headers","source":"export/csv_writer.go:89","verify":"Generate a report where a column header contains a comma, a double-quote, and a newline. Download the CSV and open it in Excel and LibreOffice Calc:\n\n```bash\ncurl -s 'http://localhost:8080/export?fmt=csv&report=special-chars' -o /tmp/test.csv\nopen /tmp/test.csv\n```\n\nConfirm both apps parse the file without corruption and the header row shows exactly the expected strings.","why_manual":"Spreadsheet rendering is application-level behaviour that cannot be asserted by the Go test suite."}
```

**UAT example 2 — multi-line `verify` with blank line:**

```
{"id":"e5f6a7b8c901","kind":"uat","created_at":"2026-05-09T15:05:00Z","created_by":"crafter-implementer","task_id":"20260509-feat-gh-16-buffer-skill","title":"Confirm session cookie is marked HttpOnly and Secure in production","source":"auth/session.go:31","verify":"Deploy to the staging environment and log in.\n\nOpen browser DevTools → Application → Cookies. Locate the session cookie and confirm:\n- HttpOnly flag is set\n- Secure flag is set\n- SameSite is Strict or Lax (not None)\n\nRepeat with an HTTP (non-TLS) origin and confirm the cookie is NOT sent.","why_manual":"Cookie attribute inspection requires a live browser and a production-equivalent TLS terminator; the test suite runs over plain HTTP with a test key."}
```

### Gap examples

**Gap example 1 — multi-line `detail`:**

```
{"id":"f6a7b8c901aa","kind":"gap","created_at":"2026-05-09T15:10:00Z","created_by":"crafter-reviewer","task_id":"20260509-feat-gh-16-buffer-skill","title":"Database migrations lack a rollback path","source":"db/migrations/0042_add_audit_log.sql","detail":"Migration 0042 adds an audit_log table and a trigger. The up migration is clean, but there is no corresponding down migration.\n\nIn a rollback scenario the trigger would be left dangling, which breaks the schema version check on startup and prevents the previous binary from running.","followup":"Add a down migration for 0042 that drops the trigger before dropping the table. Consider making the CI pipeline enforce that every up migration has a matching down migration."}
```

**Gap example 2 — fenced code block in `detail`:**

```
{"id":"a7b8c9d0e1f2","kind":"gap","created_at":"2026-05-09T15:15:00Z","created_by":"crafter-implementer","task_id":"20260509-feat-gh-16-buffer-skill","title":"Pagination cursor is not opaque — leaks internal row ID","source":"api/handlers/list.go:204","detail":"The list endpoint returns a `next_cursor` that is just a base64-encoded row ID:\n\n```bash\n$ echo 'eyJpZCI6IDQyfQ==' | base64 -d\n{\"id\": 42}\n```\n\nClients can enumerate records by decoding and incrementing the cursor. This is a minor information-disclosure issue and a brittle API contract (cursor format tied to DB schema).","followup":"Replace the cursor with an HMAC-signed token that encodes the row ID and an expiry; reject cursors that fail verification. Document the new cursor format in the API changelog."}
```

## Creation behavior

The subcommand creates the buffer file if it does not exist. The first line written to a new file is a fixed marker object so that future tooling (PR composer, debug tools) can detect the file kind without inspecting the filename:

- `uat-buffer.jsonl` first line: `{"_marker":"uat-buffer","_format":"ndjson-v1"}`
- `gaps-buffer.jsonl` first line: `{"_marker":"gaps-buffer","_format":"ndjson-v1"}`

Subsequent lines are full schema-1 NDJSON entries as defined above. File names are fixed by kind (`uat-buffer.jsonl`, `gaps-buffer.jsonl`); the filename is not a flag. The `--run-dir` directory must exist before `crafter buffer` is called; the subcommand does not create it.

## Concurrency

Writes are sequential under the orchestrator (per Phase 1 Decision 3). The subcommand uses `O_APPEND | O_WRONLY | O_CREAT` and issues a single `write(2)` per call (the very first call on a fresh file writes the marker line and the first entry in one syscall; subsequent calls write only the entry). Steady-state entry-only writes are kept under the POSIX PIPE_BUF bound (conservatively 512 bytes on macOS, 4096 on Linux); the one-time coalesced first write (marker + first entry, up to ~560 bytes) relies on the stronger regular-file page-cache atomicity that Linux ext4/XFS and macOS APFS provide for sub-page writes. Cross-filesystem / NFS scenarios are explicitly NOT supported in this PoC. Concurrent calls from multiple sub-agents in the same run are out of scope.
