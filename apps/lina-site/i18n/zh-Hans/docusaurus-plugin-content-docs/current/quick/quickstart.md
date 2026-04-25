---
slug: '/quickstart'
title: '快速开始'
sidebar_position: 0
---

通过三条命令把 `LinaPro` 跑起来。读完本页，核心宿主服务、管理工作台和插件运行时都会在你的电脑上就绪，随时可用、可扩展、可替换。

## 准备工作

`LinaPro` 支持 `macOS`、`Linux`、`Windows`。开始之前，请确认以下工具已在 `PATH` 中：

- **Go**：核心宿主服务（`lina-core`）从源码编译。
- **Node.js**，配合 `pnpm` 或 `yarn`：用于运行 Vue 3 管理工作台（`lina-vben`）。
- **Make**：所有常用任务的统一入口。
- **Git**：安装脚本与 `OpenSpec` 工作流都会用到。

持久化状态需要一个关系型数据库。第一次启动时，`make mock` 生成的 `SQLite` 文件已经够用；后续随时可以切换到 `MySQL` 或 `PostgreSQL`。

## 1. 安装

在你希望创建项目的目录下，执行官方安装脚本：

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

脚本会下载 `LinaPro` CLI、生成脚手架项目，并打印它放置项目的路径。继续之前请先 `cd` 到该目录。

## 2. 初始化

生成运行时配置、执行数据库迁移，并按需写入示例数据：

```bash
make init confirm=init
make mock confirm=mock   # 可选——写入示例用户、角色、菜单
```

`confirm=` 是一个安全闸：所有会改动持久化状态的命令，必须显式声明意图才会执行。`make init` 可以重复执行，结果幂等。

## 3. 开发调试

一条命令拉起整套环境——核心宿主、管理工作台、插件运行时：

```bash
make dev
```

启动完成后，你会得到：

| 服务                              | 地址                     |
| --------------------------------- | ------------------------ |
| 管理工作台（`lina-vben`）         | http://localhost:5666    |
| 核心 API（`lina-core`）           | http://localhost:8080    |

使用默认账号登录工作台：

- **用户名：** `admin`
- **密码：** `admin123`

:::warning
在把工作台暴露到网络之前，请先修改默认密码。可在工作台 **系统管理 → 用户管理** 下完成。
:::

## 验证环境

正式动手之前，建议跑一次冒烟测试：

1. 访问 `http://localhost:5666`，使用默认账号登录。
2. 进入 **系统管理 → 插件管理**，应能看到官方插件（部门、通知、监控类等）。
3. 访问 `http://localhost:8080/api`，确认 OpenAPI 文档已经提供。

如果其中任何一步失败，请跳转到 [故障排查](#故障排查)。

## 下一步

你现在已经跑在生产团队同款的技术栈上。按照你要构建的方向继续阅读：

- **架构总览**：四个层次（`lina-core`、`lina-vben`、`lina-plugins`、`openspec`）以及它们如何咬合。
- **内置模块**：用户、角色、菜单、字典、参数、文件、任务调度、插件、API 文档、系统信息——开箱即用并已接入 RBAC。
- **插件开发**：编译期接入的源码插件，或运行期热加载的 WASM 插件。两种模式都运行在沙箱中，数据库与文件系统按命名空间隔离。
- **OpenSpec 工作流**：AI 研发闭环（`explore → propose → implement → review → archive`）。每次变更从规范出发，伴随 E2E 测试落地。

## 故障排查

- **`make: command not found`**：用系统包管理器安装 Make（`brew install make`、`apt install make` 等）。
- **`5666` 或 `8080` 端口被占用**：停掉占用端口的进程，或在生成的项目配置中调整端口后重新执行 `make dev`。
- **数据库连接失败**：打开生成的配置文件，把 `database` 段指向你的 `MySQL` / `PostgreSQL` / `SQLite` 实例，然后重新执行 `make init confirm=init`。
- **其他问题**：欢迎在 [GitHub](https://github.com/linaproai/linapro/issues) 提交 issue，团队每天都会跟进。
