## ADDED Requirements

### Requirement: 集群管理实现不得解析 Redis 连接
`cluster.Service` 生产实现 MUST 只依赖拓扑配置和注入的 `coordination.Service`。`ClusterConfig` MAY 包含 `coordination.backend` 与 `coordination.group` 作为用途选择。`ClusterConfig` 与 `cluster` 包生产实现 MUST NOT 包含 Redis 连接字段。启动编排 MUST 根据所选分组构造协调后端并注入。

#### Scenario: 集群对象只消费注入的 coordination
- **WHEN** 宿主以集群模式启动
- **THEN** 启动编排根据 `cluster.coordination` 与 `redis.<group>` 构造 Redis coordination provider
- **AND** `cluster.New` 接收该 provider 进行 primary election
- **AND** `cluster` 包生产实现不创建 Redis client

#### Scenario: 集群配置值对象不含 Redis 连接
- **WHEN** 调用方读取 `GetCluster`
- **THEN** 返回对象可包含协调 backend 与 group
- **AND** 不包含 address、password 或超时等连接字段

## MODIFIED Requirements

### Requirement: 集群拓扑必须由统一 coordination 注入
系统 SHALL 通过统一启动编排创建 coordination provider，并将其注入 cluster、locker、cachecoord、kvcache、auth、session、cron 和插件运行时等需要集群协调的组件。业务组件 MUST 不自行解析 Redis 分组连接。

#### Scenario: 启动编排注入 coordination
- **WHEN** 宿主以集群模式启动
- **THEN** 启动编排先根据所选 Redis 分组创建 Redis coordination provider
- **AND** cluster service 使用该 provider 进行 primary election
- **AND** 其他组件通过构造参数接收 provider 或 provider-backed service

#### Scenario: 禁止组件自行读取 Redis 配置
- **WHEN** `role` 或 `pluginruntimecache` 需要发布跨节点 revision
- **THEN** 它们通过 `cachecoord` 或 coordination-backed controller 完成
- **AND** 不读取 `redis.default.address`
- **AND** 不创建 Redis client

### Requirement: Primary 判定必须与 coordination lock 状态一致
系统 SHALL 在集群模式下以注入的 coordination leader lock 持有状态作为 `IsPrimary` 的权威来源。续约失败或锁丢失后，`IsPrimary` MUST 立即返回 false。

#### Scenario: 续约失败后 primary 状态变更
- **WHEN** 当前 primary 节点无法续约 leader lock
- **THEN** cluster service 将本节点降级为 follower
- **AND** `IsPrimary` 返回 false
- **AND** 主节点专属任务停止执行
