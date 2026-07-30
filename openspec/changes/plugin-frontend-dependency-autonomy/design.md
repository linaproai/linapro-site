## 背景

源码插件前端有两条不同职责：

1. 页面和插槽入口发现：宿主扫描 `frontend/pages` 和 `frontend/slots` 下的 `.vue` 文件，生成虚拟模块并注册到工作台路由或插槽。
2. 前端源码模块编译：页面可以 import 同插件 `frontend/api`、`frontend/utils`、`frontend/types`、`frontend/components` 等普通模块，这些模块也需要参与 Vite 依赖解析。

当前实现使用同一个 `isPluginFrontendSourceFile` 同时判断入口发现、热更新和依赖回落，且只允许 `pages|slots` 下的 `.vue` 文件。这使 `frontend/utils/*.ts` 的 bare import 行为与页面文件不一致。

## 设计决策

### 1. 拆分入口文件与模块文件判断

新增两个概念：

| 概念 | 范围 | 用途 |
|------|------|------|
| 插件前端入口文件 | `frontend/pages/**/*.vue`、`frontend/slots/**/*.vue` | 虚拟模块发现、入口新增删除时失效 |
| 插件前端模块文件 | `frontend/**/*.{vue,ts,tsx,js,jsx}` | bare import 解析回落、普通源码依赖判断 |

这样既能让 `frontend/utils/*.ts` 解析 bare import，又不会因为每个普通模块变化都重建页面/插槽虚拟模块。

### 2. 依赖分层

| 层级 | 示例 | 归属 |
|------|------|------|
| 宿主单例 peer | `vue`、`vue-router`、`pinia`、`ant-design-vue`、`@vben/*` | 宿主提供，插件以 peer 消费 |
| 宿主共享能力 | `#/adapter/*`、`#/api/*`、`#/utils/download`、`#/utils/time` | 宿主公共能力 |
| 插件私有业务依赖 | `markdown-it`、专用图表库、业务 SDK | 插件 `frontend/package.json` 声明 |
| 可选共享包 | 多插件共用的 Markdown 或图表封装 | 独立 workspace package，不能塞进单个宿主 app |

`markdown-it` 仅服务 marketplace 文档渲染时归入插件私有业务依赖。

### 3. 插件 frontend package 约定

源码插件可以维护 `apps/lina-plugins/<plugin-id>/frontend/package.json`：

```json
{
  "name": "@lina-plugin/<plugin-id>-frontend",
  "private": true,
  "type": "module",
  "dependencies": {
    "markdown-it": "14.1.1"
  },
  "devDependencies": {
    "@types/markdown-it": "14.1.2"
  },
  "peerDependencies": {
    "ant-design-vue": "^4.0.0",
    "vue": "^3.5.0",
    "vue-router": "^4.0.0"
  }
}
```

宿主 Vite 解析时优先使用宿主强制单例 alias，插件私有依赖通过插件 frontend package 所在目录解析。宿主已安装依赖的回落仅作为共享依赖和 P0 兼容行为，不作为单插件业务库长期落点。

### 4. 依赖准备链路

`linactl` 是默认跨平台开发工具入口，因此插件前端依赖准备必须集中到 `hack/tools/linactl/internal/frontend` 或等价内部组件中：

- `make env.setup` 和 `linactl env.setup` 安装宿主前端依赖后，同时准备已安装源码插件的 frontend package 依赖。
- `make dev` 和 `linactl dev` 在启动 Vite 前检查并准备插件前端依赖。
- `make build` 和 `linactl build` 在宿主前端构建前准备插件前端依赖，不能依赖插件自定义 `build.commands`，因为源码插件页面会在宿主前端构建阶段被编译。

实现应使用 Go 标准库扫描 `apps/lina-plugins/*/frontend/package.json`，并通过 `pnpm --dir <plugin-frontend-dir> install` 这类跨平台子进程调用完成安装。长期可再演进为 pnpm workspace 纳入插件 frontend，但本轮先保持最小可落地方案。

### 5. 不改变动态插件运行时资产模型

动态插件通过 WASM artifact 携带前端资产，再由 `public_assets` 显式声明公开路径。源码插件前端依赖自治只处理宿主构建时的源码编译和 npm 解析，不改变 `/x-assets`、动态插件前端缓存或公开资产授权边界。

## 复杂度判断

本次新增抽象集中在两个已确认变化点：

- Vite 中入口文件与普通模块文件的职责不同，拆分判断能降低调用方理解成本。
- `linactl` 前端依赖准备需要同时服务 env、dev、build 三个入口，放入内部 frontend 组件可避免命令文件重复扫描和安装逻辑。

未引入联邦前端、远程运行时加载、插件级独立 Vite 应用或 chunk 懒加载改造。这些属于后续包体和分发模式优化，不是当前根因闭环所需。

## 影响分析

- `i18n`：不新增运行时 UI 文案或翻译键；迁移 Markdown 工具不改变用户可见文案。插件 `linapro-plugin-marketplace` 已启用 i18n，但本次不修改语言资源。
- 缓存一致性：不修改后端缓存、前端运行时缓存或动态插件 frontend bundle 缓存。
- 数据权限：不修改后端 API、数据读写或可见性边界。
- 开发工具跨平台：有影响；依赖准备逻辑必须通过 Go 工具和跨平台 `pnpm` 子进程实现，不新增 shell 脚本。
- DI：无新增后端运行期依赖或宿主服务构造函数变更。
- 测试策略：补充 Vite 解析静态单测、`linactl` 前端依赖准备单测、marketplace 前端静态测试，并运行 OpenSpec 严格校验。
