## ADDED Requirements

### Requirement: 源码插件前端依赖必须声明在插件 frontend package 中

系统 SHALL 允许源码插件在 `apps/lina-plugins/<plugin-id>/frontend/package.json` 中声明仅该插件前端需要的 npm 依赖。该 package 文件属于插件前端资源治理边界。插件私有业务依赖不得默认写入宿主 `apps/lina-vben/apps/web-antd/package.json`。

#### Scenario: 插件声明私有前端依赖

- **WHEN** 源码插件前端只自身使用 `markdown-it`
- **THEN** 插件在自己的 `frontend/package.json` 中声明 `markdown-it`
- **AND** 宿主 `web-antd/package.json` 不需要包含该业务依赖

#### Scenario: 插件声明宿主单例 peer

- **WHEN** 源码插件前端使用 `vue`、`vue-router`、`pinia` 或 `ant-design-vue`
- **THEN** 插件 frontend package 将这些依赖表达为 `peerDependencies` 或等价 peer 消费约束
- **AND** 宿主构建只加载一份宿主提供的单例依赖

#### Scenario: 单插件业务依赖进入宿主时被治理阻断

- **WHEN** 变更向 `apps/lina-vben/apps/web-antd/package.json` 新增明显只服务某个插件的业务依赖
- **THEN** 审查或治理扫描必须要求说明该依赖是临时例外或改为插件 frontend package 声明
- **AND** 临时例外必须记录迁回计划
