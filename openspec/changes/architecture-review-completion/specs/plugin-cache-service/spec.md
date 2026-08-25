## ADDED Requirements

### Requirement: 缓存宿主服务使用 JSON

动态插件缓存读写 MUST 通过 JSON 信封调用宿主 cache 服务。guest 客户端 MUST NOT 再调用 dedicated cache marshal 函数。

#### Scenario: cache.set 往返

- **WHEN** guest 写入一个带 TTL 的缓存值
- **THEN** 请求与响应都是 JSON
- **AND** WASM 分发用同一 JSON 结构解码
