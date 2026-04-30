---
slug: '/docs'
title: '开发手册'
sidebar_position: 0
description: '开发手册是 LinaPro 面向生产应用开发者的长期参考文档，系统说明核心宿主、管理工作台、插件系统、AI 原生 OpenSpec 工作流、内置能力、架构边界、本地命令、测试策略、部署准备以及面向贡献的工程实践。'
keywords:
  - LinaPro
  - 开发手册
  - 开发者指南
  - AI 原生框架
  - 全栈框架
  - lina-core
  - lina-vben
  - lina-plugins
  - OpenSpec
  - 架构设计
  - 插件开发
  - RBAC
  - 测试
  - 部署
  - 最佳实践
  - 可持续交付
---

开发手册面向基于`LinaPro`构建真实产品的开发者。完成快速开始检查后，如果你需要理解结构、边界、扩展点和交付工作流，请从这里继续。

## 手册结构

| 领域 | 覆盖内容 |
| --- | --- |
| 核心概念 | 框架四个层次、内置能力和所有权边界。 |
| 架构设计 | 管理工作台、核心宿主、插件、数据库与工作流之间的运行关系。 |
| 插件开发 | 源码插件、动态`WASM`插件、清单、生命周期和升级行为。 |
| 测试与部署 | 本地命令、`E2E`要求、构建准备和运维检查。 |

## 推荐阅读顺序

1. 阅读[框架层次](/docs/core-concepts/layers)，理解主要仓库区域。
2. 阅读[内置能力](/docs/core-concepts/built-in-capabilities)，再判断哪些功能需要自己实现。
3. 阅读[运行时架构](/docs/architecture/runtime-architecture)，再修改共享契约。
4. 阅读[插件开发](/docs/plugin-development)，判断功能是否应放在宿主之外。
5. 阅读[测试与部署](/docs/testing-deployment)，再准备进入评审。

## 事实来源

官网文档必须与`linapro`源码仓库保持一致。当源码行为发生变化时，先更新中文文档，再用自然英文同步英文版本。
