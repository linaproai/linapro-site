## 1. 目录结构创建

- [x] 1.1 创建目录`apps/lina-site/docs/docs/2000-components/5000-plugin-system/9000-capabilities/`

## 2. 总览文档

- [x] 2.1 创建`9000-capabilities.md`：插件能力分层、mermaid架构图、普通能力矩阵、源码插件专属能力、管理命令、动态`hostServices`目录和按场景选型指南

## 3. 认证与上下文类服务文档

- [x] 3.1 创建`6600-apidoc.md`：APIDocService架构设计——操作键构建机制、在请求链路中的位置
- [x] 3.2 创建`6700-auth.md`：AuthService架构设计——两阶段认证模型、模拟令牌治理边界
- [x] 3.3 创建`6800-bizctx.md`：BizCtxService架构设计——业务上下文投影、与Auth和Session的关系
- [x] 3.4 创建`7100-i18n.md`：I18nService架构设计——运行时翻译机制、翻译键命名约定

## 4. 配置与资源类服务文档

- [x] 4.1 创建`7000-config.md`：配置管理能力关系介绍——插件配置、宿主配置和运行时配置管理的领域边界
- [x] 4.2 创建`7200-manifest.md`：ManifestService架构设计——资源路径语义、资源管线边界

## 5. 数据与治理类服务文档

- [x] 5.1 创建`6900-cache.md`：CacheService架构设计——插件作用域隔离、值类型和过期策略
- [x] 5.2 创建`7300-notify.md`：NotifyService架构设计——通知发布模型、SourceType和CategoryCode分类
- [x] 5.3 在`7500-plugins.md`中介绍`Plugins().Lifecycle()`子能力：租户级生命周期编排、与`pluginhost.Lifecycle()`的区别
- [x] 5.4 在`7500-plugins.md`中介绍`Plugins().State()`子能力：本地快照、权威读取和Provider启用状态
- [x] 5.5 创建`7700-route.md`：RouteService架构设计——动态路由元数据、与动态插件的关系
- [x] 5.6 创建`7800-session.md`：SessionService架构设计——在线会话管理、Session投影模型

## 6. 能力提供方类服务文档

- [x] 6.1 创建`7400-org.md`：orgcap.Service架构设计——组织能力消费侧接口、Provider扩展机制、能力降级策略
- [x] 6.2 创建`7900-tenant.md`：tenantcap.Service架构设计——租户能力消费侧接口、Provider和Resolver扩展机制、降级策略
- [x] 6.3 在`7900-tenant.md`中介绍`TenantFilter()`源码插件专属子能力：数据库查询注入模式、与`pluginhost.Services`的关系

## 7. 文档关联与验证

- [x] 7.1 更新`6000-source-plugins.md`：在插件配置和清单资源章节添加指向新文档的链接
- [x] 7.2 验证所有文档的frontmatter符合`markdown-format.instructions.md`规范（description不少于100字、keywords不少于15个）
- [x] 7.3 验证侧边栏自动生成正确，`9000-capabilities/`显示为`5000-plugin-system/`的子分类

## Feedback

- [ ] **FB-1**: 按领域能力重组插件能力文档，覆盖所有已暴露给插件的普通能力、源码插件专属能力、管理命令和动态`hostServices`
- [x] **FB-2**: 拆分`Infra`基础设施领域能力与动态插件专属`Runtime`能力文档边界
- [x] **FB-3**: 根据`linapro`源码修正领域能力文档中的接口签名、动态插件SDK入口、`hostServices`资源结构和示例代码
- [x] **FB-4**: 根据最新`linapro`源码将声明期`Governance()`能力文档更名并修正为`Access()`/`AccessDeclarations`
- [x] **FB-5**: 修复`make build`报告的当前版和`0.3.x`版本文档坏链
