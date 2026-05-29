## ADDED Requirements

### Requirement: PluginState文档说明启用状态查询的两种策略
文档`7600-plugin-state.md` SHALL 说明`IsEnabled`（本地快照）和`IsEnabledAuthoritative`（权威读取）的设计区别和适用场景。

#### Scenario: 两种策略可理解
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解本地快照适合高频判断（菜单、路由），权威读取适合全局控制（中间件、写保护）

### Requirement: PluginState文档说明Provider启用状态
文档 SHALL 说明`IsProviderEnabled`的独立语义：判断插件是否可作为框架能力提供方，与租户业务入口可见性无关。

#### Scenario: Provider状态语义清晰
- **WHEN** 插件开发者阅读设计约束
- **THEN** 理解`IsEnabled`控制业务入口，`IsProviderEnabled`控制能力供给

### Requirement: PluginState文档包含主要能力概览
文档 SHALL 以表格形式简要列出`IsEnabled`、`IsProviderEnabled`、`IsEnabledAuthoritative`三个方法的职责。

#### Scenario: 方法概览可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能快速了解每个方法的适用场景
