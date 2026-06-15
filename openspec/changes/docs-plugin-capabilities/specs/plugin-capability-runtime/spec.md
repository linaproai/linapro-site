## ADDED Requirements

### Requirement: Runtime文档说明动态插件专属运行时能力
文档`8050-runtime.md` SHALL 说明`Runtime()`是`pluginbridge.Services`上的动态插件专属能力，不属于源码插件`capability.Services`普通领域目录。

#### Scenario: Runtime专属边界可查阅
- **WHEN** 插件开发者阅读动态`Runtime`能力文档
- **THEN** 理解源码插件使用宿主原生日志、上下文和注入领域服务，不通过`Runtime()`包装

### Requirement: Runtime文档说明动态runtime服务方法
文档 SHALL 说明动态`service: runtime`支持日志、插件作用域状态和基础信息读取。

#### Scenario: runtime方法覆盖完整
- **WHEN** 动态插件开发者查看`runtime`能力
- **THEN** 文档列出`log.write`、`state.get`、`state.set`、`state.delete`、`info.now`、`info.uuid`和`info.node`

### Requirement: Runtime文档说明与Infra的边界
文档 SHALL 明确`Runtime()`不用于读取基础设施组件状态；组件状态读取应使用`Infra()`和`service: infra`。

#### Scenario: Runtime不替代Infra
- **WHEN** 动态插件开发者需要读取组件状态
- **THEN** 文档引导其查看`Infra`文档和声明`service: infra`
