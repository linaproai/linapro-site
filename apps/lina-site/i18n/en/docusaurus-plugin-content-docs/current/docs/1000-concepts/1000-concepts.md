---
slug: '/docs/concepts'
title: 'Concepts'
hide_title: true
description: 'AI-native design, modular architecture, AI spec-driven development, AI engineering quality assurance, and framework-plugin capability boundaries, helping developers build the right mental model before going deeper into the framework.'
keywords:
  - design principles
  - design philosophy
  - AI-native
  - AI-native design
  - modular design
  - modular architecture
  - spec-driven development
  - AI spec-driven development
  - AI engineering quality
  - engineering quality assurance
  - SDD
  - OpenSpec
  - test coverage
  - API contract
  - framework architecture
  - framework design
  - LinaPro
---

## Design Principles

A framework's behavior is shaped by the choices made when it was designed. Understanding the reasoning behind those choices helps developers grasp the framework's boundaries more accurately, work with it more naturally, and get more out of it.

`LinaPro`'s design principles revolve around five core themes:

- **[AI-Native Design](/docs/ai-native)**: Treating `AI` as the primary engine of engineering productivity — not a supplementary tool. The AI-native design manifests in two independent dimensions: a spec-driven development workflow that lets `AI` participate deeply at every stage from requirements analysis to implementation and testing; and a built-in `AI` skill system covering the full development lifecycle, enabling `AI` to make framework-aware decisions in every specific work context — from backend development and frontend design to test coverage, performance audits, and version upgrades. Together, these two dimensions form `LinaPro`'s core productivity engine.

- **[Modular Design](/docs/modular-design)**: Every capability in the framework exists as a decoupled module that interacts with others through stable interfaces. Developers assemble systems like connecting building blocks rather than building from scratch each time — a fundamentally better way to deliver reliably at speed.

- **[AI Spec-Driven Development](/docs/spec-driven-development)**: Built on the principle that specs come before code, the spec-driven workflow captures every iteration's design decisions and implementation context as persistent documents, ensuring code, documentation, and tests are produced in the same iteration cycle — preventing architectural drift at the root. `OpenSpec` is the recommended tool for implementing this workflow.

- **[AI Engineering Quality Assurance](/docs/ai-engineering-quality)**: A systematic look at the engineering management challenges that emerge when `AI` becomes part of software development, and how `LinaPro` builds a complete quality assurance system across four dimensions: the `SDD` spec-driven workflow, full project-level specifications, interface abstraction with anti-leakage contracts, and high-density test coverage, where test code accounts for `39%` of the total codebase.

- **[Framework and Plugin Capability Boundaries](/docs/capability-boundary)**: Describes the division of responsibilities between the core framework and plugins — the core framework focuses on lightweight foundational capabilities and stable extension interfaces, while business capabilities are delivered through plugin extensions. Details how source plugins integrate through the `pluginhost` contract, dynamic plugins communicate through the `pluginbridge` sandbox, and the conventions governing route namespaces, public asset hosting, and admin workspace integration.


## Related Documents

import DocCardList from '@theme/DocCardList';

<DocCardList />
