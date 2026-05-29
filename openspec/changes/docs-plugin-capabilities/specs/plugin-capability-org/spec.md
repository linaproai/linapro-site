## ADDED Requirements

### Requirement: Org文档说明组织能力消费侧设计
文档`7400-org.md` SHALL 说明`orgcap.Service`作为只读组织能力消费接口的设计定位：插件通过`Services.Org()`获取用户部门、岗位等组织投影。

#### Scenario: 消费侧设计可理解
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解组织能力采用"能力提供方+消费侧服务"模式，普通插件只消费只读投影

### Requirement: Org文档说明Provider扩展机制
文档 SHALL 说明`Provider`接口和`ProviderFactory`注册机制，即如何通过插件扩展组织能力。

#### Scenario: Provider机制可理解
- **WHEN** 插件开发者阅读提供方指引
- **THEN** 理解`orgcap.Provide()`注册工厂、`ProviderEnv`构造环境、`Provider`接口的核心方法

### Requirement: Org文档说明能力降级策略
文档 SHALL 说明组织能力不可用时的安全降级行为：`Available()`返回`false`，查询方法返回空结果。

#### Scenario: 降级策略清晰
- **WHEN** 插件开发者阅读设计约束
- **THEN** 理解插件应先检查`Available()`再使用组织能力，避免依赖可能不存在的数据

### Requirement: Org文档包含主要能力概览
文档 SHALL 以表格形式简要列出`Service`接口的主要方法。

#### Scenario: 方法概览可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能快速了解消费侧可用的方法
