## Context

现有中文官网已经有插件系统、源码插件、动态插件、开发指令和`AI`原生设计文档。其中源码插件文档包含“插件配置和清单资源”小节，但内容较短，只列出`Config()`、`HostConfig()`和`Manifest()`三个服务，没有系统解释配置来源优先级、动态插件发布快照、`manifest`资源边界，也没有和`AI Coding`规范管理策略形成完整说明。

本次文档事实依据来自`/Users/john/Workspace/github/linaproai/linapro`最新源码：

- 插件配置服务位于`apps/lina-core/pkg/plugin/capability/config/`，只读取当前插件作用域内的`config.yaml`，不会回退到宿主完整`GoFrame`配置树。
- 插件配置读取顺序为生产配置、开发期配置、动态插件`artifact`默认配置。生产配置路径为生产配置根下`plugins/<plugin-id>/config.yaml`；生产配置根优先来自显式注入，其次来自`GF_GCFG_PATH`，再回退到仓库开发态`apps/lina-core/manifest/config`或二进制目录`config`。
- 开发期配置路径为仓库根下`apps/lina-plugins/<plugin-id>/manifest/config/config.yaml`。
- 动态插件构建产物会携带`manifest/config/config.yaml`，运行时在没有生产和开发期配置文件时可作为当前有效发布版本的默认配置；`manifest/config/config.example.yaml`只是模板，不作为运行时默认值。
- 插件`Manifest()`服务读取的是当前插件`manifest/`下的原始资源，调用路径相对`manifest/`，例如`profile.yaml`、`resources/policy.yaml`、`config/config.example.yaml`、`sql/*.sql`或`i18n/*.json`。`config/`、`sql/`、`i18n/`可以作为原文读取，但运行配置、`SQL`执行和语言包加载仍由各自专用管线负责。
- 根目录`AGENTS.md`是`AI Coding`项目规范入口，`.agents/rules/*.md`承载渐进披露的领域细则，插件目录`apps/lina-plugins/<plugin-id>/AGENTS.md`在该插件目录内优先于全局规范。
- `make agents`通过`linactl agents`管理`.agents/skills`、`.agents/prompts`和`AGENTS.md`到不同`AI Coding`工具私有路径的软链。

用户明确要求本次只更新或创建中文官网文档，不同步英文文档，待中文内容审阅后再翻译。

## Goals / Non-Goals

**Goals:**

- 新增一篇插件公共文档，完整介绍插件独立配置和`manifest`资源的定义、读取、使用方式与边界。
- 在源码插件和动态插件文档中保留必要入口，把详细说明收敛到新增公共文档，减少重复。
- 扩写开发指令文档中的`make agents`说明，让用户理解命令背后的规范资源桥接模型。
- 在`AI原生设计`文档补充渐进式项目规范管理策略，以及插件本地`AGENTS.md`优先级。
- 所有新增中文内容必须符合`apps/lina-site/docs/`同级文档结构和 Markdown 格式规范。

**Non-Goals:**

- 不修改`/Users/john/Workspace/github/linaproai/linapro`主仓库源码。
- 不修改`apps/lina-site/i18n/`英文或其他语言内容。
- 不新增站点运行时代码、Docusaurus 插件、路由重定向或构建依赖。
- 不把源码级内部实现细节完整暴露为开发者教程，只抽取足够解释使用方式和设计边界的事实。

## Decisions

### 决策一：新增插件公共文档承载配置与manifest资源

**Decision**

在`apps/lina-site/docs/docs/2000-components/5000-plugin-system/`下新增`9000-plugin-config-and-metadata.md`，标题为“插件配置与manifest资源”。该文档作为源码插件和动态插件共享的公共说明，覆盖：

- 插件配置文件职责和目录位置。
- 生产、开发期、动态`artifact`默认配置的读取优先级。
- `Config()`、`HostConfig()`、`Manifest()`的职责边界。
- `Manifest()`读取插件`manifest/`原始资源的路径语义、来源和授权边界。
- 使用配置和`manifest`资源实现插件个性化行为的示例。
- 常见误区，例如不要把插件业务配置写入宿主`config.yaml`，不要把`Manifest()`读取原文理解为配置、`SQL`或语言包已经运行时生效。

**Rationale**

配置和`manifest`资源不是源码插件专属能力。动态插件构建和运行时也保留同类路径语义，把内容放在源码插件文档会让动态插件开发者漏读关键规则。

**Alternatives considered**

- 只扩写源码插件文档：修改小，但会让动态插件路径语义和`artifact`快照说明位置不自然。
- 分别新建配置文档和`manifest`资源文档：颗粒度更细，但当前内容规模不需要拆成两篇，合并更方便开发者一次理解资源模型。

### 决策二：AI规范管理放在开发工具和AI原生设计中

**Decision**

不把`AI Coding`项目规范管理策略塞进插件系统文档。具体落点为：

- 在`docs/5000-tools/1000-commands.md`的`AI工具集成`章节中扩写`make agents`和三类资源桥接：`skills`、`prompts`、`md`。
- 在`docs/1000-concepts/1000-ai-native.md`中新增或扩写“渐进式项目规范管理”小节，说明`AGENTS.md`、`.agents/rules/*.md`、插件本地`AGENTS.md`和`make agents`之间的关系。

**Rationale**

`AI Coding`规范管理是`LinaPro`的`AI`原生工程能力，插件本地规范只是其中一个应用场景。放到开发工具和`AI`原生设计文档更符合用户查找路径。

**Alternatives considered**

- 新建独立`AI规范管理`页面：可扩展性更好，但当前已有`AI原生设计`和`开发指令`页面，先扩写现有文档更克制。
- 放入插件文档：能覆盖插件本地`AGENTS.md`，但会削弱全局规范和多工具兼容的整体视角。

### 决策三：本次明确不更新英文文档

**Decision**

本次实现只修改`apps/lina-site/docs/`中文主稿，不修改`apps/lina-site/i18n/en/`或其他`i18n`目录。

**Rationale**

用户明确要求先审阅中文内容，再进行英文翻译。插件配置和`AI`规范管理涉及较多源码事实，先保证中文准确性比同步半成品英文更重要。

**Alternatives considered**

- 同步英文镜像：结构一致，但违反用户要求，也可能产生不自然翻译。

## Risks / Trade-offs

- [中文与英文短期不一致] → 在最终说明中明确本次未同步英文，后续由中文定稿后再处理。
- [插件配置运行时路径写错] → 文档以源码读取顺序为准，写清“生产配置根下`plugins/<plugin-id>/config.yaml`”，避免简化成仓库根路径。
- [`manifest`资源路径边界混淆] → 示例使用`services.Manifest().Scan(ctx, "profile.yaml", "", &target)`和`services.Manifest().Get(ctx, "config/config.example.yaml")`，并明确调用路径相对`manifest/`，文件名没有框架级特殊语义。
- [文档重复] → 源码插件和动态插件页面只保留摘要和跳转，详细规则集中在新增公共文档。

## Migration Plan

1. 新增插件配置与`manifest`资源中文文档，按同级插件系统文档风格补齐`front matter`、`description`和`keywords`。
2. 更新源码插件文档中“插件配置和清单资源”小节，替换为简短摘要、关键示例和新文档链接。
3. 更新动态插件文档中目录结构或资源打包说明，补充到新文档的引用。
4. 扩写开发指令文档的`AI工具集成`章节，解释`make agents`聚合入口和资源软链策略。
5. 扩写`AI原生设计`文档，补充渐进式规范管理策略和插件本地规范优先级。
6. 运行`openspec validate`和站点构建或文档检查，确认文档结构有效。

回滚方式为删除新增文档并恢复被修改的中文 Markdown 文件；本次不涉及数据迁移或运行时代码变更。

## Open Questions

- 中文审阅后，英文翻译是否使用现有`linasite-sync-i18n`流程同步，还是由人工先翻译再让`AI`审校？本次不处理该问题。
