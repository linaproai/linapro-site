---
slug: '/quick/agent-tools'
title: 'AI Tool Integration'
hide_title: true
description: 'This page explains how mainstream AI coding tools support the .agents/skills/ skills directory, the AGENTS.md project instruction file, and tool-specific files such as CLAUDE.md in the LinaPro framework. It also documents the LinaPro-provided `make skills` command suite for managing per-tool symlinks consistently across Windows, Linux and macOS, helping teams onboard Claude Code, OpenAI Codex, Cursor, GitHub Copilot, Gemini CLI, Windsurf, Amp, OpenCode, Devin, Qoder, CodeBuddy, Roo Code and other tools.'
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
  - make skills
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

> The following tools are commonly used by community team members in `AI Coding` practice.

| Tool | Skills Directory | Project Instructions | Integration Guidance |
|------|------------------|----------------------|----------------------|
| Claude Code | `.claude/skills/` | ✅ Supported | Reads `CLAUDE.md` by default; LinaPro already supports this through a symlink. Run `make skills.link AGENT=claude-code` |
| OpenAI Codex | ✅ Supported | ✅ Supported | Natively matches LinaPro's current directory structure |
| Cursor | ✅ Supported | ✅ Supported | Put simple conventions in `AGENTS.md` and more complex rules in `.cursor/rules/` |
| OpenCode | ✅ Supported | ✅ Supported | Supports both `AGENTS.md` and `Claude Code` compatible files |
| Windsurf | `.windsurf/skills/` | ✅ Supported | Run `make skills.link AGENT=windsurf` |
| Cline | ✅ Supported | ✅ Supported | Put shared guidance in `AGENTS.md` and `Cline`-specific rules in `.clinerules/` |
| GitHub Copilot | ✅ Supported | ✅ Supported | Also supports `.github/copilot-instructions.md` configuration |

## Complete Tool Matrix

> If anything is missing or inaccurate, please submit a `PR` or `Issue` to help update it.

| Tool | Skills Directory | Project Instructions | Integration Guidance |
|----------------------|----------|--------------|------------------|
| Amp | ✅ Supported | ✅ Supported | Reuse the current `AGENTS.md` and standard skills directory directly |
| Antigravity | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| Claude Code | ✅ Supported | ✅ Supported | Reads `CLAUDE.md` by default; LinaPro already supports this through a symlink. Run `make skills.link AGENT=claude-code` |
| OpenClaw | `skills/` | ⚠️ To be confirmed | Conflicts with the repo-root `skills/` path; explicitly opt in via `make skills.link AGENT=openclaw FORCE=1`. Validate instruction file loading behavior |
| Cline | ✅ Supported | ✅ Supported | Put shared guidance in `AGENTS.md` and `Cline`-specific rules in `.clinerules/` |
| CodeBuddy | `.codebuddy/skills/` | ⚠️ To be confirmed | Run `make skills.link AGENT=codebuddy`; validate instruction file loading behavior |
| OpenAI Codex | ✅ Supported | ✅ Supported | Natively compatible with the current `LinaPro` conventions |
| Cursor | ✅ Supported | ✅ Supported | Put simple rules in `AGENTS.md` and scoped rules in `.cursor/rules/` |
| Deep Agents | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| Devin | `.devin/skills/` | ✅ Supported | Run `make skills.link AGENT=devin` |
| Dexto | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| Droid | `.factory/skills/` | ✅ Supported | Run `make skills.link AGENT=droid` |
| Firebender | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| ForgeCode | `.forge/skills/` | ⚠️ To be confirmed | Run `make skills.link AGENT=forgecode`; validate instruction file loading behavior |
| Gemini CLI | ✅ Supported | ⚙️ Configurable | Recommended symlink: `ln -s AGENTS.md GEMINI.md` |
| GitHub Copilot | ✅ Supported | ✅ Supported | Also supports `.github/copilot-instructions.md` configuration |
| Goose | `.goose/skills/` | ✅ Supported | Run `make skills.link AGENT=goose` |
| Hermes Agent | `.hermes/skills/` | ⚠️ To be confirmed | Run `make skills.link AGENT=hermes-agent`; validate instruction file loading behavior |
| Kimi Code CLI | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| OpenCode | ✅ Supported | ✅ Supported | Reads `AGENTS.md` natively and can also migrate `Claude Code` projects |
| Pi | `.pi/skills/` | ⚠️ To be confirmed | Run `make skills.link AGENT=pi`; validate instruction file loading behavior |
| Qoder | `.qoder/skills/` | ✅ Supported | Run `make skills.link AGENT=qoder` |
| Qwen Code | `.qwen/skills/` | ⚠️ To be confirmed | Run `make skills.link AGENT=qwen-code`; for instruction file compatibility also `ln -s AGENTS.md QWEN.md` |
| Replit | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| Rovo Dev | `.rovodev/skills/` | ⚠️ To be confirmed | Run `make skills.link AGENT=rovodev`; validate instruction file loading behavior |
| Roo Code | `.roo/skills/` | ✅ Supported | Run `make skills.link AGENT=roo` |
| Trae | `.trae/skills/` | ⚠️ To be confirmed | Run `make skills.link AGENT=trae`; validate instruction file loading behavior |
| Warp | ✅ Supported | ✅ Supported | Reuse the standard skills directory first; validate project instruction loading behavior |
| Windsurf | `.windsurf/skills/` | ✅ Supported | Run `make skills.link AGENT=windsurf` |
| Universal | ✅ Supported | Depends on the target tool | Use this for `AGENTS.md`-compatible tools that are not listed explicitly |

## Using LinaPro's Unified Management Commands

`LinaPro` ships a built-in `make skills` command suite that unifies symlink management for the agents above, removing the need for hand-written `ln -s` invocations and guaranteeing consistent behavior across `Windows`, `Linux` and `macOS`. The commands operate strictly inside the repository root and never touch `HOME` directories or any system-global paths.

```bash
# Interactive action menu on a TTY (link / unlink / quit)
make skills

# Interactive selection on a TTY; read-only status table on CI/pipes
make skills.link

# Non-interactive: create symlinks for selected agents
make skills.link AGENT=claude-code
make skills.link AGENT=claude-code,codebuddy,qoder
make skills.link AGENT=all                # link every link-class agent
make skills.link AGENT=all FORCE=1        # rebuild mismatched symlinks

# Interactive selection of currently managed symlinks to remove
make skills.unlink

# Non-interactive: remove symlinks for selected agents
make skills.unlink AGENT=claude-code
make skills.unlink AGENT=all
```

**Agent categories** (aligned with the [vercel-labs/skills](https://github.com/vercel-labs/skills#supported-agents) project-path table):

- `native` — the project path is already `.agents/skills` (e.g. `cursor`, `gemini-cli`, `codex`, `amp`, `opencode`, `cline`, `github-copilot`). No symlink is needed.
- `link` — the project path lives under a tool-specific directory (e.g. `claude-code` → `.claude/skills`, `codebuddy` → `.codebuddy/skills`, `windsurf` → `.windsurf/skills`). A relative symlink to `.agents/skills` is created on demand.
- `rootCollision` — the project path is `skills/` at the repository root (currently only `openclaw`). Skipped by default; pass `AGENT=openclaw FORCE=1` to opt in.

**Conflict protection rules**:

- The commands never auto-remove existing real directories or files — even with `FORCE=1`.
- `FORCE=1` only rebuilds symlinks that already exist but point at a non-managed target.
- All per-tool symlink directories are listed in `.gitignore`, so creating them locally never pollutes the repository.

The interactive `make skills.link` selection uses a 3-column grid with status glyphs:

- `[+]` linked — already pointing at `.agents/skills`
- `[~]` mismatch — symlink exists but targets another location
- `[.]` absent — no symlink yet
- `[!]` conflict — a real directory or file blocks linking
- `[*]` root-collision — agent uses the repo-root `skills/` path
- `[?]` error — inspection failed; see the non-interactive listing for details

See [the `linactl` README](https://github.com/linaproai/linapro/blob/main/hack/tools/linactl/README.md) for additional details.
