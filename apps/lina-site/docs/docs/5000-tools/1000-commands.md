---
slug: '/docs/commands'
title: '开发指令'
hide_title: true
description: '支持的参数选项和使用示例，涵盖 linactl、Makefile 兼容入口、Windows make.cmd 包装入口、本地环境检查与初始化、开发服务管理、完整构建、WASM 插件构建、Docker 镜像构建、插件工作区管理、智能体资源软链管理、测试验证、国际化检查、数据库初始化和发布治理等场景，帮助开发者在 macOS、Linux 和 Windows 上稳定使用同一套项目工具链。'
keywords:
  - make指令
  - 开发指令
  - 构建命令
  - LinaPro开发
  - make env.check
  - make env.setup
  - make dev
  - make build
  - make pack.assets
  - make test
  - make image
  - make image.build
  - make db.init
  - make db.upgrade
  - make wasm
  - make plugins.install
  - make plugins.status
  - make agents
  - make agents.skills.link
  - make agents.prompts.link
  - make agents.md.link
  - make release.tag.check
  - make lint
  - make lint.go
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
  - 静态检查
  - golangci-lint
  - 代码质量
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

后续文档中所有`make <指令>`示例，都可以等价替换为`cd hack/tools/linactl && go run . <指令>`，参数格式保持一致。`make`兼容入口会转发常用变量；如果需要使用`linactl`专属参数，可直接调用`go run . <指令>`。

## 指令总览

| 指令 | 分类 | 说明 |
|------|------|------|
| `make env.check` | 环境 | 检查本地开发工具、项目本地前端工具和`PostgreSQL`版本 |
| `make env.setup` | 环境 | 安装前端依赖和`Playwright Chromium`浏览器及系统依赖 |
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
| `make lint` | 静态检查 | 运行仓库静态检查门禁（当前转发到`lint.go`） |
| `make lint.go` | 静态检查 | 对选定`Go workspace`模块运行`golangci-lint`静态检查 |
| `make i18n.check` | 国际化 | 扫描运行时硬编码文案并校验语言包`key`覆盖 |
| `make plugins.init` | 插件工作区 | 将官方插件子模块转换为普通目录 |
| `make plugins.install` | 插件工作区 | 安装配置中的源码插件到`apps/lina-plugins` |
| `make plugins.update` | 插件工作区 | 更新`apps/lina-plugins`中的源码插件 |
| `make plugins.status` | 插件工作区 | 查看源码插件工作区状态 |
| `make agents` | 智能体资源 | 为单个智能体一键创建或移除技能、提示词和`AGENTS.md`相关软链 |
| `make db.init` | 数据库 | 初始化数据库表结构和种子数据 |
| `make db.upgrade` | 数据库 | 重放宿主`SQL`文件升级数据库表结构 |
| `make db.mock` | 数据库 | 加载演示`Mock`数据 |
| `make release.tag.check` | 发布治理 | 校验`release tag`与`metadata.yaml`中`framework.version`一致 |
| `make help` | 其他 | 查看所有可用指令 |


## 环境管理

### make env.check

检查本地开发环境是否满足默认开发工作流要求。该命令只读取工具版本和数据库连接信息，不修改工作区。除`Vite`和`Playwright`使用项目本地依赖外，`PostgreSQL`版本会通过`apps/lina-core/manifest/config/config.yaml`中的`database.default.link`连接数据库后查询。

```bash
make env.check
```

当前检查项如下：

| 检查项 | 最低版本 | 说明 |
|--------|----------|------|
| `Go` | `1.25.0` | 后端、工具链和`WASM`插件构建使用的`Go`工具链 |
| `Node.js` | `20.19.0` | 前端开发和构建所需的`Node.js`运行时 |
| `pnpm` | `10.0.0` | 前端依赖管理工具 |
| `Vite` | `7.3.1` | 项目本地前端构建工具，缺失时先执行`make env.setup` |
| `Playwright` | `1.58.2` | `E2E`测试运行器，缺失时先执行`make env.setup` |
| `PostgreSQL` | `14.0.0` | 通过`database.default.link`探测服务端版本 |

### make env.setup

安装开发与`E2E`测试所需的所有前置依赖，包括`Go`静态检查工具（`golangci-lint`和`staticcheck`）、前端依赖、`Playwright Chromium`浏览器和浏览器运行所需系统依赖。在首次克隆仓库或`CI`环境初始化时执行一次即可，后续通常不需要重复运行。

```bash
make env.setup
```

## 开发服务

### make dev

重启后端和前端开发服务器。执行前会先校验命令参数中的`backend_port`、后端`server.address`和前端`Vite proxy target`是否一致，避免端口错配后才暴露为健康检查超时或接口错误。随后命令会停止已有服务，按插件模式决定是否构建`WASM`插件，准备前端静态资源，编译后端并等待两端健康检查通过。

```bash
make dev

# 强制主框架模式，跳过官方源码插件和 WASM 插件构建
make dev plugins=0

# 强制启用官方源码插件模式
make dev plugins=1

# 执行指定目录的定向构建，不启动或重启开发服务
make dev dir=tools/custom-builder
```

`plugins=auto`是默认模式：当`apps/lina-plugins/`中存在可用插件`manifest`时自动启用官方插件模式，否则使用主框架模式。直接调用`linactl dev`时还可以传入`skip_wasm=true`只跳过`WASM`构建步骤。

```bash
cd hack/tools/linactl
go run . dev skip_wasm=true
```

传入`dir=<path>`时，`make dev`会复用`make build dir=<path>`的定向构建逻辑，读取目标目录下的`hack/config.yaml`或支持的构建入口，不执行默认的服务停止、编译启动和健康检查流程。

后端默认监听`http://localhost:9120`，前端开发服务器默认监听`http://localhost:5666`，管理工作台默认开发入口为`http://localhost:5666/admin`。运行日志分别写入`temp/lina-core.log`和`temp/lina-vben.log`。

### make stop

停止后端和前端开发服务器，并清理残留`PID`文件。对于仍占用端口的僵尸进程，会强制终止。

```bash
make stop

# 执行指定目录的自定义停止逻辑
make stop dir=tools/custom-builder
```

传入`dir=<path>`时，`make stop`不执行默认宿主前后端停止流程，而是读取`<path>/hack/config.yaml`中的`stop.commands`并在目标目录执行。该模式适合自定义工具、插件或本地扩展目录维护自己的停止逻辑；目标目录缺少`hack/config.yaml`时命令会快速失败。

### make status

打印前后端当前的运行状态及日志文件路径，便于快速确认服务是否正常启动。

```bash
make status

# 执行指定目录的自定义状态查询逻辑
make status dir=tools/custom-builder
```

传入`dir=<path>`时，`make status`不打印默认宿主服务状态表，而是读取`<path>/hack/config.yaml`中的`status.commands`并在目标目录执行。该模式用于让独立工具或扩展目录提供自己的状态检查输出。

输出示例：

```text
+----------+---------+------------------------+-------+------------------------+--------------------+
| Service  | Status  | URL                    | PID   | PID File               | Log File           |
+----------+---------+------------------------+-------+------------------------+--------------------+
| Backend  | running | http://127.0.0.1:9120/ | 87739 | temp/pids/backend.pid  | temp/lina-core.log |
| Frontend | running | http://127.0.0.1:5666/ | 87740 | temp/pids/frontend.pid | temp/lina-vben.log |
+----------+---------+------------------------+-------+------------------------+--------------------+
```

## 代码构建

### make build

完整构建流程，依次执行：前端静态资源构建、嵌入到后端的静态资源和`manifest`资源准备、按插件模式构建动态`WASM`插件，最后编译后端主框架二进制。构建产物输出到`temp/output/`目录。

```bash
# 默认构建（当前平台）
make build

# 强制主框架模式，不构建官方源码插件
make build plugins=0

# 只构建指定官方插件
make build dir=apps/lina-plugins/my-plugin

# 执行任意目录下 hack/config.yaml 定义的定向构建
make build dir=tools/custom-builder

# 指定目标平台（交叉编译）
make build platforms=linux/amd64,linux/arm64

# 覆盖构建产物目录和二进制名称
make build output_dir=temp/release binary_name=linapro

# 覆盖配置文件
make build config=hack/config.yaml

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

插件构建模式通过`plugins`参数控制：

| `plugins`值 | 说明 |
|-------------|------|
| `auto`（默认） | 当`apps/lina-plugins/`存在可用插件`manifest`时启用官方源码插件模式 |
| `0` | 强制主框架模式，移除官方插件构建标签并跳过官方插件`WASM`构建 |
| `1` | 强制启用官方源码插件模式；如果插件工作区不可用，命令会快速失败 |

`dir=<path>`用于定向执行单个目标目录。内置目标保留快捷路径：`apps/lina-vben`只构建默认前端并同步嵌入资源；`apps/lina-core`构建默认前端、准备嵌入资源并编译后端二进制；`apps/lina-plugins/<plugin-id>`使用官方插件构建环境。其他目录优先读取自身`hack/config.yaml`；没有该配置时，`make build dir=<path>`才回退到本地`package.json`的`build`脚本。

#### 定向目录配置

目标目录可以在自身`hack/config.yaml`中为不同命令维护自定义指令。`make build dir=<path>`和`make dev dir=<path>`读取`build.commands`；`make stop dir=<path>`读取`stop.commands`；`make status dir=<path>`读取`status.commands`。这些指令都在所选目录执行，并支持以下变量：

| 变量 | 含义 |
|------|------|
| `$(TARGET_DIR)` | 当前`dir`解析后的绝对目录 |
| `$(BUILD_DIR)` | 同`$(TARGET_DIR)`，用于构建命令命名 |
| `$(PLUGIN_ROOT)` | 同`$(TARGET_DIR)`，作为插件旧配置的兼容别名 |
| `$(REPO_ROOT)` | 当前仓库根目录 |

```yaml
build:
  commands:
    - pnpm --dir "$(BUILD_DIR)/frontend" run build
stop:
  commands:
    - node scripts/stop.mjs --root "$(TARGET_DIR)"
status:
  commands:
    - node scripts/status.mjs --root "$(TARGET_DIR)"
```

当`dir=apps/lina-plugins/<plugin-id>`指向官方插件时，`plugins`参数仍只控制官方插件模式，不用于选择具体插件。源码插件会使用官方插件构建环境；动态插件会在`build.commands`完成后继续生成自身`WASM`产物并写入本次构建输出目录。`make dev dir=<path>`只执行上述定向构建路径，不启动或重启开发服务；`make stop dir=<path>`和`make status dir=<path>`则用目录配置替代默认宿主服务停止和状态查询流程。

### make wasm

单独构建运行时`WASM`插件。`make wasm`是兼容入口，默认把产物输出到`temp/output/`，支持使用`p=<plugin-id>`只构建指定插件，并可通过`dry_run=true`只检查可构建插件而不写产物。需要构建任意路径下的单个插件目录时，直接使用`linactl wasm plugin_dir=<path>`。

```bash
# 构建所有 WASM 插件
make wasm

# 只构建指定插件（plugin-id 为插件目录名）
make wasm p=my-plugin

# 只检查构建计划
make wasm dry_run=true

# 直接构建指定插件目录
cd hack/tools/linactl
go run . wasm plugin_dir=../../apps/lina-plugins/my-plugin out=../../temp/output
```

### make pack.assets

准备主框架`manifest`资产，用于`Go`嵌入。该命令会刷新`apps/lina-core/internal/packed/manifest/`下的`config`、`sql`和`i18n`资源，通常由`make build`或`make dev`自动调用。需要单独检查或准备嵌入资源时可以手动执行：

```bash
make pack.assets
```

### make tidy

整理主框架、开发工具和插件相关`Go`模块依赖，适合在升级依赖或初始化插件完整模式后执行：

```bash
make tidy
```

## 镜像编译

### make image

完整的`Docker`镜像构建流程：先执行镜像构建预检，再执行`make build`生成所有构建产物，最后调用`hack/tools/image-builder`封装成镜像。镜像名称、标签、镜像仓库地址、基础镜像和插件构建模式等均通过参数配置。

```bash
# 使用默认配置构建镜像
make image

# 指定标签和镜像仓库
make image tag=v0.6.0 registry=ghcr.io/linaproai

# 构建后直接推送
make image tag=v0.6.0 registry=ghcr.io/linaproai push=1

# 多平台构建
make image platforms=linux/amd64,linux/arm64 tag=v0.6.0

# 覆盖运行时基础镜像并使用主框架模式
make image base_image=alpine:3.22 plugins=0
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

仅准备镜像构建的所有产物（等价于先执行`make build`并生成镜像构建上下文），不执行`Docker build`步骤。适用于需要手动检查产物或自定义镜像构建步骤的场景。

```bash
make image.build
```

## 测试管理

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

运行所有受维护`Go`模块的单元测试，并默认启用竞态检测和详细日志。命令会先发现当前`Go workspace`中的模块，把有测试文件的包作为真实测试执行，把没有测试文件的包作为编译冒烟检查执行，并按模块输出汇总。支持通过`plugins=0`强制主框架模式，或通过`race=false`关闭竞态检测。

```bash
make test.go
make test.go plugins=0
make test.go race=false
make test.go verbose=false
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

## 静态检查

### make lint

运行仓库静态检查门禁。当前转发到`lint.go`，未来可扩展集成其他语言的静态检查工具。

```bash
make lint
```

### make lint.go

对选定`Go workspace`模块运行`golangci-lint`静态检查。检查规则在仓库根目录`.golangci.yml`中集中配置，涵盖缺陷检测、资源管理、错误处理、安全检查、性能优化和可维护性等多个质量维度。

```bash
# 检查宿主工作区（apps/lina-core 和 hack/tools/linactl）
make lint.go plugins=0

# 检查宿主和官方插件工作区
make lint.go plugins=1

# 自动修复可纠正的问题
make lint.go plugins=0 fix=true
```

`plugins`参数控制检查范围：

| `plugins`值 | 说明 |
|-------------|------|
| `auto`（默认） | 自动探测：当`apps/lina-plugins/`存在可用插件`manifest`时启用官方插件模式 |
| `0` | 强制主框架模式，只检查宿主工作区 |
| `1` | 强制启用官方源码插件模式，检查宿主、工具和官方插件`Go`模块 |

静态检查工具版本由仓库文件精确锁定：`golangci-lint`版本记录在`.golangci-lint-version`，`staticcheck`版本记录在`.staticcheck-version`。`linactl lint.go`会自动安装缺失或不匹配的版本，确保团队成员和`CI`环境使用完全一致的分析行为。

`fix=true`是显式开发者操作，允许`golangci-lint`在支持时自动修复导入和格式问题。`CI`流水线不会启用该参数。

## 插件管理

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

## I18N国际化

### make i18n.check

扫描运行时可见的代码路径，检测未被纳入国际化体系的硬编码文案，并校验主框架和各插件运行时语言包的消息`key`覆盖情况。适合在提交新功能前进行`i18n`合规自查。

```bash
make i18n.check
```


## AI工具集成

`agents`系列命令用于管理仓库内面向不同`AI Coding Agent`的资源软链。`LinaPro`把项目规范、技能和提示词统一维护在仓库标准位置，再通过软链适配不同工具自己的项目路径。

| 统一来源 | 作用 | 典型桥接示例 |
|----------|------|--------------|
| `AGENTS.md` | 项目顶层规范入口 | `CLAUDE.md`、`GEMINI.md`、`QWEN.md` |
| `.agents/skills` | 项目内置`AI`技能 | `.claude/skills`、`.goose/skills`、`.augment/skills` |
| `.agents/prompts` | 项目命令和提示词目录 | `.claude/commands`、`.codex/prompts`、`.cursor/commands` |

这样做的目标是让团队只维护一份规范和一套技能，同时兼容开发者习惯使用的不同`AI Coding`工具。原生支持`AGENTS.md`或`.agents/skills`的工具会被识别为原生支持，命令只展示状态，不会重复创建不必要的软链。

这些命令只管理自己创建的软链；移除命令不会删除真实目录或文件，也不会移除指向外部目标的非托管软链。遇到已存在但目标不一致的软链时，通常需要传入`FORCE=1`才会重建。

### make agents

推荐使用的聚合入口。无参数且连接到交互式终端时，会先用方向键选择智能体，再选择`link`或`unlink`操作；在非交互环境中，通过`agent=<name>`指定单个智能体。聚合入口会对该智能体支持的所有资源类型执行同一操作，不支持`agent=all`或逗号列表。

```bash
# 交互式选择智能体和操作
make agents

# 为单个智能体创建所有可用资源软链
make agents agent=claude-code

# 移除单个智能体的托管软链
make agents agent=claude-code action=unlink

# 重建目标不一致的托管软链
make agents agent=claude-code force=1
```

如果智能体原生读取某类资源，例如原生读取`AGENTS.md`，聚合入口会跳过该资源并在汇总中说明原因。

## 数据库

:::caution 破坏性操作

`init`和`mock`均会对数据库执行破坏性操作，因此要求显式传入`confirm`参数才能执行，防止误操作。

:::

### make db.init

初始化数据库的表结构（`DDL`）和系统必需的种子数据。命令读取`apps/lina-core/manifest/config/config.yaml`中的`database.default.link`，当前仅支持`PostgreSQL 14+`方言；`sqlite:`、`mysql:`或未知链接会在方言解析阶段快速失败，不会创建本地数据库文件或继续执行`SQL`。

```bash
# 仅初始化（保留现有数据）
make db.init confirm=init

# 重建数据库（清空后重新初始化）
make db.init confirm=init rebuild=true
```

如果`PostgreSQL`无法连接，命令会提示先启动`PostgreSQL`，并输出本地`docker run`示例。`rebuild=true`会先终止目标数据库连接，再执行`DROP DATABASE IF EXISTS`和`CREATE DATABASE`，只应在确认可以清空目标库时使用。

### make db.upgrade

重放宿主框架的`SQL`文件，将数据库表结构升级到最新状态。该命令读取`apps/lina-core/manifest/sql/`目录下所有`.sql`文件，按文件名排序依次执行。所有`SQL`语句使用`CREATE TABLE IF NOT EXISTS`和`CREATE INDEX IF NOT EXISTS`等幂等语法，因此重放不会破坏已有数据，只会将缺失的表和索引补全。

```bash
make db.upgrade confirm=upgrade
```

该命令只处理宿主框架的`SQL`文件，不处理插件自带的`SQL`。如果需要升级插件的表结构，应参考各插件文档中的说明。如果`PostgreSQL`无法连接，命令会提示先启动数据库并输出`docker run`示例。

### make db.mock

在`make db.init`完成之后，加载用于本地演示和开发验证的可选`Mock`数据。

```bash
make db.mock confirm=mock
```

## 其他

### make help

打印默认可用的跨平台开发指令列表，输出按指令名称排序。`linactl`内部还注册了`cli`、`cli.install`、`ctrl`和`dao`等维护命令，默认帮助不会展示；只有直接执行`go run . help --all`时才会列出。

```bash
make help
```
