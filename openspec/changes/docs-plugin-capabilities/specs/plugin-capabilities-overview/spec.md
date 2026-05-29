## ADDED Requirements

### Requirement: 总览文档包含Services架构说明
总览文档`6500-capabilities.md` SHALL 包含`Services`接口的设计原则和架构定位说明，帮助插件开发者理解基础能力服务在LinaPro插件系统中的角色。

#### Scenario: 文档结构完整
- **WHEN** 插件开发者阅读总览文档
- **THEN** 文档包含以下章节：基本介绍、服务架构、获取方式、服务分类速查、快速参考、相关内容

### Requirement: 总览文档包含mermaid架构图
总览文档 SHALL 使用mermaid图展示17个服务的分类关系和协作模式。

#### Scenario: 架构图展示服务分类
- **WHEN** 插件开发者查看架构图
- **THEN** 图中展示认证与上下文、配置与资源、数据与存储、插件治理、通知、能力提供方等服务分类

### Requirement: 总览文档包含服务获取方式说明
总览文档 SHALL 说明源码插件通过`registrar.Services()`获取服务的代码模式。

#### Scenario: 代码示例展示获取方式
- **WHEN** 插件开发者查看获取方式章节
- **THEN** 文档包含`registrar.Services()`的Go代码示例

### Requirement: 总览文档包含服务分类速查表
总览文档 SHALL 包含17个服务的分类速查表格，列出服务名称、合约类型和一句话描述。

#### Scenario: 速查表覆盖所有服务
- **WHEN** 插件开发者查看速查表
- **THEN** 表格包含APIDoc、Auth、BizCtx、Cache、Config、HostConfig、I18n、Manifest、Notify、Org、PluginLifecycle、PluginState、Route、Session、Tenant、TenantFilter共16个服务条目

### Requirement: 总览文档包含按场景选服务指南
总览文档 SHALL 提供按使用场景推荐服务的快速参考指南。

#### Scenario: 场景指南覆盖常见需求
- **WHEN** 插件开发者需要选择合适的服务
- **THEN** 文档提供如"读取插件配置→Config"、"租户过滤→TenantFilter"等场景到服务的映射
