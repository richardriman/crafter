# Workspace Mode — Phase 0: Per-Task Workspace Instances

> Status: Findings (spike)
> Date: 2026-11-14
> Settles: [workspace-mode.md](workspace-mode.md) § Open Questions — "Where do per-task worktrees live, and how does the devcontainer see them?"

**Result:** the candidate works. A task runs in a **workspace instance** — a sibling directory of the main workspace that holds a worktree of the workspace root and a worktree of every nested repo, with its own container stack. Git works inside the container once the main checkouts' git dirs are mounted at their host paths. An instance costs seconds and a few hundred MB of files; the expensive parts (toolchains, gems, the compiled backend target dir) are shared caches.

---

## The workspace under test

A real workspace, described generically:

- **Workspace root** repo: `.devcontainer/` (Dockerfile, `compose.yml`, `devcontainer.json`), `Makefile`, `Procfile`, docs. `compose.yml` bind-mounts the root folder (`..`) at `/workspace`.
- **Nested repos** cloned at fixed paths inside the root, untracked by the root repo (not submodules): a Ruby backend repo, a Rust backend repo, an Elm frontend repo, and four smaller repos. Run tooling uses fixed container paths (`/workspace/<repo>`).
- Named volumes nested inside the bind (`node_modules`, Elm build cache, the Rust `target` dir) and in `$HOME` (`mise` toolchains + gems, cargo registry).
- MariaDB data is a host bind (`${MYSQL_DATA_DIR:-~/volumes/…}`), shared by default by every checkout.
- `ws-run` (host → container exec) finds the root by walking up to `.devcontainer/devcontainer.json` and calls `devcontainer exec --workspace-folder <root>`.
- The compose project name comes from `COMPOSE_PROJECT_NAME` in the root `.env` (`make set-namespace` writes `${USER}-<dirname>`).
- Host git 2.50 (macOS); container git 2.39 (Debian bookworm). Docker via OrbStack.

---

## What was tried

| # | Experiment | Outcome |
|---|------------|---------|
| 1 | Instance = sibling dir `<parent>/<ws>--<slug>/`: `git worktree add` of the root on the task branch, worktrees of two touched repos on the same branch, the other five nested repos as detached worktrees at their `origin/HEAD` base | Works. Main checkouts untouched (their branches and working trees unchanged). |
| 2 | Container sees only the instance (`..:/workspace`), nothing else | `git status` fails: `fatal: not a git repository: /<host path>/<repo>/.git/worktrees/<repo>` — the worktree's `.git` file holds an absolute host path. |
| 3 | Also bind-mount each main checkout's `.git` dir at **the same absolute host path** | Git works: status, log, commit. A commit made in the container is immediately visible from the host and in the main repo's branch list. |
| 4 | Relative worktrees (`git worktree add --relative-paths`, git ≥ 2.48) | Rejected. Git writes `extensions.relativeWorktrees = true` into the **shared** repo config; container git 2.39 then refuses the whole repo — `fatal: unknown repository extension found: relativeworktrees` — including the main checkout inside the main container. Relative paths would also need the container to mirror the host's relative layout (instance and main side by side), which `/workspace` does not. |
| 5 | Alternative: local clones (`git clone <main>/<repo>`) instead of worktrees | Works without any extra mount (the clone's `.git` is a real dir inside the instance). Cheap on disk: pack files are hardlinked (≈ 58 MB real for the Ruby backend repo, whose `.git` is 229 MB). But `origin` points at the local main repo and its `origin/*` refs are the main repo's *local* branches; the task branch is invisible from the main checkout until pushed; hardlinked packs diverge after the next repack. Workable, more plumbing than it saves. |
| 6 | `devcontainer up` of the instance with its own compose project, shared toolchain/cargo/target volumes, own MariaDB volume | See measurements. |

---

## Measured results

All numbers from one instance with two touched repos (both backend repos) and five untouched nested repos, on a laptop with Docker via OrbStack. "Warm" = the dev image layers were already in the build cache.

| Step | Time | Disk (instance-owned) |
|------|------|------|
| Create instance: 8 worktrees (root + 7 nested repos) | 3.4 s | 327 MB of files |
| `devcontainer up` (5 services), warm | 23 s | — |
| `devcontainer up`, cold build cache | ~6 min, one-time per Dockerfile change | +4.3 GB image if its tag is separate |
| Gems: `bundle check` against the shared toolchain volume | 0 s | 0 (shared) |
| `yarn install` into fresh per-instance `node_modules` (two repos) | 24 s + 39 s | 295 MB + 444 MB |
| Own MariaDB: ready on first query, `make db-setup` (schema load) | 10 s | 184 MB |
| Rust backend: incremental build on the **shared** `target` dir (task branch differs from the main checkout's branch), GraphQL answering | 85 s | +0.46 GB in the shared target |
| Ruby backend listening | < 5 s | — |
| **Whole stack running, from nothing** | **≈ 3.5 min warm** | **≈ 1.3–1.4 GB** |
| Teardown: `compose down -v` + volume sweep | 5 s | — |
| Teardown: remove worktrees | 1 s | — |

Checked through the instance's own proxy, inside its container: the frontend dev server returned 200, a GraphQL query through the proxy returned data, the Ruby backend rendered its own pages. Git status/log/commit worked in the container for the root and nested worktrees.

What sharing saves: the caches the instance did **not** duplicate are the Rust `target` dir (29 GB), the toolchain + gem volume (1.4 GB) and the cargo registry (1.2 GB) — a fresh copy of these per instance would mean ~32 GB and a from-scratch Rust build plus a full gem install per task.

The main stack kept running untouched throughout; the main checkouts never changed branch.

---

## Pitfalls found

- **`git worktree prune` inside any container that mounts the main git dirs deletes other worktrees' metadata.** Inside the container, every worktree whose path is a host path (other instances, ad-hoc worktrees) looks missing. A dry run in the spike listed eight unrelated worktrees as removable. `git gc` runs the same prune for entries older than `gc.worktreePruneExpire` (3 months). **Fix: create instance worktrees with `--lock`.** Locked worktrees are never pruned. (The main container already has this exposure for ad-hoc worktrees today.)
- **Repo config is shared between the main checkout and every worktree.** A setup step that runs `git config` (e.g. a hooks installer setting `core.hooksPath`) in an instance changes the main checkout too. Instance setup must not write repo config, or must use `git config --worktree`. Paths baked into config (`core.hooksPath=/workspace/...`) resolve to the instance's files inside an instance container — usually what you want, but the main repo's hooks dir (`/workspace/.git/hooks`) does not exist in an instance, where `.git` is a file, so such hooks silently stop running.
- **`git worktree add -b <branch> <path> origin/<base>` sets the base as upstream** (`branch.autoSetupMerge`). Use `--no-track`; the first push then sets the real upstream with `-u`.
- **The compose project name is read from the `.env` in the CLI's current directory** (`devcontainer up` checks the `COMPOSE_PROJECT_NAME` env var, then `<cwd>/.env`, then `name:` in `compose.yml`). Run `devcontainer up` from the instance root, or export the variable. Otherwise every instance falls back to the compose file's `name:` and they collide.
- **The dev image tag is global** (`${USER}-<ws>-dev:<variant>`), not per project. An instance whose branch changes the Dockerfile retags the image the main checkout uses next time. Harmless while Dockerfiles are equal; an instance that changes `.devcontainer/` should use its own tag.
- **Default MariaDB data dir is a shared host bind.** Two stacks on one data dir is two servers on one InnoDB directory. Each instance needs its own data dir or named volume.
- **Build args:** passing the host's GID (macOS `staff` = 20) as `USER_GID` fails the image build (`groupadd: GID '20' already exists`). Keep the defaults and let the devcontainer CLI remap the UID, as the main checkout does.
- `devcontainer.json` `mounts` of type `volume` are prefixed with the compose project name (`<project>_<name>`) — they are per checkout, not global as their comments suggest.

---

## Recommended rule

### Layout

```
<parent>/
  <ws>/                        main workspace — never switched, never edited by a task
  <ws>--<task-slug>/           workspace instance = worktree of the workspace root, task branch
    <touched-repo>/            worktree, task branch (same name in every touched repo)
    <untouched-repo>/          worktree, detached at the repo's base (origin/HEAD)
```

- **A sibling directory, not a subfolder of the workspace.** Nothing inside the workspace folder changes, and the existing tooling works unmodified: `ws-run` walks up to the instance's `.devcontainer/`, the devcontainer CLI keys the container on the instance path, `make set-namespace` derives a unique compose project from the directory name.
- **Every nested repo gets a worktree**, touched or not, so the instance is the whole system at known commits. Untouched repos are detached at their base, not at whatever the main checkout has checked out.
- **One instance per task.** The slug includes the tracker ID when there is one, so names stay unique.

### Create

1. `git -C <ws> worktree add --lock --reason "<task>" --no-track -b <branch> <instance> origin/<root-base>`
2. For each nested repo: `git -C <ws>/<repo> worktree add --lock --reason "<task>" …` — `--no-track -b <branch> … origin/<base>` if touched, `--detach … origin/<base>` otherwise.
3. Copy the untracked local config the stack needs from the main checkout (`.devcontainer/.env` with its secrets, per-repo `.env` and local config files). Never commit them.
4. Write the instance's env:
   - root `.env`: `COMPOSE_PROJECT_NAME=${USER}-<ws>--<task-slug>`
   - `.devcontainer/.env`: `WS_MAIN=<absolute path of the main workspace>`, `MYSQL_DATA_DIR=<per-instance data dir>`
5. From the instance root: `devcontainer up --workspace-folder .`, then the workspace's normal setup (deps install, DB setup). Skip any step that writes git config (hooks installers): repo config is shared with the main checkout and already set there.

Adding a repo to a running task later is step 2 for that repo, switching its worktree from detached to the task branch (`git -C <instance>/<repo> switch -c <branch>`).

### Git inside the container

- **Absolute worktree paths (git's default), plus each main checkout's `.git` dir bind-mounted read-write at its identical host path.** Worktree links resolve unchanged on both sides; a commit in the container is a commit in the main repo.
- **Never relative worktrees** while any container git is older than 2.48: the repo extension they set breaks every checkout of that repo for the older git.
- **Lock every instance worktree** (`--lock`) so a `git worktree prune` or `git gc` in any container cannot remove them; do not run `git worktree prune` inside containers.

### Shared vs. per instance

| Shared across the main checkout and all instances | Per instance |
|---|---|
| Dev image (by tag) | Working files (worktrees) |
| Toolchains + gems volume | `node_modules`, frontend build caches (lockfiles differ per branch; install is under a minute) |
| Cargo registry | Database (own data dir/volume; schema load in seconds) |
| Rust `target` dir (cargo locks it; a different branch costs an incremental rebuild) | Redis, mail catcher, proxy — the whole compose stack |
| Git object stores (via worktrees) | Compose project, network, container |

- **Never share a database data dir between stacks.** An instance starts from the schema (`db-setup`); to start from the developer's data, dump and restore from the main stack (not measured here).
- No port conflicts: services share the dev container's network namespace and publish nothing to the host. Reaching an instance from a host browser needs a port forward (the IDE does it; the CLI does not).

### Tear down

1. `docker compose -p <project> -f <instance>/.devcontainer/compose.yml down -v --remove-orphans`
2. Remove the remaining `<project>_*` volumes — volumes declared in `devcontainer.json` `mounts` are not in `compose.yml`, so `down -v` leaves them (observed: four were left).
3. Remove the instance's DB data dir if it is a bind.
4. For each repo: `git worktree unlock` + `git worktree remove --force`; root last.
5. Keep the task branches (they back the PRs); delete them only when the task is abandoned.

### Changes a workspace needs

To its devcontainer (one-time, backward compatible — with `WS_MAIN` unset the main checkout behaves as today):

```yaml
# compose.yml, dev service volumes: the main checkout's git dirs at their host paths
      - ..:/workspace
      - ${WS_MAIN:-..}/.git:${WS_MAIN:-/nonexistent-ws-main}/.git
      - ${WS_MAIN:-..}/<repo>/.git:${WS_MAIN:-/nonexistent-ws-main}/<repo>/.git   # one line per nested repo

# compose.yml, shared caches named after the main checkout's project, so the main
# checkout keeps its existing volumes and every instance resolves the same ones
volumes:
  mise:           { name: "${USER:-dev}-<ws>_mise" }
  cargo_registry: { name: "${USER:-dev}-<ws>_cargo_registry" }
  rust_target:    { name: "${USER:-dev}-<ws>_rust_target" }
```

- Move the Rust `target` volume from `devcontainer.json` `mounts` into `compose.yml` (`- rust_target:/workspace/rust/target` on the dev service): `devcontainer.json` volume mounts are always project-prefixed, so they cannot be shared.
- The MariaDB data dir is already overridable (`MYSQL_DATA_DIR`); the instance sets it. No change.
- Git config holding container paths (e.g. `core.hooksPath=/workspace/.git/hooks` in the root repo) points at nothing in an instance, where `.git` is a file; hooks set that way silently stop running there. Prefer a path relative to the worktree (as the Rust repo's `.githooks` already is).
- A periodic cache sweeper that runs once per container would run once per instance against the shared `target` dir. Guard it (a lock file in the volume) or run it only in the main checkout.

To its tooling: two Makefile targets (or scripts), `instance-create <slug> <branch> <repos…>` and `instance-remove <slug>`, doing the Create and Tear down steps above. An existing agent-runner integration in the workspace already has the same lifecycle (create, patch compose, start, stop, remove) built on fresh shallow clones; it can switch to these targets.

To Crafter (Phase 2): `/crafter-do` in a workspace creates the instance after plan approval, runs agents with the instance root as the working directory (and each touched repo's instance path as `repository`), runs checks through the instance's container, and offers teardown after the PRs are open. The instance path and compose project go into the task file so a resumed run finds them.

---

## Alternatives compared

| Option | Verdict |
|--------|---------|
| **Instance of worktrees + main git dirs mounted at host paths** | **Recommended.** Seconds to create, one object store, branches visible from the main checkout; costs N bind-mount lines and `--lock`. |
| Instance of local clones (hardlinked objects) | Works with zero container changes; similar disk. Rejected for plumbing: `origin` must be re-pointed and re-fetched, branches are invisible from the main checkout until pushed, packs diverge after repacks. Reasonable fallback where the compose file cannot change. |
| Instance of fresh shallow clones from the remote (the agent-runner approach) | Needs network and full dependency setup per task, no shared objects; fine for unattended agents, slow for interactive work. Not measured. |
| Relative worktrees | Rejected — breaks older container git for the whole repo (see experiment 4). |
| A second full checkout per developer, reused across tasks | Already works, but it is a checkout that switches branches — the problem this design avoids. |
