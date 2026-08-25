## ADDED Requirements

### Requirement: 锁宿主服务使用 JSON

动态插件锁获取、续期和释放 MUST 通过 JSON 信封调用宿主 lock 服务。

#### Scenario: lock.acquire 往返

- **WHEN** guest 申请一把命名锁
- **THEN** 请求与响应都是 JSON
- **AND** 票据字段仍由 hostlock 签发
