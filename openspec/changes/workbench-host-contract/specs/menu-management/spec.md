## MODIFIED Requirements

### Requirement:创建菜单

系统 SHALL 支持创建新菜单，包括目录、菜单和按钮三种类型。菜单类型的页面资源地址 MUST 视为不透明资源标识，宿主 MUST NOT 按默认 Vben 工作台规则补全`#/views`前缀、`BasicLayout`或`IFrameView`。

#### Scenario:创建目录类型菜单
- **当** 用户填写目录信息（名称、图标、排序、是否显示、状态）并提交时
- **则** 系统创建目录类型菜单，type 为 "D"
- **且** 系统自动设置 created_at 和 updated_at
- **且** 目录类型的 path 字段用于路由分组

#### Scenario:创建菜单类型菜单
- **当** 用户填写菜单信息（名称、路由地址、页面资源地址、权限标识、图标、排序、是否显示、是否缓存、状态）并提交时
- **则** 系统创建菜单类型菜单，type 为 "M"
- **且** 菜单类型必须有 path 和页面资源地址
- **且** 宿主不得把页面资源地址改写为`#/views/...`或 Vben 布局组件名

#### Scenario:创建按钮类型菜单
- **当** 用户填写按钮信息（名称、权限标识、上级菜单）并提交时
- **则** 系统创建按钮类型菜单，type 为 "B"
- **且** 按钮类型没有 path、页面资源地址、icon 等字段

#### Scenario:创建外链菜单
- **当** 用户创建菜单并将打开方式设置为外链时
- **则** 系统将 path 视为外链地址
- **且** 前端点击该菜单时新窗口打开外链

#### Scenario:菜单名称重复
- **当** 用户创建菜单时使用已存在的菜单名称
- **则** 系统返回错误消息"菜单名称已存在"

## ADDED Requirements

### Requirement: 当前用户导航必须返回通用资源而不是 Vben 路由

系统 SHALL 为当前登录用户提供通用导航资源。该接口 MUST NOT 按 Vben 路由记录投影，MUST NOT 在控制器层把组件路径补成`#/views/...`，也 MUST NOT 为外链或内嵌页指定`BasicLayout`或`IFrameView`。

#### Scenario: 已登录用户获取导航资源

- **WHEN** 当前用户请求可访问导航
- **THEN** 系统返回通用导航节点
- **AND** 节点不包含`hideInMenu`、`keepAlive`或 Vben 组件名

#### Scenario: 外链与内嵌页使用打开方式

- **WHEN** 菜单以 iframe 或外链方式打开
- **THEN** 宿主使用通用打开方式与目标地址表达
- **AND** 不得返回名为`IFrameView`的组件字段

### Requirement: 插件过滤菜单必须使用稳定投影

系统 SHALL 在插件启用态过滤菜单时使用稳定菜单投影，至少包含过滤所需的编号、菜单键、插件归属和可见性字段。跨模块调用 MUST NOT 传递`SysMenu`实体切片。

#### Scenario: 禁用插件的菜单被过滤

- **WHEN** 菜单装配需要隐藏已禁用插件的菜单
- **THEN** 插件集成层接收菜单投影并返回过滤后的投影
- **AND** 不得以`[]*entity.SysMenu`作为跨模块契约
