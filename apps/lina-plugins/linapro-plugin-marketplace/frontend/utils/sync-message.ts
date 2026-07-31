/**
 * Marketplace Git lastSyncMessage display helpers.
 *
 * Discovery and pipeline paths persist English source diagnostics in
 * `plugin_marketplace_plugin.last_sync_message`. UI locales must map known
 * patterns through plugin i18n keys so Chinese (and other locales) never show
 * the stored English source when a translation exists. Unknown free-form
 * diagnostics (err.Error / gitPublicErrorMessage) remain as fallback.
 */

/** Runtime i18n translate function used by Vue pages (`t` from useI18n). */
export type MarketplaceSyncTranslate = (
  key: string,
  params?: Record<string, unknown>,
) => string;

/** Optional `te` existence check from vue-i18n; when omitted, key equality is used. */
export type MarketplaceSyncTranslateExists = (key: string) => boolean;

const syncMessageKeyPrefix =
  "plugin.linapro-plugin-marketplace.detail.syncMessage.";

/**
 * Known sync diagnostic keys under `detail.syncMessage.*`.
 * Keep in sync with backend writers in marketplace_git.go / marketplace_process.go.
 */
export const MARKETPLACE_SYNC_MESSAGE_CODES = [
  "discoveredDrafts",
  "discoveredImmutableOnly",
  "discoveredWithFailures",
  "failedImportRefs",
  "credentialLoadFailed",
  "repositoryUrlInvalid",
  "platformTokenConfigFailed",
  "queuedForVerification",
  "docsIndexIncomplete",
  "immutableDocsIndexIncomplete",
] as const;

export type MarketplaceSyncMessageCode =
  (typeof MARKETPLACE_SYNC_MESSAGE_CODES)[number];

/**
 * Full runtime i18n keys for every known sync message pattern. Kept as a static
 * string list so i18n coverage checks can discover keys without executing Vue.
 */
export const MARKETPLACE_SYNC_MESSAGE_I18N_KEYS =
  MARKETPLACE_SYNC_MESSAGE_CODES.map(
    (code) => `${syncMessageKeyPrefix}${code}` as const,
  );

function syncMessageKey(code: MarketplaceSyncMessageCode): string {
  return `${syncMessageKeyPrefix}${code}`;
}

function translateOrFallback(
  t: MarketplaceSyncTranslate,
  key: string,
  fallback: string,
  params?: Record<string, unknown>,
  te?: MarketplaceSyncTranslateExists,
): string {
  if (typeof te === "function" && !te(key)) {
    return fallback;
  }
  const translated = params ? t(key, params) : t(key);
  if (!translated || translated === key) {
    return fallback;
  }
  return translated;
}

function withOptionalDetail(detail: string | undefined): string {
  const trimmed = (detail || "").trim();
  if (!trimmed) {
    return "";
  }
  return `: ${trimmed}`;
}

type SyncPattern = {
  re: RegExp;
  code: MarketplaceSyncMessageCode;
  params?: (match: RegExpMatchArray) => Record<string, unknown>;
};

/**
 * Patterns ordered from most specific to least. Match backend English source
 * text written by updateGitSyncStatus / setPluginProcessStatus.
 */
const SYNC_MESSAGE_PATTERNS: SyncPattern[] = [
  {
    re: /^discovered 0 new draft releases \((\d+) existing immutable version\(s\)\)$/,
    code: "discoveredImmutableOnly",
    params: (match) => ({ count: Number(match[1]) }),
  },
  {
    re: /^discovered (\d+) draft releases$/,
    code: "discoveredDrafts",
    params: (match) => ({ count: Number(match[1]) }),
  },
  {
    re: /^discovered (\d+) drafts with (\d+) ref failures(?::\s*(.*))?$/s,
    code: "discoveredWithFailures",
    params: (match) => ({
      count: Number(match[1]),
      failures: Number(match[2]),
      detail: withOptionalDetail(match[3]),
    }),
  },
  {
    re: /^failed to import (\d+) refs(?::\s*(.*))?$/s,
    code: "failedImportRefs",
    params: (match) => ({
      count: Number(match[1]),
      detail: withOptionalDetail(match[2]),
    }),
  },
  {
    re: /^release imported; documentation indexing incomplete(?::\s*(.*))?$/s,
    code: "docsIndexIncomplete",
    params: (match) => ({ detail: withOptionalDetail(match[1]) }),
  },
  {
    re: /^immutable release kept; documentation indexing incomplete(?::\s*(.*))?$/s,
    code: "immutableDocsIndexIncomplete",
    params: (match) => ({ detail: withOptionalDetail(match[1]) }),
  },
  {
    re: /^credential load failed$/,
    code: "credentialLoadFailed",
  },
  {
    re: /^repository url is invalid$/,
    code: "repositoryUrlInvalid",
  },
  {
    re: /^platform git token config failed$/,
    code: "platformTokenConfigFailed",
  },
  {
    re: /^new git draft queued for verification$/,
    code: "queuedForVerification",
  },
];

/**
 * Localize one Git/pipeline lastSyncMessage for the current UI locale.
 *
 * Prefer known English source patterns → `detail.syncMessage.*` so historical
 * rows and new discovery writes both render in the active locale. Unknown
 * free-form diagnostics fall back to the stored source text.
 */
export function formatMarketplaceLastSyncMessage(
  t: MarketplaceSyncTranslate,
  message: null | string | undefined,
  te?: MarketplaceSyncTranslateExists,
): string {
  const fallback = (message || "").trim();
  if (!fallback) {
    return "";
  }
  for (const pattern of SYNC_MESSAGE_PATTERNS) {
    const match = fallback.match(pattern.re);
    if (!match) {
      continue;
    }
    const key = syncMessageKey(pattern.code);
    const params = pattern.params ? pattern.params(match) : undefined;
    return translateOrFallback(t, key, fallback, params, te);
  }
  return fallback;
}
