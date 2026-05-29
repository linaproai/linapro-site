## ADDED Requirements

### Requirement: APIDoc文档说明设计动机
文档`6600-apidoc.md` SHALL 说明`APIDocService`解决的问题：API文档本地化和操作键翻译。

#### Scenario: 设计动机清晰
- **WHEN** 插件开发者阅读基本介绍
- **THEN** 理解该服务用于将路由操作键解析为本地化的模块标签和操作摘要

### Requirement: APIDoc文档说明操作键构建机制
文档 SHALL 说明操作键的构建方式：静态路由从DTO类型派生，动态路由从路径派生。

#### Scenario: 操作键机制可理解
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解`BuildOperationKeyFromHandler`和`BuildOperationKeyFromPath`的区别和适用场景

### Requirement: APIDoc文档包含主要能力概览
文档 SHALL 以表格形式简要列出`ResolveRouteText`、`ResolveRouteTexts`、`FindRouteTitleOperationKeys`三个方法的职责。

#### Scenario: 方法概览可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能快速了解每个方法的用途
