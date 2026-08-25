## Why

集群部署配置不应持有 Redis 连接参数，也不该把连接嵌在顶层 `coordination` 段里让缓存等消费者去翻协调配置。连接是共享基础设施，用途选择属于各产品配置；实现对象由启动编排注入。

## What Changes

- `cluster` 只保留拓扑、选举时序，以及协调用途选择：`coordination.backend`、`coordination.group`。
- 删除顶层 `coordination` 配置段。Redis 连接放到顶层具名分组 `redis.<group>`。
- `cluster.coordination.group` 指向 `redis` 分组名；程序按 `cluster.enabled` 和所选分组构造 Redis 实现并注入。
- `address` 遵循 GoFrame Redis 配置：单地址为普通实例，逗号分隔多地址按 Redis Cluster 客户端连接。
- `cluster.Service` 运行时仍只消费注入的 `coordination.Service`，不解析 Redis 连接。
- 额外 Redis 分组可以并存；当前 kvcache/会话/修订号仍使用集群协调所选分组，不自动拆到 `cache` 分组。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `cluster-coordination-config`：Redis 连接改为顶层具名分组；集群只选择 backend 与 group。
- `cluster-deployment-mode`：集群启用条件为 `cluster.enabled=true` 且 `cluster.coordination.backend=redis`，并解析对应 Redis 分组。
- `cluster-topology-boundaries`：业务组件不得自建 Redis；集群实现不解析连接参数。
- `leader-election`：通过注入的 coordination lock 选主。
- `session-hot-state`：热状态仍走注入的 coordination。
- `system-info`：诊断报告实际 backend 名。

## Impact

- 代码：`cluster`、`coordination` Redis 客户端、`config`、`httpstartup`、配置模板与文档。
- API：无 HTTP 契约变更。`config.yaml` BREAKING：连接参数为 `redis.<group>`，集群选择为 `cluster.coordination.backend/group`。
- 数据：无表结构变更。
- 测试：配置加载、分组解析、多地址、启动装配单测。
- i18n：无运行时文案变更。
- 数据权限：无影响。
- 缓存：修订号与 KV 仍走注入的 coordination；默认与集群协调共用所选 Redis 分组。
