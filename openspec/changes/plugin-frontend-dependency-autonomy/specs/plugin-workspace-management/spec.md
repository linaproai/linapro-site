## ADDED Requirements

### Requirement: `linactl` 必须准备源码插件前端 package 依赖

系统 SHALL 在默认开发工具入口中发现并准备源码插件 `frontend/package.json` 声明的前端依赖。依赖准备必须通过跨平台 `linactl` 工具实现，不得依赖 shell 脚本或要求开发者手动逐个插件安装。

#### Scenario: 环境初始化安装插件前端依赖

- **WHEN** 开发者运行 `make env.setup` 或 `linactl env.setup`
- **AND** `apps/lina-plugins/<plugin-id>/frontend/package.json` 存在
- **THEN** 工具在安装宿主 `apps/lina-vben` 依赖后准备该插件 frontend 依赖
- **AND** 插件私有依赖可被后续宿主 Vite 构建解析

#### Scenario: 开发服务启动前准备插件前端依赖

- **WHEN** 开发者运行 `make dev` 或 `linactl dev`
- **AND** 源码插件 frontend package 依赖尚未安装
- **THEN** 工具在启动 Vite 前准备这些依赖
- **AND** Vite 启动后插件普通模块的 bare import 不因缺少依赖安装而失败

#### Scenario: 宿主构建前准备插件前端依赖

- **WHEN** 开发者运行 `make build` 或 `linactl build`
- **AND** 源码插件启用参与宿主前端构建
- **THEN** 工具在执行宿主前端构建前准备插件 frontend 依赖
- **AND** 不能依赖插件 `hack/config.yaml build.commands` 才安装这些依赖

#### Scenario: 无 frontend package 的插件跳过依赖安装

- **WHEN** `apps/lina-plugins/<plugin-id>/frontend/package.json` 不存在
- **THEN** 工具不为该插件执行前端依赖安装
- **AND** 不将缺少 frontend package 视为错误

### Requirement: 插件前端依赖准备必须保持跨平台

系统 SHALL 使用 Go 标准库扫描插件目录并通过明确子进程调用 `pnpm` 完成安装。实现不得新增 `.sh`、`.ps1` 或依赖平台专属命令的默认开发入口。

#### Scenario: Windows、Linux 和 macOS 使用同一入口

- **WHEN** 开发者在 Windows、Linux 或 macOS 上运行默认开发命令
- **THEN** 插件 frontend package 发现逻辑使用 Go 文件系统 API
- **AND** 子进程工作目录使用平台正确路径
- **AND** 不要求 POSIX shell、PowerShell 专属语法或 Unix 工具存在
