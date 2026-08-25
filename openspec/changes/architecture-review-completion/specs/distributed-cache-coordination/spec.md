## ADDED Requirements

### Requirement: 模块和插件使用一份带命名空间的 TTL 缓存

系统 SHALL 向模块和插件提供一份带命名空间和过期时间的缓存产品。调用方 MUST NOT 同时面对 coordination 键值 API 和另一套并列的 kvcache 产品。领域修订策略可以留在`cachecoord`内部，修订转发器 MUST NOT 再作为独立产品。`cachecoord.New`构造的修订状态 MUST 属于该实例，MUST NOT 写入进程全局表。

#### Scenario: 两次 New 互不影响

- **WHEN** 测试连续构造两个`cachecoord.New`实例并 bump 其中一个领域
- **THEN** 另一个实例的修订号不变

#### Scenario: 插件写缓存

- **WHEN** 插件或模块写入带 TTL 的缓存项
- **THEN** 它只调用命名空间 TTL 缓存入口
- **AND** 不需要同时导入 coordination KV 与 kvcache 两套产品 API

### Requirement: sys_cache_revision 不是现行路径

系统 MUST NOT 把`sys_cache_revision`表、DAO 或「行锁递增修订」注释当作现行实现。集群修订走注入的 coordination revision store。

#### Scenario: 阅读 cachecoord 修订注释

- **WHEN** 开发者阅读修订 bump 实现
- **THEN** 注释描述 Redis 或进程内 revision store
- **AND** 不得再写持久修订行或行锁
