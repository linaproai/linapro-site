## ADDED Requirements

### Requirement: BizCtx文档说明业务上下文投影设计
文档`6800-bizctx.md` SHALL 说明`BizCtxService`作为请求级只读快照的设计定位。

#### Scenario: 投影设计可理解
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解该服务返回`CurrentContext`只读快照，不暴露宿主内部上下文模型

### Requirement: BizCtx文档说明CurrentContext字段含义
文档 SHALL 说明`CurrentContext`结构体中的关键字段：`UserID`、`TenantID`、`ActingUserID`、`IsImpersonation`、`PlatformBypass`等。

#### Scenario: 字段含义清晰
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 理解各字段在不同请求场景下的取值语义

### Requirement: BizCtx文档说明与Auth和Session的关系
文档 SHALL 在架构位置章节说明`BizCtx`与`Auth`、`Session`服务的协作关系。

#### Scenario: 服务关系可理解
- **WHEN** 插件开发者查看架构图
- **THEN** 理解BizCtx在请求链路中作为上下文投影层的角色
