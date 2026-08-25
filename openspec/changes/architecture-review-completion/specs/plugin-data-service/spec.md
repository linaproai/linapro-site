## ADDED Requirements

### Requirement: 数据宿主服务使用 JSON

动态插件`data`方法 MUST 通过 JSON 信封传输查询与变更请求。MUST NOT 再为 list/get/create/update/delete/transaction/batch_get 维护 dedicated binary codec。

#### Scenario: data.list 往返

- **WHEN** guest 列出授权表中的记录
- **THEN** 请求与响应都是 JSON

### Requirement: RecordStore 是实验能力

官方插件文档 SHALL 把`RecordStore`标明为实验能力，不得把它写成动态插件主路径或必须使用的数据访问方式。实现可以保留，查询计划器不得借此继续加功能。

#### Scenario: 阅读当前官网数据能力文档

- **WHEN** 读者打开当前版 RecordStore 文档
- **THEN** 文档标明实验能力
- **AND** 不得把它与 Runtime/Network 并列写成必选主路径
