## ADDED Requirements

### Requirement: Tenant文档说明租户能力消费侧设计
文档`7900-tenant.md` SHALL 说明`tenantcap.Service`作为租户能力消费接口的设计定位：获取当前租户、校验租户可见性、列出用户可访问租户。

#### Scenario: 消费侧设计可理解
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解租户能力采用"提供方+消费侧"模式，`Service`接口面向普通插件消费

### Requirement: Tenant文档说明Provider和Resolver扩展机制
文档 SHALL 说明`Provider`接口（租户解析、校验、切换）和`Resolver`接口（HTTP请求租户解析）的扩展机制。

#### Scenario: 扩展机制可理解
- **WHEN** 插件开发者阅读提供方指引
- **THEN** 理解`tenantcap.Provide()`注册工厂、`Provider`和`Resolver`的不同职责

### Requirement: Tenant文档说明租户能力降级策略
文档 SHALL 说明租户能力不可用时的安全降级行为：默认返回平台租户（`PlatformTenantID = 0`）。

#### Scenario: 降级策略清晰
- **WHEN** 插件开发者阅读设计约束
- **THEN** 理解单租户模式或插件禁用时系统自动降级到平台租户

### Requirement: Tenant文档包含主要能力概览
文档 SHALL 以表格形式简要列出`Service`接口的主要方法。

#### Scenario: 方法概览可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能快速了解消费侧可用的方法
