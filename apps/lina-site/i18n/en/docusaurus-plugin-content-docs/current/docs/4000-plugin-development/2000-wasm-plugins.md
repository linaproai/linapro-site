---
slug: '/docs/wasm-plugins'
title: 'Dynamic Plugins (WASM)'
hide_title: true
description: 'A complete guide to developing LinaPro WASM dynamic plugins, covering directory structure, cross-platform linactl builds, host service bridge interfaces (runtime, cron, storage, network, data), hostServices permission declarations in the plugin manifest, multi-tenant manifest fields, runtime upload and installation workflow, and key differences from source plugins.'
keywords:
  - WASM dynamic plugins
  - dynamic plugins
  - WebAssembly
  - hot-loading
  - plugin bridge
  - pluginbridge
  - host services
  - hostServices
  - storage service
  - network service
  - data service
  - cron service
  - runtime service
  - plugin upload
  - linactl wasm
  - LinaPro plugin
  - dynamic injection
  - plugin sandbox
---

## Overview

`WASM` dynamic plugins are `LinaPro`'s unique runtime extension capability. Plugins are compiled into `WebAssembly` format and can be dynamically uploaded, installed, enabled, disabled, and uninstalled while the host is running — all without downtime or restarts.

Dynamic plugins run in a full `WASM` sandbox. All access to host resources (filesystem, database, network) must go through governed bridge interfaces provided by the host.

## Directory Structure

Dynamic plugins share the same directory structure as source plugins. The only difference is that a dynamic plugin's `main.go` implements the `WASM` plugin protocol:

```text
apps/lina-plugins/<plugin-id>/
├── main.go                     # Plugin WASM entry (func main)
├── plugin_embed.go             # Embedded resource registration
├── plugin.yaml                 # Plugin manifest (includes host service permission declarations)
├── backend/
│   └── internal/
│       └── service/            # Business logic (accesses host capabilities via bridge interfaces)
├── frontend/
│   └── pages/                  # Plugin frontend pages (independent static assets)
└── manifest/
    └── sql/                    # Installation SQL (optional; dynamic plugins typically don't create tables)
        └── uninstall/          # Uninstallation SQL (optional)
```

## Plugin Manifest (Host Service Declarations)

Dynamic plugins must declare their multi-tenant boundaries, dependencies, and required host services in `plugin.yaml`. The host validates these permissions during installation and enablement:

```yaml
# Unique plugin identifier (kebab-case), globally unique
id: my-dynamic-plugin
# Plugin display name
name: Dynamic Plugin Example
# Semantic version (semver)
version: v0.1.0
# Plugin type: dynamic means WASM dynamic plugin
type: dynamic
# Multi-tenant scope and default install mode
scope_nature: tenant_aware
supports_multi_tenant: true
default_install_mode: tenant_scoped
# Brief description
description: A demo plugin showcasing dynamic plugin capabilities

# Declare required host service permissions; administrators confirm these at install time
hostServices:
  - service: runtime
    methods:
      - log.write
      - info.now
      - info.node
  - service: storage
    methods:
      - put
      - get
      - delete
    resources:
      paths:
        - demo-record-files/
  - service: network
    methods:
      - request
    resources:
      - url: https://example.com
  - service: data
    methods:
      - list
      - get
      - create
      - update
      - delete
    resources:
      tables:
        - plugin_demo_dynamic_record

# Menu declarations
menus:
  - key: plugin:my-dynamic-plugin:main   # Unique menu key, format: plugin:<plugin-id>:<feature>
    name: Dynamic Plugin Example          # Display name
    path: my-dynamic-plugin-main          # Frontend route path, globally unique
    component: system/plugin/dynamic-page # Fixed component for dynamic plugin pages
    type: M                               # Menu type: D=directory, M=menu item, B=button
    sort: 1                               # Sort weight, lower values appear first
```

Host service permission reference:

| Service | Description | Typical Use |
|---------|-------------|-------------|
| `runtime` | Access runtime metadata, write to host logs, maintain plugin state | Display framework version, node info, record runtime logs |
| `cron` | Register built-in scheduled tasks for the dynamic plugin | Periodic sync, cleanup, stats tasks |
| `storage` | Read and write files within the plugin's namespace | Upload attachments, generate report files |
| `network` | Make governed outbound `HTTP` requests | Call third-party `API`s, `Webhook`s |
| `data` | Access database data within authorized table scope | Read and write plugin-owned data |

## Building Dynamic Plugins

### Prerequisites

Building `WASM` plugins uses the standard `Go` toolchain (`Go 1.22+`) with the `GOOS=wasip1 GOARCH=wasm` compilation target. No additional tools like `TinyGo` are required — the host's `hack/tools/build-wasm` build tool encapsulates all compilation parameters.

### Build Commands

```bash
# Cross-platform entry: run from the hack/tools/linactl directory
cd hack/tools/linactl
go run . wasm
go run . wasm p=my-dynamic-plugin

# Linux/macOS compatible entry
make wasm
make wasm p=my-dynamic-plugin

# Windows compatible entry
make.cmd wasm
make.cmd wasm p=my-dynamic-plugin
```

Build artifacts are output to `temp/output/my-dynamic-plugin.wasm`.

### Build Tool Code (hack/tools/build-wasm)

The host provides a unified `WASM` build tool that handles compilation parameters, resource embedding, and artifact packaging automatically — no manual `TinyGo` configuration needed.

## Developing a Dynamic Plugin

### Plugin Entry Point

A `WASM` plugin's entry is `main.go`, implementing the plugin protocol defined by `LinaPro`:

```go
// main.go
package main

import (
    bridge "github.com/linaproai/linapro/apps/lina-core/pkg/pluginbridge"
)

func main() {
    // Register plugin route handlers
    bridge.RegisterRoutes(func(router bridge.Router) {
        router.GET("/hello", handleHello)
    })
}

func handleHello(ctx bridge.Context) {
    // Access host capabilities via bridge interfaces
    info, _ := ctx.Runtime().GetFrameworkInfo()
    ctx.JSON(200, map[string]string{
        "message": "Hello from WASM plugin!",
        "version": info.Version,
    })
}
```

### Accessing Host Services

Dynamic plugins access host capabilities through bridge interfaces, all constrained by the sandbox:

**File storage (storage):**

```go
// Write a file to the plugin namespace
err := ctx.Storage().Write("reports/2024-01.csv", csvData)

// Read a file from the plugin namespace
data, err := ctx.Storage().Read("reports/2024-01.csv")

// List files in a plugin directory
files, err := ctx.Storage().List("reports/")
```

**Database access (data):**

```go
// Query data within the plugin namespace
// Plugins can only access tables prefixed with the plugin ID
rows, err := ctx.Data().Query("SELECT * FROM my_dynamic_plugin_records WHERE status = ?", 1)

// Execute a write operation
affected, err := ctx.Data().Exec("INSERT INTO my_dynamic_plugin_records (title) VALUES (?)", "Demo")
```

**HTTP network requests (network):**

```go
// Make an external HTTP request (constrained by domain allowlist)
resp, err := ctx.Network().GET("https://api.example.com/data")
body, _ := io.ReadAll(resp.Body)
```

**Runtime information (runtime):**

```go
// Get framework runtime information
info, err := ctx.Runtime().GetFrameworkInfo()
// info.Version, info.NodeID, info.StartedAt
```

## Frontend Pages

Dynamic plugin frontend pages are **independent static files** that do not depend on the host's `Vue` framework. They are displayed in the management workspace via `iframe` or direct embedding:

```html
<!-- frontend/pages/main.html -->
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8" />
    <title>Dynamic Plugin Example</title>
    <!-- Any frontend framework or vanilla HTML can be used -->
</head>
<body>
    <div id="app">
        <h1>Dynamic Plugin Example</h1>
        <button onclick="fetchData()">Load Data</button>
        <div id="result"></div>
    </div>
    <script>
    async function fetchData() {
        const res = await fetch('/plugin/my-dynamic-plugin/hello')
        const data = await res.json()
        document.getElementById('result').textContent = JSON.stringify(data)
    }
    </script>
</body>
</html>
```

## Installation and Usage

The complete workflow for dynamic plugins:

1. Build the `.wasm` file: `go run . wasm p=<plugin-id>` — artifacts are output to `temp/output/<plugin-id>.wasm`
2. Log in to the management workspace and navigate to **Extension Center → Plugin Management**
3. Click **Upload Plugin** and select the built `.wasm` file
4. Confirm the permission declarations (i.e., the `hostServices` requested in `plugin.yaml`) — the administrator reviews and proceeds
5. Click **Install** — the host runs installation `SQL` (if any) and writes governance records
6. Click **Enable** — the plugin functionality takes effect immediately, and the plugin entry appears in the left sidebar

## Version Upgrades

Dynamic plugins support independent version upgrades without redeploying the host:

1. Build a new version `.wasm` file (update the `version` field in `plugin.yaml`)
2. Upload the new version file on the Plugin Management page
3. The host automatically detects the version change and offers an upgrade option
4. After confirming the upgrade, the new version takes effect immediately

## Key Differences from Source Plugins

| Feature | Source Plugin | `WASM` Dynamic Plugin |
|---------|--------------|----------------------|
| Load at startup | ✅ Enabled plugins load automatically at startup | ✅ Installed and enabled plugins load at startup |
| Hot-loading | ❌ | ✅ Available immediately after upload |
| Frontend framework | Shares host `Vue3` ecosystem | Independent static pages, any tech stack |
| Database access | Direct `GoFrame` ORM usage | Via bridge interfaces, scoped to namespace |
| Debugging tools | Standard `Go` debugging tools | Limited debugging capabilities |

Reference implementation: `apps/lina-plugins/plugin-demo-dynamic/`
