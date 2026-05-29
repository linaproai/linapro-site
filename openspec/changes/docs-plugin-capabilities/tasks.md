## 1. 目录结构创建

- [x] 1.1 创建目录`apps/lina-site/docs/docs/2000-components/5000-plugin-system/6500-capabilities/`

## 2. 总览文档

- [x] 2.1 创建`6500-capabilities.md`：Services接口设计原则、mermaid架构图、获取方式代码示例、17个服务分类速查表、按场景选服务指南

## 3. 认证与上下文类服务文档

- [x] 3.1 创建`6600-apidoc.md`：APIDocService架构设计——操作键构建机制、在请求链路中的位置
- [x] 3.2 创建`6700-auth.md`：AuthService架构设计——两阶段认证模型、模拟令牌治理边界
- [x] 3.3 创建`6800-bizctx.md`：BizCtxService架构设计——业务上下文投影、与Auth和Session的关系
- [x] 3.4 创建`7100-i18n.md`：I18nService架构设计——运行时翻译机制、翻译键命名约定

## 4. 配置与资源类服务文档

- [x] 4.1 创建`7000-config.md`：ConfigService与HostConfigService架构设计——插件配置读取优先级、宿主配置白名单边界
- [x] 4.2 创建`7200-manifest.md`：ManifestService架构设计——资源路径语义、资源管线边界

## 5. 数据与治理类服务文档

- [x] 5.1 创建`6900-cache.md`：CacheService架构设计——插件作用域隔离、值类型和过期策略
- [x] 5.2 创建`7300-notify.md`：NotifyService架构设计——通知发布模型、SourceType和CategoryCode分类
- [x] 5.3 创建`7500-plugin-lifecycle.md`：PluginLifecycleService架构设计——租户级生命周期编排、与pluginhost.Lifecycle的区别
- [x] 5.4 创建`7600-plugin-state.md`：PluginStateService架构设计——本地快照与权威读取策略、Provider启用状态
- [x] 5.5 创建`7700-route.md`：RouteService架构设计——动态路由元数据、与动态插件的关系
- [x] 5.6 创建`7800-session.md`：SessionService架构设计——在线会话管理、Session投影模型

## 6. 能力提供方类服务文档

- [x] 6.1 创建`7400-org.md`：orgcap.Service架构设计——组织能力消费侧接口、Provider扩展机制、能力降级策略
- [x] 6.2 创建`7900-tenant.md`：tenantcap.Service架构设计——租户能力消费侧接口、Provider和Resolver扩展机制、降级策略
- [x] 6.3 创建`8000-tenant-filter.md`：TenantFilterService架构设计——数据库查询注入模式、与pluginhost.Services的关系

## 7. 文档关联与验证

- [x] 7.1 更新`6000-source-plugins.md`：在插件配置和清单资源章节添加指向新文档的链接
- [x] 7.2 验证所有文档的frontmatter符合`markdown-format.instructions.md`规范（description不少于100字、keywords不少于15个）
- [x] 7.3 验证侧边栏自动生成正确，`6500-capabilities/`显示为`5000-plugin-system/`的子分类
