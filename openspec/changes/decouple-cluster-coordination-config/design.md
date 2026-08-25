## Context

连接参数若放在 `cluster` 或顶层 `coordination` 下，缓存等平级产品要么去读别人的配置，要么再抄一份 Redis。用户确认的形状是：集群只选择后端和分组，Redis 连接按 GoFrame 风格做成顶层具名分组。

## Goals / Non-Goals

**Goals:**

- YAML 采用 `cluster.coordination.backend/group` + 顶层 `redis.<group>`。
- 启动编排按分组构造 Redis 客户端并注入 `cluster.Service` 与其他 coordination 消费者。
- `address` 支持逗号分隔多节点，多地址时按 GoFrame 规则使用 Redis Cluster 客户端。
- 单机不要求 Redis 分组可连。

**Non-Goals:**

- 不把 kvcache 自动绑到名为 `cache` 的分组；未单独声明用途绑定时，缓存仍用集群协调所选分组。
- 不实现 Sentinel / 第二协调后端。
- 不引入 backend 工厂注册表。

## Decisions

1. **集群选择 vs 连接**：`cluster.coordination` 只含 `backend`、`group`。连接字段只出现在 `redis.<group>`。
2. **分组形状**：`redis` 段是分组名到连接参数的映射，与 GoFrame `redis.default` / `redis.cache` 一致。`password` 保持本仓库字段名，不改成 GoFrame 的 `pass`。
3. **默认分组**：`group` 缺省为 `default`。集群启用时该分组必须存在且 `address` 非空。
4. **多地址**：`address` 用逗号分隔；多于一个地址时使用 `go-redis` Cluster/Universal 客户端，行为对齐 GoFrame `SplitAndTrim(address, ",")` 且 `len(Addrs)>1` 走 Cluster。
5. **注入**：`httpstartup` 读取集群选择和 Redis 分组，调用 `coordination.NewRedis`，再 `cluster.New(cfg, coordinationSvc)`。`cluster` 实现不读 `redis.*`。

## Risks / Trade-offs

- [配置 BREAKING] 旧的 `cluster.redis` 与顶层 `coordination.redis` 均不再读取。
- [分组未绑定] YAML 可声明 `redis.cache` 但当前没有独立 cache 用途选择器，避免静默拆连接。
- [Redis Cluster] 多地址走 Cluster 客户端后，逻辑 db 在 Cluster 模式下通常被忽略，需在注释中说明。
