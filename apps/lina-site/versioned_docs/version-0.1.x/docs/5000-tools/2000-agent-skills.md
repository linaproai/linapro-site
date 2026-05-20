---
slug: '/docs/agent-skills'
title: '开发技能'
hide_title: true
description: '本文介绍 LinaPro 的 AI 专属技能体系，包含可选安装的 OpenSpec CLI、goframe-v2 和 find-skills 三项外部工具或技能，以及项目内置的前端设计、Git 工作流、代码审查、E2E 测试、性能审计、OpenSpec 实施归档等自动加载技能，并说明源码插件升级等场景可能依赖 AI 工具侧提供的 lina-upgrade 技能，帮助开发者全面了解各技能的作用、安装方法和使用方式。'
keywords:
  - AI技能
  - 开发技能
  - LinaPro技能
  - Claude Code技能
  - OpenSpec
  - goframe-v2
  - find-skills
  - 技能安装
  - 前端设计技能
  - E2E测试技能
  - 性能审计技能
  - 代码审查技能
  - Git提交技能
  - Playwright技能
  - AI协作
  - 技能体系
  - 框架技能
---

`LinaPro`的`AI`技能体系由两类技能组成：

- **外部安装技能**：需通过`npm`/`npx`命令手动安装，安装后在任意项目中全局生效，例如`OpenSpec`、`goframe-v2`和`find-skills`。
- **项目内置技能**：随项目源码一起存放于`.agents/skills/`目录，无需额外安装，使用例如`Claude Code`等`AI`工具在项目目录中工作时会自动加载。

## 技能概览

| 技能名称 | 适用场景 | 必要依赖 |
|---------|---------|---------|
| `openspec` | 可选的规范驱动工作流工具，推荐配合使用 | 建议安装 |
| `goframe-v2` | `GoFrame`专属`AI`技能，提升后端代码生成质量 | 建议安装 |
| `find-skills` | `AI`技能市场搜索工具，辅助技能选型 | 建议安装 |
| `frontend-design` | 创作高品质、有辨识度的前端界面 | 无 |
| `frontend-patterns` | `React`/前端开发最佳实践与模式 | 无 |
| `git-commit-push` | 生成规范提交信息并推送到远端 | 无 |
| `git-worktree` | 创建独立`Git`工作树隔离开发任务 | 无 |
| `karpathy-guidelines` | 减少`AI`编码错误的行为守则 | 无 |
| `lina-archive-consolidate` | 聚合`OpenSpec`归档迭代为统一目录（通常用于CI自动化） | 无 |
| `lina-auto-archive` | 自动扫描并归档已完成的`OpenSpec`变更（通常用于CI自动化） | 需安装`OpenSpec CLI` |
| `lina-e2e` | `E2E`测试用例命名规范与管理标准 | 需安装`Playwright` |
| `lina-feedback` | 追踪并修复`OpenSpec`变更后报告的问题 | 无 |
| `lina-perf-audit` | 全面的后端`API`性能审计（手动触发） | 需运行中的`LinaPro`服务 |
| `lina-review` | 代码与规范的结构化审查 | 无 |
| `openspec-explore` | 需求探索与问题分析的思维伙伴模式 | 无 |
| `openspec-propose` | 一步生成完整`OpenSpec`变更提案 | 无 |
| `openspec-apply-change` | 按`OpenSpec`任务清单实施变更 | 需安装`OpenSpec CLI` |
| `openspec-archive-change` | 归档完成的`OpenSpec`变更 | 需安装`OpenSpec CLI` |
| `playwright-cli` | 浏览器自动化交互与`Playwright`测试 | 需安装`Playwright` |


## 如何使用技能

**外部安装技能**：按各技能说明中的命令完成全局安装后，`Claude Code`在对应场景下会自动识别并应用。

**项目内置技能**：`AI`工具在项目目录中运行时会自动加载`.agents/skills/`下的所有技能，无需额外操作。使用技能时，直接在对话中说明场景即可触发，或使用斜杠命令（如`/git-commit-push`）显式调用。

## 外部安装技能

以下三项技能需要手动安装，安装完成后在所有项目中全局生效。

### openspec

`OpenSpec`是`LinaPro`规范驱动工作流的推荐命令行工具，安装后`/opsx:explore`、`/opsx:propose`、`/opsx:apply`、`/opsx:archive`等工作流技能将自动使用它作为底层引擎。这是`LinaPro AI`研发工作流中**建议安装**的技能。

**适用场景**：执行`OpenSpec`探索、提案、实施和归档等完整工作流时。

**安装方法**：

```bash
npm install -g @fission-ai/openspec@latest
```

安装完成后，可通过`opsx --version`确认安装成功。

### goframe-v2

`goframe-v2`是专为`LinaPro`后端`Go`代码定制的`Claude Code`技能，内置`GoFrame`编码规范、`ORM`使用模式和最佳实践示例。编写或修改后端`Go`文件、实现服务接口、创建`API`或进行数据库操作时，该技能会自动激活，提供代码生成、错误诊断和性能优化建议，大幅提升与框架规范的一致性。

**适用场景**：开发`lina-core`后端代码、实现业务接口、编写数据访问层时。

**安装方法**：

```bash
npx skills add github.com/gogf/skills -g
```

### find-skills

`find-skills`是`AI`技能市场的搜索工具，帮助开发者快速在技能市场中查找和评估适合当前项目的`AI`技能。在需要引入新的`AI`辅助能力、不确定是否存在合适的现成技能时，通过`find-skills`可以显著提升技能选型效率。

**适用场景**：需要为项目引入新`AI`技能、评估现有技能生态时。

**安装方法**：

```bash
npx skills add vercel-labs/skills --skill find-skills -g
```

## 项目内置技能

以下技能随项目源码一起维护于`.agents/skills/`目录，无需安装，在项目目录中使用`AI`工具时自动加载。

### frontend-design

高品质前端界面设计指导。帮助`AI`避免产出千篇一律的"通用感"界面，创作出有辨识度、有设计质感的前端页面。

技能内容涵盖：设计意图分析（页面目的、调性、差异化方向）、字体与配色选择、动效与空间构图、背景与视觉细节处理，以及界面概念方向的确立与精确执行。

**适用场景**：开发新页面、重新设计现有界面、需要前端视觉风格指导时。

**使用方式**：在对话中描述界面需求，`AI`会自动应用设计指导；或使用`/frontend-design`命令显式触发。

### frontend-patterns

`React`及现代前端开发的最佳实践与设计模式。为`AI`提供组件组合、自定义`Hook`、状态管理、性能优化、表单处理、错误边界和动效等方面的规范参考。

**适用场景**：开发前端功能组件、需要状态管理方案、实现性能优化或可访问性改进时。

**使用方式**：在前端代码开发场景中自动激活，或使用`/frontend-patterns`命令显式触发。

### git-commit-push

审查当前`Git`工作区变更，按仓库提交规范自动生成提交信息，将所有改动提交到当前分支并推送到`origin`。

提交信息遵循约定式提交格式，支持`fix`、`feat`、`build`、`docs`等类型前缀，自动从`diff`提炼提交摘要。

**适用场景**：完成功能开发或修复后需要提交并推送代码时。

**使用方式**：告知`AI`"提交代码"或"commit and push"，或使用`/git-commit-push`命令触发。

### git-worktree

创建独立的`Git`工作树，让`AI`在新的隔离目录中继续工作，而不影响当前检出的分支状态。

对于并行处理多个功能分支、需要在不干扰主工作区的情况下探索新方案时特别有用。

**适用场景**：需要隔离任务开发环境、并行处理多个分支时。

**使用方式**：使用`/git-worktree`命令触发，`AI`会自动派生任务标识作为工作树分支名。

### karpathy-guidelines

一套减少`AI`编码常见错误的行为守则，灵感来源于`Andrej Karpathy`对`LLM`编码行为的总结。

四条核心原则：先思考再编码（显式声明假设和权衡）、简单优先（最小化代码解决问题）、外科手术式修改（只改必要的地方）、目标驱动执行（定义可验证的完成标准并循环直到通过）。

**适用场景**：所有编码任务，作为`AI`辅助开发的基础行为约束。

**使用方式**：作为基础守则自动激活，通常无需显式触发。

### lina-archive-consolidate

将`openspec/changes/archive/`下已归档的多轮`OpenSpec`迭代，按功能职责聚合整理为统一的归档目录。

执行语义合并时，保留所有迭代的设计决策和实现信息，将多次迭代压缩为单份内聚的归档，降低历史目录的噪音。归档条目达到`8`条及以上时，会自动进入完整的`OpenSpec`工作流（探索→提案→实施）。

**适用场景**：归档目录积累了大量零散迭代记录，需要定期整理合并时。

**使用方式**：使用`/lina-archive-consolidate`命令触发。

### lina-auto-archive

自动扫描`openspec/changes/`目录，将所有已完成的活跃变更批量归档，并汇总处理结果。

技能会严格检查每个变更的完成状态——`OpenSpec`状态完成且任务全部完成才允许归档，任何未完成项一律跳过，不会强制放行，也不会绕过校验。归档完成后以中文汇报成功归档列表和未归档变更的原因。

**适用场景**：需要批量清理已完成变更、定期整理活跃变更目录时。

**必要依赖**：需要安装`OpenSpec CLI`（`npm install -g @fission-ai/openspec@latest`）。

**使用方式**：使用`/lina-auto-archive`命令触发。

### lina-e2e

`Playwright E2E`测试用例的命名规范、目录结构和管理标准。为`AI`提供测试用例`ID`分配规则（`TC{NNNN}`格式，全局唯一且不复用）、按模块组织的目录布局、测试文件模板、独立性要求和页面对象模型使用规范。

**适用场景**：编写、维护或审查`E2E`测试用例时。

**必要依赖**：需要安装`Playwright`及测试依赖：

```bash
cd hack/tests
pnpm install
npx playwright install --with-deps chromium
```

**使用方式**：`AI`在编写`E2E`测试时自动应用本技能规范；或使用`/lina-e2e`命令触发。

### lina-feedback

结构化的问题追踪与修复工作流，用于处理`OpenSpec`变更实施后报告的`bug`、缺失功能、体验问题和测试漏洞。

工作流覆盖：确定目标变更、按类型分析问题（`bug`/缺失/体验/测试漏洞）、更新增量规范、组织任务列表、循环执行修复、回归分析，以及完成标记前的全面验证。

**适用场景**：实施`OpenSpec`变更后收到反馈报告，需要系统性处理和修复时。

**使用方式**：使用`/lina-feedback`命令触发，并告知`AI`需要处理的反馈内容。

### lina-perf-audit

全面的`LinaPro`后端`API`性能审计工作流（**仅手动触发**）。审计内容包括`N+1`查询、缺少索引、无限制列表响应、重复读取、可合并`SQL`调用、循环中的阻塞操作，以及读接口中执行写`SQL`的问题。

审计流程分三阶段：阶段`0`（准备环境：重置数据库、重启服务、安装插件、植入压测数据）、阶段`1`（并发子代理审计各模块）、阶段`2`（汇总生成报告和持久化问题卡片）。

:::caution 注意

此技能会重置数据库、重启服务，耗时较长（几十分钟到数小时），消耗大量`Token`，仅在明确需要全面性能审计时使用。

:::

**适用场景**：显式需要对所有后端`API`进行系统性性能审计时。

**必要依赖**：需要运行中的`LinaPro`开发环境（通过`make dev`启动）。

**使用方式**：在对话中明确要求"运行`lina-perf-audit`"或"对所有后端`API`进行完整性能审计"来触发。

### lina-review

代码与规范的结构化审查工作流，在`OpenSpec`实施或反馈任务完成后、归档之前触发，确保代码质量和规范一致性。

审查范围：后端代码审查（`GoFrame`规范、接口`i18n`合规、分布式缓存一致性）、`RESTful API`审查、`SQL`审查、`E2E`测试审查，以及最终报告生成。以`AGENTS.md`为所有审查标准的唯一依据。

**适用场景**：`OpenSpec`变更完成实施后、准备归档前的质量把关环节。

**使用方式**：使用`/lina-review`命令触发，`AI`会根据变更范围自动确定审查对象。

### openspec-explore

探索模式——需求分析和问题调查的`AI`思维伙伴。采用好奇、非预设的方式帮助理清问题空间、调查代码库、对比方案、可视化架构，并识别风险与未知因素。

探索模式侧重于**思考和调查**而非实施：读取文件、探查代码库、用`ASCII`图表可视化，可以生成`OpenSpec`提案等思考产物，但不直接实现业务功能。

**适用场景**：开始一项复杂任务前的调研阶段、需要厘清需求或技术方案时。

**使用方式**：使用`/openspec-explore`命令触发，描述你想探索的问题或方向。

### openspec-propose

一步生成完整`OpenSpec`变更提案，自动创建包含`proposal.md`、`design.md`、`tasks.md`的变更目录及`specs/`规范文件。

`AI`会按依赖顺序生成各产物，确保设计文档和任务列表之间的信息一致性，无需手工拼装整套变更文件。

**适用场景**：需要为一项新功能或改动创建完整的`OpenSpec`变更提案时。

**使用方式**：使用`/openspec-propose`命令触发，描述你想实现的变更内容。

### openspec-apply-change

按照`OpenSpec`变更目录中的`tasks.md`逐项实施代码、文档和测试改动。该技能通常在提案和设计已经完成后触发，负责把变更从规范落到可运行实现。

**适用场景**：已经存在`openspec/changes/<change-id>/`，需要开始或继续实施任务时。

**必要依赖**：建议安装`OpenSpec CLI`。

**使用方式**：使用`/openspec-apply-change`命令触发，或直接告诉`AI`继续实施当前`OpenSpec`变更。

### openspec-archive-change

归档已经完成的`OpenSpec`变更，将增量规范合并到基线规范，并把变更目录移动到归档区。

**适用场景**：变更实现、审查和验证全部完成，需要沉淀为长期基线时。

**必要依赖**：建议安装`OpenSpec CLI`。

**使用方式**：使用`/openspec-archive-change`命令触发，或让`AI`归档当前完成的变更。

### playwright-cli

`Playwright`浏览器自动化`CLI`指令集，覆盖打开页面、点击、输入、表单提交、多标签页管理、`Cookie`/`localStorage`/`sessionStorage`操作、网络路由拦截和`DevTools`等完整操作集。

**适用场景**：自动化浏览器交互、编写或调试`Playwright`测试时。

**必要依赖**：需要安装`Playwright`及相关浏览器：

```bash
cd hack/tests
pnpm install
npx playwright install --with-deps chromium
```

**使用方式**：描述需要执行的浏览器操作，或使用`/playwright-cli`命令触发。
