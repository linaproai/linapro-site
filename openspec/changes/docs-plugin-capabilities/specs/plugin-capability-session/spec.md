## ADDED Requirements

### Requirement: Session文档说明在线会话投影读取
文档`7800-session.md` SHALL 说明`Sessions()`提供在线会话搜索和批量读取投影，不暴露会话存储表或`JWT`内部实现。

#### Scenario: 会话投影可查阅
- **WHEN** 插件开发者阅读在线会话能力文档
- **THEN** 理解会话投影字段和`SearchSessions`、`BatchGetSessions`的用途

### Requirement: Session文档说明会话吊销管理命令
文档 SHALL 说明会话吊销通过`Admin().Sessions().RevokeSession`执行，并由宿主即时影响对应令牌。

#### Scenario: 会话吊销入口清晰
- **WHEN** 插件开发者需要吊销会话
- **THEN** 文档引导其使用可信源码插件管理命令，而不是普通`Sessions()`能力

### Requirement: Session文档说明动态插件关系
文档 SHALL 说明动态`hostServices`目录没有独立`session`服务，动态请求的当前身份和会话标识来自桥接请求信封。

#### Scenario: 动态会话边界清晰
- **WHEN** 动态插件开发者需要当前会话信息
- **THEN** 文档引导其读取请求信封`Identity`而不是声明`service: session`
