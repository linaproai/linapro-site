---
slug: '/docs/commands'
title: '开发指令'
hide_title: true
description: '本文介绍 LinaPro 项目中跨平台开发指令集的作用、支持的参数选项和使用示例，涵盖 linactl、Makefile 兼容入口、Windows make.cmd 包装入口、开发服务管理、完整构建、WASM 插件构建、Docker 镜像构建、测试验证、国际化检查和数据库初始化等场景，帮助开发者在 macOS、Linux 和 Windows 上稳定使用同一套项目工具链。'
keywords:
  - make指令
  - 开发指令
  - 构建命令
  - LinaPro开发
  - make dev
  - make build
  - make test
  - make image
  - make init
  - make wasm
  - linactl
  - 开发工作流
  - 后端构建
  - 前端构建
  - Docker镜像
  - E2E测试
  - 数据库初始化
  - Windows支持
  - make.cmd
  - 跨平台
---

`LinaPro`项目提供了一套跨平台开发指令集。长期维护的任务编排集中在`hack/tools/linactl`中，以`Go`程序实现；根目录`Makefile`和`make.cmd`只是兼容入口，都会转发到底层`linactl`。因此同一套命令可以在`macOS`、`Linux`和`Windows`上使用，不再依赖`GNU Make`或`POSIX Shell`作为唯一入口。

## 平台说明

**跨平台原生命令**：所有平台都可以直接使用`linactl`：

```bash
cd hack/tools/linactl
go run . help
go run . status
go run . init confirm=init
go run . dev
```

**macOS / Linux**：可以继续使用根目录`make`兼容入口：

```bash
make help
make init confirm=init
make dev
```

**Windows cmd.exe**：使用项目根目录的`make.cmd`包装入口。在`cmd.exe`中会按可执行文件扩展名查找当前目录脚本，因此可直接省略`.cmd`后缀：

```cmd
make dev
make build
make help
```

**Windows PowerShell**：需要加当前目录前缀。默认`Windows`环境下可写成`.\make`；如需避免与本机已安装的其他`make`命令混淆，可显式使用`.\make.cmd`：

```powershell
.\make help
.\make init confirm=init
.\make dev
```

后续文档中所有`make <指令>`示例，都可以等价替换为`cd hack/tools/linactl && go run . <指令>`，参数格式保持一致。

## 指令总览

| 指令 | 分类 | 说明 |
|------|------|------|
| `make dev` | 开发服务 | 重启前后端开发服务器 |
| `make stop` | 开发服务 | 停止前后端开发服务器 |
| `make status` | 开发服务 | 查看前后端运行状态及日志路径 |
| `make build` | 构建 | 完整构建前端、插件和后端二进制 |
| `make wasm` | 构建 | 构建所有或指定运行时`WASM`插件 |
| `make tidy` | 构建 | 整理宿主、工具和插件相关`Go`模块依赖 |
| `make image` | 镜像 | 构建生产`Docker`镜像 |
| `make image-build` | 镜像 | 仅准备镜像产物，不执行`Docker`构建 |
| `make test` | 测试 | 运行完整`E2E`测试套件 |
| `make test.go` | 测试 | 运行`Go`单元测试 |
| `make test.host` | 测试 | 只运行宿主自有`E2E`测试 |
| `make test.plugins` | 测试 | 运行官方插件自有`E2E`测试 |
| `make test.scripts` | 测试 | 运行工具脚本的单元与`smoke`测试 |
| `make i18n.check` | 国际化 | 扫描运行时硬编码文案并校验语言包`key`覆盖 |
| `make init` | 数据库 | 初始化数据库表结构和种子数据 |
| `make mock` | 数据库 | 加载演示`Mock`数据 |
| `make help` | 其他 | 查看所有可用指令 |

`linactl`还直接提供`plugins.init`、`plugins.install`、`plugins.update`和`plugins.status`等插件工作区管理命令，用于需要把官方插件子模块转换为普通插件目录、安装配置中的源码插件或查看插件工作区状态的高级场景。

## 插件模式参数

官方插件目录`apps/lina-plugins/`是`Git submodule`。当该目录已初始化且包含插件清单时，`dev`、`build`、`image`和相关`Go`测试命令会自动进入插件完整模式，并基于根目录宿主专用`go.work`生成已忽略的`temp/go.work.plugins`，再通过`GOWORK`解析源码插件模块。

如果只需要运行主框架，可以跳过子模块初始化，或在命令中传入`plugins=0`强制宿主模式：

```bash
make dev plugins=0
make build plugins=0
make image plugins=0
```

构建或测试官方插件前，需要先初始化子模块：

```bash
git submodule update --init --recursive
```

## 开发服务

### make dev

重启后端和前端开发服务器。执行前会先停止已有服务，依次完成`WASM`插件构建、前端静态资源准备和后端编译，等待两端健康检查通过后打印运行状态。

```bash
make dev
```

后端默认监听`http://localhost:8080`，前端默认监听`http://localhost:5666`。运行日志分别写入`temp/lina-core.log`和`temp/lina-vben.log`。

### make stop

停止后端和前端开发服务器，并清理残留`PID`文件。对于仍占用端口的僵尸进程，会强制终止。

```bash
make stop
```

### make status

打印前后端当前的运行状态及日志文件路径，便于快速确认服务是否正常启动。

```bash
make status
```

输出示例：

```text
╔══════════════════════════════════════════════╗
║         LinaPro Framework Status             ║
╠══════════════════════════════════════════════╣
║  Backend:  ✓ running  http://localhost:8080  ║
║  Frontend: ✓ running  http://localhost:5666  ║
╠══════════════════════════════════════════════╣
║  Backend log:   temp/lina-core.log           ║
║  Frontend log:  temp/lina-vben.log           ║
╚══════════════════════════════════════════════╝
```

## 构建

### make build

完整构建流程，依次执行：前端静态资源构建、嵌入到后端的`manifest`资源准备、所有`WASM`插件构建，最后编译后端宿主二进制。构建产物输出到`temp/output/`目录。

```bash
# 默认构建（当前平台）
make build

# 指定目标平台（交叉编译）
make build platforms=linux/amd64,linux/arm64

# 开启详细日志
make build verbose=1
# 或
make build v=1
```

构建行为的默认值由仓库根目录的`hack/config.yaml`集中管理，命令行参数会覆盖文件中的对应字段。

```yaml
build:
  # 目标平台列表，使用goos/goarch格式，make build platforms=...可覆盖
  platforms:        
    - "auto"
  # 是否启用 CGO
  cgoEnabled: false 
  # 构建产物输出路径，相对于仓库根目录
  outputDir: "temp/output" 
  # 编译后生成的二进制文件名
  binaryName: "lina"       
```

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `build.platforms` | `["auto"]` | 目标平台列表，使用`goos/goarch`格式，`auto`表示`linux/<当前架构>`，`make build platforms=...`可覆盖 |
| `build.cgoEnabled` | `false` | 是否启用`CGO` |
| `build.outputDir` | `temp/output` | 构建产物输出路径，相对于仓库根目录 |
| `build.binaryName` | `lina` | 宿主二进制文件名 |

### make wasm

单独构建运行时`WASM`插件。`make wasm`是兼容入口，默认把产物输出到`temp/output/`，支持使用`p=<plugin-id>`只构建指定插件。需要覆盖输出目录或只做构建探测时，直接使用`linactl wasm`。

```bash
# 构建所有 WASM 插件
make wasm

# 只构建指定插件（plugin-id 为插件目录名）
make wasm p=my-plugin

# 指定输出目录（linactl 原生命令）
cd hack/tools/linactl
go run . wasm p=my-plugin out=../../temp/output

# 仅检查可构建的动态插件（linactl 原生命令）
go run . wasm dry_run=true
```

### linactl prepare-packed-assets

准备后端嵌入发布所需的前端静态资源和`manifest`资产，通常由`make build`或`make dev`自动调用。需要单独检查嵌入资源时可以手动执行：

```bash
cd hack/tools/linactl
go run . prepare-packed-assets
```

### make tidy

整理宿主、开发工具和插件相关`Go`模块依赖，适合在升级依赖或初始化插件完整模式后执行：

```bash
make tidy
```

## 镜像

### make image

完整的`Docker`镜像构建流程：先执行`make build`生成所有构建产物，再调用`hack/tools/image-builder`封装成镜像。镜像名称、标签、镜像仓库地址等均通过参数配置。

```bash
# 使用默认配置构建镜像
make image

# 指定标签和镜像仓库
make image tag=v0.6.0 registry=ghcr.io/linaproai

# 构建后直接推送
make image tag=v0.6.0 registry=ghcr.io/linaproai push=1

# 多平台构建
make image platforms=linux/amd64,linux/arm64 tag=v0.6.0
```

镜像构建的默认值同样由`hack/config.yaml`管理，命令行参数可覆盖本次执行。

```yaml
image:
  # 镜像名称，构建时拼接 registry 前缀
  name: "linapro"       
  # 默认标签，留空时根据 git 信息自动推导             
  tag: "dev"  
  # 远端仓库前缀，例如 ghcr.io/linaproai                       
  registry: ""        
  # 是否默认推送，push=1 可覆盖本次执行               
  push: false      
  # 运行时基础镜像                  
  baseImage: "alpine:3.22"         
  # Dockerfile 路径，相对于仓库根目录  
  dockerfile: "hack/docker/Dockerfile" 
```

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `image.name` | `linapro` | 镜像名称，构建时会在前面拼接`registry`前缀 |
| `image.tag` | `dev` | 默认标签，留空时根据`git`信息自动推导 |
| `image.registry` | 空 | 远端仓库前缀，例如`ghcr.io/linaproai` |
| `image.push` | `false` | 是否默认推送，命令行`push=1`可覆盖本次执行 |
| `image.baseImage` | `alpine:3.22` | 运行时基础镜像 |
| `image.dockerfile` | `hack/docker/Dockerfile` | `Dockerfile`路径，相对于仓库根目录 |

### make image-build

仅准备镜像构建的所有产物（等价于先执行`make build`），不执行`Docker build`步骤。适用于需要手动检查产物或自定义镜像构建步骤的场景。

```bash
make image-build
```

## 测试

### make test

运行`hack/tests/`下完整的`Playwright E2E`测试套件。执行前请确保开发服务已通过`make dev`启动。

```bash
make test
```

### make test.go

运行所有受维护`Go`模块的单元测试，支持通过`plugins=0`强制宿主模式。

```bash
make test.go
make test.go plugins=0
```

### make test.host

只运行宿主自有`Playwright E2E`测试，不要求初始化官方插件子模块。

```bash
make test.host
```

### make test.plugins

运行官方插件自有`Playwright E2E`测试，执行前需要先初始化`apps/lina-plugins/`子模块。

```bash
make test.plugins
```

### make test.scripts

运行跨平台仓库工具的单元和`smoke`检查，用于验证`linactl`、`make.cmd`等辅助入口的基本正确性。

```bash
make test.scripts
```

## 国际化

### make i18n.check

扫描运行时可见的代码路径，检测未被纳入国际化体系的硬编码文案，并校验宿主和各插件运行时语言包的消息`key`覆盖情况。适合在提交新功能前进行`i18n`合规自查。

```bash
make i18n.check
```

## 插件工具

### linactl plugins.status

查看当前官方插件工作区状态。高级开发者也可以使用`plugins.init`、`plugins.install`和`plugins.update`管理非子模块形态的源码插件工作区。

```bash
cd hack/tools/linactl
go run . plugins.status
```

## 数据库

:::caution 破坏性操作

`init`和`mock`均会对数据库执行破坏性操作，因此要求显式传入`confirm`参数才能执行，防止误操作。

:::

### make init

初始化数据库的表结构（`DDL`）和系统必需的种子数据。后端会按`config.yaml`中`database.default.link`的配置自动选择`PostgreSQL`或`SQLite`方言，其中`PostgreSQL 14+`是默认数据存储，`SQLite`仅用于本地演示或冒烟验证。

```bash
# 仅初始化（保留现有数据）
make init confirm=init

# 重建数据库（清空后重新初始化）
make init confirm=init rebuild=true
```

### make mock

在`make init`完成之后，加载用于本地演示和开发验证的可选`Mock`数据。

```bash
make mock confirm=mock
```

## 其他

### make help

打印根`Makefile`及所有引入目标文件中的可用指令列表，输出按指令名称排序。

```bash
make help
```
