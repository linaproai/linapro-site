import ExecutionEnvironment from '@docusaurus/ExecutionEnvironment';

// 站点支持的语言列表
const SUPPORTED_LOCALES = ['en', 'zh-Hans'] as const;
type SupportedLocale = (typeof SUPPORTED_LOCALES)[number];

// localStorage key
const LOCALE_STORAGE_KEY = 'user-locale-preference';

// 浏览器语言到站点语言的映射
const LOCALE_MAP: Record<string, SupportedLocale> = {
  zh: 'zh-Hans',
  'zh-cn': 'zh-Hans',
  'zh-hans': 'zh-Hans',
  'zh-tw': 'zh-Hans',
  'zh-hant': 'zh-Hans',
  en: 'en',
  'en-us': 'en',
  'en-gb': 'en',
};

// 语言标识到 URL 路径的映射（与 docusaurus.config.ts 中的 localeConfigs 一致）
const LOCALE_URL_PATH: Record<SupportedLocale, string> = {
  en: '',
  'zh-Hans': '/zh',
};

/**
 * 从 localStorage 获取用户保存的语言偏好
 */
function getSavedLocalePreference(): SupportedLocale | null {
  if (!ExecutionEnvironment.canUseDOM) return null;
  
  try {
    const saved = localStorage.getItem(LOCALE_STORAGE_KEY);
    if (saved && SUPPORTED_LOCALES.includes(saved as SupportedLocale)) {
      return saved as SupportedLocale;
    }
  } catch {
    // localStorage 不可用时忽略
  }
  return null;
}

/**
 * 保存用户语言偏好到 localStorage
 */
function saveLocalePreference(locale: SupportedLocale): void {
  if (!ExecutionEnvironment.canUseDOM) return;
  
  try {
    localStorage.setItem(LOCALE_STORAGE_KEY, locale);
  } catch {
    // localStorage 不可用时忽略
  }
}

/**
 * 检测浏览器语言并返回匹配的站点语言
 */
function detectBrowserLocale(): SupportedLocale | null {
  if (!ExecutionEnvironment.canUseDOM) return null;
  
  // 优先使用 navigator.languages（用户偏好列表）
  const languages = navigator.languages || [navigator.language];
  
  for (const lang of languages) {
    const normalizedLang = lang.toLowerCase();
    
    // 精确匹配
    if (LOCALE_MAP[normalizedLang]) {
      return LOCALE_MAP[normalizedLang];
    }
    
    // 前缀匹配（如 zh-CN 匹配 zh）
    const prefix = normalizedLang.split('-')[0];
    if (LOCALE_MAP[prefix]) {
      return LOCALE_MAP[prefix];
    }
  }
  
  return null;
}

/**
 * 获取当前页面的语言
 */
function getCurrentLocale(): SupportedLocale {
  const path = window.location.pathname;
  
  // 检查路径中是否包含语言 URL 前缀
  if (path.startsWith('/zh/') || path === '/zh') {
    return 'zh-Hans';
  }
  
  // 默认返回英文
  return 'en';
}

/**
 * 执行语言重定向
 */
function redirectToLocale(targetLocale: SupportedLocale): void {
  const currentPath = window.location.pathname;
  const currentLocale = getCurrentLocale();
  
  // 如果目标语言与当前语言相同，不需要重定向
  if (targetLocale === currentLocale) return;
  
  // 构建新的 URL
  let newPath: string;
  
  // 移除当前语言的 URL 前缀
  const currentUrlPath = LOCALE_URL_PATH[currentLocale];
  let pathWithoutLocale: string;
  
  if (currentUrlPath && currentPath.startsWith(currentUrlPath)) {
    pathWithoutLocale = currentPath.substring(currentUrlPath.length) || '/';
  } else {
    pathWithoutLocale = currentPath || '/';
  }
  
  // 添加目标语言的 URL 前缀
  const targetUrlPath = LOCALE_URL_PATH[targetLocale];
  if (targetUrlPath) {
    newPath = `${targetUrlPath}${pathWithoutLocale}`;
  } else {
    newPath = pathWithoutLocale;
  }
  
  // 使用 replace 避免在历史记录中留下记录
  window.location.replace(newPath);
}

/**
 * 主函数：执行语言检测和重定向
 */
function initLocaleDetection(): void {
  if (!ExecutionEnvironment.canUseDOM) return;
  
  // 1. 检查用户是否已有保存的语言偏好
  const savedLocale = getSavedLocalePreference();
  if (savedLocale) {
    // 用户已有偏好，直接使用，不执行自动检测
    return;
  }
  
  // 2. 检测浏览器语言
  const detectedLocale = detectBrowserLocale();
  if (!detectedLocale) {
    // 无法检测到语言，使用默认语言
    return;
  }
  
  // 3. 获取当前页面语言
  const currentLocale = getCurrentLocale();
  
  // 4. 如果检测到的语言与当前语言不同，执行重定向
  if (detectedLocale !== currentLocale) {
    // 保存检测到的语言偏好，避免重复检测
    saveLocalePreference(detectedLocale);
    redirectToLocale(detectedLocale);
  }
}

// 监听语言切换事件，保存用户手动选择的语言
function setupLocaleChangeListener(): void {
  if (!ExecutionEnvironment.canUseDOM) return;
  
  // 监听点击事件，检测语言切换按钮
  document.addEventListener('click', (event) => {
    const target = event.target as HTMLElement;
    
    // 检查是否点击了语言切换按钮
    // Docusaurus 的 localeDropdown 会生成带有特定属性的链接
    const localeLink = target.closest('a[href*="/"], a[href*="/zh/"]');
    if (localeLink) {
      const href = localeLink.getAttribute('href');
      if (href) {
        // 根据 href 判断目标语言
        if (href.startsWith('/zh/') || href === '/zh') {
          saveLocalePreference('zh-Hans');
        } else if (href === '/' || href.startsWith('/en/') || href === '/en') {
          saveLocalePreference('en');
        }
      }
    }
  });
}

// 初始化
initLocaleDetection();
setupLocaleChangeListener();
