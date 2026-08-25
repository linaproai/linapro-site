## ADDED Requirements

### Requirement: 任务登记 CRUD 与调度是一个操作面

系统 SHALL 让处理函数登记、任务增删改查和调度属于同一操作面。宿主 MUST 继续通过`cron.Start`启动内建任务，以满足后端启动入口约定。`jobmeta` MUST NOT 再把 HTTP 接口枚举别名当成领域层。

#### Scenario: 启动宿主内建任务

- **WHEN** HTTP 运行时调用`cron.Start`
- **THEN** 内建任务被登记并交给调度器
- **AND** 不需要第三个包才能完成登记
