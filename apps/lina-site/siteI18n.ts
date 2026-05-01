export const DEFAULT_LOCALE = 'en';

export const SITE_LOCALES = ['en', 'zh-Hans'] as const;

type SiteLocale = (typeof SITE_LOCALES)[number];

type SiteSeo = {
  title: string;
  tagline: string;
  description: string;
  keywords: string;
};

type LocaleConfig = {
  label: string;
  direction: 'ltr' | 'rtl';
  htmlLang: string;
  calendar: string;
  path: string;
  baseUrl?: string;
  translate: boolean;
};

export const SITE_LOCALE_CONFIGS = {
  en: {
    label: 'English',
    direction: 'ltr',
    htmlLang: 'en-US',
    calendar: 'gregory',
    path: 'en',
    translate: true,
  },
  'zh-Hans': {
    label: '简体中文',
    direction: 'ltr',
    htmlLang: 'zh-CN',
    calendar: 'gregory',
    path: 'zh-Hans',
    baseUrl: '/zh/',
    translate: true,
  },
} satisfies Record<SiteLocale, LocaleConfig>;

const SITE_SEO = {
  en: {
    title:
      'LinaPro.AI - AI-native full-stack framework for sustainable delivery',
    tagline: 'AI-native full-stack framework for sustainable delivery',
    description:
      'LinaPro makes AI a core engine of delivery: AI leads analysis, design, and implementation while teams set direction and make the critical calls. With a core host service, admin workspace, plugin runtime, and spec-driven AI-native R&D workflow built in, teams can ship production-grade applications quickly while keeping architecture, testing, and governance ready to evolve.',
    keywords:
      'LinaPro,AI-native full-stack framework,sustainable delivery,AI-driven development,Go framework,Vue 3 admin,WASM plugins,RBAC,specification-driven,OpenSpec,AI-native R&D workflow,lina-core,lina-vben,lina-plugins,enterprise governance',
  },
  'zh-Hans': {
    title: 'LinaPro.AI - 面向可持续交付的 AI 原生全栈框架',
    tagline: '面向可持续交付的 AI 原生全栈框架',
    description:
      '把 AI 作为全栈研发的核心生产力，以高效、易用、可维护的方式帮助每一位开发者交付生产级应用',
    keywords:
      'LinaPro,AI 原生全栈框架,可持续交付,AI 驱动开发,Go 后端框架,Vue 3 管理后台,RBAC 权限,WASM 插件,源码插件,规范驱动开发,OpenSpec,AI 原生研发工作流,lina-core,lina-vben,lina-plugins',
  },
} satisfies Record<SiteLocale, SiteSeo>;

function isSiteLocale(locale: string | undefined): locale is SiteLocale {
  return SITE_LOCALES.includes(locale as SiteLocale);
}

export function getCurrentSiteSeo() {
  const locale = process.env.DOCUSAURUS_CURRENT_LOCALE;
  return SITE_SEO[isSiteLocale(locale) ? locale : DEFAULT_LOCALE];
}
