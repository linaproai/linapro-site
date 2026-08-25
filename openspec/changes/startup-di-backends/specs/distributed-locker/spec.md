## ADDED Requirements

### Requirement: 锁服务必须在构造时绑定 LockStore

系统 SHALL 让宿主`locker.Service`在构造时绑定唯一的`LockStore`实现。单机 MAY 绑定 SQL 表锁实现，集群 MUST 绑定 coordination 锁实现。运行期方法 MUST 只使用构造时注入的 store，MUST NOT 读取包级可变后端指针。

#### Scenario: 测试创建独立锁服务

- **WHEN** 单测构造 locker 并传入内存或假`LockStore`
- **THEN** 该实例只使用传入的 store
- **AND** 不受其他测试对生产全局补丁的影响

#### Scenario: 集群节点获取锁走注入的 coordination store

- **WHEN** 集群模式节点调用 locker 获取锁
- **THEN** 请求进入构造时注入的 coordination`LockStore`
- **AND** 不得回退到未注入的 SQL 表锁
