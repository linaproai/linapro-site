## ADDED Requirements

### Requirement: 锁会话和缓存协调不得在构造后用进程全局切换后端

系统 SHALL 要求`locker`、在线会话存储和`cachecoord`在构造函数中接收部署选定的后端。生产路径 MUST NOT 使用`ConfigureCoordination`、`cachecoord.Default`或等价包内全局变量，在`New()`之后把表锁换成 Redis、把本地修订换成共享修订，或让后续`New()`继承进程脏状态。`config`读取面若依赖缓存协调，MUST 通过构造参数接收启动期同一实例。

#### Scenario: 集群启动时锁后端写在构造参数上

- **WHEN** `cluster.enabled=true`且 HTTP 启动装配构造 locker
- **THEN** 构造函数接收 coordination 提供的`LockStore`
- **AND** 生产代码不得在构造返回后再调用`ConfigureCoordination`

#### Scenario: 单机启动时会话存储不偷看全局补丁

- **WHEN** `cluster.enabled=false`且启动装配构造会话存储
- **THEN** 构造函数接收单机存储实现
- **AND** `NewDBStore()`不得根据进程级`ConfigureCoordination`结果改写行为

#### Scenario: 配置服务复用启动期 cachecoord

- **WHEN** 配置读取需要集群修订协调
- **THEN** `config`构造函数接收启动期`cachecoord.Service`
- **AND** 不得在内部调用`cachecoord.Default`创建或取出另一份进程单例
