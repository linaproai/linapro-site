---
slug: '/docs/plugin-development'
title: 'Plugin Development'
hide_title: true
description: 'An overview of LinaPro dual-mode plugin system design — when to choose source plugins vs. WASM dynamic plugins, a detailed feature comparison, selection guidance, and the unique value of dynamic plugins for hot-loading, production hotfixes, and commercial distribution.'
keywords:
  - plugin development
  - plugin extension
  - source plugins
  - WASM dynamic plugins
  - plugin selection
  - hot-loading
  - hot-swap
  - commercial plugins
  - LinaPro
  - plugin system
  - dual-mode plugins
  - plugin comparison
  - non-intrusive plugins
  - dynamic injection
  - plugin distribution
  - plugin extensions
---

## Why Two Plugin Modes Exist

`LinaPro`'s plugin system supports two fundamentally different delivery modes because real-world extension needs fall into two genuinely distinct categories.

### Category 1: Long-lived, stable business features

Organization management, content management, order processing — these features require sustained iteration and maintenance, close collaboration with the system core, and have some performance requirements. Source plugins are compiled together with the host, call host-provided `Go` packages directly, and deliver the best performance and the simplest toolchain.

### Category 2: Dynamic, temporary, or commercially distributed extensions

- Let AI develop and inject a new feature into a production environment temporarily, decide whether to keep it after validation
- Discover a production bug and inject a fix without taking the service down
- As a commercial ISV, distribute plugin capabilities in binary form without exposing source code
- Let third-party developers build and distribute plugins that end users install by uploading

Source plugins cannot meet these needs — they require recompiling the entire host to take effect. `WASM` dynamic plugins can be uploaded and activated at runtime immediately.

## Detailed Comparison

| Dimension | Source plugin | `WASM` dynamic plugin |
|-----------|---------------|----------------------|
| **Hot-loading** | ❌ Requires host restart | ✅ Hot-swap at runtime, no restart |
| **Performance** | ✅ Native `Go` performance | ⚠️ ~10–30% overhead from sandbox calls |
| **Development complexity** | ✅ Low — standard Go development | ⚠️ Requires understanding WASM builds and restricted interfaces |
| **Host service access** | ✅ Direct host package calls | ⚠️ Accessed through bridge interfaces with restrictions |
| **Debug experience** | ✅ Standard Go debugging tools | ⚠️ More complex to debug |
| **Source code protection** | ❌ Source managed with the host | ✅ Distribute `.wasm` binary only |
| **Independent upgrades** | ❌ Requires redeploying the host | ✅ Upload a new `.wasm` version independently |
| **Cross-platform** | ❌ Requires recompilation | ✅ Compile once, run on any platform |
| **Isolation level** | Namespace isolation | Full `WASM` sandbox isolation |
| **Recommended use case** | Long-term business features | Temporary features, hotfixes, commercial distribution |

## Choosing Between Them

**Default to source plugins.** They are the right choice for most business development — better developer experience, higher performance, easier debugging, and a perfect fit with `LinaPro`'s AI-native workflow and built-in AI skills (`goframe-v2`, `lina-e2e`, `lina-review`, etc.). AI can make framework-compliant decisions at every stage of plugin development.

**Choose dynamic plugins when any of the following applies:**

- Hot-loading is required — you cannot restart the host
- Emergency production fix with minimal downtime risk
- You do not want to distribute source code (commercial plugin)
- End users should be able to upload and manage plugins themselves
- Let AI rapidly prototype and validate a feature that may be discarded afterward

## The Core Value of Dynamic Plugins

Dynamic plugin hot-loading is especially valuable in the following scenarios:

### Temporary development and dynamic injection

You can ask `Claude Code` (or any other AI tool) to develop a `WASM` plugin, build it, and upload it directly to a running system — no downtime. Once you are satisfied with the result, decide whether to convert it into a source plugin for long-term maintenance.

```
# Example: ask AI to quickly build a temporary stats dashboard
Ask Claude Code: Build a WASM plugin that shows today's new user count on a stats panel
```

### Production hotfixes

When an urgent issue arises in production, injecting a fix via a dynamic plugin reduces the downtime window from "redeploy the entire host" (potentially minutes) to "upload a `.wasm` file" (seconds).

### Commercial plugin distribution

Plugin vendors can compile their plugins to `.wasm` binaries for distribution. End users upload and install them directly — no source code exposure, no host recompilation required.

## Related Documents

import DocCardList from '@theme/DocCardList';

<DocCardList />
