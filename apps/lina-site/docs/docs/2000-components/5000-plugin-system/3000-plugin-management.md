---
slug: '/docs/plugin-management'
title: '插件安装、升级与治理'
hide_title: true
description: '插件来源配置、本地插件工作区、plugins.init、plugins.install、plugins.update、plugins.status、管理端安装启用、禁用卸载、动态插件上传和运行时升级的完整链路，帮助开发者理解插件从代码获取到运行时生效的治理过程。'
keywords:
  - 插件管理
  - 插件工作区
  - plugins.init
  - plugins.install
  - plugins.update
  - plugins.status
  - hack/config.yaml
  - 插件来源
  - 插件安装
  - 插件升级
  - 插件状态
  - 动态插件上传
  - 运行时升级
  - 插件列表缓存
  - plugin.autoEnable
  - 插件禁用
  - 插件卸载
  - apps/lina-plugins
  - LinaPro插件
---

## 基本介绍

插件管理连接两条链路：

- **开发期链路**：从`hack/config.yaml`读取插件来源，将源码插件同步到`apps/lina-plugins/`工作区。
- **运行时链路**：主框架扫描插件清单后，在管理工作台完成发现、安装、启用、禁用、卸载和升级。

这两条链路分工明确。代码同步只代表本地工作区出现了插件文件；插件是否安装、启用、升级成功，仍由主框架运行时治理记录决定。

```mermaid
flowchart LR
    Config["hack/config.yaml<br/>插件来源"]
    Workspace["apps/lina-plugins<br/>本地插件工作区"]
    Host["主框架扫描plugin.yaml"]
    Govern["插件治理记录<br/>发现、安装、启用"]
    Runtime["运行时投影<br/>菜单、路由、Hook、Cron"]

    Config -->|"plugins.install / update"| Workspace
    Workspace --> Host
    Host --> Govern
    Govern --> Runtime
```

插件管理页读取的是主框架构建的完整插件列表投影。列表包含发现版本、有效版本、依赖检查、`hostServices`授权、声明路由、是否存在演示数据、运行时升级状态和租户开通策略等信息，前端筛选只是在完整投影上派生，不会让接口丢失详情字段。

## 插件来源配置

插件来源在仓库根目录的`hack/config.yaml`中配置：

```yaml
plugins:
  sources:
    official:
      repo: "https://github.com/linaproai/official-plugins.git"
      root: "."
      ref: "main"
      items:
        - "*"
```

| 字段 | 说明 |
|------|------|
| `repo` | 插件来源仓库地址 |
| `root` | 插件目录在来源仓库中的根路径 |
| `ref` | 分支、标签或`commit` |
| `items` | 需要安装的插件列表，`"*"`表示全部插件 |

可以配置多个来源，例如官方插件源和企业内部插件源。命令支持通过`source=<名称>`筛选。

:::info 提示
目前版本仅支持开源仓库作为插件来源，后续会支持私有仓库和动态插件来源的安装机制。
:::
## 插件工作区

`apps/lina-plugins/`是固定插件工作区。官方插件通常以`Git submodule`形式挂载；如果项目要自行管理插件源码，可以先转换为普通目录：

```bash
make plugins.init
```

该命令会解除子模块关联并保留已有插件代码。若目录不存在，则创建空工作区。

:::info 提示
该指令是可选的，在执行安装指令时会自动执行目录转换。执行后，`apps/lina-plugins/`就成为一个普通目录，可以直接在其中添加、修改或删除插件源码。
:::

## 安装和更新源码插件

安装插件：

```bash
make plugins.install
```

更新插件：

```bash
make plugins.update
```

更新前工具会检查本地插件目录是否存在未提交改动。默认情况下，存在本地改动会阻断更新，避免覆盖开发者修改。如确实需要覆盖：

```bash
make plugins.update force=1
```

查看状态：

```bash
make plugins.status
```

状态输出会展示插件`ID`、来源、版本、本地是否已安装、是否存在本地改动，以及远端同步状态。

## 管理端生命周期

插件代码进入工作区后，主框架启动时扫描`plugin.yaml`并将插件呈现为“已发现”。随后管理员在扩展中心完成运行时生命周期操作：

| 操作 | 运行时行为 |
|------|------------|
| **安装** | 检查依赖，执行安装`SQL`，写入插件治理记录 |
| **启用** | 投影菜单、权限、路由、钩子、定时任务和前端资源 |
| **禁用** | 隐藏菜单和业务路由，保留数据和治理记录 |
| **卸载** | 清理治理记录，并按选择保留或清理插件自有数据 |
| **升级** | 对待升级插件执行预览、确认、迁移、版本切换和缓存失效 |

禁用和卸载的区别很重要：禁用只是撤出运行入口，数据仍在；卸载会进入清理流程，若选择清理数据，插件自有数据可能无法恢复。

`GET /api/v1/plugins`是纯读取接口，不会因为打开列表而隐式执行同步写入。插件同步、动态插件上传、安装、卸载、启用、禁用、源码插件升级、动态插件升级和租户插件开通策略变更会主动失效列表缓存；集群模式下通过插件运行时修订通知其他节点刷新。主框架启动后会异步预热列表缓存，预热失败只记录日志，请求到达时仍可按需重建。

## 动态插件上传

`WASM`动态插件不依赖源码工作区交付。构建产物为`.wasm`文件，管理员在扩展中心上传后，主框架会验证产物并读取内嵌清单、路由、资源和授权声明。

动态插件安装时需要管理员确认`hostServices`授权。授权确认后，主框架才允许插件通过`pluginbridge`访问对应的服务和资源范围。

如果动态插件声明了资源型`hostServices`，例如`storage`的`resources.paths`、`network`的`resources[].url`、`data`的`resources.tables`、`hostconfig`的`resources.keys`或`manifest`的`resources.paths`，启动自动启用也必须拥有已确认的授权快照，否则不会绕过治理直接启用。

## 运行时升级

插件文件更新后，主框架可能发现“有效版本”和“发现版本”不一致。此时插件进入运行时升级状态：

| 状态 | 说明 |
|------|------|
| `normal` | 有效版本与发现版本一致 |
| `pending_upgrade` | 发现更高版本，等待管理员显式升级 |
| `upgrade_running` | 正在执行升级 |
| `upgrade_failed` | 升级失败，保留旧有效版本并记录诊断 |
| `abnormal` | 文件版本低于有效版本或状态异常，需要人工处理 |

升级前可以查看预览，包括版本差异、依赖检查、`SQL`数量、`hostServices`差异和风险提示。执行升级时，主框架会进行确认校验、获取运行时升级锁、执行生命周期回调、运行升级`SQL`、同步治理资源、切换有效版本并刷新缓存。

## 启动自动启用

主框架支持通过`config.yaml`中的`plugin.autoEnable`在启动时自动安装并启用指定插件：

```yaml
plugin:
  autoEnable:
    - id: "linapro-demo-source"
      withMockData: false
    - id: "linapro-demo-dynamic"
      withMockData: true
```

`autoEnable`条目必须是对象结构`{id, withMockData}`，不再接受裸字符串。`withMockData`缺省为`false`；设为`true`时，只在启动自动安装阶段加载插件`manifest/sql/mock-data`下的演示数据，对已经安装的插件不会重复加载。重复插件`ID`会按首次出现的条目生效并去重。

对于租户级插件，启动自动启用还会同步`autoEnableForNewTenants`等租户开通策略，并请求租户`Provider`为既有租户补齐开通状态。生产环境应谨慎使用演示数据，动态插件还应确保授权快照已确认。

## 多租户治理

租户感知插件可以选择全局启用或租户级启用：

| 模式 | 行为 |
|------|------|
| `global` | 插件安装启用一次，对平台或所有租户生效 |
| `tenant_scoped` | 插件可按租户单独启用或停用 |

具体启用策略由主框架治理记录和`multi-tenant`插件控制，不由前端单独决定。

## 管理界面展示

扩展中心的列表展示以实际治理字段为准。插件类型列显示“源码插件”或“动态插件”；即便动态产物底层是`WASM`，治理类型仍归为“动态插件”。列表默认更强调安装时间、更新时间、版本和运行时升级状态，已弱化旧版“交付模式”“集成模式”“入口”等容易与治理状态混淆的概念。

## 常见建议

- 开发期同步插件代码后，仍要在管理端安装和启用插件。
- 更新插件源码前先检查本地改动，避免误用`force=1`覆盖开发内容。
- 动态插件上传更高版本后，不要假设新版本已经生效；需要执行运行时升级。
- 不要依赖插件列表请求触发同步；需要刷新发现结果时使用显式同步操作。
- `plugin.autoEnable`只适合开发、演示或受控初始化流程，生产环境应显式审查`hostServices`和演示数据开关。
- 卸载生产插件前确认是否保留数据，并检查是否存在反向依赖。
- 插件来源建议锁定到稳定分支、标签或`commit`，避免生产环境使用不可控浮动版本。
