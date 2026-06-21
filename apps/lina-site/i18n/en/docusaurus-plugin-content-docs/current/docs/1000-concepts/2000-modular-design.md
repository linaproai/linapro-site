---
slug: '/docs/modular-design'
title: 'Modular Design'
hide_title: true
description: 'How the framework follows the "framework minimalism principle" — only placing core generic capabilities in the core, delegating what plugins can provide to plugins — through domain decoupling of built-in capabilities, plugins as independent module units, a stable extension interface system, and real-world examples like cloud storage and LDAP/OIDC, enabling developers to build systems like assembling building blocks, significantly reducing development costs and improving delivery quality and system maintainability.'
keywords:
  - modular design
  - framework minimalism principle
  - building blocks
  - decoupled design
  - plugin ecosystem
  - plugin extension
  - storage backend plugin
  - authentication plugin
  - interface abstraction
  - module lifecycle
  - on-demand assembly
  - pluggable
  - LinaPro
  - AI-native
  - modular architecture
  - extension interface
  - sustainable delivery
---

## What Is Modular Design

Modular design is one of `LinaPro`'s core design philosophies. Its central idea is: **all framework capabilities exist as decoupled modules, and modules collaborate through stable interfaces rather than depending on each other's internal implementations**.

Each module is a complete, independent functional unit that can be installed, enabled, disabled, or uninstalled on demand — the entire process is non-intrusive to the rest of the system. Boundaries between modules are always clear, responsibilities don't overlap, and modules don't interfere with each other.

## The Framework's Modular Design

`LinaPro` established modular design as a core principle from its inception, rather than bolting it on as an afterthought once the framework was already formed.

- **Built-in capabilities split by domain boundaries**: Each capability domain of the main framework (`lina-core`) — authentication, authorization, users, menus, dictionaries, parameters, file storage, scheduling, plugin governance — is independently designed and evolves independently, with no cross-contamination. Each capability domain has a clear entry point and contract; external code can only access it through public interfaces and cannot bypass the boundary to call internal implementations directly.

- **The plugin system is a first-class citizen**: Plugins are not an add-on to the framework — they are the core of the framework's extension mechanism. From the very beginning of the framework's design, the plugin runtime was designed alongside the main framework services, with complete lifecycle management and governance. Any official business feature follows the exact same module specification as third-party plugins.

- **Frontend and backend modularized together**: Each plugin owns its own backend services, frontend pages, database resources, and permission declarations. Frontend and backend are delivered as a complete module unit, not two separate halves. This means installing a plugin installs a fully self-contained functional module with a complete capability loop.

- **Stable extension interfaces**: The main framework exposes well-defined extension points to plugins. Plugins can only interact with the main framework through these extension interfaces — registering routes, responding to events, filtering menus, etc. — and never directly touch the main framework's internal implementation. This boundary design keeps dependency relationships between modules always clear and controllable, and the main framework's internal evolution won't accidentally break existing plugins.

## The Framework Minimalism Principle

The `LinaPro` main framework follows a simple design principle: **only include core, generic capabilities; delegate what plugins can provide to plugins**.

The main framework is not a catch-all for features — it is a stable capability foundation. As business requirements grow, the path to expansion is enriching the plugin ecosystem and adding new plugins, not piling more implementations into the main framework. This keeps the main framework lightweight, core logic controllable, and external evolution stable.

**Cloud storage support as an example**: When integrating object storage like Qiniu Cloud or `AWS S3`, the specific integration logic is not written directly into the main framework. Instead, the corresponding storage plugin provides the implementation. Once the plugin is enabled, file writes automatically route to the cloud. The main framework does only one thing in this process — **define the storage interface abstraction**. This abstraction layer allows seamless switching between different storage backends, and callers don't need to know about differences in underlying implementations.

**Authentication works the same way**: For protocols like `LDAP` and `OIDC`, the main framework provides only a thin authentication interface abstraction. The specific protocol handshake and user mapping logic is handled by plugins. This division means the main framework doesn't need to bloat for every third-party integration, and developers can flexibly extend authentication methods on demand without depending on the main framework's release cycle.

This principle is the foundation of `LinaPro`'s plugin ecosystem: **the framework provides abstracted, interface-based capabilities; plugins provide composable, concrete implementations**. Both respect their boundaries and work together, keeping the system structurally clear in the face of diverse requirements.

## Building Blocks vs. Building from Scratch

Even in the AI era, developing a complete feature from scratch is expensive: requirements analysis consumes tokens, database design consumes tokens, frontend and backend implementation consumes tokens, testing and debugging consume more tokens. After significant time, effort, and compute, the resulting feature module may still not be mature or stable — it often needs multiple rounds of real-world refinement before it becomes reliable.

`LinaPro`'s modular design offers an alternative. **The modules and plugins provided by the framework are mature building blocks that have already undergone extensive development, testing, and validation over significant time, effort, and tokens**. Developers use `AI` with minimal tokens to "assemble" them — describing requirements, selecting blocks, configuring combinations — to quickly build stable, reliable systems, rather than reinventing the wheel every time.

Mature building blocks are inherently more stable and predictable than freshly reinvented wheels.

## Two Types of Building Blocks

`LinaPro` provides two types of building blocks for developers to choose from:

- **Main framework built-in modules**: Capabilities available out of the box in the framework's core services, including the user and permission system, dictionary and parameter management, file storage, task scheduling, plugin governance, and more. These modules are delivered with the framework and require no additional installation — they are the stable foundation for every project.

- **Official plugin ecosystem**: Extension capabilities delivered as independent plugins, including general business plugins (such as organizational structure, content management, monitoring and auditing) and vertical industry plugins. The official plugin ecosystem will continue to grow, covering an increasing number of common business scenarios. In the future, developers will be able to combine official plugins to rapidly build fully functional business systems at minimal cost.

## Module Composability

The value of modular design lies not just in "what building blocks are available" but in "how flexibly you can use them."

`LinaPro`'s modules have complete lifecycle management capabilities:

| Operation | Description |
|-----------|-------------|
| **Install** | Registers the module's menus, routes, permissions, and database resources |
| **Enable** | Activates the module, making it visible and usable to users |
| **Disable** | Hides the module's capabilities; existing data is fully preserved |
| **Uninstall** | Removes the module; optionally cleans up associated data |

Each operation is atomic and non-intrusive to the rest of the system. Disabling only makes capabilities invisible — it doesn't destroy any existing data. Re-enabling restores everything to its previous state. This design lets the system flexibly adjust capability boundaries without downtime or code changes, significantly reducing change risk. It is also a concrete embodiment of the "sustainable delivery" philosophy.

## Combining Modular Design with AI-Native

Under AI-native development workflows, the value of modular design is amplified further.

In traditional development, `AI` needs to understand business requirements from scratch, design data structures, and write complete implementations — every step consumes compute, and output quality varies. In `LinaPro`'s modular system, `AI`'s role fundamentally shifts: **from "making bricks" to "building houses."**

`AI` no longer needs to reinvent wheels that have already been validated. Instead, it builds on mature building blocks, focusing on understanding business combination logic, selecting appropriate modules, and completing targeted configuration and extension. This division lets `AI`'s energy concentrate where it truly creates value — understanding the business, making decisions, and combining building blocks into systems that meet requirements.
