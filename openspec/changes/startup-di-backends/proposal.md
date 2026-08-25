## Why

规范要求锁、会话、缓存协调这类敏感对象在构造函数里接收启动期那一份后端。现状是`locker.New()`、`session.NewDBStore()`、`cachecoord.Default()`先拿到空或进程单例，再由`ConfigureCoordination`在 HTTP 启动时打补丁；`config.New()`还会内部去拿默认协调器。读启动代码看不出`New()`拿到的是表锁还是 Redis，测试顺序也会改变行为。

## What Changes

- 锁、在线会话存储、缓存协调在构造时接收明确后端（表锁/本地或`coordination`），删除生产路径上的`ConfigureCoordination`和`cachecoord.Default`事后打补丁入口。
- `config`读取面构造时注入缓存协调器，不得在内部调用`cachecoord.Default`。
- 启动装配在`httpstartup`里一次性创建并下发这些实例；集群与单机差异写在构造参数上，不写在进程可变全局状态里。
- 注释、域说明和运行路径对齐既有规范：集群修订走`coordination`/Redis，不把未使用的`sys_cache_revision`表写成现行实现。若确认服务层零调用，从模型和注释中拿掉这条死路径。
- 不合并`cluster`与`coordination`，不删除`hostlock`，不把单机表锁从产品里拿掉，只去掉中间空门面和全局后门。
- 本变更不改插件总门面，不退役动态插件二进制编解码。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `service-dependency-injection-governance`：禁止锁/会话/缓存协调在`New()`之后用包内全局变量切换后端。
- `distributed-locker`：锁服务构造时绑定`LockStore`实现，不再依赖启动后补丁。
- `session-hot-state`：会话存储构造时绑定协调后端或数据库投影，不再依赖进程级`ConfigureCoordination`。
- `distributed-cache-coordination`：`cachecoord`必须由启动编排注入拓扑与协调后端；生产路径不得使用进程单例`Default()`补拓扑。

## Impact

- 代码：`internal/service/locker`、`session`、`cachecoord`、`config`、`internal/cmd/internal/httpstartup`，以及所有`New()`调用点与单测夹具。
- API：无对外 HTTP 契约变更。
- 数据：源码 SQL 不再创建`sys_cache_revision`；不为已部署库保留 DROP 迁移。集群行为仍以 Redis revision 为准。
- 测试：启动装配、单机/集群锁与会话、缓存修订、配置读取缓存域单测；必须覆盖「先`New`再启动」顺序不再改变后端。
- i18n：无用户可见文案变更。
- 数据权限：会话带范围列表不在本变更重划模块边界。
- 缓存：本变更直接治理缓存协调实例来源，必须验证单机与集群分支。
- 依赖：建议在 `workbench-host-contract` 之后、`plugin-hostservice-catalog` 之前合入。
