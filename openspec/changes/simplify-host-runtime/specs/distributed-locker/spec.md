## ADDED Requirements

### Requirement: 单机表锁必须实现统一 LockStore

系统 SHALL 将`sys_locker`表锁作为`coordination.LockStore`的一种实现。`locker.Service` MUST 在构造时绑定`LockStore`；方法实现 MUST NOT 再按`store==nil`走另一套 SQL 方法名。

#### Scenario: 单机构造锁服务

- **WHEN** `cluster.enabled=false`且启动装配构造`locker.New`
- **THEN** 构造结果绑定 SQL `LockStore`
- **AND** 后续`Lock`/`Unlock`/`Renew`只通过该 store
- **AND** HTTP 启动不得再改写该后端
