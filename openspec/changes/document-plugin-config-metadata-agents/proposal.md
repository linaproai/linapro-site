## Why

当前中文官网对插件配置、插件`manifest`资源和`AI Coding`项目规范管理的介绍仍偏分散：源码插件页面已有简短说明，但没有完整解释配置读取优先级、动态插件`artifact`快照、`manifest`资源访问边界和插件本地`AGENTS.md`优先级。用户需要一组基于最新`linapro`源码事实的中文文档，帮助插件开发者正确理解和使用这些能力。

## What Changes

- 新增中文插件文档，系统介绍插件独立配置文件的定义、读取顺序、生产覆盖路径、开发期默认路径、动态插件`artifact`默认配置和推荐使用方式。
- 在同一插件文档中补充插件`manifest`原始资源的读取模型、路径规则、专用管线边界和个性化设计场景。
- 更新源码插件、动态插件或插件系统相关中文文档，建立到新增文档的清晰入口，避免重复堆叠。
- 扩写开发工具中文文档中的`AI工具集成`说明，介绍`AGENTS.md`、`.agents/rules/*.md`、`.agents/skills`、`.agents/prompts`和`make agents`的协作关系。
- 在`AI原生设计`中文文档中补充渐进式项目规范管理策略，说明全局规范、规则文件和插件本地规范的优先级。
- 本次只更新或新增中文文档，不修改`apps/lina-site/i18n/`下英文或其他语言内容。

## Capabilities

### New Capabilities

- `docs-plugin-config-metadata-agents`: 定义中文官网对插件配置、插件`manifest`资源和`AI Coding`规范管理策略的文档覆盖要求、源码事实口径和`i18n`排除规则。

### Modified Capabilities

## Impact

- 影响`apps/lina-site/docs/`下中文主稿的插件系统、开发工具和`AI`原生设计相关文档。
- 不影响站点运行时代码、主框架源码、插件源码、`API`契约、构建依赖或数据库结构。
- 不修改`apps/lina-site/i18n/`，英文内容由中文定稿后再单独同步。
