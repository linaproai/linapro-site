## ADDED Requirements

### Requirement: Lock文档说明动态分布式锁能力
文档`8300-lock.md` SHALL 说明动态插件通过`hostServices.lock`使用受治理分布式锁，并覆盖获取、续租和释放方法。

#### Scenario: lock方法覆盖完整
- **WHEN** 动态插件开发者查看分布式锁能力
- **THEN** 文档列出`lock.acquire`、`lock.renew`和`lock.release`，并说明`ticket`和`leaseMillis`语义

### Requirement: Lock文档说明锁不是权限边界
文档 SHALL 说明锁只用于并发协调，不替代授权、租户过滤或数据可见性校验。

#### Scenario: 并发边界清晰
- **WHEN** 插件开发者使用锁保护业务处理
- **THEN** 文档要求业务仍执行权限和数据边界校验
