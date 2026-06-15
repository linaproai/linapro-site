## ADDED Requirements

### Requirement: Config文档说明配置能力关系
文档`7000-config.md` SHALL 作为配置管理关系介绍页，说明`Plugins().Config()`、`HostConfig()`、`Admin().Config()`、动态`config.get`和`hostconfig.get`的领域边界。

#### Scenario: 配置入口可区分
- **WHEN** 插件开发者阅读基本介绍
- **THEN** 文档通过表格区分插件静态配置、宿主配置读取和运行时配置管理的入口、使用者和动态服务

### Requirement: Config文档说明插件配置读取优先级
文档 SHALL 说明插件静态配置的读取优先级：生产覆盖 > 开发默认 > 动态插件产物默认。

#### Scenario: 优先级清晰
- **WHEN** 插件开发者阅读插件静态配置章节
- **THEN** 理解三层配置覆盖关系和`manifest/config/config.example.yaml`不参与运行时默认读取

### Requirement: Config文档说明宿主配置读取边界
文档 SHALL 说明源码插件通过`HostConfig()`读取宿主配置值，动态插件通过`hostconfig.get`和`keys`声明授权范围。

#### Scenario: 宿主配置边界清晰
- **WHEN** 插件开发者阅读宿主配置读取章节
- **THEN** 理解源码插件不应随意依赖宿主全局配置，动态插件必须声明可读取的配置键

### Requirement: Config文档说明运行时配置管理命令
文档 SHALL 说明运行时配置写入属于`Admin().Config().SetConfigJSON`管理命令，不属于普通插件读取能力。

#### Scenario: 写入入口清晰
- **WHEN** 插件开发者阅读运行时配置管理章节
- **THEN** 理解写入配置需要可信源码插件、`CapabilityContext`和宿主领域治理
