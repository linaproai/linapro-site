## ADDED Requirements

### Requirement: Route文档说明动态路由元数据设计
文档`7700-route.md` SHALL 说明`RouteService`的设计定位：为源码插件提供当前动态路由请求的元数据访问。

#### Scenario: 元数据设计可理解
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解`DynamicRouteMetadata`包含插件ID、HTTP方法、公开路径、标签、摘要等路由声明信息

### Requirement: Route文档说明与动态插件的关系
文档 SHALL 说明该服务主要服务于动态插件路由场景，源码插件通常不需要主动使用。

#### Scenario: 适用场景清晰
- **WHEN** 插件开发者阅读设计约束
- **THEN** 理解该服务在动态插件请求分发链路中的位置

### Requirement: Route文档包含主要能力概览
文档 SHALL 以表格形式简要列出`DynamicRouteMetadata`方法和`DynamicRouteMetadata`结构体字段。

#### Scenario: 方法概览可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能快速了解可用的元数据字段
