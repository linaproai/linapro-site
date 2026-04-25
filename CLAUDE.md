# 仓库用途

本仓库承载`LinaPro`官方网站（`apps/lina-site/`，基于`Docusaurus 3.10`）以及配套的 `OpenSpec` 治理文件。`LinaPro`本体——也就是官网所描述的 AI 驱动全栈框架 —— 位于单独的只读参考仓库`linapro`；编写站点文案时，应以该路径中的事实内容为依据。

官网发布地址为`https://linapro.ai/`，源仓库地址为`https://github.com/linaproai/linapro`（目前均未公开）。

`AGENTS.md`是指向本文件的符号链接 —— `Claude Code`、`Codex`以及其他`AI Coding Agent`会读取`AGENTS.md`的智能体应共享同一套说明。

# 架构说明

## 站点结构（`apps/lina-site/`）

- `docusaurus.config.ts`——站点配置的唯一事实源。标题、标语、描述、关键词等会通过`geti18nTitle`、`geti18nTagline`等函数按 `locale` 计算，这些函数依赖`process.env.DOCUSAURUS_CURRENT_LOCALE`。`docs.routeBasePath: '/'`表示**文档直接挂在站点根路径**，不存在单独的`/docs`前缀；根路由首页由`src/pages/index.tsx`覆盖。
- `i18n/`——国际化资源目录。当前`defaultLocale: 'en'`，可用 `locale` 为`['en', 'zh-Hans']`。各 `locale` 资源分别位于`i18n/<locale>/docusaurus-plugin-content-docs/`、`docusaurus-plugin-content-blog/`、`docusaurus-theme-classic/`以及`code.json`。**新增规范：所有站点文档、博客、页面文案的编写以`i18n/zh-Hans/`中的中文内容为主稿；中文内容确认后，再同步到默认语言目录中的英文主内容。即便当前`defaultLocale`仍为`en`，也必须遵守“中文先行、英文同步”的流程。**
- `docs/`——默认 `locale` 的 **`current`（最新）版本**，当前标签为`v0.1.x(Latest)`。旧版本位于`versioned_docs/`、`versioned_sidebars/`和`versions.json`，由`node version.js`生成。`editUrl`目前指向 GitHub 上的`linaproai/linapro`——后续站点源码预期会回到 `LinaPro` 主 `monorepo` 中。
- `sidebars.ts`——定义两个自动生成的侧边栏：`mainSidebar`（覆盖`docs/`下的全部内容）和`quickSidebar`（覆盖`docs/quick/`）。导航栏中的`Get Started`链接指向`/quickstart`，对应`docs/quick/quickstart.md`。
- `src/pages/index.tsx`——自定义首页；`src/components/HomepageFeatures/`和`src/theme/`中包含对经典主题的 `swizzle` 内容。仓库已启用 `Mermaid`（`themes: ['@docusaurus/theme-mermaid']`、`markdown.mermaid: true`），文档中的架构图和流程图应优先使用 `Mermaid`。
- 当前启用的插件包括：`docusaurus-plugin-image-zoom`（缩放选择器为`.markdown :not(em) > img`）以及`@docusaurus/plugin-ideal-image`（开发环境下禁用）。

## OpenSpec 治理（`openspec/`）

本仓库使用 `OpenSpec` 管理需求提案、设计和落地过程，目录结构如下：

- `openspec/changes/<change-id>/`——进行中的提案目录，包含`proposal.md`、`design.md`、`tasks.md`和`specs/`。
- `openspec/changes/archive/`——已完成并归档的变更。
- `openspec/specs/`——变更归档后沉淀的基线规范。
- `openspec/config.yaml`——项目级语言与生成策略配置。

涉及 `OpenSpec` 工作流时，优先使用`opsx:propose`、`opsx:explore`、`opsx:apply`、`opsx:archive`技能（或其`openspec-*`别名），不要手工拼装整套变更文件。

# 文档规范

- 编写官网文档时，需要严格遵守该文件的规范`linapro-site/.agents/instructions/markdown-format.instructions.md`。
- 编写官网文档时，优先以中文为主稿，确认中文内容后再同步到英文版本。即便当前`defaultLocale`仍为`en`，也必须遵守“中文先行、英文同步”的流程。
- 中文文档内容符合中文习惯的表达方式，同时保持内容的准确性和专业性，不能生硬表达含义。
- 英文版本需要保持为地道的英文表达，不要逐字翻译中文内容。

