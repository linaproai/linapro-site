## Why

`LinaPro`定位为面向可持续交付的`AI`原生全栈框架，默认管理工作台只是入口，用户必须能换成自己的管理壳。当前宿主却在替 Vben 编译路由、用进程全局补丁切换锁/会话/缓存后端、并用三本账维护动态插件宿主方法。这些偏差会在换壳和后续加能力时继续放大，需要一份可拆开合入的减肥纲领，而不是再往上加门面。

## What Changes

- 建立主框架减肥纲领：第一波只做三件战略，落地为三个可独立审查、合入和回滚的子变更。
- 约定合入顺序：`workbench-host-contract` → `startup-di-backends` → `plugin-hostservice-catalog`。
- **BREAKING**（由子变更 `workbench-host-contract` 执行）：宿主菜单、身份和插件页面契约改为通用资源，本仓库默认工作台一次迁完，不为仓外未知壳保留双轨。
- 明确第一波不做：插件总门面大拆、锁/缓存/任务包合并、`usermsg`收口、声明门面压缩、仪表盘假数据、验证码/二维码死路由、`RecordStore`功能扩展或删除、二进制编解码退役。这些由第二波`simplify-host-runtime`按可立即实施的子集落地，高风险包合并仍排除。
- 本变更本身不改生产代码，只冻结范围、决策和子变更边界。

## Capabilities

### New Capabilities

- `host-framework-simplification-program`：主框架减肥纲领的范围、子变更关系、合入顺序和明确排除项。

### Modified Capabilities

- `project-positioning-governance`：把可替换管理工作台写成框架硬约束，而不是文档定位声明。宿主契约的具体 delta 由`workbench-host-contract`修改`core-host-boundary-governance`。

## Impact

- OpenSpec：本变更与 `workbench-host-contract`、`startup-di-backends`、`plugin-hostservice-catalog` 四个活跃变更同时存在；实现只发生在后三个目录。
- 代码：本变更无生产代码。后续子变更分别影响 `apps/lina-core` 菜单/身份/插件页面契约、启动装配与锁会话缓存构造、动态插件 host service 目录。
- API：**BREAKING** 由子变更承担，见 `workbench-host-contract`。
- 默认工作台：`apps/lina-vben` 必须在第一个子变更内改为消费通用宿主契约。
- i18n：纲领无文案变更。子变更若改用户可见字段，各自评估。
- 缓存 / 数据权限：纲领无直接影响。`startup-di-backends` 会触及缓存协调构造；`workbench-host-contract` 会触及菜单过滤投影。
- 测试：纲领只做`openspec validate`。实现与 E2E 落在子变更。
