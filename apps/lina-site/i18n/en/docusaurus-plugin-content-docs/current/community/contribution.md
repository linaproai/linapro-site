---
slug: '/community/contribution'
title: 'Contribution Guide'
sidebar_position: 2
description: 'This contribution guide explains practical ways to help LinaPro, including reporting reproducible issues, improving documentation, proposing framework changes through OpenSpec, contributing code with focused tests, and keeping changes aligned with the repository architecture, plugin boundaries, and bilingual documentation workflow.'
keywords:
  - LinaPro
  - contribution guide
  - open source contribution
  - GitHub issues
  - pull request
  - documentation contribution
  - code contribution
  - OpenSpec
  - E2E tests
  - plugin boundaries
  - architecture governance
  - bilingual documentation
  - issue report
  - feature proposal
  - sustainable delivery
  - community
---

Contributions are most useful when they are small, reproducible, and aligned with the framework boundaries.

## Ways To Contribute

| Contribution | What good looks like |
| --- | --- |
| Bug report | Includes version, command, expected behavior, actual behavior, and logs without secrets. |
| Documentation fix | Updates Chinese and English content together when the affected page is bilingual. |
| Feature proposal | Starts from an `OpenSpec` change when behavior, architecture, or tests need to evolve. |
| Code change | Keeps ownership narrow, adds focused tests, and avoids unrelated refactors. |
| Plugin example | Follows the structure used by `plugin-demo-source` or `plugin-demo-dynamic`. |

## Documentation Workflow

Official site content follows a Chinese-first process. Confirm the Chinese content first, then synchronize the English version with natural English instead of word-for-word translation.

When editing official documentation:

1. Keep front matter complete, including `description` and `keywords`.
2. Use `Mermaid` for architecture and workflow diagrams.
3. Prefer concise tables for configuration, module, or command references.
4. Keep examples tied to facts in the `linapro` source repository.

## Code Workflow

For code changes, prefer a focused pull request that explains the problem, the chosen approach, and the verification performed. Shared behavior, plugin lifecycle changes, and user-facing workflows should include tests that can catch regressions.
