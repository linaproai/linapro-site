import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'LinaPro',
  tagline: 'AI-driven full-stack development framework',
  favicon: 'img/linapro-mark.svg',

  future: {
    v4: true,
  },

  url: 'https://linapro.ai',
  baseUrl: '/',
  trailingSlash: false,

  organizationName: 'linaproai',
  projectName: 'linapro-site',

  onBrokenLinks: 'throw',

  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'zh-CN'],
    localeConfigs: {
      en: {
        htmlLang: 'en-US',
      },
      'zh-CN': {
        htmlLang: 'zh-CN',
      },
    },
  },

  markdown: {
    format: 'mdx',
    hooks: {
      onBrokenMarkdownLinks: 'throw',
    },
  },

  stylesheets: [
    'https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500;600&family=IBM+Plex+Sans:wght@400;500;600;700&family=Noto+Sans+SC:wght@400;500;700&family=Sora:wght@400;500;600;700;800&display=swap',
  ],

  plugins: [
    [
      require.resolve('@easyops-cn/docusaurus-search-local'),
      {
        hashed: true,
        indexDocs: true,
        indexBlog: true,
        docsRouteBasePath: '/docs',
        blogRouteBasePath: '/blog',
        highlightSearchTermsOnTargetPage: true,
        explicitSearchResultPath: true,
        searchResultLimits: 8,
        searchBarShortcut: true,
      },
    ],
  ],

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          routeBasePath: 'docs',
          editUrl: 'https://github.com/linaproai/linapro-site/tree/main/apps/lina-site/',
        },
        blog: {
          showReadingTime: true,
          blogSidebarCount: 'ALL',
          blogSidebarTitle: 'Recent writing',
          feedOptions: {
            type: ['rss', 'atom'],
            xslt: true,
          },
          editUrl: 'https://github.com/linaproai/linapro-site/tree/main/apps/lina-site/',
          onInlineTags: 'warn',
          onInlineAuthors: 'warn',
          onUntruncatedBlogPosts: 'warn',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/linapro-social-card.svg',
    metadata: [
      {
        name: 'keywords',
        content:
          'LinaPro, AI-driven full-stack development framework, OpenSpec, plugin system, GoFrame, Docusaurus, monorepo',
      },
    ],
    colorMode: {
      defaultMode: 'light',
      disableSwitch: false,
      respectPrefersColorScheme: false,
    },
    navbar: {
      title: 'LinaPro',
      logo: {
        alt: 'LinaPro',
        src: 'img/linapro-wordmark.svg',
      },
      items: [
        {
          type: 'doc',
          docId: 'intro',
          position: 'left',
          label: 'Docs',
        },
        {
          to: '/blog',
          label: 'Blog',
          position: 'left',
        },
        {
          to: '/about',
          label: 'About',
          position: 'left',
        },
        {
          href: 'https://github.com/linaproai/linapro',
          label: 'GitHub',
          position: 'right',
        },
        {
          to: '/login',
          label: 'Login',
          position: 'right',
        },
        {
          to: '/register',
          label: 'Register',
          position: 'right',
        },
        {
          type: 'localeDropdown',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'light',
      links: [
        {
          title: 'Project / 项目',
          items: [
            {
              label: 'Docs',
              to: '/docs',
            },
            {
              label: 'Blog',
              to: '/blog',
            },
            {
              label: 'About',
              to: '/about',
            },
          ],
        },
        {
          title: 'Explore / 浏览',
          items: [
            {
              label: 'Architecture',
              to: '/docs/architecture',
            },
            {
              label: 'Plugins',
              to: '/docs/plugins',
            },
            {
              label: 'OpenSpec Workflow',
              to: '/docs/openspec-workflow',
            },
          ],
        },
        {
          title: 'Community / 社区',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/linaproai/linapro',
            },
            {
              label: 'Contact details soon',
              to: '/about',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} LinaPro.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'go', 'toml', 'yaml'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
