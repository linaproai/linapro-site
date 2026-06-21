---
slug: '/docs/source-plugins'
title: '源码插件'
hide_title: true
description: '适用场景、目录结构、plugin.yaml清单、插件ID命名规范、安装SQL、API契约、服务层实现、pluginhost注册、数据库访问、前端页面、事件钩子、运行时升级和最佳实践，帮助开发者用原生Go方式扩展长期业务能力。'
keywords:
  - 源码插件
  - 插件开发
  - pluginhost
  - plugin.yaml
  - 插件目录结构
  - 插件ID命名规范
  - GoFrame插件
  - 插件注册
  - 安装SQL
  - 运行时升级
  - 插件前端
  - 插件DAO
  - 多租户插件
  - tenant_id
  - 菜单声明
  - 权限声明
  - 生命周期回调
  - LinaPro插件
  - 插件配置
  - manifest资源
  - 原始资源读取
---

## 基本介绍

源码插件是`LinaPro`推荐的默认扩展方式。它以`Go`源码形式与主框架一起编译部署，使用`pluginhost`注册路由、钩子、定时任务、生命周期回调和治理逻辑，适合长期维护、性能要求高、需要完整工程体验的业务模块。

官方源码插件位于`apps/lina-plugins/`。主仓库通过该目录挂载官方插件工作区；用户项目也在该目录下维护自己的业务插件。

## 适用场景

| 场景 | 是否推荐源码插件 | 原因 |
|------|------------------|------|
| 长期业务模块 | 推荐 | 可测试、可审查、性能最好 |
| 组织、内容、监控等后台能力 | 推荐 | 与主框架权限、菜单、调度和多租户协作紧密 |
| 运行时热加载 | 不优先 | 源码插件需要重新构建和部署主框架 |
| 商业二进制分发 | 不优先 | 源码插件通常暴露源码 |

## 标准目录结构

```text
apps/lina-plugins/<plugin-id>/
├── plugin.yaml
├── plugin_embed.go
├── backend/
│   ├── api/                         # API DTO与路由契约
│   ├── hack/
│   │   └── config.yaml              # make dao等插件开发配置
│   ├── internal/
│   │   ├── controller/              # HTTP控制器
│   │   ├── service/                 # 业务服务层
│   │   ├── dao/                     # make dao生成
│   │   └── model/                   # do/entity模型
│   └── plugin.go                    # 插件注册入口
├── frontend/
│   ├── pages/                       # 插件页面
│   └── slots/                       # 插槽页面，可选
├── manifest/
│   ├── config/
│   │   ├── config.yaml              # 开发期默认配置
│   │   └── config.example.yaml      # 配置模板，不作为运行时默认值
│   ├── sql/                         # 安装与升级SQL
│   │   ├── mock-data/               # 演示数据，可选
│   │   └── uninstall/               # 卸载SQL
│   └── i18n/                        # 插件语言包
└── README.md
```

`backend/internal/service/`是插件服务逻辑的固定位置，不要在插件根目录或`backend/`根目录另建`service/`包。

## 插件清单

`plugin.yaml`声明插件身份、运行形态、多租户边界、菜单和权限：

```yaml
id: linapro-content-notice
name: 内容通知
version: v0.1.0
type: source
scope_nature: tenant_aware
supports_multi_tenant: true
default_install_mode: tenant_scoped
description: 提供内容变更通知的发布与订阅能力
author: linapro
license: Apache-2.0
menus:
  - key: plugin:linapro-content-notice:list
    name: 内容通知
    path: linapro-content-notice-list
    component: system/plugin/dynamic-page
    perms: content-notice:notice:view
    icon: ant-design:notification-outlined
    type: M
    sort: 1
  - key: plugin:linapro-content-notice:create
    parent_key: plugin:linapro-content-notice:list
    name: 创建通知
    perms: content-notice:notice:create
    type: B
```

插件`ID`推荐使用`<author>-<domain>-<capability>`三段式`kebab-case`结构，例如`linapro-content-notice`中`linapro`为作者、`content`为领域、`notice`为能力。`<domain>`段建议从`content`、`monitor`、`org`、`tenant`、`auth`、`oidc`、`ai`、`storage`、`workflow`、`message`等常见业务领域中选取，完整领域建议参见[插件系统](./5000-plugin-system.md)。菜单`key`必须全局唯一，推荐使用`plugin:<plugin-id>:<menu-key>`格式。按钮权限通过`type: B`挂在菜单下，不直接出现在侧边栏中。

## 数据库与SQL

插件安装`SQL`位于`manifest/sql/`，卸载`SQL`位于`manifest/sql/uninstall/`。安装和升级脚本必须幂等，常用`CREATE TABLE IF NOT EXISTS`、`CREATE INDEX IF NOT EXISTS`等写法。

```sql
CREATE TABLE IF NOT EXISTS linapro_content_notice_record (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id" INT NOT NULL DEFAULT 0,
    "title" VARCHAR(255) NOT NULL DEFAULT '',
    "content" TEXT NOT NULL DEFAULT '',
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_linapro_content_notice_record_tenant
    ON linapro_content_notice_record ("tenant_id");
```

需要支持多租户的插件表应包含`tenant_id`列。未启用`multi-tenant`插件时，`tenant_id = 0`表示平台上下文。

## API与服务层

插件`API`定义同样使用`g.Meta`声明路径、方法、权限和文档说明：

```go
type NoticeListReq struct {
    g.Meta `path:"/notices" method:"get" tags:"Notice" summary:"List notices" permission:"content-notice:notice:view"`
    Page     int `json:"page" v:"min:1"`
    PageSize int `json:"pageSize" v:"min:1,max:100"`
}
```

这里的`path`是控制器绑定到插件路由组后的相对路径。源码插件业务`API`应统一挂载到`/x/{plugin-id}/...`，而不是占用主框架`/api/v1`控制面。例如上面的接口在`linapro-content-notice`插件中推荐暴露为：

```text
/x/linapro-content-notice/notices
```

服务层通过插件自有`DAO`访问数据库。需要租户隔离时，应使用主框架发布的`TenantFilterService`追加租户条件，而不是手写不一致的过滤规则。

## 注册入口

源码插件在`backend/plugin.go`中通过`init()`注册：

```go
func init() {
    plugin := pluginhost.NewSourcePlugin("linapro-content-notice")

    plugin.Assets().UseEmbeddedFiles(contentnotice.EmbeddedFiles)

    plugin.HTTP().RegisterRoutes(
        pluginhost.ExtensionPointHTTPRouteRegister,
        pluginhost.CallbackExecutionModeBlocking,
        registerRoutes,
    )

    plugin.Cron().RegisterCron(
        pluginhost.ExtensionPointCronRegister,
        pluginhost.CallbackExecutionModeBlocking,
        registerCronJobs,
    )

    plugin.Lifecycle().RegisterBeforeUpgradeHandler(beforeUpgrade)
    plugin.Lifecycle().RegisterAfterUpgradeHandler(afterUpgrade)

    pluginhost.RegisterSourcePlugin(plugin)
}
```

主框架在插件完整模式下生成聚合入口，空白导入已配置插件，使这些`init()`注册逻辑进入主框架进程。

### 路由注册

`HTTPRegistrar.Routes()`返回插件路由注册器。注册器的`APIPrefix()`当前返回插件专属命名空间`/x/{plugin-id}`，后续路径段由插件自行组织：

```go
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
    routes := registrar.Routes()
    middlewares := routes.Middlewares()

    routes.Group(routes.APIPrefix(), func(group pluginhost.RouteGroup) {
        group.Middleware(
            middlewares.NeverDoneCtx(),
            middlewares.HandlerResponse(),
            middlewares.CORS(),
            middlewares.RequestBodyLimit(),
            middlewares.Ctx(),
        )

        group.Group("/", func(group pluginhost.RouteGroup) {
            group.Middleware(
                middlewares.Auth(),
                middlewares.Tenancy(),
                middlewares.Permission(),
            )
            group.Bind(articleController)
        })
    })
    return nil
}
```

源码插件仍可注册非保留公开路由，例如`/portal/...`、`/assets/...`或`/`，用于门户页面、自管静态文件或`SPA fallback`。这类路由不属于插件`API`，不会自动投影为工作台菜单、权限或`OpenAPI`接口。源码插件不得在`/x`下注册其他插件的路径；注册冲突或越界路径会在启动阶段被拒绝。

### 插件配置和清单资源

源码插件通过`registrar.HostServices()`获取插件作用域主框架服务。与配置和`manifest`资源相关的能力包括：

| 服务 | 用途 | 架构设计 |
|------|------|----------|
| `Config()` | 读取当前插件自己的配置，生产覆盖路径为生产配置根下`plugins/<plugin-id>/config.yaml`，开发期默认路径为`manifest/config/config.yaml` | [ConfigService](./9000-capabilities/7000-config.md) |
| `HostConfig()` | 读取宿主公开配置白名单键，例如`workspace.basePath`、`i18n.default`和`i18n.enabled` | [HostConfigService](./9000-capabilities/7000-config.md) |
| `Manifest()` | 读取当前插件`manifest/`下的原始资源，例如`profile.yaml`、`config/config.example.yaml`或`i18n/zh-CN/plugin.json` | [ManifestService](./9000-capabilities/7200-manifest.md) |

`manifest/config/config.example.yaml`只是模板，不参与默认读取。插件不应通过`g.Cfg()`扫描宿主完整配置树，也不应把插件业务配置写进主框架`config.yaml`。完整的配置读取优先级、`manifest`资源路径语义和专用资源管线边界，参见[插件配置与manifest资源](./4000-plugin-config-and-manifest.md)。各能力服务的架构设计和使用约束，参见[插件基础能力](./9000-capabilities/9000-capabilities.md)。

```go
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
    services := registrar.HostServices()

    timeout, err := services.Config().Duration(ctx, "sync.timeout", 5*time.Second)
    if err != nil {
        return err
    }
    _ = timeout

    workspaceBase, err := services.HostConfig().String(ctx, "workspace.basePath", "/admin")
    if err != nil {
        return err
    }
    _ = workspaceBase

    var profile struct {
        Category string `yaml:"category"`
    }
    if err := services.Manifest().Scan(ctx, "profile.yaml", "", &profile); err != nil {
        return err
    }
    return nil
}
```

这里的`profile.yaml`只是普通`YAML`资源示例。路径相对`manifest/`，不应写成`manifest/profile.yaml`。

## 前端页面

源码插件的前端页面位于`frontend/pages/`，由主框架工作台动态页壳加载。插件菜单中的`component`通常使用：

```yaml
component: system/plugin/dynamic-page
```

页面可以复用默认工作台的前端生态和设计规范。插件禁用后，主框架菜单接口不再返回该插件入口，工作台侧边栏会自动隐藏。

源码插件也可以声明`public_assets`，让主框架把插件内可公开的静态资源托管到`/x-assets/{plugin-id}/{version}/...`。只有`plugin.yaml`中显式声明的资源目录会被公开；租户文件、用户私有文件、安装脚本和配置文件不应放入`public_assets`。

## 事件钩子与定时任务

源码插件可以订阅主框架事件，例如登录成功、插件启用、系统启动等。钩子可以同步阻断，也可以异步执行，取决于注册时选择的执行模式。

插件也可以注册自己的定时任务处理器，供管理工作台创建任务时选择：

```go
plugin.Cron().RegisterCron(
    pluginhost.ExtensionPointCronRegister,
    pluginhost.CallbackExecutionModeBlocking,
    func(registry pluginhost.CronRegistry) error {
        registry.Register("content-notice:cleanup", cleanupExpiredNotices)
        return nil
    },
)
```

## 运行时升级

源码插件文件更新后，主框架会比较数据库中的有效版本和当前发现版本。发现更高版本时，插件进入`pending_upgrade`运行时状态，主框架基础治理能力保持可用，插件业务入口进入受控状态。

管理员在插件管理页执行显式运行时升级。升级流程会重新读取有效清单和目标清单，执行依赖检查、`BeforeUpgrade`回调、插件自定义升级逻辑、升级`SQL`、治理资源同步、有效版本切换和缓存失效。失败后进入`upgrade_failed`，可以查看诊断信息并重试。

这种模型避免把文件覆盖误认为数据和治理资源已经完成升级。

## 最佳实践

- 插件`ID`使用`<author>-<domain>-<capability>`三段式`kebab-case`结构，数据库表前缀使用对应的`snake_case`。
- 安装和升级`SQL`必须幂等，避免保留数据后重新安装失败。
- 服务逻辑放在`backend/internal/service/`。
- 插件只使用`pluginhost`和`pluginservice`等稳定契约，不直接依赖主框架的`internal/`包。
- 多租户插件表预留`tenant_id`列，并使用主框架租户过滤服务。
- 菜单和按钮权限一并声明，避免页面可见但操作权限缺失。
- 卸载时区分治理记录、数据库数据和文件数据，避免误删。
