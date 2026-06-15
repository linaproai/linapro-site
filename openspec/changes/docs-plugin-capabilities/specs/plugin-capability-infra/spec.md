## ADDED Requirements

### Requirement: Infra文档说明基础设施状态能力
文档`8000-infra.md` SHALL 说明`Infra()`提供基础设施组件状态视图，源码插件通过`services.Infra()`读取状态，动态插件通过`service: infra`和`pluginbridge.Services.Infra()`读取同一类状态视图。

#### Scenario: Infra能力可查阅
- **WHEN** 插件开发者阅读基础设施能力文档
- **THEN** 理解`StatusProjection`、`BatchGetStatus`、`service: infra`和动态`pluginbridge.Services.Infra()`的边界

### Requirement: Infra文档说明可信刷新命令
文档 SHALL 说明`Admin().Infra().RefreshStatus`属于可信源码插件管理命令，不通过动态插件普通`hostServices`开放。

#### Scenario: 刷新命令边界清晰
- **WHEN** 插件开发者查看`RefreshStatus`
- **THEN** 文档说明该命令只面向可信源码插件，并可能触发宿主状态重算

### Requirement: Infra文档不得混入Runtime运行时原语
文档 SHALL 明确`Runtime()`不属于`Infra()`领域能力；日志、插件状态、宿主时间、`UUID`和节点身份读取应链接到独立`Runtime`文档。

#### Scenario: Infra与Runtime边界清晰
- **WHEN** 动态插件开发者需要判断组件可用性
- **THEN** 文档引导其声明`service: infra`
- **AND** 不把`service: runtime`描述为基础设施能力入口
