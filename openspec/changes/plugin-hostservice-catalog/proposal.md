## Why

动态插件能调哪些宿主方法，现在至少有协议目录、能力映射和`WASM`手工注册表三本账。`catalog.go`已经自称唯一来源，测试也在对账，但注册表和 guest 客户端仍靠人手同步。每加一个核心领域方法就要改多处，漂移只会随插件表面变大而变贵。

## What Changes

- 以`pkg/plugin/pluginbridge/protocol/hostservices`目录为可生成的唯一来源，推导`WASM`分发注册和动态插件 guest 客户端方法表。
- 源码插件继续使用`capability.Services`。动态侧`pluginbridge.Services`只做传输适配；`Runtime`、`Network`、`RecordStore`等仅服务动态插件的入口保持为动态侧额外能力，不再手写一份几乎相同的核心方法清单。
- 新领域方法继续只走 JSON envelope，与既有`plugin-host-layer-simplification`一致。
- 本变更不退役 cache/lock/storage/data 的存量二进制编解码，不扩展或删除`RecordStore`，不拆`plugin.Service`总门面。
- 资源种类等枚举以目录为单一来源，禁止能力注册表再抄一份互相漂移的定义。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `plugin-host-layer-simplification`：catalog 不仅是治理对账，还必须能推导 WASM 注册与 guest 客户端。
- `plugin-host-service-extension`：动态插件宿主方法的发布面以 catalog 为唯一规格，分发器不得维护平行方法表。
- `pluginbridge-subcomponent-architecture`：guest SDK 与 host dispatcher 的方法集合从 protocol catalog 推导，而不是在根 facade 再抄一份。

## Impact

- 代码：`pkg/plugin/pluginbridge/protocol/hostservices`、`internal/service/plugin/internal/wasm/wasm_host_service_registry.go`、动态 guest 客户端生成或推导代码、相关治理测试。
- API：无 HTTP 变更。动态插件 wire 方法名保持不变。
- 数据：无。
- 测试：catalog 覆盖 WASM 注册与 guest 客户端；新增方法若只改 catalog 未推导，治理测试必须失败。
- i18n / 数据权限：无影响。
- 缓存：不改变 cache host service 语义，不退役其 dedicated codec。
- 依赖：建议在前两个子变更之后合入；不阻塞默认工作台换壳，但会降低后续加宿主方法的成本。
