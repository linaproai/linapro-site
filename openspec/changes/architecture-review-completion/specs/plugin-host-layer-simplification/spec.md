## ADDED Requirements

### Requirement: cache lock storage data 必须走 JSON 信封

系统 SHALL 将缓存、锁、对象存储和数据访问宿主方法改为 JSON 信封。这些领域 MUST NOT 再保留 dedicated 二进制编解码作为生产路径。catalog 中对应方法的`PayloadKind` MUST 为 JSON。

#### Scenario: 动态插件读取缓存

- **WHEN** 动态插件调用 cache.get
- **THEN** guest 与 WASM 分发都使用 JSON 请求/响应
- **AND** 仓库中不得再存在该领域的 dedicated codec 生产入口

#### Scenario: 一侧仍使用 dedicated codec

- **WHEN** catalog 将 cache/lock/storage/data 标为 JSON 但分发仍解码 binary
- **THEN** 治理测试或包测试失败

### Requirement: 源码插件声明面通过领域方法引用领域接口

系统 SHALL 在`pluginhost.Declarations`上提供`Assets()`、`Lifecycle()`、`Hooks()`、`HTTP()`、`Jobs()`、`Providers()`、`Access()`领域方法，分别返回对应的领域声明接口。`Declarations` MUST NOT 通过 embed 领域接口提升注册方法。源码插件 MUST 通过这些领域方法引用对应领域接口上的注册方法。

#### Scenario: 源码插件注册路由和任务

- **WHEN** 源码插件调用`NewDeclarations`
- **THEN** 它通过`HTTP()`获得`HTTPDeclarations`并注册路由
- **AND** 通过`Jobs()`获得`JobDeclarations`并登记任务
- **AND** `Declarations`接口不嵌入上述领域接口
