## ADDED Requirements

### Requirement: 动态插件宿主方法发布面以 catalog 为唯一规格

系统 SHALL 把动态插件可调用的宿主 service/method、资源种类和 payload kind 维护在 hostservices catalog。分发器、授权校验和 guest SDK MUST 从该目录读取发布面，MUST NOT 在 WASM 注册表或客户端工厂再维护一套互相漂移的方法表。资源种类枚举 MUST 只有一个来源。

#### Scenario: 资源种类只有一份定义

- **WHEN** 清单校验与 catalog 需要判断`table`、`path`或`key`等资源形状
- **THEN** 双方使用同一`ResourceKind`定义
- **AND** 不得在能力注册表再抄一份独立枚举

#### Scenario: 未知方法仍被拒绝

- **WHEN** 插件调用 catalog 未发布的`service.method`
- **THEN** 宿主返回显式不支持或未授权错误
- **AND** 不得因为手工注册表更宽而执行该方法
