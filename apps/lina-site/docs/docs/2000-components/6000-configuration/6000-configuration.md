---
slug: '/docs/configuration'
title: '服务配置管理'
hide_title: true
description: 'LinaPro 分层配置体系概览，涵盖主框架静态配置、主框架动态配置和插件业务配置三个层次，介绍 config.yaml 与 sys_config 的分工协作、插件配置隔离机制，以及生产环境最佳实践，帮助开发者和运维人员全面理解配置管理架构。'
keywords:
  - 配置管理
  - config.yaml
  - LinaPro配置
  - 运行时配置
  - sys_config
  - 分层配置体系
  - 静态配置
  - 动态配置
  - 插件配置
  - 配置隔离
  - 生产配置
  - 最佳实践
  - 配置优先级
  - 热更新
  - 集群同步
---

## 基本介绍

`LinaPro` 采用**分层配置体系**，将配置分为三个层次：主框架静态配置、主框架动态配置和插件业务配置。这种设计既保证了核心框架的稳定性，又为运行时调整和插件扩展提供了灵活性。

### 配置层次

| 层次 | 来源 | 说明 |
|------|------|------|
| <span style={{whiteSpace: 'nowrap'}}><strong>主框架静态配置</strong></span> | `config.yaml` | 启动时加载，进程生命周期内不变，涵盖服务、日志、数据库、认证等核心组件 |
| <span style={{whiteSpace: 'nowrap'}}><strong>主框架动态配置</strong></span> | <span style={{whiteSpace: 'nowrap'}}>`sys_config`数据表</span> | 可在运行时热更新，覆盖静态默认值；进程内以`1`小时`TTL`缓存，集群模式通过`Redis`同步修订号保持一致 |
| <span style={{whiteSpace: 'nowrap'}}><strong>插件业务配置</strong></span> | 插件独立配置文件 | 插件拥有独立的配置作用域，通过优先级机制读取配置，与主框架配置隔离 |

### 配置文件位置

主框架默认配置文件位于：

```text
apps/lina-core/manifest/config/config.yaml
```

仓库同时提供了一份完整双语注释的配置模板，适合作为逐字段参考：

```text
apps/lina-core/manifest/config/config.template.yaml
```

插件配置文件位于各自插件目录下，优先级和读取方式详见[插件业务配置](/docs/plugin-configuration)。

## 相关文档

import DocCardList from '@theme/DocCardList';

<DocCardList />
