---
slug: '/docs/source-plugins'
title: '源码插件'
hide_title: true
description: '本文详细介绍 LinaPro 源码插件的开发流程、目录结构规范、官方插件子模块工作区、接口定义方式、PostgreSQL 安装 SQL、多租户清单字段、数据库和文件访问隔离机制、插件生命周期钩子注册、前端页面集成、菜单与权限声明，以及插件版本升级流程，提供完整的开发示例，帮助开发者掌握源码插件的核心开发技能和最佳实践。'
keywords:
  - 源码插件
  - 插件开发
  - plugin.yaml
  - 插件目录结构
  - GoFrame插件
  - 插件注册
  - pluginhost
  - PostgreSQL
  - 多租户插件
  - 官方插件子模块
  - 插件DAO
  - 插件控制器
  - 插件服务
  - 插件前端
  - 菜单声明
  - 权限声明
  - 安装SQL
  - 卸载SQL
  - LinaPro插件
---

## 概述

源码插件是`LinaPro`最常用的扩展形式，与宿主一起编译部署，通过`pluginhost`包提供的稳定接口与宿主协作，插件之间保持隔离，共享宿主的治理能力（认证、权限、日志、调度、多租户上下文等）。

官方源码插件位于`apps/lina-plugins/`，该目录在主框架仓库中以`Git submodule`形式按需挂载。只运行宿主核心能力时可以不初始化该子模块；开发、构建或测试官方插件前，需要先执行`git submodule update --init --recursive`拉取插件工作区。

## 目录结构

每个源码插件遵循统一的目录结构：

```text
apps/lina-plugins/<plugin-id>/
├── plugin.yaml                     # 插件清单（元数据、菜单声明）
├── plugin_embed.go                 # 嵌入式资源注册入口
├── backend/
│   ├── api/                        # 插件 API DTO 与路由契约
│   ├── internal/
│   │   ├── controller/             # HTTP 控制器
│   │   ├── service/                # 业务逻辑层（唯一允许的服务实现位置）
│   │   ├── dao/                    # 数据访问层（gf gen dao 生成）
│   │   └── model/
│   │       ├── do/                 # 数据操作对象（gf gen dao 生成）
│   │       └── entity/             # 数据库实体（gf gen dao 生成）
│   ├── hack/config.yaml            # 插件代码生成配置
│   └── plugin.go                   # 插件后端注册入口
├── frontend/
│   └── pages/                      # 插件前端页面（Vue 单文件组件）
├── manifest/
│   ├── sql/                        # 安装 SQL（CREATE TABLE 等）
│   │   ├── mock-data/              # 演示数据（可选）
│   │   └── uninstall/              # 卸载 SQL（DROP TABLE 等）
│   └── i18n/                       # 插件运行时语言包（可选）
│       └── <locale>/apidoc/        # 插件接口文档语言包（可选）
├── README.md
└── README.zh-CN.md
```

:::tip
`backend/internal/service/`是唯一允许放置插件服务组件的位置，不要在`backend/`根目录下创建`service/`目录。
:::

## 创建第一个源码插件

以创建`content-article`（文章管理）插件为例，完整开发流程如下：

:::info 注意
以下均是`AI Coding`工具自动开发插件的实施步骤，开发者仅做参考学习即可。
:::

### 第一步：创建目录

```bash
mkdir -p apps/lina-plugins/content-article/{backend/api,backend/internal/{controller,service,dao,model/{do,entity}},frontend/pages,manifest/{sql,sql/mock-data,sql/uninstall,i18n}}
```

### 第二步：编写插件清单

`plugin.yaml`是插件的元数据和菜单声明：

```yaml
# 插件唯一标识（kebab-case），全局唯一，用作数据库表前缀和文件命名空间
id: content-article
# 插件显示名称
name: 文章管理
# 语义化版本号（semver 格式），升级时宿主会比对此字段
version: v0.1.0
# 插件类型：source 表示源码插件，随宿主编译部署
type: source
# 多租户作用域：platform_only 表示仅平台上下文治理，tenant_aware 表示可进入租户上下文
scope_nature: tenant_aware
# 是否支持租户级安装、开通和数据隔离
supports_multi_tenant: true
# 默认安装模式：global 表示全局启用，tenant_scoped 表示按租户独立启停
default_install_mode: tenant_scoped
# 插件功能简介
description: 提供文章内容的增删改查管理功能
# 插件作者
author: linapro
# 插件许可证
license: Apache-2.0

# 插件菜单声明，宿主在启用插件时自动注册
menus:
  - key: plugin:content-article:list   # 菜单唯一标识，格式：plugin:<插件ID>:<功能>
    name: 文章管理                      # 菜单显示名称（支持 i18n Key）
    path: content-article-list             # 前端路由路径，全局唯一
    component: system/plugin/dynamic-page  # 插件页面固定使用此组件，由宿主动态加载
    perms: content-article:article:view # 访问该菜单所需的权限标识
    icon: ant-design:file-text-outlined # 菜单图标，使用 Iconify 图标名
    type: M                             # 菜单类型：D=目录，M=菜单项，B=按钮
    sort: 1                             # 排序权重，数值越小越靠前
    remark: 文章管理菜单                  # 菜单备注说明（可选）
  - key: plugin:content-article:create
    parent_key: plugin:content-article:list  # 父菜单 key，挂载在列表菜单下
    name: 创建文章
    perms: content-article:article:create    # 按钮级权限标识
    type: B                                  # B=按钮权限，不在侧边栏显示
    sort: 1
  - key: plugin:content-article:update
    parent_key: plugin:content-article:list
    name: 编辑文章
    perms: content-article:article:update
    type: B
    sort: 2
  - key: plugin:content-article:delete
    parent_key: plugin:content-article:list
    name: 删除文章
    perms: content-article:article:delete
    type: B
    sort: 3
```

### 第三步：编写安装 SQL

安装`SQL`必须具备幂等性（使用`IF NOT EXISTS`、`IF NOT EXISTS`索引和可重复执行的数据写入语句），确保重复执行不会失败。当前默认数据库为`PostgreSQL 14+`，插件安装脚本也应优先使用`PostgreSQL`方言：

```sql
-- manifest/sql/001-init.sql
CREATE TABLE IF NOT EXISTS content_article_record (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT          NOT NULL DEFAULT 0,
    "title"      VARCHAR(255) NOT NULL DEFAULT '',
    "content"    TEXT         NOT NULL DEFAULT '',
    "status"     SMALLINT     NOT NULL DEFAULT 0,
    "created_by" BIGINT       NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE content_article_record IS 'Article record table';
COMMENT ON COLUMN content_article_record."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';

CREATE INDEX IF NOT EXISTS idx_content_article_record_tenant_status
    ON content_article_record ("tenant_id", "status");
```

卸载`SQL`：

```sql
-- manifest/sql/uninstall/001-cleanup.sql
DROP TABLE IF EXISTS content_article_record;
```

### 第四步：定义 API 接口

```go
// backend/api/article/v1/article_list.go
package v1

import "github.com/gogf/gf/v2/frame/g"

type ArticleListReq struct {
    g.Meta   `path:"/plugins/content-article/article" method:"get" tags:"文章管理" summary:"文章列表" permission:"content-article:article:view"`
    Page     int    `json:"page"     v:"min:1"          dc:"页码"`
    PageSize int    `json:"pageSize" v:"min:1,max:100"  dc:"每页数量"`
}

// backend/api/article/v1/article_create.go
type ArticleCreateReq struct {
    g.Meta  `path:"/plugins/content-article/article" method:"post" tags:"文章管理" summary:"创建文章" permission:"content-article:article:create"`
    Title   string `json:"title"   v:"required|length:1,255" dc:"文章标题"`
    Content string `json:"content" v:"required"              dc:"文章内容"`
}
```

### 第五步：实现服务层

```go
// backend/internal/service/article.go
package service

import (
    "context"
    "github.com/linaproai/linapro/apps/lina-plugins/content-article/backend/internal/dao"
)

type ArticleService struct{}

func (s *ArticleService) List(ctx context.Context, page, pageSize int) (list interface{}, total int, err error) {
    // 使用插件自有的 DAO 层访问数据库
    // 数据库表已在安装 SQL 中创建
    return dao.ContentArticleRecord.Ctx(ctx).Page(page, pageSize).All()
}
```

### 第六步：注册插件

每个插件的`backend/`入口文件在`init()`中完成自注册，无需外部显式调用：

```go
// backend/plugin.go
package backend

import (
    "context"

    contentarticle "lina-plugin-content-article"
    "lina-core/pkg/pluginhost"
    articlectrl "lina-plugin-content-article/backend/internal/controller/article"
    articlesvc "lina-plugin-content-article/backend/internal/service/article"
)

const pluginID = "content-article"

func init() {
    plugin := pluginhost.NewSourcePlugin(pluginID)

    // 绑定插件嵌入资源（plugin.yaml、manifest/、frontend/）
    plugin.Assets().UseEmbeddedFiles(contentarticle.EmbeddedFiles)

    // 注册插件 HTTP 路由
    plugin.HTTP().RegisterRoutes(
        pluginhost.ExtensionPointHTTPRouteRegister,
        pluginhost.CallbackExecutionModeBlocking,
        registerRoutes,
    )

    // 注册卸载清理逻辑（在 SQL 删表之前清理文件等资源）
    plugin.Lifecycle().RegisterUninstallHandler(func(ctx context.Context, input pluginhost.SourcePluginUninstallInput) error {
        if !input.PurgeStorageData() {
            return nil
        }
        // 清理插件上传的文件
        return articlesvc.PurgeStorageData(ctx)
    })

    pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes 使用宿主发布的中间件目录注册插件路由。
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
    svc := articlesvc.New(registrar.HostServices().I18n(), registrar.HostServices().TenantFilter())
    ctrl := articlectrl.NewV1(svc)
    routes := registrar.Routes()
    middlewares := routes.Middlewares()
    routes.Group("/api/v1", func(group pluginhost.RouteGroup) {
        group.Middleware(
            middlewares.NeverDoneCtx(),
            middlewares.HandlerResponse(),
            middlewares.CORS(),
            middlewares.Ctx(),
        )
        group.Group("/", func(group pluginhost.RouteGroup) {
            group.Middleware(
                middlewares.Auth(),
                middlewares.Tenancy(),
                middlewares.Permission(),
            )
            group.Bind(
                ctrl.ListArticles,
                ctrl.CreateArticle,
            )
        })
    })
    return nil
}
```

插件通过在`init()`中调用`pluginhost.RegisterSourcePlugin(plugin)`完成自注册。宿主在`plugin-full`构建时通过`linactl`自动生成聚合入口文件`plugins.go`，其中包含所有已配置插件的空白导入：

```go
// apps/lina-plugins/plugins.go（由 linactl 自动生成，勿手动编辑）
package linaplugins

import (
    _ "lina-plugin-content-article/backend"
    // ... 其他插件
)
```

## 数据库访问

插件使用`GoFrame`的`gf gen dao`命令生成数据访问层。插件的`DAO`层与宿主`DAO`层完全独立，通过数据库命名空间和租户判别列隔离。需要支持多租户的插件表应包含`tenant_id`列，并在查询时通过宿主发布的租户过滤能力追加租户条件；未启用多租户插件时，`tenant_id = 0`表示平台上下文。

```bash
# 在插件 backend/ 目录下执行
cd apps/lina-plugins/content-article/backend
gf gen dao
```

配置文件`hack/config.yaml`指向插件自有的数据表：

```yaml
gfcli:
  gen:
    dao:
      - link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
        path: "internal"
        tables: "content_article_record"        # 明确列出插件自有表，不使用通配符
        importPrefix: "lina-plugin-content-article/backend/internal"
        descriptionTag: true
        noModelComment: true
```

## 前端页面

插件前端页面放在`frontend/pages/`目录下，框架会自动发现并挂载到宿主工作台的动态路由中：

```vue
<!-- frontend/pages/article-list.vue -->
<template>
  <div class="article-list">
    <!-- 使用 VXE Table 展示文章列表 -->
    <vxe-grid v-bind="gridOptions" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
// 通过相对路径导入，或使用宿主提供的共享包
import { useRequest } from '@vben/hooks'

const { loading, data, run } = useRequest(() =>
  fetch('/api/v1/plugin/content-article/article')
)

onMounted(() => run())
</script>
```

## 事件钩子

插件可以监听宿主发布的生命周期事件，在适当时机执行自己的逻辑：

```go
// 监听用户登录成功事件，记录登录日志
plugin.Hooks().RegisterHook(
    pluginhost.ExtensionPointAuthLoginSucceeded,
    pluginhost.CallbackExecutionModeAsync,  // 异步执行，不阻塞登录流程
    func(ctx context.Context, payload pluginhost.HookPayload) error {
        // 从 payload 中提取登录信息
        userName := pluginhost.HookPayloadStringValue(payload.Values(), pluginhost.HookPayloadKeyUserName)
        ip := pluginhost.HookPayloadStringValue(payload.Values(), pluginhost.HookPayloadKeyIP)
        return recordLoginLog(ctx, userName, ip)
    },
)
```

## 插件版本升级

当插件发布新版本时（修改`plugin.yaml`中的`version`字段），升级流程如下：

1. 在`manifest/sql/`中添加幂等的增量`SQL`文件（新增列、新增表等，所有语句使用`IF NOT EXISTS`）
2. 将`plugin.yaml`中的`version`字段改为新版本号
3. 重新编译宿主；宿主启动时会检测到数据库中的已安装版本与二进制中的声明版本不一致
4. 版本不匹配时，宿主会**拒绝启动**并在日志中列出所有待升级的插件及版本差异
5. 执行升级操作（运行增量 SQL 并更新注册表），使数据库版本与二进制版本一致
6. 重启宿主，新版本正式生效

升级 SQL 的设计原则与安装 SQL 相同——必须保证幂等性，已有内容重复执行不会报错。宿主在执行升级时会重跑安装 SQL 目录下的全部文件，仅新增的语句会实际产生变更。

## 最佳实践

- 插件`ID`使用`kebab-case`，数据表前缀使用对应的`snake_case`（如`content-article` → `content_article_`）
- 安装`SQL`使用`IF NOT EXISTS`，确保幂等性
- 插件所有服务逻辑放在`backend/internal/service/`下，不要在其他位置创建`service/`
- 菜单`key`必须使用`plugin:<plugin-id>:<menu-key>`格式
- 卸载`SQL`覆盖完整的清理逻辑，并在`RegisterUninstallHandler`中处理文件等非数据库资源
- 参考`apps/lina-plugins/plugin-demo-source`作为开发模板
