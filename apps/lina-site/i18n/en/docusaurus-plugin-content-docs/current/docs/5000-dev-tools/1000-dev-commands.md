---
slug: '/docs/dev-commands'
title: 'Dev Commands'
description: 'A complete reference for every make target in the LinaPro project — what each command does, which parameters it accepts, and usage examples covering dev server management, full builds, WASM plugin builds, Docker image builds, testing, i18n checks, and database initialization.'
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

The `LinaPro` project root provides a unified set of `make` targets, organized across module files under `hack/makefiles/`. Run `make help` at any time to list all available targets.

## Command Overview

| Command | Category | Description |
|---------|----------|-------------|
| `make dev` | Dev server | Restart the backend and frontend dev servers |
| `make stop` | Dev server | Stop the backend and frontend dev servers |
| `make status` | Dev server | Show runtime status and log file paths |
| `make build` | Build | Full build: frontend, plugins, and backend binary |
| `make wasm` | Build | Build all or a specific runtime `WASM` plugin |
| `make image` | Image | Build the production `Docker` image |
| `make image-build` | Image | Stage image artifacts without running `Docker` build |
| `make test` | Test | Run the full `E2E` test suite |
| `make test-scripts` | Test | Run unit and smoke tests for tooling scripts |
| `make check-runtime-i18n` | i18n | Scan for hard-coded strings in runtime code paths |
| `make check-runtime-i18n-messages` | i18n | Validate runtime i18n message key coverage |
| `make init` | Database | Initialize database schema and seed data |
| `make mock` | Database | Load demo mock data |
| `make help` | Other | List all available targets |

## Dev Server

### make dev

Restarts the backend and frontend development servers. It stops any running services first, then sequentially builds `WASM` plugins, prepares frontend static assets, and compiles the backend binary. The command waits for both health checks to pass before printing the status panel.

```bash
make dev
```

The backend listens on `http://localhost:8080` by default; the frontend listens on `http://localhost:5666`. Logs are written to `temp/lina-core.log` and `temp/lina-vben.log`.

### make stop

Stops the backend and frontend development servers and removes stale `PID` files. Any processes still occupying the configured ports are force-killed.

```bash
make stop
```

### make status

Prints the current runtime status and log paths for both services — handy for quickly checking whether everything started correctly.

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

Runs the full build pipeline in order: frontend asset compilation, embedding manifest resources into the backend, all `WASM` plugin builds, and finally compiling the backend host binary. Artifacts land in `temp/output/`.

```bash
# Default build (current platform)
make build

# Cross-compile for specific platforms
make build platforms=linux/amd64,linux/arm64

# Show verbose output
make build verbose=1
# or
make build v=1
```

Build parameters like target platforms, binary name, output directory, and `CGO` settings are driven by `hack/config.yaml` but can be overridden on the command line with `platforms`, `output_dir`, `binary_name`, and `cgo_enabled`.

### make wasm

Builds the runtime `WASM` plugins only, outputting to `temp/output/`. Pass `p=<plugin-id>` to build a single plugin; omitting `p` builds all of them.

```bash
# Build all WASM plugins
make wasm

# Build a specific plugin (plugin-id is the plugin directory name)
make wasm p=my-plugin

# Show verbose output
make wasm verbose=1
```

## Image

### make image

Runs the full `Docker` image pipeline: calls `make build` internally to produce all artifacts, then invokes `hack/tools/image-builder` to package them into an image. Image name, tag, and registry are all configurable.

```bash
# Build using default config
make image

# Specify a tag and registry
make image tag=v0.6.0 registry=ghcr.io/linaproai

# Build and push in one step
make image tag=v0.6.0 registry=ghcr.io/linaproai push=1

# Multi-platform build
make image platforms=linux/amd64,linux/arm64 tag=v0.6.0
```

### make image-build

Stages all image build artifacts (equivalent to running `make build`) without executing `Docker build`. Useful when you want to inspect the artifacts first or plug in a custom image build step.

```bash
make image-build
```

## Test

### make test

Runs the full `Playwright E2E` test suite under `hack/tests/`. Make sure the dev servers are up (`make dev`) before running.

```bash
make test
```

### make test-scripts

Runs all unit and smoke tests for repository tooling scripts under `hack/tests/scripts/`, verifying that helper scripts behave correctly.

```bash
make test-scripts
```

## i18n

### make check-runtime-i18n

Scans runtime-visible code paths for hard-coded strings that have not been routed through the i18n system. Run this before committing new features as a self-audit for i18n compliance.

```bash
make check-runtime-i18n
```

### make check-runtime-i18n-messages

Validates the message key coverage of the host's and plugins' runtime language packs, surfacing missing or extraneous translation keys.

```bash
make check-runtime-i18n-messages
```

## Database

:::caution Destructive operations

Both `init` and `mock` make destructive changes to the database. They require an explicit `confirm` parameter to prevent accidental execution.

:::

### make init

Initializes the database schema (`DDL`) and required seed data. The backend automatically dispatches to the correct dialect (`MySQL` or `SQLite`) based on `database.default.link` in `config.yaml`.

```bash
# Initialize only (preserve existing data)
make init confirm=init

# Rebuild the database from scratch
make init confirm=init rebuild=true
```

### make mock

Loads optional demo mock data for local development and verification. Requires `make init` to have run first.

```bash
make mock confirm=mock
```

## Other

### make help

Prints all available targets from the root `Makefile` and every included sub-makefile, sorted by name.

```bash
make help
```
