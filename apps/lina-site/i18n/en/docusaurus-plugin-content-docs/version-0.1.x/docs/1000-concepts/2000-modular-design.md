---
slug: '/docs/modular-design'
title: 'Modular Design'
hide_title: true
description: 'An exploration of LinaPro modular design — how domain-bounded built-in capabilities, first-class plugin system, stable extension seams, and lifecycle-managed modules let developers assemble systems from proven building blocks rather than building everything from scratch.'
keywords:
  - modular design
  - building blocks
  - decoupled design
  - plugin ecosystem
  - module lifecycle
  - composable modules
  - pluggable
  - LinaPro
  - AI-native
  - modular architecture
  - extension seams
  - sustainable delivery
---

## What Modular Design Means

Modular design is one of `LinaPro`'s core design principles. The central idea is: **every capability in the framework exists as a decoupled module, and modules interact through stable interfaces rather than depending directly on each other's internal implementation**.

Each module is a complete, self-contained functional unit that can be installed, enabled, disabled, or uninstalled on demand — with zero intrusive effect on any other part of the system. Module boundaries remain clear at all times; responsibilities do not overlap, and modules do not interfere with each other.

## How the Framework Implements Modularity

`LinaPro` treats modularity as a core principle from day one, not as a bolt-on extension added after the framework took shape.

- **Built-in capabilities are split by domain boundary**: Each capability domain of the core framework (`lina-core`) — authentication, authorization, users, menus, dictionaries, parameters, files, scheduling, plugin governance — is designed independently and evolves independently, with no cross-domain leakage. Each domain has a clear entry point and contract; external consumers can only access it through its public interface and cannot bypass the boundary to reach its internal implementation.

- **The plugin system is a first-class citizen**: Plugins are not an afterthought — they are the core of the framework's extension mechanism. From the very beginning of the project, the plugin runtime was designed alongside the core framework service with full lifecycle management and governance. Every official business capability is delivered under the same module specification as any third-party plugin.

- **Frontend and backend are modularized together**: Each plugin ships its own backend service, frontend pages, database resources, and permission declarations as a single, coherent module unit — not as two separate halves. Installing a plugin means installing a fully self-contained capability.

- **Stable extension seams**: The core framework exposes well-defined extension points to plugins. Plugins can only interact with the core framework through these seams — registering routes, responding to events, filtering menus, and so on — and can never reach into the core framework's internal implementation. This boundary design keeps inter-module dependencies clear and controlled, so the core framework can evolve internally without accidentally breaking existing plugins.

## Building Blocks vs. Building From Scratch

Even in the AI era, the cost of developing a complete feature from scratch remains meaningful: requirements analysis consumes tokens, database design consumes tokens, frontend and backend implementation consume tokens, and testing and debugging consume more. After all that investment, the new module may not be mature or stable — it typically goes through multiple rounds of real-world use before it becomes reliable.

`LinaPro`'s modular design offers a different path. **The modules and plugins the framework provides are themselves proven building blocks — the product of extensive development, testing, and validation over time.** With AI, developers need only a small number of tokens to do the "assembly" work: describe the requirement, select the right blocks, configure the combination. The result is a stable, reliable system without reinventing the wheel every time.

Mature building blocks are naturally more stable and predictable than freshly built wheels.

## Two Kinds of Building Blocks

`LinaPro` provides two categories of building blocks:

- **Core framework built-in modules**: Capabilities available out of the box from the core service — user and permission systems, dictionaries and parameter management, file storage, job scheduling, plugin governance, and more. These modules ship with the framework and require no additional installation; they form the stable foundation of every project.

- **Official plugin ecosystem**: Extension capabilities delivered as standalone plugins — covering common business domains (organization management, content management, monitoring and auditing, etc.) and vertical-industry use cases. The official plugin ecosystem will continue to grow, covering more and more common business scenarios. In the future, developers will be able to combine official plugins to quickly assemble a fully functional business system at minimal cost.

## Module Composability

The value of modular design lies not only in "which blocks are available" but also in "how flexibly you can combine them."

`LinaPro` modules support full lifecycle management:

| Operation | Description |
|-----------|-------------|
| Install | Register the module's menus, routes, permissions, and database resources |
| Enable | Activate the module so it is visible and usable |
| Disable | Hide the module's capabilities; existing data is preserved in full |
| Uninstall | Remove the module; optionally clean up its data |

Each operation is atomic and non-intrusive to the rest of the system. Disabling simply makes capabilities invisible — it does not destroy any existing data, and re-enabling restores everything. This design lets the system adjust its capability boundaries without downtime or code changes, dramatically reducing change risk. It is also the practical expression of the "sustainable delivery" principle.

## Modularity Amplified by AI-Native Development

Under an AI-native development workflow, the value of modular design is magnified further.

In traditional development, AI must start from scratch to understand business requirements, design data structures, and write complete implementations — every step consuming compute, with inconsistent output quality. In `LinaPro`'s modular system, AI's role fundamentally shifts: **from "making bricks" to "building with bricks"**.

AI no longer needs to reinvent already-validated wheels. Instead, it works on top of proven building blocks, focusing on understanding the business composition logic, selecting the right modules, and completing targeted configuration and extension. This division of labor lets AI concentrate where it creates real value — understanding the business, making decisions, and assembling the building blocks into a system that meets the requirement.
