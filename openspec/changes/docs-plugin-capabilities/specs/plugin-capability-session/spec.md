## ADDED Requirements

### Requirement: Session文档说明在线会话管理设计
文档`7800-session.md` SHALL 说明`SessionService`的设计定位：管理在线会话的查询和撤销。

#### Scenario: 会话管理设计可理解
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解该服务提供会话的分页查询和按TokenID撤销能力

### Requirement: Session文档说明Session投影模型
文档 SHALL 说明`Session`结构体的投影字段：TokenId、TenantId、UserId、Username、ClientType、LoginTime等。

#### Scenario: 投影模型清晰
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 理解会话投影包含哪些信息及其用途

### Requirement: Session文档包含主要能力概览
文档 SHALL 以表格形式简要列出`ListPage`和`Revoke`两个方法的职责。

#### Scenario: 方法概览可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能快速了解每个方法的用途
