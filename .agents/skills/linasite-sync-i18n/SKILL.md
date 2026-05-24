---
name: linasite-sync-i18n
description: >-
  检测 apps/lina-site/docs/ 中文主稿与 apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/
  英文翻译之间的缺漏和内容差异，并进行补全修复。当 docs/ 中新增或更新了文档、或需要全量同步审查时使用。
  执行时必须优先检查 apps/lina-site/docs/ 的 git status；如果存在已修改、新增、重命名或删除的 Markdown 文档，
  必须先把这些变更同步到英文 i18n 中，再考虑全量审查。
  翻译和文档标题都必须采用地道、专业、自然的英文表达；标题应尽可能与 slug/路径表达的英文路由名称一致，
  绝不逐字翻译中文，确保英文版本在英语环境下自然流畅。
---

# linasite-sync-i18n：文档 i18n 同步与翻译

对中文主稿进行全面审查，确保每篇文档都有对应的地道英文翻译，并与中文内容保持同步。本技能输出自然、专业的英文文档和标题——而非机械翻译。

**核心原则：**
1. **路径镜像** — `docs/<path>.md` 对应 `i18n/en/docusaurus-plugin-content-docs/current/<path>.md`
2. **slug 是唯一标识** — 每篇文档的 `slug` 字段是规范 URL，翻译时绝不修改
3. **标题贴近路由** — frontmatter `title` 应尽可能与 `slug` 或镜像路径所表达的英文路由名称一致，例如 `/docs/architecture/core-host` 对应 `Core Host`
4. **地道英文优先** — 先理解内容，再以英文习惯表达；正文、frontmatter 标题和各级标题都绝不逐字翻译
5. **中文主稿为准** — `docs/` 是唯一事实来源，英文版跟随中文版更新
6. **标题质量同等重要** — 文档标题是导航、SEO 和读者第一印象的一部分，必须符合英语技术文档的命名习惯
7. **当前变更优先** — 每次执行先检查 `apps/lina-site/docs/` 的 git status；凡是当前工作区已修改、新增、重命名或删除的中文 Markdown 文档，都属于强制同步范围

---

## 目录路径

| 角色 | 路径 |
|------|------|
| 中文主稿 | `apps/lina-site/docs/` |
| 英文 i18n | `apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/` |

---

## 工作流程

### 0. 优先检查当前 docs 变更

在做任何全量扫描、差异报告或翻译前，先检查当前中文主稿目录的 git status：

```bash
git status --short -- apps/lina-site/docs
```

把输出中影响 `apps/lina-site/docs/**/*.md` 的条目整理为**当前变更同步清单**：

- `M` / `MM` / `AM`：读取当前工作区版本，按镜像路径更新对应英文 i18n 文件
- `A` / `??`：新增中文文档，创建对应英文 i18n 文件
- `R`：确认新路径的镜像文件存在并同步当前内容；必要时说明旧英文路径可能需要清理，但不要自动删除用户未确认的文件
- `D`：中文主稿已删除，报告对应英文 i18n 文件可能需要删除；除非用户明确要求，否则不要自动删除英文文件

处理规则：

- 如果当前变更同步清单非空，必须先同步这些文档；不要因为还没有完成全量审查而跳过它们。
- 当前变更同步清单中的修改和新增 Markdown 文档不属于可选差异，必须被翻译或重新翻译到英文 i18n 镜像路径。
- 对当前变更同步清单中的文件，允许在开始前简要报告将处理的文件，但不要把“是否同步这些变更”作为可拒绝选项；只在遇到路径不匹配、删除文件、或语义不清的情况时请求用户决策。
- 完成当前变更同步后，再根据用户请求决定是否继续进行全量缺漏和差异审查。

### 1. 扫描中文主稿

列出中文主稿目录中的所有 Markdown 文件：

```bash
find apps/lina-site/docs -name "*.md" | sort
```

跳过非内容文件（`.DS_Store`、`_category_.json` 等）。

对每个 `.md` 文件记录：
- **相对路径**（相对于 `apps/lina-site/docs/`，例如 `docs/1000-concepts/1000-ai-native.md`）
- **期望的 i18n 路径**：`apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/<relative-path>`
- **slug**：从 frontmatter 的 `slug:` 字段读取
- **路由名称候选**：优先从 `slug` 的最后一段推导；没有 slug 时，从文件名去掉数字前缀后推导
- **中文标题**：frontmatter `title` 以及正文中的 `##`/`###` 标题

---

### 2. 扫描已有英文翻译

列出 i18n/en 目录中已存在的所有 `.md` 文件：

```bash
find apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current -name "*.md" | sort
```

对每个文件读取其 frontmatter 中的 **slug** 和 **title**，构建索引：`slug → i18n/en 文件路径`，并记录正文中的 `##`/`###` 标题用于自然度和路由一致性检查。

---

### 3. 对每篇中文文档进行分类

将每篇中文主稿归入以下四个状态之一：

| 状态 | 判断条件 | 处理方式 |
|------|----------|----------|
| **已同步** | 镜像路径存在且内容与中文版基本一致 | ✓ 无需操作 |
| **内容不同步** | 镜像路径存在，但内容与中文版差异较大 | 按翻译规范重新翻译并覆盖 |
| **路径不匹配** | 镜像路径不存在，但 slug 在其他 i18n/en 文件中匹配到 | 提交用户决策——可能是有意重命名 |
| **缺失** | 镜像路径不存在，且 slug 在任何 i18n/en 文件中均未找到 | 需要创建英文翻译 |

#### 内容差异判断标准

当文件存在于镜像路径时，对比中英文内容：

1. **读取中文源文件**的正文内容（去除 frontmatter）
2. **读取对应英文文件**的正文内容
3. 判断为**差异较大**（需要重新翻译）的条件：
   - 英文文件的章节数量与中文明显不符（差异超过 1 个一级标题）
   - 英文文件缺少中文中存在的主要内容段落
   - 英文文件中存在大量未翻译的中文字符（代码块外）
   - 英文文件内容篇幅不足中文的 60%（排除代码块）
   - frontmatter `title` 或正文标题明显机械直译、不符合英语技术文档习惯，或与同类页面标题风格明显不一致
   - frontmatter `title` 与 `slug` / 路径所表达的英文路由名称明显不一致，且内容没有支持这种差异
4. 判断为**基本一致**（无需操作）的条件：
   - 章节结构对应
   - 核心内容已翻译
   - frontmatter `title` 和正文标题自然、专业、准确，符合英语技术文档命名习惯
   - 少量措辞差异属正常范围

#### 标题与路由一致性判断标准

检查英文文档时，必须同时审查 frontmatter `title` 与正文 `##`/`###` 标题。标题不只是翻译结果，也是站点导航、侧边栏、搜索和 SEO 的入口。

先为每篇文档推导**英文路由名称候选**：

1. 优先读取 frontmatter `slug`，取最后一个有意义的路径段。
   - `/docs/architecture/core-host` → `core-host` → `Core Host`
   - `/quickstart` → `quickstart` → `Quickstart` 或按站点习惯写作 `Quick Start`
2. 如果没有 `slug`，从文件名推导，先去掉排序数字和后缀。
   - `1000-ai-native.md` → `ai-native` → `AI-Native`
   - `2000-wasm-plugins.md` → `wasm-plugins` → `WASM Plugins`
3. 如果路由段是产品名、API 名、协议名或固定术语，按既有英文写法保留大小写。
   - `openapi` → `OpenAPI`
   - `wasm` → `WASM`
   - `i18n` → `i18n`
4. 将该候选与中文标题和全文内容一起判断。标题不要求机械等于路由段，但应表达同一个页面入口概念。

标题命名优先级：

1. **路由名称优先**：当 `slug` / 路径已经给出清晰英文概念时，frontmatter `title` 优先采用该概念的自然标题形式。
2. **内容校准**：如果中文标题比路由更具体，或路由只是缩写，阅读正文后补足必要限定词。
3. **英语读者优先**：不要为了贴合中文标题而牺牲英语自然度；也不要为了贴合路由而产生命名不清的标题。
4. **一致性优先**：同目录、同类型页面保持标题风格一致；必要时读取邻近 2 到 3 篇英文文档作为参照。

优先使用英语技术文档中自然的名词短语或动作短语：
- `slug: /docs/architecture/core-host` 的标题应接近 `Core Host`，不要译成 `Core Main Machine`。
- `slug: /docs/development/commands` 的标题应接近 `Development Commands`，不要仅按中文写成 `Commands for Development`。
- 中文“开发指令”应译为 `Development Commands`，不要机械写成 `Dev Commands`，除非同一导航体系明确使用短标签。
- 中文“OpenAPI接口文档”更自然的是 `OpenAPI Reference`，不要写成 `OpenAPI Interface Documentation`。
- 中文“服务配置管理”应表达为 `Service Configuration` 或 `Service Configuration Management`，不要只写 `Configuration` 导致语义过泛。
- 中文“原生多租户能力”应表达为 `Native Multi-Tenancy`，不要省略关键定位。
- 正文标题避免重复、笨重或直译表达，例如用 `Scheduled Task Execution` 代替 `Scheduled Task Scheduling`，用 `Importing into Third-Party Tools` 代替 `Third-Party Tool Import`。

标题润色原则：
- **准确**：保留中文标题的技术指向和范围，不为了简短丢失核心限定词。
- **自然**：优先选择英语技术文档常见说法，例如 `Reference`、`Configuration`、`Execution`、`Development`、`Internationalization`。
- **专业**：避免口语缩写和内部简称，如 `Dev`，除非该页面是窄导航标签或同级文档统一采用该表达。
- **一致**：同一目录下同类标题使用一致风格，例如 `Development Commands`、`Development Tools`。
- **贴近路由**：`slug` / 路径中的英文路由名通常代表站点希望用户记住的页面名称，标题应优先采用其自然可读形式。
- **不过度翻译**：产品名、命令、API 名称、配置键名、插件 ID 和约定术语保持原样。

---

### 4. 报告差异分析

在执行任何修改前，先向用户展示结构化摘要：

```
## i18n 同步报告

### 已同步（无需操作）
- docs/quick/1000-overview.md ✓
- docs/community/1000-community.md ✓
...

### 内容不同步（需要重新翻译）
- docs/docs/1000-concepts/1000-ai-native.md
  原因：英文版缺少"规范驱动开发"章节，且篇幅仅为中文版的 45%
  操作：按翻译规范重新翻译并覆盖

### 路径不匹配（slug 匹配到其他路径）
- docs/docs/1000-concepts/1000-ai-native.md
  Slug: /docs/ai-native
  发现于：i18n/en/.../docs/core-concepts/built-in-capabilities.md
  → 请确认：是否有意为之？（y = 跳过，n = 在镜像路径重新创建）

### 缺失（需新建英文翻译）
- docs/docs/2000-architecture/1000-core-host.md  [slug: /docs/architecture/core-host]
- docs/docs/3000-plugin-development/2000-wasm-plugins.md  [slug: /docs/plugin-development/wasm]
...

总计：X 已同步，Y 内容不同步，Z 路径不匹配，W 缺失
```

**在执行前向用户确认：**
- 当前变更同步清单中的修改、新增或重命名 Markdown 文档：报告将同步，但不要询问是否跳过
- 非当前变更范围内的内容不同步：确认是否重新翻译并覆盖
- 路径不匹配：确认是跳过还是在镜像路径新建
- 非当前变更范围内的缺失：确认是否继续翻译
- 删除的中文文档：确认是否删除对应英文 i18n 文件

---

### 5. 翻译或重新翻译文件

对每个确认需要处理的文件（包括**缺失**和**内容不同步**两类）：

**a. 声明开始：** `## 翻译中：<source-path>`

**b. 完整阅读中文源文件** — 通读全文，理解：
- 文档主题及其在 LinaPro 体系中的定位
- 技术术语及其准确英文对应词
- slug / 文件名表达的英文路由名称，以及它与中文标题、正文主题之间的关系
- 文档语气（概念介绍、操作指引、参考手册等）
- 结构元素：标题层级、代码块、提示框、Mermaid 图表

**c. 以地道英文撰写** — 把自己当作原创技术文档作者，而非翻译者：

| 翻译原则 | 说明 |
|----------|------|
| 先理解再动笔 | 通读全文后再写第一个英文单词 |
| 自然表达 | 使用英语技术写作中惯用的句式结构 |
| 保持技术精度 | LinaPro 专属术语、API 名称、配置键名保持原样 |
| 匹配语气 | 概念文档保持概念性，操作指引保持步骤性 |
| 标题与结构 | 保持相同层级结构；frontmatter `title` 应尽可能采用 slug/路径路由名称的自然英文形式，正文标题必须自然、专业、准确，必要时重写为英语技术文档习惯表达 |
| 代码块 | 绝不翻译代码、代码注释或 CLI 命令 |
| Mermaid 图表 | 翻译节点标签和边标签，保留图表类型和语法 |
| Frontmatter | 地道翻译 `title`、`description`、`keywords`；`slug` 和 `sidebar_position` 保持不变 |

**d. Frontmatter 处理示例：**

```yaml
# 中文源文件
---
slug: '/docs/architecture/core-host'
title: '核心主机'
hide_title: true
description: '本文介绍...'
keywords:
  - 核心主机
  - 插件
---

# 英文输出
---
slug: '/docs/architecture/core-host'      # ← 不变
title: 'Core Host'                         # ← 自然翻译
hide_title: true                           # ← 原样保留
description: 'This page covers...'        # ← 意译（地道英文）
keywords:                                  # ← 翻译为英文术语
  - core host
  - plugin
---
```

标题翻译时不要把中文词序硬搬到英文中。先根据 `slug` 或文件名推导英文路由名称，再结合全文内容决定自然标题。允许在不改变含义的前提下改写为更自然的英文技术文档标题：

```yaml
# 中文源文件
slug: '/docs/reference/openapi'
title: 'OpenAPI接口文档'

# 英文输出
slug: '/docs/reference/openapi'
title: 'OpenAPI Reference'
```

再例如：

```yaml
# 中文源文件
slug: '/docs/architecture/core-host'
title: '核心主机'

# 英文输出
slug: '/docs/architecture/core-host'
title: 'Core Host'
```

**e. 写入文件：**
- 确保父目录存在：`mkdir -p <parent-dir>`
- 将翻译后内容写入镜像路径（缺失时新建，内容不同步时覆盖）

**f. 完成声明：** `✓ 完成：<i18n-path>`

---

### 6. 验证输出

所有翻译完成后：

```bash
# 确认期望的文件均已存在
find apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current -name "*.md" | sort
```

重新执行差异分析，确认缺失和不同步文件均已归零。

输出最终摘要：

```
## 同步完成

- 新建翻译：N 篇
- 重新翻译（内容不同步）：M 篇
- 跳过（已同步）：K 篇
- 用户决策（路径不匹配）：P 篇

### 本次处理的文件
- apps/lina-site/i18n/en/.../docs/2000-architecture/1000-core-host.md（新建）
- apps/lina-site/i18n/en/.../docs/1000-concepts/1000-ai-native.md（重新翻译）
...
```

---

## 翻译质量检查清单

写入每个译文文件前，在脑内逐项确认：

- [ ] 已完整阅读并理解中文文档
- [ ] 已从 `slug` 或路径推导路由名称，并确认英文 `title` 与该路由概念一致
- [ ] 英文读起来对母语者自然流畅
- [ ] 技术术语与现有英文文档保持一致
- [ ] frontmatter `title` 自然、专业、准确，不是机械直译
- [ ] 正文 `##`/`###` 标题自然、专业、准确，与同级文档风格一致
- [ ] 输出中不含中文字符（代码块内的中文除外）
- [ ] `slug` 与中文源文件完全一致
- [ ] 所有代码块内容未被修改
- [ ] 标题层级与中文版保持一致
- [ ] Mermaid 图表内容已适当翻译

---

## 执行守则

- **严禁修改 `slug` 字段** — 这是规范 URL，必须与中文源文件保持一致
- **通读全文再翻译** — 不允许逐段翻译，必须在理解全文后再动笔
- **标题先理解再命名** — 标题必须传达页面定位和技术范围，并尽可能贴近 `slug` / 路径中的英文路由名称；必要时意译或改写为英语技术文档常用标题
- **修改前先确认** — 展示差异报告并获得用户确认后再执行任何写入
- **路径不匹配须人工决策** — 不得静默跳过或覆盖，必须提交用户决策
- **禁止机械直译** — 若某句话或某个标题用英文说起来不自然，重写它，只要意思相同
- **保护代码完整性** — 不得修改代码块、行内代码或 CLI 示例中的任何内容
- **参考现有英文文档风格** — 翻译前先阅读 2～3 篇已有的 i18n/en 文档，匹配用词和语气
- **逐文件处理** — 声明 → 翻译 → 写入 → 确认，完成后再处理下一篇
- **内容不同步时覆盖整文** — 不做局部 patch，重新翻译整篇以确保一致性
