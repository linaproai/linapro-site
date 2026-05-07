---
slug: '/docs/commands'
title: 'Dev Commands'
hide_title: true
description: 'A complete reference for every make command in the LinaPro project — what each command does, what options it accepts, and usage examples covering dev server management, full builds, WASM plugin builds, Docker image builds, testing, i18n validation, and database initialization.'
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
  - development workflow
  - backend build
  - frontend build
  - Docker image
  - E2E testing
  - database initialization
---

The LinaPro project root provides a complete set of `make` commands, organized across module files under `hack/makefiles/`. Run `make help` in the project root at any time to see all available commands.

## Command Overview

| Command | Category | Description |
|---------|----------|-------------|
| `make dev` | Dev server | Restart the frontend and backend dev servers |
| `make stop` | Dev server | Stop the frontend and backend dev servers |
| `make status` | Dev server | Show the running status and log paths for both servers |
| `make build` | Build | Full build: frontend, plugins, and backend binary |
| `make wasm` | Build | Build all or specific runtime `WASM` plugins |
| `make image` | Image | Build a production `Docker` image |
| `make image-build` | Image | Prepare image artifacts only, skip the `Docker` build step |
| `make test` | Test | Run the full `E2E` test suite |
| `make test-scripts` | Test | Run unit and smoke tests for tooling scripts |
| `make check-runtime-i18n` | i18n | Scan for hardcoded strings in runtime code |
| `make check-runtime-i18n-messages` | i18n | Validate message key coverage in runtime language packs |
| `make init` | Database | Initialize database schema and seed data |
| `make mock` | Database | Load demo `Mock` data |
| `make help` | Other | List all available commands |

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
    - "linux/amd64"
  # Whether to enable CGO
  cgoEnabled: false
  # Build output path, relative to the repository root
  outputDir: "temp/output"
  # Filename of the compiled host binary
  binaryName: "lina"
```

| Field | Default | Description |
|-------|---------|-------------|
| `build.platforms` | `["linux/amd64"]` | Target platform list in `goos/goarch` format; override with `make build platforms=...` |
| `build.cgoEnabled` | `false` | Whether to enable `CGO` |
| `build.outputDir` | `temp/output` | Build output path, relative to the repository root |
| `build.binaryName` | `lina` | Host binary filename |

### make wasm

Builds the runtime `WASM` plugins separately, with output written to `temp/output/`. Use `p=<plugin-id>` to build a single plugin; omit it to build all plugins.

```bash
# Build all WASM plugins
make wasm

# Build a specific plugin (plugin-id is the plugin directory name)
make wasm p=my-plugin

# Enable verbose logging
make wasm verbose=1
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

### make test-scripts

Runs unit and smoke tests for all tooling scripts under `hack/tests/scripts/`, verifying the basic correctness of repository helper scripts.

```bash
make test-scripts
```

## i18n

### make check-runtime-i18n

Scans runtime-visible code paths for hardcoded strings not covered by the i18n system. Run this before committing new features to catch compliance issues early.

```bash
make check-runtime-i18n
```

### make check-runtime-i18n-messages

Validates message key coverage across the host and plugin runtime language packs, detecting missing or extraneous translation keys.

```bash
make check-runtime-i18n-messages
```

## Database

:::caution Destructive operations

Both `init` and `mock` make destructive changes to the database. To prevent accidental execution, they require an explicit `confirm` argument.

:::

### make init

Initializes the database schema (`DDL`) and system seed data. The backend automatically selects the `MySQL` or `SQLite` dialect based on the `database.default.link` setting in `config.yaml`.

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
