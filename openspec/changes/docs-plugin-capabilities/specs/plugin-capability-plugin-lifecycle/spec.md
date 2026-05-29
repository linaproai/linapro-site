## ADDED Requirements

### Requirement: PluginLifecycle文档说明生命周期编排设计
文档`7500-plugin-lifecycle.md` SHALL 说明`PluginLifecycleService`在租户级插件治理中的角色：租户禁用插件和租户删除的前置检查与后置通知。

#### Scenario: 编排设计可理解
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解该服务用于编排跨插件的租户级生命周期事件，而非单个插件的生命周期回调

### Requirement: PluginLifecycle文档说明与pluginhost.Lifecycle的区别
文档 SHALL 明确区分`PluginLifecycleService`（宿主编排）和`pluginhost.Lifecycle()`（插件自身回调）的不同职责。

#### Scenario: 职责区分清晰
- **WHEN** 插件开发者阅读设计约束
- **THEN** 理解`PluginLifecycleService`面向治理模块消费，`pluginhost.Lifecycle()`面向单个插件注册回调

### Requirement: PluginLifecycle文档包含主要能力概览
文档 SHALL 以表格形式简要列出四个方法的职责。

#### Scenario: 方法概览可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能快速了解每个方法的用途和调用时机
