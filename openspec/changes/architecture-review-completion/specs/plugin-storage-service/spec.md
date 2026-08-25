## ADDED Requirements

### Requirement: 对象存储宿主服务使用 JSON

动态插件对象存储的 put/get/list/stat/delete 以及分片上传方法 MUST 通过 JSON 信封传输元数据和小载荷约定。MUST NOT 再为这些方法维护 dedicated binary codec。

#### Scenario: storage.stat 往返

- **WHEN** guest 查询一个对象元数据
- **THEN** 请求与响应都是 JSON
