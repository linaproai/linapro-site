---
slug: '/docs/commands'
title: '开发指令'
hide_title: true
description: '本文介绍 LinaPro 项目中所有 make 指令的作用、支持的参数选项和使用示例，涵盖开发服务管理、完整构建、WASM 插件构建、Docker 镜像构建、测试验证、国际化检查和数据库初始化等所有场景，帮助开发者熟练使用项目工具链。'
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
  - 开发工作流
  - 后端构建
  - 前端构建
  - Docker镜像
  - E2E测试
  - 数据库初始化
---

`LinaPro`项目根目录提供了一套完整的`make`指令，通过`hack/makefiles/`下的分模块文件统一管理。在项目根目录执行`make help`可随时查看所有可用指令。

## 指令总览

| 指令 | 分类 | 说明 |
|------|------|------|
| `make dev` | 开发服务 | 重启前后端开发服务器 |
| `make stop` | 开发服务 | 停止前后端开发服务器 |
| `make status` | 开发服务 | 查看前后端运行状态及日志路径 |
| `make build` | 构建 | 完整构建前端、插件和后端二进制 |
| `make wasm` | 构建 | 构建所有或指定运行时`WASM`插件 |
| `make image` | 镜像 | 构建生产`Docker`镜像 |
| `make image-build` | 镜像 | 仅准备镜像产物，不执行`Docker`构建 |
| `make test` | 测试 | 运行完整`E2E`测试套件 |
| `make test-scripts` | 测试 | 运行工具脚本的单元与`smoke`测试 |
| `make check-runtime-i18n` | 国际化 | 扫描运行时硬编码文案 |
| `make check-runtime-i18n-messages` | 国际化 | 校验运行时语言包`key`覆盖情况 |
| `make init` | 数据库 | 初始化数据库表结构和种子数据 |
| `make mock` | 数据库 | 加载演示`Mock`数据 |
| `make help` | 其他 | 查看所有可用指令 |

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
    - "linux/amd64"
  # 是否启用 CGO
  cgoEnabled: false 
  # 构建产物输出路径，相对于仓库根目录
  outputDir: "temp/output" 
  # 编译后生成的二进制文件名
  binaryName: "lina"       
```

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `build.platforms` | `["linux/amd64"]` | 目标平台列表，使用`goos/goarch`格式，`make build platforms=...`可覆盖 |
| `build.cgoEnabled` | `false` | 是否启用`CGO` |
| `build.outputDir` | `temp/output` | 构建产物输出路径，相对于仓库根目录 |
| `build.binaryName` | `lina` | 宿主二进制文件名 |

### make wasm

单独构建运行时`WASM`插件，产物输出到`temp/output/`。支持使用`p=<plugin-id>`只构建指定插件，省略时构建全部。

```bash
# 构建所有 WASM 插件
make wasm

# 只构建指定插件（plugin-id 为插件目录名）
make wasm p=my-plugin

# 开启详细日志
make wasm verbose=1
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

### make test-scripts

运行`hack/tests/scripts/`下所有工具脚本的单元和`smoke`测试，用于验证仓库辅助脚本的基本正确性。

```bash
make test-scripts
```

## 国际化

### make check-runtime-i18n

扫描运行时可见的代码路径，检测未被纳入国际化体系的硬编码文案。适合在提交新功能前进行`i18n`合规自查。

```bash
make check-runtime-i18n
```

### make check-runtime-i18n-messages

校验宿主和各插件运行时语言包的消息`key`覆盖情况，检测缺失或多余的翻译`key`。

```bash
make check-runtime-i18n-messages
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
