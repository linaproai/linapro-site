## 1. 插件投影、启停与声明面

- [x] 1.1 `FilterPermissionMenus`与`ShouldKeepPermission`改为`menu.FilterItem`，更新`role`与插件集成调用点
- [x] 1.2 启用/禁用只同步当前插件目标清单，禁止全集`ScanManifests`+`SyncManifest`
- [x] 1.3 生产钩子分发移除`insert`/`sleep`/`error`；测试保留假注册器覆盖
- [x] 1.4 `Declarations`嵌入各注册方法，删除只转发的分组包装结构体，更新本仓库源码插件注册

## 2. 锁、收件箱、会话存储

- [x] 2.1 表锁实现`LockStore`；`locker.New`构造期绑定，方法不再按 nil store 分叉
- [x] 2.2 `notify`增加当前用户可见详情读取；`usermsg`控制器改为改写层并删除`internal/service/usermsg`
- [x] 2.3 会话`Store`核心方法与带数据权限的管理查询分离

## 3. 默认工作台与文档

- [x] 3.1 分析页去掉虚构指标数字，保留标题与图表结构
- [x] 3.2 删除`code-login`/`qrcode-login`路由，更新 TC006
- [x] 3.3 `pkg/plugin`中英文 README 将`RecordStore`标为实验能力

## 4. 验证

- [x] 4.1 运行变更包 Go 测试与宿主启动绑定测试
- [x] 4.2 更新或沿用相关 E2E（分析页 TC001、登录 TC006）
- [x] 4.3 运行`openspec validate simplify-host-runtime --strict`

### 任务记录

- **DI 来源检查**：`usermsg.NewV1`改为接收已有`bizCtxSvc`、`notifySvc`、`i18nSvc`，均来自`httpstartup.newHTTPRuntime`启动期实例。`locker.NewSQLStore()`在`locker.New(nil)`构造时绑定，不是进程全局补丁。`notify.InboxGet`无新缓存敏感单例。
- **i18n**：新增`NOTIFY_INBOX_NOT_FOUND`（`error.notify.inbox.not.found`），已写入`manifest/i18n/en-US/error.json`与`zh-CN/error.json`。分析页去掉虚构数字，保留既有指标标题键。登录死路由不再当正式入口。`make i18n.check`消息覆盖通过；frontend-keys 中`notice-preview-modal.vue`的插件键缺失为既有问题，本变更未改该文件。
- **缓存一致性**：启停改为`GetDesiredManifest`+`SyncManifest`当前插件，启用快照仍由后续`SetPluginEnabledState`更新，不新增缓存域。
- **数据权限**：`notify.InboxGet`在查询阶段按投递用户和租户约束。在线用户带范围列表留在`session.Directory`，认证热路径只用`session.Store`。
- **开发工具跨平台**：无新脚本。
- **测试策略**：`go test`覆盖 menu/role/locker/session/notify/plugin/integration/lifecycle/pluginhost/httpstartup/capabilityhost/auth/catalog/runtime。E2E 沿用 TC001（标题仍在）并更新 TC006；本环境未跑 Playwright。
- **`make lint dir=apps/lina-core plugins=0`**：通过。
- **`openspec validate simplify-host-runtime --strict`**：通过。
