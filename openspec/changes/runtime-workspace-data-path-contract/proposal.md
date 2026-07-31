## Why

开发态下 `make dev` 将后端进程工作目录设为 `apps/lina-core`，而 `linactl` 日志/二进制已落在仓库根 `temp/`。宿主上传与动态插件存储已通过内部 `resolveRuntimePath` 锚到仓库根，但插件市场等业务仍用 `os.Getwd()` + `filepath.Abs` 解析可写路径，导致文档快照落在 `apps/lina-core/temp/`，路径语义分裂。需要建立全框架统一的「工作区根 / 数据路径」契约，避免后续每个新存储配置再踩同一问题。

## What Changes

- 将仓库根锚点与相对可写路径解析提升为公共契约（`pkg/runtimepath`），作为唯一实现，取代各处复制的 `findRepoRoot` / `Getwd`+`Abs`。
- 宿主 `GetUploadPath`、`GetPluginDynamicStoragePath` 等改为调用公共解析器；删除 i18n/apidoc 中重复探测逻辑。
- 支持环境变量覆盖：`LINAPRO_WORKSPACE_ROOT`、`LINAPRO_DATA_ROOT`（DataRoot 默认 `{WorkspaceRoot}/temp`，首期相对路径仍按 WorkspaceRoot 解析以兼容现有 `temp/*` 配置）。
- `linactl dev` 启动后端时注入工作区/数据根环境变量，使探测结果可显式固定。
- 插件配置中的可写相对路径（至少 marketplace `storage.root`）必须经统一解析，开发数据落到仓库根 `temp/`。
- 通过插件能力窄接口暴露路径解析（源码插件可调用），禁止业务服务自行 `Getwd` 拼数据路径。
- 更新宿主/marketplace/工具文档中的路径语义说明；说明旧 `apps/lina-core/temp/plugin-marketplace` 数据迁移方式。
- **非目标**：不改变后端进程 CWD 为仓库根；不改变前端 Vite WorkDir；不改变 GoFrame 配置文件搜索布局。

## Capabilities

### New Capabilities

- `runtime-path-resolution`：工作区根发现、数据根、相对/绝对可写路径解析规则与环境变量覆盖。
- `plugin-data-path-capability`：源码插件解析可写相对路径的宿主能力契约。

### Modified Capabilities

- （无。`plugin-marketplace` 尚未进入 `openspec/specs/` 基线；市场存储解析要求作为新能力场景覆盖。）

## Impact

- 代码：`apps/lina-core/pkg/runtimepath`（新）、`internal/service/config/config_path.go`、i18n/apidoc 重复 root 逻辑；`apps/lina-plugins/linapro-plugin-marketplace` 制品存储；`hack/tools/linactl` dev 服务环境注入。
- API：无对外 HTTP 契约变更；插件能力可能新增路径解析方法。
- 数据：marketplace 开发态落盘从 `apps/lina-core/temp/...` 迁到 `<repo>/temp/...`（需迁移或重新同步）。
- 测试：runtimepath 单测、config path 单测、marketplace artifact store 单测、linactl dev env 单测。
- 文档：marketplace README、宿主/工具 README 中路径说明。
- i18n：无用户可见文案变更（仅开发者文档）。
- 数据权限 / 缓存：无影响。
