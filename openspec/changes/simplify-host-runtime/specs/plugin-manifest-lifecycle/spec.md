## ADDED Requirements

### Requirement: 启用或禁用插件不得重扫全部清单

系统 SHALL 在单个插件启用或禁用时只同步该插件的目标清单。MUST NOT 扫描已发现的全部插件清单并逐个`SyncManifest`。

#### Scenario: 启用一个已安装插件

- **WHEN** 管理员启用某一个已安装插件
- **THEN** 宿主只读取并同步该插件 ID 的目标清单
- **AND** 不得为无关插件打开清单同步写路径
- **AND** 启用快照只更新该插件条目或基于已持久化治理状态刷新
