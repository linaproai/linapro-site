---
slug: '/docs/architecture'
title: '🏗️ 架构设计'
hide_title: true
description: '本文详细介绍 LinaPro 的整体架构设计，包括核心宿主服务、默认管理工作台、双模式插件系统和 AI 研发工作流四个层次的职责边界、运行时交互关系和设计原则，帮助开发者深入理解框架的分层模型、松耦合设计思路以及各组件如何协作完成业务功能交付。'
keywords:
  - LinaPro架构
  - 四层架构
  - 核心宿主服务
  - lina-core
  - lina-vben
  - 插件系统
  - AI研发工作流
  - 架构设计
  - 松耦合
  - GoFrame
  - Vue3
  - 分层设计
  - 插件运行时
  - RBAC权限
  - 设计原则
  - 扩展点
---



## 架构总览

`LinaPro`采用四层分离的架构模型，每一层都遵循松耦合原则独立设计，层与层之间通过定义良好的接口协作，而不是硬依赖内部实现。关键的层级组件包括：

```mermaid
graph TB
    subgraph Workflow["AI 研发工作流  openspec/"]
        direction LR
        Explore["🔍 探索"] --> Propose["📋 提案"] --> Implement["⚙️ 实现"] --> Review["🔎 审查"] --> Archive["📦 归档"]
    end

    subgraph Frontend["默认管理工作台  lina-vben"]
        UI["Vue 3 + Vben5 + Ant Design"]
    end

    subgraph Host["核心宿主服务  lina-core"]
        direction TB
        API["API 接口层\n(g.Meta 路由定义 + DTO)"]
        Ctrl["控制器层\n(HTTP 请求处理)"]
        Svc["业务服务层\n(核心业务逻辑)"]
        Plugin["插件运行时\n(生命周期编排 · 沙箱隔离)"]
        Gov["治理服务\n(JWT · RBAC · 日志 · 会话)"]
        API --> Ctrl --> Svc
        Svc --> Plugin
        Svc --> Gov
    end

    subgraph Plugins["插件层  lina-plugins"]
        direction LR
        Source["源码插件\n随宿主编译交付"]
        Dynamic["WASM 动态插件\n运行时热加载"]
    end

    DB[("数据存储\nMySQL")]

    Workflow -.->|规范驱动| Frontend
    Workflow -.->|规范驱动| Host
    UI -->|"HTTP / WebSocket"| API
    Plugin -->|加载| Source
    Plugin -->|沙箱执行| Dynamic
    Svc --> DB
    Gov --> DB
```

## 目录结构

```text
/                                   
├── AGENTS.md                         # AI 智能体协作规范
├── hack/                             # 开发工具脚本
├── apps/                             # 应用管理
│    ├── lina-core/                   # 核心宿主服务（Go）
│    ├── lina-vben/                   # 默认管理工作台（Vue 3）
│    └── lina-plugins/                # 插件系统
└── openspec/                         # AI 研发工作流
```

## 核心宿主服务（lina-core）

`lina-core`是框架的稳定地基，基于`Go`语言构建，承担所有通用平台能力。宿主服务不了解具体的业务域，只提供通用的基础设施。

**主要职责：**

- 提供系统管理、插件治理、用户认证、权限控制等`RESTful API`
- 拥有数据库迁移脚本、初始化数据和运行时配置
- 加载和管理源码插件及`WASM`动态插件的完整生命周期
- 运行宿主级定时任务和持久化定时调度子系统
- 提供国际化运行时，聚合宿主和插件的语言资源


## 默认管理工作台（lina-vben）

`lina-vben`是`lina-core`的标准消费者，基于`Vue 3 + Vben5 + Ant Design Vue`构建。工作台本身没有业务逻辑，它只是宿主`API`和插件`API`的`UI`表达层。

**关键特点：**

- 与宿主在接口契约、权限模型和设计规范上完全对齐
- 插件菜单通过宿主动态路由自动注入，无需修改工作台代码
- 开发者可以通过插件机制替换任意模块，甚至替换整个工作台

## 双模式插件系统（lina-plugins）

插件层是`LinaPro`扩展能力的核心，支持两种截然不同的插件交付模式。宿主通过`pluginhost`包暴露稳定的扩展接缝，插件只能通过这些接缝与宿主交互，永远不直接访问宿主内部实现。

详见[双模式插件系统](/docs/plugin-system)。

## AI 研发工作流（openspec/）

`AI`研发工作流凌驾于所有层次之上，它是让规范、代码与测试保持同步的连接纽带。工作流不直接属于任何运行时组件，而是作为治理机制约束整个开发过程。

详见[AI规范驱动开发](/docs/spec-driven-development)。

## 分层边界原则

理解以下边界原则，有助于避免常见的架构错误：

### 宿主不了解业务域

`lina-core`不包含任何具体的业务逻辑（如文章管理、订单管理等），这些能力在理想情况下是通过插件提供，而宿主只提供稳定的通用基础设施。

### 插件不修改宿主内部

插件通过`pluginhost`包暴露的稳定接口（路由注册、事件钩子、菜单过滤、权限过滤）与宿主协作。插件不直接调用宿主的`internal/`包，不直接访问宿主的数据库表。

### 宿主拥有顶级菜单目录

宿主拥有稳定的顶级菜单目录键（`dashboard`、`iam`、`setting`、`scheduler`、`extension`、`developer`），插件菜单只能挂载在宿主已发布的目录下或插件自己的菜单树中。

### 工作台是消费者，不是定义者

`lina-vben`消费宿主的`API`和插件的`API`，但不定义它们的行为。工作台的`UI`变化不应该影响后端行为，后端的`API`变化通过接口契约文档传达给工作台。

## 数据交互

```mermaid
sequenceDiagram
    participant User as 用户浏览器
    participant Vben as lina-vben
    participant Core as lina-core
    participant Plugin as 源码插件
    participant DB as MySQL

    User->>Vben: 访问管理工作台
    Vben->>Core: GET /api/v1/menu（获取菜单）
    Core->>Plugin: 触发菜单过滤钩子
    Plugin-->>Core: 注入插件菜单
    Core-->>Vben: 返回完整菜单树
    Vben-->>User: 渲染侧边栏（含插件菜单）

    User->>Vben: 点击插件页面
    Vben->>Core: 请求插件 API
    Core->>Plugin: 路由分发到插件控制器
    Plugin->>DB: 访问插件命名空间数据
    DB-->>Plugin: 返回数据
    Plugin-->>Core: 返回响应
    Core-->>Vben: 返回 API 响应
    Vben-->>User: 渲染页面数据
```

## 部署模型

`LinaPro`支持两种部署模式，可以根据业务规模随时切换：

### 单机模式（默认）

```mermaid
graph TB
    Nginx["Nginx"]
    Core["lina-core :8080"]
    DB[("MySQL")]

    Nginx --> Core
    Core --> DB
```

适合开发环境和小规模生产部署，配置简单，资源占用低。

### 集群模式

```mermaid
graph TB
    LB["负载均衡"]
    N1["lina-core 节点 1"]
    N2["lina-core 节点 2\n分布式选主"]
    NN["lina-core 节点 N"]
    DB[("MySQL")]

    LB --> N1
    LB --> N2
    LB --> NN
    N1 --> DB
    N2 --> DB
    NN --> DB
```

集群模式通过配置`cluster.enabled: true`启用，框架基于 `MySQL` 实现分布式选主、分布式锁和权限拓扑缓存（`sys_locker`、`sys_kv_cache` 表），无需引入额外中间件，也无需修改业务代码。详见[原生分布式架构](/docs/distributed-architecture)。

## 相关文档

import DocCardList from '@theme/DocCardList';

<DocCardList />