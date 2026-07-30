## 1. OpenSpec 与规范

- [x] 1.1 创建插件前端依赖自治 OpenSpec 变更文档和增量规范
- [x] 1.2 更新插件开发规则，明确源码插件 `frontend/package.json`、宿主单例 peer 和禁止单插件业务依赖进入 `web-antd`

## 2. Vite 解析

- [x] 2.1 拆分插件前端入口文件与模块文件判断
- [x] 2.2 扩展 `lina-plugin-source-deps`，支持 `frontend/**/*.{vue,ts,tsx,js,jsx}` 的 bare import 解析
- [x] 2.3 保持页面/插槽虚拟模块发现仅扫描入口 `.vue` 文件
- [x] 2.4 补充 Vite 配置静态测试覆盖 `frontend/utils/*.ts` 解析场景

## 3. 插件前端依赖准备

- [x] 3.1 在 `linactl` frontend 内部组件中扫描 `apps/lina-plugins/*/frontend/package.json`
- [x] 3.2 让 `env.setup`、`dev`、`build` 在宿主 Vite 启动或构建前准备插件 frontend 依赖
- [x] 3.3 补充 `linactl` 单元测试覆盖插件 frontend package 安装顺序和无 package 时跳过

## 4. Marketplace Markdown 迁移

- [x] 4.1 为 `linapro-plugin-marketplace/frontend` 增加 `package.json` 并声明 `markdown-it`
- [x] 4.2 将 Markdown 工具迁回插件 `frontend/utils/markdown.ts`
- [x] 4.3 修改详情页从插件工具导入 Markdown 渲染函数
- [x] 4.4 从宿主 `web-antd` 移除临时 `markdown-it`、`@types/markdown-it` 和 `#/utils/markdown`
- [x] 4.5 更新 marketplace 前端静态测试断言

## 5. 验证与审查

- [x] 5.1 运行相关前端静态测试、`linactl` Go 单测和 marketplace 单测
- [x] 5.2 运行 `openspec validate plugin-frontend-dependency-autonomy --strict`
- [x] 5.3 记录 i18n、缓存一致性、数据权限、开发工具跨平台、DI 和测试影响结论
- [x] 5.4 完成后调用 `lina-review` 审查

## Feedback

- [x] **FB-1**: 前端启动后一直停留在加载页
