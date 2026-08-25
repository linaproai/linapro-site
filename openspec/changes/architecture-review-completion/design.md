## Context

第一波`workbench-host-contract`、`startup-di-backends`、`plugin-hostservice-catalog`和第二波`simplify-host-runtime`已经把通用导航、身份能力标志、构造期绑定和 catalog 推导落到代码。它们同时把公开布局词、图标唯一、dedicated codec、插件总门面、任务/配置/缓存包合并写成冻结排除项。审查文档`localdocs/linapro-framework-architecture-review.md`是本目标的权威，覆盖这些冻结项。`cluster`与`coordination`、`config`与`sysconfig`、`hostlock`、表锁和`RecordStore`实现继续保留。

## Goals / Non-Goals

**Goals:**

- 宿主公开契约回到通用资源：导航、身份、公开配置、字典、角色菜单编号。
- 默认工作台编译布局词、图标审美、标签颜色、树勾选和分析页展示。
- 动态插件 cache/lock/storage/data 只走 JSON；catalog 仍是唯一方法规格。
- 插件根包用私有 facet 组合默认`Service`；HTTP 控制器和启动上下文依赖该`Service`。声明面不再要求七个分组 getter。
- 锁、缓存、任务、配置读取面按真实后端合并，构造期绑定，无进程全局补丁。

**Non-Goals:**

- 合并`cluster`与`coordination`，或`config`与`sysconfig`。
- 删除`hostlock`、`sys_locker`或`RecordStore`实现；不扩展查询计划器。
- 建设第二套非 Vben 工作台，或为仓外壳保留双轨`GET /menus/all`。
- 让每个宿主模块都按同一套前端禁用规则隐藏。
- 仅为别名/小文件清理而改行为。

## Decisions

### D1：审查文档覆盖冻结保留项

第一波 D3/D4 和第二波排除项在本变更中作废。增量规范必须写明「撤销」而不是静默改代码。

备选：继续遵守冻结项，只补测试。否决：与目标验收标准冲突。

### D2：公开配置去掉布局词，工作台用本地偏好

从`GET /config/public/frontend`和`GetPublicFrontend`投影中删除`panelLayout`、`sloganImage`、`ui.layout`。`sys_config`里已有键可以留在管理面，但不再进入宿主必填公开契约。默认工作台登录页和水印布局使用 Vben 本地 preferences 默认值。主题明暗与水印开关仍属通用展示，保留。

备选：另开工作台专用未登录接口。否决：会再造一条壳契约。

### D3：图标唯一退出宿主完整性

删除`checkIconUnique`。父级/名称冲突仍由菜单服务校验。工作台若要避免侧栏同图标，在壳里处理，不写回宿主。

### D4：字典核心去掉展示字段

公开 DTO、导入导出模板不再暴露`tagStyle`/`cssClass`。表列可保留以免强制迁移。工作台用本地映射或不再依赖宿主颜色字段。

### D5：角色授权只给编号列表

停用或改写`GET /menu/role/{roleId}`和`GET /menu/treeselect`的树选择/`checkedKeys`形状。角色详情已有`menuIds`。工作台用通用菜单列表自己拼树。不得为此把菜单全量先查再在内存里丢授权外节点。

### D6：cache/lock/storage/data 改 JSON

沿用已有`callHostServiceJSONRequest`/`decodeCapabilityJSONRequest`/`capabilityJSONResponse`。删除对应`protocol_hostservice_{cache,lock,storage,data}_codec.go`。catalog 将这些方法标为`PayloadKindJSON`。治理测试禁止再为这四个领域保留 dedicated allowlist。runtime/network 等其它存量 dedicated 本波不强制退役。

### D7：插件总门面压扁但不拆包

在`internal/service/plugin`根包保留一份组合实现，用私有 facet 组合默认`Service`。HTTP 启动上下文和 HTTP 控制器依赖`plugin.Service`，不得为了读列表或启停把拆出的子接口作为控制器依赖。不把插件包拆成多个顶层`internal/service`包。

声明面：`Declarations`通过`Assets()`/`Lifecycle()`/`Hooks()`/`HTTP()`/`Jobs()`/`Providers()`/`Access()`引用对应领域接口，不得 embed 这些领域接口。源码插件经领域方法再调用注册 API。

### D8：锁调用面就是 LockStore

`locker.Service`收成`LockStore`加`LockFunc`（若仍需要闭包助手）。删除按 SQL id 的`Unlock`/`Renew`与按 name 的第二套方法。插件构造只收一份锁依赖。`httpstartup`单机显式传入`locker.NewSQLStore()`，禁止`New(nil)`再暗中选表锁。`hostlock`继续用同一`LockStore`。

### D9：一份 TTL 缓存产品

插件与模块的 KV/TTL 入口只暴露`kvcache`（或等价命名空间 TTL API），内部使用注入的 coordination KV 或进程内存。`revisionctrl`并入`cachecoord`，不再作为并列产品。`cachecoord`修订状态放在实例上，删除`processLocalRevisions`进程表。`setCoordination`内联进`New`。

### D10：任务操作面与 cron.Start

`jobhandler.Registry`并入`jobmgmt`。`jobmeta`类型迁入`jobmgmt`或同包领域类型，不再反向依赖`api/job/v1`别名。`cron`保留为宿主启动入口：注册内建任务、预热、调用`jobmgmt`调度。不把`Start`从`service/cron`删掉，以满足`backend-go`。

### D11：配置读取面按领域组合

`config.Service`改为嵌入领域接口（Cluster、Tokens、Session、Login、I18n、Jobs、Plugins、Upload、Workspace、PublicFrontend、Raw、RuntimeParams）。组合根仍可传入完整实现；非控制器调用方可依赖用到的领域接口。HTTP 控制器必须依赖完整`config.Service`。公开前端领域不含 Vben 布局词。

### D12：会话热存储变窄

`session.Store`：`Set`/`Get`/`Delete`/`Touch`（及认证需要的按用户删除）。`Count`、带数据权限的列表、清理不活跃会话放到`Directory`管理投影。热存储构造返回`Store`，不要把管理投影类型泄漏成热路径参数。

## Risks / Trade-offs

- [动态插件线协议破坏] → 仓内动态插件与 guest SDK 同步改 JSON；治理测试锁死 payload kind。
- [登录页不再吃 sys_config 布局键] → 工作台本地默认；更新 TC004/TC006/TC007，避免假绿。
- [字典颜色丢失] → 工作台本地映射；列可留在表里。
- [角色树接口删除影响工作台抽屉] → 同期改`role-drawer.vue`用列表+`menuIds`。
- [locker 方法名变化] → 编译失败暴露所有调用点，按`LockStore`改。
- [任务包合并触达面大] → 先搬 Registry 与类型，保留`cron.Start`。

## Migration Plan

1. 先改宿主契约测试，再改实现，再改默认工作台。
2. 动态插件四个领域同时改 guest 与 WASM，不允许一侧 JSON 一侧 binary。
3. 无数据库强制 DROP。不恢复`sys_cache_revision`。
4. 回滚：恢复公开 DTO 与 codec 文件；本变更不提供双轨协议。

## Open Questions

- 无。审查文档与验收标准已覆盖冻结项覆盖、cron 入口保留和 RecordStore 不删除。
