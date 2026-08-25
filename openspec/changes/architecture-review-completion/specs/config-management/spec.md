## ADDED Requirements

### Requirement: 运行期配置读取面按领域拆分

系统 SHALL 把`config`运行期读取面拆成领域接口，至少覆盖集群、令牌、会话、登录策略、i18n、任务、插件、上传、工作台路径和公开品牌。完整`Service`可以嵌入这些接口，供非控制器调用方按领域依赖。HTTP 控制器 MUST 依赖完整`config.Service`，不得把领域读取面作为控制器字段或构造函数参数。`config`与`sysconfig`继续分开。公开品牌读取 MUST NOT 包含 Vben 布局词。

#### Scenario: 上传模块读取上传配置

- **WHEN** 上传服务被构造
- **THEN** 它可以只依赖上传读取面
- **AND** 不需要登录页品牌方法才能编译

#### Scenario: 公开配置控制器构造

- **WHEN** 公开前端配置控制器被构造
- **THEN** 它依赖`config.Service`
- **AND** 不得把`PublicFrontendReader`作为控制器字段类型

#### Scenario: 公开品牌读取

- **WHEN** 未登录客户端读取公开前端配置
- **THEN** 读取面返回品牌和功能开关
- **AND** 不含`panel-left`、`sidebar-mixed-nav`、slogan 插画字段
