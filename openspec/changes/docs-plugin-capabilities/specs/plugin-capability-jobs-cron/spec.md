## ADDED Requirements

### Requirement: JobsCron文档说明任务读取和管理能力
文档`8100-jobs-cron.md` SHALL 说明`Jobs()`读取任务投影，`Admin().Jobs()`执行任务或修改任务状态。

#### Scenario: 任务运行期入口清晰
- **WHEN** 插件开发者阅读任务能力文档
- **THEN** 理解`BatchGetJobs`、`RunJob`和`SetJobStatus`的职责

### Requirement: JobsCron文档说明Cron注册入口
文档 SHALL 说明源码插件通过`SourcePlugin.Cron()`注册定时任务，动态插件通过`cron.register`在发现阶段提交`CronContract`。

#### Scenario: Cron注册边界清晰
- **WHEN** 插件开发者需要声明插件定时任务
- **THEN** 文档区分源码插件注册门面和动态`hostServices.cron.register`
