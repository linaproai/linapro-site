## ADDED Requirements

### Requirement: 第三波按审查文档覆盖先前冻结项

系统 SHALL 在`architecture-review-completion`落地审查文档中仍未完成的宿主/工作台拆分、插件编解码与门面、以及基础设施合并。先前第一波/第二波对公开布局词、图标唯一、dedicated codec、插件总门面、任务包合并、配置拆面、缓存产品合并的排除项 MUST 被本变更覆盖。`cluster`与`coordination`、`config`与`sysconfig`仍 MUST 分开。`hostlock`、表锁和`RecordStore`实现 MUST 保留。

#### Scenario: 以冻结排除为由拒绝去掉布局词

- **WHEN** 开发者引用`workbench-host-contract` D4 保留`panel-left`
- **THEN** 本变更要求从宿主公开契约删除该布局词
- **AND** 审查文档优先于该冻结项

#### Scenario: 保留 cron 启动入口

- **WHEN** 任务登记并入`jobmgmt`
- **THEN** 宿主仍通过`cron.Start`启动内建任务
- **AND** 不得删除该启动入口
