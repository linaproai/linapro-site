## ADDED Requirements

### Requirement: Network文档说明动态出站网络能力
文档`8200-network.md` SHALL 说明动态插件通过`hostServices.network`和`Network().Request`发起受治理出站`HTTP`请求。

#### Scenario: network授权边界清晰
- **WHEN** 动态插件开发者声明`service: network`
- **THEN** 文档要求声明`resources[].ref`，并说明宿主可限制方法、请求头、超时和响应大小

### Requirement: Network文档说明预留服务状态
文档 SHALL 说明`secret`、`event`和`queue`当前为预留治理条目，不是已发布的`guest`可调用服务。

#### Scenario: 预留服务不被误用
- **WHEN** 插件开发者查找外部能力
- **THEN** 文档不把预留条目描述为可执行动态插件调用
