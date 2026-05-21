---
slug: '/quick/agent-tools'
title: 'AI Tool Integration'
hide_title: true
description: 'This page explains how mainstream AI coding tools support the .agents/skills/ skills directory, the AGENTS.md project instruction file, and tool-specific files such as CLAUDE.md in the LinaPro framework. It also documents the LinaPro-provided `make agents` command suite for managing per-tool symlinks consistently across Windows, Linux and macOS, helping teams onboard Claude Code, OpenAI Codex, Cursor, GitHub Copilot, Gemini CLI, Windsurf, Amp, OpenCode, Devin, Qoder, CodeBuddy, Roo Code and other tools.'
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
  - make agents
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
| Claude Code | `.claude/skills/` | ✅ Supported | Reads `CLAUDE.md` by default; LinaPro already supports this through a symlink. Run `make agents.skills.link AGENT=claude-code` |
| OpenAI Codex | ✅ Supported | ✅ Supported | Natively matches LinaPro's current directory structure |
| Cursor | ✅ Supported | ✅ Supported | Put simple conventions in `AGENTS.md` and more complex rules in `.cursor/rules/` |
| OpenCode | ✅ Supported | ✅ Supported | Supports both `AGENTS.md` and `Claude Code` compatible files |
| Windsurf | `.windsurf/skills/` | ✅ Supported | Run `make agents.skills.link AGENT=windsurf` |
| Cline | ✅ Supported | ✅ Supported | Put shared guidance in `AGENTS.md` and `Cline`-specific rules in `.clinerules/` |
| GitHub Copilot | ✅ Supported | ✅ Supported | Also supports `.github/copilot-instructions.md` configuration |

## Complete Tool Matrix

> If anything is missing or inaccurate, please submit a `PR` or `Issue` to help update it.

| Tool | Skills Directory | Project Instructions | Integration Guidance |
|----------------------|----------|--------------|------------------|
| Amp | ✅ Supported | ✅ Supported | Reuse the current `AGENTS.md` and standard skills directory directly |
| Antigravity | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| Claude Code | ✅ Supported | ✅ Supported | Reads `CLAUDE.md` by default; LinaPro already supports this through a symlink. Run `make agents.skills.link AGENT=claude-code` |
| OpenClaw | `skills/` | ⚠️ To be confirmed | Conflicts with the repo-root `skills/` path; explicitly opt in via `make agents.skills.link AGENT=openclaw FORCE=1`. Validate instruction file loading behavior |
| Cline | ✅ Supported | ✅ Supported | Put shared guidance in `AGENTS.md` and `Cline`-specific rules in `.clinerules/` |
| CodeBuddy | `.codebuddy/skills/` | ✅ Supported | Prefers `CODEBUDDY.md` and falls back to `AGENTS.md` automatically when `CODEBUDDY.md` is absent (per Tencent's official docs); run `make agents.skills.link AGENT=codebuddy` to wire the skills directory — no `md` symlink required |
| OpenAI Codex | ✅ Supported | ✅ Supported | Natively compatible with the current `LinaPro` conventions |
| Cursor | ✅ Supported | ✅ Supported | Put simple rules in `AGENTS.md` and scoped rules in `.cursor/rules/` |
| Deep Agents | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| Devin | `.devin/skills/` | ✅ Supported | Run `make agents.skills.link AGENT=devin` |
| Dexto | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| Droid | `.factory/skills/` | ✅ Supported | Run `make agents.skills.link AGENT=droid` |
| Firebender | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| ForgeCode | `.forge/skills/` | ⚠️ To be confirmed | Run `make agents.skills.link AGENT=forgecode`; validate instruction file loading behavior |
| Gemini CLI | ✅ Supported | ⚙️ Configurable | Run `make agents.md.link AGENT=gemini-cli` to link `GEMINI.md` → `AGENTS.md` |
| GitHub Copilot | ✅ Supported | ✅ Supported | Also supports `.github/copilot-instructions.md` configuration |
| Goose | `.goose/skills/` | ✅ Supported | Run `make agents.skills.link AGENT=goose` |
| Hermes Agent | `.hermes/skills/` | ⚠️ To be confirmed | Run `make agents.skills.link AGENT=hermes-agent`; validate instruction file loading behavior |
| Kimi Code CLI | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| OpenCode | ✅ Supported | ✅ Supported | Reads `AGENTS.md` natively and can also migrate `Claude Code` projects |
| Pi | `.pi/skills/` | ⚠️ To be confirmed | Run `make agents.skills.link AGENT=pi`; validate instruction file loading behavior |
| Qoder | `.qoder/skills/` | ✅ Supported | Run `make agents.skills.link AGENT=qoder` |
| Qwen Code | `.qwen/skills/` | ⚠️ To be confirmed | Run `make agents.skills.link AGENT=qwen-code`; for instruction file compatibility also run `make agents.md.link AGENT=qwen-code` |
| Replit | ✅ Supported | ⚠️ To be confirmed | Reuse the standard skills directory first; validate project instruction loading behavior |
| Rovo Dev | `.rovodev/skills/` | ⚠️ To be confirmed | Run `make agents.skills.link AGENT=rovodev`; validate instruction file loading behavior |
| Roo Code | `.roo/skills/` | ✅ Supported | Run `make agents.skills.link AGENT=roo` |
| Trae | `.trae/skills/` | ⚠️ To be confirmed | Run `make agents.skills.link AGENT=trae`; validate instruction file loading behavior |
| Warp | ✅ Supported | ✅ Supported | Reuse the standard skills directory first; validate project instruction loading behavior |
| Windsurf | `.windsurf/skills/` | ✅ Supported | Run `make agents.skills.link AGENT=windsurf` |
| Universal | ✅ Supported | Depends on the target tool | Use this for `AGENTS.md`-compatible tools that are not listed explicitly |

## Using LinaPro's Unified Management Commands

`LinaPro` ships a built-in `make agents` one-shot command that automatically wires up every applicable symlink for the chosen agent across skills/prompts/md, alongside a per-resource `make agents.<resource>.<action>` advanced command tree for batch flows. The commands operate strictly inside the repository root and never touch `HOME` directories or any system-global paths, and guarantee consistent behavior across `Windows`, `Linux` and `macOS`.

Supported resources:

- **`skills`**: directory bridge from `.<tool>/skills` to `.agents/skills`. Matches the tool compatibility matrix above.
- **`prompts`**: directory bridge from `.<tool>/.../opsx` to `.agents/prompts/opsx` (each agent declares its own source path) for managing slash commands / prompts catalogs.
- **`md`**: single-file bridge from `.<tool>.md` (e.g. `CLAUDE.md`, `GEMINI.md`) to the repo-root `AGENTS.md`, so agents that only read a private guide file name can reuse the same `AGENTS.md`.

```bash
# Recommended: agent-first one-shot command that touches every applicable resource

# Interactive mode (TTY only):
#   Step 1: arrow-key pick the agent (filter by typing).
#   Step 2: arrow-key pick `link` or `unlink`.
make agents

# One-shot mode (works in any environment, including CI):
make agents AGENT=claude-code                 # link claude-code across skills + md (prompts skipped per registry)
make agents AGENT=claude-code FORCE=1         # rebuild mismatched links during the same run
make agents AGENT=claude-code ACTION=unlink   # remove every managed symlink for claude-code

# Advanced: per-resource subcommands (used when the aggregate command intentionally
# does not support a flow, in particular AGENT=all and comma-separated lists)

# skills: directory bridge for the tool compatibility matrix above
make agents.skills.link                            # interactive selection on a TTY; read-only listing on CI/pipes
make agents.skills.link AGENT=claude-code          # non-interactive: link a single agent
make agents.skills.link AGENT=claude-code,codebuddy,qoder
make agents.skills.link AGENT=all                  # link every link-class agent
make agents.skills.link AGENT=all FORCE=1          # rebuild mismatched symlinks
make agents.skills.unlink AGENT=claude-code        # remove managed symlinks
make agents.skills.unlink AGENT=all

# prompts: agents' slash commands / prompts catalogs (initial coverage: claude-code / cursor / codex / gemini-cli)
make agents.prompts.link AGENT=claude-code
make agents.prompts.unlink AGENT=claude-code

# md: let agents read AGENTS.md via their private guide file name
make agents.md.link AGENT=claude-code              # link CLAUDE.md -> AGENTS.md
make agents.md.link AGENT=all                      # link every link-class agent's private guide file at once
make agents.md.unlink AGENT=claude-code
```

**Agent categories** (aligned with the [vercel-labs/skills](https://github.com/vercel-labs/skills#supported-agents) project-path table):

- `native` — the project path is already `.agents/skills` (e.g. `cursor`, `gemini-cli`, `codex`, `amp`, `opencode`, `cline`, `github-copilot`). No symlink is needed.
- `link` — the project path lives under a tool-specific directory (e.g. `claude-code` → `.claude/skills`, `codebuddy` → `.codebuddy/skills`, `windsurf` → `.windsurf/skills`). A relative symlink to `.agents/skills` is created on demand.
- `rootCollision` — the project path is `skills/` at the repository root (currently only `openclaw`). Skipped by default; pass `AGENT=openclaw FORCE=1` to opt in.

**Conflict protection rules**:

- The commands never auto-remove existing real directories or files — even with `FORCE=1`.
- `FORCE=1` only rebuilds symlinks that already exist but point at a non-managed target.
- All per-tool symlink directories are listed in `.gitignore`, so creating them locally never pollutes the repository.

**Aggregate command safety guards**:

- `make agents` accepts a single agent name only. `AGENT=all` and comma-separated lists are explicitly rejected (use the per-resource subcommands above for batch flows).
- Without `AGENT`, non-TTY invocations print a usage hint instead of blocking on input.
- Resources where the chosen agent is `native` or unregistered are skipped automatically with an explicit reason in the final summary.

All interactive entry points (the aggregate command and every per-resource subcommand) are powered by [charmbracelet/huh](https://github.com/charmbracelet/huh): use the **arrow keys** to navigate, **space** to toggle multi-select rows, **enter** to confirm, **type** to filter, and **Esc** / **Ctrl+C** to cancel. Option labels follow two different conventions depending on the prompt:

- The aggregate `make agents` command's "pick an agent" step is a single-select across the cross-resource registry. Each option embeds the agent name plus a **resource roles summary with runtime status glyphs** (e.g. `claude-code (Claude Code) — skills: link[+], prompts: link[!], md: link[+]`) so you can see at a glance which resource types the chosen agent will touch and whether each one is already linked. Link-class resources show the runtime glyph; `native` resources are listed without a glyph (no work needed); unregistered resources are omitted.
- Per-resource subcommands `make agents.<resource>.<action>` operate within a single resource and embed a **single-character status glyph** plus a short descriptor (e.g. `[~] claude-code  (mismatch)`) so you can see each candidate's current binding state without leaving the prompt.

Status glyphs embedded in per-resource option labels:

- `[+]` linked — already pointing at `.agents/skills`
- `[~]` mismatch — symlink exists but targets another location
- `[.]` absent — no symlink yet
- `[!]` conflict — a real directory or file blocks linking
- `[*]` root-collision — agent uses the repo-root `skills/` path
- `[?]` error — inspection failed; see the non-interactive listing for details

See [the `linactl` README](https://github.com/linaproai/linapro/blob/main/hack/tools/linactl/README.md) for additional details.
