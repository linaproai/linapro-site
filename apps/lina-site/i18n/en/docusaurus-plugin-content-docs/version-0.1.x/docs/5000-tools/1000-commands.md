---
slug: '/docs/commands'
title: 'Development Commands'
hide_title: true
description: 'A complete reference for the cross-platform development command set in the LinaPro project — covering linactl, Makefile compatibility entry, Windows make.cmd wrapper, development server management, full builds, WASM plugin builds, Docker image builds, plugin workspace management, testing, i18n validation, database initialization, and release governance on macOS, Linux, and Windows.'
keywords:
  - make commands
  - development commands
  - build commands
  - LinaPro development
  - make dev
  - make dev.setup
  - make build
  - make pack.assets
  - make test
  - make image
  - make image.build
  - make db.init
  - make wasm
  - make plugins.install
  - make plugins.status
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

The `LinaPro` project provides a cross-platform development command set. Long-term task orchestration lives in `hack/tools/linactl`, implemented as a `Go` program; the root `Makefile` and `make.cmd` are compatibility entries that forward to the underlying `linactl`. This means the same commands work on `macOS`, `Linux`, and `Windows`, without depending on `GNU Make` or `POSIX Shell` as the sole entry point.

## Platform Notes

**Cross-platform native commands**: All platforms can use `linactl` directly:

```bash
cd hack/tools/linactl
go run . help
go run . status
go run . init confirm=init
go run . dev
```

**macOS / Linux**: The root `make` compatibility entry continues to work:

```bash
make help
make db.init confirm=init
make dev
```

**Windows cmd.exe**: Use the `make.cmd` wrapper at the project root. In `cmd.exe`, executable extensions are resolved from the current directory, so the `.cmd` suffix can be omitted:

```cmd
make dev
make build
make help
```

**Windows PowerShell**: Requires a current-directory prefix. In a default `Windows` environment, `.\make` works; to avoid confusion with any locally installed `make`, use `.\make.cmd` explicitly:

```powershell
.\make help
.\make db.init confirm=init
.\make dev
```

All `make <command>` examples in this document can be equivalently replaced with `cd hack/tools/linactl && go run . <command>` — argument formats remain the same.

## Command Overview

| Command | Category | Description |
|---------|----------|-------------|
| `make dev.setup` | Development server | Install frontend dependencies and `Playwright` browsers |
| `make dev` | Development server | Restart frontend and backend development servers |
| `make stop` | Development server | Stop frontend and backend development servers |
| `make status` | Development server | Show running status and log paths for both servers |
| `make build` | Build | Full build: frontend, plugins, and backend binary |
| `make pack.assets` | Build | Prepare frontend static assets and `manifest` for core framework embedding |
| `make wasm` | Build | Build all or specific runtime `WASM` plugins |
| `make tidy` | Build | Tidy `Go` module dependencies for the core framework, tools, and plugins |
| `make image` | Image | Build a production `Docker` image |
| `make image.build` | Image | Prepare image artifacts only, skip the `Docker` build step |
| `make test` | Test | Run the full `E2E` test suite |
| `make test.go` | Test | Run `Go` unit tests |
| `make test.host` | Test | Run core framework-only `E2E` tests |
| `make test.plugins` | Test | Run official plugin `E2E` tests |
| `make test.scripts` | Test | Run unit and smoke tests for tooling scripts |
| `make i18n.check` | i18n | Scan for hardcoded strings and validate language pack key coverage |
| `make plugins.init` | Plugin workspace | Convert official plugin submodules to ordinary directories |
| `make plugins.install` | Plugin workspace | Install configured source plugins into `apps/lina-plugins` |
| `make plugins.update` | Plugin workspace | Update source plugins in `apps/lina-plugins` |
| `make plugins.status` | Plugin workspace | View source plugin workspace status |
| `make db.init` | Database | Initialize database schema and seed data |
| `make db.mock` | Database | Load demo `Mock` data |
| `make release.tag.check` | Release governance | Verify `release tag` matches `framework.version` in `metadata.yaml` |
| `make help` | Other | List all available commands |

## Development Server

### make dev.setup

Installs all prerequisites for development and `E2E` testing, including frontend `npm` packages and `Playwright` browsers. Run once after the initial clone or in `CI` environment initialization — no need to repeat.

```bash
make dev.setup
```

### make dev

Restarts the backend and frontend dev servers. Before starting, it stops any running instances, then sequentially builds `WASM` plugins, prepares frontend static assets, and compiles the backend. It waits for both servers to pass health checks before printing their status. Use `skip_wasm=true` to skip the `WASM` build step for faster startup.

```bash
make dev

# Skip WASM plugin build (for development scenarios not involving dynamic plugins)
make dev skip_wasm=true
```

The backend listens on `http://localhost:8080` by default; the frontend listens on `http://localhost:5666`. Logs are written to `temp/lina-core.log` and `temp/lina-vben.log` respectively.

### make stop

Stops the backend and frontend dev servers and cleans up any leftover `PID` files. Zombie processes still holding the ports are forcibly terminated.

```bash
make stop
```

### make status

Prints the current running status and log file paths for both servers, giving you a quick way to confirm whether they started correctly.

```bash
make status
```

Sample output:

```text
+----------+---------+------------------------+-------+------------------------+--------------------+
| Service  | Status  | URL                    | PID   | PID File               | Log File           |
+----------+---------+------------------------+-------+------------------------+--------------------+
| Backend  | running | http://127.0.0.1:8080/ | 87739 | temp/pids/backend.pid  | temp/lina-core.log |
| Frontend | running | http://127.0.0.1:5666/ | 87740 | temp/pids/frontend.pid | temp/lina-vben.log |
+----------+---------+------------------------+-------+------------------------+--------------------+
```

## Build

### make build

Runs the full build pipeline in order: frontend static asset build, `manifest` resource preparation for backend embedding, all `WASM` plugin builds, and finally compilation of the backend core framework binary. Build artifacts are written to `temp/output/`.

```bash
# Default build (current platform)
make build

# Target specific platforms (cross-compilation)
make build platforms=linux/amd64,linux/arm64

# Enable verbose logging
make build verbose=1
# or
make build v=1
```

Default values for the build are managed centrally in `hack/config.yaml` at the repository root. Command-line arguments override the corresponding fields.

```yaml
build:
  # Target platform list in goos/goarch format; overridable with make build platforms=...
  platforms:
    - "auto"
  # Whether to enable CGO
  cgoEnabled: false
  # Build output path, relative to the repository root
  outputDir: "temp/output"
  # Filename of the compiled core framework binary
  binaryName: "lina"
```

| Field | Default | Description |
|-------|---------|-------------|
| `build.platforms` | `["auto"]` | Target platform list in `goos/goarch` format; `auto` means `linux/<current-arch>`; override with `make build platforms=...` |
| `build.cgoEnabled` | `false` | Whether to enable `CGO` |
| `build.outputDir` | `temp/output` | Build output path, relative to the repository root |
| `build.binaryName` | `lina` | Core framework binary filename |

### make wasm

Builds runtime `WASM` plugins separately. `make wasm` is a compatibility entry that outputs artifacts to `temp/output/` by default, with `p=<plugin-id>` support for building a specific plugin. To override the output directory or perform a build-only probe, use `linactl wasm` directly.

```bash
# Build all WASM plugins
make wasm

# Build a specific plugin (plugin-id is the plugin directory name)
make wasm p=my-plugin
```

### make pack.assets

Prepares core framework `manifest` assets for `Go` embedding, usually called automatically by `make build` or `make dev`. Run manually when you need to inspect or prepare embedded resources in isolation:

```bash
make pack.assets
```

### make tidy

Tidies `Go` module dependencies for the core framework, dev tools, and plugins. Useful after upgrading dependencies or initializing full plugin mode:

```bash
make tidy
```

## Container Images

### make image

Runs the complete `Docker` image build pipeline: first executes `make build` to generate all artifacts, then calls `hack/tools/image-builder` to package them into an image. The image name, tag, registry, and other settings are all configurable.

```bash
# Build with default configuration
make image

# Specify a tag and registry
make image tag=v0.6.0 registry=ghcr.io/linaproai

# Build and push immediately
make image tag=v0.6.0 registry=ghcr.io/linaproai push=1

# Multi-platform build
make image platforms=linux/amd64,linux/arm64 tag=v0.6.0
```

Image build defaults are also managed by `hack/config.yaml`. Command-line arguments override the values for the current invocation.

```yaml
image:
  # Image name; the registry prefix is prepended at build time
  name: "linapro"
  # Default tag; leave empty to derive one from git metadata
  tag: "dev"
  # Remote registry prefix, e.g. ghcr.io/linaproai
  registry: ""
  # Whether to push by default; pass push=1 to override for one run
  push: false
  # Runtime base image
  baseImage: "alpine:3.22"
  # Dockerfile path, relative to the repository root
  dockerfile: "hack/docker/Dockerfile"
```

| Field | Default | Description |
|-------|---------|-------------|
| `image.name` | `linapro` | Image name; the `registry` prefix is prepended at build time |
| `image.tag` | `dev` | Default tag; leave empty to derive one from git metadata |
| `image.registry` | _(empty)_ | Remote registry prefix, e.g. `ghcr.io/linaproai` |
| `image.push` | `false` | Whether to push by default; `push=1` overrides for a single run |
| `image.baseImage` | `alpine:3.22` | Runtime base image |
| `image.dockerfile` | `hack/docker/Dockerfile` | `Dockerfile` path, relative to the repository root |

### make image.build

Prepares all image build artifacts (equivalent to running `make build`) without executing the `Docker build` step. Useful when you need to inspect artifacts manually or customize the image build process.

```bash
make image.build
```

## Testing

### make test

Runs the full `Playwright E2E` test suite. Make sure the dev servers are running via `make dev` before executing. Use the `scope` parameter to narrow the test scope:

| `scope` value | Description |
|---------------|-------------|
| `full` (default) | Run all `E2E` tests |
| `host` | Run core framework-only tests |
| `plugins` | Run all official plugin tests |
| `plugin:<id>` | Run tests for a specific plugin |

```bash
make test

# Core framework tests only
make test scope=host

# Specific plugin tests only
make test scope=plugin:multi-tenant
```

### make test.go

Runs unit tests for all maintained `Go` modules with race detection enabled. Pass `plugins=0` to force core framework-only mode, or `race=false` to disable race detection.

```bash
make test.go
make test.go plugins=0
make test.go race=false
```

### make test.host

Runs only the core framework-owned `Playwright E2E` tests. Does not require the official plugin submodules to be initialized.

```bash
make test.host
```

### make test.plugins

Runs the official plugins' own `Playwright E2E` tests. Requires the `apps/lina-plugins/` submodules to be initialized first.

```bash
make test.plugins
```

### make test.scripts

Runs unit and smoke tests for cross-platform repository tooling, verifying the basic correctness of `linactl`, `make.cmd`, and other helper entry points.

```bash
make test.scripts
```

## i18n

### make i18n.check

Scans runtime-visible code paths for hardcoded strings not covered by the i18n system, and validates message key coverage across the core framework and plugin runtime language packs. Run this before committing new features to catch compliance issues early.

```bash
make i18n.check
```

## Plugin Workspace

The `apps/lina-plugins/` directory holds official plugins — it can be either a `Git submodule` or managed as a source plugin directory through the plugin workspace commands. The following commands provide full lifecycle management of the plugin workspace:

### make plugins.init

Converts `apps/lina-plugins` from a `Git submodule` to an ordinary plugin directory while preserving the plugin code. After conversion, you can freely modify plugin code and commit changes without submodule constraints.

```bash
make plugins.init
```

### make plugins.install

Clones configured plugins from remote repositories into `apps/lina-plugins/` based on the `plugins.sources` configuration in `hack/config.yaml`. Supports `p=<plugin-id>` to install a single plugin and `source=<name>` to process only a specific source.

```bash
# Install all configured source plugins
make plugins.install

# Install a specific plugin only
make plugins.install p=multi-tenant

# Force-overwrite existing plugin directories
make plugins.install force=1
```

### make plugins.update

Updates configured source plugins in `apps/lina-plugins/`, pulling the latest version from the remote. Plugins with uncommitted local changes are blocked from updating unless `force=1` is passed.

```bash
make plugins.update
make plugins.update p=multi-tenant
make plugins.update force=1
```

### make plugins.status

Views the current official plugin workspace status, including configured plugin versions, local changes, and remote update state.

```bash
make plugins.status
```

## Database

:::caution Destructive operations

Both `init` and `mock` make destructive changes to the database. To prevent accidental execution, they require an explicit `confirm` argument.

:::

### make db.init

Initializes the database schema (`DDL`) and system seed data. The backend automatically selects the `PostgreSQL` or `SQLite` dialect based on the `database.default.link` setting in `config.yaml`; `PostgreSQL 14+` is the default data store, while `SQLite` is only for local demos or smoke testing.

```bash
# Initialize only (preserve existing data)
make db.init confirm=init

# Rebuild the database (wipe and reinitialize)
make db.init confirm=init rebuild=true
```

### make db.mock

After `make db.init` completes, loads optional `Mock` data for local demo and development verification.

```bash
make db.mock confirm=mock
```

## Other Commands

### make help

Prints all available commands from the root `Makefile` and every included module file, sorted by command name.

```bash
make help
```
