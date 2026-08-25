## ADDED Requirements

### Requirement: 处理函数登记属于任务操作面

系统 SHALL 把定时任务处理函数登记放进与 CRUD、调度同一操作面的`jobmgmt`。MUST NOT 再要求调用方为了登记处理函数而依赖独立的`jobhandler`包作为生产入口。

#### Scenario: 宿主注册内建处理函数

- **WHEN** 宿主启动注册会话清理等内建处理函数
- **THEN** 登记发生在任务操作面上
- **AND** `cron.Start`仍能启动这些内建任务
