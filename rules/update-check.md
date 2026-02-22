# Crafter Update Check

Run this check silently at command startup. It must never block the main workflow or produce errors visible to the user.

## Procedure

1. Resolve the installation path:

   Check `~/.claude/crafter/VERSION` first (global install). If that file does not exist, check `.claude/crafter/VERSION` (local install). If neither exists, skip the entire update check.

   Use the directory containing the found `VERSION` file as `$CRAFTER_BASE` for the remaining steps.

2. Read the installed version:

   ```bash
   cat "$CRAFTER_BASE/VERSION"
   ```

3. Check the cache:

   ```bash
   cat ~/.claude/crafter/.update-cache 2>/dev/null
   ```

   The cache file contains two lines: the timestamp of the last check (Unix seconds) and the latest version string found. Example:

   ```
   1708300000
   0.2.1
   ```

   Get the current time:

   ```bash
   date +%s
   ```

   If the cache exists and `(current_time - cached_timestamp) < 86400`, skip the API call and use the cached version. Go to step 5.

4. Fetch the latest release from GitHub:

   ```bash
   curl -sf --max-time 5 https://api.github.com/repos/richardriman/crafter/releases/latest
   ```

   - If the command fails (non-zero exit, timeout, or empty output), skip the rest of the update check entirely — do not show any error.
   - On success, extract the `tag_name` field value (e.g. `0.2.1`). Strip a leading `v` if present.
   - Write the cache file:

     ```bash
     printf '%s\n%s\n' "$(date +%s)" "<latest_version>" > ~/.claude/crafter/.update-cache
     ```

5. Compare versions:

   If the latest version is newer than the installed version, display a short, non-blocking notice **before** the command's first output:

   ```
   Note: Crafter update available (installed: X.Y.Z, latest: A.B.C). Run: curl -fsSL https://raw.githubusercontent.com/richardriman/crafter/main/install.sh | bash
   ```

   If versions are equal, display nothing.
