## ADDED Requirements

### Requirement: 缓存锁存储数据领域新老方法都是 JSON

系统 SHALL 对 cache、lock、storage、data 宿主方法使用 JSON payload。新增这些领域的方法 MUST NOT 引入 dedicated codec。

#### Scenario: catalog 声明 lock.acquire

- **WHEN** 治理测试检查 lock.acquire 的 payload kind
- **THEN** 它是 JSON
- **AND** 不得出现在 dedicated allowlist
