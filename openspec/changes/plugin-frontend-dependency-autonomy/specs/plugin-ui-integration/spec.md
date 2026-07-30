## ADDED Requirements

### Requirement: 源码插件前端依赖解析必须覆盖插件 frontend 普通模块

系统 SHALL 在宿主 Vite 编译源码插件前端时，将插件 `frontend/` 下的页面、插槽、组件、工具、API 和类型辅助模块视为同一插件前端源码边界。源码插件前端模块中的 bare import MUST 具有一致解析语义，不得仅支持 `frontend/pages/**/*.vue` 或 `frontend/slots/**/*.vue`。

#### Scenario: 插件工具模块导入插件私有依赖

- **WHEN** 源码插件 `apps/lina-plugins/<plugin-id>/frontend/utils/markdown.ts` 导入 `markdown-it`
- **AND** 该插件 `frontend/package.json` 声明了 `markdown-it`
- **THEN** 宿主 Vite 能够从该插件前端依赖上下文解析该依赖
- **AND** 不要求插件把工具函数移动到宿主 `#/utils`

#### Scenario: 插件普通模块导入宿主单例依赖

- **WHEN** 源码插件 `frontend/api`、`frontend/utils` 或 `frontend/components` 中导入 `vue`、`vue-router`、`pinia`、`ant-design-vue` 或 `@vben/*`
- **THEN** 宿主 Vite 使用宿主提供的单例依赖
- **AND** 不允许插件解析出第二份 `vue` 或 `ant-design-vue` 实例

### Requirement: 插件入口发现与普通模块依赖解析必须分离

系统 SHALL 区分源码插件前端入口文件和普通模块文件。入口文件只用于页面、插槽虚拟模块发现和入口新增删除时的失效；普通模块文件用于 Vite 源码依赖解析。系统不得因为普通 `frontend/utils/*.ts` 文件变化就将其当作新增页面或插槽入口。

#### Scenario: 普通工具文件变化不重建入口清单

- **WHEN** 源码插件修改 `frontend/utils/routes.ts`
- **THEN** 宿主 Vite 可以正常热更新受影响模块
- **AND** 页面和插槽虚拟模块清单不因该工具文件被误判为入口而重建

#### Scenario: 页面入口变化重建入口清单

- **WHEN** 源码插件新增或删除 `frontend/pages/detail/index.vue`
- **THEN** 宿主 Vite 重新生成插件页面虚拟模块
- **AND** 工作台插件页面注册表可以观察到入口变化

### Requirement: 单插件业务前端依赖不得污染宿主共享工具

系统 SHALL 要求只服务单个源码插件的业务前端依赖和工具函数保留在该插件 `frontend/` 目录。宿主 `#/utils`、`#/adapter` 和宿主 package 依赖只承载跨插件或工作台壳相关的稳定共享能力。

#### Scenario: Marketplace Markdown 渲染归属插件

- **WHEN** `linapro-plugin-marketplace` 需要使用 `markdown-it` 渲染版本文档
- **THEN** Markdown 渲染工具位于该插件 `frontend/utils`
- **AND** `markdown-it` 由该插件 `frontend/package.json` 声明
- **AND** 宿主 `web-antd` 不保留仅服务 marketplace 文档的 `#/utils/markdown` 后门
