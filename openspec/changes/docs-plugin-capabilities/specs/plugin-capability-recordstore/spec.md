## ADDED Requirements

### Requirement: RecordStore文档说明动态data服务
文档`7600-data-recordstore.md` SHALL 说明动态插件通过`hostServices.data`和`guest.RecordStore()`访问插件自有表，覆盖读取、变更和事务方法。

#### Scenario: data服务覆盖完整
- **WHEN** 动态插件开发者查看数据记录能力文档
- **THEN** 文档列出`list`、`get`、`create`、`update`、`delete`和`transaction`方法及`host:data:read`、`host:data:mutate`能力含义

### Requirement: RecordStore文档说明源码插件替代路径
文档 SHALL 说明源码插件没有普通`Services.Data()`入口，应通过自身领域服务访问插件自有表，并在需要租户隔离时使用`TenantFilter()`。

#### Scenario: 源码和动态数据路径可区分
- **WHEN** 插件开发者比较源码插件和动态插件数据访问
- **THEN** 文档说明源码插件用领域服务加`TenantFilter()`，动态插件用`data`服务和`tables`
