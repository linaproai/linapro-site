## ADDED Requirements

### Requirement: 总览文档包含插件能力分层说明
总览文档`9000-capabilities.md` SHALL 说明`capability.Services`、`pluginhost.Services`、`AdminServices`和动态`hostServices`之间的分层关系，帮助插件开发者理解源码插件、可信源码插件和动态插件各自可用的能力入口。

#### Scenario: 文档结构完整
- **WHEN** 插件开发者阅读总览文档
- **THEN** 文档包含基本介绍、组件结构、能力分层、领域能力概览、源码插件专属能力、管理命令、动态`hostServices`、选型建议和设计约束

### Requirement: 总览文档按父领域能力组织覆盖矩阵
总览文档 SHALL 使用表格列出所有已发布给插件的父领域能力，并链接到对应领域文档；子能力不得作为独立文档入口要求。

#### Scenario: 普通能力覆盖完整
- **WHEN** 插件开发者查看领域能力概览
- **THEN** 表格包含`AI()`、`APIDoc()`、`Auth()`、`Users()`、`BizCtx()`、`Cache()`、`Dict()`、`Files()`、`HostConfig()`、`I18n()`、`Infra()`、`Jobs()`、`Manifest()`、`Notifications()`、`Org()`、`Plugins()`、`Route()`、`Sessions()`和`Tenant()`

#### Scenario: 子能力收敛到父领域
- **WHEN** 插件开发者查找`AI`文本生成、插件配置、插件状态、插件生命周期或租户过滤能力
- **THEN** 文档分别引导到`AI`、`Plugins`、`Tenant`等父领域页面，而不是要求独立子能力页面

### Requirement: 总览文档覆盖源码插件专属能力和管理命令
总览文档 SHALL 说明`pluginhost.Services.Admin()`和`pluginhost.Services.TenantFilter()`的可用范围，并列出`AdminServices`暴露的领域管理命令。

#### Scenario: 源码插件专属入口清晰
- **WHEN** 插件开发者查看源码插件专属能力
- **THEN** 文档说明`Admin()`面向可信源码插件，`TenantFilter()`面向插件自有表租户过滤且携带`gdb.Model`

### Requirement: 总览文档覆盖动态hostServices目录
总览文档 SHALL 列出动态插件已发布的`hostServices`服务、资源声明形态、方法集合和预留未发布条目。

#### Scenario: 动态服务目录覆盖完整
- **WHEN** 插件开发者查看动态服务目录
- **THEN** 表格包含`runtime`、`cron`、`storage`、`network`、`data`、`cache`、`lock`、`notify`、`config`、`hostconfig`、`manifest`、`ai`、`org`和`tenant`

#### Scenario: 预留服务不被误用
- **WHEN** 插件开发者查看`secret`、`event`或`queue`
- **THEN** 文档说明它们是描述符中的预留治理条目，不是已发布的`guest`可调用动态服务

### Requirement: 总览文档包含按场景选能力指南
总览文档 SHALL 提供按使用场景推荐能力入口的快速参考指南。

#### Scenario: 场景指南覆盖常见需求
- **WHEN** 插件开发者需要选择合适入口
- **THEN** 文档提供读取插件配置、读取宿主配置、访问插件自有表、动态对象存储、外部`HTTP`请求、分布式锁、定时任务、通知和租户过滤等场景映射
