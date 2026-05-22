---
slug: '/docs/commands'
title: 'Development Commands'
hide_title: true
description: 'This page describes the cross-platform development command set used by LinaPro, including linactl, the Makefile compatibility entry, the Windows make.cmd wrapper, local environment checks and setup, development server management, full builds, WASM plugin builds, Docker image builds, plugin workspace management, agent resource symlink management, test validation, i18n checks, database initialization, and release governance. It helps developers use the same project toolchain consistently on macOS, Linux, and Windows.'
keywords:
  - make commands
  - development commands
  - build commands
  - LinaPro development
  - make env.check
  - make env.setup
  - make dev
  - make build
  - make pack.assets
  - make test
  - make image
  - make image.build
  - make init
  - make wasm
  - make plugins.install
  - make plugins.status
  - make agents
  - make agents.skills.link
  - make agents.prompts.link
  - make agents.md.link
  - make release.tag.check
  - linactl
  - development workflow
  - backend build
  - frontend build
  - Docker image
  - E2E testing
  - database initialization
  - plugin workspace
  - release governance
  - Windows support
  - make.cmd
  - cross-platform
---

`LinaPro` provides a cross-platform development command set. Long-term task orchestration lives in `hack/tools/linactl`, implemented as a `Go` program; the root `Makefile` and `make.cmd` are compatibility entries that forward to the underlying `linactl`. This means the same commands work on `macOS`, `Linux`, and `Windows`, without relying on `GNU Make` or `POSIX Shell` as the only entry point.

## Platform Notes

**Cross-platform native commands**: Every platform can use `linactl` directly:

```bash
cd hack/tools/linactl
go run . help
go run . status
go run . init confirm=init
go run . dev
```

**macOS / Linux**: The root `make` compatibility entry remains available:

```bash
make help
make init confirm=init
make dev
```

**Windows cmd.exe**: Use the `make.cmd` wrapper at the project root. In `cmd.exe`, executable extensions are resolved from the current directory, so the `.cmd` suffix can be omitted:

```cmd
make dev
make build
make help
```

**Windows PowerShell**: A current-directory prefix is required. In a default `Windows` environment, `.\make` works; to avoid ambiguity with another installed `make` command, use `.\make.cmd` explicitly:

```powershell
.\make help
.\make init confirm=init
.\make dev
```

Every `make <command>` example below can be replaced with `cd hack/tools/linactl && go run . <command>` with the same argument format. The `make` compatibility entry forwards common variables; for `linactl`-specific arguments, invoke `go run . <command>` directly.

## Command Overview

| Command | Category | Description |
|---------|----------|-------------|
| `make env.check` | Environment | Check local development tools, project-local frontend tools, and the `PostgreSQL` version |
| `make env.setup` | Environment | Install frontend dependencies, the `Playwright Chromium` browser, and required system dependencies |
| `make dev` | Development server | Restart the backend and frontend development servers |
| `make stop` | Development server | Stop the backend and frontend development servers |
| `make status` | Development server | Show backend/frontend running status and log paths |
| `make build` | Build | Fully build the frontend, plugins, and backend binary |
| `make pack.assets` | Build | Prepare frontend static assets and `manifest` resources for host embedding |
| `make wasm` | Build | Build all or selected runtime `WASM` plugins |
| `make tidy` | Build | Tidy `Go` module dependencies for the host, tools, and plugins |
| `make image` | Image | Build a production `Docker` image |
| `make image.build` | Image | Prepare image artifacts only, without running the `Docker` build |
| `make test` | Test | Run the full `E2E` test suite |
| `make test.go` | Test | Run `Go` unit tests |
| `make test.host` | Test | Run only the host-owned `E2E` tests |
| `make test.plugins` | Test | Run the official plugins' own `E2E` tests |
| `make test.scripts` | Test | Run unit and smoke tests for tooling scripts |
| `make i18n.check` | i18n | Scan runtime hardcoded text and validate language pack key coverage |
| `make plugins.init` | Plugin workspace | Convert official plugin submodules into ordinary directories |
| `make plugins.install` | Plugin workspace | Install configured source plugins into `apps/lina-plugins` |
| `make plugins.update` | Plugin workspace | Update source plugins under `apps/lina-plugins` |
| `make plugins.status` | Plugin workspace | Show source plugin workspace status |
| `make agents` | Agent resources | Create or remove skill, prompt, and `AGENTS.md` symlinks for a single agent |
| `make agents.skills.link` | Agent resources | Symlink supported agents' project skill directories to `.agents/skills` |
| `make agents.skills.unlink` | Agent resources | Remove symlinks managed by `agents.skills.link` |
| `make agents.prompts.link` | Agent resources | Symlink supported agents' command or prompt directories to `.agents/prompts/...` |
| `make agents.prompts.unlink` | Agent resources | Remove symlinks managed by `agents.prompts.link` |
| `make agents.md.link` | Agent resources | Symlink supported agents' private rule files to the root `AGENTS.md` |
| `make agents.md.unlink` | Agent resources | Remove the `AGENTS.md` rule-file symlinks managed by `agents.md.link` |
| `make init` | Database | Initialize database schema and seed data |
| `make mock` | Database | Load demo `Mock` data |
| `make release.tag.check` | Release governance | Verify the `release tag` matches `framework.version` in `metadata.yaml` |
| `make help` | Other | Show all available commands |

## Environment Management

### make env.check

Checks whether the local development environment satisfies the default development workflow requirements. This command only reads tool versions and database connection information; it does not modify the workspace. `Vite` and `Playwright` are checked from the project's local dependencies, while the `PostgreSQL` version is queried through the `database.default.link` connection configured in `apps/lina-core/manifest/config/config.yaml`.

```bash
make env.check
```

The current checks are:

| Check | Minimum version | Description |
|-------|-----------------|-------------|
| `Go` | `1.25.0` | `Go` toolchain used to build the backend, tooling, and `WASM` plugins |
| `Node.js` | `20.19.0` | `Node.js` runtime required for frontend development and builds |
| `pnpm` | `10.0.0` | Frontend dependency manager |
| `Vite` | `7.3.1` | Project-local frontend build tool; run `make env.setup` first if missing |
| `Playwright` | `1.58.2` | `E2E` test runner; run `make env.setup` first if missing |
| `PostgreSQL` | `14.0.0` | Server version detected through `database.default.link` |

### make env.setup

Installs the frontend dependencies, the `Playwright Chromium` browser, and the system dependencies required by the browser for development and `E2E` testing. Run it once after the first clone or during `CI` environment initialization; it is usually not needed again.

```bash
make env.setup
```

## Development Servers

### make dev

Restarts the backend and frontend development servers. Before starting, it checks that the command-line `backend_port`, backend `server.address`, and frontend `Vite proxy target` agree, so port mismatches fail early instead of surfacing later as health-check timeouts or API errors. The command then stops existing services, decides whether to build `WASM` plugins based on plugin mode, prepares frontend static assets, compiles the backend, and waits for both health checks to pass.

```bash
make dev

# Force host mode and skip official source plugins and WASM plugin builds
make dev plugins=0

# Force official source plugin mode
make dev plugins=1
```

`plugins=auto` is the default mode: if `apps/lina-plugins/` contains usable plugin `manifest` files, official plugin mode is enabled automatically; otherwise host mode is used. When invoking `linactl dev` directly, you can also pass `skip_wasm=true` to skip only the `WASM` build step.

```bash
cd hack/tools/linactl
go run . dev skip_wasm=true
```

The backend listens on `http://localhost:9120` by default, and the frontend listens on `http://localhost:5666`. Logs are written to `temp/lina-core.log` and `temp/lina-vben.log`.

### make stop

Stops the backend and frontend development servers and cleans up stale `PID` files. Any remaining processes still occupying the ports are force-terminated.

```bash
make stop
```

### make status

Prints the current backend/frontend running status and log file paths, making it easy to confirm whether both services started correctly.

```bash
make status
```

Sample output:

```text
+----------+---------+------------------------+-------+------------------------+--------------------+
| Service  | Status  | URL                    | PID   | PID File               | Log File           |
+----------+---------+------------------------+-------+------------------------+--------------------+
| Backend  | running | http://127.0.0.1:9120/ | 87739 | temp/pids/backend.pid  | temp/lina-core.log |
| Frontend | running | http://127.0.0.1:5666/ | 87740 | temp/pids/frontend.pid | temp/lina-vben.log |
+----------+---------+------------------------+-------+------------------------+--------------------+
```

## Code Builds

### make build

Runs the full build pipeline: frontend static asset build, static asset and `manifest` preparation for backend embedding, dynamic `WASM` plugin builds based on plugin mode, and finally compilation of the backend host binary. Build artifacts are written to `temp/output/`.

```bash
# Default build for the current platform
make build

# Force host mode without building official source plugins
make build plugins=0

# Target specific platforms through cross-compilation
make build platforms=linux/amd64,linux/arm64

# Override output directory and binary name
make build output_dir=temp/release binary_name=linapro

# Override the config file
make build config=hack/config.yaml

# Enable verbose logs
make build verbose=1
# or
make build v=1
```

Build defaults are managed centrally in the root `hack/config.yaml`. Command-line arguments override the corresponding fields.

```yaml
build:
  # Target platform list in goos/goarch format; make build platforms=... overrides this
  platforms:
    - "auto"
  # Whether to enable CGO
  cgoEnabled: false
  # Build output path, relative to the repository root
  outputDir: "temp/output"
  # Filename of the compiled binary
  binaryName: "lina"
```

| Field | Default | Description |
|-------|---------|-------------|
| `build.platforms` | `["auto"]` | Target platform list in `goos/goarch` format; `auto` means `linux/<current-arch>`; override with `make build platforms=...` |
| `build.cgoEnabled` | `false` | Whether to enable `CGO` |
| `build.outputDir` | `temp/output` | Build output path, relative to the repository root |
| `build.binaryName` | `lina` | Host binary filename |

Plugin build mode is controlled by the `plugins` argument:

| `plugins` value | Description |
|-----------------|-------------|
| `auto` (default) | Enable official source plugin mode when `apps/lina-plugins/` contains usable plugin `manifest` files |
| `0` | Force host mode, remove official plugin build tags, and skip official plugin `WASM` builds |
| `1` | Force official source plugin mode; fail fast if the plugin workspace is unavailable |

### make wasm

Builds runtime `WASM` plugins separately. `make wasm` is a compatibility entry that writes artifacts to `temp/output/` by default, supports `p=<plugin-id>` to build only one plugin, and supports `dry_run=true` to check the buildable plugin plan without writing artifacts. To build a single plugin directory at any path, invoke `linactl wasm plugin_dir=<path>` directly.

```bash
# Build all WASM plugins
make wasm

# Build only a specific plugin, where plugin-id is the plugin directory name
make wasm p=my-plugin

# Check the build plan only
make wasm dry_run=true

# Build a specific plugin directory directly
cd hack/tools/linactl
go run . wasm plugin_dir=../../apps/lina-plugins/my-plugin out=../../temp/output
```

### make pack.assets

Prepares host `manifest` assets for `Go` embedding. This command refreshes `config`, `sql`, and `i18n` resources under `apps/lina-core/internal/packed/manifest/`. It is normally called automatically by `make build` or `make dev`, but can be run manually when you need to inspect or prepare embedded resources in isolation:

```bash
make pack.assets
```

### make tidy

Tidies `Go` module dependencies for the host, development tools, and plugins. Run it after dependency upgrades or after initializing full plugin mode:

```bash
make tidy
```

## Image Builds

### make image

Runs the full `Docker` image build pipeline: first performs the image build preflight, then runs `make build` to generate all build artifacts, and finally calls `hack/tools/image-builder` to package them into an image. Image name, tag, registry, base image, and plugin build mode are all configurable through arguments.

```bash
# Build with default configuration
make image

# Specify a tag and image registry
make image tag=v0.6.0 registry=ghcr.io/linaproai

# Build and push immediately
make image tag=v0.6.0 registry=ghcr.io/linaproai push=1

# Multi-platform build
make image platforms=linux/amd64,linux/arm64 tag=v0.6.0

# Override the runtime base image and use host mode
make image base_image=alpine:3.22 plugins=0
```

Image build defaults are also managed by `hack/config.yaml`; command-line arguments override values for the current invocation.

```yaml
image:
  # Image name; registry prefix is prepended at build time
  name: "linapro"
  # Default tag; leave empty to derive it from git metadata
  tag: "dev"
  # Remote registry prefix, for example ghcr.io/linaproai
  registry: ""
  # Whether to push by default; push=1 overrides this invocation
  push: false
  # Runtime base image
  baseImage: "alpine:3.22"
  # Dockerfile path, relative to the repository root
  dockerfile: "hack/docker/Dockerfile"
```

| Field | Default | Description |
|-------|---------|-------------|
| `image.name` | `linapro` | Image name; the `registry` prefix is prepended during build |
| `image.tag` | `dev` | Default tag; leave empty to derive it from `git` metadata |
| `image.registry` | Empty | Remote registry prefix, for example `ghcr.io/linaproai` |
| `image.push` | `false` | Whether to push by default; command-line `push=1` overrides the current invocation |
| `image.baseImage` | `alpine:3.22` | Runtime base image |
| `image.dockerfile` | `hack/docker/Dockerfile` | `Dockerfile` path relative to the repository root |

### make image.build

Prepares all image build artifacts (equivalent to running `make build` and generating the image build context), but does not run the `Docker build` step. Use it when you need to inspect artifacts manually or customize the image build process.

```bash
make image.build
```

## Test Management

### make test

Runs the full `Playwright E2E` test suite. Make sure the development services have been started with `make dev` first. Use the `scope` argument to narrow the test scope:

| `scope` value | Description |
|---------------|-------------|
| `full` (default) | Run all `E2E` tests |
| `host` | Run only host-owned tests |
| `plugins` | Run all official plugin tests |
| `plugin:<id>` | Run tests for a specific plugin |

```bash
make test

# Run only host tests
make test scope=host

# Run only tests for a specific plugin
make test scope=plugin:multi-tenant
```

### make test.go

Runs unit tests for all maintained `Go` modules with race detection and verbose logs enabled by default. The command first discovers modules in the current `Go workspace`, runs real tests for packages that contain test files, performs compile smoke checks for packages without test files, and prints a per-module summary. Use `plugins=0` to force host mode, or `race=false` to disable race detection.

```bash
make test.go
make test.go plugins=0
make test.go race=false
make test.go verbose=false
```

### make test.host

Runs only the host-owned `Playwright E2E` tests and does not require official plugin submodules to be initialized.

```bash
make test.host
```

### make test.plugins

Runs the official plugins' own `Playwright E2E` tests. Initialize the `apps/lina-plugins/` submodules before running this command.

```bash
make test.plugins
```

### make test.scripts

Runs unit and smoke checks for cross-platform repository tooling, validating the basic correctness of `linactl`, `make.cmd`, and other helper entries.

```bash
make test.scripts
```

## Plugin Management

The `apps/lina-plugins/` directory stores official plugins. It can be a `Git submodule`, or it can be managed as a source plugin workspace through the plugin workspace commands below.

### make plugins.init

Converts `apps/lina-plugins` from a `Git submodule` into an ordinary plugin directory while preserving the plugin code. After conversion, plugin code can be modified and committed freely without submodule constraints.

```bash
make plugins.init
```

### make plugins.install

Clones selected plugins from remote repositories into `apps/lina-plugins/` according to `plugins.sources` in `hack/config.yaml`. Use `p=<plugin-id>` to install only one plugin, or `source=<name>` to process only one configured source.

```bash
# Install all configured source plugins
make plugins.install

# Install only one plugin
make plugins.install p=multi-tenant

# Force overwrite an existing plugin directory
make plugins.install force=1
```

### make plugins.update

Updates configured source plugins under `apps/lina-plugins/` by pulling the latest remote version. Plugins with uncommitted local changes are blocked from updating unless `force=1` is passed.

```bash
make plugins.update
make plugins.update p=multi-tenant
make plugins.update force=1
```

### make plugins.status

Shows the current official plugin workspace status, including configured plugin versions, local changes, and remote update state.

```bash
make plugins.status
```

## i18n

### make i18n.check

Scans runtime-visible code paths for hardcoded text that has not been added to the internationalization system, and validates message `key` coverage across host and plugin runtime language packs. Run it before submitting new features to check i18n compliance.

```bash
make i18n.check
```

## AI Tool Integration

The `agents` command family manages repository-local resource symlinks for different `AI Coding Agent` tools. The repository uses `.agents/skills`, `.agents/prompts/`, and the root `AGENTS.md` as canonical resource sources, then creates symlinks at tool-specific paths such as `.claude/skills`, `.codex/prompts/opsx`, `CLAUDE.md`, or `GEMINI.md`.

These commands manage only the symlinks they create. Unlink commands do not delete real directories or files, and they do not remove unmanaged symlinks that point elsewhere. If a symlink already exists but points to a different target, pass `FORCE=1` to rebuild it.

### make agents

Recommended aggregate entry. When called without arguments from an interactive terminal, it first lets you choose an agent with arrow keys, then choose the `link` or `unlink` action. In non-interactive environments, specify a single agent with `agent=<name>`. The aggregate entry applies the same action to every resource type supported by that agent; it does not support `agent=all` or comma-separated lists.

```bash
# Choose agent and action interactively
make agents

# Create every available resource symlink for one agent
make agents agent=claude-code

# Remove managed symlinks for one agent
make agents agent=claude-code action=unlink

# Rebuild managed symlinks that point to the wrong target
make agents agent=claude-code force=1
```

If an agent natively reads a resource type, such as `AGENTS.md`, the aggregate entry skips that resource and explains why in the summary.

### make agents.skills.link

Symlinks supported agents' project skill directories to the canonical `.agents/skills` source. Without `agent`, a non-interactive environment prints support status and guidance, while an interactive terminal opens the selection flow. Use `agent=<name|all|csv>` to process one or more agents.

```bash
# Show skill symlink status
make agents.skills.link

# Create the skill symlink for one agent
make agents.skills.link agent=claude-code

# Create skill symlinks for every symlink-capable agent
make agents.skills.link agent=all

# Rebuild a skill symlink that points to the wrong target
make agents.skills.link agent=claude-code force=1
```

### make agents.skills.unlink

Removes skill directory symlinks managed by `agents.skills.link`. Non-interactive environments must pass `agent=<name|all|csv>` explicitly; interactive terminals can choose from existing managed symlinks.

```bash
make agents.skills.unlink agent=claude-code
make agents.skills.unlink agent=all
```

### make agents.prompts.link

Symlinks supported agents' command or prompt directories to canonical sources under `.agents/prompts/`. This is mainly used to bridge `.agents/prompts/opsx` to each tool's own command directory, such as `.claude/commands/opsx`, `.cursor/commands/opsx`, `.codex/prompts/opsx`, or `.gemini/commands/opsx`.

```bash
# Show prompt symlink status
make agents.prompts.link

# Create prompt symlinks for one agent
make agents.prompts.link agent=codex

# Create prompt symlinks in batches
make agents.prompts.link agent=claude-code,codex,cursor,gemini-cli
```

### make agents.prompts.unlink

Removes prompt directory symlinks managed by `agents.prompts.link` without deleting the real prompt directories.

```bash
make agents.prompts.unlink agent=codex
make agents.prompts.unlink agent=all
```

### make agents.md.link

Symlinks supported agents' private rule files to the root `AGENTS.md`. For example, `claude-code` maps to `CLAUDE.md`, `gemini-cli` maps to `GEMINI.md`, `qwen-code` maps to `QWEN.md`, and `junie` maps to `.junie/guidelines.md`. Agents that natively read `AGENTS.md` are shown in status output and do not need symlinks.

```bash
# Show AGENTS.md rule-file symlink status
make agents.md.link

# Create the rule-file symlink for one agent
make agents.md.link agent=claude-code

# Create rule-file symlinks for every supported agent
make agents.md.link agent=all
```

### make agents.md.unlink

Removes rule-file symlinks managed by `agents.md.link`. This command does not delete real hand-written files such as `CLAUDE.md` or `GEMINI.md`.

```bash
make agents.md.unlink agent=claude-code
make agents.md.unlink agent=all
```

## Database

:::caution Destructive Operations

Both `init` and `mock` perform destructive database operations. They require an explicit `confirm` argument to prevent accidental execution.

:::

### make init

Initializes the database schema (`DDL`) and required system seed data. The command reads `database.default.link` from `apps/lina-core/manifest/config/config.yaml`; currently only the `PostgreSQL 14+` dialect is supported. `sqlite:`, `mysql:`, or unknown links fail fast during dialect parsing without creating local database files or continuing to execute `SQL`.

```bash
# Initialize only, preserving existing data
make init confirm=init

# Rebuild the database after wiping it
make init confirm=init rebuild=true
```

If `PostgreSQL` cannot be reached, the command prompts you to start `PostgreSQL` and prints a local `docker run` example. `rebuild=true` first terminates connections to the target database, then runs `DROP DATABASE IF EXISTS` and `CREATE DATABASE`; use it only when the target database can be cleared.

### make mock

After `make init` completes, loads optional `Mock` data for local demos and development validation.

```bash
make mock confirm=mock
```

## Other

### make help

Prints the default cross-platform development command list, sorted by command name. `linactl` also registers maintenance commands such as `cli`, `cli.install`, `ctrl`, and `dao`; default help does not show them. Run `go run . help --all` directly to list those commands.

```bash
make help
```

### make release.tag.check

Verifies that the release tag matches `framework.version` in `apps/lina-core/manifest/config/metadata.yaml`. When `tag` is not passed explicitly, the command reads `GITHUB_REF_NAME`, making it suitable for release pipeline validation.

```bash
make release.tag.check tag=v0.6.0

# Print only framework.version from metadata.yaml
make release.tag.check print_version=1
```
