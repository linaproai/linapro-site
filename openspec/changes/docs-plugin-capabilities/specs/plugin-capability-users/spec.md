## ADDED Requirements

### Requirement: Users文档说明用户领域读取能力
文档`6850-users.md` SHALL 说明`Users()`提供用户投影、搜索和可见性校验，不暴露宿主用户表、密码字段、角色关系或用户实体。

#### Scenario: 用户投影边界清晰
- **WHEN** 插件开发者阅读用户能力文档
- **THEN** 理解`UserProjection`字段、`BatchGetUsers`、`SearchUsers`和`EnsureUsersVisible`的用途

### Requirement: Users文档说明用户管理命令
文档 SHALL 说明用户状态变更通过`Admin().Users().SetUserStatus`执行，并需要领域治理上下文。

#### Scenario: 用户写入入口清晰
- **WHEN** 插件开发者需要改变用户状态
- **THEN** 文档引导其使用可信源码插件管理命令，而不是普通`Users()`能力
