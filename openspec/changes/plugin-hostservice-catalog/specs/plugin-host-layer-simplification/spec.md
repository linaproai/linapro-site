## ADDED Requirements

### Requirement: WASM 分发注册必须从 hostservices catalog 推导

系统 SHALL 以`protocol/hostservices`目录为动态插件宿主方法的唯一规格来源。WASM 分发注册表 MUST 从 catalog 的已发布 dispatcher 方法推导，MUST NOT 再维护一份平行的手工方法清单。catalog 中已发布但缺少 handler、或 handler 已注册但不在 catalog 中的情况 MUST 使构建或治理测试失败。

#### Scenario: 新增 catalog 方法未绑定 handler

- **WHEN** 开发者在 catalog 增加已发布 dispatcher 方法但未提供 WASM handler
- **THEN** 构建或治理测试失败
- **AND** 失败信息包含`service.method`

#### Scenario: 手工注册孤儿方法

- **WHEN** WASM 注册表包含 catalog 未发布的方法
- **THEN** 构建或治理测试失败

### Requirement: 动态 guest 核心方法表必须从同一 catalog 推导

系统 SHALL 让动态插件 guest 客户端的 core-owned 方法集合与 catalog 发布面一致。`pluginbridge.Services` MUST NOT 再手写一份与`capability.Services`平行的核心方法清单。`Runtime`、`Network`、`RecordStore`等仅动态侧入口 MUST 在 catalog 中标记为动态专用，不得并入源码插件`capability.Services`。

#### Scenario: 源码目录与动态核心方法对齐

- **WHEN** catalog 发布一个 core-owned JSON 方法
- **THEN** 动态 guest 客户端可调用该方法
- **AND** 不需要在 pluginbridge 根 facade 再抄方法名

#### Scenario: 动态专用入口保持额外

- **WHEN** 动态插件调用`RecordStore`或`Network`
- **THEN** 这些入口仍只存在于动态侧
- **AND** 源码插件`capability.Services`不增加对应方法
