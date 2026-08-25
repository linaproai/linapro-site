## Why

主框架减肥第一波已经把工作台编译器请出宿主契约、把锁/会话/缓存改成构造期注入，并用 catalog 推导动态插件方法表。审查文档第四到第六块仍在：权限菜单继续泄漏`SysMenu`、启用插件会重扫全部清单、生产分发仍解释演示钩子、`usermsg`叠在`notify`之上、单机表锁还走另一套方法名、分析页和死登录路由仍是模板残留。这些会让第一波刚拉开的宿主/工作台边界重新缠回去。

## What Changes

- 权限菜单过滤改吃稳定`FilterItem`投影，禁止跨模块传递`SysMenu`。
- 启用/禁用只同步当前插件的目标清单，不再扫描并逐个同步全部已发现清单。
- 演示用`insert`/`sleep`/`error`钩子动作移出生产分发；测试用假注册器或测试专用入口。
- 源码插件`Declarations`改为单一声明面：挂资源、注册回调、绑路由、登记任务。分组门面不再作为作者必须经过的一层。
- 单机表锁实现`coordination.LockStore`。`locker.New`构造时绑定后端，不再在方法里分叉 SQL 与 coordination 两套路径。
- 当前用户收件箱改由`notify`提供详情读取；`usermsg`服务退回控制器改写。HTTP 路径保持不变。
- 会话热状态存储只保留写入、读取、删除、续期；带数据权限的在线用户列表留在会话管理投影，不再塞进通用`Store`核心方法集。
- 默认工作台分析页去掉虚构指标数字；删除已关闭入口的验证码/二维码登录路由。
- 动态插件`RecordStore`标成可选实验能力，从主路径插件文档降级。不删除实现，不扩展查询计划器。
- 不退役存量 dedicated 二进制编解码，不合并`cron`/`jobmgmt`/`jobhandler`（`backend-go`仍要求`cron`作为宿主定时任务入口），不把`config.Service`再切成窄接口，不撤销第一波对图标唯一和公开布局词的保留决定。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `plugin-ui-integration`：权限菜单过滤必须使用稳定投影。
- `plugin-manifest-lifecycle`：状态变更只同步当前插件目标清单。
- `plugin-hook-slot-extension`：生产分发不得解释演示钩子动作。
- `plugin-host-layer-simplification`：源码插件声明面收成单一`Declarations`。
- `distributed-locker`：表锁作为`LockStore`实现，构造期绑定。
- `user-message`：收件箱详情走`notify`，不再经独立`usermsg`服务查表。
- `online-user`：带范围列表不属于通用会话存储核心契约。
- `dashboard-workbench`：分析页不得展示虚构业务数字。
- `login-page-presentation`：不再注册验证码/二维码登录路由。
- `plugin-data-service`：`RecordStore`为实验能力，官方文档不得把它写成主路径。
- `host-framework-simplification-program`：第二波范围与排除项。

## Impact

- 代码：`apps/lina-core`插件集成/生命周期/`pluginhost`/`locker`/`notify`/`session`；`apps/lina-vben`分析页与认证路由；`pkg/plugin` README。
- API：用户消息 HTTP 路径不变。登录子路由`/auth/code-login`、`/auth/qrcode-login`从默认工作台消失。
- 插件作者：源码插件可直接在`Declarations`上注册，不必再经过七个分组门面。
- 数据：无新表。`sys_locker`仍作单机锁实现。
- 测试：插件过滤/启停、锁构造、收件箱详情、分析页、登录路由相关单测与既有 E2E（沿用 TC 编号）。
- i18n：分析页若去掉虚构数字，保留指标标题键；删除死登录路由后可停止把它们当正式入口。收件箱分类文案仍走既有`notify.category.*`。
- 数据权限：收件箱详情与在线用户列表仍必须在查询阶段约束到当前用户或角色范围。
- 缓存：启用快照仍按单插件更新加刷新，不改为全量重扫清单。
- 依赖：必须在`simplify-host-framework`纲领之下作为第二波实施；第一波三个子变更保持不动。
