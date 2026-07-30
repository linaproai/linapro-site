## Why

源码插件前端源码已经由宿主 `apps/lina-vben/apps/web-antd` 的 Vite 构建管线直接扫描和编译，但插件前端依赖仍只能隐式依赖宿主 `package.json`。当前 `lina-plugin-source-deps` 只对 `frontend/pages` 和 `frontend/slots` 下的 `.vue` 文件做 bare import 回落，导致插件 `frontend/utils/*.ts` 等普通模块无法解析宿主已安装依赖。`linapro-plugin-marketplace` 的 Markdown 文档渲染因此把 `markdown-it` 和工具函数临时放入宿主，形成插件业务依赖污染宿主依赖图的问题。

## What Changes

- 扩展源码插件前端模块识别范围，使 `frontend/**/*.{vue,ts,tsx,js,jsx}` 的 bare import 行为一致。
- 将页面/插槽入口发现与普通前端模块依赖解析拆分，避免普通工具文件变更误触发虚拟模块重建。
- 为插件 `frontend/package.json` 建立约定，插件私有业务依赖在插件侧声明，宿主单例依赖以 peer 方式消费。
- 在 `linactl` 的开发、环境准备和构建链路中，在宿主 Vite 启动或构建前准备插件前端依赖。
- 将 marketplace Markdown 工具迁回插件 `frontend/utils`，由插件 `frontend/package.json` 声明 `markdown-it`，并清理宿主临时依赖和工具入口。
- 更新插件规范和回归测试，阻止后续单插件业务依赖继续进入 `web-antd`。

## Impact

- 影响 `apps/lina-vben/apps/web-antd/vite.config.mts` 的源码插件 Vite 解析逻辑。
- 影响 `hack/tools/linactl` 前端依赖准备和构建前置流程。
- 影响 `apps/lina-plugins/linapro-plugin-marketplace/frontend` 的 Markdown 工具归属。
- 影响 `apps/lina-vben` 的依赖清单和 lockfile。
- 影响 `.agents/rules/plugin.md` 的插件前端依赖规范。
- 不改变动态插件 `/x-assets`、`public_assets` 或 WASM 运行时前端资产托管语义。
