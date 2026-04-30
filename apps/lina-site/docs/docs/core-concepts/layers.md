---
slug: '/docs/core-concepts/layers'
title: '框架层次'
sidebar_position: 0
description: '本文说明 LinaPro 的四个核心层次：Go 核心宿主服务、Vue 管理工作台、官方插件工作区和 OpenSpec AI 原生研发工作流，并明确每一层的职责、协作方式，以及为什么框架需要保持宿主、界面、插件和规范的所有权边界。'
keywords:
  - LinaPro
  - 框架层次
  - 核心概念
  - lina-core
  - lina-vben
  - lina-plugins
  - OpenSpec
  - 核心宿主
  - 管理工作台
  - 插件系统
  - AI 原生工作流
  - 所有权边界
  - 松耦合
  - 全栈架构
  - 可持续交付
  - 仓库结构
---

`LinaPro`围绕四个层次组织。这些层次共同演进，但各自保持清晰的所有权边界。

| 层次 | 仓库路径 | 职责 |
| --- | --- | --- |
| 核心宿主服务 | `apps/lina-core` | 后端`API`契约、服务治理、认证鉴权、权限、插件生命周期、数据库迁移和定时任务。 |
| 管理工作台 | `apps/lina-vben` | 基于`Vue 3`的工作台，以及内置能力的参考`UI`实现。 |
| 插件工作区 | `apps/lina-plugins` | 官方源码插件、插件示例、清单、前端页面和插件自有资源。 |
| `AI`研发工作流 | `openspec/` | 规范驱动的变更记录，让需求、实现、评审和归档后的能力基线保持对齐。 |

## 为什么需要分层

宿主应保持稳定和通用。业务模块如果可以独立安装、升级或移除，优先放入插件边界。工作台消费后端契约，但不应单独定义宿主行为。`OpenSpec`位于实现工作之上，让变更背景不会随着一次对话结束而丢失。

## 仓库地图

```text
apps/
  lina-core/      Core host service
  lina-vben/      Management workspace
  lina-plugins/   Official and sample plugins
hack/
  scripts/install/ Bootstrap installers
  tests/          Playwright E2E suite
openspec/
  changes/        Active and archived changes
  specs/          Current baseline specifications
```

新增功能应放在真正拥有该行为的层次中。如果变更跨越多个层次，请先写清楚契约。
