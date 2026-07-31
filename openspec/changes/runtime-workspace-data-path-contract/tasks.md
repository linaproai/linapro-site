## 1. 公共运行时路径包

- [x] 1.1 新增 `apps/lina-core/pkg/runtimepath`：WorkspaceRoot / DataRoot / Resolve / 环境变量与探测逻辑
- [x] 1.2 为 `runtimepath` 补充单测（monorepo 子目录 CWD、env 覆盖、绝对路径、回退）

## 2. 宿主统一接入

- [x] 2.1 改造 `internal/service/config/config_path.go` 委托 `runtimepath`，保持对外路径语义
- [x] 2.2 删除/收敛 i18n 与 apidoc 中重复的 repo root 探测与二次相对路径解析
- [x] 2.3 运行宿主相关单测确认上传/动态插件路径仍锚仓库根

## 3. 插件能力与 marketplace

- [x] 3.1 扩展 `plugincap.ConfigService` 增加 `ResolvePath`，并实现源码插件宿主适配与测试 double 更新
- [x] 3.2 marketplace `resolveArtifactStoreRoot` / 制品存储改为统一解析（capability 或 `runtimepath`）
- [x] 3.3 更新 marketplace 单测：相对路径落在仓库根 temp，绝对路径不受 CWD 影响

## 4. 开发工具与文档

- [x] 4.1 `linactl dev` 后端 env 注入 `LINAPRO_WORKSPACE_ROOT` 与 `LINAPRO_DATA_ROOT`
- [x] 4.2 更新 marketplace README 与配置注释中的路径语义与迁移说明
- [x] 4.3 运行 `openspec validate runtime-workspace-data-path-contract --strict` 与关键单测门禁

### 任务记录

- **DI 来源检查**：无新增运行期服务依赖；`runtimepath` 为无状态公共包，`ConfigService.ResolvePath` 在既有 `pluginconfig.serviceAdapter` 与 domainhostcall 适配器上实现，不引入新构造函数依赖。
- **i18n**：无用户可见文案变更。
- **缓存一致性**：无影响。
- **数据权限**：无影响。
- **开发工具跨平台**：路径使用 `filepath`；`linactl` 注入绝对 `root`/`temp`；macOS symlink 在测试中用 canonicalize 比较。
- **测试策略**：`pkg/runtimepath`、config path、i18n/apidoc 路径、marketplace artifact store、linactl devservice 单测；marketplace backend 全量包测试通过。
