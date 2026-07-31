## ADDED Requirements

### Requirement: 工作区根必须可稳定发现

系统 SHALL 提供公共运行时路径解析能力，用于确定工作区根（WorkspaceRoot）。解析优先级 MUST 为：环境变量 `LINAPRO_WORKSPACE_ROOT`（绝对路径）优先；否则自给定起始目录向上探测仓库标记（至少包括 `go.work` 与 `apps/lina-core` 布局，或与现有安装布局等价标记）；若仍无法确定，则 MUST 回退到起始目录（通常为进程工作目录）。业务模块 MUST NOT 各自复制互不一致的仓库根探测实现。

#### Scenario: 从 apps/lina-core 工作目录发现 monorepo 根

- **WHEN** 进程工作目录为 monorepo 下的 `apps/lina-core`
- **AND** 仓库根存在 `go.work` 与 `apps/lina-core`
- **AND** 未设置 `LINAPRO_WORKSPACE_ROOT`
- **THEN** WorkspaceRoot MUST 解析为该 monorepo 根目录

#### Scenario: 环境变量覆盖工作区根

- **WHEN** 环境变量 `LINAPRO_WORKSPACE_ROOT` 设置为某绝对路径
- **THEN** WorkspaceRoot MUST 使用该绝对路径
- **AND** MUST 不再依赖目录向上探测结果覆盖该值

### Requirement: 相对可写路径必须锚在工作区根

系统 SHALL 将配置中的相对文件系统路径解析为绝对路径。绝对路径 MUST 仅做规范化清理；相对路径 MUST 相对于 WorkspaceRoot 拼接（首期语义，兼容配置值形如 `temp/upload`、`temp/output`、`temp/plugin-marketplace/artifacts`）。系统 MUST NOT 将业务可写相对路径解析为「仅相对进程 CWD」的结果（当 WorkspaceRoot 可发现且与 CWD 不同时）。

#### Scenario: 相对 temp 路径锚到仓库根

- **WHEN** WorkspaceRoot 为 monorepo 根 `/repo`
- **AND** 配置路径为 `temp/upload`
- **AND** 进程 CWD 为 `/repo/apps/lina-core`
- **THEN** 解析结果 MUST 为 `/repo/temp/upload`
- **AND** MUST NOT 为 `/repo/apps/lina-core/temp/upload`

#### Scenario: 绝对路径保持不变

- **WHEN** 配置路径为绝对路径 `/var/lib/linapro/upload`
- **THEN** 解析结果 MUST 为规范化后的该绝对路径
- **AND** MUST 忽略 WorkspaceRoot 拼接

### Requirement: 数据根环境变量可覆盖默认 temp 目录

系统 SHALL 支持环境变量 `LINAPRO_DATA_ROOT` 指定数据根。当未设置时，DataRoot MUST 默认为 `{WorkspaceRoot}/temp`。DataRoot 主要用于工具注入、文档约定与后续短名语义；首期带 `temp/` 前缀的配置值仍按 WorkspaceRoot 解析。

#### Scenario: 默认数据根

- **WHEN** 未设置 `LINAPRO_DATA_ROOT`
- **AND** WorkspaceRoot 为 `/repo`
- **THEN** DataRoot MUST 为 `/repo/temp`

#### Scenario: 显式数据根

- **WHEN** `LINAPRO_DATA_ROOT` 为 `/var/lib/linapro`
- **THEN** DataRoot MUST 为 `/var/lib/linapro`

### Requirement: 宿主上传与动态插件存储必须使用统一解析

宿主配置读取路径（至少包括上传目录与动态插件存储目录）MUST 通过统一运行时路径解析能力得到绝对路径，且行为 MUST 与 WorkspaceRoot 锚定语义一致。

#### Scenario: 上传目录与动态插件目录共享锚点

- **WHEN** 配置 `upload.path` 为 `temp/upload` 且 `plugin.dynamic.storagePath` 为 `temp/output`
- **AND** WorkspaceRoot 可发现
- **THEN** 宿主对外返回的上传目录与动态插件存储目录 MUST 均位于同一 WorkspaceRoot 之下

### Requirement: 开发启动必须注入工作区与数据根环境变量

本地开发入口（`linactl dev` / `make dev`）在启动后端进程时 MUST 注入 `LINAPRO_WORKSPACE_ROOT` 与 `LINAPRO_DATA_ROOT`，使其分别指向仓库根与 `{仓库根}/temp`，从而固定开发态路径语义且不改变后端进程工作目录。

#### Scenario: make dev 注入环境变量

- **WHEN** 通过 `linactl dev` 启动后端
- **THEN** 后端进程环境 MUST 包含指向仓库根的 `LINAPRO_WORKSPACE_ROOT`
- **AND** MUST 包含指向 `{仓库根}/temp` 的 `LINAPRO_DATA_ROOT`
- **AND** 后端进程工作目录仍 MUST 为 `apps/lina-core`（或等价宿主模块目录）
