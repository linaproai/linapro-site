## MODIFIED Requirements

### Requirement: 菜单图标全局唯一仍是宿主导航规则

系统 MUST NOT 把目录或菜单图标全局唯一当作宿主数据完整性规则。图标重复是默认工作台侧栏审美，换壳后没有通用语义。创建或更新菜单时，宿主 MAY 继续校验父级和名称冲突，MUST NOT 因图标与其它菜单相同而拒绝保存。

#### Scenario: 两个菜单使用同一图标

- **WHEN** 管理员创建或更新菜单并填写已有菜单使用的图标
- **THEN** 宿主允许保存
- **AND** 不得返回图标重复的业务错误

### Requirement: 壳布局偏好可以留在宿主公开配置

系统 MUST NOT 在`GET /config/public/frontend`中要求默认 Vben 工作台的布局词汇。公开配置 MUST 提供品牌、语言相关展示、注册/忘记密码等功能开关。宿主 MUST NOT 把`panel-left`、`sidebar-mixed-nav`、`sloganImage`或等价 slogan 字段当作换壳前置或通用契约字段。主题明暗与水印开关可以保留为通用展示设置。

#### Scenario: 自定义壳读取公开配置

- **WHEN** 任意管理工作台读取`GET /config/public/frontend`
- **THEN** 响应包含应用名、Logo 和功能开关
- **AND** 响应不得出现`panel-left`、`sidebar-mixed-nav`或`sloganImage`

## ADDED Requirements

### Requirement: 默认工作台编译壳布局偏好

默认管理工作台 SHALL 用自己的 preferences 或本地适配编译登录面板位置、导航布局和 slogan 插画。这些值 MUST NOT 作为宿主公开契约的必填字段。

#### Scenario: 默认工作台启动登录页

- **WHEN** 默认工作台打开登录页且公开配置不再返回面板布局
- **THEN** 工作台使用本地默认布局完成渲染
- **AND** 不得要求宿主返回`panelLayout`
