## 1. catalog 作为可推导唯一来源

- [x] 1.1 确认 catalog 能表达 dispatcher/guest 发布与动态专用标记
- [x] 1.2 WASM 分发注册改为由 catalog + handler 图构建，删除平行方法清单
- [x] 1.3 缺 handler 与孤儿 handler 必须让构建或治理测试失败
- [x] 1.4 动态 guest 核心方法表从同一 catalog 推导

## 2. 收敛重复定义

- [x] 2.1 资源种类枚举只保留 catalog 一份
- [x] 2.2 `Runtime`/`Network`/`RecordStore`保持动态专用，不并入`capability.Services`
- [x] 2.3 不退役存量 dedicated codec，生成/绑定逻辑同时支持 JSON 与 dedicated

## 3. 验证

- [x] 3.1 升级 catalog 覆盖测试，证明只改一处方法名会失败
- [x] 3.2 运行`pluginbridge`与 wasm 相关包测试，确认 wire 行为不变
- [x] 3.3 运行`openspec validate plugin-hostservice-catalog --strict`

### 任务记录

- **DI 来源检查**：无新增运行期服务。`buildHostServiceDispatchRegistry`在`wasm`包内用 catalog 方法集绑定已有`dispatchXxxHostService`；`defaultHostServiceDispatchRegistry`仍是包内`sync.Once`，首次分发时构造，不进入`httpstartup`依赖图，不另建 cache/lock/session 实例。
- **i18n**：无影响。无运行时文案、apidoc 源文本、语言包或`plugin.yaml`变更。
- **缓存一致性**：不改变 cache host service 语义、修订或失效路径；仅把`cache`方法注册改为由 catalog 推导。
- **数据权限**：无影响。无列表/详情/导出或授权边界变更。
- **开发工具跨平台**：未新增 generate 脚本；注册胶水是显式`BuildRegistry`，Unix/Windows 无新入口。
- **测试策略**：治理测试 + 协议/wasm/guest 单测。无用户可观察行为，不新增 E2E。未触发 E2E 质量审查。
- **文档**：同步更新`pkg/plugin/README.md`与`README.zh-CN.md`维护说明；生成 catalog 表未改列。
