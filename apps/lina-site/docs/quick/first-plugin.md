---
slug: '/quick/first-plugin'
title: '第一个插件路径'
sidebar_position: 3
description: '本文面向准备扩展 LinaPro 的新手开发者，说明应如何理解插件开发入口、何时选择源码插件或动态 WASM 插件、优先阅读哪些仓库目录，以及在正式编写生产插件前需要掌握的清单、生命周期、前端页面、后端接口、SQL 资源和升级概念。'
keywords:
  - LinaPro
  - 第一个插件
  - 插件开发
  - 源码插件
  - 动态插件
  - WASM 插件
  - plugin.yaml
  - 插件生命周期
  - lina-plugins
  - 后端 API
  - 前端页面
  - SQL 清单
  - 菜单注入
  - 源码升级
  - 扩展点
  - 沙箱
---

插件是`LinaPro`最主要的扩展路径。一个插件可以携带自己的后端路由、服务逻辑、数据库资源、前端页面、菜单、`Hook`以及生命周期行为。

## 选择插件模式

| 模式 | 适用场景 |
| --- | --- |
| 源码插件 | 插件随产品源码一起交付，并在开发期显式接线到宿主中。官方插件使用这种模式。 |
| 动态`WASM`插件 | 插件需要通过插件管理完成运行期安装、启用、停用或卸载。 |

如果业务模块与产品源码在同一仓库中长期演进，建议先从源码插件开始，并参考`apps/lina-plugins/plugin-demo-source/`。

## 优先阅读的文件

| 路径 | 用途 |
| --- | --- |
| `apps/lina-plugins/README.md` | 插件工作区规则和官方插件列表。 |
| `apps/lina-plugins/plugin-demo-source/README.md` | 源码插件结构与注册方式。 |
| `apps/lina-plugins/plugin-demo-dynamic/README.md` | 动态插件结构与生命周期模型。 |
| `apps/lina-plugins/OPERATIONS.md` | 插件运维操作流程。 |

## 最小心智模型

生产插件需要保持清晰的所有权边界：

1. `plugin.yaml`声明元信息、菜单、页面、资源和生命周期要求。
2. `backend/`承载路由、控制器、服务以及本插件本地生成的数据访问工件。
3. `frontend/pages/`承载由宿主菜单挂载的页面。
4. `manifest/sql/`拥有安装、演示数据和卸载数据库资源。
5. 源码插件版本升级需要显式执行`make upgrade confirm=upgrade scope=source-plugin plugin=<plugin-id>`。

理解这个结构后，再进入开发手册中的插件开发章节继续深入。
