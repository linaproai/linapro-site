## MODIFIED Requirements

### Requirement: 集群模式会话热状态必须存储在 Redis
系统 SHALL 在 `cluster.enabled=true` 且 `cluster.coordination.backend=redis` 时，将请求路径使用的在线会话热状态存储到 Redis。PostgreSQL `sys_online_session` SHALL 保留为在线用户管理、数据权限过滤、登录信息展示和清理投影。

#### Scenario: 登录写入 Redis hot state 和 PostgreSQL 投影
- **WHEN** 用户登录成功并签发正式 JWT
- **THEN** 系统写入 Redis session hot key
- **AND** Redis session TTL 等于有效会话超时时长
- **AND** 系统写入或更新 `sys_online_session` 投影行

#### Scenario: 受保护请求验证 Redis hot state
- **WHEN** 已登录用户访问受保护 API
- **AND** JWT 签名有效
- **THEN** 认证链读取 Redis session hot key
- **AND** 仅当 Redis session 存在且未被撤销时请求继续处理
