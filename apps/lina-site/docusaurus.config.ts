const LATEST_VERSION_LABEL = 'v0.1.x(Latest)';

import type { Options as IdealImageOptions } from '@docusaurus/plugin-ideal-image';
import type * as Preset from '@docusaurus/preset-classic';
import type { Config } from '@docusaurus/types';
import { themes as prismThemes } from 'prism-react-renderer';
import {
  DEFAULT_LOCALE,
  SITE_LOCALE_CONFIGS,
  SITE_LOCALES,
  getCurrentSiteSeo,
} from './siteI18n';

const siteSeo = getCurrentSiteSeo();

// https://docusaurus.io/docs/api/plugins/@docusaurus/plugin-content-docs#markdown-front-matter
// https://docusaurus.io/zh-CN/docs/api/docusaurus-config
const config: Config = {
  title: siteSeo.title,
  tagline: siteSeo.tagline,
  favicon: '/favicon.ico',
  url: 'https://linapro.ai/',
  baseUrl: '/',
  trailingSlash: false,
  organizationName: 'linaproai',
  projectName: 'linapro',
  onBrokenLinks: 'warn',
  // 多语言配置
  i18n: {
    defaultLocale: DEFAULT_LOCALE,
    locales: [...SITE_LOCALES],
    path: 'i18n',
    localeConfigs: SITE_LOCALE_CONFIGS,
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
        content: siteSeo.keywords,
      },
      {
        name: 'description',
        content: siteSeo.description,
      },
      {
        property: 'og:title',
        content: siteSeo.title,
      },
      {
        property: 'og:description',
        content: siteSeo.description,
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
        content: siteSeo.title,
      },
      {
        name: 'twitter:description',
        content: siteSeo.description,
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
          type: 'docSidebar',
          sidebarId: 'quickSidebar',
          label: '🚀 Get Started',
          position: 'left',
        },
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          label: '📖 Documentation',
          position: 'left',
        },
        {
          type: 'docSidebar',
          sidebarId: 'communitySidebar',
          label: '💬 Community',
          position: 'left',
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
      copyright: `Copyright ${new Date().getFullYear()} LinaPro.AI`,
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
