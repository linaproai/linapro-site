<div align="center">
<img src="https://linapro.ai/img/linapro-logo.png" width="300" alt="linapro logo"/>

[English](README.md) | 简体中文

</div>

# 项目介绍

本仓库包含 [LinaPro.AI](https://linapro.ai/) 官方网站的源码，基于 [Docusaurus 3.10](https://docusaurus.io/) 构建。`LinaPro` 框架源码位于独立仓库：[linaproai/linapro](https://github.com/linaproai/linapro)。

# 快速链接

| 资源 | 地址 |
|------|------|
| **开源仓库** | https://github.com/linaproai/linapro-site |
| **官方网站** | https://linapro.ai/ |
| **框架仓库** | https://github.com/linaproai/linapro |

# 技术栈

| 类别 | 技术 | 说明 |
|------|------|------|
| 框架 | `Docusaurus` | `v3.10.0`， classic  预设 |
| 语言 | `TypeScript` | `v5.6.3` |
| 运行时 | `Node.js` | `>= 20` |
| 包管理 | `yarn` | 亦支持 `pnpm` / `npm` |
| 搜索 | `Algolia DocSearch` | |
| 图表 | `Mermaid` | `Markdown` 内嵌 |
| 评论 | `Giscus` | 基于 `GitHub Discussions` |

# 项目结构

```
linapro-site/
├── apps/lina-site/              # Docusaurus 站点
│   ├── docs/                    # 中文文档主稿（内容基准）
│   ├── i18n/                    # 国际化资源（en、zh-Hans）
│   ├── src/                     # 自定义页面与组件
│   ├── static/                  # 静态资源
│   ├── docusaurus.config.ts     # 站点配置
│   ├── sidebars.ts              # 侧边栏定义
│   └── siteI18n.ts              # 多语言 SEO 元信息
├── openspec/                    # OpenSpec 治理文件
├── Makefile                     # 顶层开发/构建命令
└── AGENTS.md                    # AI Coding Agent 指令
```

# 快速开始

**前置条件**：`Node.js >= 20`，已安装 `yarn`（或 `pnpm` / `npm`）。

```bash
# 克隆仓库
git clone https://github.com/linaproai/linapro-site.git
cd linapro-site

# 安装依赖
cd apps/lina-site && yarn install

# 启动中文站点（默认，localhost:3000/zh/）
make dev

# 启动英文站点（localhost:3000/）
make dev locale=en
```

# 常用命令

| 命令 | 说明 |
|------|------|
| `make dev` | 启动本地开发服务（中文，默认） |
| `make dev locale=en` | 启动英文站点 |
| `make build` | 构建静态站点至 `apps/lina-site/build/` |
| `make check` | 检查中文文档的 i18n 翻译完整性 |
| `make webp` | 将内容图片转换为 WebP 格式 |

# 国际化

站点采用**中文先行、英文同步**的工作流：

- `apps/lina-site/docs/` 维护中文文档主稿（内容基准）
- 英文翻译位于 `apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/`
- 中文站点（`/zh/`）直接复用中文主稿
- 每个语言版本的 `SEO` 元信息在 `siteI18n.ts` 中独立定义

# 贡献指南

文档贡献需遵循中文先行的工作流：先更新中文主稿，再同步至英文及其他语言版本。

# 许可证

[Apache License 2.0](LICENSE)
