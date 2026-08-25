## ADDED Requirements

### Requirement: 宿主必须向管理工作台提供通用导航资源

系统 SHALL 为当前登录用户提供通用导航资源列表或树。每个可导航节点 MUST 包含稳定编号、父级、标题、路径、权限标识、图标、可见性、排序、打开方式和资源地址。打开方式 MUST 使用宿主通用枚举，至少覆盖页内打开、内嵌资源、iframe 和外链。宿主 MUST NOT 在该契约中返回默认 Vben 工作台的组件路径、布局组件名或路由元字段，包括`#/views`前缀、`hideInMenu`、`keepAlive`、`IFrameView`和`BasicLayout`。

#### Scenario: 当前用户拉取可访问导航

- **WHEN** 已登录用户请求当前可访问导航资源
- **THEN** 系统只返回该用户授权范围内的通用导航节点
- **AND** 响应中不得出现`#/views/`或`IFrameView`

#### Scenario: 打开方式不靠旧工作台字段推断

- **WHEN** 宿主从已存菜单投影打开方式
- **THEN** 只使用路径、外链标志和托管脚本扩展名
- **AND** 不得再读取`system/plugin/dynamic-page`或`pluginAccessMode`来决定打开方式

#### Scenario: 默认工作台编译路由

- **WHEN** 默认管理工作台`lina-vben`收到通用导航资源
- **THEN** 工作台自行编译为壳所需的路由记录
- **AND** 动态插件页的壳组件路径只存在于工作台，不存在于宿主公开包

#### Scenario: 源码插件页按路径编译到工作台壳

- **WHEN** 默认管理工作台收到`openMode=page`且路径对应已注册源码插件页面的导航资源
- **THEN** 工作台按页面路径编译到自己的动态页壳
- **AND** 不得要求宿主`resource`仍为`system/plugin/dynamic-page`
- **AND** 不得把以`#`开头的字符串当作已经编译好的 Vben 组件路径原样采用

### Requirement: 插件页面发现可以走约定目录，但写入宿主的必须是通用页面资源

系统 SHALL 允许按`frontend/pages`等约定目录发现源码插件页面，并按动态插件前端资产约定发现动态页面入口。发现结果写入菜单或页面存储时 MUST 只保存通用页面资源。`pluginhost` MUST NOT 导出默认工作台 Vue 组件路径常量。

#### Scenario: 源码插件页面被发现

- **WHEN** 已启用源码插件在`frontend/pages`下提供页面文件
- **THEN** 宿主同步出带页面编号、路径、权限、打开方式和资源地址的通用页面资源
- **AND** 不得把`#/views/<plugin>/...`写入宿主导航契约

#### Scenario: 动态插件页面被发现

- **WHEN** 已启用动态插件提供可托管的前端入口资产
- **THEN** 宿主保存打开方式与资源地址
- **AND** 不得要求插件或宿主公开包引用`system/plugin/dynamic-page`

### Requirement: 菜单图标全局唯一仍是宿主导航规则

系统 SHALL 在创建或更新目录和菜单类型记录时校验图标全局唯一。该规则属于宿主导航模型，不因换壳而取消。

#### Scenario: 两个菜单使用同一图标

- **WHEN** 管理员创建或更新菜单并填写已有菜单使用的图标
- **THEN** 宿主拒绝保存
- **AND** 返回图标重复的业务错误

### Requirement: 壳布局偏好可以留在宿主公开配置

系统 SHALL 允许公开前端配置继续提供登录面板位置、导航布局名、slogan、主题和水印等可选壳偏好。自定义管理工作台 MAY 忽略这些字段。宿主 MUST NOT 把这些偏好当成换壳的前置条件。

#### Scenario: 自定义壳忽略导航布局名

- **WHEN** 非默认工作台读取`GET /config/public/frontend`
- **THEN** 它可以忽略`ui.layout`等壳偏好
- **AND** 仍必须能使用通用导航资源、品牌和能力开关完成启动
