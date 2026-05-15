---
slug: '/docs/plugin-management'
title: '插件管理'
hide_title: true
description: '本文介绍 LinaPro 插件功能的日常使用，包括在 hack/config.yaml 中配置插件来源、使用 make 指令安装、升级和查看插件状态、通过 p 和 source 参数精确筛选目标插件、force 参数强制覆盖本地修改，以及通过管理端界面完成插件启用、禁用和卸载操作，帮助开发者快速上手官方插件和自定义插件的全流程管理。'
keywords:
  - 插件管理
  - plugins.install
  - plugins.update
  - plugins.status
  - plugins.init
  - hack/config.yaml
  - 插件配置
  - 官方插件
  - 源码插件
  - 插件安装
  - 插件升级
  - 插件状态
  - 插件工作区
  - 插件来源
  - force参数
  - 插件筛选
  - 插件启用
  - 插件禁用
  - LinaPro
---

## 概述

`LinaPro`提供一套完整的插件生命周期管理工具，覆盖"从代码下载到功能上线"的完整链路。插件来源在`hack/config.yaml`中统一配置，`make`指令负责将插件代码同步到本地工作区，管理端界面负责运行时的启用、禁用和卸载操作。

## 配置插件来源

插件来源在仓库根目录的`hack/config.yaml`中的`plugins.sources`节配置：

```yaml
plugins:
  sources:
    official:                                              # 来源名称，自定义，用于 --source 筛选
      repo: "https://github.com/linaproai/official-plugins.git"  # 插件 Git 仓库地址
      root: "."                                           # 插件目录所在的仓库子路径
      ref: "main"                                         # 要拉取的分支、标签或 commit
      items:                                              # 要安装的插件列表
        - "*"                                             # "*" 表示安装 root 下全部插件目录
```

`items`列表接受两种写法：

| 写法 | 说明 |
|------|------|
| `"*"` | 安装`root`下的全部插件目录 |
| `"org-center"` | 只安装指定`ID`的插件 |

可以同时配置多个来源，每个来源有独立的名称，方便后续通过`source=<名称>`参数单独操作：

```yaml
plugins:
  sources:
    official:
      repo: "https://github.com/linaproai/official-plugins.git"
      root: "."
      ref: "main"
      items:
        - "multi-tenant"
        - "org-center"
    internal:
      repo: "https://git.example.com/my-company/lina-plugins.git"
      root: "plugins"
      ref: "release/1.0"
      items:
        - "*"
```

## 初始化插件工作区

`apps/lina-plugins/`是固定的本地插件工作区。若该目录当前以`Git`子模块形式存在，需要先执行以下指令将其转换为普通目录，才能进行插件管理操作：

```bash
make plugins.init
```

该指令会将子模块解除关联并保留已有的插件代码，不会丢失本地内容。若目录不存在，则自动创建。

## 安装插件

按照`hack/config.yaml`中的配置，将插件代码拉取到`apps/lina-plugins/`：

```bash
make plugins.install
```

安装完成后，`apps/lina-plugins/`下会出现对应的插件目录，每个插件均包含`plugin.yaml`清单和完整的前后端代码。工具同时写入`.linapro-plugins.lock.yaml`记录安装状态，用于后续升级时的变更检测。

## 升级插件

将已安装插件更新到来源仓库当前`ref`指向的最新版本：

```bash
make plugins.update
```

升级前，工具会检查本地是否有未提交的修改。若插件目录存在本地改动，默认会阻止升级，以防意外覆盖：

```bash
# 强制覆盖本地修改直接升级
make plugins.update force=1
```

## 查看插件状态

查看当前工作区中已配置插件的安装状态、版本信息和本地改动情况：

```bash
make plugins.status
```

输出示例：

```text
Plugin workspace: apps/lina-plugins (ordinary)
Querying configured plugin sources...
Rendering status for 3 configured plugin(s)...

Plugin          Source    Version  Installed  Dirty  Remote
multi-tenant    official  v0.1.0   true       false  up-to-date
org-center      official  v0.1.0   true       true   up-to-date
content-notice  official  v0.1.0   false      -      up-to-date
```

各列说明：

| 列名 | 说明 |
|------|------|
| `Plugin` | 插件`ID` |
| `Source` | 来源名称（对应`hack/config.yaml`中的键） |
| `Version` | 本地已安装版本（取自`plugin.yaml`） |
| `Installed` | 插件目录是否存在 |
| `Dirty` | 本地是否有未提交的修改 |
| `Remote` | 与来源仓库当前`ref`的同步状态 |

## 在管理端启用插件

插件代码同步到`apps/lina-plugins/`并重新启动宿主后，宿主会自动扫描发现新插件。进入管理端的**扩展管理**页面：

1. 在插件列表中找到目标插件，其状态为**已发现**
2. 点击**安装**，宿主执行依赖检查并运行安装`SQL`
3. 安装成功后点击**启用**，宿主注册菜单、路由和钩子，功能即刻生效

无需重启宿主，启用后菜单和路由实时可见。

## 禁用与卸载

在管理端的**扩展管理**页面找到已启用的插件：

- **禁用**：隐藏该插件的菜单和路由，插件数据完整保留，随时可重新启用。
- **卸载**：系统会弹窗询问是否同时清理插件自有数据。选择清理后执行卸载`SQL`，数据无法恢复；选择保留则仅移除治理记录，数据表保持不动，后续重新安装时数据可复用。

## 使用自定义配置文件

默认情况下，`make`指令读取`hack/config.yaml`。若需要使用其他路径的配置文件，可通过`config`参数指定：

```bash
make plugins.install config=hack/config.staging.yaml
```

## 插件来源与工作区的关系

```mermaid
flowchart LR
    Config["hack/config.yaml<br/>声明插件来源"]
    Workspace["apps/lina-plugins/<br/>本地源码工作区"]
    Host["宿主启动<br/>自动扫描 plugin.yaml"]
    Govern["插件治理面<br/>已发现"]
    Lifecycle["安装 → 启用<br/>禁用 → 卸载"]

    Config -->|"make plugins.install<br/>make plugins.update"| Workspace
    Workspace -->|"随宿主一起编译"| Host
    Host -->|"进入"| Govern
    Govern -->|"管理端操作"| Lifecycle
```

关于插件系统的架构设计和双模式差异，参见[双模式插件系统](/docs/plugin-system)。关于在插件清单中声明多租户字段，参见[多租户能力](/docs/multi-tenant)。
