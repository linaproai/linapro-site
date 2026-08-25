## 1. 配置契约拆分

- [x] 1.1 将 `coordination.backend` 与 `coordination.redis` 作为独立配置段加载，并从 `cluster` 配置值对象中删除 Redis 字段
- [x] 1.2 集群启用时校验独立协调配置；单机不要求 Redis；更新静态配置默认值和 panic 诊断字段名

## 2. 集群实现解耦

- [x] 2.1 `cluster.New` 只接收拓扑配置和已构造的 `coordination.Service`；生产实现去掉 Redis 类型、字段、常量和标识
- [x] 2.2 启动编排从独立协调配置构造后端并注入集群对象；业务组件不得再读取 `cluster.redis`

## 3. 验证与文档

- [x] 3.1 更新配置、集群构造、启动装配和相关 fake 测试
- [x] 3.2 同步配置模板、宿主 README 和当前官网文档中的配置键

## Feedback

- [x] **FB-1**: 将 Redis 配置从集群部署配置中拆出并独立注入，且集群管理实现不得出现 Redis 标识
- [x] **FB-2**: 按用户 YAML 收口为 cluster.coordination 选择 backend/group，连接放到顶层 redis 具名分组，address 支持逗号分隔集群节点

### 任务记录

- **DI 来源检查**：
  - Owner：`httpstartup.newHTTPRuntime`。
  - 创建位置：读取 `cluster.enabled` 与 `cluster.coordination.backend/group`，从 `config.GetRedis()` 取对应分组，构造 `coordination.NewRedis`，再 `cluster.New(clusterCfg, coordinationSvc)`。
  - 传递路径：同一 `coordinationSvc` 继续注入 cachecoord、kvcache、locker、session。
  - 共享实例：生产路径仍只构造一份 coordination；当前 kvcache 不自动改绑 `redis.cache` 分组。
- **i18n**：无运行时文案、前端 UI、API 文档源文本、插件清单或语言包变更。
- **缓存一致性**：权威数据源仍是业务库；集群修订/KV/锁仍走注入的 coordination。多分组只提供连接选择，不改变失效模型。
- **数据权限**：无数据操作边界变更。
- **开发工具跨平台**：无 Makefile/脚本变更。
- **测试策略**：配置加载、分组并存、逗号地址解析、启动装配、panic 诊断单测。无 UI 变更，不新增 E2E。
- **外部规则**：已读 `openspec.md`、`architecture.md`、`cache-consistency.md`、`backend-go.md`、`testing.md`、`i18n.md`、`documentation.md`。`data-permission.md` 无影响；`plugin.md` 无插件目录契约变更。
