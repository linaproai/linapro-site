## ADDED Requirements

### Requirement: 会话存储必须在构造时绑定热状态后端

系统 SHALL 在构造在线会话存储时绑定热状态后端与投影存储。集群 MUST 注入 coordination 热状态；单机 MUST 注入本地/SQL 实现。生产路径 MUST NOT 用`session.ConfigureCoordination`在构造后切换后端。

#### Scenario: 集群登录写入注入的热状态

- **WHEN** 集群模式用户登录成功
- **THEN** 会话写入构造时注入的 Redis 或 coordination 热状态
- **AND** 同时更新 PostgreSQL 投影
- **AND** 不得依赖进程全局事后补丁才切换到 Redis

#### Scenario: 单测会话存储隔离

- **WHEN** 单测构造会话存储并传入假热状态后端
- **THEN** 该实例行为只取决于传入依赖
- **AND** 不得读取其他测试留下的`ConfigureCoordination`状态
