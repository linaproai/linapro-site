## ADDED Requirements

### Requirement: 主框架减肥第一波必须拆成可独立合入的子变更

系统 SHALL 将主框架减肥第一波限制为三件战略，并分别由独立 OpenSpec 变更落地：`workbench-host-contract`、`startup-di-backends`、`plugin-hostservice-catalog`。纲领变更`simplify-host-framework` MUST NOT 修改生产代码。合入顺序 MUST 为工作台契约、启动期依赖注入、动态插件 catalog 生成。

#### Scenario: 实现进入错误的变更目录

- **WHEN** 开发者需要修改`GET /menus/all`、身份菜单树或插件页面资源契约
- **THEN** 这些修改 MUST 落在`workbench-host-contract`
- **AND** MUST NOT 作为`simplify-host-framework`的实现任务

#### Scenario: 三个子变更可单独回滚

- **WHEN** `startup-di-backends`合入后需要回滚
- **THEN** 回滚 MUST NOT 强制同时回滚已合入的`workbench-host-contract`
- **AND** MUST NOT 要求尚未合入的`plugin-hostservice-catalog`一并处理

### Requirement: 第一波必须排除非战略重构

系统 SHALL 把下列工作排除在第一波之外：插件总门面大拆、锁/缓存/任务包合并、`usermsg`收口、源码插件声明门面压缩、仪表盘假数据清理、验证码/二维码死路由清理、`RecordStore`功能扩展或删除、存量 dedicated 二进制编解码退役。

#### Scenario: 审查文档第六块被提议并入第一波

- **WHEN** 开发者计划在第一波中压扁`plugin.Service`或合并`cron`/`jobmgmt`
- **THEN** 该工作 MUST 被拒绝或另开变更
- **AND** MUST NOT 写入三个战略子变更的实现任务

### Requirement: 第二波由 simplify-host-runtime 承接可立即实施的排除项

系统 SHALL 在第一波完成后由`simplify-host-runtime`承接权限投影、单插件启停、演示钩子、声明面、表锁`LockStore`、收件箱、会话核心存储、分析页假数字、死登录路由和`RecordStore`实验标注。`cron`/`jobmgmt`合并、`plugin.Service`拆包、dedicated 编解码退役仍 MUST 排除。

#### Scenario: 第二波实现进入正确目录

- **WHEN** 开发者需要把收件箱详情并入`notify`
- **THEN** 修改 MUST 落在`simplify-host-runtime`
- **AND** MUST NOT 回写第一波三个战略子变更

### Requirement: 可替换管理工作台是第一波硬约束

系统 SHALL 将用户自定义管理工作台视为框架硬约束。默认`lina-vben`只是本仓库提供的参考壳。第一波 MUST 由`workbench-host-contract`把宿主导航与插件页面契约改为通用资源，并在本仓库内一次迁移默认工作台。

#### Scenario: 为仓外未知壳保留 Vben 路由双轨

- **WHEN** 开发者提议继续让宿主产出 Vben`RouteRecord`以便兼容未知壳
- **THEN** 该双轨 MUST 被拒绝
- **AND** 默认工作台 MUST 在同一子变更中改为消费通用导航资源
