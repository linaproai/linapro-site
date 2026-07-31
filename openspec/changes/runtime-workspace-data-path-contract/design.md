## Context

`make dev` 通过 `linactl` 将后端 `WorkDir` 设为 `apps/lina-core`，同时把二进制、日志、wasm 产物写在仓库根 `temp/`。宿主 `internal/service/config/config_path.go` 已对 `upload.path` 与 `plugin.dynamic.storagePath` 做「相对路径锚仓库根」解析，但该逻辑是包内私有实现；i18n / apidoc 又各自复制了类似 `findRepoRoot`；插件市场 `storage.root` 则直接 `filepath.Abs`，落在 `apps/lina-core/temp/`。路径语义分裂导致开发者与后续插件配置反复踩坑。

约束：

- 不改变后端进程 CWD 与 GoFrame 配置布局。
- 相对路径默认值形状保持 `temp/...`，避免首期大规模改配置文件。
- 插件不得依赖宿主 `internal/` 包。

## Goals / Non-Goals

**Goals:**

- 提供公共、唯一的工作区根与可写路径解析实现（`pkg/runtimepath`）。
- 宿主与官方插件（至少 marketplace）共用同一解析语义。
- 环境变量可显式固定 WorkspaceRoot / DataRoot；`linactl dev` 注入。
- 通过插件能力窄接口让源码插件解析可写路径。
- 清理重复 root 探测代码。

**Non-Goals:**

- 不把后端 `WorkDir` 改成仓库根。
- 不改前端 Vite 工作目录。
- 不引入完整的「配置短名相对 DataRoot」二期语义（可选后续变更）。
- 不自动迁移磁盘上的历史 marketplace 快照（文档说明即可）。

## Decisions

### D1：公共包 `pkg/runtimepath` 作为唯一实现

- **选择**：新建 `apps/lina-core/pkg/runtimepath`，导出 `WorkspaceRoot`、`DataRoot`、`Resolve`、`ResolveFrom` 等。
- **理由**：宿主 service 与插件 backend 均可依赖 `pkg`；避免 `internal/service/config` 泄漏给插件。
- **替代**：仅 env 注入绝对路径、不建公共包 → 插件仍会各自 `Getwd`。

### D2：首期相对路径锚 WorkspaceRoot（兼容 `temp/*`）

- **选择**：`Resolve("temp/upload")` → `{WorkspaceRoot}/temp/upload`；`DataRoot()` 默认 `{WorkspaceRoot}/temp`，供工具与文档约定，但首期配置值仍带 `temp/` 前缀。
- **理由**：与现 `resolveRuntimePath` 测试与配置一致，零配置迁移。
- **替代**：配置改成相对 DataRoot 的短名（`upload`）→ 需同步所有 yaml/文档，二期再做。

### D3：环境变量优先级

```text
WorkspaceRoot:
  1. LINAPRO_WORKSPACE_ROOT（绝对路径，存在或可创建父级）
  2. 自 startDir 向上探测 go.work + apps/lina-core（或安装布局标记）
  3. 回退 startDir（通常为 Getwd）

DataRoot:
  1. LINAPRO_DATA_ROOT（绝对路径）
  2. 默认 Join(WorkspaceRoot, "temp")

Resolve(path):
  绝对路径 → Clean
  相对路径 → Join(WorkspaceRoot, path)  // 首期
```

- **理由**：CI/安装布局可用 env 固定；开发默认同 monorepo。

### D4：插件能力扩展方式

- **选择**：在 `plugincap.ConfigService` 增加 `ResolvePath(ctx, path string) (string, error)`（或独立 PathService 若接口膨胀）。优先挂在 ConfigService，因路径常来自插件配置键。
- **宿主适配**：源码插件宿主实现调用 `runtimepath.Resolve`。
- **替代**：仅文档约定插件复制解析代码 → 无法门禁。

### D5：marketplace 接入

- `resolveArtifactStoreRoot` 读取配置后调用 `runtimepath.Resolve`（或 capability），不再 `filepath.Abs` 相对 CWD。
- 默认值保持 `temp/plugin-marketplace/artifacts`。

### D6：`linactl dev` 注入

- 后端 service env 增加：
  - `LINAPRO_WORKSPACE_ROOT=<repo root>`
  - `LINAPRO_DATA_ROOT=<repo root>/temp`
- 不改变 `WorkDir=apps/lina-core`。

### D7：重复逻辑清理

- `config_path.go` 委托 `runtimepath`。
- `i18n_plugin_dynamic` / `apidoc_i18n_dynamic` 删除本地 `findRepoRoot` 与「Stat 候选路径」分支，直接使用 `GetPluginDynamicStoragePath` 已解析绝对路径（避免二次相对解析）。

## Risks / Trade-offs

- [Risk] 开发机已有 `apps/lina-core/temp/plugin-marketplace` 数据在新路径下不可见 → [Mitigation] README 说明拷贝或重新 Git 同步；可选启动日志打印解析后的 storage root。
- [Risk] 非 monorepo 部署找不到 WorkspaceRoot → [Mitigation] 生产使用绝对路径配置 + `LINAPRO_WORKSPACE_ROOT`；探测失败回退 CWD。
- [Risk] 扩展 `plugincap.ConfigService` 破坏 mock 实现 → [Mitigation] 全仓更新 test doubles；编译门禁。
- [Risk] symlink / macOS /var 临时目录路径比较失败 → [Mitigation] 测试中 EvalSymlinks，与现 config_path_test 一致。

## Migration Plan

1. 发布后相对路径统一到仓库根 `temp/`。
2. 开发者若需保留旧 marketplace 快照：  
   `mv apps/lina-core/temp/plugin-marketplace temp/`（在仓库根执行）。
3. 生产若已使用绝对 `storage.root` / `upload.path`，行为不变。
4. 回滚：恢复旧代码后路径回到 CWD 语义；磁盘文件不会自动搬回。

## Open Questions

- 无阻塞项。二期是否引入配置 `data.root` 与短名默认值另开变更。
