---
slug: '/quick/agent-tools'
title: 'AI工具集成'
hide_title: true
description: '本文介绍当前主流 AI Coding 工具对 LinaPro 框架中 .agents/skills/ 技能目录、AGENTS.md 项目规范文件以及 CLAUDE.md 等工具专属规范文件的支持情况，并提供 LinaPro 统一的 make skills 软链管理命令，帮助团队判断 Claude Code、OpenAI Codex、Cursor、GitHub Copilot、Gemini CLI、Windsurf、Amp、OpenCode、Devin、Qoder、CodeBuddy、Roo Code 等工具如何快速接入 LinaPro 的 AI 原生研发规范。'
keywords:
  - LinaPro
  - AI工具集成
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
  - AI Coding工具
  - 项目规范
  - 工具兼容性
  - AI原生开发
---

## 背景

`LinaPro`框架内置了完整的`AI`原生研发支撑体系，其中两个核心机制是：

- **`.agents/skills/`技能目录**：存放结构化的`Agent Skills`，遵循开放的[Agent Skills 规范](https://agentskills.io/)，最常用的技能目录。
- **`AGENTS.md`规范文件**：项目级`AI`智能体规范文档，集中描述代码库架构、开发流程、技术约束和验证要求，为接入的`AI Coding`工具提供持续的项目上下文。

当今的`AI Coding`工具百花齐放，各个工具对技能目录和项目规范文件的读取路径并不完全一致，并无统一规范，本文用于梳理和澄清，帮助大家更好地对接常用的开发工具。

## 状态说明

| 标记 | 含义 |
|------|------|
| ✅ 支持 | 工具官方文档明确说明支持`LinaPro`默认提供的技能目录和规范文件 |
| ⚙️ 可配置 | 默认不是该文件名，但可通过配置、导入或软连接复用 |
| ⚠️ 待确认 | 仅确认`skills CLI`支持安装技能，未确认项目规范文件读取行为 |

## 主流工具推荐

> 以下是社区团队成员在`AI Coding`实践中使用较多的工具，供大家参考。

| 工具名称 | 技能目录 | 项目规范 | 集成建议 |
|---------|---------|---------|---------|
| Claude Code | `.claude/skills/` | ✅ 支持 | 默认读取`CLAUDE.md`，已做软连支持 |
| OpenAI Codex | ✅ 支持 | ✅ 支持 | 原生匹配`LinaPro`当前目录结构 |
| Cursor | ✅ 支持 | ✅ 支持 | 简单规范放`AGENTS.md`，复杂规则放`.cursor/rules/` |
| OpenCode | ✅ 支持 | ✅ 支持 | 同时支持`AGENTS.md`与 `Claude Code` 兼容文件 |
| Windsurf | `.windsurf/skills/` | ✅ 支持 | 执行`make skills.link AGENT=windsurf` |
| Cline | ✅ 支持 | ✅ 支持 | 共享规范放`AGENTS.md`，`Cline` 专属规则放`.clinerules/` |
| GitHub Copilot | ✅ 支持 | ✅ 支持 | 也支持`.github/copilot-instructions.md`配置 |

## 完整工具矩阵

> 若有遗漏或错误，欢迎提交 `PR` 或 `Issue` 进行补充和修正。

| 工具名称 | 技能目录 | 项目规范 | 集成建议 |
|----------------------|----------|--------------|------------------|
| Amp | ✅ 支持 | ✅ 支持 | 直接复用当前`AGENTS.md`和标准技能目录 |
| Antigravity | ✅ 支持 | ⚠️ 待确认 | 可先复用标准技能目录，项目规范读取需实测 |
| Claude Code | ✅ 支持 | ✅ 支持 | 默认读取`CLAUDE.md`，已做软连支持 |
| OpenClaw | `skills/` | ⚠️ 待确认 | `make skills.link AGENT=openclaw FORCE=1`，规范文件读取需实测 |
| Cline | ✅ 支持 | ✅ 支持 | 共享规范放`AGENTS.md`，`Cline` 专属规则放`.clinerules/` |
| CodeBuddy | `.codebuddy/skills/` | ⚠️ 待确认 | 执行`make skills.link AGENT=codebuddy`，规范文件读取需实测 |
| OpenAI Codex | ✅ 支持 | ✅ 支持 | 原生兼容当前`LinaPro`规范 |
| Cursor | ✅ 支持 | ✅ 支持 | 简单规则用`AGENTS.md`，范围规则用`.cursor/rules/` |
| Deep Agents | ✅ 支持 | ⚠️ 待确认 | 可先复用标准技能目录，项目规范读取需实测 |
| Devin | `.devin/skills/` | ✅ 支持 | 执行`make skills.link AGENT=devin` |
| Dexto | ✅ 支持 | ⚠️ 待确认 | 可先复用标准技能目录，项目规范读取需实测 |
| Droid | `.factory/skills/` | ✅ 支持 | 执行`make skills.link AGENT=droid` |
| Firebender | ✅ 支持 | ⚠️ 待确认 | 可先复用标准技能目录，项目规范读取需实测 |
| ForgeCode | `.forge/skills/` | ⚠️ 待确认 | 执行`make skills.link AGENT=forgecode`，规范文件读取需实测 |
| Gemini CLI | ✅ 支持 | ⚙️ 可配置 | 建议软连`ln -s AGENTS.md GEMINI.md` |
| GitHub Copilot | ✅ 支持 | ✅ 支持 | 也支持`.github/copilot-instructions.md`配置  |
| Goose | `.goose/skills/` | ✅ 支持 | 执行`make skills.link AGENT=goose` |
| Hermes Agent | `.hermes/skills/` | ⚠️ 待确认 | 执行`make skills.link AGENT=hermes-agent`，规范文件读取需实测 |
| Kimi Code CLI | ✅ 支持 | ⚠️ 待确认 | 可先复用标准技能目录，项目规范读取需实测 |
| OpenCode | ✅ 支持 | ✅ 支持 | 原生读取`AGENTS.md`，也能迁移 `Claude Code` 项目 |
| Pi | `.pi/skills/` | ⚠️ 待确认 | 执行`make skills.link AGENT=pi`，规范文件读取需实测 |
| Qoder | `.qoder/skills/` | ✅ 支持 | 执行`make skills.link AGENT=qoder` |
| Qwen Code | `.qwen/skills/` | ⚠️ 待确认 | `make skills.link AGENT=qwen-code; ln -s AGENTS.md QWEN.md` |
| Replit | ✅ 支持 | ⚠️ 待确认 | 可先复用标准技能目录，项目规范读取需实测 |
| Rovo Dev | `.rovodev/skills/` | ⚠️ 待确认 | 执行`make skills.link AGENT=rovodev`，规范文件读取需实测 |
| Roo Code | `.roo/skills/` | ✅ 支持 | 执行`make skills.link AGENT=roo` |
| Trae | `.trae/skills/` | ⚠️ 待确认 | 执行`make skills.link AGENT=trae`，规范文件读取需实测 |
| Warp | ✅ 支持 | ✅ 支持 | 可先复用标准技能目录，项目规范读取需实测 |
| Windsurf | `.windsurf/skills/` | ✅ 支持 | 执行`make skills.link AGENT=windsurf` |

## make skills 管理命令

`LinaPro`提供了仓库内置的`make skills`命令体系，统一管理上述工具的`Agent`项目路径软链，避免开发者手写`ln -s`，并保证`Windows`/`Linux`/`macOS`三端一致行为。命令仅在仓库根目录范围内操作，不会修改`HOME`目录或任何系统全局路径。

```bash
# 终端下进入交互式操作菜单（link / unlink / quit）
make skills

# 终端下交互式选择要链接的 Agent；CI 或管道下显示只读状态列表
make skills.link

# 非交互式：为指定 Agent 创建软链
make skills.link AGENT=claude-code
make skills.link AGENT=claude-code,codebuddy,qoder
make skills.link AGENT=all                # 为所有 link 类 Agent 创建软链
make skills.link AGENT=all FORCE=1        # 强制重建指向错误源的旧软链

# 终端下交互式选择要移除的受管软链
make skills.unlink

# 非交互式：移除指定 Agent 的受管软链
make skills.unlink AGENT=claude-code
make skills.unlink AGENT=all
```
