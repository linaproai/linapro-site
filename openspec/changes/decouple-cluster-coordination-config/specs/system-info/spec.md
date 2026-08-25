## MODIFIED Requirements

### Requirement: 系统信息必须暴露 coordination 状态
系统 SHALL 在系统信息诊断中暴露集群 coordination 状态。响应至少包含 cluster enabled、coordination backend、Redis 健康、当前 node ID、primary 状态和最近错误。

#### Scenario: 集群 Redis 健康状态
- **WHEN** 运维查询系统信息
- **AND** 宿主以 `cluster.coordination.backend=redis` 运行
- **THEN** 响应包含 coordination backend `redis`
- **AND** 响应包含 Redis ping 状态
- **AND** 响应包含当前节点 primary 状态

#### Scenario: Redis 最近错误可见
- **WHEN** Redis coordination 最近发生连接错误
- **THEN** 系统信息响应包含最近错误摘要
- **AND** 不暴露 Redis 密码或敏感连接串
