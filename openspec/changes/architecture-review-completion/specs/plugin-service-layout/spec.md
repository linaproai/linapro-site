## ADDED Requirements

### Requirement: HTTP 控制器依赖插件默认 Service

系统 SHALL 要求宿主插件 HTTP 控制器依赖 `internal/service/plugin` 导出的默认 `Service`。根包可以继续用私有 facet 组合 `Service`，但插件管理控制器 MUST NOT 把从 `Service` 拆出的管理、启动或运行时子接口作为构造函数参数或结构体字段。`httpRuntime` MUST 继续保存单个 `plugin.Service` 统一入口，并把它传给控制器。

#### Scenario: 插件管理控制器构造

- **WHEN** 插件管理控制器通过 `NewV1` 被构造
- **THEN** 构造函数和控制器字段使用 `pluginsvc.Service`
- **AND** 不得依赖从 `Service` 拆出的管理、启动或运行时子接口才能编译
