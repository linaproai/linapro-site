---
slug: '/docs/commands'
title: 'Dev Commands'
hide_title: true
description: 'A complete reference for the cross-platform dev command set in the LinaPro project — covering linactl, Makefile compatibility entry, Windows make.cmd wrapper, dev server management, full builds, WASM plugin builds, Docker image builds, testing, i18n validation, plugin tools, and database initialization on macOS, Linux, and Windows.'
keywords:
  - make commands
  - dev commands
  - build commands
  - LinaPro development
  - make dev
  - make build
  - make test
  - make image
  - make init
  - make wasm
  - linactl
  - development workflow
  - backend build
  - frontend build
  - Docker image
  - E2E testing
  - database initialization
  - Windows support
  - make.cmd
  - cross-platform
---

The `LinaPro` project provides a cross-platform dev command set. Long-term task orchestration lives in `hack/tools/linactl`, implemented as a `Go` program; the root `Makefile` and `make.cmd` are compatibility entries that forward to the underlying `linactl`. This means the same commands work on `macOS`, `Linux`, and `Windows`, without depending on `GNU Make` or `POSIX Shell` as the sole entry point.

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
make init confirm=init
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
.\make init confirm=init
.\make dev
```

All `make <command>` examples in this document can be equivalently replaced with `cd hack/tools/linactl && go run . <command>` — argument formats remain the same.

## Command Overview

| Command | Category | Description |
|---------|----------|-------------|
| `make dev` | Dev server | Restart the frontend and backend dev servers |
| `make stop` | Dev server | Stop the frontend and backend dev servers |
| `make status` | Dev server | Show the running status and log paths for both servers |
| `make build` | Build | Full build: frontend, plugins, and backend binary |
| `make wasm` | Build | Build all or specific runtime `WASM` plugins |
| `make tidy` | Build | Tidy `Go` module dependencies for host, tools, and plugins |
| `make image` | Image | Build a production `Docker` image |
| `make image-build` | Image | Prepare image artifacts only, skip the `Docker` build step |
| `make test` | Test | Run the full `E2E` test suite |
| `make test.go` | Test | Run `Go` unit tests |
| `make test.host` | Test | Run host-only `E2E` tests |
| `make test.plugins` | Test | Run official plugin `E2E` tests |
| `make test.scripts` | Test | Run unit and smoke tests for tooling scripts |
| `make i18n.check` | i18n | Scan for hardcoded strings and validate language pack key coverage |
| `make init` | Database | Initialize database schema and seed data |
| `make mock` | Database | Load demo `Mock` data |
| `make help` | Other | List all available commands |

`linactl` also provides `plugins.init`, `plugins.install`, `plugins.update`, and `plugins.status` for advanced scenarios such as converting official plugin submodules into regular plugin directories, installing configured source plugins, or inspecting the plugin workspace state.

## Plugin Mode Parameters

The official plugin directory `apps/lina-plugins/` is a `Git submodule`. When this directory is initialized and contains plugin manifests, the `dev`, `build`, `image`, and related `Go` test commands automatically enter full plugin mode. A `temp/go.work.plugins` file (git-ignored) is generated from the root host-only `go.work`, and source plugin modules are resolved via `GOWORK`.

If you only need to run the main framework, skip the submodule initialization or pass `plugins=0` to force host-only mode:

```bash
make dev plugins=0
make build plugins=0
make image plugins=0
```

Before building or testing official plugins, initialize the submodules first:

```bash
git submodule update --init --recursive
```

## Dev Server

### make dev

Restarts the backend and frontend dev servers. Before starting, it stops any running instances, then sequentially builds the `WASM` plugins, prepares frontend static assets, and compiles the backend. It waits for both servers to pass health checks before printing their status.

```bash
make dev
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
╔══════════════════════════════════════════════╗
║         LinaPro Framework Status             ║
╠══════════════════════════════════════════════╣
║  Backend:  ✓ running  http://localhost:8080  ║
║  Frontend: ✓ running  http://localhost:5666  ║
╠══════════════════════════════════════════════╣
║  Backend log:   temp/lina-core.log           ║
║  Frontend log:  temp/lina-vben.log           ║
╚══════════════════════════════════════════════╝
```

## Build

### make build

Runs the full build pipeline in order: frontend static asset build, `manifest` resource preparation for backend embedding, all `WASM` plugin builds, and finally compilation of the backend host binary. Build artifacts are written to `temp/output/`.

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
  # Filename of the compiled host binary
  binaryName: "lina"
```

| Field | Default | Description |
|-------|---------|-------------|
| `build.platforms` | `["auto"]` | Target platform list in `goos/goarch` format; `auto` means `linux/<current-arch>`; override with `make build platforms=...` |
| `build.cgoEnabled` | `false` | Whether to enable `CGO` |
| `build.outputDir` | `temp/output` | Build output path, relative to the repository root |
| `build.binaryName` | `lina` | Host binary filename |

### make wasm

Builds runtime `WASM` plugins separately. `make wasm` is a compatibility entry that outputs artifacts to `temp/output/` by default, with `p=<plugin-id>` support for building a specific plugin. To override the output directory or perform a build-only probe, use `linactl wasm` directly.

```bash
# Build all WASM plugins
make wasm

# Build a specific plugin (plugin-id is the plugin directory name)
make wasm p=my-plugin

# Specify output directory (linactl native command)
cd hack/tools/linactl
go run . wasm p=my-plugin out=../../temp/output

# Dry-run: list buildable dynamic plugins (linactl native command)
go run . wasm dry_run=true
```

### linactl prepare-packed-assets

Prepares frontend static assets and `manifest` resources for backend embedding. Usually called automatically by `make build` or `make dev`. Run manually when you need to inspect embedded assets in isolation:

```bash
cd hack/tools/linactl
go run . prepare-packed-assets
```

### make tidy

Tidies `Go` module dependencies for the host, dev tools, and plugins. Useful after upgrading dependencies or initializing full plugin mode:

```bash
make tidy
```

## Image

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

### make image-build

Prepares all image build artifacts (equivalent to running `make build`) without executing the `Docker build` step. Useful when you need to inspect artifacts manually or customize the image build process.

```bash
make image-build
```

## Test

### make test

Runs the full `Playwright E2E` test suite under `hack/tests/`. Make sure the dev servers are running via `make dev` before executing.

```bash
make test
```

### make test.go

Runs unit tests for all maintained `Go` modules. Pass `plugins=0` to force host-only mode.

```bash
make test.go
make test.go plugins=0
```

### make test.host

Runs only the host's own `Playwright E2E` tests. Does not require the official plugin submodules to be initialized.

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

Scans runtime-visible code paths for hardcoded strings not covered by the i18n system, and validates message key coverage across the host and plugin runtime language packs. Run this before committing new features to catch compliance issues early.

```bash
make i18n.check
```

## Plugin Tools

### linactl plugins.status

View the current official plugin workspace status. Advanced developers can also use `plugins.init`, `plugins.install`, and `plugins.update` to manage source plugin workspaces in non-submodule form.

```bash
cd hack/tools/linactl
go run . plugins.status
```

## Database

:::caution Destructive operations

Both `init` and `mock` make destructive changes to the database. To prevent accidental execution, they require an explicit `confirm` argument.

:::

### make init

Initializes the database schema (`DDL`) and system seed data. The backend automatically selects the `PostgreSQL` or `SQLite` dialect based on the `database.default.link` setting in `config.yaml`; `PostgreSQL 14+` is the default data store, while `SQLite` is only for local demos or smoke testing.

```bash
# Initialize only (preserve existing data)
make init confirm=init

# Rebuild the database (wipe and reinitialize)
make init confirm=init rebuild=true
```

### make mock

After `make init` completes, loads optional `Mock` data for local demo and development verification.

```bash
make mock confirm=mock
```

## Other

### make help

Prints all available commands from the root `Makefile` and every included module file, sorted by command name.

```bash
make help
```
