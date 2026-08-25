## MODIFIED Requirements

### Requirement: 集群模式领导选举必须使用 Redis 锁
系统 SHALL 在 `cluster.enabled=true` 且 `cluster.coordination.backend=redis` 时使用 Redis lock store 参与领导选举。领导锁 MUST 使用固定 lock name、节点 owner token 和 TTL 租约。`cluster` 实现 MUST 通过注入的 `coordination.LockStore` 使用该锁，不得直接解析 Redis 分组连接。

#### Scenario: 首个节点成为 primary
- **WHEN** 集群模式下第一个节点启动
- **AND** Redis 中不存在领导锁
- **THEN** 节点获取领导锁
- **AND** `IsPrimary` 返回 true

#### Scenario: 第二个节点成为 follower
- **WHEN** 集群模式下已有节点持有领导锁
- **AND** 第二个节点启动
- **THEN** 第二个节点无法获取领导锁
- **AND** `IsPrimary` 返回 false
