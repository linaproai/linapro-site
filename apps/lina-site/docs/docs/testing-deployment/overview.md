---
slug: '/docs/testing-deployment'
title: '测试与部署'
sidebar_position: 0
description: '本文汇总 LinaPro 的本地命令工作流、涉及数据库变更命令的安全确认参数、安装脚本冒烟测试、E2E 测试要求、构建准备、默认服务端口、运维检查，以及提交生产变更前应该完成的验证实践。'
keywords:
  - LinaPro
  - 测试
  - 部署
  - 本地命令
  - make init
  - make mock
  - make dev
  - make test
  - make test-install
  - E2E 测试
  - Playwright
  - 数据库初始化
  - 构建准备
  - 服务端口
  - 运维检查
  - 生产就绪
---

本地命令应保持统一使用方式，确保安装、验证和评审都可以复现。

## 常用命令

| 命令 | 用途 |
| --- | --- |
| `make init confirm=init` | 初始化数据库结构和种子数据。 |
| `make init confirm=init rebuild=true` | 明确需要重建本地数据库时使用。 |
| `make mock confirm=mock` | 初始化后加载可选演示数据。 |
| `make dev` | 本地启动后端和前端服务。 |
| `make stop` | 停止本地服务。 |
| `make status` | 查看本地服务状态。 |
| `make test-install` | 运行安装脚本冒烟测试。 |
| `make test` | 运行完整`E2E`测试套件。 |

会修改持久化状态的命令都需要显式确认参数，避免误操作直接落库。

## 默认本地端口

| 服务 | 端口 |
| --- | --- |
| 管理工作台 | `5666` |
| 核心宿主`API` | `8080` |

## 验证清单

准备提交评审前，请完成下面的检查：

1. 先运行与变更最相关的最小测试。
2. 如果修改了共享流程、权限、插件或用户可见行为，运行`make test`。
3. 如果修改了路由、菜单或权限行为，手动检查管理工作台和`API`文档。
4. 如果跳过某项验证，需要记录具体原因。

面向部署的专门文档会随着公开发布打包流程稳定后继续扩展。
