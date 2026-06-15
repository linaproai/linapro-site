## ADDED Requirements

### Requirement: Plugins文档说明插件治理父领域能力
文档`7500-plugins.md` SHALL 说明`Plugins()`作为插件治理父领域能力，聚合注册表投影、插件配置、插件启用状态、能力提供方状态和生命周期编排。

#### Scenario: 父领域结构清晰
- **WHEN** 插件开发者阅读基本介绍
- **THEN** 文档以表格说明`Registry()`、`Config()`、`State()`和`Lifecycle()`子能力，而不要求这些子能力各自单独成页

### Requirement: Plugins文档说明插件配置子能力
文档 SHALL 说明`Plugins().Config()`的读取优先级、主要方法和与`HostConfig()`的关系。

#### Scenario: 插件配置可查阅
- **WHEN** 插件开发者查看`Config`子能力
- **THEN** 文档列出生产配置、开发期`manifest/config/config.yaml`和动态插件产物配置的优先级，并列出`Get`、`Exists`、`Scan`、`String`、`Bool`、`Int`和`Duration`

### Requirement: Plugins文档说明插件状态子能力
文档 SHALL 说明`Plugins().State()`中`IsEnabled`、`IsEnabledAuthoritative`和`IsProviderEnabled`的语义差异。

#### Scenario: 状态策略可理解
- **WHEN** 插件开发者阅读`State`子能力
- **THEN** 理解本地快照适合高频判断，权威读取适合全局控制，提供方启用状态独立于租户业务入口可见性

### Requirement: Plugins文档说明生命周期子能力
文档 SHALL 说明`Plugins().Lifecycle()`在租户级插件禁用和租户删除中的前置检查与后置通知角色，并区分它与`pluginhost.SourcePlugin.Lifecycle()`。

#### Scenario: 生命周期职责清晰
- **WHEN** 插件开发者阅读`Lifecycle`子能力
- **THEN** 理解`Plugins().Lifecycle()`面向治理模块编排跨插件事件，`SourcePlugin.Lifecycle()`面向单个插件注册自身回调
