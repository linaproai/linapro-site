## ADDED Requirements

### Requirement: 源码插件必须能通过宿主能力解析可写路径

系统 SHALL 向源码插件暴露路径解析能力，使插件配置中的相对可写路径与宿主使用同一 WorkspaceRoot 锚定语义。插件业务代码 MUST NOT 依赖 `os.Getwd` 与 `filepath.Abs` 组合作为可写数据根的权威解析方式。

#### Scenario: 插件解析相对制品路径

- **WHEN** 源码插件请求将配置值 `temp/plugin-marketplace/artifacts` 解析为绝对路径
- **AND** WorkspaceRoot 为 monorepo 根 `/repo`
- **THEN** 返回路径 MUST 为 `/repo/temp/plugin-marketplace/artifacts`

#### Scenario: 插件解析绝对路径

- **WHEN** 源码插件请求解析绝对路径 `/data/marketplace`
- **THEN** 返回路径 MUST 为规范化后的 `/data/marketplace`

### Requirement: 插件市场制品存储根必须使用统一路径解析

插件市场（`linapro-plugin-marketplace`）的本地制品与 Git 文档快照存储根 MUST 通过统一运行时路径解析（宿主能力或等价公共解析）得到绝对路径。默认相对配置 `temp/plugin-marketplace/artifacts` 在 monorepo 开发态下 MUST 落在仓库根 `temp/` 树下，而非 `apps/lina-core/temp/`（当 WorkspaceRoot 为仓库根时）。

#### Scenario: Git 文档快照写入仓库根 temp

- **WHEN** 开发态 WorkspaceRoot 为 monorepo 根
- **AND** marketplace `storage.root` 为默认相对路径 `temp/plugin-marketplace/artifacts`
- **AND** 系统完成某插件版本的 Git 文档索引
- **THEN** 文档快照文件 MUST 写入 `{WorkspaceRoot}/temp/plugin-marketplace/artifacts/<plugin-id>/<version>/docs/...`（或同根下的版本级 `meta/docs-manifest.json`）
- **AND** MUST NOT 以进程 CWD 为唯一锚点写入 `{CWD}/temp/plugin-marketplace/artifacts/...`（当 CWD 为 `apps/lina-core` 时）

#### Scenario: 绝对 storage.root 不受 CWD 影响

- **WHEN** marketplace `storage.root` 配置为绝对路径
- **THEN** 制品存储 MUST 使用该绝对路径
- **AND** 行为 MUST 与进程工作目录无关
