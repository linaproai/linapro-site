## MODIFIED Requirements

### Requirement: 集群模式必须使用 Redis coordination
系统 SHALL 在 PostgreSQL 集群模式下使用 Redis coordination 作为当前唯一支持的分布式协调实现。`cluster.enabled=true` MUST 与 `cluster.coordination.backend=redis` 同时成立，并成功解析 `cluster.coordination.group` 对应的 `redis` 分组后才允许进入集群启动流程。

#### Scenario: PostgreSQL 集群模式启用 Redis coordination
- **WHEN** 数据库链接为 PostgreSQL
- **AND** `cluster.enabled=true`
- **AND** `cluster.coordination.backend=redis`
- **AND** 所选 Redis 分组探活成功
- **THEN** 宿主进入集群模式
- **AND** leader election、cache coordination、session hot state 和 kvcache 均使用 coordination provider

#### Scenario: PostgreSQL 集群模式未配置 coordination backend
- **WHEN** 数据库链接为 PostgreSQL
- **AND** `cluster.enabled=true`
- **AND** `cluster.coordination.backend` 缺失
- **THEN** 宿主启动失败
- **AND** 不得回退到 PostgreSQL 表协调实现
- **AND** 不得回退到节点本地 `memory`

### Requirement: 集群模式不得使用 PostgreSQL 作为跨节点协调主实现
系统 SHALL 禁止集群模式依赖 `sys_locker`、`sys_cache_revision` 或 `sys_kv_cache` 完成跨节点一致性。

#### Scenario: 集群模式 cachecoord 不写 sys_cache_revision
- **WHEN** `cluster.enabled=true` 且 `cluster.coordination.backend=redis`
- **AND** 业务写路径发布缓存 revision
- **THEN** 系统使用 Redis revision store
- **AND** 不依赖 `sys_cache_revision` 递增来通知其他节点

#### Scenario: 集群模式 leader election 不写 sys_locker
- **WHEN** `cluster.enabled=true` 且 `cluster.coordination.backend=redis`
- **AND** 节点参与 primary election
- **THEN** 系统使用 Redis lock store
- **AND** 不依赖 `sys_locker` 判断 primary

#### Scenario: 集群模式 kvcache 不写 sys_kv_cache
- **WHEN** `cluster.enabled=true` 且 `cluster.coordination.backend=redis`
- **AND** 插件、认证短期状态或宿主模块写入 kvcache
- **THEN** 系统使用 Redis coordination KV backend
- **AND** 不写入 `sys_kv_cache`
