---
slug: '/docs/source-plugin'
title: '源码插件'
sidebar_position: 2
hide_title: true
description: '本文详细介绍 LinaPro 源码插件的开发流程、目录结构规范、接口定义方式、数据库和文件访问隔离机制、插件生命周期钩子注册、前端页面集成、菜单与权限声明，以及插件版本升级流程，提供完整的开发示例，帮助开发者掌握源码插件的核心开发技能和最佳实践。'
keywords:
  - 源码插件
  - 插件开发
  - plugin.yaml
  - 插件目录结构
  - GoFrame插件
  - 插件注册
  - pluginhost
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

源码插件是`LinaPro`最常用的扩展形式，与宿主一起编译部署，通过`pluginhost`包提供的稳定接口与宿主协作，插件之间保持隔离，共享宿主的治理能力（认证、权限、日志、调度等）。

## 目录结构

每个源码插件遵循统一的目录结构：

```text
apps/lina-plugins/<plugin-id>/
  plugin.yaml               插件清单（元数据、菜单声明）
  plugin_embed.go           嵌入式资源注册入口
  backend/
    api/                    插件 API DTO 与路由契约
    internal/
      controller/           HTTP 控制器
      service/              业务逻辑层（唯一允许的服务实现位置）
      dao/                  数据访问层（gf gen dao 生成）
      model/
        do/                 数据操作对象（gf gen dao 生成）
        entity/             数据库实体（gf gen dao 生成）
    hack/config.yaml        插件 codegen 配置
    plugin.go               插件后端注册入口
  frontend/
    pages/                  插件前端页面（Vue 单文件组件）
  manifest/
    sql/                    安装 SQL（CREATE TABLE 等）
    sql/mock-data/          演示数据（可选）
    sql/uninstall/          卸载 SQL（DROP TABLE 等）
    i18n/                   插件运行时语言包（可选）
    i18n/<locale>/apidoc/   插件接口文档语言包（可选）
  README.md
  README.zh_CN.md
```

:::tip
`backend/internal/service/`是唯一允许放置插件服务组件的位置，不要在`backend/`根目录下创建`service/`目录。
:::

## 创建第一个源码插件

以创建`content-article`（文章管理）插件为例，完整开发流程如下：

### 第一步：创建目录

```bash
mkdir -p apps/lina-plugins/content-article/{backend/api,backend/internal/{controller,service,dao,model/{do,entity}},frontend/pages,manifest/{sql,sql/mock-data,sql/uninstall,i18n}}
```

### 第二步：编写插件清单

`plugin.yaml`是插件的元数据和菜单声明：

```yaml
id: content-article
name: 文章管理
version: v0.1.0
type: source
description: 提供文章内容的增删改查管理功能
author: linapro
license: Apache-2.0

menus:
  - key: plugin:content-article:list
    name: 文章管理
    path: content-article-list
    component: system/plugin/dynamic-page
    perms: content-article:article:view
    icon: ant-design:file-text-outlined
    type: M
    sort: 1
    remark: 文章管理菜单
  - key: plugin:content-article:create
    parent_key: plugin:content-article:list
    name: 创建文章
    perms: content-article:article:create
    type: B
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

安装`SQL`必须具备幂等性（使用`IF NOT EXISTS`），确保重复执行不会失败：

```sql
-- manifest/sql/001-init.sql
CREATE TABLE IF NOT EXISTS `content_article_record` (
    `id`         BIGINT       NOT NULL AUTO_INCREMENT COMMENT '文章ID',
    `title`      VARCHAR(255) NOT NULL               COMMENT '文章标题',
    `content`    LONGTEXT     NOT NULL               COMMENT '文章内容',
    `status`     TINYINT      NOT NULL DEFAULT 0     COMMENT '状态：0=草稿 1=已发布',
    `created_by` BIGINT       NOT NULL DEFAULT 0     COMMENT '创建人ID',
    `created_at` DATETIME     NOT NULL               COMMENT '创建时间',
    `updated_at` DATETIME     NOT NULL               COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章记录表';
```

卸载`SQL`：

```sql
-- manifest/sql/uninstall/001-cleanup.sql
DROP TABLE IF EXISTS `content_article_record`;
```

### 第四步：定义 API 接口

```go
// backend/api/article/article.go
package article

import "github.com/gogf/gf/v2/frame/g"

type ArticleListReq struct {
    g.Meta   `path:"/plugin/content-article/article" method:"get"
               tags:"文章管理" summary:"文章列表"
               middleware:"AuthMiddleware,PermMiddleware"
               perm:"content-article:article:view"`
    Page     int    `json:"page"     v:"min:1"          dc:"页码"`
    PageSize int    `json:"pageSize" v:"min:1,max:100"  dc:"每页数量"`
}

type ArticleCreateReq struct {
    g.Meta  `path:"/plugin/content-article/article" method:"post"
              tags:"文章管理" summary:"创建文章"
              middleware:"AuthMiddleware,PermMiddleware"
              perm:"content-article:article:create"`
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

```go
// backend/plugin.go
package backend

import (
    "github.com/linaproai/linapro/apps/lina-core/pkg/pluginhost"
    "github.com/linaproai/linapro/apps/lina-plugins/content-article/backend/internal/controller"
)

func Register(p pluginhost.SourcePlugin) {
    // 绑定前端资源（从 plugin_embed.go 引入）
    p.Assets().UseEmbeddedFiles(embedFS)

    // 注册插件 HTTP 路由
    p.HTTP().RegisterRoutes(
        pluginhost.ExtensionPointHTTPRouteRegister,
        pluginhost.CallbackExecutionModeBlocking,
        func(group *ghttp.RouterGroup) {
            group.Bind(controller.NewArticleController())
        },
    )

    // 注册卸载清理逻辑（在 SQL 删表之前清理文件等资源）
    p.Lifecycle().RegisterUninstallHandler(func(ctx context.Context, keepData bool) error {
        if !keepData {
            // 清理插件上传的文件
        }
        return nil
    })
}
```

在宿主插件注册文件中接线：

```go
// apps/lina-plugins/lina-plugins.go
import contentarticle "github.com/linaproai/linapro/apps/lina-plugins/content-article/backend"

func init() {
    pluginregistry.Register("content-article", contentarticle.Register)
}
```

## 数据库访问

插件使用`GoFrame`的`gf gen dao`命令生成数据访问层。插件的`DAO`层与宿主`DAO`层完全独立，通过数据库命名空间隔离：

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
      - link: "mysql:..."
        tables: "content_article_*"  # 只生成插件自有表的 DAO
        descriptionTag: true
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
p.Hooks().RegisterHook(
    pluginhost.ExtensionPointAuthLoginSucceeded,
    pluginhost.CallbackExecutionModeAsync,  // 异步执行，不阻塞登录流程
    func(ctx context.Context, payload *pluginhost.HookPayload) error {
        // 将登录信息写入插件自有的日志表
        return recordLoginLog(ctx, payload)
    },
)
```

## 插件版本升级

当插件发布新版本时（修改`plugin.yaml`中的`version`字段），升级流程如下：

1. 宿主启动时扫描发现新版本（`plugin.yaml`版本 > 数据库中已安装版本）
2. 如果发现版本不匹配，**宿主启动失败**并提示执行升级命令
3. 通过`lina-upgrade AI`技能执行升级：

```
# 在 Claude Code 中
upgrade source plugin content-article
```

4. 升级完成后重启宿主，新版本正式生效

## 最佳实践

- 插件`ID`使用`kebab-case`，数据表前缀使用对应的`snake_case`（如`content-article` → `content_article_`）
- 安装`SQL`使用`IF NOT EXISTS`，确保幂等性
- 插件所有服务逻辑放在`backend/internal/service/`下，不要在其他位置创建`service/`
- 菜单`key`必须使用`plugin:<plugin-id>:<menu-key>`格式
- 卸载`SQL`覆盖完整的清理逻辑，并在`RegisterUninstallHandler`中处理文件等非数据库资源
- 参考`apps/lina-plugins/plugin-demo-source`作为开发模板
