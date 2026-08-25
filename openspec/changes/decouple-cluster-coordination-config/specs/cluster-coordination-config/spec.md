## MODIFIED Requirements

### Requirement: 集群模式必须声明协调后端和 Redis 分组
系统 SHALL 在 `cluster.enabled=true` 时要求 `cluster.coordination.backend`。当前版本唯一合法值 MUST 为 `redis`。系统 SHALL 使用 `cluster.coordination.group` 选择顶层 `redis` 分组；未声明时 MUST 默认 `default`。当 `cluster.enabled=false` 时，系统 MUST 不要求 Redis 分组可连接，也不得因为 Redis 缺失而影响单机启动。连接参数 MUST NOT 出现在 `cluster` 段。

#### Scenario: 集群模式缺少 coordination backend
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **AND** 未声明 `cluster.coordination.backend`
- **THEN** 宿主启动失败
- **AND** 错误信息明确指出必须配置 `cluster.coordination.backend=redis`

#### Scenario: 集群模式配置非法 coordination backend
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **AND** `cluster.coordination.backend=postgres`
- **THEN** 宿主启动失败
- **AND** 错误信息明确指出当前仅支持 `redis`

#### Scenario: 单机模式不要求 Redis
- **WHEN** 配置文件声明 `cluster.enabled=false`
- **AND** 未声明 `redis` 分组
- **THEN** 宿主按单机模式启动成功
- **AND** 系统不得尝试连接 Redis

### Requirement: Redis 连接必须使用顶层具名分组
系统 SHALL 从顶层 `redis.<group>` 读取连接配置。每个分组 MUST 支持 `address`、`db`、`password`、`connectTimeout`、`readTimeout`、`writeTimeout`。时间长度 MUST 使用带单位的时长字符串并解析为 `time.Duration`。`address` MUST 支持单个 `host:port`，也 MUST 支持逗号分隔的多个节点；多于一个节点时系统 MUST 按 GoFrame Redis 配置规则使用 Redis Cluster 客户端。

#### Scenario: Redis 分组解析成功
- **WHEN** 配置文件声明 `cluster.coordination.backend=redis`
- **AND** `cluster.coordination.group=default`
- **AND** `redis.default.address="127.0.0.1:6379"`
- **THEN** 配置服务返回该分组连接对象
- **AND** 超时字段均为 `time.Duration`
- **AND** `GetCluster` 结果不含 Redis 连接字段

#### Scenario: 多分组并存
- **WHEN** 配置文件同时声明 `redis.default` 和 `redis.cache`
- **THEN** 配置服务可分别读取两个分组
- **AND** 未绑定用途的分组不阻止启动

#### Scenario: 所选分组缺失
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **AND** `cluster.coordination.group=cache`
- **AND** 未声明 `redis.cache`
- **THEN** 宿主启动失败
- **AND** 错误信息包含缺失分组 `redis.cache`

#### Scenario: Redis address 缺失
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **AND** `cluster.coordination.backend=redis`
- **AND** 所选分组 `address` 为空
- **THEN** 宿主启动失败
- **AND** 错误信息包含缺失字段 `redis.<group>.address`

#### Scenario: 逗号分隔地址按集群客户端处理
- **WHEN** 所选分组 `address="127.0.0.1:6379,127.0.0.1:6370"`
- **THEN** 系统将该地址解析为两个节点
- **AND** 使用 Redis Cluster 客户端而不是单节点客户端

#### Scenario: Redis timeout 格式非法
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **AND** `redis.default.readTimeout=2000`
- **THEN** 宿主启动失败
- **AND** 错误信息要求使用带单位的时长字符串

### Requirement: 集群启动必须先完成 Redis 探活
系统 SHALL 在 HTTP 服务、定时任务、插件运行时和业务路由启动前完成所选 Redis 分组探活。探活失败时，系统 MUST 拒绝以集群模式启动。

#### Scenario: Redis 不可达时拒绝启动
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **AND** `cluster.coordination.backend=redis`
- **AND** 所选 Redis 分组地址不可连接
- **THEN** 宿主启动失败
- **AND** 不注册 HTTP 业务路由
- **AND** 不启动 leader election、cron、插件 runtime reconciler 或缓存 watcher

#### Scenario: Redis 探活成功后继续启动
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **AND** `cluster.coordination.backend=redis`
- **AND** 所选分组 ping 成功
- **THEN** 宿主继续初始化 cluster、coordination、cron 和插件运行时组件
- **AND** 系统信息诊断中显示 coordination backend 为 `redis`

### Requirement: 非 PostgreSQL 数据库链接必须在 coordination 启动前失败
系统仅支持 PostgreSQL 运行时数据库。`sqlite:`、`mysql:` 或未知数据库链接 MUST 在方言解析阶段失败，不得进入 Redis 探活、集群配置覆盖或业务启动流程。

#### Scenario: SQLite 配置了 Redis coordination
- **WHEN** `database.default.link` 以 `sqlite:` 开头
- **AND** 配置文件声明 `cluster.enabled=true`
- **AND** 配置文件声明 `cluster.coordination.backend=redis`
- **THEN** 宿主启动失败并返回 SQLite 不再支持的明确错误
- **AND** 系统不得连接 Redis

#### Scenario: SQLite 配置了单机模式
- **WHEN** `database.default.link` 以 `sqlite:` 开头
- **AND** 配置文件声明 `cluster.enabled=false`
- **THEN** 宿主启动失败并返回 SQLite 不再支持的明确错误

### Requirement: 配置模板必须展示 Redis 分组和集群选择
系统 SHALL 在 `manifest/config/config.template.yaml` 中提供顶层 `redis` 分组示例，以及 `cluster.coordination.backend` / `group` 选择示例。注释 MUST 说明单机模式不需要连接 Redis，集群模式必须选择合法分组。

#### Scenario: 配置模板包含 Redis 分组
- **WHEN** 开发者查看 `config.template.yaml`
- **THEN** 文件包含 `cluster.coordination.backend: redis` 和 `cluster.coordination.group`
- **AND** 文件包含顶层 `redis.default` 的 `address`、`db`、`password`、超时字段
- **AND** `cluster` 段不含 Redis 连接字段
- **AND** 注释说明 `cluster.enabled=false` 时不需要 Redis
