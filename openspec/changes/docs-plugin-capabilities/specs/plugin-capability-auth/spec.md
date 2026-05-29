## ADDED Requirements

### Requirement: Auth文档说明两阶段认证模型
文档`6700-auth.md` SHALL 说明`AuthService`的PreToken→TenantToken两阶段认证设计。

#### Scenario: 两阶段模型可理解
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解登录产生PreToken、选择租户后签发TenantToken的流程

### Requirement: Auth文档说明模拟令牌治理边界
文档 SHALL 说明模拟令牌（ImpersonationToken）的设计约束：仅平台管理员可用、需要审计记录。

#### Scenario: 模拟令牌边界清晰
- **WHEN** 插件开发者阅读设计约束
- **THEN** 理解插件在调用`IssueImpersonationToken`前需自行完成业务授权和审计检查

### Requirement: Auth文档包含主要能力概览
文档 SHALL 以表格形式简要列出`SelectTenant`、`SwitchTenant`、`IssueImpersonationToken`、`RevokeImpersonationToken`四个方法的职责。

#### Scenario: 方法概览可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能快速了解每个方法的用途
