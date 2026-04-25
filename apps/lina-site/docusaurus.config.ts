const LATEST_VERSION_LABEL = 'v0.1.x(Latest)';

import type { Options as IdealImageOptions } from '@docusaurus/plugin-ideal-image';
import type * as Preset from '@docusaurus/preset-classic';
import type { Config } from '@docusaurus/types';
import { themes as prismThemes } from 'prism-react-renderer';

function isZhLocale() {
  return process.env.DOCUSAURUS_CURRENT_LOCALE === 'zh-Hans';
}

function geti18nTitle() {
  return isZhLocale()
    ? 'LinaPro - AI 驱动的全栈开发框架'
    : 'LinaPro - AI-driven full-stack development framework';
}

function geti18nTagline() {
  return isZhLocale()
    ? '面向可持续交付的 AI 驱动全栈开发框架——AI 主导执行，人类把握方向'
    : 'AI-driven full-stack development framework engineered for sustainable delivery — AI leads execution, humans steer direction.';
}

function geti18nDescription() {
  return isZhLocale()
    ? 'LinaPro 是一款 AI 驱动的全栈开发框架，由通用核心宿主服务（lina-core）、Vue 3 管理工作台（lina-vben）、双模式插件运行时（lina-plugins）与规范驱动的 AI 研发工作流（openspec）四层紧密协作组成，让 AI 在生产级软件交付中的效率随产品迭代持续保持。'
    : 'LinaPro is an AI-driven full-stack development framework: a Go core host service (lina-core), a Vue 3 management workspace (lina-vben), a dual-mode source/WASM plugin runtime (lina-plugins), and a specification-driven AI R&D workflow (openspec) — engineered so AI productivity compounds with every iteration of production software.';
}

function geti18nKeywords() {
  return isZhLocale()
    ? 'LinaPro,AI 全栈开发框架,AI 驱动开发,Go 后端框架,Vue 3 管理后台,RBAC 权限,WASM 插件,源码插件,规范驱动开发,OpenSpec,AI 研发工作流,lina-core,lina-vben,lina-plugins'
    : 'LinaPro,AI-driven development,full-stack framework,Go framework,Vue 3 admin,WASM plugins,RBAC,specification-driven,OpenSpec,AI R&D workflow,lina-core,lina-vben,lina-plugins,enterprise governance';
}

// https://docusaurus.io/docs/api/plugins/@docusaurus/plugin-content-docs#markdown-front-matter
// https://docusaurus.io/zh-CN/docs/api/docusaurus-config
const config: Config = {
  title: geti18nTitle(),
  tagline: geti18nTagline(),
  favicon: '/favicon.ico',
  url: 'https://linapro.ai/',
  baseUrl: '/',
  trailingSlash: false,
  organizationName: 'linaproai',
  projectName: 'linapro',
  onBrokenLinks: 'warn',
  // 多语言配置
  i18n: {
    defaultLocale: 'en',
    locales: ['en','zh-Hans'],
    path: 'i18n',
    localeConfigs: {
      'en': {
        label: 'English',
        direction: 'ltr',
        htmlLang: 'en-US',
        calendar: 'gregory',
        path: 'en',
      },
      'zh-Hans': {
        label: '简体中文',
        direction: 'ltr',
        htmlLang: 'zh-CN',
        calendar: 'gregory',
        path: 'zh-Hans',
        baseUrl: '/zh/',
      },
    },
  },
  // https://www.docusaurus.cn/blog/releases/3.6#docusaurus-faster
  future: {
    v4: {
      removeLegacyPostBuildHeadAttribute: true,
    },
    faster: true,
  },
  // 启用 Markdown 中的 Mermaid 支持
  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
    mermaid: true,
  },
  // 配置 Mermaid 主题
  themes: ['@docusaurus/theme-mermaid'],
  presets: [
    [
      'classic',
      {
        // Will be passed to @docusaurus/plugin-content-docs (false to disable)
        docs: {
          routeBasePath: '/',
          sidebarPath: require.resolve('./sidebars.ts'),
          lastVersion: 'current',
          // https://docusaurus.io/docs/versioning
          versions: {
            current: {
              label: LATEST_VERSION_LABEL,
            },
          },
          // 编辑当前页面的配置
          editUrl: 'https://github.com/linaproai/linapro/blob/main/apps/lina-site/',
          // 显示更新时间和作者
          showLastUpdateTime: true,
          showLastUpdateAuthor: true,
        },
        // Will be passed to @docusaurus/plugin-content-blog (false to disable)
        blog: {},
        // Will be passed to @docusaurus/plugin-content-pages (false to disable)
        pages: {},
        // Will be passed to @docusaurus/theme-classic.
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      } satisfies Preset.Options,
    ],
  ],
  plugins: [
    require.resolve('docusaurus-plugin-image-zoom'),
    [
      'ideal-image',
      {
        quality: 70,
        max: 1030,
        min: 640,
        steps: 2,
        // Use false to debug, but it incurs huge perf costs
        disableInDev: true,
      } satisfies IdealImageOptions,
    ],
  ],
  themeConfig: {
    metadata: [
      {
        name: 'keywords',
        content: geti18nKeywords(),
      },
      {
        name: 'description',
        content: geti18nDescription(),
      },
      {
        property: 'og:title',
        content: geti18nTitle(),
      },
      {
        property: 'og:description',
        content: geti18nDescription(),
      },
      {
        property: 'og:type',
        content: 'website',
      },
      {
        property: 'og:url',
        content: 'https://linapro.ai/',
      },
      {
        property: 'og:image',
        content: 'https://linapro.ai/img/linapro-logo.png',
      },
      {
        name: 'twitter:card',
        content: 'summary_large_image',
      },
      {
        name: 'twitter:title',
        content: geti18nTitle(),
      },
      {
        name: 'twitter:description',
        content: geti18nDescription(),
      },
      {
        name: 'twitter:image',
        content: 'https://linapro.ai/img/linapro-logo.png',
      },
    ],
    colorMode: {
      defaultMode: 'light',
      disableSwitch: false,
      respectPrefersColorScheme: false,
    },
    zoom: {
      selector: '.markdown :not(em) > img',
      config: {
        // options you can specify via https://github.com/francoischalifour/medium-zoom#usage
        background: {
          light: 'rgb(255, 255, 255)',
          dark: 'rgb(50, 50, 50)',
        },
      },
    },
    navbar: {
      title: '',
      logo: {
        alt: 'LinaPro Logo',
        src: '/img/logo-banner.png',
      },
      items: [
        {
          label: 'Get Started',
          position: 'left',
          // type: 'docSidebar',
          // sidebarId: 'quickSidebar',
          to: '/quickstart',
        },

        // 右边导航栏
        {
          type: 'docsVersionDropdown',
          position: 'right' as const,
          dropdownActiveClassDisabled: true,
        },
        {
          type: 'localeDropdown',
          position: 'right' as const,
        },
        {
          href: 'https://github.com/linaproai/linapro',
          position: 'right' as const,
          className: 'header-github-link',
          'aria-label': 'GitHub repository',
        },
      ],
    },
    // toc目录层级显示设置
    tableOfContents: {
      minHeadingLevel: 2,
      maxHeadingLevel: 3,
    },
    footer: {
      copyright: `Copyright ${new Date().getFullYear()} LinaPro`,
    },
    // 代码块配置
    prism: {
      theme: prismThemes.okaidia,
      darkTheme: prismThemes.vsDark,
      defaultLanguage: 'go',
      additionalLanguages: ['bash', 'javascript', 'toml', 'ini'], // 添加语言
      // 默认支持的语言 https://github.com/FormidableLabs/prism-react-renderer/blob/master/packages/generate-prism-languages/index.ts#L9-L23
      // 默认支持的语言 "markup","jsx","tsx","swift","kotlin","objectivec","js-extras","reason","rust","graphql","yaml","go","cpp","markdown","python","json"
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
