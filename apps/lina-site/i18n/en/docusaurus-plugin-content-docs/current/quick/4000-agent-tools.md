---
slug: '/quick/agent-tools'
title: 'AI Tool Integration'
hide_title: true
description: 'This page explains how mainstream AI coding tools support the .agents/skills/ skills directory, the AGENTS.md project instruction file, and tool-specific files such as CLAUDE.md in the LinaPro framework. It summarizes Agent Skills installation targets from vercel-labs/skills and public documentation findings, helping teams decide how Claude Code, OpenAI Codex, Cursor, GitHub Copilot, Gemini CLI, Windsurf, Amp, OpenCode, Devin, Qoder, and other tools can consume LinaPro AI-native development conventions.'
keywords:
  - LinaPro
  - AI tool integration
  - Agent Skills
  - AGENTS.md
  - CLAUDE.md
  - GEMINI.md
  - Claude Code
  - OpenAI Codex
  - Cursor
  - GitHub Copilot
  - Gemini CLI
  - Windsurf
  - Amp
  - OpenCode
  - Devin
  - Qoder
  - Cline
  - Roo Code
  - vercel-labs
  - skills CLI
  - .agents/skills
  - AI coding tools
  - project instructions
  - tool compatibility
  - AI-native development
---

## Background

The `LinaPro` framework includes a complete `AI`-native development support system. Two mechanisms are central to that system:

- **The `.agents/skills/` skills directory**: Stores structured `Agent Skills`, follows the open [Agent Skills specification](https://agentskills.io/), and is the most commonly used skills directory.
- **The `AGENTS.md` instruction file**: Provides project-level instructions for `AI` agents, covering repository architecture, development workflows, technical constraints, and verification requirements, so connected `AI Coding` tools receive durable project context.

Today, `AI Coding` tools are highly diverse. Each tool may read skills directories and project instruction files from different paths, and there is no single universal convention. This page summarizes and clarifies those differences so teams can connect common development tools more effectively.

## Status Legend

| Mark | Meaning |
|------|---------|
| ✅ Supported | The tool's official documentation explicitly supports the skills directory and instruction files provided by `LinaPro` by default |
| ⚙️ Configurable | The default file name is different, but the tool can reuse it through configuration, import, or a symlink |
| ⚠️ To be confirmed | Only `skills CLI` skill installation support has been confirmed; project instruction file loading behavior has not been confirmed |

## Recommended Mainstream Tools

| Tool | Skills Directory | Project Instructions | Integration Guidance |
|------|------------------|----------------------|----------------------|
| Claude Code | `.claude/skills/` | ✅ Supported | Reads `CLAUDE.md` by default; LinaPro already supports this through a symlink |
| OpenAI Codex | ✅ Supported | ✅ Supported | Natively matches LinaPro's current directory structure |
| Cursor | ✅ Supported | ✅ Supported | Put simple conventions in `AGENTS.md` and more complex rules in `.cursor/rules/` |
| OpenCode | ✅ Supported | ✅ Supported | Supports both `AGENTS.md` and `Claude Code` compatible files |
| Windsurf | `.windsurf/skills/` | ✅ Supported | Recommended symlink: `ln -s .agents .windsurf` |
| Cline | ✅ Supported | ✅ Supported | Put shared guidance in `AGENTS.md` and `Cline`-specific rules in `.clinerules/` |
| GitHub Copilot | ✅ Supported | ✅ Supported | Also supports `.github/copilot-instructions.md` configuration |

## Complete Tool Matrix

| Tool | Skills Directory | Project Instructions | Integration Guidance |
|----------------------|----------|--------------|------------------|
| Amp | ✅ Supported | ✅ Supported | Reuse the current `AGENTS.md` and standard skills directory directly |
| Antigravity | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| Claude Code | ✅ Supported | ✅ Supported | Reads `CLAUDE.md` by default; LinaPro already supports this through a symlink |
| OpenClaw | `skills/` | ⚠️ To be confirmed | Install skills under the tool-specific directory; validate instruction file loading behavior |
| Cline | ✅ Supported | ✅ Supported | Put shared guidance in `AGENTS.md` and `Cline`-specific rules in `.clinerules/` |
| CodeBuddy | `.codebuddy/skills/` | ⚠️ To be confirmed | Recommended symlink: `ln -s .agents .codebuddy`; validate instruction file loading behavior |
| OpenAI Codex | ✅ Supported | ✅ Supported | Natively compatible with the current `LinaPro` conventions |
| Cursor | ✅ Supported | ✅ Supported | Put simple rules in `AGENTS.md` and scoped rules in `.cursor/rules/` |
| Deep Agents | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| Devin | `.devin/skills/` | ✅ Supported | Recommended symlink: `ln -s .agents .devin` |
| Dexto | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| Droid | `.factory/skills/` | ✅ Supported | Recommended symlink: `ln -s .agents .factory` |
| Firebender | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| ForgeCode | `.forge/skills/` | ⚠️ To be confirmed | Recommended symlink: `ln -s .agents .forge`; validate instruction file loading behavior |
| Gemini CLI | ✅ Supported | ⚙️ Configurable | Recommended symlink: `ln -s AGENTS.md GEMINI.md` |
| GitHub Copilot | ✅ Supported | ✅ Supported | Also supports `.github/copilot-instructions.md` configuration |
| Goose | `.goose/skills/` | ✅ Supported | Recommended symlink: `ln -s .agents .goose` |
| Hermes Agent | `.hermes/skills/` | ⚠️ To be confirmed | Recommended symlink: `ln -s .agents .hermes`; validate instruction file loading behavior |
| Kimi Code CLI | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| OpenCode | ✅ Supported | ✅ Supported | Reads `AGENTS.md` natively and can also migrate `Claude Code` projects |
| Pi | `.pi/skills/` | ⚠️ To be confirmed | Recommended symlink: `ln -s .agents .pi`; validate instruction file loading behavior |
| Qoder | `.qoder/skills/` | ✅ Supported | Recommended symlink: `ln -s .agents .qoder` |
| Qwen Code | `.qwen/skills/` | ⚠️ To be confirmed | Recommended symlinks: `ln -s .agents .qwen`, `ln -s AGENTS.md QWEN.md` |
| Replit | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| Rovo Dev | `.rovodev/skills/` | ⚠️ To be confirmed | Recommended symlink: `ln -s .agents .rovodev`; validate instruction file loading behavior |
| Roo Code | `.roo/skills/` | ✅ Supported | Recommended symlink: `ln -s .agents .roo` |
| Trae | `.trae/skills/` | ⚠️ To be confirmed | Recommended symlink: `ln -s .agents .trae`; validate instruction file loading behavior |
| Warp | ✅ Supported | ✅ Supported | Reuse the standard skills directory first; validate project instruction loading behavior |
| Windsurf | `.windsurf/skills/` | ✅ Supported | Recommended symlink: `ln -s .agents .windsurf` |
| Universal | ✅ Supported | Depends on the target tool | Use this for `AGENTS.md`-compatible tools that are not listed explicitly |
