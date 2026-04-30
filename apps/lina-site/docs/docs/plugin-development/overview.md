---
slug: '/docs/plugin-development'
title: '插件开发'
sidebar_position: 0
description: '本文说明 LinaPro 插件开发的所有权模型、源码插件与动态 WASM 插件的选择方式、插件目录结构、清单职责、前后端边界、SQL 资源归属、生命周期行为，以及显式的源码插件版本升级流程。'
keywords:
  - LinaPro
  - 插件开发
  - 插件系统
  - 源码插件
  - 动态插件
  - WASM 插件
  - plugin.yaml
  - 生命周期 Hook
  - 插件清单
  - 前端页面
  - 后端服务
  - SQL 资源
  - 源码升级
  - lina-plugins
  - 扩展点
  - 沙箱
---

插件是自包含的扩展单元。插件应拥有自己的路由、服务、页面、数据库资源、菜单和生命周期行为。

## 插件类型

| 类型 | 特点 |
| --- | --- |
| 源码插件 | 位于`apps/lina-plugins/<plugin-id>/`，显式接线到产品源码中，并使用源码插件升级流程。 |
| 动态`WASM`插件 | 通过运行期上传、安装、启用、停用和卸载完成生命周期管理。 |

## 标准源码插件结构

```text
apps/lina-plugins/<plugin-id>/
  backend/
    api/
    internal/
      controller/
      service/
      dao/
      model/do/
      model/entity/
    hack/config.yaml
    plugin.go
  frontend/pages/
  manifest/sql/
  manifest/sql/mock-data/
  manifest/sql/uninstall/
  plugin.yaml
  plugin_embed.go
  README.md
  README.zh_CN.md
```

`backend/internal/service/`是源码插件业务`service`的合法目录。不要再创建并行的`backend/service/`目录。

## 升级流程

当已安装源码插件的`plugin.yaml`版本更新后，宿主不会在启动时静默切换到新版本。需要显式执行升级：

```bash
make upgrade confirm=upgrade scope=source-plugin plugin=<plugin-id>
```

只有在明确要升级所有已安装且发现更高版本的源码插件时，才使用`plugin=all`。

## 扩展纪律

插件代码应依赖宿主公开包和扩展点。如果插件需要新的宿主能力，应先定义宿主契约，再把插件实现放在该契约之后。
