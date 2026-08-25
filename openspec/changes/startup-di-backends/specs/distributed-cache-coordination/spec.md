## ADDED Requirements

### Requirement: cachecoord 必须由启动编排注入而不是进程 Default

系统 SHALL 由启动编排构造`cachecoord.Service`并注入拓扑与 coordination 后端。生产路径 MUST NOT 使用`cachecoord.Default`或`DefaultWithCoordination`作为共享实例工厂。集群修订事实源仍 MUST 遵循既有 Redis revision 规则，MUST NOT 把未使用的`sys_cache_revision`表重新写成现行实现。

#### Scenario: 启动期只存在一份 cachecoord

- **WHEN** HTTP 运行时完成启动装配
- **THEN** 配置、权限、插件等调用方持有同一`cachecoord`实例
- **AND** 该实例的拓扑和 coordination 后端来自构造参数

#### Scenario: 注释不得宣称 SQL 修订表仍在生产路径

- **WHEN** 开发者阅读 cachecoord、config 或权限修订相关注释
- **THEN** 注释描述的运行路径必须与实际后端一致
- **AND** 不得把`sys_cache_revision`写成集群现行实现

#### Scenario: 源码 SQL 不创建修订表

- **WHEN** 从当前仓库 SQL 初始化空库
- **THEN** 初始化结果中不存在`sys_cache_revision`
- **AND** 不得再提供仅为已部署库删除该表的迁移文件
