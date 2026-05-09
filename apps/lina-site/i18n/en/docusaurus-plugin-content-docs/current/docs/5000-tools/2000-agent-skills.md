---
slug: '/docs/agent-skills'
title: 'Dev Skills'
hide_title: true
description: 'A complete guide to LinaPro AI skills — the optional OpenSpec CLI, goframe-v2, and find-skills that can be installed externally, plus 13 project-bundled skills covering frontend design, Git workflows, code review, E2E testing, and performance auditing. Includes installation instructions and usage guidance for every skill.'
keywords:
  - AI skills
  - dev skills
  - LinaPro skills
  - Claude Code skills
  - OpenSpec
  - goframe-v2
  - find-skills
  - skill installation
  - frontend design skill
  - E2E test skill
  - performance audit skill
  - code review skill
  - git commit skill
  - Playwright skill
  - AI collaboration
  - skills library
  - framework skills
---

`LinaPro`'s AI skill library consists of two categories:

- **Externally installed skills**: Installed globally via `npm`/`npx` and available across all projects — `OpenSpec`, `goframe-v2`, and `find-skills`.
- **Project-bundled skills**: Shipped with the project source under `.agents/skills/`, loaded automatically when you work inside the project directory with `Claude Code` or another compatible AI tool.

## Skills Overview

| Skill | Use case | Prerequisites |
|-------|----------|---------------|
| `openspec` | Optional spec-driven workflow tool — recommended for AI dev workflows | Recommended (`npm`) |
| `goframe-v2` | GoFrame-specific AI skill for higher-quality backend code | Recommended (`npx skills`) |
| `find-skills` | AI skill marketplace search to help select the right skills | Recommended (`npx skills`) |
| `frontend-design` | Create distinctive, high-quality frontend interfaces | None |
| `frontend-patterns` | React / frontend best practices and design patterns | None |
| `git-commit-push` | Generate a conventional commit message and push | None |
| `git-worktree` | Create an isolated Git worktree for a task | None |
| `karpathy-guidelines` | Behavioral guardrails to reduce AI coding mistakes | None |
| `lina-archive-consolidate` | Consolidate archived OpenSpec iterations | None |
| `lina-e2e` | E2E test case naming conventions and management | Playwright required |
| `lina-feedback` | Track and fix issues reported after an OpenSpec change | None |
| `lina-perf-audit` | Comprehensive backend API performance audit (manual only) | Running LinaPro environment |
| `lina-review` | Structured code and spec review | None |
| `openspec-explore` | Thinking-partner mode for exploring ideas and problems | None |
| `openspec-propose` | Generate a complete OpenSpec change proposal in one step | None |
| `playwright-cli` | Browser automation and Playwright test authoring | Playwright required |

## How to Use Skills

**Externally installed skills**: Install them globally using the commands in each skill's section. Once installed, `Claude Code` automatically recognizes and applies them in the relevant contexts.

**Project-bundled skills**: Loaded automatically when an AI tool runs inside the project directory — no setup needed. Invoke a skill by describing your scenario in the conversation, or use a slash command (e.g. `/git-commit-push`) to trigger it explicitly.

## Externally Installed Skills

These three skills must be installed manually. Once installed, they take effect globally across all your projects.

### openspec

`OpenSpec` is the recommended command-line tool for `LinaPro`'s spec-driven development workflow. Once installed, workflow skills such as `/opsx:explore`, `/opsx:propose`, `/opsx:apply`, and `/opsx:archive` will automatically use it as their underlying engine. It is **recommended** for the `LinaPro` AI dev workflow.

**When to use**: Running any step of the `OpenSpec` workflow — exploration, proposals, implementation, and archiving.

**Installation**:

```bash
npm install -g @fission-ai/openspec@latest
```

Verify the installation with `opsx --version`.

### goframe-v2

`goframe-v2` is a `Claude Code` skill purpose-built for `LinaPro` backend `Go` code. It embeds `GoFrame` coding conventions, `ORM` usage patterns, and best-practice examples. The skill activates automatically when you write or modify backend `Go` files, implement service interfaces, create APIs, or work with data access layers — improving code generation quality and keeping output aligned with framework conventions.

**When to use**: Developing `lina-core` backend code, implementing business interfaces, or writing the data access layer.

**Installation**:

```bash
npx skills add github.com/gogf/skills -g
```

### find-skills

`find-skills` is an AI skill marketplace search tool that helps you quickly discover and evaluate skills suited to your project. When you need to extend your AI toolchain and aren't sure what already exists, `find-skills` makes the selection process much faster.

**When to use**: Looking for new AI skills to bring into the project or evaluating the available skill ecosystem.

**Installation**:

```bash
npx skills add find-skills -g
```

## Project-Bundled Skills

The following skills are maintained in `.agents/skills/` alongside the project source. No installation is required — they load automatically when you use an AI tool inside the project directory.

### frontend-design

Guides the AI toward creating distinctive, production-grade frontend interfaces rather than generic "AI slop" aesthetics. Covers design intent analysis (page purpose, tone, differentiation), typography and color palettes, motion design, spatial composition, and background / visual detail treatment.

**When to use**: Building a new page, redesigning an existing interface, or whenever you want frontend work to have a considered visual direction.

**How to invoke**: Describe what you want to build and the AI applies these guidelines automatically, or use `/frontend-design`.

### frontend-patterns

Provides the AI with a comprehensive reference for modern frontend development: component composition, compound components, render props, custom hooks, `Context + Reducer` state management, memoization, code splitting, virtualization, form validation, error boundaries, `Framer Motion` animations, and keyboard accessibility patterns.

**When to use**: Implementing frontend components, choosing a state management approach, optimizing performance, or adding accessibility support.

**How to invoke**: Activates automatically in frontend development contexts, or use `/frontend-patterns`.

### git-commit-push

Reviews the current `Git` workspace, generates a conventional commit message from the actual diff following the repository's commit format (with type prefixes like `fix`, `feat`, `build`, `docs`), stages all changes, commits, and pushes to `origin`.

**When to use**: After finishing a feature or fix and ready to commit and push.

**How to invoke**: Say "commit and push" or use `/git-commit-push`.

### git-worktree

Creates a dedicated `Git` worktree so the AI continues work in a fresh, isolated directory without touching the current checkout. Especially useful when juggling multiple feature branches in parallel.

**When to use**: Isolating a development task from the main working tree, or working on multiple branches simultaneously.

**How to invoke**: Use `/git-worktree` — the AI derives a branch name from the task description automatically.

### karpathy-guidelines

A set of behavioral guardrails that reduce common LLM coding mistakes, inspired by Andrej Karpathy's observations on LLM coding behavior.

Four core principles: **Think before coding** (state assumptions and tradeoffs explicitly), **Simplicity first** (write the minimal code that solves the problem), **Surgical changes** (only touch what is necessary), **Goal-driven execution** (define verifiable success criteria and loop until verified).

**When to use**: All coding tasks — these guidelines act as a baseline behavioral constraint for AI-assisted development.

**How to invoke**: Applied automatically as a baseline; rarely needs explicit triggering.

### lina-archive-consolidate

Reads archived `OpenSpec` iterations from `openspec/changes/archive/`, groups them by functional responsibility, and performs semantic merging — combining multiple iterations into a single cohesive archive while preserving all design decisions and implementation details. When the archive count reaches 8 or more, it automatically enters the full `OpenSpec` workflow (explore → propose → apply).

**When to use**: The archive directory has accumulated many scattered iteration records and needs periodic cleanup.

**How to invoke**: Use `/lina-archive-consolidate`.

### lina-e2e

Defines the project's `Playwright E2E` test case management standards: file naming (globally unique `TC{NNNN}` IDs that are never reused), module-based directory layout, test file templates, independence requirements (one test per file, fully self-contained), and page object model usage.

**When to use**: Writing, maintaining, or reviewing E2E test cases.

**Prerequisites**: Install `Playwright` and its test dependencies first:

```bash
cd hack/tests
pnpm install
npx playwright install --with-deps chromium
```

**How to invoke**: The AI applies these standards automatically when writing E2E tests, or use `/lina-e2e`.

### lina-feedback

A structured workflow for capturing and resolving issues reported after an `OpenSpec` change lands. Covers problem classification by type (bug / missing / UX / test gap), incremental spec updates, task organization in `tasks.md`, fix execution loops, regression analysis, and pre-completion verification.

**When to use**: After an `OpenSpec` change has been implemented and feedback reports need systematic triaging and fixing.

**How to invoke**: Use `/lina-feedback` and describe the feedback to address.

### lina-perf-audit

A comprehensive `LinaPro` backend API performance audit workflow (**manual trigger only**). Audits for N+1 queries, missing indexes, unbounded list responses, redundant reads, mergeable SQL calls, blocking operations inside loops, and write SQL executed inside read endpoints.

The audit runs in three phases: Phase 0 (environment setup — database reset, service restart, plugin installation, load-test data seeding), Phase 1 (concurrent sub-agent audits per module), Phase 2 (aggregation — summary report and persistent issue cards under `perf-issues/`).

:::caution Heads up

This skill resets the database, restarts services, and spawns multiple sub-agents. It takes tens of minutes to several hours and consumes significant token budget. Only trigger it when you explicitly need a full performance audit.

:::

**When to use**: When you explicitly need a systematic performance audit of all backend APIs.

**Prerequisites**: A running `LinaPro` dev environment started with `make dev`.

**How to invoke**: Explicitly say "run `lina-perf-audit`" or "run a full backend API performance audit" in the conversation.

### lina-review

A structured code and spec review workflow, triggered after `OpenSpec` apply or feedback tasks complete and before archiving. Covers backend code review (`GoFrame` conventions, API i18n compliance, distributed cache consistency), RESTful API review, SQL review, E2E test review, and report generation. `AGENTS.md` is the single source of truth for all review standards.

**When to use**: After an `OpenSpec` change has been implemented, as the quality gate before archiving.

**How to invoke**: Use `/lina-review` — the AI determines the review scope automatically based on the change.

### openspec-explore

Thinking-partner mode for exploring ideas, investigating problems, and clarifying requirements. The AI approaches problems with curiosity and without preconceptions — it reads files, probes the codebase, compares options, draws ASCII diagrams, and surfaces risks and unknowns.

Explore mode is for **thinking and investigation**, not implementation. It can produce `OpenSpec` artifacts (proposals, designs) as outputs of the thinking process, but it does not implement production features.

**When to use**: Before starting a complex task, or whenever you need to think through requirements or architectural options.

**How to invoke**: Use `/openspec-explore` and describe what you want to explore.

### openspec-propose

Generates a complete `OpenSpec` change with all required artifacts in a single step: `proposal.md`, `design.md`, `tasks.md`, and any `specs/` files. The AI creates artifacts in dependency order to keep design docs and task lists internally consistent.

**When to use**: Kicking off a new feature or change that needs a full `OpenSpec` proposal.

**How to invoke**: Use `/openspec-propose` and describe the change you want to make.

### playwright-cli

A comprehensive `Playwright` CLI command reference for browser automation: navigating to URLs, clicking, typing, form submission, multi-tab workflows, cookie / `localStorage` / `sessionStorage` operations, network route interception, and `DevTools` integration.

**When to use**: Automating browser interactions, writing Playwright tests, or debugging browser-based workflows.

**Prerequisites**: Install `Playwright` and the required browsers:

```bash
cd hack/tests
pnpm install
npx playwright install --with-deps chromium
```

**How to invoke**: Describe the browser interactions you need, or use `/playwright-cli`.
