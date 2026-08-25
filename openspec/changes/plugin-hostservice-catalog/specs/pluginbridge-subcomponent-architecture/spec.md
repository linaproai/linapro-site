## ADDED Requirements

### Requirement: guest SDK 与 dispatcher 方法集合不得在根 facade 再抄一份

系统 SHALL 让`pluginbridge`根 facade 只保留稳定入口和动态专用额外能力。core-owned 方法集合 MUST 来自`protocol/hostservices` catalog 的推导结果。本要求 MUST NOT 改变既有 wire 字符串、JSON envelope 与 dedicated codec 的运行语义。

#### Scenario: 根 facade 不再维护平行方法名表

- **WHEN** 开发者新增一个 catalog 已发布的 core-owned 方法
- **THEN** guest 客户端与 WASM dispatcher 通过推导接入
- **AND** 不需要在`pluginbridge.go`手工追加方法清单

#### Scenario: 协议行为保持不变

- **WHEN** 使用推导后的客户端编码并调用既有 host service 方法
- **THEN** wire 方法名与 payload 编解码结果与变更前等价
- **AND** 既有协议测试必须继续通过
