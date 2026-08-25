## ADDED Requirements

### Requirement: 源码插件声明面必须是单一 Declarations

系统 SHALL 让源码插件在一个`Declarations`对象上挂嵌入文件、注册生命周期回调、绑定路由和登记定时任务。MUST NOT 要求插件作者先取七个分组门面再给同一个对象赋值。

#### Scenario: 源码插件注册 HTTP 与任务

- **WHEN** 源码插件在`Register`中声明路由和定时任务
- **THEN** 它可以直接对`Declarations`调用注册方法
- **AND** 不得再强制经过只做转发的包装结构体
