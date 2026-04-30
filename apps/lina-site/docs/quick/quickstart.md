---
slug: '/quickstart'
title: '快速开始'
sidebar_position: 1
description: '本页面向第一次体验 LinaPro 的开发者，按最短路径说明如何安装框架源码、初始化数据库、加载可选演示数据、启动 Go 后端与 Vue 管理工作台、使用默认账号登录，并完成核心 API、插件运行时与管理界面的基础检查。'
keywords:
  - LinaPro
  - 快速开始
  - AI 原生全栈框架
  - 本地开发
  - 安装脚本
  - make init
  - make mock
  - make dev
  - Go 后端
  - Vue 3 工作台
  - lina-core
  - lina-vben
  - 插件运行时
  - OpenAPI 文档
  - RBAC
  - 演示数据
---

通过三条命令把`LinaPro`跑起来。读完本页后，核心宿主服务、管理工作台和插件运行时都会在本地就绪，可以继续开发、扩展或验证功能。

## 准备工作

`LinaPro`支持`macOS`、`Linux`、`Windows`。开始之前，请确认以下工具已在`PATH`中：

- `Go 1.21+`：用于核心宿主服务。
- `Node.js 18+`与`pnpm 9+`：用于`Vue 3`管理工作台。
- `Make`：本地常用任务的统一入口。
- `Git`：源码获取与`OpenSpec`工作流都会用到。
- `MySQL 8.0+`：用于持久化运行状态。

默认项目使用关系型数据库。第一次本地体验时，建议直接准备一个本地`MySQL`实例，启动路径最稳定。

## 1. 安装

在你希望创建项目的目录下，执行官方安装脚本：

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

`Windows PowerShell`环境使用下面的命令：

```powershell
irm https://linapro.ai/install.ps1 | iex
```

安装脚本会下载仓库源码归档，默认创建`./linapro`目录，并执行环境体检。脚本不会自动安装依赖，也不会自动启动服务。继续之前请先`cd`到生成的项目目录。

## 2. 初始化

生成运行时配置、执行数据库迁移，并按需写入示例数据：

```bash
make init confirm=init
make mock confirm=mock   # 可选——写入示例用户、角色、菜单
```

`confirm=`是一个安全闸：所有会改动持久化状态的命令，必须显式声明意图才会执行。只有在明确要重建本地数据库时，才使用`make init confirm=init rebuild=true`。

## 3. 开发调试

一条命令拉起后端宿主和前端工作台：

```bash
make dev
```

启动完成后，你会得到：

| 服务                              | 地址                     |
| --------------------------------- | ------------------------ |
| 管理工作台（`lina-vben`）         | http://localhost:5666    |
| 核心`API`（`lina-core`）          | http://localhost:8080    |

使用默认账号登录工作台：

- **用户名：** `admin`
- **密码：** `admin123`

:::warning
在把工作台暴露到网络之前，请先修改默认密码。可在工作台 **系统管理 → 用户管理** 下完成。
:::

## 验证环境

正式动手之前，建议跑一次冒烟测试：

1. 访问`http://localhost:5666`，使用默认账号登录。
2. 进入 **扩展中心 → 插件管理**，确认官方插件已经可见。
3. 访问`http://localhost:8080/api`，确认`OpenAPI`文档已经提供。

如果其中任何一步失败，请跳转到 [故障排查](#故障排查)。

## 下一步

你现在已经跑在生产团队同款的技术栈上。按照你要构建的方向继续阅读：

- 阅读[工作台导览](/quick/workspace-tour)，了解内置管理模块。
- 进入[开发手册](/docs)，继续查看架构、插件、测试与部署内容。
- 访问[开源社区](/community)，获取帮助或参与贡献。

## 故障排查

- **`make: command not found`**：用系统包管理器安装`Make`，例如`brew install make`或`apt install make`。
- **`5666`或`8080`端口被占用**：停掉占用端口的进程，或在生成的项目配置中调整端口后重新执行`make dev`。
- **数据库连接失败**：打开生成的配置文件，把`database`段指向你的本地`MySQL`实例，然后重新执行`make init confirm=init`。
- **其他问题**：欢迎在[GitHub](https://github.com/linaproai/linapro/issues)提交`issue`，团队会持续跟进。
