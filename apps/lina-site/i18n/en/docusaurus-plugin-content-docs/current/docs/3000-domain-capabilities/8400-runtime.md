---
slug: '/docs/domain-capability-runtime'
title: 'Runtime'
hide_title: true
description: '`Runtime()` is a runtime host service client exclusive to dynamic plugins, corresponding to the `service: runtime` declaration in `plugin.yaml`. It provides WASM guests with structured log writing, plugin-scoped state read/write, host time reading, UUID generation, and node identity reading. This capability does not belong to the `capability.Services` standard domain directory; source plugins should use native host logging, context, and injected domain services. Infrastructure component state reads use the separate `Infra()` and `service: infra` capability.'
keywords:
  - Runtime capability
  - dynamic plugin runtime
  - RuntimeHostService
  - HostServiceRuntime
  - host:runtime
  - "service: runtime"
  - pluginbridge.Default().Runtime
  - WASI host service
  - log.write
  - state.get
  - state.set
  - state.delete
  - info.now
  - info.uuid
  - info.node
  - plugin runtime state
  - structured logging
  - node identity
  - dynamic plugins
  - LinaPro
---

## Overview

`Runtime()` is a capability exclusive to dynamic plugins. Dynamic plugins run within the WASM guest boundary and cannot directly access host-process logging, time, node identity, or local state implementations. They therefore need to call host-provided runtime primitives through `service: runtime`.

Source plugins already run inside the host process and typically use native host logging, request context, and injected domain services directly, so they do not need the `Runtime()` wrapper.

**Capability Phase**: Runtime

**Supported Types**: Dynamic plugins

## Design Philosophy

### Dynamic Plugin Exclusive Entry Point

`Runtime()` lives in `pluginbridge.Services` and does not belong to the `capability.Services` standard domain capability directory. It serves the dynamic plugin bridge runtime environment rather than business domain data access.

```mermaid
graph TB
    Dynamic["Dynamic Plugin WASM guest"] --> Runtime["pluginbridge.Default().Runtime()"]
    Runtime --> HostService["hostServices.runtime"]
    HostService --> Log["Structured logging"]
    HostService --> State["Plugin-scoped state"]
    HostService --> Info["Time, UUID, node identity"]
    Source["Source Plugin"] --> Native["Native host logging and context"]
```

### Boundary with `Infra()`

`Runtime()` reads dynamic plugin runtime primitives; `Infra()` reads infrastructure component state. The two are not interchangeable:

| Capability | Boundary | Primary Responsibility |
|------------|----------|----------------------|
| `Runtime()` | Dynamic plugin exclusive | Logging, plugin state, time, UUID, and node identity |
| `Infra()` | Shared between source and dynamic plugins | Infrastructure component state views |

If a dynamic plugin needs to check whether an infrastructure component is available, it should declare `service: infra` rather than treating `runtime.info.*` as a component status capability.

## Core Capabilities

| Dynamic Method | Dynamic SDK Method | Description |
|----------------|-------------------|-------------|
| `log.write` | `Runtime().Log` | Writes a structured runtime log entry |
| `state.get` | `Runtime().StateGet`, `Runtime().StateGetInt` | Reads plugin-scoped runtime state |
| `state.set` | `Runtime().StateSet`, `Runtime().StateSetInt` | Writes plugin-scoped runtime state |
| `state.delete` | `Runtime().StateDelete` | Deletes plugin-scoped runtime state |
| `info.now` | `Runtime().Now` | Reads the host time string |
| `info.uuid` | `Runtime().UUID` | Generates a host-side unique identifier |
| `info.node` | `Runtime().Node` | Reads the current host node identity |

`runtime` is a `none` resource type and does not declare `paths`, `tables`, `keys`, or `resources`. Authorization boundaries are controlled by `service + method`.

## Usage

Dynamic plugins declare the `runtime` service in `plugin.yaml`:

```yaml
hostServices:
  - service: runtime
    methods:
      - log.write
      - state.get
      - state.set
      - state.delete
      - info.now
      - info.uuid
      - info.node
```

Call through `pluginbridge.Default().Runtime()` on the dynamic plugin side:

```go
runtime := pluginbridge.Default().Runtime()

err := runtime.Log(protocol.LogLevelInfo, "export started", map[string]string{
    "taskID": taskID,
})
if err != nil {
    return err
}

now, err := runtime.Now()
if err != nil {
    return err
}

err = runtime.StateSet("last_export_time", now)
if err != nil {
    return err
}

lastExport, found, err := runtime.StateGet("last_export_time")
if err != nil {
    return err
}

id, err := runtime.UUID()
if err != nil {
    return err
}

node, err := runtime.Node()
```

## Design Constraints

- **Dynamic plugin exclusive.** `Runtime()` only exists in the `pluginbridge` dynamic SDK and does not enter the source plugin `capability.Services` directory.
- **Runtime state is not a cache substitute.** `runtime.state.*` is suitable for storing small amounts of plugin-scoped state. Cross-plugin sharing, counters, and expiration policies should use the [Cache capability](/docs/domain-capability-cache).
- **Log fields should be short and stable.** Dynamic plugins should not put large bodies, secrets, tokens, or personally identifiable information into log fields.
- **Info reads are not health checks.** `info.now`, `info.uuid`, and `info.node` only provide basic runtime information and do not express component availability.
- **Infrastructure state goes through `Infra()`.** Component state reads should use the [Infra capability](/docs/domain-capability-infra) and `service: infra`.

## Related Services

- [Infra Capability](/docs/domain-capability-infra)
- [Cache Capability](/docs/domain-capability-cache)
- [Dynamic Plugins and WASM Runtime](/docs/wasm-plugins)
- [Domain Capability Overview](/docs/domain-capabilities)
