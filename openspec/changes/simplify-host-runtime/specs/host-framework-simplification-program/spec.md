## ADDED Requirements

### Requirement: 主框架减肥第二波由 simplify-host-runtime 落地

系统 SHALL 将审查文档第四到第六块中可立即实施的部分放入`simplify-host-runtime`。第二波 MUST 覆盖权限投影、单插件启停同步、演示钩子移出生产分发、声明面压缩、表锁`LockStore`、收件箱并入`notify`、会话核心存储与管理投影分离、分析页假数字和死登录路由，以及`RecordStore`实验标注。

#### Scenario: 第二波进入错误的变更目录

- **WHEN** 开发者需要把`FilterPermissionMenus`改为投影
- **THEN** 这些修改 MUST 落在`simplify-host-runtime`
- **AND** MUST NOT 写入第一波三个战略子变更

### Requirement: 第二波必须排除高风险包合并

系统 SHALL 把下列工作排除在第二波之外：`plugin.Service`拆成多包、`cron`/`jobmgmt`/`jobhandler`合并、`config.Service`窄接口拆分、`cachecoord`/`kvcache`包合并、dedicated 编解码退役、`RecordStore`删除、撤销图标唯一和公开布局词。

#### Scenario: 提议在第二波合并 cron 与 jobmgmt

- **WHEN** 开发者计划删除`service/cron`并把`Start`搬进`jobmgmt`
- **THEN** 该工作 MUST 被拒绝
- **AND** 必须继续遵守`backend-go`对`cron.Start`入口的要求
