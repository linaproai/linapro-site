---
slug: '/docs/commands'
title: '开发指令'
hide_title: true
description: '本文介绍 LinaPro 项目中跨平台开发指令集的作用、支持的参数选项和使用示例，涵盖 linactl、Makefile 兼容入口、Windows make.cmd 包装入口、开发服务管理、完整构建、WASM 插件构建、Docker 镜像构建、插件工作区管理、测试验证、国际化检查、数据库初始化和发布治理等场景，帮助开发者在 macOS、Linux 和 Windows 上稳定使用同一套项目工具链。'
keywords:
  - make指令
  - 开发指令
  - 构建命令
  - LinaPro开发
  - make dev
  - make dev.setup
  - make build
  - make pack.assets
  - make test
  - make image
  - make image.build
  - make db.init
  - make wasm
  - make plugins.install
  - make plugins.status
  - make release.tag.check
  - linactl
  - 开发工作流
  - 后端构建
  - 前端构建
  - Docker镜像
  - E2E测试
  - 数据库初始化
  - 插件工作区
  - 发布治理
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
make db.init confirm=init
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
.\make db.init confirm=init
.\make dev
```

后续文档中所有`make <指令>`示例，都可以等价替换为`cd hack/tools/linactl && go run . <指令>`，参数格式保持一致。

## 指令总览

| 指令 | 分类 | 说明 |
|------|------|------|
| `make dev.setup` | 开发服务 | 安装前端依赖和`Playwright`浏览器 |
| `make dev` | 开发服务 | 重启前后端开发服务器 |
| `make stop` | 开发服务 | 停止前后端开发服务器 |
| `make status` | 开发服务 | 查看前后端运行状态及日志路径 |
| `make build` | 构建 | 完整构建前端、插件和后端二进制 |
| `make pack.assets` | 构建 | 准备主框架嵌入所需的前端静态资源和`manifest` |
| `make wasm` | 构建 | 构建所有或指定运行时`WASM`插件 |
| `make tidy` | 构建 | 整理主框架、工具和插件相关`Go`模块依赖 |
| `make image` | 镜像 | 构建生产`Docker`镜像 |
| `make image.build` | 镜像 | 仅准备镜像产物，不执行`Docker`构建 |
| `make test` | 测试 | 运行完整`E2E`测试套件 |
| `make test.go` | 测试 | 运行`Go`单元测试 |
| `make test.host` | 测试 | 只运行主框架自有`E2E`测试 |
| `make test.plugins` | 测试 | 运行官方插件自有`E2E`测试 |
| `make test.scripts` | 测试 | 运行工具脚本的单元与`smoke`测试 |
| `make i18n.check` | 国际化 | 扫描运行时硬编码文案并校验语言包`key`覆盖 |
| `make plugins.init` | 插件工作区 | 将官方插件子模块转换为普通目录 |
| `make plugins.install` | 插件工作区 | 安装配置中的源码插件到`apps/lina-plugins` |
| `make plugins.update` | 插件工作区 | 更新`apps/lina-plugins`中的源码插件 |
| `make plugins.status` | 插件工作区 | 查看源码插件工作区状态 |
| `make db.init` | 数据库 | 初始化数据库表结构和种子数据 |
| `make db.mock` | 数据库 | 加载演示`Mock`数据 |
| `make release.tag.check` | 发布治理 | 校验`release tag`与`metadata.yaml`中`framework.version`一致 |
| `make help` | 其他 | 查看所有可用指令 |


## 开发服务

### make dev.setup

安装开发与`E2E`测试所需的全部前置依赖，包括前端`npm`包和`Playwright`浏览器。在首次克隆仓库或`CI`环境初始化时执行一次即可，后续不需要重复运行。

```bash
make dev.setup
```

### make dev

重启后端和前端开发服务器。执行前会先停止已有服务，依次完成`WASM`插件构建、前端静态资源准备和后端编译，等待两端健康检查通过后打印运行状态。支持通过`skip_wasm=true`跳过`WASM`构建步骤以加快启动速度。

```bash
make dev

# 跳过 WASM 插件构建（适用于不涉及动态插件的开发场景）
make dev skip_wasm=true
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
+----------+---------+------------------------+-------+------------------------+--------------------+
| Service  | Status  | URL                    | PID   | PID File               | Log File           |
+----------+---------+------------------------+-------+------------------------+--------------------+
| Backend  | running | http://127.0.0.1:8080/ | 87739 | temp/pids/backend.pid  | temp/lina-core.log |
| Frontend | running | http://127.0.0.1:5666/ | 87740 | temp/pids/frontend.pid | temp/lina-vben.log |
+----------+---------+------------------------+-------+------------------------+--------------------+
```

## 构建

### make build

完整构建流程，依次执行：前端静态资源构建、嵌入到后端的`manifest`资源准备、所有`WASM`插件构建，最后编译后端主框架二进制。构建产物输出到`temp/output/`目录。

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
| `build.binaryName` | `lina` | 主框架二进制文件名 |

### make wasm

单独构建运行时`WASM`插件。`make wasm`是兼容入口，默认把产物输出到`temp/output/`，支持使用`p=<plugin-id>`只构建指定插件。需要覆盖输出目录或只做构建探测时，直接使用`linactl wasm`。

```bash
# 构建所有 WASM 插件
make wasm

# 只构建指定插件（plugin-id 为插件目录名）
make wasm p=my-plugin
```

### make pack.assets

准备主框架`manifest`资产，用于`Go`嵌入，通常由`make build`或`make dev`自动调用。需要单独检查或准备嵌入资源时可以手动执行：

```bash
make pack.assets
```

### make tidy

整理主框架、开发工具和插件相关`Go`模块依赖，适合在升级依赖或初始化插件完整模式后执行：

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

### make image.build

仅准备镜像构建的所有产物（等价于先执行`make build`），不执行`Docker build`步骤。适用于需要手动检查产物或自定义镜像构建步骤的场景。

```bash
make image.build
```

## 测试

### make test

运行完整的`Playwright E2E`测试套件。执行前请确保开发服务已通过`make dev`启动。支持通过`scope`参数缩小测试范围：

| `scope`值 | 说明 |
|-----------|------|
| `full`（默认） | 运行全部`E2E`测试 |
| `host` | 仅运行主框架自有测试 |
| `plugins` | 仅运行所有官方插件测试 |
| `plugin:<id>` | 仅运行指定插件测试 |

```bash
make test

# 只运行主框架测试
make test scope=host

# 只运行指定插件测试
make test scope=plugin:multi-tenant
```

### make test.go

运行所有受维护`Go`模块的单元测试，并启用竞态检测。支持通过`plugins=0`强制主框架模式，或通过`race=false`关闭竞态检测。

```bash
make test.go
make test.go plugins=0
make test.go race=false
```

### make test.host

只运行主框架自有`Playwright E2E`测试，不要求初始化官方插件子模块。

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

扫描运行时可见的代码路径，检测未被纳入国际化体系的硬编码文案，并校验主框架和各插件运行时语言包的消息`key`覆盖情况。适合在提交新功能前进行`i18n`合规自查。

```bash
make i18n.check
```

## 插件工作区

`apps/lina-plugins/`目录用于存放官方插件，既可以是`Git submodule`，也可以通过插件工作区命令管理为源码插件。以下命令提供对插件工作区的完整生命周期管理：

### make plugins.init

将`apps/lina-plugins`从`Git submodule`转换为普通插件目录，同时保留插件代码。转换后可以自由修改插件代码并提交变更，不再受`submodule`约束。

```bash
make plugins.init
```

### make plugins.install

根据`hack/config.yaml`中的`plugins.sources`配置，将远端插件仓库中的指定插件克隆到`apps/lina-plugins/`。支持通过`p=<plugin-id>`只安装一个插件，通过`source=<name>`只处理指定来源。

```bash
# 安装所有配置的源码插件
make plugins.install

# 只安装指定插件
make plugins.install p=multi-tenant

# 强制覆盖已存在的插件目录
make plugins.install force=1
```

### make plugins.update

更新`apps/lina-plugins/`中已配置的源码插件，拉取远端最新版本。有本地未提交改动的插件会被阻止更新，除非传入`force=1`。

```bash
make plugins.update
make plugins.update p=multi-tenant
make plugins.update force=1
```

### make plugins.status

查看当前官方插件工作区状态，包括已配置的插件版本、本地变更和远端更新情况。

```bash
make plugins.status
```

## 数据库

:::caution 破坏性操作

`init`和`mock`均会对数据库执行破坏性操作，因此要求显式传入`confirm`参数才能执行，防止误操作。

:::

### make db.init

初始化数据库的表结构（`DDL`）和系统必需的种子数据。后端会按`config.yaml`中`database.default.link`的配置自动选择`PostgreSQL`或`SQLite`方言，其中`PostgreSQL 14+`是默认数据存储，`SQLite`仅用于本地演示或冒烟验证。

```bash
# 仅初始化（保留现有数据）
make db.init confirm=init

# 重建数据库（清空后重新初始化）
make db.init confirm=init rebuild=true
```

### make db.mock

在`make db.init`完成之后，加载用于本地演示和开发验证的可选`Mock`数据。

```bash
make db.mock confirm=mock
```

## 其他

### make help

打印根`Makefile`及所有引入目标文件中的可用指令列表，输出按指令名称排序。

```bash
make help
```
