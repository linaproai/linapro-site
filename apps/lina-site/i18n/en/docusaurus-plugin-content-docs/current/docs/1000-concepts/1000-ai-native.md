---
slug: '/docs/ai-native'
title: 'AI-Native Design'
hide_title: true
description: 'A detailed look at LinaPro as an AI-native framework — what AI-native truly means, how it differs from AI-assisted development, how AI participates in every stage of the development lifecycle, the built-in AI skill system covering backend, frontend, testing, performance auditing, and version management, and why spec-driven development prevents architectural drift over time.'
keywords:
  - AI-native
  - AI-native framework
  - spec-driven development
  - LinaPro
  - OpenSpec
  - AI workflow
  - sustainable delivery
  - AI as primary path
  - specification-driven
  - code consistency
  - architectural drift
  - test gaps
  - AI-led development
  - human decision-making
  - SDD
  - full-lifecycle AI
  - AI skill system
  - goframe-v2
  - lina-review
  - lina-e2e
  - lina-perf-audit
  - lina-upgrade
  - frontend-design
  - performance audit
  - skill ecosystem
---

## What "AI-Native" Means

"AI-native" does not mean adding an AI assistant button to your product, nor does it mean using AI to help with one specific feature. AI-native means **making AI the primary path of the product, not an optional add-on**. In an AI-native design, AI is not a tool — AI is the team's core engineering productivity. The developer's role shifts from code writer to **direction-setter and key decision-maker**.

`LinaPro` was designed from the ground up with AI-driven development in mind. On one hand, it provides a spec-driven AI development workflow so that developers can hand off requirements analysis, system design, code implementation, and test verification entirely to AI. On the other hand, it ships a built-in AI skill system covering the full development lifecycle, giving AI the domain-specific knowledge to make framework-compliant decisions in every concrete work context — backend development, frontend design, test coverage, performance auditing, and more. These two dimensions complement each other and together form the foundation of `LinaPro`'s AI-native design.

## AI-Native vs. AI-Assisted

| Dimension | AI-Assisted | AI-Native |
|-----------|-------------|-----------|
| **AI's role** | Optional support tool | Core engineering productivity, primary path |
| **Where code comes from** | Humans write, AI assists | AI writes, humans decide |
| **Documentation** | Humans maintain, prone to drifting from code | AI maintains in sync, tightly coupled to code |
| **Architecture consistency** | Relies on manual review | Spec-anchored, AI automatically follows |
| **Test coverage** | Written as needed, gaps possible | Mandatory E2E tests, a prerequisite for every change |
| **Iteration speed** | Bounded by human coding speed | Bounded by AI processing speed |

## AI Across the Full Development Lifecycle

In `LinaPro`'s AI-native workflow, AI is deeply involved at every stage of development:

```mermaid
flowchart TD
    A["Describe the requirement"] --> B["AI explores & analyzes\nUnderstands existing architecture\nIdentifies risk boundaries"]
    B --> C["AI generates proposal\nproposal + design + tasks\nIncremental spec documents"]
    C --> D{"User confirms direction"}
    D -->|"Adjust direction"| B
    D -->|"Confirm & begin"| E["AI implements code\nBackend API + service layer\nFrontend pages + routes\nSQL scripts + tests"]
    E --> F["AI auto-reviews\nCode quality + spec compliance\nE2E tests passing"]
    F --> G{"Review passed?"}
    G -->|"Issues found"| E
    G -->|"All passed"| H["User validates feature"]
    H --> I{"Feature meets expectations?"}
    I -->|"Needs adjustment"| J["User feedback\nAI corrects implementation"]
    J --> F
    I -->|"Confirmed complete"| K["AI archives change\nSpecs consolidated as baseline"]
```

- **Requirements analysis:** `AI` reads the project specs (`AGENTS.md`/`CLAUDE.md`), existing `OpenSpec` documents, and source code to deeply understand the current architecture. It proactively asks clarifying questions to help you think through scope and impact.

- **System design:** `AI` produces database schema designs, API interface definitions, and frontend page structures, forming a complete design document (`design.md`) and incremental specs (`specs/`) for your review and approval.

- **Implementation:** `AI` works through the task list (`tasks.md`) item by item — backend services, API routes, database DAO layer, frontend page components, permission declarations, and i18n resources, all in one pass.

- **Testing:** `AI` writes E2E test cases, runs them, and ensures the implementation matches the design specification.

- **Documentation archival:** `AI` organizes the design decisions and API specs from the change into baseline documentation for future iterations to reference.

## Built-in AI Skill System

AI-native is not just about a development workflow — it also means the framework provides deep, purpose-built skills for every stage of AI work. `LinaPro` ships a built-in AI skill system covering the full development lifecycle, so AI can work at maximum effectiveness in every concrete scenario without requiring developers to re-explain project conventions at the start of every session.

| Category | Skill | Core capability |
|----------|-------|----------------|
| **Dev workflow** | `openspec-*` | Requirements exploration, change proposals, design document generation |
| **Dev workflow** | `lina-review` | Code and spec compliance review — the quality gate before archiving an `OpenSpec` change |
| **Dev workflow** | `lina-feedback` | Issue tracking, bug fixes, and E2E test coverage closure |
| **Backend** | `goframe-v2` | `GoFrame` development conventions, best practices for API, service, and DAO layers |
| **Frontend** | `frontend-design` | Production-grade frontend design that avoids generic AI-generated interface patterns |
| **Frontend** | `frontend-patterns` | Frontend development patterns: state management, performance optimization, form handling |
| **Testing** | `lina-e2e` | `Playwright` E2E test case naming, organization, and writing conventions |
| **Testing** | `playwright-cli` | Browser automation for E2E test execution and debugging |
| **Quality** | `lina-perf-audit` | Backend API performance auditing — surfaces N+1 queries, missing indexes, and similar issues |
| **Quality baseline** | `karpathy-guidelines` | Guards against common LLM coding pitfalls to keep implementations precise and lean |
| **Version management** | `lina-upgrade` | Framework and source plugin version upgrades, AI-guided and script-driven |
| **Version management** | `git-commit-push` | Analyzes the diff and generates commit messages that match the repository's conventions |
| **Version management** | `git-worktree` | Creates isolated `Git` worktrees for parallel tasks to avoid conflicts |

These skills are embedded in the framework's AI collaboration spec (`.agents/skills/`) as domain knowledge — no installation required. AI tools activate the relevant skill automatically when handling the corresponding scenario. The result: **in every concrete work context, AI makes decisions that comply with `LinaPro`'s framework constraints** rather than falling back on general model knowledge to produce code or documentation that doesn't fit the project.

## Spec-Driven Development

`LinaPro`'s AI-native workflow is built on **Specification-Driven Development (SDD)**.

The problem with traditional development is that, over time, documentation falls behind code, code drifts from design, test coverage develops gaps, and the codebase becomes legacy that nobody dares touch.

SDD addresses this through the following mechanisms:

### Every change has a corresponding spec anchor

Each feature iteration produces a complete documentation record under `openspec/changes/`, including design decisions, API specs, and implementation details. These documents serve as AI's context for the next iteration.

### Specs first, code follows

Before a single line of code is written, AI must finalize the spec (`proposal.md` and `design.md` reviewed and approved), and only then begins implementation. This ensures code and documentation are produced in sync within the same iteration cycle.

### Mandatory E2E tests as a prerequisite for every change

Every change must have corresponding E2E test coverage. Passing tests is a required condition for archiving a change — not an optional step. This fundamentally prevents the accumulation of test gaps.

### Archiving is consolidation

When a change is archived, its incremental specs are merged into the `openspec/specs/` directory, building a continuously growing baseline spec library. In future iterations, AI uses these validated baselines as constraints to maintain architectural consistency.

## Why Direct Code Edits Are Discouraged

This is a common question from developers new to `LinaPro`: can I just edit the code directly without going through the AI workflow?

**You can — but it's not recommended.** Here's why:

### Documentation and code immediately diverge

`LinaPro` uses `SDD OpenSpec` spec documents (`design.md`, `specs/`) to describe the current state of the system. If you modify code without updating the specs, the next AI iteration will generate designs based on outdated specs, producing increasingly divergent results.

### AI cannot sense "implicit rules"

When humans modify code directly, they often introduce unstated constraints — a special case added here, a limitation bypassed there — that exist only in the developer's head, not in the spec documents. When AI next modifies the same module, it will likely break these implicit constraints.

### AI can guarantee implementation-documentation consistency

When AI leads the implementation, code and documentation are produced in the same conversation turn. They are necessarily consistent — a level of synchronization that is hard for humans to match.

### Exceptions

The following scenarios may warrant direct code changes:

- Urgent production hotfixes (but add an `OpenSpec` record afterward)
- Highly specialized business constraints that AI cannot adequately understand
- Configuration file and environment variable adjustments

Even in these cases, it is recommended to use the `/lina-feedback` skill afterward to record the change in the currently active `OpenSpec` change.

## The Value of an AI-Native Framework

`LinaPro`'s AI-native design delivers value along five dimensions:

- **Delivery speed**: AI can complete a full CRUD module — backend API, database layer, frontend pages, and test cases — in minutes, work that would take a human engineer several hours.

- **Consistent quality**: Every iteration follows the same standards and best practices. Code style, naming conventions, error handling, and test coverage remain highly consistent, independent of individual developer habits.

- **Sustainable evolution**: Spec documents are updated alongside code. Architectural decisions are documented. New team members can quickly understand the system's design intent by reading `openspec/specs/`.

- **Controlled risk**: Every change has E2E tests as a safety net. AI's review mechanism catches spec-divergent implementations before merge.

- **Context-aware execution**: More than ten built-in AI skills cover the full development lifecycle. Whether the task is backend API development, frontend interface design, E2E test writing, performance auditing, version upgrades, or code commits, AI has dedicated domain knowledge for every specific work context. Developers don't need to re-explain project conventions at the start of each session, which dramatically reduces the cognitive overhead of AI collaboration.
