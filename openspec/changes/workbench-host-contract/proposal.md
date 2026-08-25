## Why

第二套管理工作台是框架硬约束：用户使用`LinaPro`时可以换成自己的壳。当前`GET /menus/all`直接吐 Vben 路由，`pluginhost`写死`system/plugin/dynamic-page`，身份接口还带一份偏 RuoYi 的菜单树，登录再用路径猜测多租户。宿主在替默认工作台编译页面，换壳的人必须假装自己还是 Vben。

## What Changes

- **BREAKING**：宿主菜单启动接口改为通用导航资源（编号、父级、标题、路径、权限、图标、可见性、排序、打开方式、资源地址）。不再返回`#/views`、`hideInMenu`、`IFrameView`、`BasicLayout`、`keepAlive`等 Vben 路由形状。
- 默认工作台`lina-vben`在本仓库内一次迁完：自己把通用导航资源编译成 Vben 路由。不考虑仓外兼容，不保留`/menus/all`双轨。
- **BREAKING**：`GET /user/info`不再返回菜单树。身份接口返回用户、角色、权限、落地路径，以及组织/租户等能力标志。
- 登录与壳启动用能力标志判断组织/租户是否启用，禁止再从菜单 path/name 猜测。
- 插件页面（源码页与动态页一起改完）：约定目录发现`frontend/pages`，但宿主只写入通用页面资源。公开包不再出现`system/plugin/dynamic-page`，菜单不再携带 Vben 组件路径或`pluginAccessMode`这类壳专用 query。
- 插件过滤菜单时改吃稳定菜单投影，不得再传入`SysMenu`实体。
- 登录布局、侧栏导航模式等壳偏好可以继续留在`GET /config/public/frontend`；新壳自行决定是否应用。
- 菜单图标全局唯一作为通用导航规则保留在宿主校验。
- 工作台入口路径、品牌、注册开关、水印、主题明暗继续留在宿主。

## Capabilities

### New Capabilities

- `host-navigation-resource`：宿主向任意管理工作台提供的通用导航与插件页面资源契约，以及默认工作台的编译职责。

### Modified Capabilities

- `menu-management`：当前用户可访问菜单改为通用资源，不再按 Vben 路由投影。
- `user-auth`：登录后用户信息不再携带菜单树，改为能力标志。
- `plugin-ui-integration`：插件页面发现产物与打开方式改为通用资源；动态页组件路径离开宿主公开包。
- `core-host-boundary-governance`：宿主公开 HTTP 契约不得按默认 Vben 工作台的路由 JSON 或 Vue 组件路径设计。
- `module-decoupling`：组织/租户启用态必须是明确能力标志，不得从菜单路径推断。

## Impact

- 代码：`apps/lina-core/api/menu`、`api/user`、`pkg/plugin/pluginhost`、菜单控制器与插件`FilterMenus`；`apps/lina-vben`路由装配、`auth.ts`、插件页面运行时。
- API：**BREAKING** `GET /menus/all`响应形状；**BREAKING** `GET /user/info`去掉`menus`；插件菜单同步写入的页面字段。
- 数据：菜单表可继续保存打开方式与资源地址；不要求拆新表才能起步。图标唯一校验保留。
- 测试：菜单启动、登录落地、插件启停后侧栏刷新、组织/租户能力显隐相关单测与 E2E。
- i18n：若身份/菜单字段对前端文案有影响，随工作台编译层评估；宿主公开配置布局词本变更不撤。
- 数据权限：菜单过滤改投影后仍必须在授权范围内返回资源。
- 缓存：菜单/插件启用快照若随投影结构调整，复用现有失效路径。
- 依赖：必须在 `simplify-host-framework` 纲领之下实施；建议先于 `startup-di-backends` 合入。
