---
slug: '/docs/architecture/runtime-architecture'
title: '运行时架构'
sidebar_position: 0
description: '本文说明 LinaPro 的运行时架构，展示 Vue 管理工作台如何消费 Go 核心宿主 API，宿主如何编排治理能力和插件生命周期，官方插件如何加载，以及 OpenSpec 如何让需求、实现、测试和归档规范保持一致。'
keywords:
  - LinaPro
  - 运行时架构
  - 架构设计
  - lina-core
  - lina-vben
  - lina-plugins
  - OpenSpec
  - 管理工作台
  - 核心宿主
  - 插件生命周期
  - API 契约
  - 治理能力
  - 数据库
  - Mermaid 图
  - AI 原生工作流
  - 全栈架构
---

运行时架构需要让宿主保持通用，让工作台围绕契约开发，让插件拥有独立边界。

```mermaid
graph TB
    subgraph Workflow["AI 研发工作流"]
        Explore["探索"] --> Propose["提案"] --> Implement["实现"] --> Review["评审"] --> Archive["归档"]
    end

    subgraph Frontend["管理工作台  lina-vben"]
        UI["Vue 3 + Ant Design Vue"]
    end

    subgraph Host["核心宿主服务  lina-core"]
        API["API 契约"]
        Service["领域服务"]
        Governance["认证 · RBAC · 审计 · 任务"]
        Runtime["插件运行时"]
        API --> Service
        Service --> Governance
        Service --> Runtime
    end

    subgraph Plugins["插件工作区  lina-plugins"]
        Source["源码插件"]
        Dynamic["动态 WASM 插件"]
    end

    DB[("数据库")]

    Workflow -.-> Frontend
    Workflow -.-> Host
    UI --> API
    Runtime --> Source
    Runtime --> Dynamic
    Service --> DB
```

## 运行时契约

| 契约 | 所有者 |
| --- | --- |
| 后端路由和`DTO` | `apps/lina-core/api`以及插件本地`backend/api`。 |
| 管理页面 | `apps/lina-vben`以及插件本地`frontend/pages`。 |
| 权限和菜单 | 宿主治理数据与插件清单。 |
| 数据库资源 | 宿主`SQL`清单或插件本地`manifest/sql`。 |
| 变更意图 | 实现前位于`openspec/changes`，归档后位于`openspec/specs`。 |

## 设计原则

不要让宿主依赖插件内部实现。插件应通过清单、宿主公开包、生命周期`Hook`和已发布扩展点接入。
