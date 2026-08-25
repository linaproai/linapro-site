## ADDED Requirements

### Requirement: 锁调用面就是 LockStore

系统 SHALL 让表锁与 Redis/内存锁实现同一个`LockStore`契约。生产调用方 MUST 使用`Acquire`/`Renew`/`Release`/`IsHeld`这一套方法，MUST NOT 再维持按 SQL id 解锁和按 name 解锁两套方法名。`locker.New` MUST 接收已构造的 store；单机启动 MUST 显式传入 SQL store，不得靠`nil`回落再`NewSQLStore()`。`hostlock`继续存在并复用同一 store。

#### Scenario: 单机构造后再启动 HTTP

- **WHEN** 测试构造`locker.New(sqlStore)`后再执行 HTTP 启动装配
- **THEN** 锁后端仍是构造时的 SQL store
- **AND** 不得被启动逻辑换成另一实现

#### Scenario: 集群与单机方法名一致

- **WHEN** 调用方获取或释放一把锁
- **THEN** 单机表锁和 Redis 锁使用同一方法名
- **AND** 不存在只对 SQL 生效的 id 解锁入口
