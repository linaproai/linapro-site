## Why

审查文档`localdocs/linapro-framework-architecture-review.md`要求宿主停止替 Vben 编译页面、锁和缓存在构造时说清后端、动态插件宿主方法只维护一份规格。第一波和第二波已经落地通用导航、身份能力标志、构造期绑定、catalog 推导、单插件启停、演示钩子跳过、表锁`LockStore`、收件箱并入`notify`、分析页概览清零和死登录路由删除。仍留在代码里的，是第一波/第二波明确冻结的保留项：公开配置布局词、图标全局唯一、字典展示字段、控制台形角色树接口、dedicated 编解码、插件总门面、声明分组门面、锁/缓存/任务/配置分层，以及官网仍把`RecordStore`写成主路径。本变更以审查文档为目标，覆盖这些冻结项。

## What Changes

- **审查覆盖冻结规则**：`workbench-host-contract` D4/D3 保留的公开布局词与图标全局唯一、`plugin-hostservice-catalog` 保留的 dedicated codec、`simplify-host-runtime` 排除的总门面/任务包合并/配置拆面/`cachecoord`/`kvcache` 合并，一律按审查文档改掉。本变更是对该覆盖的显式记录，不是静默回滚。
- **BREAKING**：`GET /config/public/frontend`去掉宿主必填的 Vben 布局词：`panelLayout`（`panel-left` 等）、`ui.layout`（`sidebar-mixed-nav` 等）、`sloganImage`。保留品牌、语言、注册/忘记密码开关、主题明暗、水印、工作台入口路径。
- 菜单写入不再把图标全局唯一当作数据完整性规则。
- **BREAKING**：字典核心契约只保留值与标签（外加排序/状态/备注等通用字段），公开 DTO、导入导出不再把`tagStyle`/`cssClass`当作核心模型。
- **BREAKING**：角色授权不再通过树选择/`checkedKeys`页面形状接口交付。宿主返回菜单资源列表和角色`menuIds`；默认工作台自己拼树和勾选。
- 默认工作台继续编译通用导航；分析页图表去掉虚构业务序列；登录布局改由工作台本地偏好承担。
- **BREAKING**：缓存、锁、对象存储、数据访问的 dedicated 二进制编解码退役，改走 JSON 信封。catalog 为唯一方法规格；一侧改名仍必须让治理测试失败。
- 插件根包仍构造一份实现，HTTP 启动与控制器依赖默认`plugin.Service`，不得把拆出的管理/启动/运行时子接口作为控制器依赖。源码插件`Declarations`不再要求七个分组门面/getter。
- 生产校验拒绝演示钩子`insert`/`sleep`/`error`；官网与`pkg/plugin`文档把`RecordStore`标成实验能力。
- 锁调用面收成`LockStore`（表锁与 Redis/内存是两种实现）。保留`hostlock`与`sys_locker`。
- 模块/插件使用一份带命名空间和 TTL 的缓存产品；修订转发器不再作为并列产品。
- 定时任务处理函数登记、CRUD、调度收成一个操作面；宿主仍通过`cron.Start`启动内建任务。
- `config`读取面按领域拆开（集群、令牌、会话、登录、i18n、任务、插件、上传、工作台路径、公开品牌）。`cluster`与`coordination`、`config`与`sysconfig`继续分开。
- 会话热存储只保留写入/读取/删除/续期；带范围的在线用户列表留在管理投影。`sys_cache_revision`不得再作为活模型或注释路径。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `host-navigation-resource`：撤销图标全局唯一与公开布局词保留；布局词离开宿主公开契约。
- `core-host-boundary-governance`：宿主公开契约不得要求 Vben 布局词、字典标签样式或角色树勾选形状。
- `login-page-presentation`：登录面板位置与 slogan 改由默认工作台本地偏好承担。
- `dict-management`：字典核心为值+标签，不再把`tagStyle`/`cssClass`当核心字段。
- `menu-management`：图标唯一不再是宿主完整性规则；角色授权不再返回树选择/`checkedKeys`。
- `role-management`：角色菜单授权只暴露菜单编号列表。
- `dashboard-workbench`：分析页不得展示虚构图表序列。
- `plugin-host-layer-simplification`：退役 cache/lock/storage/data dedicated codec；声明面去掉七个必经分组门面。
- `plugin-host-service-extension`：上述领域方法改为 JSON-only。
- `plugin-cache-service`：cache host service 走 JSON。
- `plugin-lock-service`：lock host service 走 JSON。
- `plugin-storage-service`：storage host service 走 JSON。
- `plugin-data-service`：data host service 走 JSON；`RecordStore`为实验能力，官网不得写成主路径。
- `plugin-hook-slot-extension`：生产校验拒绝演示动作。
- `plugin-service-layout`：HTTP 控制器和启动上下文依赖默认`plugin.Service`；根包用私有 facet 组合，不把子接口暴露给控制器。
- `distributed-locker`：去掉第二套 locker 方法名；表锁实现`LockStore`。
- `distributed-cache-coordination`：一份带命名空间 TTL 的缓存产品；修订状态不得是进程全局表。
- `cron-handler-registry`：处理函数登记并入任务操作面。
- `cron-job-management`：CRUD 与调度属于同一操作面；`cron.Start`仍是宿主启动入口。
- `session-hot-state`：热存储只保留写/读/删/续期。
- `config-management`：运行期读取面按领域拆分，公开前端配置不含 Vben 布局词。
- `host-framework-simplification-program`：第三波覆盖原先排除的冻结项，并以审查文档为准。

## Impact

- 代码：`apps/lina-core`公开配置/菜单/字典/角色/锁/会话/缓存/任务/配置/插件包与`pkg/plugin`；`apps/lina-vben`公开配置客户端、字典标签、角色树、分析页、登录偏好；`apps/lina-site`插件文档；`hack/tests`既有 TC。
- API：**BREAKING** 公开前端配置字段；字典 DTO；角色菜单树/treeselect 形状；动态插件 cache/lock/storage/data 线协议。
- 数据：不强制删`tag_style`/`css_class`/`sys_locker`列。不恢复`sys_cache_revision`。
- 测试：契约单测必须驱动真实入口；更新 TC004/TC006/TC007 等既有 E2E，不新造 TC 编号除非`lina-e2e`要求新文件。
- i18n：公开配置/字典/菜单 apidoc 源文本与`zh-CN`翻译；图标重复错误若不再返回可停止当正式入口。
- 数据权限：角色菜单列表、在线用户列表、收件箱详情仍在查询阶段约束。
- 缓存：cachecoord 实例状态不得跨`New()`共享；插件缓存走同一 TTL 产品。
- 依赖：在`simplify-host-framework`纲领下作为第三波；不合并`cluster`与`coordination`，不合并`config`与`sysconfig`，不删除`hostlock`/`RecordStore`实现。
