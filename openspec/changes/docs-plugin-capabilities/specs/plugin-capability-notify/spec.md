## ADDED Requirements

### Requirement: Notify文档说明通知领域三类入口
文档`7300-notify.md` SHALL 说明通知能力分为普通`Notifications()`投影读取、可信源码插件`Admin().Notifications()`管理命令和动态插件`notify.send`宿主服务。

#### Scenario: 通知入口可区分
- **WHEN** 插件开发者阅读基本介绍
- **THEN** 理解普通能力只读，源码插件发送和删除通知走管理命令，动态插件发送通知走`notify.send`

### Requirement: Notify文档说明发送数据模型
文档 SHALL 说明`SendInput`和`SendResult`的关键字段，以及`SourceType`、`CategoryCode`的分类语义。

#### Scenario: 发送模型清晰
- **WHEN** 插件开发者查看数据模型
- **THEN** 理解收件人、来源、标题、内容、分类、发送者、消息标识和投递数量的用途

### Requirement: Notify文档说明动态notify授权
文档 SHALL 说明动态`notify`服务使用`resources[].ref`表达可发送的消息场景或分类。

#### Scenario: 动态通知授权清晰
- **WHEN** 动态插件开发者声明`service: notify`
- **THEN** 文档要求声明`send`方法和资源引用
