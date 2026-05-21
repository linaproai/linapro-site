---
slug: '/docs/agent-skills'
title: 'Development Skills'
hide_title: true
description: 'This page introduces LinaPro''s AI-focused skill system, including the optional externally installed OpenSpec CLI, goframe-v2, and find-skills tools or skills, plus project-bundled skills for frontend design, Git workflows, code review, E2E testing, performance auditing, OpenSpec implementation and archiving, and other workflows that load automatically in the project. It also notes that source plugin upgrade scenarios may depend on the lina-upgrade skill provided by the AI tool environment, helping developers understand what each skill does, how to install it, and how to invoke it.'
keywords:
  - AI skills
  - development skills
  - LinaPro skills
  - Claude Code skills
  - OpenSpec
  - goframe-v2
  - find-skills
  - skill installation
  - frontend design skill
  - E2E testing skill
  - performance audit skill
  - code review skill
  - Git commit skill
  - Playwright skill
  - AI collaboration
  - skill system
  - framework skills
---

`LinaPro`'s `AI` skill system has two categories:

- **Externally installed skills**: Installed manually with `npm`/`npx` commands. Once installed, they apply globally across projects, such as `OpenSpec`, `goframe-v2`, and `find-skills`.
- **Project-bundled skills**: Stored with the project source under `.agents/skills/`. They require no extra installation and are loaded automatically when you work in the project directory with `AI` tools such as `Claude Code`.

## Skills Overview

| Skill | Use case | Prerequisites |
|-------|----------|---------------|
| `openspec` | Optional spec-driven workflow tool, recommended for this project | Recommended |
| `goframe-v2` | `GoFrame`-specific `AI` skill that improves backend code generation quality | Recommended |
| `find-skills` | `AI` skill marketplace search tool for choosing suitable skills | Recommended |
| `frontend-design` | Create high-quality, distinctive frontend interfaces | None |
| `frontend-patterns` | `React` / frontend development best practices and patterns | None |
| `git-commit-push` | Generate a conventional commit message and push to the remote | None |
| `git-worktree` | Create an isolated `Git` worktree for a development task | None |
| `karpathy-guidelines` | Behavioral guardrails that reduce `AI` coding mistakes | None |
| `lina-archive-consolidate` | Consolidate archived `OpenSpec` iterations into unified directories, typically for `CI` automation | None |
| `lina-auto-archive` | Automatically scan and archive completed `OpenSpec` changes, typically for `CI` automation | Requires `OpenSpec CLI` |
| `lina-e2e` | `E2E` test case naming and management standards | Requires `Playwright` |
| `lina-feedback` | Track and fix issues reported after `OpenSpec` changes | None |
| `lina-perf-audit` | Comprehensive backend `API` performance audit, manually triggered | Requires a running `LinaPro` service |
| `lina-review` | Structured code and specification review | None |
| `openspec-explore` | Thinking-partner mode for requirement exploration and problem analysis | None |
| `openspec-propose` | Generate a complete `OpenSpec` change proposal in one step | None |
| `openspec-apply-change` | Implement work according to an `OpenSpec` task list | Requires `OpenSpec CLI` |
| `openspec-archive-change` | Archive completed `OpenSpec` changes | Requires `OpenSpec CLI` |
| `playwright-cli` | Browser automation and `Playwright` testing | Requires `Playwright` |

## How to Use Skills

**Externally installed skills**: Follow each skill's installation command. After installation, `Claude Code` recognizes and applies the skill automatically in matching scenarios.

**Project-bundled skills**: `AI` tools automatically load all skills under `.agents/skills/` when running in the project directory. To use a skill, describe the scenario directly in the conversation, or invoke it explicitly with a slash command such as `/git-commit-push`.

## Externally Installed Skills

The following three skills require manual installation. After installation, they apply globally across all projects.

### openspec

`OpenSpec` is the recommended command-line tool for `LinaPro`'s spec-driven workflow. After installation, workflow skills such as `/opsx:explore`, `/opsx:propose`, `/opsx:apply`, and `/opsx:archive` automatically use it as the underlying engine. It is a **recommended** part of the `LinaPro AI` development workflow.

**When to use**: Running the full `OpenSpec` workflow, including exploration, proposal creation, implementation, and archiving.

**Installation**:

```bash
npm install -g @fission-ai/openspec@latest
```

After installation, run `opsx --version` to confirm it is available.

### goframe-v2

`goframe-v2` is a `Claude Code` skill tailored for `LinaPro` backend `Go` code. It includes `GoFrame` coding conventions, `ORM` usage patterns, and best-practice examples. When writing or modifying backend `Go` files, implementing service interfaces, creating `API`s, or working on database operations, the skill activates automatically and improves alignment with framework conventions through code generation, error diagnosis, and performance optimization guidance.

**When to use**: Developing `lina-core` backend code, implementing business interfaces, or writing data access layers.

**Installation**:

```bash
npx skills add github.com/gogf/skills -g
```

### find-skills

`find-skills` is an `AI` skill marketplace search tool that helps developers quickly discover and evaluate skills suited to the current project. When you need to introduce new `AI` assistance capabilities but are unsure whether an existing skill is available, `find-skills` makes skill selection more efficient.

**When to use**: Introducing new `AI` skills into the project or evaluating the current skill ecosystem.

**Installation**:

```bash
npx skills add vercel-labs/skills --skill find-skills -g
```

## Project-Bundled Skills

The following skills are maintained with the project source under `.agents/skills/`. They require no installation and load automatically when an `AI` tool is used in the project directory.

### frontend-design

High-quality frontend interface design guidance. This skill helps `AI` avoid generic interface output and produce frontend pages with a recognizable visual direction and stronger design quality.

It covers design intent analysis, including page purpose, tone, and differentiation; typography and color selection; motion and spatial composition; background and visual detail treatment; and the definition and precise execution of interface concepts.

**When to use**: Building a new page, redesigning an existing interface, or needing visual style guidance for frontend work.

**How to invoke**: Describe the interface requirement in the conversation and the `AI` applies the design guidance automatically, or invoke `/frontend-design` explicitly.

### frontend-patterns

Best practices and design patterns for `React` and modern frontend development. This skill gives the `AI` guidance for component composition, custom `Hook`s, state management, performance optimization, form handling, error boundaries, motion, and related frontend concerns.

**When to use**: Building frontend components, choosing a state management approach, implementing performance optimizations, or improving accessibility.

**How to invoke**: Activates automatically during frontend development work, or invoke `/frontend-patterns` explicitly.

### git-commit-push

Reviews the current `Git` workspace changes, generates a repository-compliant commit message, commits all changes to the current branch, and pushes the branch to `origin`.

Commit messages follow the conventional commit format, supporting type prefixes such as `fix`, `feat`, `build`, and `docs`, and the summary is derived from the actual `diff`.

**When to use**: After completing feature development or a fix and needing to commit and push code.

**How to invoke**: Tell the `AI` to "commit the code" or "commit and push", or invoke `/git-commit-push`.

### git-worktree

Creates an isolated `Git` worktree so the `AI` can continue work in a separate directory without affecting the state of the current checkout.

This is especially useful for handling multiple feature branches in parallel or exploring a new approach without disturbing the main working directory.

**When to use**: Isolating a task-specific development environment or working on multiple branches in parallel.

**How to invoke**: Invoke `/git-worktree`; the `AI` derives a task identifier for the worktree branch name.

### karpathy-guidelines

A set of behavioral guardrails for reducing common `AI` coding mistakes, inspired by `Andrej Karpathy`'s observations about `LLM` coding behavior.

The four principles are: think before coding (state assumptions and tradeoffs explicitly), simplicity first (solve the problem with minimal code), surgical changes (change only what is necessary), and goal-driven execution (define verifiable completion criteria and iterate until they pass).

**When to use**: All coding tasks, as a baseline constraint for `AI`-assisted development.

**How to invoke**: Applied automatically as a baseline guardrail; explicit invocation is rarely needed.

### lina-archive-consolidate

Consolidates multiple archived `OpenSpec` iterations under `openspec/changes/archive/` into unified archive directories grouped by functional responsibility.

During semantic consolidation, it preserves the design decisions and implementation details from every iteration while compressing multiple iterations into one cohesive archive, reducing historical directory noise. When the archive count reaches `8` or more, it automatically enters the full `OpenSpec` workflow: exploration, proposal, and implementation.

**When to use**: The archive directory has accumulated many scattered iteration records that should be periodically merged.

**How to invoke**: Invoke `/lina-archive-consolidate`.

### lina-auto-archive

Automatically scans `openspec/changes/`, batch-archives active changes that are complete, and summarizes the result.

The skill strictly checks each change's completion status: both the `OpenSpec` status and every task must be complete before archiving is allowed. Incomplete changes are skipped rather than forced through, and the final report lists successful archives and reasons for skipped changes in Chinese.

**When to use**: Batch-cleaning completed changes or periodically tidying the active change directory.

**Prerequisite**: Install `OpenSpec CLI` with `npm install -g @fission-ai/openspec@latest`.

**How to invoke**: Invoke `/lina-auto-archive`.

### lina-e2e

Naming conventions, directory structure, and management standards for `Playwright E2E` test cases. This skill defines test case `ID` allocation (`TC{NNNN}`, globally unique and never reused), module-based organization, test file templates, independence requirements, and page object model usage.

**When to use**: Writing, maintaining, or reviewing `E2E` test cases.

**Prerequisites**: Install `Playwright` and test dependencies:

```bash
cd hack/tests
pnpm install
npx playwright install --with-deps chromium
```

**How to invoke**: The `AI` applies these standards automatically when writing `E2E` tests, or invoke `/lina-e2e`.

### lina-feedback

A structured issue tracking and remediation workflow for `bug`s, missing functionality, experience problems, and test gaps reported after an `OpenSpec` change has been implemented.

The workflow covers target change identification, issue classification (`bug` / missing / UX / test gap), incremental spec updates, task list organization, iterative fixes, regression analysis, and comprehensive verification before completion.

**When to use**: Feedback is reported after an `OpenSpec` change and needs to be handled systematically.

**How to invoke**: Invoke `/lina-feedback` and describe the feedback to address.

### lina-perf-audit

A comprehensive `LinaPro` backend `API` performance audit workflow, **manual trigger only**. It audits for `N+1` queries, missing indexes, unbounded list responses, duplicate reads, mergeable `SQL` calls, blocking operations inside loops, and write `SQL` executed from read endpoints.

The audit runs in three phases: phase `0`, environment preparation with database reset, service restart, plugin installation, and load-test data seeding; phase `1`, concurrent sub-agent audits for each module; phase `2`, report aggregation and persistent issue card generation.

:::caution Note

This skill resets the database, restarts services, can take a long time from tens of minutes to hours, and consumes significant `Token` budget. Use it only when a full performance audit is explicitly needed.

:::

**When to use**: Explicitly requested systematic performance auditing for all backend `API`s.

**Prerequisite**: A running `LinaPro` development environment started with `make dev`.

**How to invoke**: Say "run `lina-perf-audit`" or "run a full backend `API` performance audit" in the conversation.

### lina-review

A structured code and specification review workflow triggered after `OpenSpec` implementation or feedback work is complete and before archiving, ensuring code quality and spec consistency.

Review scope includes backend code review (`GoFrame` conventions, `API` i18n compliance, distributed cache consistency), `RESTful API` review, `SQL` review, `E2E` test review, and final report generation. `AGENTS.md` is the single source of truth for all review standards.

**When to use**: After an `OpenSpec` change is implemented and before it is archived as a quality gate.

**How to invoke**: Invoke `/lina-review`; the `AI` determines the review scope from the change.

### openspec-explore

Explore mode is an `AI` thinking-partner mode for requirement analysis and problem investigation. It helps clarify the problem space, investigate the codebase, compare approaches, visualize architecture, and identify risks and unknowns with a curious, non-presumptive approach.

Explore mode focuses on **thinking and investigation** rather than implementation: it may read files, inspect the codebase, visualize with `ASCII` diagrams, and produce thinking artifacts such as `OpenSpec` proposals, but it does not directly implement production functionality.

**When to use**: Before starting a complex task, or when requirements or technical direction need clarification.

**How to invoke**: Invoke `/openspec-explore` and describe the problem or direction to explore.

### openspec-propose

Generates a complete `OpenSpec` change proposal in one step, automatically creating the change directory with `proposal.md`, `design.md`, `tasks.md`, and `specs/`.

The `AI` generates artifacts in dependency order so the design document and task list remain internally consistent, avoiding manual assembly of the full change file set.

**When to use**: Creating a full `OpenSpec` proposal for a new feature or change.

**How to invoke**: Invoke `/openspec-propose` and describe the change you want to implement.

### openspec-apply-change

Implements code, documentation, and test changes according to `tasks.md` in an `OpenSpec` change directory. This skill is usually triggered after the proposal and design are complete, and is responsible for moving the change from specification to runnable implementation.

**When to use**: An `openspec/changes/<change-id>/` directory already exists and you need to start or continue implementation.

**Prerequisite**: `OpenSpec CLI` is recommended.

**How to invoke**: Invoke `/openspec-apply-change`, or tell the `AI` to continue implementing the current `OpenSpec` change.

### openspec-archive-change

Archives a completed `OpenSpec` change by merging incremental specifications into baseline specifications and moving the change directory into the archive area.

**When to use**: The change implementation, review, and verification are complete and should be consolidated into the long-term baseline.

**Prerequisite**: `OpenSpec CLI` is recommended.

**How to invoke**: Invoke `/openspec-archive-change`, or ask the `AI` to archive the completed change.

### playwright-cli

A `Playwright` browser automation `CLI` command set covering navigation, clicks, typing, form submission, multi-tab management, `Cookie` / `localStorage` / `sessionStorage` operations, network route interception, `DevTools`, and other browser automation workflows.

**When to use**: Automating browser interactions, writing tests, or debugging `Playwright` tests.

**Prerequisites**: Install `Playwright` and the required browsers:

```bash
cd hack/tests
pnpm install
npx playwright install --with-deps chromium
```

**How to invoke**: Describe the browser operation you need, or invoke `/playwright-cli`.
