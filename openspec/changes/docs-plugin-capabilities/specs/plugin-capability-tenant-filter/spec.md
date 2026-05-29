## ADDED Requirements

### Requirement: TenantFilter文档说明数据库查询注入设计
文档`8000-tenant-filter.md` SHALL 说明`TenantFilterService`的核心设计：为插件自有表注入`tenant_id`过滤条件。

#### Scenario: 查询注入设计可理解
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解`Apply`方法如何向`*gdb.Model`追加租户条件，以及`PlatformBypass`的豁免逻辑

### Requirement: TenantFilter文档说明与pluginhost.Services的关系
文档 SHALL 说明`TenantFilter`不在`capability.Services`中，而是通过`pluginhost.Services`扩展暴露，因为其携带数据库查询构建器。

#### Scenario: 接口层级清晰
- **WHEN** 插件开发者阅读设计约束
- **THEN** 理解为什么`TenantFilter`是源码插件独有能力，不通过普通`capability.Services`暴露

### Requirement: TenantFilter文档说明TenantFilterContext字段
文档 SHALL 说明`TenantFilterContext`结构体中的关键字段及其在不同请求场景下的取值。

#### Scenario: 上下文字段清晰
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 理解`UserID`、`TenantID`、`ActingUserID`、`PlatformBypass`等字段的语义

### Requirement: TenantFilter文档包含主要能力概览
文档 SHALL 以表格形式简要列出`Context`和`Apply`两个方法的职责。

#### Scenario: 方法概览可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能快速了解每个方法的用途
