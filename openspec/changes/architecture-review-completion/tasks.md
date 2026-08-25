## 1. 宿主与工作台契约

- [x] 1.1 从`GET /config/public/frontend`和`GetPublicFrontend`去掉`panelLayout`、`sloganImage`、`ui.layout`；工作台改用本地 preferences
- [x] 1.2 删除菜单`checkIconUnique`及创建/更新调用
- [x] 1.3 字典公开 DTO、导入导出去掉`tagStyle`/`cssClass`核心字段；工作台侧不再依赖宿主样式列
- [x] 1.4 角色授权改为菜单列表加`menuIds`；工作台自己拼树；去掉树选择/`checkedKeys`核心契约
- [x] 1.5 分析页图表去掉虚构序列；登录页不再依赖宿主布局词

## 2. 插件编解码与门面

- [x] 2.1 cache/lock/storage/data 改为 JSON 信封，删除 dedicated codec 生产入口，治理测试锁定 payload kind
- [x] 2.2 导出插件 Management/Lifecycle/Runtime 子能力；启动与控制器按需依赖
- [x] 2.3 `Declarations`去掉七个公开分组 getter；源码插件直接注册
- [x] 2.4 生产校验拒绝演示钩子`insert`/`sleep`/`error`；文档不再教学
- [x] 2.5 官网中文主稿与英文镜像将`RecordStore`标为实验能力

## 3. 基础设施

- [x] 3.1 锁调用面收成`LockStore`+`LockFunc`；单机显式传入 SQL store；去掉 nil 回落和第二套方法名
- [x] 3.2 cachecoord 修订状态实例化；折叠 revisionctrl 产品；模块/插件只暴露一份 TTL 缓存
- [x] 3.3 处理函数登记并入`jobmgmt`；保留`cron.Start`启动内建任务
- [x] 3.4 `config.Service`按领域接口组合；公开品牌读取不含布局词
- [x] 3.5 会话热存储只保留写/读/删/续期；列表与清理留在管理投影；修订注释不再提`sys_cache_revision`

## 4. 测试与验证

- [x] 4.1 增加/更新 Go 与工作台单测，驱动真实入口覆盖导航、身份、公开配置、菜单写入、字典、锁/会话/cachecoord 构造、catalog、启停、钩子、收件箱、会话存储
- [x] 4.2 更新既有 E2E（TC004/TC006/TC007、分析页、导航、收件箱），不新造 TC 编号除非`lina-e2e`要求
- [x] 4.3 运行`openspec validate architecture-review-completion --strict`与变更包 Go 测试

### 任务记录

- **DI 来源检查**：`locker.New`显式接收`LockStore`；单机由`httpstartup`传入`locker.NewSQLStore()`。`cachecoord.New`实例持有本地修订。`config.New`仍注入启动期`cachecoord`。`jobmgmt.NewRegistry()`在`httpstartup`构造，传入`jobmgmt.New`、`NewScheduler`和`cron.New`；`RegisterHostHandlers`与`AttachPluginLifecycle`使用同一 registry。无新进程全局后端。
- **i18n**：更新公开配置/菜单/字典 apidoc 源文本与`zh-CN`翻译及 packed 镜像。图标唯一错误不再作为正式入口。`jobhandler`错误码 ID 保持不变。已从`config.json`去掉布局词 seed 对应翻译。无新用户可见文案。
- **缓存一致性**：cachecoord 修订状态按实例隔离；两次`New()`互不影响。插件 TTL 仍走`kvcache`。
- **数据权限**：角色菜单编号与在线用户列表仍在查询阶段过滤。收件箱详情仍走`notify.InboxGet`。
- **开发工具跨平台**：无新脚本。linactl wasm 钩子校验与宿主一致拒绝演示动作。参数类型与工作台菜单写进`005`/`006`最终态，不另增兼容回填 SQL。
- **测试策略**：契约单测驱动真实入口；更新 TC001q/TC002a/TC004/TC006/TC007；分析页去掉虚构指标；动态插件演示钩子改为事件声明并拒绝`insert`/`sleep`/`error`上传。`cron.Start`仍启动内建任务。
- **FB-4 DI 来源检查**：无新增运行期依赖。插件控制器改回接收启动期同一份`plugin.New()`实例，类型为`pluginsvc.Service`；`httpstartup`仍通过`runtime.pluginSvc`传入。`Management`/`Startup`/`Runtime`收回为私有 facet，无新实例。
- **FB-4 i18n**：无运行时文案、菜单、API 文档源文本、插件清单或语言包影响。
- **FB-4 缓存一致性**：无缓存影响；插件服务仍复用启动期共享实例。
- **FB-4 数据权限**：无数据操作边界变化。
- **FB-4 开发工具跨平台**：无脚本或工具链影响。
- **FB-4 外部规则**：已读取`openspec.md`、`architecture.md`、`backend-go.md`、`testing.md`、`i18n.md`。`api-contract`/`database`/`plugin`/`frontend-ui`/`dev-tooling`/`data-permission`/`cache-consistency`对本反馈无影响。
- **FB-4 测试策略**：治理类修复；新增`plugin_new_test.go`锁定控制器依赖`pluginsvc.Service`并扫描其他宿主控制器构造函数。无用户可观察行为变化，不新增 E2E。
- **FB-5 DI 来源检查**：无新增运行期依赖。`notify.New`仍只接收`tenantSvc`。控制器继续从启动期`bizctx`取当前用户 ID，把校验失败交给`notify` inbox 入口。
- **FB-5 i18n**：未新增错误码机器码。公开`errorCode`仍为`USERMSG_NOT_AUTHENTICATED`，继续使用已有`error.usermsg.not.authenticated`翻译。无新语言包、菜单或 API 文档源文本。
- **FB-5 缓存一致性**：无缓存影响。
- **FB-5 数据权限**：inbox 仍按传入`userID`自隔离；查询阶段`ApplyTenantScope`未改。仅把`userID<=0`从`CodeNotifyUserNotFound`改为未登录错误。
- **FB-5 开发工具跨平台**：无脚本或工具链影响。
- **FB-5 外部规则**：已读取`openspec.md`、`architecture.md`、`backend-go.md`、`testing.md`、`i18n.md`、`data-permission.md`。`api-contract`无 DTO/路由变化且公开`errorCode`保持不变，无影响。`plugin`/`frontend-ui`/`dev-tooling`/`database`/`cache-consistency`无影响。
- **FB-5 测试策略**：新增`TestInboxMethodsRejectMissingUserID`与`TestCountRejectsMissingCurrentUser`。无用户可观察页面变化，不新增 E2E。
- **FB-6 DI 来源检查**：无新增运行期依赖。
- **FB-6 i18n**：仅补注释，无运行时文案、语言包或 API 文档源文本变化。
- **FB-6 缓存一致性**：无缓存影响。
- **FB-6 数据权限**：无数据操作变化。
- **FB-6 开发工具跨平台**：无脚本影响。
- **FB-6 外部规则**：已读取`openspec.md`、`backend-go.md`、`testing.md`、`i18n.md`。`architecture`/`api-contract`/`database`/`plugin`/`frontend-ui`/`data-permission`/`cache-consistency`/`dev-tooling`对本反馈无影响。
- **FB-6 测试策略**：治理类注释修复；静态扫描确认缺注释项清零。回归编译`usermsg`/`notify`/`locker`/`menuopen`/`cmd`。不新增 E2E。

## Feedback

- [x] **FB-1**: 删除`Declarations`分组兼容 getter，源码插件直接注册
- [x] **FB-2**: 从`dictcap`去掉`TagStyle`/`CssClass`
- [x] **FB-3**: 从`005` seed 与宿主公开配置规格去掉`loginPanelLayout`/`sloganImage`/`sys.ui.layout`
- [x] **FB-4**: 宿主 HTTP 控制器依赖目标组件默认 Service，不得依赖从 Service 拆出的子接口
- [x] **FB-5**: 站内信未登录错误码归属 notify Service，控制器不再定义 bizerr.Code
- [x] **FB-6**: 为搬迁/新增源码补齐常量、内部方法和文件用途注释
- [x] **FB-7**: 恢复 Declarations 领域 getter，不改为 embed 领域接口

### FB-7 任务记录

- **DI 来源检查**：无新增运行期依赖。声明面仍由`pluginhost.NewDeclarations`构造。
- **i18n**：无运行时文案、菜单、API 文档源文本或语言包变更。
- **缓存一致性**：无缓存影响。
- **数据权限**：无数据操作变化。
- **开发工具跨平台**：无脚本影响。
- **外部规则**：已读取`openspec.md`、`architecture.md`、`backend-go.md`、`plugin.md`、`testing.md`、`i18n.md`。`api-contract`/`database`/`frontend-ui`/`data-permission`/`cache-consistency`/`dev-tooling`对本反馈无影响。
- **测试策略**：更新`TestDeclarationsExposesDomainGettersWithoutEmbedding`；回归`pluginhost`、plugin 子包、i18n、apidoc 与 marketplace 后端单测。无用户可观察 UI 变化，不新增 E2E。
