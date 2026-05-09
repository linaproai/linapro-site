<div align="center">
<img src="https://linapro.ai/img/linapro-logo.png" width="300" alt="linapro logo"/>

English | [简体中文](README.zh-CN.md)

</div>

# Overview

This repository contains the source code for the [LinaPro.AI](https://linapro.ai/) official website, built with [Docusaurus 3.10](https://docusaurus.io/). The `LinaPro` framework itself lives in a separate repository: [linaproai/linapro](https://github.com/linaproai/linapro).

# Quick Links

| Resource | URL |
|----------|-----|
| **Repository** | https://github.com/linaproai/linapro-site |
| **Website** | https://linapro.ai/ |
| **Framework** | https://github.com/linaproai/linapro |

# Tech Stack

| Category | Technology | Notes |
|----------|------------|-------|
| Framework | `Docusaurus` | `v3.10.0`, classic preset |
| Language | `TypeScript` | `v5.6.3` |
| Runtime | `Node.js` | `>= 20` |
| Package Manager | `yarn` | `pnpm` / `npm` also supported |
| Search | `Algolia DocSearch` | |
| Diagrams | `Mermaid` | Inline in `Markdown` |
| Comments | `Giscus` | `GitHub Discussions`-backed |

# Project Structure

```
linapro-site/
├── apps/lina-site/              # Docusaurus site
│   ├── docs/                    # Chinese primary docs (source of truth)
│   ├── i18n/                    # i18n resources (en, zh-Hans)
│   ├── src/                     # Custom pages & components
│   ├── static/                  # Static assets
│   ├── docusaurus.config.ts     # Site configuration
│   ├── sidebars.ts              # Sidebar definitions
│   └── siteI18n.ts              # Per-locale SEO metadata
├── openspec/                    # OpenSpec governance files
├── Makefile                     # Top-level dev/build commands
└── AGENTS.md                    # AI Coding Agent instructions
```

# Quick Start

**Prerequisites**: Node.js >= 20, with yarn (or pnpm / npm) installed.

```bash
# Clone the repository
git clone https://github.com/linaproai/linapro-site.git
cd linapro-site

# Install dependencies
cd apps/lina-site && yarn install

# Start the English site (default, localhost:3000/)
make dev locale=en

# Start the Chinese site (localhost:3000/zh/)
make dev
```

# Common Commands

| Command | Description |
|---------|-------------|
| `make dev` | Start local dev server (Chinese, default) |
| `make dev locale=en` | Start English site |
| `make build` | Build static site to `apps/lina-site/build/` |
| `make check` | Verify i18n translation completeness |
| `make webp` | Convert content images to WebP format |

# Internationalization

The site follows a **Chinese-first, English-synced** workflow:

- Primary docs are maintained in `apps/lina-site/docs/` (Chinese)
- English translations live in `apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/`
- The Chinese site (`/zh/`) reuses the Chinese primary docs directly
- Each locale has its own SEO metadata defined in `siteI18n.ts`

# Contributing

Documentation must follow the Chinese-first workflow: update the Chinese primary docs first, then sync to English and other locales.

# License

[Apache License 2.0](LICENSE)
