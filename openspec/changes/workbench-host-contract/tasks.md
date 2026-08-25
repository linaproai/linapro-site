## 1. 宿主通用导航契约

- [x] 1.1 定义通用导航/页面资源 DTO 与打开方式枚举，去掉 Vben 路由元字段
- [x] 1.2 改造当前用户导航接口：停止`#/views`补全、`BasicLayout`/`IFrameView`编译
- [x] 1.3 菜单写入把页面资源视为不透明地址；保留图标全局唯一校验
- [x] 1.4 插件菜单同步按约定目录发现，但只写入通用页面资源
- [x] 1.5 从`pluginhost`移除`system/plugin/dynamic-page`等默认壳组件常量
- [x] 1.6 `FilterMenus`改为稳定菜单投影，禁止跨模块传递`SysMenu`

## 2. 身份与能力标志

- [x] 2.1 `GET /user/info`去掉`menus`，保留身份、角色、权限与`homePath`
- [x] 2.2 身份或能力接口返回组织/租户启用标志
- [x] 2.3 默认工作台删除按菜单 path/name 猜测租户的逻辑

## 3. 默认工作台编译器

- [x] 3.1 在`lina-vben`实现通用导航资源到 Vben 路由的编译
- [x] 3.2 动态页壳组件路径与内嵌 query 仅由工作台持有
- [x] 3.3 公开前端配置中的布局偏好继续读取，不作为换壳前置
- [x] 3.4 更新工作台启动、插件启停侧栏刷新与登录落地相关测试

## 4. 验证

- [x] 4.1 覆盖变更包的 Go 测试与宿主启动绑定测试
- [x] 4.2 更新相关 E2E（沿用现有用例编号，遵循`lina-e2e`）
- [x] 4.3 运行`openspec validate workbench-host-contract --strict`

### 任务记录

- **DI 来源检查**：`user.NewV1`新增`tenantspi.Service`参数；实例来自`httpstartup`已有的`tenantSvc`（`tenantspi.New`），与菜单/认证共用同一启动期实例。无新增缓存敏感单例。
- **i18n**：更新了`GET /menus/all`与`GET /user/info`的英文 DTO/`dc`，并同步`zh-CN` apidoc 与 packed 镜像。无新运行时 UI `$t`键。
- **缓存一致性**：菜单过滤仍走既有插件启用快照，无新缓存域。
- **数据权限**：当前用户导航仍按角色菜单 ID 过滤后返回，不先查全量再丢弃授权外数据。
- **开发工具跨平台**：无新脚本。
- **测试策略**：`menuopen`、菜单/用户控制器、plugin FilterMenus、apidoc i18n、工作台`compile-nav-resources`单测已跑。E2E 夹具改为通用导航资源形状（沿用 TC002/TC008/TC011/TC015/TC016）。未在本环境跑 Playwright。
- **`internal/cmd`**：`go test ./internal/cmd -count=1`中`TestProductionPanicsMatchAllowlist`失败指向`linapro-plugin-marketplace/backend/plugin.go` init panic 扫描，与本次导航契约无关。路由构造包可编译。
- **FB-2**：宿主 GET 对 `system/plugin/dynamic-page` 仍做契约清洗（不是双轨 API）。工作台按 `getPluginPageByRoute` 编译源码页；`#` 前缀不再原样当作已编译组件。`plugin.yaml` 与 E2E 夹具不再把工作台壳路径写成宿主 resource。`make lint dir=apps/lina-core plugins=0` 通过。`go test ./pkg/menuopen ./internal/controller/menu ./internal/controller/user ./internal/cmd/internal/httpstartup ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/frontend` 通过。Vitest `compile-nav-resources.test.ts` 7 项通过。`openspec validate workbench-host-contract --strict` 通过。
- **DI 来源检查（FB-2）**：无新增运行期依赖。工作台编译器读取已有 `virtual:lina-plugin-pages` 注册表。

## Feedback

- [x] **FB-1**: 打开方式只按路径、外链标志和托管脚本推导，不再读取旧 Vben 组件名或 `pluginAccessMode`
- [x] **FB-2**: 去掉工作台编译器对 `#` 前缀资源的双轨直通；源码插件页按路径编译，插件清单与夹具不再把 `system/plugin/dynamic-page` 当作宿主 resource
